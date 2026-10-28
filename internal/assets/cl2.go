package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// clxFillMax is the last fill control byte; 0xBF already means a 65-pixel
// literal run (clx_decode.hpp:25).
const clxFillMax = 0xBE

// DecodeCL2 decodes a CL2 file into indexed frames. A plain list has the CLX
// list header (cl2_to_clx.cpp:17-19); a sheet carries a group-offset table
// instead (:23-28) and decodes into GroupSizes with one entry per list.
// Frame width and height are not stored in CL2 frames, so they must be given.
func DecodeCL2(data []byte, width, height int) (*ClxSpriteList, error) {
	if !isFrameList(data) {
		return decodeCL2Sheet(data, width, height)
	}

	frames, err := decodeCL2Group(data, width, height)
	if err != nil {
		return nil, err
	}

	return &ClxSpriteList{Frames: frames}, nil
}

// decodeCL2Sheet decodes every list of a CL2 sheet, flattening the frames and
// recording each list's frame count.
func decodeCL2Sheet(data []byte, width, height int) (*ClxSpriteList, error) {
	groups, err := groupRanges(data)
	if err != nil {
		return nil, fmt.Errorf("CL2 %w", err)
	}

	list := &ClxSpriteList{GroupSizes: make([]int, 0, len(groups))}
	for i, group := range groups {
		frames, err := decodeCL2Group(group, width, height)
		if err != nil {
			return nil, fmt.Errorf("CL2 group %d: %w", i, err)
		}

		list.Frames = append(list.Frames, frames...)
		list.GroupSizes = append(list.GroupSizes, len(frames))
	}

	return list, nil
}

// isFrameList reports whether the buffer is a single frame list: a list's last
// frame offset equals the buffer size (clx_sprite.hpp:447-457).
func isFrameList(data []byte) bool {
	if len(data) < 8 {
		return true
	}

	numFrames := int(binary.LittleEndian.Uint32(data[0:4]))
	last := 4 + 4*(numFrames+1)
	if numFrames < 1 || len(data) < last {
		return true
	}

	return binary.LittleEndian.Uint32(data[last-4:last]) == uint32(len(data))
}

// decodeCL2Group decodes one CL2 frame list.
func decodeCL2Group(data []byte, width, height int) ([]ClxFrame, error) {
	numFrames, offsets, err := parseFrameOffsets(data, "CL2")
	if err != nil {
		return nil, err
	}

	frames := make([]ClxFrame, 0, numFrames)
	for i := range numFrames {
		pixels, err := decodeCL2Frame(data, offsets[i], offsets[i+1], width, height)
		if err != nil {
			return nil, fmt.Errorf("CL2 frame %d: %w", i, err)
		}

		frames = append(frames, ClxFrame{Width: width, Height: height, Pixels: pixels})
	}

	return frames, nil
}

// decodeCL2Frame decodes one CL2 frame: a uint16 pixel-data offset header,
// then the CL2 command stream (CL2.md §2/§4.2). CL2 frames store no width or
// height, unlike CLX.
func decodeCL2Frame(data []byte, start, end uint32, width, height int) ([]byte, error) {
	from, to := int(start), int(end)
	if from+2 > to || from+2 > len(data) {
		return nil, errors.New("too short for header")
	}

	headerSize := int(binary.LittleEndian.Uint16(data[from : from+2]))
	if from+headerSize > to {
		return nil, fmt.Errorf("header size %d overruns the frame", headerSize)
	}

	return decodeCl2Pixels(data[from+headerSize:to], width, height)
}

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

	// A single list's final offset is the data size; anything else means the
	// caller should have split the buffer into groups first
	// (clx_sprite.hpp:447-457, cl2_to_clx.cpp:23-28).
	if offsets[numFrames] != uint32(len(data)) {
		return 0, nil, fmt.Errorf("%s: last frame offset %d != data size %d",
			format, offsets[numFrames], len(data))
	}

	return numFrames, offsets, nil
}

// groupRanges splits a sheet buffer into per-group slices: the leading u32
// values are group offsets and group i ends where group i+1 starts
// (cl2_to_clx.cpp:23-28, clx_sprite.hpp:447-457).
func groupRanges(data []byte) ([][]byte, error) {
	numGroups := int(binary.LittleEndian.Uint32(data[0:4])) / 4
	if numGroups < 1 || len(data) < 4*numGroups {
		return nil, fmt.Errorf("sheet: group count %d does not fit %d bytes", numGroups, len(data))
	}

	groupOffsets := make([]int, numGroups+1)
	for i := range numGroups {
		groupOffsets[i] = int(binary.LittleEndian.Uint32(data[4*i : 4+4*i]))
	}

	groupOffsets[numGroups] = len(data)

	groups := make([][]byte, numGroups)
	for i := range numGroups {
		start, end := groupOffsets[i], groupOffsets[i+1]
		if start < 4*numGroups || start > end || end > len(data) {
			return nil, fmt.Errorf("sheet: group %d range [%d,%d) is invalid", i, start, end)
		}

		groups[i] = data[start:end]
	}

	return groups, nil
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

		if err := decodeOneCommand(src, &i, pixels, &pos); err != nil {
			return nil, err
		}
	}

	if pos != len(pixels) {
		return nil, fmt.Errorf("decoded %d pixels, expected %d", pos, len(pixels))
	}

	return pixels, nil
}

// decodeFillRun handles the 0x80..0xBE fill command (clx_decode.hpp:25,31-35).
// control is the command byte; runLen = 0xBF - control, i.e. 1..63.
func decodeFillRun(src []byte, i *int, pixels []byte, pos *int, control byte) error {
	runLen := int(0xBF) - int(control)

	if *i >= len(src) {
		return errors.New("truncated fill run")
	}

	value := src[*i]
	*i++

	if *pos+runLen > len(pixels) {
		return fmt.Errorf("fill run of %d overruns frame at pixel %d", runLen, *pos)
	}

	for j := range runLen {
		pixels[*pos+j] = value
	}

	*pos += runLen

	return nil
}

// decodeLiteralRun handles the 0xBF and 0xC0..0xFF literal command
// (clx_decode.hpp:28-30); runLen = 256 - control, i.e. 65 for 0xBF and 1..64
// above it.
func decodeLiteralRun(src []byte, i *int, pixels []byte, pos *int, control byte) error {
	runLen := int(-int8(control))
	if *i+runLen > len(src) {
		return fmt.Errorf("literal run of %d overruns source at byte %d", runLen, *i)
	}

	if *pos+runLen > len(pixels) {
		return fmt.Errorf("literal run of %d overruns frame at pixel %d", runLen, *pos)
	}

	copy(pixels[*pos:*pos+runLen], src[*i:*i+runLen])
	*i += runLen
	*pos += runLen

	return nil
}

// decodeTransparentRun handles the 0x00..0x7F transparent command (CL2.md §4.2).
// control is the number of transparent pixels to skip.
func decodeTransparentRun(pos *int, control byte) error {
	skip := int(control)
	// transparent run just advances position, no bounds check on pixels
	// since transparent means "skip/wrap"
	*pos += skip

	return nil
}

// decodeOneCommand dispatches a single CL2 command byte to the appropriate
// decoder (transparent / fill / literal). Mirrors clx_decode.hpp:10-35.
func decodeOneCommand(src []byte, i *int, pixels []byte, pos *int) error {
	if *i >= len(src) {
		return errors.New("unexpected EOF in command stream")
	}

	control := src[*i]
	*i++

	switch {
	case control < 0x80: // transparent run
		return decodeTransparentRun(pos, control)
	case control <= clxFillMax: // fill run, 0x80..0xBE
		return decodeFillRun(src, i, pixels, pos, control)
	default: // literal run, 0xBF and 0xC0..0xFF
		return decodeLiteralRun(src, i, pixels, pos, control)
	}
}
