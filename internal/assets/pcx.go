// Package assets implements decoders for Diablo asset formats (CLX, CL2,
// CEL, PCX, TRN, palette) into engine-ready indexed pixel data.
//
// All decoders are pure Go and produce 256-colour indexed []byte frames;
// palette expansion happens in the render layer, not here.
package assets

import (
	"encoding/binary"
	"fmt"
)

// PCXHeader represents the PCX file header (128 bytes).
type PCXHeader struct {
	Manufacturer  uint8
	Version       uint8
	Encoding      uint8
	BitsPerPixel  uint8
	Xmin          uint16
	Ymin          uint16
	Xmax          uint16
	Ymax          uint16
	HDpi          uint16
	VDpi          uint16
	Colormap      [48]uint8
	Reserved      uint8
	NPlanes       uint8
	BytesPerLine  uint16
	PaletteInfo   uint16
	HscreenSize   uint16
	VscreenSize   uint16
	Filler        [54]uint8
}

const pcxHeaderSize = 128

// PCXPalette is a 256-color palette (768 bytes RGB).
type PCXPalette [256][3]uint8

// PcxFrame is one decoded PCX sprite frame as 256-colour indexed pixels.
type PcxFrame struct {
	Width    int
	Height   int
	Pixels   []byte
	Palette  PCXPalette
	HasPalette bool
}

// DecodePCX parses a PCX file and returns the frame with palette.
// PCX format (from pcx.hpp): 128-byte header + RLE compressed pixel data + 769-byte palette (0x0C + 256*3)
func DecodePCX(data []byte) (*PcxFrame, error) {
	if len(data) < pcxHeaderSize {
		return nil, fmt.Errorf("PCX data too short for header (need %d, have %d)", pcxHeaderSize, len(data))
	}

	header := parsePCXHeader(data[:pcxHeaderSize])

	// Validate header
	if header.Manufacturer != 0x0A {
		return nil, fmt.Errorf("PCX: invalid manufacturer byte 0x%02X (expected 0x0A)", header.Manufacturer)
	}
	if header.Encoding != 1 {
		return nil, fmt.Errorf("PCX: unsupported encoding %d (only RLE=1 supported)", header.Encoding)
	}
	if header.BitsPerPixel != 8 {
		return nil, fmt.Errorf("PCX: unsupported bits per pixel %d (only 8-bit indexed supported)", header.BitsPerPixel)
	}
	if header.NPlanes != 1 {
		return nil, fmt.Errorf("PCX: unsupported number of planes %d (only 1 plane supported)", header.NPlanes)
	}

	width := int(header.Xmax) - int(header.Xmin) + 1
	height := int(header.Ymax) - int(header.Ymin) + 1
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("PCX: invalid dimensions %dx%d", width, height)
	}

	// Decode RLE pixel data
	pixelData := data[pcxHeaderSize:]
	pixels, palette, hasPalette, err := decodePCXPixels(pixelData, width, height)
	if err != nil {
		return nil, fmt.Errorf("PCX: pixel decode: %w", err)
	}

	return &PcxFrame{
		Width:      width,
		Height:     height,
		Pixels:     pixels,
		Palette:    palette,
		HasPalette: hasPalette,
	}, nil
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
func decodePCXPixels(data []byte, width, height int) ([]byte, PCXPalette, bool, error) {
	expectedPixels := width * height
	pixels := make([]byte, expectedPixels)
	pixelIdx := 0

	var palette PCXPalette
	hasPalette := false

	// Check if there's a palette at the end (769 bytes: 0x0C + 256*3)
	if len(data) >= 769 {
		paletteStart := len(data) - 769
		if data[paletteStart] == 0x0C {
			// Palette present
			paletteData := data[paletteStart+1 : paletteStart+769]
			for i := 0; i < 256; i++ {
				palette[i][0] = paletteData[i*3]
				palette[i][1] = paletteData[i*3+1]
				palette[i][2] = paletteData[i*3+2]
			}
			hasPalette = true
			data = data[:paletteStart] // Remove palette from pixel data
		}
	}

	for pixelIdx < expectedPixels && len(data) > 0 {
		b := data[0]
		data = data[1:]

		if b >= 0xC0 {
			// Run length encoded
			runLen := int(b & 0x3F)
			if runLen == 0 {
				continue
			}
			if len(data) < 1 {
				return nil, palette, hasPalette, fmt.Errorf("PCX: RLE run needs value byte")
			}
			value := data[0]
			data = data[1:]
			for j := 0; j < runLen && pixelIdx < expectedPixels; j++ {
				pixels[pixelIdx] = value
				pixelIdx++
			}
		} else {
			// Literal
			pixels[pixelIdx] = b
			pixelIdx++
		}
	}

	if pixelIdx != expectedPixels {
		return nil, palette, hasPalette, fmt.Errorf("PCX: decoded %d pixels, expected %d", pixelIdx, expectedPixels)
	}

	return pixels, palette, hasPalette, nil
}