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

// DecodeCEL parses a CEL sprite list and converts it to CL2 frames.
// CEL files use a different RLE encoding (CelToClx in devilutionX).
// They can be single frames or multi-frame groups.
func DecodeCEL(data []byte) (*CelSpriteList, error) {
	groups, err := celGroups(data)
	if err != nil {
		return nil, err
	}

	list := &CelSpriteList{Frames: make([]CelFrame, 0)}

	for _, groupData := range groups {
		frames, err := decodeCelGroup(groupData)
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

// decodeCelGroup decodes a single CEL group (which uses CL2-like encoding after conversion)
func decodeCelGroup(data []byte) ([]CelFrame, error) {
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

	offsets := make([]uint32, numFrames+1)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(data[4+4*i : 8+4*i])
	}

	// List-vs-sheet disambiguation: last offset should equal data size for single list
	// Multi-frame list: last offset is end of last frame, not end of data
	// This is normal for multi-frame CEL files

	frames := make([]CelFrame, 0, numFrames)
	for i := range numFrames {
		frameData := data[offsets[i]:offsets[i+1]]

		frame, err := decodeCelFrame(frameData)
		if err != nil {
			return nil, fmt.Errorf("CEL frame %d: %w", i, err)
		}

		frames = append(frames, *frame)
	}

	return frames, nil
}

// decodeCelFrame decodes one CEL frame using CEL RLE encoding.
func decodeCelFrame(data []byte) (*CelFrame, error) {
	width, height, err := celFrameHeader(data)
	if err != nil {
		return nil, err
	}

	pixels, err := decodeCelFrameBody(data[6:], width*height)
	if err != nil {
		return nil, err
	}

	return &CelFrame{Width: width, Height: height, Pixels: pixels}, nil
}

// celFrameHeader validates the 6-byte CEL frame header and returns its dimensions.
func celFrameHeader(data []byte) (width, height int, err error) {
	if len(data) < 6 {
		return 0, 0, fmt.Errorf("CEL frame header too short (need 6, have %d)", len(data))
	}

	headerSize := int(data[0])
	if headerSize != 6 {
		return 0, 0, fmt.Errorf("CEL frame header size mismatch (expected 6, got %d)", headerSize)
	}

	width = int(binary.LittleEndian.Uint16(data[2:4]))
	height = int(binary.LittleEndian.Uint16(data[4:6]))

	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("CEL: invalid dimensions %dx%d", width, height)
	}

	return width, height, nil
}

// decodeCelFrameBody expands CEL RLE pixel data, requiring exactly nominalPixels output.
func decodeCelFrameBody(data []byte, nominalPixels int) ([]byte, error) {
	pixels := make([]byte, 0, nominalPixels)

	for i := 0; i < len(data); {
		next, err := appendCelRun(&pixels, data, i)
		if err != nil {
			return nil, err
		}

		i = next
	}

	if len(pixels) != nominalPixels {
		return nil, fmt.Errorf("CEL: pixel count mismatch (got %d, expected %d)", len(pixels), nominalPixels)
	}

	return pixels, nil
}

// appendCelRun consumes the RLE run starting at offset i, appends its pixels and
// returns the offset of the next run.
func appendCelRun(pixels *[]byte, data []byte, i int) (int, error) {
	ctrl := data[i]
	next := i + 1

	if ctrl >= 0x80 {
		// Transparent run
		*pixels = appendTransparentRun(*pixels, int(-int8(ctrl)))

		return next, nil
	}

	// Literal run (control byte = count-1)
	count := int(ctrl) + 1
	if count > len(data)-next {
		return 0, fmt.Errorf("CEL literal run overflow: need %d, have %d", count, len(data)-next)
	}

	*pixels = append(*pixels, data[next:next+count]...)

	return next + count, nil
}

// appendTransparentRun appends n transparent (zero) pixels.
func appendTransparentRun(pixels []byte, n int) []byte {
	for range n {
		pixels = append(pixels, 0)
	}

	return pixels
}
