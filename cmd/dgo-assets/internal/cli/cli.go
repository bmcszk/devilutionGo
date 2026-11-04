// Package cli implements the dgo-assets command: manifest coverage checking,
// listing and legacy→CLX conversion.
//
// It is a cmd-layer package: it may import internal/assets and
// internal/assets/manifest, and nothing Diablo-specific.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/assets/manifest"
)

// Process exit codes shared by every subcommand.
const (
	ExitOK      = 0
	ExitFailure = 1
	ExitUsage   = 2
)

const usage = `dgo-assets: graphics asset inventory tool

usage:
  dgo-assets check   --manifest <yaml> --root <dir>
  dgo-assets list    --manifest <yaml> [--markdown]
  dgo-assets convert --format cel|cl2|pcx --in <file> --out <file>
                     [--widths 1,2,3] [--frames N] [--transparent N]

check exits non-zero and prints the loader file:line for every entry whose
path does not resolve under --root.
`

const convertUsage = `convert flags: --format cel|cl2|pcx --in <file> --out <file>
  --widths       frame widths, comma separated (cel, cl2)
  --frames       pcx frame count (positive) or frame height (negative)
  --transparent  pcx transparent palette index
`

// convertRequest is one validated convert invocation.
type convertRequest struct {
	format      string
	inPath      string
	outPath     string
	widths      []int
	frames      int
	transparent int
}

// Run dispatches one dgo-assets invocation and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stderr, usage)

		return ExitUsage
	}

	switch args[0] {
	case "check":
		return runCheck(args[1:], stdout, stderr)
	case "list":
		return runList(args[1:], stdout, stderr)
	case "convert":
		return runConvert(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		_, _ = fmt.Fprint(stdout, usage)

		return ExitOK
	default:
		_, _ = fmt.Fprintf(stderr, "dgo-assets: unknown command %q\n\n%s", args[0], usage)

		return ExitUsage
	}
}

// loadManifest reads and validates a manifest file.
func loadManifest(path string) (manifest.Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("open manifest %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	m, err := manifest.Read(file)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("read manifest %s: %w", path, err)
	}
	if err := m.Validate(); err != nil {
		return manifest.Manifest{}, fmt.Errorf("invalid manifest %s: %w", path, err)
	}

	return m, nil
}

// runCheck reports every manifest entry that does not resolve under root.
func runCheck(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "assets/manifest.yaml", "manifest YAML to check")
	root := flags.String("root", "assets/loose", "asset root the entry paths resolve under")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}

	m, err := loadManifest(*manifestPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "dgo-assets check: %v\n", err)

		return ExitFailure
	}

	missing := m.Missing(*root)
	for _, entry := range missing {
		_, _ = fmt.Fprintf(stdout, "missing %s (requested by %s)\n", entry.Path, entry.Loader)
	}
	if len(missing) > 0 {
		_, _ = fmt.Fprintf(stderr, "dgo-assets check: %d of %d entries do not resolve under %s\n",
			len(missing), len(m.Entries), *root)

		return ExitFailure
	}

	_, _ = fmt.Fprintf(stdout, "dgo-assets check: %d entries in %d families resolve under %s\n",
		len(m.Entries), m.Summary.Families, *root)

	return ExitOK
}

// runList prints the manifest entries, one per line or as a markdown table.
func runList(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "assets/manifest.yaml", "manifest YAML to list")
	markdown := flags.Bool("markdown", false, "print a markdown table instead of tab-separated lines")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}

	m, err := loadManifest(*manifestPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "dgo-assets list: %v\n", err)

		return ExitFailure
	}

	if *markdown {
		_, _ = fmt.Fprintln(stdout, "| family | path | format | frames | loader |")
		_, _ = fmt.Fprintln(stdout, "| --- | --- | --- | --- | --- |")
	}
	for _, entry := range m.Entries {
		if *markdown {
			_, _ = fmt.Fprintf(stdout, "| %s | %s | %s | %d | %s |\n",
				entry.Family, entry.Path, entry.Format, entry.Frames, entry.Loader)

			continue
		}
		_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\t%d\t%s\n",
			entry.Family, entry.Path, entry.Format, entry.Frames, entry.Loader)
	}

	return ExitOK
}

// runConvert converts one legacy file into CLX.
func runConvert(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("convert", flag.ContinueOnError)
	flags.SetOutput(stderr)
	format := flags.String("format", "", "source format: cel, cl2 or pcx")
	inPath := flags.String("in", "", "source file")
	outPath := flags.String("out", "", "CLX output file")
	widths := flags.String("widths", "", "frame widths, comma separated (cel, cl2)")
	frames := flags.Int("frames", 1, "pcx frame count (positive) or frame height (negative)")
	transparent := flags.Int("transparent", -1, "pcx transparent palette index")
	if err := flags.Parse(args); err != nil {
		return ExitUsage
	}
	if *format == "" || *inPath == "" || *outPath == "" {
		_, _ = fmt.Fprintf(stderr, "dgo-assets convert: --format, --in and --out are required\n%s", convertUsage)

		return ExitUsage
	}

	if !validFormat(*format) {
		_, _ = fmt.Fprintf(stderr, "dgo-assets convert: unknown format %q (want cel, cl2 or pcx)\n", *format)

		return ExitUsage
	}

	request, err := newConvertRequest(*format, *inPath, *outPath, *widths, *frames, *transparent)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "dgo-assets convert: %v\n", err)

		return ExitUsage
	}

	written, err := request.convert()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "dgo-assets convert: %v\n", err)

		return ExitFailure
	}

	_, _ = fmt.Fprintf(stdout, "dgo-assets convert: wrote %d bytes to %s\n", written, *outPath)

	return ExitOK
}

// newConvertRequest validates the convert arguments before any file is read.
func newConvertRequest(format, inPath, outPath, widths string, frames, transparent int) (convertRequest, error) {
	if !validFormat(format) {
		return convertRequest{}, fmt.Errorf("unknown format %q (want cel, cl2 or pcx)", format)
	}

	frameWidths, err := parseInts(widths)
	if err != nil {
		return convertRequest{}, err
	}

	return convertRequest{
		format: format, inPath: inPath, outPath: outPath,
		widths: frameWidths, frames: frames, transparent: transparent,
	}, nil
}

// convert reads the source, converts it and writes the CLX output.
func (r convertRequest) convert() (int, error) {
	source, err := os.ReadFile(r.inPath)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", r.inPath, err)
	}

	converted, err := r.convertSource(source)
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(r.outPath, converted, 0o644); err != nil {
		return 0, fmt.Errorf("write %s: %w", r.outPath, err)
	}

	return len(converted), nil
}

// convertSource dispatches to the converter for the request's format.
func (r convertRequest) convertSource(source []byte) ([]byte, error) {
	switch r.format {
	case "cel":
		return convertCEL(source, r.widths)
	case "cl2":
		return convertCL2(source, r.widths)
	default:
		return convertPCX(source, r.frames, r.transparent)
	}
}

// convertCEL converts a CEL file to CLX.
func convertCEL(source []byte, widths []int) ([]byte, error) {
	converted, err := assets.ConvertCELToCLX(source, widths)
	if err != nil {
		return nil, fmt.Errorf("convert CEL: %w", err)
	}

	return converted, nil
}

// convertCL2 converts a CL2 file to CLX.
func convertCL2(source []byte, widths []int) ([]byte, error) {
	converted, err := assets.ConvertCL2ToCLX(source, widths)
	if err != nil {
		return nil, fmt.Errorf("convert CL2: %w", err)
	}

	return converted, nil
}

// validFormat reports whether format is a supported source container.
func validFormat(format string) bool {
	switch format {
	case "cel", "cl2", "pcx":
		return true
	default:
		return false
	}
}

// convertPCX converts a PCX with an optional transparent palette index.
func convertPCX(source []byte, frames, transparent int) ([]byte, error) {
	var transparentColor []byte
	if transparent >= 0 {
		transparentColor = []byte{byte(transparent)}
	}

	converted, err := assets.ConvertPCXToCLX(source, frames, transparentColor...)
	if err != nil {
		return nil, fmt.Errorf("convert PCX: %w", err)
	}

	return converted, nil
}

// parseInts parses a comma-separated integer list; empty means none.
func parseInts(list string) ([]int, error) {
	if list == "" {
		return nil, nil
	}

	fields := strings.Split(list, ",")
	values := make([]int, len(fields))
	for i, field := range fields {
		value, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil {
			return nil, fmt.Errorf("width %q is not an integer: %w", field, err)
		}
		values[i] = value
	}

	return values, nil
}
