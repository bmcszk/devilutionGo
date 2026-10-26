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

// CelFrame is one decoded CEL sprite frame as 256-colour indexed pixels.
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
	if len(data) < 4 {
		return nil, fmt.Errorf("CEL data too short for frame count")
	}

	// First 4 bytes: number of frames or pointer to group offsets
	maybeNumFrames := int(binary.LittleEndian.Uint32(data[0:4]))

	// Check if this is a multi-group file (group offsets table)
	// In devilutionX: if LoadLE32(data[(maybeNumFrames*4)+4]) != size, it's a multi-group file
	var groups [][]byte
	if maybeNumFrames > 0 && len(data) >= 4+maybeNumFrames*4 {
		lastOffset := int(binary.LittleEndian.Uint32(data[4+maybeNumFrames*4 : 8+maybeNumFrames*4]))
		if lastOffset != len(data) {
			// Multi-group file: maybeNumFrames is the address of first group
			numGroups := maybeNumFrames / 4
			groupOffsets := make([]uint32, numGroups+1)
			for i := range groupOffsets {
				groupOffsets[i] = binary.LittleEndian.Uint32(data[4+4*i : 8+4*i])
			}
			groupOffsets[numGroups] = uint32(len(data))

			for i := 0; i < int(numGroups); i++ {
				start := groupOffsets[i]
				end := groupOffsets[i+1]
				if start < end && end <= uint32(len(data)) {
					groups = append(groups, data[start:end])
				}
			}
		} else {
			// Single group
			groups = append(groups, data)
		}
	} else {
		groups = append(groups, data)
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

// decodeCelGroup decodes a single CEL group (which uses CL2-like encoding after conversion)
func decodeCelGroup(data []byte) ([]CelFrame, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("CEL group data too short for frame count")
	}

	numFrames := int(binary.LittleEndian.Uint32(data[0:4]))
	if numFrames < 1 {
		return nil, fmt.Errorf("CEL: invalid frame count %d", numFrames)
	}

	need := 4 + 4*(numFrames+1)
	if len(data) < need {
		return nil, fmt.Errorf("CEL: data too short for %d frame offsets (need %d, have %d)", numFrames+1, need, len(data))
	}

	offsets := make([]uint32, numFrames+1)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(data[4+4*i : 8+4*i])
	}

	// List-vs-sheet disambiguation: last offset should equal data size for single list
	// Multi-frame list: last offset is end of last frame, not end of data
	// This is normal for multi-frame CEL files

	frames := make([]CelFrame, 0, numFrames)
	for i := 0; i < numFrames; i++ {
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
	if len(data) < 6 {
		return nil, fmt.Errorf("CEL frame header too short (need 6, have %d)", len(data))
	}

	headerSize := int(data[0])
	if headerSize != 6 {
		return nil, fmt.Errorf("CEL frame header size mismatch (expected 6, got %d)", headerSize)
	}
	width := int(binary.LittleEndian.Uint16(data[2:4]))
	height := int(binary.LittleEndian.Uint16(data[4:6]))

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("CEL: invalid dimensions %dx%d", width, height)
	}

	nominalPixels := width * height
	pixels := make([]byte, 0, nominalPixels)

	for i := 6; i < len(data); {
		ctrl := data[i]
		i++
		if ctrl < 0x80 {
			// Literal run (control byte = count-1)
			count := int(ctrl) + 1
			if count > len(data)-i {
				return nil, fmt.Errorf("CEL literal run overflow: need %d, have %d", count, len(data)-i)
			}
			for j := 0; j < count; j++ {
				pixels = append(pixels, data[i+j])
			}
			i += count
		} else {
			// Transparent run
			run := int(-int8(ctrl))
			for j := 0; j < run; j++ {
				pixels = append(pixels, 0)
			}
		}
	}

	if len(pixels) != nominalPixels {
		return nil, fmt.Errorf("CEL: pixel count mismatch (got %d, expected %d)", len(pixels), nominalPixels)
	}

	return &CelFrame{Width: width, Height: height, Pixels: pixels}, nil
}