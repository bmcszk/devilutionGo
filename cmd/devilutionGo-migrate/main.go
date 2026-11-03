// Command devilutionGo-migrate migrates retail Diablo data into the form the
// game reads: MPQ archives into the assets dir plus the CSV tables and YAML
// asset list (ADR-013), and, in their own lanes, saves into SQLite and
// diablo.ini settings (ADR-014).
package main

import (
	"fmt"
	"io"
	"os"
)

// Exit codes: 2 mirrors the flag package's usage-error convention, 1 marks a
// migration that ran and failed, so a caller script can tell them apart.
const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

const usage = `usage: devilutionGo-migrate <subcommand> [flags]

subcommands:
  assets    extract MPQ archives into the assets dir, CSV tables and manifest.yaml
  saves     import .sv/.sp saves into SQLite (not implemented yet)
  settings  carry the diablo.ini settings over (not implemented yet)

Run "devilutionGo-migrate assets -h" for the assets flags.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches one subcommand and reports the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return writeUsage(stderr)
	}

	switch args[0] {
	case "assets":
		return runAssets(args[1:], stdout, stderr)
	case "saves", "settings":
		return pendingSubcommand(args[0], stderr)
	default:
		return writeUsage(stderr)
	}
}

// pendingSubcommand reports a subcommand whose lane has not landed yet.
func pendingSubcommand(name string, stderr io.Writer) int {
	_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate: %s: not implemented yet\n", name)

	return exitUsage
}

// writeUsage prints the top-level usage and returns the usage exit code.
func writeUsage(stderr io.Writer) int {
	_, _ = fmt.Fprint(stderr, usage)

	return exitUsage
}
