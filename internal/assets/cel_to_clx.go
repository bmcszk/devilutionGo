package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// celLayout is the header geometry of a CEL file: a plain frame list, or a
// group offset table whose entries point at one CEL frame list each
// (cel_to_clx.cpp:36-58).
type celLayout struct {
	numGroups      int
	frameCount     int
	groupTableSize int
}

// ConvertCELToCLX converts a CEL file into CLX bytes, ported 1:1 from
// CelToClx (utils/cel_to_clx.cpp:35-127).
//
// widths holds the frame widths the CEL headers do not carry: a single value
// applies to every frame, otherwise one value per frame is required
// (cel_to_clx.cpp:86).
//
// The return value is a CLX sprite list, or a sprite sheet whose leading
// numGroups uint32 group offsets point at one CLX list per group, matching
// upstream's in-place layout (cel_to_clx.cpp:51-58, :126).
func ConvertCELToCLX(data []byte, widths []int) ([]byte, error) {
	layout, err := parseCELLayout(data)
	if err != nil {
		return nil, err
	}

	out := make([]byte, layout.groupTableSize)

	for group := range layout.numGroups {
		if err := layout.appendGroup(&out, data, group, widths); err != nil {
			return nil, err
		}
	}

	return out, nil
}

// parseCELLayout reads the leading value as a frame count when the offset after
// the last frame equals the file size, and as a group table size otherwise
// (cel_to_clx.cpp:39-58).
func parseCELLayout(data []byte) (celLayout, error) {
	maybeNumFrames, err := le32(data, 0)
	if err != nil {
		return celLayout{}, fmt.Errorf("CEL: %w", err)
	}
	frameCountOrTable := int(maybeNumFrames)

	lastOffset, offsetErr := le32(data, frameCountOrTable*4+4)
	if offsetErr != nil {
		return celLayout{}, fmt.Errorf("CEL: %w", offsetErr)
	}
	if int(lastOffset) == len(data) {
		return celLayout{numGroups: 1, frameCount: frameCountOrTable}, nil
	}

	numGroups := frameCountOrTable / 4
	if numGroups == 0 || frameCountOrTable > len(data) {
		return celLayout{}, fmt.Errorf(
			"CEL: group table of %d bytes does not fit the file",
			frameCountOrTable,
		)
	}

	return celLayout{
		numGroups:      numGroups,
		frameCount:     frameCountOrTable,
		groupTableSize: frameCountOrTable,
	}, nil
}

// group returns the frame list a group offset points at and its frame count
// (cel_to_clx.cpp:44-50).
func (l celLayout) group(data []byte, group int) ([]byte, int, error) {
	if l.numGroups == 1 {
		return data, l.frameCount, nil
	}

	groupOffset, err := le32(data, group*4)
	if err != nil {
		return nil, 0, fmt.Errorf("CEL: %w", err)
	}
	if int(groupOffset) > len(data) {
		return nil, 0, fmt.Errorf("CEL group %d: offset %d is past the end of %d bytes",
			group, groupOffset, len(data))
	}
	groupData := data[groupOffset:]

	numFrames, err := le32(groupData, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("CEL group %d: %w", group, err)
	}

	return groupData, int(numFrames), nil
}

// appendGroup writes one group's CLX list at its slot in the group offset
// table (cel_to_clx.cpp:44-126).
func (l celLayout) appendGroup(out *[]byte, data []byte, group int, widths []int) error {
	groupData, numFrames, err := l.group(data, group)
	if err != nil {
		return err
	}
	if l.numGroups != 1 {
		binary.LittleEndian.PutUint32((*out)[4*group:], uint32(len(*out)))
	}
	if err := encodeCELGroup(out, groupData, numFrames, widths); err != nil {
		return fmt.Errorf("CEL group %d: %w", group, err)
	}

	return nil
}

// encodeCELGroup writes one CLX list (frame count, offsets, frames)
// (cel_to_clx.cpp:64-124).
func encodeCELGroup(out *[]byte, data []byte, numFrames int, widths []int) error {
	frameWidths, err := frameWidthList(numFrames, widths)
	if err != nil {
		return err
	}

	listOffset := len(*out)
	*out = append(*out, make([]byte, 4*(2+numFrames))...)
	binary.LittleEndian.PutUint32((*out)[listOffset:], uint32(numFrames))

	for frame := 1; frame <= numFrames; frame++ {
		frameBegin, frameEnd, rangeErr := celFrameRange(data, frame)
		if rangeErr != nil {
			return rangeErr
		}

		binary.LittleEndian.PutUint32((*out)[listOffset+4*frame:], uint32(len(*out)-listOffset))
		if err := encodeCELFrame(out, data, frameBegin, frameEnd, frameWidths[frame-1]); err != nil {
			return fmt.Errorf("frame %d: %w", frame, err)
		}
	}

	return patchListSize(*out, listOffset, numFrames)
}

// celFrameRange returns the byte range of one frame inside its group; offset 0
// points at the frame list header (cel_to_clx.cpp:68-70).
func celFrameRange(data []byte, frame int) (frameBegin, frameEnd int, err error) {
	beginOffset, err := le32(data, 4*frame)
	if err != nil {
		return 0, 0, err
	}
	endOffset, err := le32(data, 4*(frame+1))
	if err != nil {
		return 0, 0, err
	}

	return int(beginOffset), int(endOffset), nil
}

// encodeCELFrame re-encodes one CEL frame as a CLX frame (cel_to_clx.cpp:70-121).
func encodeCELFrame(out *[]byte, data []byte, frameBegin, frameEnd, frameWidth int) error {
	if frameBegin < 0 || frameBegin > frameEnd || frameEnd > len(data) {
		return fmt.Errorf(
			"frame range %d..%d does not fit %d bytes",
			frameBegin,
			frameEnd,
			len(data),
		)
	}

	src := frameBegin
	// A CEL frame carries the 10-byte header only when its first uint16 says
	// so; the fields are unused (cel_to_clx.cpp:76-84).
	if headerSize, err := le16(data, src); err == nil && int(headerSize) == celFrameHeaderSize {
		src += celFrameHeaderSize
	}
	if src > frameEnd {
		return errors.New("CEL frame header overruns the frame")
	}

	headerPos := appendCLXFrameHeader(out, frameWidth)
	frameHeight, transparentRunWidth, err := encodeCELLines(out, data, src, frameEnd, frameWidth)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint16((*out)[headerPos+4:], uint16(frameHeight))
	appendCLXTransparentRun(out, transparentRunWidth)

	return nil
}

// encodeCELLines re-encodes the command stream up to frameEnd, one line at a
// time, and reports the frame height and the open transparent run
// (cel_to_clx.cpp:86-119).
func encodeCELLines(
	out *[]byte,
	data []byte,
	src, frameEnd, frameWidth int,
) (frameHeight, transparentRunWidth int, err error) {
	for src != frameEnd {
		src, transparentRunWidth, err = encodeCELLine(
			out,
			data,
			src,
			frameEnd,
			frameWidth,
			transparentRunWidth,
		)
		if err != nil {
			return 0, 0, err
		}
		frameHeight++
	}

	return frameHeight, transparentRunWidth, nil
}

// encodeCELLine decodes exactly one pixel line and re-encodes it as CLX runs
// (cel_to_clx.cpp:87-114). The caller guarantees that the runs fill the line.
//
// Deviations from upstream, which has no bounds checks: a command stream that
// ends before the line is filled or a run that overruns the frame is an error
// instead of an out-of-bounds read.
func encodeCELLine(
	out *[]byte,
	data []byte,
	src, frameEnd, frameWidth, transparentRunWidth int,
) (next, newTransparentRunWidth int, err error) {
	remaining := frameWidth

	for remaining > 0 {
		if src >= frameEnd {
			return src, transparentRunWidth,
				fmt.Errorf("command stream ends at byte %d before the line is filled", src)
		}
		control := data[src]
		src++

		if control >= celTransparentMin {
			width := 256 - int(control)
			transparentRunWidth += width
			remaining -= width

			continue
		}

		appendCLXTransparentRun(out, transparentRunWidth)
		transparentRunWidth = 0
		width := int(control)
		if src+width > frameEnd {
			return src, transparentRunWidth,
				fmt.Errorf("pixel run of %d at byte %d overruns the frame", width, src)
		}
		appendCLXPixelsOrFillRun(out, data[src:src+width])
		src += width
		remaining -= width
	}

	return src, transparentRunWidth, nil
}
