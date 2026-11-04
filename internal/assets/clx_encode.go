package assets

import (
	"encoding/binary"
	"fmt"
)

// CLX run encoding limits (clx_encode.hpp:16-63). A control byte below 0x80 is
// a transparent run, 0x80..0xBE a fill run, 0xBF a 65-pixel literal run and
// 0xC0..0xFF a shorter literal run.
const (
	// clxFrameHeaderSize is the CLX per-frame header: header size, width,
	// height as three little-endian uint16s (clx_encode.hpp:6-16).
	clxFrameHeaderSize = 6

	// clxOpaqueMin is the first control byte that means an opaque run
	// (clx_decode.hpp:11-15).
	clxOpaqueMin            = 0x80
	clxTransparentSingleRun = 0x7F
	clxFillSingleRun        = 0x3F
	clxLiteralSingleRun     = 0x41
	// clxOpaqueFillMax is the largest fill control byte; the fill width is
	// clxOpaqueFillEnd - control (clx_decode.hpp:32-41).
	clxOpaqueFillMax = 0xBE
	clxOpaqueFillEnd = 0xBF
	// minFillRunLength is the tunable fill-run threshold: "3 appears to be
	// optimal for most of our data" (clx_encode.hpp:78-80).
	minFillRunLength = 3
)

// EncodeCLXSpriteList builds a CLX sprite list from a stack of frames, exactly
// like SurfaceToClx (surface_to_clx.cpp:20-80): frames are stacked vertically in
// pixels, each frame's rows are stored bottom-up, and the optional transparent
// palette index becomes a transparent run.
//
// pixels holds width x (frameHeight*numFrames) palette indices, top row first
// (surface order). Every frame is frameHeight = height/numFrames rows tall.
func EncodeCLXSpriteList(pixels []byte, width, numFrames int, transparentColor ...byte) ([]byte, error) {
	if width <= 0 || numFrames <= 0 {
		return nil, fmt.Errorf("encode CLX: invalid geometry %dx%d frames", width, numFrames)
	}
	if len(pixels)%width != 0 {
		return nil, fmt.Errorf("encode CLX: %d pixels is not a multiple of width %d", len(pixels), width)
	}

	height := len(pixels) / width
	if height%numFrames != 0 {
		return nil, fmt.Errorf("encode CLX: height %d is not a multiple of frames %d", height, numFrames)
	}
	frameHeight := height / numFrames

	out := make([]byte, 4*(2+numFrames))
	binary.LittleEndian.PutUint32(out, uint32(numFrames))

	for frame := range numFrames {
		encodeCLXFrame(&out, pixels, width, frameHeight, frame, transparentColor)
	}
	binary.LittleEndian.PutUint32(out[4*(1+numFrames):], uint32(len(out)))

	return out, nil
}

// encodeCLXFrame appends one frame header plus its bottom-up command stream
// (surface_to_clx.cpp:40-77).
func encodeCLXFrame(out *[]byte, pixels []byte, width, frameHeight, frame int, transparentColor []byte) {
	binary.LittleEndian.PutUint32((*out)[4*(frame+1):], uint32(len(*out)))

	headerPos := appendCLXFrameHeader(out, width)
	binary.LittleEndian.PutUint16((*out)[headerPos+4:], uint16(frameHeight))

	transparentRunWidth := 0
	frameBase := frame * frameHeight * width
	for line := range frameHeight {
		rowStart := frameBase + (frameHeight-(line+1))*width
		row := pixels[rowStart : rowStart+width]
		if len(transparentColor) == 0 {
			appendCLXPixelsOrFillRun(out, row)

			continue
		}
		transparentRunWidth = appendCLXTransparentAwareLine(out, row, transparentColor[0], transparentRunWidth)
	}
	appendCLXTransparentRun(out, transparentRunWidth)
}

// appendCLXFrameHeader appends a CLX frame header (header size, width, reserved
// height) and returns the position it was written at (clx_encode.hpp:6-16).
func appendCLXFrameHeader(out *[]byte, width int) int {
	headerPos := len(*out)
	*out = append(*out, 0, 0, 0, 0, 0, 0)
	binary.LittleEndian.PutUint16((*out)[headerPos:], clxFrameHeaderSize)
	binary.LittleEndian.PutUint16((*out)[headerPos+2:], uint16(width))

	return headerPos
}

// appendCLXTransparentAwareLine appends one line that may contain the
// transparent index, carrying the open transparent run into the next line
// (surface_to_clx.cpp:60-74). The solid-run and transparent-run flushes are
// no-ops for empty runs, so the boundary between them needs no guard.
func appendCLXTransparentAwareLine(out *[]byte, row []byte, transparent byte, transparentRunWidth int) int {
	solidStart := 0
	solidRunWidth := 0

	for x, value := range row {
		if value == transparent {
			appendCLXPixelsOrFillRun(out, row[solidStart:solidStart+solidRunWidth])
			solidRunWidth = 0
			transparentRunWidth++

			continue
		}

		appendCLXTransparentRun(out, transparentRunWidth)
		transparentRunWidth = 0
		if solidRunWidth == 0 {
			solidStart = x
		}
		solidRunWidth++
	}

	appendCLXPixelsOrFillRun(out, row[solidStart:solidStart+solidRunWidth])

	return transparentRunWidth
}

// appendCLXTransparentRun emits a transparent run of width pixels
// (clx_encode.hpp:20-28).
func appendCLXTransparentRun(out *[]byte, width int) {
	for width >= clxTransparentSingleRun {
		*out = append(*out, clxTransparentSingleRun)
		width -= clxTransparentSingleRun
	}
	if width == 0 {
		return
	}
	*out = append(*out, byte(width))
}

// appendCLXFillRun emits a run of width copies of color (clx_encode.hpp:30-40).
func appendCLXFillRun(out *[]byte, color byte, width int) {
	for width >= clxFillSingleRun {
		*out = append(*out, 0x80, color)
		width -= clxFillSingleRun
	}
	if width == 0 {
		return
	}
	*out = append(*out, clxOpaqueFillEnd-byte(width), color)
}

// appendCLXPixelsRun emits literal pixels (clx_encode.hpp:42-56).
func appendCLXPixelsRun(out *[]byte, src []byte) {
	for len(src) >= clxLiteralSingleRun {
		*out = append(*out, 0xBF)
		*out = append(*out, src[:clxLiteralSingleRun]...)
		src = src[clxLiteralSingleRun:]
	}
	if len(src) == 0 {
		return
	}
	*out = append(*out, byte(256-len(src)))
	*out = append(*out, src...)
}

// appendCLXPixelsOrFillRun encodes a solid run, choosing fill runs for repeated
// colours (clx_encode.hpp:58-100). The caller passes exactly the run it wants;
// width above 4 is the same variable upstream calls `length`.
func appendCLXPixelsOrFillRun(out *[]byte, src []byte) {
	if len(src) == 0 {
		return
	}

	begin := 0
	prevColorBegin := 0
	prevColorRunLength := 1
	prevColor := src[0]

	for i := 1; i < len(src); i++ {
		color := src[i]
		if prevColor == color {
			prevColorRunLength++

			continue
		}
		if prevColorRunLength >= minFillRunLength {
			appendCLXPixelsRun(out, src[begin:prevColorBegin])
			appendCLXFillRun(out, prevColor, prevColorRunLength)
			begin = i
		}
		prevColorBegin = i
		prevColorRunLength = 1
		prevColor = color
	}

	// Here 2 instead of minFillRunLength: this run is followed by transparent
	// pixels, where a width-2 fill command costs 2 bytes and a pixels command 3
	// (clx_encode.hpp:92-95).
	if prevColorRunLength >= 2 {
		appendCLXPixelsRun(out, src[begin:prevColorBegin])
		appendCLXFillRun(out, prevColor, prevColorRunLength)

		return
	}
	// begin, not prevColorBegin: everything from the last flush up to the end
	// is still pending (clx_encode.hpp:98-99).
	appendCLXPixelsRun(out, src[begin:])
}

// skipSize returns how many whole lines a run overran and the x offset it
// landed on (clx_decode.hpp:46-60). Requires remainingWidth <= 0.
func skipSize(remainingWidth, srcWidth int) (wholeLines, xOffset int) {
	if remainingWidth == 0 {
		return 1, 0
	}

	overrun := -remainingWidth

	return overrun/srcWidth + 1, overrun % srcWidth
}
