package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// cliHelperEnv makes the child half of a process-level check run the real
// dispatch boundary, the same pattern cmd/devilutionGo uses for
// appfat's exit codes.
const cliHelperEnv = "DEVILUTIONGO_MIGRATE_HELPER"

// cliArgsEnv carries the arguments the child passes to run().
const cliArgsEnv = "DEVILUTIONGO_MIGRATE_ARGS"

func TestCLIUnknownSubcommandExitsTwo(t *testing.T) {
	// when the command runs with a subcommand that does not exist
	code, stderr := runChild(t, "extract")

	// then the process exits 2 and prints the usage text (ADR-014: one tool,
	// three subcommands)
	if code != exitUsage {
		t.Fatalf("exit code = %d, want %d", code, exitUsage)
	}

	if !strings.Contains(stderr, "usage: devilutionGo-migrate") {
		t.Errorf("stderr = %q, want the usage text", stderr)
	}
}

func TestCLIPendingLaneExitsTwo(t *testing.T) {
	// when the saves subcommand runs (its lane is a separate bean)
	code, stderr := runChild(t, "saves")

	// then the process exits 2 and says the lane has not landed
	if code != exitUsage {
		t.Fatalf("exit code = %d, want %d", code, exitUsage)
	}

	if !strings.Contains(stderr, "saves: not implemented yet") {
		t.Errorf("stderr = %q, want the not-implemented report", stderr)
	}
}

func TestCLIAssetsWithoutArchivesExitsOne(t *testing.T) {
	// when the assets subcommand runs with no archive configured for the child
	code, stderr := runChild(t, "assets", "--dry-run", t.TempDir())

	// then a migration that cannot run exits 1, distinct from a usage error
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitFailure, stderr)
	}

	if !strings.Contains(stderr, assetArchivesEnv) {
		t.Errorf("stderr = %q, want it to name %s", stderr, assetArchivesEnv)
	}
}

// TestMigrateCLIHelper is the child half of the process-level checks: it runs the
// real dispatch and turns its code into the process exit status.
func TestMigrateCLIHelper(t *testing.T) {
	// given the child is the command under test, not a test
	if os.Getenv(cliHelperEnv) != "1" {
		t.Skip("child process only; driven by the TestCLI* process checks")
	}

	// when the dispatch runs the arguments the parent chose
	code := run(strings.Fields(os.Getenv(cliArgsEnv)), os.Stdout, os.Stderr)

	// then the child ends with that code and the parent asserts on it
	os.Exit(code) //nolint:revive // child process stands in for main()
}

// runChild runs the command in a child process and reports its exit code and
// stderr. The environment is cleared of the archive setting so the child cannot
// pick up the caller's retail archives.
func runChild(t *testing.T, args ...string) (int, string) {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestMigrateCLIHelper$")
	cmd.Env = append(childEnv(), cliHelperEnv+"=1", cliArgsEnv+"="+strings.Join(args, " "))

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatalf("child did not run: %v (output: %s)", err, out.String())
	}

	return cmd.ProcessState.ExitCode(), out.String()
}

// childEnv is the parent environment without the archive setting the child must
// not inherit.
func childEnv() []string {
	env := make([]string, 0, len(os.Environ()))

	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, assetArchivesEnv+"=") {
			env = append(env, entry)
		}
	}

	return env
}
