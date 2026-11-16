package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/render"
	"github.com/bmcszk/devilutionGo/internal/render/parity"
	"github.com/bmcszk/devilutionGo/internal/render/paritytest"
)

const (
	// frameName is the dump frame name the parity subcommands record and compare.
	frameName = "frame"
	// frameWidth is the scripted world's surface width.
	frameWidth = render.WorldLogicalWidth
	// frameHeight is the scripted world's surface height.
	frameHeight = render.WorldLogicalHeight
)

// errParityDiff marks a completed compare with unmasked differences; main maps
// every refresher error to exit 1, so it only suppresses the redundant error
// log line.
var errParityDiff = errors.New("parity frames differ")

// runParityRecord renders the scripted parity frame through the presentation
// pass and writes the dump family into DIR (the record half of the retired
// dgo-parity CLI, devilGo-7aul).
//
// Exit-code contract change: the old CLI exited 0 exact (or allow-listed),
// 1 diff, 2 usage error; goldens exits 1 on every refresher error (main.go),
// so the diff verdict and the usage error share goldens' failure exit.
func runParityRecord(paths []string) error {
	dump, err := recordFrame()
	if err != nil {
		return err
	}

	if err := parity.WriteDump(paths[0], frameName, dump); err != nil {
		return fmt.Errorf("write dump: %w", err)
	}

	slog.Info("goldens parity-record done", "dir", paths[0])

	return nil
}

// runParityCompare adapts the refresher table shape: the verdict always
// reports on stderr.
func runParityCompare(paths []string) error {
	return compareParity(paths, os.Stderr)
}

// compareParity loads both dump families and reports the comparator verdict:
// exact, allow-listed, or the first differing pixel with its owning tile (the
// compare half of the retired dgo-parity CLI).
//
// A comparison divergence is not a tool failure: it is reported and returned
// as errParityDiff so main exits 1 without slog noise.
func compareParity(paths []string, stderr io.Writer) error {
	golden, actual, err := splitCompareTargets(paths[0])
	if err != nil {
		return err
	}

	goldenDump, err := parity.LoadFrameDump(golden, frameName, frameWidth, frameHeight)
	if err != nil {
		return fmt.Errorf("load golden: %w", err)
	}

	actualDump, err := parity.LoadFrameDump(actual, frameName, frameWidth, frameHeight)
	if err != nil {
		return fmt.Errorf("load actual: %w", err)
	}

	report := parity.Diff(goldenDump, actualDump, parity.Options{})

	if err := reportVerdict(stderr, report); err != nil {
		return err
	}

	if !report.IndexExact || len(report.ColourDiffs) != 0 {
		return errParityDiff
	}

	return nil
}

// recordFrame renders the scripted world through the presentation pass. It
// needs a display: Ebitengine hosts one run loop per process, which is exactly
// what this invocation is for.
func recordFrame() (parity.FrameDump, error) {
	if !paritytest.GPUAvailable() {
		return parity.FrameDump{}, errors.New("parity-record needs a display (DISPLAY)")
	}

	world := paritytest.World()

	dump, err := paritytest.Capture(
		paritytest.PaletteState(),
		paritytest.SoftwareSurface(world),
		paritytest.FrameAfterWarmup,
	)
	if err != nil {
		return parity.FrameDump{}, fmt.Errorf("capture: %w", err)
	}

	return dump, nil
}

// splitCompareTargets splits GOLDEN_DIR:ACTUAL_DIR into the two directories:
// exactly one colon-separated pair, both sides non-empty.
func splitCompareTargets(targets string) (golden, actual string, err error) {
	parts := strings.Split(targets, ":")
	if len(parts) != 2 {
		return "", "", errors.New("parity-compare wants GOLDEN_DIR:ACTUAL_DIR")
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("parity-compare directories must both be non-empty")
	}

	return parts[0], parts[1], nil
}

// reportVerdict prints the comparator report: exact, masked, or the first
// differing pixel with its owning tile.
func reportVerdict(w io.Writer, report parity.Report) error {
	switch {
	case report.IndexDiff != nil:
		diff := *report.IndexDiff

		return reportf(w, "DIFF index (%d,%d) tile (%d,%d): got %d want %d\n",
			diff.X, diff.Y, diff.TileX, diff.TileY, diff.Got, diff.Want)
	case len(report.ColourDiffs) != 0:
		diff := report.ColourDiffs[0]

		return reportf(w, "DIFF colour at (%d,%d) tile (%d,%d): %d unmasked colour diffs\n",
			diff.X, diff.Y, diff.TileX, diff.TileY, len(report.ColourDiffs))
	case len(report.Masked) != 0:
		return reportf(w, "PASS (allow-listed: %d masked regions)\n", len(report.Masked))
	default:
		return reportf(w, "PASS exact\n")
	}
}

// reportf writes one formatted report line, checking the write.
func reportf(w io.Writer, format string, args ...any) error {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	return nil
}
