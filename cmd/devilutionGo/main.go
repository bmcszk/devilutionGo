// Command devilutionGo is the game: it runs windowed, or headless when the
// command line asks for it (ADR-020).
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
	"github.com/bmcszk/devilutionGo/internal/game"
)

const (
	// exitFatal is the code DisplayFatalErrorAndExit reaches via diablo_quit(1)
	// (appfat.cpp:66).
	exitFatal = 1
	// exitUsage is the code upstream quits with on a flag error (diablo.cpp:1093).
	exitUsage = 64
)

// runOptions is what the shared command line selects: the windowed game by
// default, the headless tick loop on request.
type runOptions struct {
	headless bool
	ticks    int
}

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run selects the run mode from the flags and returns the process exit code.
func run(args []string, stderr io.Writer) int {
	opts, err := parseFlags(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}

	if err != nil {
		return exitOnError(stderr, err)
	}

	if opts.headless {
		return runHeadless(opts.ticks)
	}

	runWindowed(args)

	return 0
}

// parseFlags reads the run-mode flags, reporting parse failures on stderr.
func parseFlags(args []string, stderr io.Writer) (runOptions, error) {
	flags := flag.NewFlagSet("devilutionGo", flag.ContinueOnError)
	flags.SetOutput(stderr)
	opts := runOptions{}

	flags.BoolVar(&opts.headless, "headless", false, "run without a window, then exit")
	flags.IntVar(&opts.ticks, "ticks", 0, "with -headless, run this many 50ms game ticks, then exit 0")

	if err := flags.Parse(args); err != nil {
		return runOptions{}, fmt.Errorf("parse flags: %w", err)
	}

	return opts, nil
}

// runWindowed starts the game with the debug surface the args ask for.
func runWindowed(args []string) {
	// Upstream only ships the cheat commands in _DEBUG builds; here the same
	// surface is an explicit opt-in switch (--debug).
	g := game.NewGame(&engine.MockTickSource{})
	g.Debug = game.NewDebugRouterFromFlag(args, game.DebugEnv{})

	slog.Info("devilutionGo started", "debug", g.Debug != nil)
}

// exitOnError presents err on stderr and returns the process exit code: a fatal
// error uses the dialog message and code 1 (appfat.cpp:62-67); anything else is
// a usage error (diablo.cpp:1093).
func exitOnError(stderr io.Writer, err error) int {
	var fatal *appfat.FatalError
	if errors.As(err, &fatal) {
		if _, writeErr := fatal.WriteTo(stderr); writeErr != nil {
			slog.Error("headless: report fatal error", "err", writeErr)
		}

		return exitFatal
	}

	slog.Error("headless: usage error", "err", err)

	return exitUsage
}
