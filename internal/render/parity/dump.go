// Package parity compares full-frame palette-index dumps: the loader for the
// C++ backbuffer dumper's <frame>.idx.bin / <frame>.pal.bin / <frame>.ppm
// families, the byte-exact comparator and the golden codec (devilGo-7aul).
//
// Dump contract (dx.cpp GlobalBackBuffer + capture.cpp flow, devilutionX
// e1ad0c2b6): the index bytes are the 8-bit PalSurface contents, row pitch
// equals Width, and the palette is the active 256-entry row the frame was
// presented with. The PPM is the colour reference (P6 binary, maxval 255) whose
// raster is the palette expansion of the index bytes.
package parity

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const (
	// TileSize is the 32px sprite grid a differing pixel is attributed to
	// (DunFrameWidth = TILE_WIDTH/2, Source/levels/dun_tile.hpp:124).
	TileSize = 32
	// paletteEntries is the number of entries in a Diablo palette (palette.h:23).
	paletteEntries = 256
	// paletteEntryBytes is the byte count of one RGBA palette record.
	paletteEntryBytes = 4
	// ppmMagic is the binary PPM magic the colour reference carries.
	ppmMagic = "P6"
	// ppmMaxval is the only maxval the colour reference may carry.
	ppmMaxval = 255
)

// RGBA is one palette record: the C++ dumper's SDL_Color layout (R, G, B, unused).
type RGBA struct {
	R, G, B, A uint8
}

// Palette is the active 256-entry palette row of one dumped frame.
type Palette [paletteEntries]RGBA

// FrameDump is one captured frame: the index bytes, the active palette and the
// colour reference. LightLevels carries the per-pixel light levels when the
// producer recorded them; they are diagnostic and are not persisted by WriteDump.
type FrameDump struct {
	Width  int
	Height int
	// Index is Width*Height palette indices, row-major, pitch Width.
	Index []byte
	// Palette is the 256-entry palette row the frame was presented with.
	Palette Palette
	// LightLevels is Width*Height light levels, or nil when not recorded.
	LightLevels []byte
}

// Region is a rectangle of pixels an allow-list entry masks out of the colour
// gate: X..X+Width and Y..Y+Height in pixel coordinates.
type Region struct {
	X, Y, Width, Height int
}

// PixelDiff is one mismatching pixel: its position, the owning TileSize tile,
// the index bytes when the index gate caught it and the expanded RGB when the
// colour gate caught it.
type PixelDiff struct {
	X, Y         int
	TileX, TileY int
	// Got and Want are the two index bytes (index-gate diffs only).
	Got, Want uint8
	// GotRGB and WantRGB are the two palette-expanded colours (colour-gate diffs).
	GotRGB, WantRGB [3]uint8
}

// Report is one comparator run: the index gate is byte-exact and never
// relaxable, the colour gate reports every unmasked palette-expanded mismatch.
type Report struct {
	// IndexExact holds when every index byte matched.
	IndexExact bool
	// IndexDiff is the first differing index byte, nil when exact.
	IndexDiff *PixelDiff
	// ColourDiffs lists every unmasked palette-expanded mismatch in scan order.
	ColourDiffs []PixelDiff
	// Masked echoes the allow-list regions that covered at least one pixel.
	Masked []Region
}

// Options carries the comparator's data-encoded inputs.
type Options struct {
	// AllowList masks regions out of the colour gate only; the index gate is
	// never masked.
	AllowList []Region
}

// frameFiles bundles the three file bodies of one frame family.
type frameFiles struct {
	// index is the raw <frame>.idx.bin body.
	index []byte
	// palette is the raw <frame>.pal.bin body.
	palette []byte
	// ppm is the raw <frame>.ppm body.
	ppm []byte
}

// decodedFrame bundles a validated dump with its PPM raster offset, so pixel
// violations can cite whole-file offsets.
type decodedFrame struct {
	dump FrameDump
	// rasterOffset is the raster start inside the PPM file body.
	rasterOffset int
}

// ppmHeader is the parsed header of one P6 frame.
type ppmHeader struct {
	width  int
	height int
	maxval int
	// offsets are the start offsets of the width, height and maxval fields.
	offsets [3]int
	// rasterOffset is the offset just past the maxval field.
	rasterOffset int
}

// LoadFrameDump reads one frame family from dir: <frame>.idx.bin, <frame>.pal.bin
// and <frame>.ppm, validated strictly against width x height. The returned error
// names the first violating file and byte offset.
func LoadFrameDump(dir, frame string, width, height int) (FrameDump, error) {
	files, err := readFrameFiles(dir, frame)
	if err != nil {
		return FrameDump{}, err
	}

	decoded, err := decodeFrameDump(files, width, height)
	if err != nil {
		return FrameDump{}, err
	}

	return decoded.dump, nil
}

// WriteDump writes one frame family into dir with the byte format the loader
// reads: index bytes, the 256x4 palette rows and the P6 colour reference whose
// raster is the palette expansion. LightLevels are not persisted.
func WriteDump(dir, frame string, dump FrameDump) error {
	if dump.Width <= 0 || dump.Height <= 0 {
		return fmt.Errorf("parity: dump: dimensions %dx%d are not positive", dump.Width, dump.Height)
	}

	files, err := encodeFrameDump(dump)
	if err != nil {
		return err
	}

	byName := map[string][]byte{
		frame + ".idx.bin": files.index,
		frame + ".pal.bin": files.palette,
		frame + ".ppm":     files.ppm,
	}

	var writeErrs []error
	for name, body := range byName {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o600); err != nil {
			writeErrs = append(writeErrs, fmt.Errorf("parity: write %s: %w", name, err))
		}
	}

	return errors.Join(writeErrs...)
}

// readFrameFiles reads the three bodies of one frame family.
func readFrameFiles(dir, frame string) (files frameFiles, err error) {
	for _, entry := range []struct {
		name string
		slot *[]byte
	}{
		{frame + ".idx.bin", &files.index},
		{frame + ".pal.bin", &files.palette},
		{frame + ".ppm", &files.ppm},
	} {
		body, readErr := os.ReadFile(filepath.Join(dir, entry.name))
		if readErr != nil {
			return frameFiles{}, fmt.Errorf("parity: %w", readErr)
		}

		*entry.slot = body
	}

	return files, nil
}

// violationf builds a first-violation error naming the file and the offending
// byte offset.
func violationf(file, format string, args ...any) error {
	return fmt.Errorf("parity: %s: %s", file, fmt.Sprintf(format, args...))
}

// decodeFrameDump validates the three file bodies against width x height and
// assembles the dump. Size violations cite the first offending byte offset.
func decodeFrameDump(files frameFiles, width, height int) (decodedFrame, error) {
	indexWant := width * height
	if len(files.index) != indexWant {
		return decodedFrame{}, violationf("index", "index buffer: have %d bytes, want %d (first violation at offset %d)",
			len(files.index), indexWant, len(files.index))
	}

	paletteWant := paletteEntries * paletteEntryBytes
	if len(files.palette) != paletteWant {
		return decodedFrame{}, violationf("palette", "palette: have %d bytes, want %d (first violation at offset %d)",
			len(files.palette), paletteWant, len(files.palette))
	}

	rasterOffset, err := expectPPMFrame(files.ppm, width, height)
	if err != nil {
		return decodedFrame{}, err
	}

	dump := FrameDump{Width: width, Height: height, Index: files.index, Palette: readPalette(files.palette)}

	if err := expectPPMRasterMatchesPalette(files.ppm[rasterOffset:], &dump, rasterOffset); err != nil {
		return decodedFrame{}, err
	}

	return decodedFrame{dump: dump, rasterOffset: rasterOffset}, nil
}

// encodeFrameDump renders the dump into its three file bodies, validating the
// buffer shapes first.
func encodeFrameDump(dump FrameDump) (frameFiles, error) {
	indexWant := dump.Width * dump.Height
	if len(dump.Index) != indexWant {
		return frameFiles{}, violationf("dump", "index buffer: have %d bytes, want %d (first violation at offset %d)",
			len(dump.Index), indexWant, len(dump.Index))
	}

	if dump.LightLevels != nil && len(dump.LightLevels) != indexWant {
		return frameFiles{}, violationf("dump", "light levels: have %d bytes, want %d (first violation at offset %d)",
			len(dump.LightLevels), indexWant, len(dump.LightLevels))
	}

	palette := make([]byte, 0, paletteEntries*paletteEntryBytes)
	for _, entry := range dump.Palette {
		palette = append(palette, entry.R, entry.G, entry.B, entry.A)
	}

	raster := make([]byte, 0, 3*indexWant)
	for i := range indexWant {
		colour := dump.expandAt(i)
		raster = append(raster, colour[0], colour[1], colour[2])
	}

	ppm := appendPPMHeader(make([]byte, 0, ppmHeaderLen(dump.Width, dump.Height)+len(raster)), dump.Width, dump.Height)
	ppm = append(ppm, raster...)

	return frameFiles{index: dump.Index, palette: palette, ppm: ppm}, nil
}

// readPalette lays the flat 256x4 records into the palette row.
func readPalette(bytes []byte) Palette {
	var palette Palette
	for entry := range paletteEntries {
		offset := entry * paletteEntryBytes
		palette[entry] = RGBA{
			R: bytes[offset], G: bytes[offset+1],
			B: bytes[offset+2], A: bytes[offset+3],
		}
	}

	return palette
}

// expandAt returns the palette-expanded RGB of the pixel at linear position i.
func (d FrameDump) expandAt(i int) [3]uint8 {
	entry := d.Palette[d.Index[i]]

	return [3]uint8{entry.R, entry.G, entry.B}
}

// expectPPMFrame validates the PPM header against the wanted frame: magic,
// dimensions in scan order, maxval. It returns the raster start offset.
func expectPPMFrame(data []byte, width, height int) (int, error) {
	if len(data) < len(ppmMagic) || string(data[:len(ppmMagic)]) != ppmMagic {
		return 0, violationf("ppm", "magic: have %q, want %q (first violation at offset 0)", magicPrefix(data), ppmMagic)
	}

	header, err := parsePPMHeader(data, len(ppmMagic))
	if err != nil {
		return 0, err
	}

	if header.width != width {
		return 0, violationf("ppm", "width: have %d, want %d (first violation at offset %d)",
			header.width, width, header.offsets[0])
	}

	if header.height != height {
		return 0, violationf("ppm", "height: have %d, want %d (first violation at offset %d)",
			header.height, height, header.offsets[1])
	}

	if header.maxval != ppmMaxval {
		return 0, violationf("ppm", "maxval: have %d, want %d (first violation at offset %d)",
			header.maxval, ppmMaxval, header.offsets[2])
	}

	return header.rasterOffset, nil
}

// parsePPMHeader scans the three decimal header fields after the magic.
func parsePPMHeader(data []byte, start int) (ppmHeader, error) {
	var header ppmHeader

	pos := start
	for field := range 3 {
		tokenStart, tokenEnd := scanToken(data, pos)
		if tokenStart == tokenEnd {
			return ppmHeader{}, violationf("ppm", "header field %d: ran out of header at offset %d", field, pos)
		}

		value, convErr := strconv.Atoi(string(data[tokenStart:tokenEnd]))
		if convErr != nil {
			return ppmHeader{}, violationf("ppm", "header field %d: %q is not decimal (first violation at offset %d)",
				field, string(data[tokenStart:tokenEnd]), tokenStart)
		}

		switch field {
		case 0:
			header.width = value
		case 1:
			header.height = value
		default:
			header.maxval = value
		}

		header.offsets[field] = tokenStart
		pos = tokenEnd
	}

	header.rasterOffset = pos + 1

	return header, nil
}

// scanToken skips whitespace and returns the span of the following decimal
// token; the span is empty when the data ends first.
func scanToken(data []byte, start int) (tokenStart, tokenEnd int) {
	pos := start
	for pos < len(data) && isSpace(data[pos]) {
		pos++
	}

	tokenStart = pos
	for pos < len(data) && !isSpace(data[pos]) {
		pos++
	}

	return tokenStart, pos
}

// isSpace reports whether the byte is one of the whitespace separators the PPM
// header allows.
func isSpace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

// expectPPMRasterMatchesPalette compares the raster byte-for-byte with the
// palette expansion of the dump's index bytes; rasterOffset is the raster's
// start inside the PPM file body, for whole-file violation offsets.
func expectPPMRasterMatchesPalette(raster []byte, dump *FrameDump, rasterOffset int) error {
	want := 3 * dump.Width * dump.Height
	if len(raster) != want {
		return violationf("ppm", "raster: have %d bytes, want %d (first violation at offset %d)",
			len(raster), want, rasterOffset+len(raster))
	}

	for i := range dump.Width * dump.Height {
		colour := dump.expandAt(i)
		for component := range colour {
			offset := 3*i + component
			if raster[offset] != colour[component] {
				x, y := i%dump.Width, i/dump.Width

				return violationf("ppm", "pixel (%d,%d): have %d, want %d (first violation at offset %d)",
					x, y, raster[offset], colour[component], rasterOffset+offset)
			}
		}
	}

	return nil
}

// magicPrefix returns the leading bytes of data for magic-mismatch messages.
func magicPrefix(data []byte) string {
	if len(data) > len(ppmMagic) {
		data = data[:len(ppmMagic)]
	}

	return string(data)
}

// appendPPMHeader writes the canonical P6 header fields.
func appendPPMHeader(dst []byte, width, height int) []byte {
	dst = append(dst, ppmMagic...)
	dst = append(dst, '\n')
	dst = strconv.AppendInt(dst, int64(width), 10)
	dst = append(dst, ' ')
	dst = strconv.AppendInt(dst, int64(height), 10)
	dst = append(dst, '\n')
	dst = strconv.AppendInt(dst, ppmMaxval, 10)
	dst = append(dst, '\n')

	return dst
}

// ppmHeaderLen is the canonical header length for a width x height frame.
func ppmHeaderLen(width, height int) int {
	return len(ppmMagic) + 1 + digitCount(width) + 1 + digitCount(height) + 1 + digitCount(ppmMaxval) + 1
}

// digitCount is the decimal digit count of a non-negative int.
func digitCount(value int) int {
	return len(strconv.Itoa(value))
}
