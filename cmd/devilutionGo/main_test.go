package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
	"github.com/bmcszk/devilutionGo/internal/engine/demo"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// helperEnv makes the child half of the process-level check run.
const helperEnv = "APPFAT_HEADLESS_HELPER"

// demoSeed fixes the kernel streams a headless replay draws from: the same file
// and seed must hash identically, run after run.
const demoSeed uint32 = 1

// demoHeaderSave is the save slot stamped into the synthetic demo file.
const demoHeaderSave uint32 = 7

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
	code := run([]string{"--headless"}, &stderr)

	// then
	if code != 0 {
		t.Fatalf("run(--headless) = %d, want 0", code)
	}
}

// TestRun_HelpShowsHeadlessSurface pins ADR-020: the one game binary advertises
// the headless mode on its own command line.
func TestRun_HelpShowsHeadlessSurface(t *testing.T) {
	// given
	var stderr bytes.Buffer

	// when
	code := run([]string{"--help"}, &stderr)

	// then
	if code != 0 {
		t.Fatalf("run(--help) = %d, want 0", code)
	}

	for _, want := range []string{"-headless", "-ticks"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("run(--help) output = %q, want it to show %q", stderr.String(), want)
		}
	}
}

// TestReplay_DemoFile_HashesIdentically proves the headless truth: a demo file
// replays end to end through the deterministic kernel with no window, and two
// runs of the same stream and seed produce the same event trace.
func TestReplay_DemoFile_HashesIdentically(t *testing.T) {
	// given
	path := filepath.Join(t.TempDir(), "replay.dmo")
	writeDemoFile(t, path)

	// when
	firstHash, firstEvents := replayDemoFile(t, path)
	secondHash, secondEvents := replayDemoFile(t, path)

	// then
	if firstEvents == 0 {
		t.Fatal("replay emitted no events, want the recorded inputs replayed")
	}

	if firstHash != secondHash || firstEvents != secondEvents {
		t.Fatalf("replays hashed %d (%d events) then %d (%d events), want identical traces",
			firstHash, firstEvents, secondHash, secondEvents)
	}
}

// TestHeadlessMode_HasNoDisplayDependency proves the command the headless mode
// runs in links no display stack: no ebiten, no render package. Both run modes
// share one command since ADR-020, so the windowed bring-up has to re-scope this
// assertion to the headless run path.
func TestHeadlessMode_HasNoDisplayDependency(t *testing.T) {
	// given
	cmd := exec.CommandContext(context.Background(), "go", "list", "-deps", "-f", "{{.ImportPath}}", ".")

	// when
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}

	// then
	for _, banned := range []string{"hajimehoshi/ebiten", "/render"} {
		if strings.Contains(string(out), banned) {
			t.Fatalf("headless dependency graph contains %q, want display-free", banned)
		}
	}
}

// writeDemoFile records one synthetic input stream to path, byte for byte the
// stream the replay test then feeds back through the kernel.
func writeDemoFile(t *testing.T, path string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create demo file: %v", err)
	}
	defer func() { _ = file.Close() }()

	if err := demo.WriteHeader(file, demoHeaderSave); err != nil {
		t.Fatalf("write demo header: %v", err)
	}

	writer := demo.NewWriter(file)

	for _, rec := range []demo.Record{
		demo.MouseMotionRecord{X: 10, Y: 20},
		demo.MouseButtonRecord{Button: 1, X: 10, Y: 20, Down: true},
		demo.TickRecord{Progress: 40},
		demo.MouseWheelRecord{X: -1, Y: 2, Mod: 4},
		demo.KeyRecord{Sym: 27, Down: true},
		demo.KeyRecord{Sym: 27},
		demo.TickRecord{Progress: 80},
		demo.QuitRecord{},
	} {
		if err := writer.WriteRecord(rec); err != nil {
			t.Fatalf("write demo record %T: %v", rec, err)
		}
	}

	if err := writer.Flush(); err != nil {
		t.Fatalf("flush demo file: %v", err)
	}
}

// replayDemoFile runs the recorded file through the kernel replay seam and
// returns its trace hash plus the event count of that run.
func replayDemoFile(t *testing.T, path string) (uint64, int) {
	t.Helper()

	stream, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read demo file: %v", err)
	}

	k := kernel.New(kernel.DropUnnamed(), kernel.NewStreams(demoSeed, "ai", "items"))
	for _, kind := range []kernel.Kind{
		kernel.KindMouseMotion,
		kernel.KindMouseButton,
		kernel.KindMouseWheel,
		kernel.KindKey,
		kernel.KindQuit,
	} {
		k.Register(kind, func(c kernel.Command, streams *kernel.Streams) []kernel.Event {
			roll, _ := streams.Draw("ai")

			return []kernel.Event{{Kind: kernel.KindPointerMoved, A: c.A, B: c.B, C: c.C, Stream: "ai", Roll: roll}}
		})
	}

	hash, err := game.NewDemoReplay(k).Replay(bytes.NewReader(stream))
	if err != nil {
		t.Fatalf("replay demo file: %v", err)
	}

	return hash, len(k.Trace())
}
