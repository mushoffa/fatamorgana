package fatamorgana

import "runtime/debug"

// modulePath is the import path of this module, as it appears in a consumer's
// build information.
const modulePath = "github.com/mushoffa/fatamorgana"

// development is reported when no release version can be determined, for
// example when the library is built from a local checkout or a replace
// directive.
const development = "(Development)"

// Version returns the version of this module that the running binary was built
// with, as recorded by the Go toolchain (for example "v0.6.0").
func Version() string {
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return development
	}

	return moduleVersion(build)
}

// moduleVersion extracts this module's version from build information.
func moduleVersion(build *debug.BuildInfo) string {
	if build == nil {
		return development
	}

	// This module is the main module, for example when building from inside the
	// repository itself.
	if build.Main.Path == modulePath {
		return normalize(build.Main.Version)
	}

	// This module is a dependency of the binary being built.
	for _, dep := range build.Deps {
		if dep == nil || dep.Path != modulePath {
			continue
		}

		// A replace directive swaps the module for another source. A replacement
		// without a version is a local path, which has no release version.
		if dep.Replace != nil {
			return normalize(dep.Replace.Version)
		}

		return normalize(dep.Version)
	}

	return development
}

// normalize maps the empty string to the development marker.
func normalize(version string) string {
	if version == "" {
		return development
	}

	return version
}
