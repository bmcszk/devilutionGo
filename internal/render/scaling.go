package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// WorldLogicalWidth is the world logical surface width: upstream renders the
// world at 640x480 and scales the finished frame to the window
// (DEVILUTIONX_DEFAULT_RESOLUTION, Source/utils/config.h; the UI rectangle is
// pinned to the same 640x480, Source/utils/display.cpp:140).
const WorldLogicalWidth = 640

// WorldLogicalHeight is the world logical surface height.
const WorldLogicalHeight = 480

// ScaleToWindow is upstream's integer-scaling window presentation
// (Source/utils/display.cpp:121-126 CalculatePreferredWindowSize with
// Graphics.integerScaling, presented through SDL_LOGICAL_PRESENTATION_INTEGER_SCALE,
// display.cpp:780-811): the last presented 640x480 logical frame is scaled by
// the largest integer factor that fits the width x height window and centred
// in a letterbox. Windows smaller than the logical surface fall back to
// factor 1 — upstream forbids sub-640x480 resolutions outright
// (display.cpp:291). There is no fractional factor and no filtering
// (scaling mode NearestPixel, display.cpp:694-713).
func (p *Presenter) ScaleToWindow(target *ebiten.Image, width, height int) {
	if p.target == nil {
		return
	}

	factor := IntegerScaleFactor(width, height, p.width, p.height)

	target.Clear()

	options := &ebiten.DrawImageOptions{}
	options.GeoM.Scale(float64(factor), float64(factor))
	options.GeoM.Translate(float64((width-p.width*factor)/2), float64((height-p.height*factor)/2))
	options.Filter = ebiten.FilterNearest
	target.DrawImage(p.target, options)
}

// IntegerScaleFactor is the letterboxed integer factor: the floored minimum of
// the per-axis ratios (display.cpp:121-125), clamped to at least 1 (the
// sub-640x480 floor of display.cpp:291).
func IntegerScaleFactor(windowWidth, windowHeight, frameWidth, frameHeight int) int {
	if frameWidth <= 0 || frameHeight <= 0 {
		return 1
	}

	factor := min(windowWidth/frameWidth, windowHeight/frameHeight)

	return max(factor, 1)
}

// ScaledFrameRect is the on-window rectangle the logical frame occupies after
// integer scaling: the scaled size centred in the window, the letterbox model
// of SDL_LOGICAL_PRESENTATION_INTEGER_SCALE (display.cpp:780-811).
func ScaledFrameRect(windowWidth, windowHeight, frameWidth, frameHeight int) image.Rectangle {
	factor := IntegerScaleFactor(windowWidth, windowHeight, frameWidth, frameHeight)

	scaledW, scaledH := frameWidth*factor, frameHeight*factor

	return image.Rect((windowWidth-scaledW)/2, (windowHeight-scaledH)/2,
		(windowWidth-scaledW)/2+scaledW, (windowHeight-scaledH)/2+scaledH)
}
