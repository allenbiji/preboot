# Test Suite

This document covers how to run the tests, what each package tests, the helpers available, and the patterns every test must follow.

---

## Running tests

```bash
make test              # full suite — race detector on, every test once (matches CI)
make test-short        # same but without race detector (faster local iteration)

# Target a single package
go test ./internal/checks/... -v

# Run a single test or subtest
go test ./internal/config/... -run TestLoad_BothFiles -v
go test ./internal/checks/... -run TestHttpReachableCheck_Execute/200_OK -v
```

---

## Package coverage

| Package | Test file(s) | What is tested |
|---|---|---|
| `internal/engine` | `run_test.go`, `color_test.go` | Check execution loop, severity handling, quick mode, timeout injection, ANSI color toggling |
| `internal/config` | `load_test.go`, `validate_test.go`, `merge_test.go` | YAML loading, two-file merge, config validation, defaults merging |
| `internal/detect` | `detect_test.go`, `docker_test.go`, `files_test.go` | Auto-detection of Go, Makefile, Docker Compose, and `.env` files; env key extraction |
| `internal/checks` | one `*_test.go` per check type, `testutil_test.go` | Build-time option validation and Execute() behaviour for all 7 check types |
| `internal/registry` | `registry_test.go` | Register/build/lookup lifecycle, duplicate-registration panic, unknown-type error |

The `cmd/preboot` and `internal/cli` packages have no test files — see [Good first issues](contributing.md) for how to contribute coverage there.

---

## Test inventory

### `internal/engine`

**`run_test.go`**

| Test | What it covers |
|---|---|
| `TestRun_EmptyChecks` | Empty checks slice returns nil |
| `TestRun_AllPass` | Single passing check returns nil |
| `TestRun_BlockerFails` | Blocker failure returns `ErrCheckFailed` |
| `TestRun_WarningNonStrict` | Warning in non-strict mode returns nil |
| `TestRun_WarningStrictMode` | Warning in strict mode returns `ErrCheckFailed` |
| `TestRun_InfoNeverBlocks` | Info severity never causes failure |
| `TestRun_UnknownCheckType` | Unknown check type returns `ErrCheckFailed` |
| `TestRun_QuickModeSkipsHttp` | `--quick` skips `http_reachable` |
| `TestRun_QuickModeSkipsTcp` | `--quick` skips `tcp_reachable` |
| `TestRun_GlobalTimeoutInjected` | Global `timeout_ms` is propagated to checks |
| `TestRun_OwnTimeoutNotOverridden` | Per-check `timeout_ms` takes precedence over global |

**`color_test.go`**

| Test | What it covers |
|---|---|
| `TestColorize_Enabled` | ANSI codes added when `colorEnabled` is true |
| `TestColorize_Disabled` | Plain text returned when `colorEnabled` is false |

---

### `internal/config`

**`load_test.go`**

| Test | What it covers |
|---|---|
| `TestLoadFrom_PathNotFound` | Error when file does not exist |
| `TestLoadFrom_InvalidYAML` | Error on malformed YAML |
| `TestLoadFrom_ValidConfig` | Successful YAML load |
| `TestLoadFrom_EmptyPath_FallsBackToLoad` | Empty path triggers `Load()` fallback |
| `TestLoad_NeitherFile` | Error when both config files missing |
| `TestLoad_OnlySageYml` | Loads from `preboot.yml` only |
| `TestLoad_OnlySageAutoYml` | Loads from `preboot-auto.yml` only |
| `TestLoad_BothFiles` | Merges checks from both files (user overrides auto) |
| `TestLoad_AutoParseError_SageMissing` | Error when auto file is invalid and user file absent |
| `TestLoad_SageParseError_AutoMissing` | Error when user file is invalid and auto file absent |
| `TestLoad_InvalidVersionRejects` | Config with `version: 2` is rejected |

**`validate_test.go`**

| Test | What it covers |
|---|---|
| `TestValidateConfig` | Table-driven: version out of range, blank/whitespace names, invalid severity, unknown check type |
| `TestValidateConfig_MultipleErrors` | Multiple validation errors accumulate into a single message |

**`merge_test.go`**

| Test | What it covers |
|---|---|
| `TestMergeDefaults` | Table-driven: nil defaults, empty map, `strict`, `timeout_ms` |

---

### `internal/detect`

**`detect_test.go`**

| Test | What it covers |
|---|---|
| `TestExtractEnvKeys_NoEqualsSign` | Lines without `=` are skipped |
| `TestExtractEnvKeys_MultipleEquals` | Values containing `=` are handled correctly |
| `TestDetectGo_GoModPresent` | `go.mod` triggers `go-installed` check |
| `TestDetectGo_GoModAbsent` | No `go.mod` → no checks emitted |
| `TestDetectEnv_NoFile` | Neither env template file → no checks |
| `TestDetectEnv_ExamplePriority` | `.env.example` takes precedence over `.env.template` |
| `TestDetectEnv_TemplateFallback` | `.env.template` used when `.env.example` missing |
| `TestGenerateEnvChecks_MissingFile` | Missing template file → only `file_exists` check emitted |
| `TestScanRepo_EmptyDir` | Empty directory → `version: 1`, no checks |
| `TestScanRepo_GoProject` | `go.mod` → `go-installed` check in output |
| `TestScanRepo_WithMakefile` | `Makefile` → `make-installed` check in output |

**`docker_test.go`**

| Test | What it covers |
|---|---|
| `TestDetectDockerCompose_NoFile` | No compose file → empty slice |
| `TestDetectDockerCompose_YmlPresent` | `docker-compose.yml` → `docker-installed` + `port-free-*` checks |
| `TestDetectDockerCompose_YamlFallback` | `compose.yaml` used when `docker-compose.yml` absent |
| `TestDetectDockerCompose_EnvVarPort` | `${PORT}` style port references are skipped |
| `TestDetectDockerCompose_InvalidYAML` | Invalid YAML still emits `docker-installed` check |
| `TestExtractHostPort` | Table-driven: `host:container`, `ip:host:container`, env-var formats |

**`files_test.go`**

| Test | What it covers |
|---|---|
| `TestExtractEnvKeys` | Table-driven: comments, blank lines, inline comments, whitespace, empty values, missing file |

---

### `internal/checks`

Each check type has a `TestBuild*` (option validation) and a `TestCheck_Execute` (runtime behaviour) test.

| Test | What it covers |
|---|---|
| `TestBuildFileCheck` | Rejects absolute paths, `~`, `..`; accepts relative and nested paths |
| `TestFileCheck_Execute` | File present, file missing, path is a directory |
| `TestBuildDirectoryExistsCheck` | Rejects `~`, `..`; accepts relative paths |
| `TestDirectoryCheck_Execute` | Directory present, directory missing, path is a file |
| `TestEnvCheck_Execute` | Key present, key missing, empty value |
| `TestBuildEnvExistsCheck` | Missing key option, missing `.env` file, valid build, concurrent cache access (20 goroutines), cache reuse |
| `TestBuildPortFreeCheck` | Missing port, empty port, non-numeric, out-of-range (0, 65536, negative), valid bounds (1, 65535) |
| `TestPortFreeCheck_Execute` | Port in use (real listener), port free |
| `TestBuildTcpReachableCheck` | Missing address, default timeout (5 s), custom `timeout_ms`, invalid fallback, missing port, empty host |
| `TestTcpReachableCheck_Execute` | Real listener open, port closed |
| `TestBuildHttpReachableCheck` | Missing address, default timeout, custom `timeout_ms`, invalid fallback, scheme validation (`http`/`https` only) |
| `TestHttpReachableCheck_Execute` | 200 OK, 301 redirect (error), 404, 500, unreachable, timeout exceeded |

---

### `internal/registry`

| Test | What it covers |
|---|---|
| `TestRegister_Build` | Registers a factory, builds an instance |
| `TestRegister_DuplicatePanics` | Duplicate registration panics with type name in message |
| `TestBuild_UnknownType` | Unknown type returns `"unknown check type"` error |
| `TestIsKnownType` | Distinguishes registered from unregistered types |

---

## Helpers

| Helper | Location | What it does |
|---|---|---|
| `chdir(t, dir)` | `detect/detect_test.go`, `detect/docker_test.go`, `config/load_test.go` | `os.Chdir(dir)` with `t.Cleanup` to restore the original directory |
| `cfg(typ, opts)` | `checks/testutil_test.go` | Builds a minimal `model.CheckConfig` for check tests |
| `passCfg(name, sev)` | `engine/run_test.go` | `CheckConfig` that targets `"go"` (always in PATH) |
| `failCfg(name, sev)` | `engine/run_test.go` | `CheckConfig` that targets a nonexistent command |
| `validCfg()` | `config/validate_test.go` | Returns a minimal valid `PrebootConfig` |
| `stubCheck` | `registry/registry_test.go` | Minimal `registry.Check` implementation for registry tests |

---

## Patterns all tests must follow

These apply to every new test in the repo:

1. **Parallel** — call `t.Parallel()` at the top of every `Test*` function and every `t.Run` subtest.
2. **Table-driven** — use `[]struct{ name, ... }` + `t.Run(tc.name, ...)` for any test with more than one case.
3. **TempDir** — use `t.TempDir()` for any test that touches the filesystem; never write to the real working directory.
4. **chdir** — use the local `chdir(t, dir)` helper (not a bare `os.Chdir`) so the working directory is always restored.
5. **Error assertions** — check `strings.Contains(err.Error(), expected)` rather than exact equality, so error messages can evolve without breaking tests.
6. **No third-party frameworks** — standard `testing` package only; no testify, gomock, etc.
7. **Real I/O over mocks** — for network tests, bind a real `net.Listen("tcp", ":0")` rather than mocking; for HTTP, use `httptest.NewServer`.
