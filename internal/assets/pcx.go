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

// DecodePCX parses a PCX file and returns the frame with palette.
// PCX format (from pcx.hpp): 128-byte header + RLE compressed pixel data + 769-byte palette (0x0C + 256*3)
func DecodePCX(data []byte) (*PcxFrame, error) {
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

	decoded, err := decodePCXPixels(data[pcxHeaderSize:], width, height)
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

// decodePCXPixels decodes PCX RLE pixel data.
// PCX RLE: byte >= 0xC0 -> run length = byte & 0x3F, next byte = value
// byte < 0xC0 -> literal byte
// Palette: last 769 bytes (0x0C + 256*3) if present
func decodePCXPixels(data []byte, width, height int) (pcxPixelData, error) {
	pixelData, palette, hasPalette := splitPCXPalette(data)

	expectedPixels := width * height
	pixels := make([]byte, expectedPixels)
	pixelIdx, err := expandPCXRLE(pixelData, pixels)
	if err != nil {
		return pcxPixelData{}, err
	}

	if pixelIdx != expectedPixels {
		return pcxPixelData{}, fmt.Errorf("PCX: decoded %d pixels, expected %d", pixelIdx, expectedPixels)
	}

	return pcxPixelData{pixels: pixels, palette: palette, hasPalette: hasPalette}, nil
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

// expandPCXRLE fills pixels from PCX RLE data and returns how many were written.
func expandPCXRLE(data, pixels []byte) (int, error) {
	pixelIdx := 0

	for pixelIdx < len(pixels) && len(data) > 0 {
		b := data[0]
		data = data[1:]

		if b < 0xC0 {
			pixels[pixelIdx] = b
			pixelIdx++

			continue
		}

		var err error
		pixelIdx, data, err = writePCXRun(data, pixels, pixelIdx, int(b&0x3F))
		if err != nil {
			return 0, err
		}
	}

	return pixelIdx, nil
}

// writePCXRun writes a run of runLen copies of the next data byte into pixels
// starting at pixelIdx and returns the updated pixel index and remaining data.
func writePCXRun(data, pixels []byte, pixelIdx, runLen int) (int, []byte, error) {
	if runLen == 0 {
		return pixelIdx, data, nil
	}

	if len(data) < 1 {
		return pixelIdx, data, errors.New("PCX: RLE run needs value byte")
	}

	value := data[0]
	data = data[1:]

	for j := 0; j < runLen && pixelIdx < len(pixels); j++ {
		pixels[pixelIdx] = value
		pixelIdx++
	}

	return pixelIdx, data, nil
}
