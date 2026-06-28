# Getting Started

## Prerequisites

| Requirement | Minimum version | Notes |
|---|---|---|
| Go toolchain | 1.24.6 | Required to build from source |
| Git | Any | Required to clone the repo |

No runtime dependencies — the compiled binary is statically linked and self-contained.

---

## Installation

### Option 1 — Build from source (recommended during development)

```bash
# Clone the repo
git clone https://github.com/allenbiji/preboot.git
cd preboot

# Build and place the binary on your PATH
go build -o /usr/local/bin/preboot ./cmd/sage

# Verify
preboot --version
```

### Option 2 — `go install`

If the module is published to a Go module proxy:

```bash
go install github.com/allenbiji/preboot/cmd/sage@latest
```

The binary lands in `$(go env GOPATH)/bin/preboot`. Make sure that directory is on your `$PATH`.

### Option 3 — Download a pre-built release binary

Once releases are published to GitHub Releases:

```bash
# Linux amd64 example
curl -L https://github.com/allenbiji/preboot/releases/latest/download/sage-linux-amd64 \
  -o /usr/local/bin/preboot
chmod +x /usr/local/bin/preboot
```

See [Deployment Guide](deployment.md) for how to build and publish release binaries.

---

## First Run

### Step 1 — Navigate to your project

```bash
cd /path/to/your-go-project
```

### Step 2 — Generate a baseline config

```bash
preboot init
```

`preboot init` scans the current directory for known frameworks and generates `preboot-auto.yml` with relevant checks:

- **Go project** (`go.mod` found) → adds `go-installed` check
- **Makefile** found → adds `make-installed` check
- **Docker Compose** (`docker-compose.yml` / `compose.yaml`) → adds `docker-installed` + port-free checks for every mapped port
- **Environment file** (`.env.example`, `.env.template`) → adds `env-file-exists` + one `env_exists` check per key

Example output:

```
✅  Detected go project - adding go checks
✅  Detected docker compose - adding docker checks  
✅  Detected .env.example - adding env checks
✅  preboot-auto.yml written (12 checks)

Run 'preboot check' to validate your setup.
```

### Step 3 — Add custom checks

Open or create `preboot.yml` in the same directory. This file is merged on top of `preboot-auto.yml`:

```yaml
version: 1

checks:
  - name: redis-running
    type: tcp_reachable
    severity: blocker
    options:
      address: "localhost:6379"
    message: "Redis is not running"
    fix: "Start Redis with: docker compose up redis -d"

  - name: stripe-key-set
    type: env_exists
    severity: blocker
    options:
      key: STRIPE_SECRET_KEY
```

### Step 4 — Run diagnostics

```bash
preboot check
```

Review any failures, apply the suggested fixes, and re-run until all blockers pass.

---

## Environment variable controls

| Variable | Effect |
|---|---|
| `NO_COLOR` | Disables ANSI color output (any non-empty value) |
| `TERM=dumb` | Also disables color output |

These are honoured automatically — no flag needed.

---

## Typical project workflow

```
┌─────────────────────────────────────────────────┐
│  Developer clones the repo for the first time   │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
            preboot init   (generates preboot-auto.yml if missing)
                     │
                     ▼
            cp .env.example .env
            # fill in secrets
                     │
                     ▼
            preboot check  ──► all green? start coding
                     │
                     ▼ (if red)
            follow Fix: guidance
            re-run preboot check
```

Commit `preboot-auto.yml` and `preboot.yml` alongside your code so every contributor gets instant feedback.
