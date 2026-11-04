package assets

import (
	"encoding/binary"
	"fmt"
)

// cl2Layout is the header geometry of a CL2 file: a single frame list, or a
// group offset table whose entries point at one CL2 frame list each
// (cl2_to_clx.cpp:15-30).
type cl2Layout struct {
	numGroups      int
	frameCount     int
	groupTableSize int
}

// ConvertCL2ToCLX converts a CL2 file into CLX bytes, ported from
// Cl2ToClx (utils/cl2_to_clx.cpp:15-112).
//
// widths holds the frame widths the file does not carry. A single value applies
// to every frame; otherwise one value per frame of the current group is
// required (cl2_to_clx.cpp:58-66).
//
// The return value is a CLX sprite list, or a sprite sheet whose leading
// numGroups uint32 group offsets point at one CLX list per group, matching
// upstream's in-place layout (:28-47).
func ConvertCL2ToCLX(data []byte, widths []int) ([]byte, error) {
	layout, err := parseCL2Layout(data)
	if err != nil {
		return nil, err
	}

	out := make([]byte, layout.groupTableSize)

	// Transient buffer for a contiguous run of non-transparent pixels.
	pixels := make([]byte, 0, 4096)

	for group := range layout.numGroups {
		if err := layout.appendGroup(&out, data, group, widths, &pixels); err != nil {
			return nil, err
		}
	}

	return out, nil
}

// parseCL2Layout reads the leading value as a frame count when the offset after
// the last frame equals the file size, and as a group table size otherwise
// (cl2_to_clx.cpp:19-30).
func parseCL2Layout(data []byte) (cl2Layout, error) {
	maybeNumFrames, err := le32(data, 0)
	if err != nil {
		return cl2Layout{}, fmt.Errorf("CL2: %w", err)
	}
	frameCountOrTable := int(maybeNumFrames)

	lastOffset, offsetErr := le32(data, frameCountOrTable*4+4)
	if offsetErr != nil {
		return cl2Layout{}, fmt.Errorf("CL2: %w", offsetErr)
	}
	if lastOffset == uint32(len(data)) {
		return cl2Layout{numGroups: 1, frameCount: frameCountOrTable}, nil
	}

	numGroups := frameCountOrTable / 4
	if numGroups == 0 || frameCountOrTable > len(data) {
		return cl2Layout{}, fmt.Errorf(
			"CL2: group table of %d bytes does not fit the file",
			frameCountOrTable,
		)
	}

	return cl2Layout{
		numGroups:      numGroups,
		frameCount:     frameCountOrTable,
		groupTableSize: frameCountOrTable,
	}, nil
}

// group returns the offset a group starts at and that group's frame count
// (cl2_to_clx.cpp:36-44).
func (l cl2Layout) group(data []byte, group int) (groupBegin, numFrames int, err error) {
	if l.numGroups == 1 {
		return 0, l.frameCount, nil
	}

	groupOffset, err := le32(data, group*4)
	if err != nil {
		return 0, 0, fmt.Errorf("CL2: %w", err)
	}
	groupFrames, err := le32(data, int(groupOffset))
	if err != nil {
		return 0, 0, fmt.Errorf("CL2 group %d: %w", group, err)
	}

	return int(groupOffset), int(groupFrames), nil
}

// appendGroup writes one group's CLX list at its slot in the group offset
// table (cl2_to_clx.cpp:35-45).
func (l cl2Layout) appendGroup(out *[]byte, data []byte, group int, widths []int, pixels *[]byte) error {
	groupBegin, numFrames, err := l.group(data, group)
	if err != nil {
		return err
	}
	if l.numGroups != 1 {
		binary.LittleEndian.PutUint32((*out)[4*group:], uint32(len(*out)))
	}
	if err := encodeCL2Group(out, data, groupBegin, numFrames, widths, pixels); err != nil {
		return fmt.Errorf("CL2 group %d: %w", group, err)
	}

	return nil
}

// encodeCL2Group writes one CLX list (frame count, offsets, frames)
// (cl2_to_clx.cpp:34-54).
func encodeCL2Group(
	out *[]byte,
	data []byte,
	groupBegin, numFrames int,
	widths []int,
	pixels *[]byte,
) error {
	frameWidths, err := frameWidthList(numFrames, widths)
	if err != nil {
		return err
	}

	listOffset := len(*out)
	*out = append(*out, make([]byte, 4*(2+numFrames))...)
	binary.LittleEndian.PutUint32((*out)[listOffset:], uint32(numFrames))

	for frame := 1; frame <= numFrames; frame++ {
		frameBegin, frameEnd, rangeErr := cl2FrameRange(data, groupBegin, frame)
		if rangeErr != nil {
			return rangeErr
		}

		binary.LittleEndian.PutUint32((*out)[listOffset+4*frame:], uint32(len(*out)-listOffset))
		if err := encodeCL2Frame(out, data, frameBegin, frameEnd, frameWidths[frame-1], pixels); err != nil {
			return fmt.Errorf("frame %d: %w", frame, err)
		}
	}

	return patchListSize(*out, listOffset, numFrames)
}

// cl2FrameRange returns the byte range of one frame inside its group; offset 0
// points at the frame list header (cl2_to_clx.cpp:46-59).
func cl2FrameRange(data []byte, groupBegin, frame int) (frameBegin, frameEnd int, err error) {
	beginOffset, err := le32(data, groupBegin+4*frame)
	if err != nil {
		return 0, 0, err
	}
	endOffset, err := le32(data, groupBegin+4*(frame+1))
	if err != nil {
		return 0, 0, err
	}

	return groupBegin + int(beginOffset), groupBegin + int(endOffset), nil
}

// encodeCL2Frame re-encodes one CL2 frame as a CLX frame
// (cl2_to_clx.cpp:60-107).
func encodeCL2Frame(
	out *[]byte,
	data []byte,
	frameBegin, frameEnd, frameWidth int,
	pixels *[]byte,
) error {
	if frameBegin > frameEnd || frameEnd > len(data) || frameBegin < 0 {
		return fmt.Errorf(
			"frame range %d..%d does not fit %d bytes",
			frameBegin,
			frameEnd,
			len(data),
		)
	}

	headerSize, err := le16(data, frameBegin)
	if err != nil {
		return err
	}
	src := frameBegin + int(headerSize)
	if src > frameEnd {
		return fmt.Errorf("frame header of %d bytes overruns the frame", headerSize)
	}

	headerPos := appendCLXFrameHeader(out, frameWidth)
	frameHeight, err := encodeCL2Lines(out, data, src, frameEnd, frameWidth, pixels)
	if err != nil {
		return err
	}

	if len(*pixels) > 0 {
		appendCLXPixelsOrFillRun(out, *pixels)
		*pixels = (*pixels)[:0]
	}
	binary.LittleEndian.PutUint16((*out)[headerPos+4:], uint16(frameHeight))

	return nil
}

// encodeCL2Lines re-encodes the command stream up to frameEnd, tracking the
// overrun into the next line the way GetSkipSize does
// (cl2_to_clx.cpp:68-100).
func encodeCL2Lines(out *[]byte, data []byte, src, frameEnd, frameWidth int,
	pixels *[]byte,
) (frameHeight int, err error) {
	transparentRunWidth := 0
	xOffset := 0

	for src != frameEnd {
		var overrun int
		src, overrun, err = encodeCL2Line(out, data, src, frameEnd, frameWidth, xOffset, pixels,
			&transparentRunWidth)
		if err != nil {
			return 0, err
		}

		lines, newXOffset := skipSize(overrun, frameWidth)
		xOffset = newXOffset
		frameHeight += lines
	}

	appendCLXTransparentRun(out, transparentRunWidth)

	return frameHeight, nil
}

// encodeCL2Line consumes commands until the current line is filled and reports
// the source offset, the line overrun (0 or negative) and the open transparent
// run.
func encodeCL2Line(out *[]byte, data []byte, src, frameEnd, frameWidth, xOffset int, pixels *[]byte,
	transparentRunWidth *int,
) (next, overrun int, err error) {
	remaining := frameWidth - xOffset

	for remaining > 0 {
		if src >= frameEnd {
			return src, remaining,
				fmt.Errorf("command stream ends at byte %d before the line is filled", src)
		}
		control := data[src]
		src++

		src, remaining, err = encodeCL2Command(out, data, src, frameEnd, control, remaining, pixels,
			transparentRunWidth)
		if err != nil {
			return src, remaining, err
		}
	}

	return src, remaining, nil
}

// encodeCL2Command applies one CL2 command to the pending runs and reports the
// new source offset and remaining line width (cl2_to_clx.cpp:73-93).
func encodeCL2Command(out *[]byte, data []byte, src, frameEnd int, control byte, remaining int,
	pixels *[]byte, transparentRunWidth *int,
) (next, remainingNow int, err error) {
	switch {
	case control < clxOpaqueMin:
		return applyCL2TransparentCommand(out, src, control, remaining, pixels, transparentRunWidth)
	case control <= clxOpaqueFillMax:
		return applyCL2FillCommand(
			out,
			data,
			src,
			frameEnd,
			control,
			remaining,
			pixels,
			transparentRunWidth,
		)
	default:
		return applyCL2LiteralCommand(
			out,
			data,
			src,
			frameEnd,
			control,
			remaining,
			pixels,
			transparentRunWidth,
		)
	}
}

// applyCL2TransparentCommand starts a transparent run (cl2_to_clx.cpp:76-83).
func applyCL2TransparentCommand(out *[]byte, src int, control byte, remaining int,
	pixels *[]byte, transparentRunWidth *int,
) (next, remainingNow int, err error) {
	if len(*pixels) > 0 {
		appendCLXPixelsOrFillRun(out, *pixels)
		*pixels = (*pixels)[:0]
	}
	*transparentRunWidth += int(control)

	return src, remaining - int(control), nil
}

// applyCL2FillCommand expands a fill command into pixels (cl2_to_clx.cpp:84-91).
func applyCL2FillCommand(out *[]byte, data []byte, src, frameEnd int, control byte, remaining int,
	pixels *[]byte, transparentRunWidth *int,
) (next, remainingNow int, err error) {
	appendCLXTransparentRun(out, *transparentRunWidth)
	*transparentRunWidth = 0

	width := int(clxOpaqueFillEnd) - int(control)
	if src >= frameEnd {
		return src, remaining, fmt.Errorf("fill run at byte %d is truncated", src)
	}
	color := data[src]
	for range width {
		*pixels = append(*pixels, color)
	}

	return src + 1, remaining - width, nil
}

// applyCL2LiteralCommand copies literal pixels (cl2_to_clx.cpp:92-99).
func applyCL2LiteralCommand(
	out *[]byte,
	data []byte,
	src, frameEnd int,
	control byte,
	remaining int,
	pixels *[]byte,
	transparentRunWidth *int,
) (next, remainingNow int, err error) {
	appendCLXTransparentRun(out, *transparentRunWidth)
	*transparentRunWidth = 0

	width := 256 - int(control)
	if src+width > frameEnd {
		return src, remaining, fmt.Errorf("literal run of %d at byte %d is truncated", width, src)
	}
	*pixels = append(*pixels, data[src:src+width]...)

	return src + width, remaining - width, nil
}

// patchListSize writes the list end offset that closes a CLX list header.
func patchListSize(out []byte, listOffset, numFrames int) error {
	pos := listOffset + 4*(1+numFrames)
	if pos+4 > len(out) {
		return fmt.Errorf("list at %d has no room for its end offset", listOffset)
	}
	binary.LittleEndian.PutUint32(out[pos:], uint32(len(out)-listOffset))

	return nil
}

// frameWidthList resolves the width of every frame: one value for all frames or
// exactly one per frame (cl2_to_clx.cpp:66).
func frameWidthList(numFrames int, widths []int) ([]int, error) {
	switch {
	case len(widths) == 1:
		resolved := make([]int, numFrames)
		for i := range resolved {
			resolved[i] = widths[0]
		}

		return resolved, nil
	case len(widths) == numFrames:
		return widths, nil
	default:
		return nil, fmt.Errorf("%d widths for %d frames", len(widths), numFrames)
	}
}

// le16 reads a little-endian uint16 at pos.
func le16(data []byte, pos int) (uint16, error) {
	if pos < 0 || pos+2 > len(data) {
		return 0, fmt.Errorf("uint16 at byte %d is past the end of %d bytes", pos, len(data))
	}

	return binary.LittleEndian.Uint16(data[pos:]), nil
}

// le32 reads a little-endian uint32 at pos.
func le32(data []byte, pos int) (uint32, error) {
	if pos < 0 || pos+4 > len(data) {
		return 0, fmt.Errorf("uint32 at byte %d is past the end of %d bytes", pos, len(data))
	}

	return binary.LittleEndian.Uint32(data[pos:]), nil
}
