// version.go
// Add this file to the ROOT of your DockLens repo.
//
// GoReleaser injects Version, Commit, and BuildDate at compile time
// using the ldflags in .goreleaser.yaml. This file receives those values.
//
// Usage: docklens --version  →  docklens v1.0.0 (abc1234, built 2025-01-15)

package main

import (
	"fmt"
	"os"
)

// These are set by GoReleaser via -ldflags at build time.
// The default values show when you build locally with `go run .`
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// Call this from your main() when --version flag is passed
func printVersion() {
	fmt.Printf("docklens %s (%s, built %s)\n", Version, Commit, BuildDate)
}

// Add this check at the very top of your main() function:
//
//   if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
//       printVersion()
//       os.Exit(0)
//   }

func init() {
	// Auto-handle --version / -v before Bubbletea starts
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		printVersion()
		os.Exit(0)
	}
}
