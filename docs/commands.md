# CLI Commands Reference

All commands are subcommands of the `preboot` binary.

---

## Global behaviour

- A gradient ASCII art banner is printed to stdout before every command **when stdout is a terminal**. When piped or redirected, the banner is suppressed.
- All diagnostic output goes to **stdout**.
- Errors (internal failures, bad flags) go to **stderr**.

---

## `preboot init`

Scans the current working directory for known project patterns and writes a `preboot-auto.yml` file with auto-detected checks.

### Synopsis

```
preboot init [flags]
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--force` | `-f` | `false` | Overwrite an existing `preboot-auto.yml` |

### Behaviour

1. Checks whether `preboot-auto.yml` already exists. If it does and `--force` is not set, the command exits with a friendly message.
2. Calls `detect.ScanRepo()` which inspects the working directory for:
   - `go.mod` → Go checks
   - `Makefile` → make checks
   - `docker-compose.yml` / `compose.yaml` → Docker + port checks
   - `.env.example`, `.env.template` → environment variable checks
3. Writes `preboot-auto.yml` in the current directory.
4. Prints which checks were added and suggests running `preboot check`.

### Output example

```
✅  Detected go project    — added: go-installed
✅  Detected docker-compose — added: docker-installed, port-free-5432, port-free-6379
✅  Detected .env.example   — added: env-file-exists, DB_URL-configured, SECRET_KEY-configured

preboot-auto.yml written (6 checks).
Run 'preboot check' to validate your setup.
```

### Exit codes

| Code | Meaning |
|---|---|
| `0` | File written successfully (or already existed without `--force`) |
| `2` | Internal error (filesystem write failure, etc.) |

### Notes

- `preboot-auto.yml` is meant to be committed to version control so every developer gets baseline checks for free.
- You should never manually edit `preboot-auto.yml`; use `preboot.yml` for custom overrides instead.

---

## `preboot check`

Loads configuration and runs all checks, printing pass/fail results.

### Synopsis

```
preboot check [flags]
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--config PATH` | `-c` | _(auto)_ | Path to a custom YAML config file. When set, only this file is loaded — no merging. |
| `--quick` | `-q` | `false` | Skip network checks (`http_reachable`, `tcp_reachable`) for a faster local run. |

### Config resolution (without `--config`)

The engine looks for two files in the current directory and merges them:

```
preboot-auto.yml   +   preboot.yml   →   effective config
   (auto)           (user)
```

- If only `preboot-auto.yml` exists → use it.
- If only `preboot.yml` exists → use it.
- If both exist → merge: user-defined checks that share a `name` with an auto check override that auto check; all others are appended.
- If neither exists → error with guidance to run `preboot init`.

### Execution flow

1. Load and merge config.
2. Apply defaults (`strict: true`, `timeout_ms: 3000` unless overridden).
3. Iterate over checks in declaration order.
4. For each check, resolve the factory from the registry, execute it, and print the result.
5. Print a summary line.
6. Exit with the appropriate code.

### Output format

```
Running Preboot Diagnostics...

✅  go-installed
✅  docker-installed
ℹ️   optional-service-running — could not connect: dial tcp [::1]:9999: connection refused
⚠️  make-installed — command "make" not found in PATH
    Fix: brew install make
❌  DB_URL-configured — key DB_URL not found in .env
    Fix: Copy .env.example to .env and set DB_URL

5 passed, 1 failed
```

**Severity icons:**

| Icon | Severity | Strict mode effect |
|---|---|---|
| ✅ | (passed) | — |
| ℹ️  | `info` | Never a blocker |
| ⚠️  | `warning` | Blocker when `strict: true` (default) |
| ❌ | `blocker` | Always a blocker |

### Exit codes

| Code | Meaning |
|---|---|
| `0` | All blocker-level checks passed |
| `1` | One or more blocker-level checks failed |
| `2` | Internal error (bad config, registry error, etc.) |

This makes `preboot check` safe to use in CI pipelines:

```yaml
- name: Validate dev environment
  run: preboot check
```

---

## `preboot validate`

Parses and validates the YAML config without running any checks. Useful for CI linting or pre-commit hooks.

### Synopsis

```
preboot validate [flags]
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--config PATH` | `-c` | _(auto)_ | Custom config file path. Same resolution logic as `preboot check`. |

### Validation rules

1. YAML must parse without errors.
2. `version` field must equal `1`.
3. Every check must have a non-empty `name`.
4. `severity` must be one of: `info`, `warning`, `blocker`.
5. `type` must be a registered check type (see [Check Types](checks.md)).

### Output

```
✅ Configuration is valid!
```

Or, on failure:

```
❌ Configuration error: check "my-check": unknown check type "typo_check"
```

### Exit codes

| Code | Meaning |
|---|---|
| `0` | Config is valid |
| `1` | Validation error |
| `2` | Internal error (file not found, YAML parse error, etc.) |

---

## Shell completion (future)

Cobra supports automatic shell completion. Once the project exposes it:

```bash
preboot completion bash  >> ~/.bashrc
preboot completion zsh   >> ~/.zshrc
preboot completion fish  > ~/.config/fish/completions/preboot.fish
```
