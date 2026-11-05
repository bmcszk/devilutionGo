package main

import (
	"strings"
	"testing"
)

func TestAssetsListPrintsEveryEntry(t *testing.T) {
	// the baseline inventory is the same for every shape
	root := t.TempDir()
	inventoryPath := writeInventory(t, root, fixtureInventory())

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "tab separated",
			args: []string{"assets", "list", "--manifest", inventoryPath},
			want: "ui.data\tdata/health.clx\tclx\t1\tSource/diablo.cpp:1359\n",
		},
		{
			name: "markdown",
			args: []string{"assets", "list", "--manifest", inventoryPath, "--markdown"},
			want: "| ui.data | data/health.clx | clx | 1 | Source/diablo.cpp:1359 |\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			code, stdout, stderr := runMigrate(t, test.args...)
			// then
			assertListRendering(t, code, stdout, stderr, test.want)
		})
	}
}

func TestAssetsCommandRejectsBadInvocations(t *testing.T) {
	// the nested asset subcommands must refuse these invocations
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown subcommand", args: []string{"frobnicate"}, want: `unknown command "frobnicate"`},
		{
			name: "convert without flags", args: []string{"convert"},
			want: "--format, --in and --out are required",
		},
		{
			name: "convert with a bad width list",
			args: []string{"convert", "--format", "cel", "--in", "x", "--out", "y", "--widths", "a,b"},
			want: `width "a" is not an integer`,
		},
		{
			name: "convert with an unknown format",
			args: []string{"convert", "--format", "gif", "--in", "x", "--out", "y"},
			want: `unknown format "gif"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// when
			code, _, stderr := runMigrate(t, append([]string{"assets"}, test.args...)...)
			// then
			assertUsageRejection(t, code, stderr, test.want)
		})
	}
}

func TestAssetsWithoutASubcommandRunsTheExtractionPass(t *testing.T) {
	// given no archive is configured anywhere
	t.Setenv(assetArchivesEnv, "")

	// when `assets` runs with neither a nested subcommand nor flags
	code, _, stderr := runMigrate(t, "assets")

	// then it takes the extraction path (the nested dispatch only claims named
	// subcommands), so a missing archive is a migration failure, not a usage
	// error: the old `dgo-assets` no-argument usage case now lives at the top
	// level, TestRunWithoutSubcommandPrintsUsage
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitFailure, stderr)
	}

	if !strings.Contains(stderr, assetArchivesEnv) {
		t.Fatalf("stderr = %q, want it to name %s", stderr, assetArchivesEnv)
	}
}

// assertListRendering checks one list run's exit code and rendering.
func assertListRendering(t *testing.T, code int, stdout, stderr, want string) {
	t.Helper()

	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, want) {
		t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
	}
}

// assertUsageRejection checks one bad assets invocation's exit code and report.
func assertUsageRejection(t *testing.T, code int, stderr, want string) {
	t.Helper()

	if code != exitUsage {
		t.Fatalf("exit code = %d, want %d", code, exitUsage)
	}

	if !strings.Contains(stderr, want) {
		t.Fatalf("stderr = %q, want it to contain %q", stderr, want)
	}
}
