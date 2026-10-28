// Package assets implements decoders for Diablo asset formats (CLX, CL2,
// CEL, PCX, TRN, palette) into engine-ready indexed pixel data.
//
// All decoders are pure Go and produce 256-color indexed []byte frames;
// palette expansion happens in the render layer, not here.
package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const pcxHeaderSize = 128

// pcxLayout is the frame geometry the pixel decoder needs.
type pcxLayout struct {
	width        int
	height       int
	bytesPerLine int
	transparent  *byte
}

// PCXHeader represents the PCX file header (128 bytes).
type PCXHeader struct {
	Manufacturer uint8
	Version      uint8
	Encoding     uint8
	BitsPerPixel uint8
	Xmin         uint16
	Ymin         uint16
	Xmax         uint16
	Ymax         uint16
	HDpi         uint16
	VDpi         uint16
	Colormap     [48]uint8
	Reserved     uint8
	NPlanes      uint8
	BytesPerLine uint16
	PaletteInfo  uint16
	HscreenSize  uint16
	VscreenSize  uint16
	Filler       [54]uint8
}

// PCXPalette is a 256-color palette (768 bytes RGB).
type PCXPalette [256][3]uint8

// PcxFrame is one decoded PCX sprite frame as 256-color indexed pixels.
type PcxFrame struct {
	Width      int
	Height     int
	Pixels     []byte
	Palette    PCXPalette
	HasPalette bool
}

// pcxPixelData is decoded PCX pixel payload: indexed pixels plus the trailing
// palette when the file carries one.
type pcxPixelData struct {
	pixels     []byte
	palette    PCXPalette
	hasPalette bool
}

// DecodePCX parses a PCX file and returns the frame with palette. Frames come
// out bottom-up because "the lines are reversed in CEL" (pcx_to_clx.cpp:123,169),
// scanlines are decoded bytesPerLine wide with the padding dropped (:91-93,141-163),
// and the optional transparent palette index becomes transparent index 0 (:94,172-190).
// PCX format: 128-byte header + RLE compressed pixel data + 769-byte palette (0x0C + 256*3).
func DecodePCX(data []byte, transparentColor ...byte) (*PcxFrame, error) {
	if len(data) < pcxHeaderSize {
		return nil, fmt.Errorf("PCX data too short for header (need %d, have %d)", pcxHeaderSize, len(data))
	}

	header := parsePCXHeader(data[:pcxHeaderSize])
	if err := validatePCXHeader(header); err != nil {
		return nil, err
	}

	width := int(header.Xmax) - int(header.Xmin) + 1

	height := int(header.Ymax) - int(header.Ymin) + 1
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("PCX: invalid dimensions %dx%d", width, height)
	}

	if int(header.BytesPerLine) < width {
		return nil, fmt.Errorf("PCX: BytesPerLine %d is less than the image width %d",
			header.BytesPerLine, width)
	}

	layout := pcxLayout{
		width:        width,
		height:       height,
		bytesPerLine: int(header.BytesPerLine),
		transparent:  optionalPCXColor(transparentColor),
	}

	decoded, err := decodePCXPixels(data[pcxHeaderSize:], layout)
	if err != nil {
		return nil, fmt.Errorf("PCX: pixel decode: %w", err)
	}

	return &PcxFrame{
		Width:      width,
		Height:     height,
		Pixels:     decoded.pixels,
		Palette:    decoded.palette,
		HasPalette: decoded.hasPalette,
	}, nil
}

// optionalPCXColor reads the variadic transparent colour.
func optionalPCXColor(transparentColor []byte) *byte {
	if len(transparentColor) == 0 {
		return nil
	}

	return &transparentColor[0]
}

// validatePCXHeader rejects headers this decoder cannot read.
func validatePCXHeader(header PCXHeader) error {
	if header.Manufacturer != 0x0A {
		return fmt.Errorf("PCX: invalid manufacturer byte 0x%02X (expected 0x0A)", header.Manufacturer)
	}

	if header.Encoding != 1 {
		return fmt.Errorf("PCX: unsupported encoding %d (only RLE=1 supported)", header.Encoding)
	}

	if header.BitsPerPixel != 8 {
		return fmt.Errorf("PCX: unsupported bits per pixel %d (only 8-bit indexed supported)", header.BitsPerPixel)
	}

	if header.NPlanes != 1 {
		return fmt.Errorf("PCX: unsupported number of planes %d (only 1 plane supported)", header.NPlanes)
	}

	return nil
}

func parsePCXHeader(data []byte) PCXHeader {
	var h PCXHeader

	h.Manufacturer = data[0]
	h.Version = data[1]
	h.Encoding = data[2]
	h.BitsPerPixel = data[3]
	h.Xmin = binary.LittleEndian.Uint16(data[4:6])
	h.Ymin = binary.LittleEndian.Uint16(data[6:8])
	h.Xmax = binary.LittleEndian.Uint16(data[8:10])
	h.Ymax = binary.LittleEndian.Uint16(data[10:12])
	h.HDpi = binary.LittleEndian.Uint16(data[12:14])
	h.VDpi = binary.LittleEndian.Uint16(data[14:16])
	copy(h.Colormap[:], data[16:64])
	h.Reserved = data[64]
	h.NPlanes = data[65]
	h.BytesPerLine = binary.LittleEndian.Uint16(data[66:68])
	h.PaletteInfo = binary.LittleEndian.Uint16(data[68:70])
	h.HscreenSize = binary.LittleEndian.Uint16(data[70:72])
	h.VscreenSize = binary.LittleEndian.Uint16(data[72:74])
	copy(h.Filler[:], data[74:128])

	return h
}

// decodePCXPixels decodes PCX RLE pixel data, one bytesPerLine scanline at a
// time, then emits the frame bottom-up with the transparent colour blanked.
func decodePCXPixels(data []byte, layout pcxLayout) (pcxPixelData, error) {
	pixelData, palette, hasPalette := splitPCXPalette(data)

	rows, err := decodePCXRows(pixelData, layout)
	if err != nil {
		return pcxPixelData{}, err
	}

	return pcxPixelData{
		pixels:     emitPCXFrame(rows, layout),
		palette:    palette,
		hasPalette: hasPalette,
	}, nil
}

// decodePCXRows decodes each scanline on bytesPerLine boundaries and keeps the
// first width bytes of every line (pcx_to_clx.cpp:141-163).
func decodePCXRows(pixelData []byte, layout pcxLayout) ([][]byte, error) {
	rows := make([][]byte, layout.height)
	line := make([]byte, layout.bytesPerLine)

	for j := range layout.height {
		rest, err := decodePCXLine(pixelData, line)
		if err != nil {
			return nil, fmt.Errorf("PCX row %d: %w", j, err)
		}

		rows[j] = append([]byte(nil), line[:layout.width]...)
		pixelData = rest
	}

	return rows, nil
}

// decodePCXLine fills one scanline of exactly len(line) pixels.
func decodePCXLine(data, line []byte) ([]byte, error) {
	for x := 0; x < len(line); {
		if len(data) == 0 {
			return nil, errors.New("PCX pixel data is truncated")
		}

		var err error
		if data[0] <= 0xBF {
			line[x] = data[0]
			data, x = data[1:], x+1

			continue
		}

		data, x, err = copyPCXRun(data, line, x)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// copyPCXRun expands one 0xC0..0xFF run into line at x and returns the new
// data offset and x. A run never crosses a scanline boundary.
func copyPCXRun(data, line []byte, x int) ([]byte, int, error) {
	runLen := int(data[0] & 0x3F)
	if runLen > len(line)-x {
		return nil, 0, errors.New("PCX run crosses a scanline boundary")
	}

	if len(data) < 2 {
		return nil, 0, errors.New("PCX pixel data is truncated")
	}

	value := data[1]

	for j := range runLen {
		line[x+j] = value
	}

	return data[2:], x + runLen, nil
}

// emitPCXFrame writes the rows into one buffer bottom-up, turning the
// transparent colour into index 0 (pcx_to_clx.cpp:169,172-190).
func emitPCXFrame(rows [][]byte, layout pcxLayout) []byte {
	pixels := make([]byte, 0, layout.width*layout.height)

	for j := range layout.height {
		row := rows[layout.height-(j+1)]
		for _, v := range row {
			if layout.transparent != nil && v == *layout.transparent {
				pixels = append(pixels, 0)

				continue
			}

			pixels = append(pixels, v)
		}
	}

	return pixels
}

// splitPCXPalette peels a trailing 256-color palette off the pixel data.
func splitPCXPalette(data []byte) ([]byte, PCXPalette, bool) {
	var palette PCXPalette

	if len(data) < 769 {
		return data, palette, false
	}

	paletteStart := len(data) - 769
	if data[paletteStart] != 0x0C {
		return data, palette, false
	}

	paletteData := data[paletteStart+1 : paletteStart+769]
	for i := range 256 {
		palette[i][0] = paletteData[i*3]
		palette[i][1] = paletteData[i*3+1]
		palette[i][2] = paletteData[i*3+2]
	}

	return data[:paletteStart], palette, true
}
