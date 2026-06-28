# Preboot Documentation

Welcome to the **Preboot** (`preboot`) documentation. This index links to every topic you need to understand, use, extend, or deploy the tool.

## What is Preboot?

Preboot is a Go CLI that diagnoses local setup failures in Go repositories. Instead of debugging "it works on my machine" by hand, you define — or auto-generate — a YAML file listing health checks for your project, then run a single command to validate the entire environment.

```
preboot check
```

A typical run looks like:

```
  ██████╗ ██████╗ ███████╗██████╗  ██████╗  ██████╗ ████████╗
  ██╔══██╗██╔══██╗██╔════╝██╔══██╗██╔═══██╗██╔═══██╗╚══██╔══╝
  ██████╔╝██████╔╝█████╗  ██████╔╝██║   ██║██║   ██║   ██║
  ██╔═══╝ ██╔══██╗██╔══╝  ██╔══██╗██║   ██║██║   ██║   ██║
  ██║     ██║  ██║███████╗██████╔╝╚██████╔╝╚██████╔╝   ██║
  ╚═╝     ╚═╝  ╚═╝╚══════╝╚═════╝  ╚═════╝  ╚═════╝   ╚═╝

Running Preboot Diagnostics...

✅ go-installed
✅ docker-installed
✅ env-file-exists
✅ DB_URL-configured
✅ port-free-5432
❌ API_KEY-configured — key API_KEY not found in .env
   Fix: Copy .env.example to .env and fill in missing values

5 passed, 1 failed
```

---

## Documentation Index

| Document | What it covers |
|---|---|
| [Getting Started](getting-started.md) | Installation, first run, quick start |
| [Configuration Reference](configuration.md) | YAML schema, merging logic, defaults |
| [CLI Commands](commands.md) | `preboot init`, `preboot check`, `preboot validate` |
| [Check Types](checks.md) | All 7 check types with option reference and examples |
| [Architecture](architecture.md) | Package layout, data flow, extension points |
| [Contributing](contributing.md) | Dev setup, testing, adding new check types |
| [Testing Guide](testing.md) | Full test inventory, helpers, and test patterns |
| [Deployment Guide](deployment.md) | Building, distributing, and shipping Preboot |

---

## Quick-reference card

```bash
# Auto-generate checks from your repo
preboot init

# Run all checks
preboot check

# Skip slow network checks
preboot check --quick

# Use a specific config file
preboot check --config path/to/preboot.yml

# Validate config syntax without running checks
preboot validate
```

Exit codes: `0` = all blockers passed · `1` = one or more blockers failed · `2` = internal error
