package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// parseFrameOffsets parses a CLX sprite-list header: a uint32 frame count
// followed by numFrames+1 little-endian uint32 frame offsets, where the last
// offset must equal the data length (clx_sprite.hpp:528-535; cl2_to_clx.cpp
// uses the same "last offset == size" test to tell a list from a sheet).
// Returns the frame count and the offsets slice (len numFrames+1).
func parseFrameOffsets(data []byte, format string) (int, []uint32, error) {
	if len(data) < 4 {
		return 0, nil, fmt.Errorf("%s data too short for frame count", format)
	}
	numFrames := int(binary.LittleEndian.Uint32(data[0:4]))
	if numFrames < 1 {
		return 0, nil, fmt.Errorf("%s: invalid frame count %d", format, numFrames)
	}
	need := 4 + 4*(numFrames+1)
	if len(data) < need {
		return 0, nil, fmt.Errorf("%s: data too short for %d frame offsets (need %d, have %d)",
			format, numFrames+1, need, len(data))
	}
	offsets := make([]uint32, numFrames+1)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(data[4+4*i : 8+4*i])
	}
	// List-vs-sheet disambiguation: in a single list, the final offset is
	// the data size. Otherwise the first uint32 was a group offset and this
	// is a sheet (clx_sprite.hpp:530-535).
	if offsets[numFrames] != uint32(len(data)) {
		return 0, nil, fmt.Errorf("%s: last frame offset %d != data size %d (sheets not supported)",
			format, offsets[numFrames], len(data))
	}
	return numFrames, offsets, nil
}

// decodeCl2Pixels decodes a CL2/CLX pixel-data command stream into
// width*height indexed pixels. Line boundaries are enforced strictly: every
// command must end exactly at the end of a row, except transparent runs,
// which may overrun into subsequent lines (clx_render.cpp:28-31 — "Transparent
// runs can cross line boundaries").
//
// Command stream (CL2.md §4.2, clx_decode.hpp:10-35):
//
//	0x00..0x7F: transparent run of `control` pixels (wrap to next line as needed)
//	0x80..0xBF: fill run: (0xBF - control) copies of the next byte
//	0xC1..0xFF: literal run: (-int8(control)) raw pixel bytes follow
func decodeCl2Pixels(src []byte, width, height int) ([]byte, error) {
	pixels := make([]byte, width*height)
	pos := 0 // write position in pixels
	i := 0
	for i < len(src) {
		if pos >= len(pixels) {
			break
		}
		control := src[i]
		i++
		switch {
		case control < 0x80: // transparent run
			pos += int(control)
		case control <= 0xBF: // fill run
			runLen := int(0xBF) - int(control)
			if i >= len(src) {
				return nil, errors.New("truncated fill run")
			}
			value := src[i]
			i++
			if pos+runLen > len(pixels) {
				return nil, fmt.Errorf("fill run of %d overruns frame at pixel %d", runLen, pos)
			}
			for j := 0; j < runLen; j++ {
				pixels[pos+j] = value
			}
			pos += runLen
		default: // literal run (control >= 0xC1)
			runLen := int(-int8(control))
			if i+runLen > len(src) {
				return nil, fmt.Errorf("literal run of %d overruns source at byte %d", runLen, i)
			}
			if pos+runLen > len(pixels) {
				return nil, fmt.Errorf("literal run of %d overruns frame at pixel %d", runLen, pos)
			}
			copy(pixels[pos:pos+runLen], src[i:i+runLen])
			i += runLen
			pos += runLen
		}
	}
	if pos != len(pixels) {
		return nil, fmt.Errorf("decoded %d pixels, expected %d", pos, len(pixels))
	}
	return pixels, nil
}

// DecodeCL2 decodes a CL2 file: a frame-count header and frame offsets in the
// same layout as CLX lists (cl2_to_clx.cpp:17-19 — CL2 and CLX share the list
// header; only the frame metadata differs), then per-frame CL2 pixel streams.
// Frame width and height are not stored in CL2 frames, so they must be given.
func DecodeCL2(data []byte, width, height int) ([]ClxFrame, error) {
	numFrames, offsets, err := parseFrameOffsets(data, "CL2")
	if err != nil {
		return nil, err
	}
	frames := make([]ClxFrame, 0, numFrames)
	for i := 0; i < numFrames; i++ {
		// CL2 frames carry only the CL2-style header (uint16 pixel-data
		// offset, CL2.md §2/§4.2); no width/height like CLX.
		start := int(offsets[i])
		if start+2 > int(offsets[i+1]) {
			return nil, fmt.Errorf("CL2 frame %d: too short for header", i)
		}
		headerSize := int(binary.LittleEndian.Uint16(data[start : start+2]))
		pixels, err := decodeCl2Pixels(data[start+headerSize:int(offsets[i+1])], width, height)
		if err != nil {
			return nil, fmt.Errorf("CL2 frame %d: %w", i, err)
		}
		frames = append(frames, ClxFrame{Width: width, Height: height, Pixels: pixels})
	}
	return frames, nil
}
