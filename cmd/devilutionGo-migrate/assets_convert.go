package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// convertRequest is one validated convert invocation.
type convertRequest struct {
	format      string
	inPath      string
	outPath     string
	widths      []int
	frames      int
	transparent int
}

// runAssetsConvert converts one legacy container into CLX and reports the
// written byte count.
func runAssetsConvert(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("assets convert", flag.ContinueOnError)
	flags.SetOutput(stderr)
	format := flags.String("format", "", "source format: cel, cl2 or pcx")
	inPath := flags.String("in", "", "source file")
	outPath := flags.String("out", "", "CLX output file")
	widths := flags.String("widths", "", "frame widths, comma separated (cel, cl2)")
	frames := flags.Int("frames", 1, "pcx frame count (positive) or frame height (negative)")
	transparent := flags.Int("transparent", -1, "pcx transparent palette index")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}

	if *format == "" || *inPath == "" || *outPath == "" {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate assets convert: --format, --in and --out are required\n%s",
			assetConvertUsage)

		return exitUsage
	}

	request, err := newConvertRequest(*format, *inPath, *outPath, *widths, *frames, *transparent)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate assets convert: %v\n", err)

		return exitUsage
	}

	written, err := request.convert()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate assets convert: %v\n", err)

		return exitFailure
	}

	_, _ = fmt.Fprintf(stdout, "devilutionGo-migrate assets convert: wrote %d bytes to %s\n", written, *outPath)

	return exitOK
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

// convertSource dispatches to the converter for the request's format. CEL and
// CL2 go through the same geometry dispatch the inline assets pass uses, so the
// folded path and `assets convert` cannot drift (ADR-019).
func (r convertRequest) convertSource(source []byte) ([]byte, error) {
	if r.format == "pcx" {
		return convertPCX(source, r.frames, r.transparent)
	}

	return spriteGeometry{widths: r.widths, frames: r.frames}.convert(r.format, source)
}
