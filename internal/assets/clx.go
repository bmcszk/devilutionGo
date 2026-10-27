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
}

// DecodeCLX parses a CLX sprite list and decodes every frame to indexed
// pixels. It does not support CLX sheets (multi-list files); devilutionX
// disambiguates list vs sheet via the last frame offset
// (clx_sprite.hpp:528-535 GetNumListsFromClxListOrSheetBuffer) — a buffer
// whose last offset does not equal the data size is a sheet and is rejected
// here until sheets are needed.
func DecodeCLX(data []byte) (*ClxSpriteList, error) {
	numFrames, offsets, err := parseFrameOffsets(data, "CLX")
	if err != nil {
		return nil, err
	}

	list := &ClxSpriteList{Frames: make([]ClxFrame, 0, numFrames)}
	for i := range numFrames {
		frame, err := decodeClxFrame(data[offsets[i]:offsets[i+1]])
		if err != nil {
			return nil, fmt.Errorf("CLX frame %d: %w", i, err)
		}

		list.Frames = append(list.Frames, *frame)
	}

	return list, nil
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
