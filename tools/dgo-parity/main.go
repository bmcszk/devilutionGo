// Command dgo-parity records and compares full-frame palette-index dumps
// (devilGo-7aul): record renders scripted parity frames into a dump directory,
// compare runs the byte-exact comparator over two dump directories.
//
//	go run ./tools/dgo-parity -record /tmp/parity-dumps
//	go run ./tools/dgo-parity -compare /tmp/parity-dumps-golden:/tmp/parity-dumps
//
// Exit codes: 0 exact (or allow-listed), 1 diff, 2 usage error.
//
// It never ships: the parity harness is a developer instrument under tools/
// per ADR-020 (cmd/ holds exactly the two user-facing commands).
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/bmcszk/devilutionGo/internal/render"
	"github.com/bmcszk/devilutionGo/internal/render/parity"
	"github.com/bmcszk/devilutionGo/internal/render/paritytest"
)

const (
	// exitExact is the exit code of an exact or allow-listed compare.
	exitExact = 0
	// exitDiff is the exit code of a compare with unmasked differences.
	exitDiff = 1
	// exitUsage is the exit code of a bad invocation.
	exitUsage = 2
	// frameName is the dump frame name the tool records and compares.
	frameName = "frame"
	// frameWidth is the scripted world's surface width.
	frameWidth = render.WorldLogicalWidth
	// frameHeight is the scripted world's surface height.
	frameHeight = render.WorldLogicalHeight
	// usageLine is the one-line invocation summary a bad flag combination gets.
	usageLine = "usage: dgo-parity -record DIR | -compare GOLDEN_DIR:ACTUAL_DIR"
)

// options are the parsed command-line flags: exactly one of the two modes.
type options struct {
	record  string
	compare [2]string
}

// main dispatches one mode and maps the verdict to the exit code.
func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run parses the flags, runs the mode and returns the process exit code.
func run(args []string, stderr io.Writer) int {
	opts, err := parseFlags(args)
	if err != nil {
		reportUsage(stderr, err)

		return exitUsage
	}

	if opts.record != "" {
		return runRecord(opts.record, stderr)
	}

	return runCompare(opts.compare, stderr)
}

// parseFlags resolves -record or -compare; both set or both unset is a usage
// error.
func parseFlags(args []string) (options, error) {
	opts := options{}
	flagSet := flag.NewFlagSet("dgo-parity", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	compareTargets := ""
	flagSet.StringVar(&opts.record, "record", "", "render the scripted parity frame into DIR")
	flagSet.StringVar(&compareTargets, "compare", "", "compare GOLDEN_DIR against ACTUAL_DIR (colon-separated)")

	if err := flagSet.Parse(args); err != nil {
		return options{}, fmt.Errorf("dgo-parity: %w", err)
	}

	if err := resolveMode(&opts, flagSet.Args(), opts.record, compareTargets); err != nil {
		return options{}, fmt.Errorf("%w\n%s", err, usageLine)
	}

	return opts, nil
}

// resolveMode validates the flag combination: exactly one mode with the right
// operand shape.
func resolveMode(opts *options, positional []string, record, compareTargets string) error {
	switch {
	case record != "" && compareTargets != "":
		return errors.New("dgo-parity: -record and -compare are exclusive")
	case record != "":
		if len(positional) != 0 {
			return errors.New("dgo-parity: -record takes no positional paths")
		}

		return nil
	case compareTargets != "":
		golden, actual, err := splitCompareTargets(compareTargets)
		if err != nil {
			return err
		}

		opts.compare = [2]string{golden, actual}

		return nil
	default:
		return errors.New("dgo-parity: one of -record or -compare is required")
	}
}

// splitCompareTargets splits GOLDEN:ACTUAL into the two directories.
func splitCompareTargets(targets string) (golden, actual string, err error) {
	golden, actual, found := cutUnique(targets, ":")
	if !found {
		return "", "", errors.New("dgo-parity: -compare wants GOLDEN_DIR:ACTUAL_DIR")
	}

	if golden == "" || actual == "" {
		return "", "", errors.New("dgo-parity: -compare directories must both be non-empty")
	}

	return golden, actual, nil
}

// cutUnique splits s at its only separator; zero or several separators are
// signalled by found=false.
func cutUnique(s, sep string) (before, after string, found bool) {
	first := indexOf(s, sep)
	if first < 0 {
		return s, "", false
	}

	if indexOf(s[first+len(sep):], sep) >= 0 {
		return "", "", false
	}

	return s[:first], s[first+len(sep):], true
}

// indexOf returns the offset of sep inside s, or -1.
func indexOf(s, sep string) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}

	return -1
}

// runRecord renders the scripted parity frame and writes the dump family.
func runRecord(dir string, stderr io.Writer) int {
	dump, err := recordFrame()
	if err != nil {
		slog.Error("dgo-parity: record failed", "err", err)

		return exitUsage
	}

	if err := parity.WriteDump(dir, frameName, dump); err != nil {
		slog.Error("dgo-parity: record failed", "err", err)

		return exitUsage
	}

	if err := reportf(stderr, "RECORDED %s\n", dir); err != nil {
		slog.Error("dgo-parity: report failed", "err", err)
	}

	return exitExact
}

// runCompare loads both dumps and reports the comparator verdict.
func runCompare(dirs [2]string, stderr io.Writer) int {
	golden, err := parity.LoadFrameDump(dirs[0], frameName, frameWidth, frameHeight)
	if err != nil {
		slog.Error("dgo-parity: compare failed", "err", err)

		return exitUsage
	}

	actual, err := parity.LoadFrameDump(dirs[1], frameName, frameWidth, frameHeight)
	if err != nil {
		slog.Error("dgo-parity: compare failed", "err", err)

		return exitUsage
	}

	report := parity.Diff(golden, actual, parity.Options{})

	if err := reportVerdict(stderr, report); err != nil {
		slog.Error("dgo-parity: report failed", "err", err)
	}

	if !report.IndexExact || len(report.ColourDiffs) != 0 {
		return exitDiff
	}

	return exitExact
}

// recordFrame renders the scripted world through the presentation pass. It
// needs a display: Ebitengine hosts one run loop per process, which is exactly
// what this CLI process is for.
func recordFrame() (parity.FrameDump, error) {
	if !paritytest.GPUAvailable() {
		return parity.FrameDump{}, errors.New("dgo-parity: -record needs a display (DISPLAY)")
	}

	world := paritytest.World()

	dump, err := paritytest.Capture(
		paritytest.PaletteState(),
		paritytest.SoftwareSurface(world),
		paritytest.FrameAfterWarmup)
	if err != nil {
		return parity.FrameDump{}, fmt.Errorf("dgo-parity: capture: %w", err)
	}

	return dump, nil
}

// reportVerdict prints the comparator report: exact, masked, or the first
// differing pixel with its owning tile.
func reportVerdict(stderr io.Writer, report parity.Report) error {
	switch {
	case report.IndexDiff != nil:
		diff := *report.IndexDiff

		return reportf(stderr, "DIFF index (%d,%d) tile (%d,%d): got %d want %d\n",
			diff.X, diff.Y, diff.TileX, diff.TileY, diff.Got, diff.Want)
	case len(report.ColourDiffs) != 0:
		diff := report.ColourDiffs[0]

		return reportf(stderr, "DIFF colour at (%d,%d) tile (%d,%d): %d unmasked colour diffs\n",
			diff.X, diff.Y, diff.TileX, diff.TileY, len(report.ColourDiffs))
	case len(report.Masked) != 0:
		return reportf(stderr, "PASS (allow-listed: %d masked regions)\n", len(report.Masked))
	default:
		return reportf(stderr, "PASS exact\n")
	}
}

// reportUsage presents a usage error on stderr.
func reportUsage(stderr io.Writer, err error) {
	if reportErr := reportf(stderr, "%s\n%s\n", err, usageLine); reportErr != nil {
		slog.Error("dgo-parity: report failed", "err", reportErr)
	}
}

// reportf writes one formatted report line, checking the write.
func reportf(w io.Writer, format string, args ...any) error {
	if _, err := fmt.Fprintf(w, format, args...); err != nil {
		return fmt.Errorf("dgo-parity: write report: %w", err)
	}

	return nil
}
