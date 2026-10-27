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

const (
	// celFrameHeaderSize is the CEL frame header length (cel_to_clx.cpp:81).
	celFrameHeaderSize = 10
	// celTransparentMin is the first control byte that means "transparent run"
	// (cel_to_clx.cpp:24-26).
	celTransparentMin = 0x80
)

// CelFrame is one decoded CEL sprite frame as 256-color indexed pixels.
type CelFrame struct {
	Width  int
	Height int
	// Pixels is row-major indexed pixel data of size Width*Height.
	Pixels []byte
}

// CelSpriteList is a decoded CEL sprite list (a single animation).
// Use DecodeCEL to parse CEL files; multi-frame CELs are returned as a list.
type CelSpriteList struct {
	Frames []CelFrame
}

// DecodeCEL parses a CEL sprite list and decodes every frame to indexed
// pixels. CEL is devilutionX's legacy container; CelToClx converts it to CLX at
// load (cel_to_clx.cpp:35-140). The caller supplies the frame width because the
// CEL frame header carries no usable width (cel_to_clx.cpp:84-86); height is
// derived from the decoded line count (:88-111). Upstream's
// PointerOrValue<uint16_t> also accepts a per-frame width array
// (cursor.cpp:448); that surface is deferred to devilGo-l5fp.
//
// DecodeCEL keeps decoded indices; devilutionX's CelToClx re-encodes to CL2 only for
// its runtime sprite type — same pixels, different container.
func DecodeCEL(data []byte, width uint16) (*CelSpriteList, error) {
	if width == 0 {
		return nil, errors.New("CEL: frame width must be positive")
	}

	groups, err := celGroups(data)
	if err != nil {
		return nil, err
	}

	list := &CelSpriteList{Frames: make([]CelFrame, 0)}

	for _, groupData := range groups {
		frames, err := decodeCelGroup(groupData, int(width))
		if err != nil {
			return nil, fmt.Errorf("CEL group: %w", err)
		}

		list.Frames = append(list.Frames, frames...)
	}

	return list, nil
}

// celGroups splits CEL data into the groups to decode: the whole buffer for a
// plain frame list, or the offset-table slices for a multi-group file.
func celGroups(data []byte) ([][]byte, error) {
	if len(data) < 4 {
		return nil, errors.New("CEL data too short for frame count")
	}

	// First 4 bytes: number of frames or pointer to group offsets.
	maybeNumFrames := int(binary.LittleEndian.Uint32(data[0:4]))
	if !celHasOffsetTable(data, maybeNumFrames) {
		return [][]byte{data}, nil
	}

	// In devilutionX: if LoadLE32(data[(maybeNumFrames*4)+4]) != size, it is a multi-group file.
	lastOffset := int(binary.LittleEndian.Uint32(data[4+maybeNumFrames*4 : 8+maybeNumFrames*4]))
	if lastOffset == len(data) {
		return [][]byte{data}, nil
	}

	return celMultiGroups(data, maybeNumFrames/4), nil
}

func celHasOffsetTable(data []byte, maybeNumFrames int) bool {
	return maybeNumFrames > 0 && len(data) >= 4+maybeNumFrames*4
}

// celMultiGroups slices a multi-group CEL into one buffer per group offset range.
// The header is the group offset table itself: maybeNumFrames is the first
// group's address, so entry i lives at data[4*i] (cel_to_clx.cpp:60-66).
func celMultiGroups(data []byte, numGroups int) [][]byte {
	groupOffsets := make([]uint32, numGroups+1)
	for i := range numGroups {
		groupOffsets[i] = binary.LittleEndian.Uint32(data[4*i : 4+4*i])
	}

	groupOffsets[numGroups] = uint32(len(data))

	groups := make([][]byte, 0, numGroups)
	for i := range numGroups {
		start := groupOffsets[i]

		end := groupOffsets[i+1]
		if start < end && end <= uint32(len(data)) {
			groups = append(groups, data[start:end])
		}
	}

	return groups
}

// decodeCelGroup decodes a single CEL group (the whole file for a plain frame
// list, or one offset-table slice for a multi-group file).
func decodeCelGroup(data []byte, width int) ([]CelFrame, error) {
	if len(data) < 4 {
		return nil, errors.New("CEL group data too short for frame count")
	}

	numFrames := int(binary.LittleEndian.Uint32(data[0:4]))
	if numFrames < 1 {
		return nil, fmt.Errorf("CEL: invalid frame count %d", numFrames)
	}

	need := 4 + 4*(numFrames+1)
	if len(data) < need {
		return nil, fmt.Errorf("CEL: data too short for %d frame offsets (need %d, have %d)",
			numFrames+1, need, len(data))
	}

	offsets, err := celFrameOffsets(data, numFrames)
	if err != nil {
		return nil, err
	}

	frames := make([]CelFrame, 0, numFrames)
	for i := range numFrames {
		frameData := data[offsets[i]:offsets[i+1]]

		frame, err := decodeCelFrame(frameData, width)
		if err != nil {
			return nil, fmt.Errorf("CEL frame %d: %w", i, err)
		}

		frames = append(frames, *frame)
	}

	return frames, nil
}

// celFrameOffsets reads the frame offset table and validates it: offsets must
// be non-decreasing and the last one must stay inside the group data. A
// malformed CEL would otherwise panic on the per-frame slice.
func celFrameOffsets(data []byte, numFrames int) ([]uint32, error) {
	offsets := make([]uint32, numFrames+1)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(data[4+4*i : 8+4*i])
	}

	for i := 1; i < len(offsets); i++ {
		if offsets[i] < offsets[i-1] {
			return nil, fmt.Errorf("CEL: frame offsets not non-decreasing: offset %d = %d after %d",
				i, offsets[i], offsets[i-1])
		}
	}

	if last := offsets[len(offsets)-1]; last > uint32(len(data)) {
		return nil, fmt.Errorf("CEL: last frame offset %d past the end of %d bytes", last, len(data))
	}

	return offsets, nil
}

// decodeCelFrame decodes one CEL frame: skip the 10-byte header when present,
// then read one line of width pixels at a time until the buffer ends.
func decodeCelFrame(data []byte, width int) (*CelFrame, error) {
	pixels, height, err := decodeCelLines(skipCelFrameHeader(data), width)
	if err != nil {
		return nil, err
	}

	return &CelFrame{Width: width, Height: height, Pixels: pixels}, nil
}

// skipCelFrameHeader drops the CEL frame header when the frame has one.
func skipCelFrameHeader(data []byte) []byte {
	if len(data) >= 2 && binary.LittleEndian.Uint16(data[0:2]) == celFrameHeaderSize {
		return data[celFrameHeaderSize:]
	}

	return data
}

// decodeCelLines decodes whole lines until the buffer is exhausted; the line
// count is the frame height (cel_to_clx.cpp:88-111).
func decodeCelLines(src []byte, width int) ([]byte, int, error) {
	pixels := make([]byte, 0, width*len(src)/(width+1)+width)

	height := 0
	for len(src) > 0 {
		line, rest, err := decodeCelLine(src, width)
		if err != nil {
			return nil, 0, fmt.Errorf("CEL line %d: %w", height, err)
		}

		pixels = append(pixels, line...)
		src = rest
		height++
	}

	if height == 0 {
		return nil, 0, errors.New("CEL: frame has no lines")
	}

	return pixels, height, nil
}

// decodeCelLine consumes runs until the line holds exactly width pixels.
func decodeCelLine(src []byte, width int) (line, rest []byte, err error) {
	line = make([]byte, 0, width)
	for len(line) < width {
		run, next, err := nextCelRun(src)
		if err != nil {
			return nil, nil, err
		}

		if len(line)+len(run) > width {
			return nil, nil, fmt.Errorf("CEL run of %d pixels overruns the %d-pixel line", len(run), width)
		}

		line = append(line, run...)
		src = next
	}

	return line, src, nil
}

// nextCelRun returns one CEL run's pixels and the offset after it. A control
// byte >= 0x80 is a transparent run of -int8(control) pixels; anything below is
// a literal run of control pixels (cel_to_clx.cpp:24-32, :88-100).
func nextCelRun(src []byte) (run, rest []byte, err error) {
	if len(src) == 0 {
		return nil, nil, errors.New("CEL: line truncated before reaching the width")
	}

	ctrl := src[0]
	rest = src[1:]

	if ctrl >= celTransparentMin {
		return make([]byte, int(-int8(ctrl))), rest, nil
	}

	count := int(ctrl)
	if count == 0 {
		return nil, nil, errors.New("CEL: zero-length literal run")
	}

	if count > len(rest) {
		return nil, nil, fmt.Errorf("CEL literal run overflow: need %d, have %d", count, len(rest))
	}

	return rest[:count], rest[count:], nil
}
