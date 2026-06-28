# Deployment Guide

This document covers everything you need to build, package, and distribute the `preboot` binary to end users.

---

## Overview of distribution options

| Method | Best for | Effort |
|---|---|---|
| `go install` | Go developers | Minimal — just publish the module |
| GitHub Releases (pre-built binaries) | All users, no Go required | Medium — add a release workflow |
| Homebrew tap | macOS/Linux users who use `brew` | Medium — create a tap repo |
| Docker image | CI environments, containers | Medium — write a Dockerfile |
| `apt`/`rpm` package | Linux sysadmin environments | High |
| Internal binary registry | Enterprise teams | Varies |

Most open-source CLI tools start with **`go install` + GitHub Releases** and add a Homebrew tap as adoption grows.

---

## Step 1 — Ensure the module is publicly accessible

Preboot's module path is:

```
github.com/allenbiji/preboot
```

This must match the actual GitHub repository URL. Verify in [go.mod](../go.mod):

```
module github.com/allenbiji/preboot
```

Make sure the repository at `https://github.com/allenbiji/preboot` is **public**. Private repos cannot be fetched by `go install` or the module proxy.

---

## Step 2 — Tag a release

Go's module proxy and `go install` resolve versions by **git tags** of the form `vMAJOR.MINOR.PATCH`:

```bash
# Make sure you are on main and it is clean
git checkout main
git pull

# Create and push an annotated tag
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
```

After pushing the tag, users can install immediately:

```bash
go install github.com/allenbiji/preboot/cmd/preboot@v0.1.0
# or the latest tagged version:
go install github.com/allenbiji/preboot/cmd/preboot@latest
```

---

## Step 3 — Build release binaries (GitHub Releases)

Many users do not have Go installed. Provide pre-built binaries via GitHub Releases.

### Add a GoReleaser workflow

[GoReleaser](https://goreleaser.com) is the standard tool for cross-compiling and publishing Go binaries.

**Install GoReleaser locally (for testing):**

```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

**Create `.goreleaser.yaml` in the repo root:**

```yaml
# .goreleaser.yaml
version: 2

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: preboot
    main: ./cmd/preboot
    binary: preboot
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: preboot
    name_template: "sage_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - LICENSE

checksum:
  name_template: "checksums.txt"

release:
  github:
    owner: allenbiji
    name: preboot
  draft: false
  prerelease: auto

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

**Create `.github/workflows/release.yml`:**

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # GoReleaser needs full git history

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**To publish a release:**

```bash
git tag -a v0.2.0 -m "v0.2.0"
git push origin v0.2.0
```

The workflow triggers automatically, cross-compiles for all platforms, and creates a GitHub Release with downloadable `.tar.gz` / `.zip` archives and a `checksums.txt`.

---

## Step 4 — Embed version information

Update `cmd/preboot/main.go` to read the version from build-time ldflags:

```go
package main

import (
    "github.com/allenbiji/preboot/internal/cli"
)

// Set by GoReleaser via ldflags
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    cli.SetVersion(version, commit, date)
    if err := cli.Execute(); err != nil {
        os.Exit(2)
    }
}
```

In `internal/cli/root.go`, expose the version via the `--version` flag:

```go
rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
```

---

## Step 5 — Homebrew tap (optional but recommended)

A Homebrew tap lets macOS and Linux users install with:

```bash
brew install allenbiji/tap/preboot
```

### Create the tap repository

1. Create a new GitHub repository named `homebrew-tap` under your account: `github.com/allenbiji/homebrew-tap`
2. Create `Formula/preboot.rb` in that repo:

```ruby
class Sage < Formula
  desc "Diagnose local setup failures in Go repositories"
  homepage "https://github.com/allenbiji/preboot"
  version "0.2.0"

  on_macos do
    on_arm do
      url "https://github.com/allenbiji/preboot/releases/download/v#{version}/sage_#{version}_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_FROM_CHECKSUMS_TXT"
    end
    on_intel do
      url "https://github.com/allenbiji/preboot/releases/download/v#{version}/sage_#{version}_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_FROM_CHECKSUMS_TXT"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/allenbiji/preboot/releases/download/v#{version}/sage_#{version}_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_FROM_CHECKSUMS_TXT"
    end
    on_intel do
      url "https://github.com/allenbiji/preboot/releases/download/v#{version}/sage_#{version}_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_FROM_CHECKSUMS_TXT"
    end
  end

  def install
    bin.install "preboot"
  end

  test do
    system "#{bin}/preboot", "--version"
  end
end
```

**GoReleaser can automate tap updates.** Add a `brews` block to `.goreleaser.yaml`:

```yaml
brews:
  - name: preboot
    repository:
      owner: allenbiji
      name: homebrew-tap
    homepage: "https://github.com/allenbiji/preboot"
    description: "Diagnose local setup failures in Go repositories"
    install: |
      bin.install "preboot"
    test: |
      system "#{bin}/preboot --version"
```

Now every release automatically opens a PR on `homebrew-tap` with the updated formula.

---

## Step 6 — Docker image (optional)

Useful for CI environments and teams that don't install tools directly on hosts.

**Create `Dockerfile`:**

```dockerfile
# Build stage
FROM golang:1.24.6-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /preboot ./cmd/preboot

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /preboot /usr/local/bin/preboot
ENTRYPOINT ["preboot"]
```

**Build and push:**

```bash
docker build -t ghcr.io/allenbiji/preboot:v0.2.0 .
docker push ghcr.io/allenbiji/preboot:v0.2.0
```

**Use in GitHub Actions CI:**

```yaml
- name: preboot check
  run: |
    docker run --rm \
      -v ${{ github.workspace }}:/workspace \
      -w /workspace \
      ghcr.io/allenbiji/preboot:latest check
```

---

## CI deployment checklist

Before shipping a release, make sure all of the following pass:

```
go build ./...           ← binary compiles
go vet ./...             ← no vet warnings
go test ./... -race -count=1  ← all tests pass
goreleaser check         ← GoReleaser config is valid
```

Add these as a pre-release gate in GitHub Actions by running them in the existing `ci.yml` and making the `release.yml` job depend on `ci.yml` passing first:

```yaml
# release.yml
jobs:
  release:
    needs: [ci]   # gate: only release if CI passes
    ...
```

---

## Versioning policy

Follow [Semantic Versioning](https://semver.org/):

| Change | Version bump |
|---|---|
| Backward-compatible new feature (new check type, new flag) | Minor: `v0.2.0 → v0.3.0` |
| Backward-compatible bug fix | Patch: `v0.2.0 → v0.2.1` |
| Breaking change to YAML schema or CLI flags | Major: `v0.x.y → v1.0.0` |

The YAML schema is currently `version: 1`. Increment it if you make breaking changes to the config format and update the validator accordingly.

---

## Summary: deployment sequence

```
1. go test ./... -race     ← tests pass
2. git tag v0.X.Y          ← create version tag
3. git push origin vX.X.Y  ← triggers release.yml
4. GoReleaser:
     - cross-compiles linux/darwin/windows × amd64/arm64
     - creates GitHub Release with binaries + checksums
     - (optional) updates Homebrew tap
5. Users install via:
     go install ...@latest
     brew install allenbiji/tap/preboot
     curl -L .../preboot-linux-amd64 | install
```
