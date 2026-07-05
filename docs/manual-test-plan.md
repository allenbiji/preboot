# Manual Test Plan — PreBoot Release Validation

Run this plan against a tagged release (e.g. `v0.1.0`) before announcing it.
Every step is copy-pasteable. Record results in the sign-off table at the end.

**Exit code contract** (assert with `echo $?` after each command):

| Code | Meaning |
|------|---------|
| 0    | No blocker failed (info failures never affect the exit code) |
| 1    | A **blocker** check failed, a **warning** check failed while `defaults.strict` is true, or a check hit an internal error (e.g. `env_exists` with no `.env` file) |
| 2    | Config or usage error |

Note on `strict`: handwritten configs default to `strict: true` (warnings
fail the run); `preboot init` writes `strict: false` into auto-configs
(warnings report but don't fail).

---

## 1. Installation matrix

Each install path must produce a binary that runs and reports the release
version — not `dev`.

| # | Path | Platform | Steps | Expected |
|---|------|----------|-------|----------|
| I1 | Release tarball | macOS (arm64 + Intel if available) | Download `preboot_0.1.0_darwin_arm64.tar.gz`, `tar xzf`, `xattr -d com.apple.quarantine ./preboot` if Gatekeeper blocks it, run `./preboot --version` | `preboot version 0.1.0` |
| I2 | Release tarball | Linux | Download `preboot_0.1.0_linux_amd64.tar.gz` from the release, `tar xzf`, run `./preboot --version` | `preboot version 0.1.0` |
| I3 | Release zip | Windows | Download `preboot_0.1.0_windows_amd64.zip`, extract, run `preboot.exe --version` | `preboot version 0.1.0` |
| I4 | go install | any | `go install github.com/allenbiji/preboot/cmd/preboot@v0.1.0` | `preboot --version` → `preboot version v0.1.0` (via Go build info) |
| I5 | Checksums | any | `sha256sum -c checksums.txt` against the downloaded archive | OK |

(Homebrew tap is deferred — add install/upgrade rows here when it ships.)

---

## 2. Fixture repositories

Create these once under a scratch directory (`mkdir -p ~/preboot-qa && cd ~/preboot-qa`).

### R1 — Go repo (go.mod + Makefile)

```bash
mkdir r1-go && cd r1-go && git init -q
printf 'module example.com/demo\n\ngo 1.25\n' > go.mod
printf 'build:\n\tgo build ./...\n' > Makefile
cd ..
```

### R2 — Docker Compose repo (published ports)

```bash
mkdir r2-compose && cd r2-compose && git init -q
cat > docker-compose.yml <<'EOF'
services:
  db:
    image: postgres:16
    ports:
      - "5432:5432"
  cache:
    image: redis:7
    ports:
      - "6379:6379"
EOF
cd ..
```

### R3 — Env repo (.env.example)

```bash
mkdir r3-env && cd r3-env && git init -q
printf 'DATABASE_URL=postgres://localhost/dev\nAPI_KEY=changeme\n' > .env.example
cd ..
```

### R4 — Kitchen sink (all detectors at once)

Combine R1 + R2 + R3 files in one repo. Also test the `compose.yaml`
filename variant here: after the first pass, rename
`docker-compose.yml` → `compose.yaml`, re-run `preboot init --force`, and
confirm the same checks are generated.

### R5 — Empty repo (no detectable artifacts)

```bash
mkdir r5-empty && cd r5-empty && git init -q && cd ..
```

### R6 — Handwritten config covering all seven check types

```bash
mkdir r6-alltypes && cd r6-alltypes && git init -q
printf 'API_KEY=secret\n' > .env        # env_exists reads keys from .env, not the shell env
cat > preboot.yml <<'EOF'
version: 1
checks:
  - name: git-installed
    type: command_exists
    severity: blocker
    options: { command: git }
  - name: api-key-env
    type: env_exists
    severity: warning
    options: { key: API_KEY }
  - name: license-file
    type: file_exists
    severity: info
    options: { path: LICENSE }
  - name: src-dir
    type: directory_exists
    severity: warning
    options: { folder: src }
  - name: example-http
    type: http_reachable
    severity: warning
    options: { address: "https://example.com", timeout_ms: "3000" }
  - name: dns-tcp
    type: tcp_reachable
    severity: warning
    options: { address: "1.1.1.1:53", timeout_ms: "3000" }
  - name: dev-port-free
    type: port_free
    severity: blocker
    options: { port: "8080" }
EOF
cd ..
```

(Option keys verified against `internal/checks/*.go`: `command`, `key`,
`path`, `folder`, `address`, `port`, `timeout_ms`.)

### R7 — Broken configs

```bash
mkdir r7-broken && cd r7-broken && git init -q
printf 'version: 2\nchecks: []\n' > bad-version.yml
printf 'version: 1\nchecks:\n  - name: x\n    type: made_up_type\n' > bad-type.yml
printf 'version: 1\n  checks: [broken\n' > bad-yaml.yml
cd ..
```

---

## 3. Functional test cases

Run inside each fixture repo. `V` = also verify exit code.

### init

| # | Repo | Command | Expected |
|---|------|---------|----------|
| F1 | R1 | `preboot init` | stderr reports detected requirements; `preboot-auto.yml` created containing `command_exists` checks for `go` and `make` |
| F2 | R1 | `preboot init` (again) | Error: `preboot-auto.yml already exists — use --force to overwrite`; exit 2 (V) |
| F3 | R1 | `preboot init --force` | Overwrites without error |
| F4 | R2 | `preboot init` | Auto-config includes `docker-installed` (blocker) plus port checks for 5432 and 6379 |
| F5 | R3 | `preboot init` | Auto-config includes `env_exists` checks for `DATABASE_URL` and `API_KEY` |
| F6 | R4 | `preboot init` | All of F1+F4+F5 checks in one file; repeat after renaming to `compose.yaml` |
| F7 | R5 | `preboot init` | "No recognised frameworks found. Generating empty baseline." — file still written; exit 0 (V) |

### check

| # | Repo | Command | Expected |
|---|------|---------|----------|
| F8 | R1 (after init) | `preboot check` | All pass (go + make installed); exit 0 (V) |
| F9a | R6 | `preboot check` | `src-dir` fails as warning; strict defaults to **true** for handwritten configs, so exit 1 (V) |
| F9b | R6 | add `defaults:\n  strict: false` to `preboot.yml`, run `preboot check` | Same warning failure but exit 0; `license-file` (info) failure never affects exit code (V) |
| F9c | R6 | temporarily `mv .env .env.bak`, run `preboot check` | `api-key-env` reports "internal error: could not read .env"; exit 1 even with `strict: false` (V). Restore `.env` after. |
| F10 | R6 | occupy port first: `python3 -m http.server 8080 &`, then `preboot check` | `dev-port-free` (blocker) fails; exit 1 regardless of strict (V). Kill the server after. |
| F11 | R6 | `preboot check --quick -f json` | `example-http` and `dns-tcp` have `"status": "skipped"` in JSON (they're omitted from text output); port/env/file checks still execute |
| F12 | R6 | `preboot check -f json \| python3 -m json.tool` | Valid JSON, one entry per executed check with name/severity/status |
| F13 | R6 | `preboot check -c preboot.yml` | Same as F9a (explicit path) |
| F14 | R4 | add a `preboot.yml` overriding one auto check, run `preboot check` | stderr shows "Merging preboot-auto.yml with preboot.yml..."; explicit config wins for the overridden check |
| F15 | R5 (no configs, no init) | `preboot check` | Clear error about missing config; exit 2 (V) |

### validate

| # | Repo | Command | Expected |
|---|------|---------|----------|
| F16 | R6 | `preboot validate` | Config reported valid; exit 0 (V) |
| F17 | R7 | `preboot validate -c bad-version.yml` | `unsupported config version: 2`; exit 2 (V) |
| F18 | R7 | `preboot validate -c bad-type.yml` | Validation error naming the unknown check type; exit 2 (V) |
| F19 | R7 | `preboot validate -c bad-yaml.yml` | YAML parse error, not a panic; exit 2 (V) |

### Cross-cutting

| # | Command | Expected |
|---|---------|----------|
| F20 | `preboot --version` | Version only — no banner above it |
| F21 | `preboot --help` / bare `preboot` | Usage listing init/check/validate |
| F22 | `preboot check -f json > out.json` (redirected/non-TTY) | JSON output not corrupted by color codes |

---

## 4. Sign-off

| Case | macOS (tarball) | Linux (tarball) | Windows (zip) | go install |
|------|-----------------|-----------------|---------------|------------|
| I1–I5 | | | | |
| F1–F7 (init) | | | | |
| F8–F15 (check) | | | | |
| F16–F19 (validate) | | | | |
| F20–F22 | | | | |

Tester: ____________  Date: ____________  Release: ____________
