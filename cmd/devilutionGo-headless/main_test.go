package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
)

// helperEnv makes the child half of the process-level check run.
const helperEnv = "APPFAT_HEADLESS_HELPER"

// Ports test/appfat_test.cpp:8-11 — app_fatal("test") exits 1 and puts the
// message text on stderr. The child runs the real exitOnError boundary.
func TestExitOnError_Fatal_ExitsOneWithStderrMessage(t *testing.T) {
	// given
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestHeadlessFatalHelper$")
	cmd.Env = append(os.Environ(), helperEnv+"=1", "APPFAT_HEADLESS_BODY=test")

	var out bytes.Buffer
	cmd.Stderr = &out
	cmd.Stdout = &out

	// when
	err := cmd.Run()

	// then
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("child exited with %v, want exit status %d", err, exitFatal)
	}

	if code := exitErr.ExitCode(); code != exitFatal {
		t.Fatalf("child exit code = %d, want %d (appfat.cpp:66 diablo_quit(1))", code, exitFatal)
	}

	if got := out.String(); !strings.Contains(got, "Error\ntest\n") {
		t.Fatalf("child stderr = %q, want it to contain %q", got, "Error\ntest\n")
	}
}

// TestHeadlessFatalHelper is the child half of TestExitOnError_Fatal: it is not
// a test on its own.
func TestHeadlessFatalHelper(t *testing.T) {
	// given
	if os.Getenv(helperEnv) != "1" {
		t.Skip("child process only; driven by TestExitOnError_Fatal_ExitsOneWithStderrMessage")
	}

	// when
	// The presentation and the exit code come from the real boundary.
	code := exitOnError(os.Stderr, appfat.AppFatal(os.Getenv("APPFAT_HEADLESS_BODY")))

	// then
	// The parent asserts on the exit code and stderr; the child just ends.
	os.Exit(code) //nolint:revive // child process stands in for main()
}

// Ports diablo.cpp:1089-1093 — a flag error reports on stderr and exits 64.
func TestRun_FlagError_ExitsSixtyFour(t *testing.T) {
	// given
	var stderr bytes.Buffer

	// when
	code := run([]string{"-nope"}, &stderr)

	// then
	if code != exitUsage {
		t.Fatalf("run(-nope) = %d, want %d (diablo.cpp:1093 diablo_quit(64))", code, exitUsage)
	}

	if !strings.Contains(stderr.String(), "nope") {
		t.Fatalf("run(-nope) stderr = %q, want it to name the bad flag", stderr.String())
	}
}

func TestRun_ZeroTicks_ExitsZero(t *testing.T) {
	// given
	var stderr bytes.Buffer

	// when
	code := run(nil, &stderr)

	// then
	if code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}
