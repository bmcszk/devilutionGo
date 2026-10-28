// Package assets implements decoders for Diablo asset formats (CLX, CL2,
// CEL, PCX, TRN, palette) into engine-ready indexed pixel data.
//
// All decoders are pure Go and produce 256-color indexed []byte frames;
// palette expansion happens in the render layer, not here.
package assets

import (
	"encoding/binary"
	"fmt"
)

// ClxFrame is one decoded CLX sprite frame as 256-color indexed pixels.
type ClxFrame struct {
	Width  int
	Height int
	// Pixels is row-major indexed pixel data of size Width*Height.
	Pixels []byte
}

// ClxSpriteList is a decoded CLX sprite list (a single animation).
type ClxSpriteList struct {
	Frames []ClxFrame
	// GroupSizes is the frame count of each list in the file; nil for a plain
	// list, one entry per list for a sheet.
	GroupSizes []int
}

// Group returns the frames of one list in a sheet, or nil when the file is a
// plain list (clx_sprite.hpp:459-500 ClxSpriteSheet).
func (l *ClxSpriteList) Group(index int) []ClxFrame {
	if index < 0 || index >= len(l.GroupSizes) {
		return nil
	}

	start := 0
	for i := range index {
		start += l.GroupSizes[i]
	}

	return l.Frames[start : start+l.GroupSizes[index]]
}

// DecodeCLX parses a CLX sprite list, or a sheet of lists, and decodes every
// frame to indexed pixels. devilutionX tells the two apart by the last frame
// offset (clx_sprite.hpp:447-457): equal to the buffer size means a list,
// otherwise the leading offsets address the sheet's lists (cl2_to_clx.cpp:23-28).
func DecodeCLX(data []byte) (*ClxSpriteList, error) {
	if !isFrameList(data) {
		return decodeCLXSheet(data)
	}

	frames, err := decodeCLXGroup(data)
	if err != nil {
		return nil, err
	}

	return &ClxSpriteList{Frames: frames}, nil
}

// decodeCLXSheet decodes every list of a sheet, flattening the frames and
// recording each list's frame count.
func decodeCLXSheet(data []byte) (*ClxSpriteList, error) {
	groups, err := groupRanges(data)
	if err != nil {
		return nil, fmt.Errorf("CLX %w", err)
	}

	list := &ClxSpriteList{GroupSizes: make([]int, 0, len(groups))}
	for i, group := range groups {
		frames, err := decodeCLXGroup(group)
		if err != nil {
			return nil, fmt.Errorf("CLX group %d: %w", i, err)
		}

		list.Frames = append(list.Frames, frames...)
		list.GroupSizes = append(list.GroupSizes, len(frames))
	}

	return list, nil
}

// decodeCLXGroup decodes one CLX frame list.
func decodeCLXGroup(data []byte) ([]ClxFrame, error) {
	numFrames, offsets, err := parseFrameOffsets(data, "CLX")
	if err != nil {
		return nil, err
	}

	frames := make([]ClxFrame, 0, numFrames)
	for i := range numFrames {
		frame, err := decodeClxFrame(data[offsets[i]:offsets[i+1]])
		if err != nil {
			return nil, fmt.Errorf("CLX frame %d: %w", i, err)
		}

		frames = append(frames, *frame)
	}

	return frames, nil
}

// decodeClxFrame decodes one CLX frame.
//
// Frame header is 6 bytes (clx_sprite.hpp:15-27):
//
//	0..2 uint16 header size (pixel data offset within the frame)
//	2..4 uint16 width
//	4..6 uint16 height
//
// Pixel data uses the CL2 RLE command stream
// (https://github.com/savagesteel/d1-file-formats/blob/master/PC-Mac/CL2.md#42-cl2-frame-data).
func decodeClxFrame(frame []byte) (*ClxFrame, error) {
	if len(frame) < 6 {
		return nil, fmt.Errorf("frame too short for CLX header: %d bytes", len(frame))
	}

	headerSize := int(binary.LittleEndian.Uint16(frame[0:2]))
	width := int(binary.LittleEndian.Uint16(frame[2:4]))

	height := int(binary.LittleEndian.Uint16(frame[4:6]))
	if headerSize < 6 || headerSize > len(frame) {
		return nil, fmt.Errorf("invalid CLX frame header size %d (frame %d bytes)", headerSize, len(frame))
	}

	if width < 1 || height < 1 {
		return nil, fmt.Errorf("invalid CLX frame dimensions %dx%d", width, height)
	}

	pixels, err := decodeCl2Pixels(frame[headerSize:], width, height)
	if err != nil {
		return nil, err
	}

	return &ClxFrame{Width: width, Height: height, Pixels: pixels}, nil
}
