// Package version reports the CLI build version.
package version

import "runtime/debug"

// version is injected at build time via:
//
//	-ldflags "-X github.com/allenbiji/preboot/internal/version.version=v1.2.3"
var version = ""

// Version returns the release version embedded at build time, falling back
// to the module version recorded by the Go toolchain (set when installed
// via `go install ...@version`), then to "dev" for plain local builds.
func Version() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
