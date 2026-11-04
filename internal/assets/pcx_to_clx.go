package assets

import (
	"encoding/binary"
	"fmt"
)

// ConvertPCXToCLX converts a PCX file into CLX bytes, ported 1:1 from
// PcxToClx (utils/pcx_to_clx.cpp:55-190).
//
// A positive numFramesOrFrameHeight is a frame count; a negative value is the
// frame height (pcx_to_clx.cpp:69-75). transparentColor, when given, is dropped
// from the stream and encoded as a CLX transparent run (:136-172).
//
// Rows are stored bottom-up in the CLX frame, and scanlines are decoded on the
// header's BytesPerLine boundary with the padding dropped, like PcxToClx's
// `srcSkip` (:117-134).
func ConvertPCXToCLX(data []byte, numFramesOrFrameHeight int, transparentColor ...byte) ([]byte, error) {
	if len(data) < pcxHeaderSize {
		return nil, fmt.Errorf("PCX: data is %d bytes, shorter than the %d-byte header", len(data), pcxHeaderSize)
	}

	header := parsePCXHeader(data[:pcxHeaderSize])
	if err := validatePCXHeader(header); err != nil {
		return nil, err
	}

	layout := pcxLayout{
		width:        int(header.Xmax) - int(header.Xmin) + 1,
		height:       int(header.Ymax) - int(header.Ymin) + 1,
		bytesPerLine: int(header.BytesPerLine),
	}
	if layout.width <= 0 || layout.height <= 0 {
		return nil, fmt.Errorf("PCX: invalid dimensions %dx%d", layout.width, layout.height)
	}

	numFrames, frameHeight, err := pcxFrameGeometry(layout.height, numFramesOrFrameHeight)
	if err != nil {
		return nil, err
	}

	rows, err := decodePCXRows(data[pcxHeaderSize:], layout)
	if err != nil {
		return nil, fmt.Errorf("PCX: pixel decode: %w", err)
	}

	return encodePCXRows(rows, layout.width, numFrames, frameHeight, transparentColor), nil
}

// pcxFrameGeometry resolves the frame count and height from the caller's
// numFramesOrFrameHeight argument (pcx_to_clx.cpp:69-75).
func pcxFrameGeometry(height, numFramesOrFrameHeight int) (numFrames, frameHeight int, err error) {
	switch {
	case numFramesOrFrameHeight > 0:
		if height%numFramesOrFrameHeight != 0 {
			return 0, 0, fmt.Errorf("PCX: height %d is not a multiple of %d frames",
				height, numFramesOrFrameHeight)
		}

		return numFramesOrFrameHeight, height / numFramesOrFrameHeight, nil
	case numFramesOrFrameHeight < 0:
		frameHeight = -numFramesOrFrameHeight
		if height%frameHeight != 0 {
			return 0, 0, fmt.Errorf("PCX: height %d is not a multiple of frame height %d",
				height, frameHeight)
		}

		return height / frameHeight, frameHeight, nil
	default:
		return 0, 0, fmt.Errorf("PCX: numFramesOrFrameHeight must not be %d", numFramesOrFrameHeight)
	}
}

// encodePCXRows stacks the decoded rows into a CLX sprite list, bottom-up
// (pcx_to_clx.cpp:96-171).
func encodePCXRows(rows [][]byte, width, numFrames, frameHeight int, transparentColor []byte) []byte {
	out := make([]byte, 4*(2+numFrames))
	binary.LittleEndian.PutUint32(out, uint32(numFrames))

	for frame := range numFrames {
		binary.LittleEndian.PutUint32(out[4*(frame+1):], uint32(len(out)))

		headerPos := appendCLXFrameHeader(&out, width)
		binary.LittleEndian.PutUint16(out[headerPos+4:], uint16(frameHeight))

		transparentRunWidth := 0
		frameBase := frame * frameHeight
		for line := range frameHeight {
			row := rows[frameBase+(frameHeight-(line+1))]
			if len(transparentColor) == 0 {
				appendCLXPixelsOrFillRun(&out, row)

				continue
			}
			transparentRunWidth = appendCLXTransparentAwareLine(&out, row, transparentColor[0], transparentRunWidth)
		}
		appendCLXTransparentRun(&out, transparentRunWidth)
	}

	// encodePCXRows cannot fail: the list header is allocated before any frame
	// is appended, so the end-offset slot always exists.
	_ = patchListSize(out, 0, numFrames)

	return out
}
