# Contributing & Development

---

## Development setup

### Prerequisites

- Go 1.24.6 or newer (`go version`)
- Git

### Clone and build

```bash
git clone https://github.com/allenbiji/preboot.git
cd preboot
make build

# Run it
./preboot --help
```

### Run tests

```bash
make test          # all tests with race detector (matches CI)

# Specific package or single test (raw go for targeting)
go test ./internal/checks/... -v
go test ./internal/config/... -run TestLoad_BothFiles -v
```

See [Testing Guide](testing.md) for the full test inventory, available helpers, and the patterns every test must follow.

### Vet and lint

```bash
make vet
```

There is no linter configuration in the repo beyond the standard `go vet`. If you add one (e.g. `golangci-lint`), document it here.

---

## Project conventions

### Makefile

Run `make help` to see all available targets. Use `make ci` before pushing to run the full build + vet + test chain locally.

### Write tests for everything you add

Every new check type, detection rule, config behaviour, or engine change must come with tests. The repo's test patterns are documented in [Testing Guide](testing.md). As a checklist:

- New check type → `TestBuild*` (option validation) + `TestCheck_Execute` (runtime behaviour) in `internal/checks/<type>_test.go`
- New detect rule → tests in `internal/detect/detect_test.go` using `t.TempDir` + `chdir`
- New config field or merge rule → table-driven case in `merge_test.go` or `validate_test.go`
- Engine change → case in `run_test.go`

If a PR adds behaviour with no test coverage it will not be merged.

### Test patterns

- All tests use the standard `testing` package — no third-party test frameworks.
- Table-driven tests with `t.Run` subtest names.
- `t.Parallel()` at the top of every test function and subtests.
- `t.TempDir()` for temporary files — automatically cleaned up.
- `t.Cleanup()` for anything that can't use `TempDir`.
- A `chdir(t, dir)` helper in `testutil_test.go` changes the working directory for a test and restores it afterwards.

### Error messages

Follow the pattern `verb "subject": reason` — for example:

```
command "go" not found in PATH
key "DB_URL" not found in .env
port 5432 is already in use
```

### YAML field names

Use `snake_case` in YAML to match the existing schema (`timeout_ms`, `check_type`, etc.).

---

## Adding a new check type

Follow these steps to add, say, a `process_running` check.

### 1. Define the check type constant

Open [internal/model/config.go](../internal/model/config.go) and add your type to the `CheckType` constants:

```go
const (
    EnvExists       CheckType = "env_exists"
    CommandExists   CheckType = "command_exists"
    // ... existing types ...
    ProcessRunning  CheckType = "process_running"  // add this
)
```

### 2. Create the check file

Create `internal/checks/process.go`:

```go
package checks

import (
    "fmt"

    "github.com/allenbiji/preboot/internal/model"
    "github.com/allenbiji/preboot/internal/registry"
)

func init() {
    registry.Register(model.ProcessRunning, newProcessCheck)
}

type processCheck struct {
    name string
}

func newProcessCheck(cfg model.CheckConfig) (registry.Check, error) {
    name := cfg.Options["name"]
    if name == "" {
        return nil, fmt.Errorf("process_running: 'name' option is required")
    }
    return &processCheck{name: name}, nil
}

func (c *processCheck) Execute() error {
    // Your implementation here.
    // Return nil on success, error on failure.
    // The error message is shown in the output.
    return fmt.Errorf("process %q is not running", c.name)
}
```

### 3. Write tests

Create `internal/checks/process_test.go`:

```go
package checks_test

import (
    "testing"

    "github.com/allenbiji/preboot/internal/model"
    "github.com/allenbiji/preboot/internal/registry"
)

func TestProcessCheck_MissingOption(t *testing.T) {
    t.Parallel()
    cfg := model.CheckConfig{Type: "process_running", Options: map[string]string{}}
    _, err := registry.Build(cfg)
    if err == nil {
        t.Fatal("expected error for missing 'name' option")
    }
}

func TestProcessCheck_Execute(t *testing.T) {
    t.Parallel()
    cfg := model.CheckConfig{
        Type:    "process_running",
        Options: map[string]string{"name": "nonexistent-process-xyz"},
    }
    check, err := registry.Build(cfg)
    if err != nil {
        t.Fatalf("unexpected build error: %v", err)
    }
    if err := check.Execute(); err == nil {
        t.Fatal("expected failure for nonexistent process")
    }
}
```

### 4. Update the config validator (if needed)

`internal/config/validate.go` calls `registry.IsKnownType()` to validate type names. Since your new type is registered in `init()`, it is automatically recognized — **no changes needed** to the validator.

### 5. Update auto-detection (optional)

If your new check type can be auto-detected (e.g., detected from a config file), add detection logic in `internal/detect/`. Follow the pattern in `detect/go.go` or `detect/docker.go`:

```go
// internal/detect/process.go
package detect

import "github.com/allenbiji/preboot/internal/model"

func detectProcesses(cfg *model.PrebootConfig) {
    // examine current directory, add checks to cfg.Checks
}
```

Call your function from `detect.ScanRepo()` in `internal/detect/repo.go`.

### 6. Update documentation

Add a section to [docs/checks.md](checks.md) documenting the new type's options, validation rules, failure messages, and examples.

### 7. Run tests

```bash
make ci
```

---

## CI pipeline

The GitHub Actions workflow at [.github/workflows/ci.yml](../.github/workflows/ci.yml) runs on every push to `main` and every PR targeting `main`:

```yaml
steps:
  - uses: actions/checkout@v4
  - uses: actions/setup-go@v5
    with:
      go-version-file: go.mod
  - run: go build ./...
  - run: go vet ./...
  - run: go test ./... -race -count=1
```

All three steps must pass. PRs that break the build, introduce vet warnings, or fail tests will not be merged.

---

## Dependency management

```bash
# Add a dependency
go get github.com/some/package@v1.2.3

# Remove unused dependencies
make tidy

# Verify the module graph
make verify
```

Commit both `go.mod` and `go.sum` with every dependency change.

---

## Commit style

This repo accepts two equivalent styles — use whichever fits the commit best:

### Conventional commits

```
feat: add process_running check type
fix: handle .env files with Windows line endings
docs: add architecture diagram
test: cover negative cases for port_free
refactor: extract timeout resolution into helper
```

### Go module style (package-scoped)

Prefix with the package or sub-system being changed, followed by a short imperative description. This mirrors the style used in the Go standard library and many Go open-source projects:

```
checks: add process_running check type
detect: auto-detect Node.js projects
engine: skip blank-name checks with a warning
config: reject version values outside 1–9
registry: panic message now includes type name
```

Both styles are accepted. Use mod-style when the change is clearly scoped to one package; use conventional commits for cross-cutting changes (`docs:`, `test:`, `chore:`).

---

## Pull request checklist

- [ ] `make ci` passes (build + vet + test)
- [ ] New behaviour is covered by tests (see [Testing Guide](testing.md))
- [ ] New public functions/types have a one-line doc comment
- [ ] New check types are documented in `docs/checks.md`
- [ ] `go.mod` and `go.sum` are up to date
