// Command devilutionGo-headless is the headless entry point for CI and parity
// testing.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
	"github.com/bmcszk/devilutionGo/internal/game"
)

const (
	tickDuration = 50 * time.Millisecond

	// exitFatal is the code DisplayFatalErrorAndExit reaches via diablo_quit(1)
	// (appfat.cpp:66).
	exitFatal = 1
	// exitUsage is the code upstream quits with on a flag error (diablo.cpp:1093).
	exitUsage = 64
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run drives the headless tick loop and returns the process exit code.
func run(args []string, stderr io.Writer) int {
	ticks, err := parseTicks(args, stderr)
	if err != nil {
		return exitOnError(stderr, err)
	}

	src := &engine.MockTickSource{}

	g := game.NewGame(src)
	for range ticks {
		src.NowValue += tickDuration
		g.Update(tickDuration)
	}

	slog.Info("headless run complete", "ticks", ticks)

	return 0
}

// parseTicks parses the headless flags, reporting parse failures on stderr.
func parseTicks(args []string, stderr io.Writer) (int, error) {
	flags := flag.NewFlagSet("devilutionGo-headless", flag.ContinueOnError)
	flags.SetOutput(stderr)
	ticks := flags.Int("ticks", 0, "run this many 50ms game ticks, then exit 0")

	if err := flags.Parse(args); err != nil {
		return 0, fmt.Errorf("headless: parse flags: %w", err)
	}

	return *ticks, nil
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
