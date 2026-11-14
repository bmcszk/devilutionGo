package render

import "github.com/bmcszk/devilutionGo/internal/assets"

// transparent is the expanded value of a keyed-out pixel.
var transparent RGBA

// Surface is an index buffer: one palette index per pixel, row-major, no padding.
// Upstream's 8-bit PalSurface has the same shape (dx.cpp:58).
type Surface struct {
	Width  int
	Height int
	Pixels []uint8
}

// Frame is a palette-expanded pixel buffer, the output of the final lookup pass.
type Frame struct {
	Width  int
	Height int
	Pixels []RGBA
}

// RGBA is one expanded pixel.
type RGBA struct {
	R, G, B, A uint8
}

// Clip is a sprite draw positioned on a surface: the sprite's decoded pixels laid
// over the surface at (X, Y), the fields the CLX rasteriser works with
// (clx_render.cpp:63-73).
type Clip struct {
	X, Y   int
	Width  int
	Height int
	// Pixels is the decoded sprite, Width*Height palette indices, row-major.
	Pixels []uint8
	// Light is the per-pixel light level buffer, Width*Height entries, or nil for
	// a fully lit draw.
	Light []uint8
}

// visibleRect is the part of a sprite that lands on the surface.
type visibleRect struct {
	x, y          int // first visible sprite column and row
	width, height int
}

// blitFunc is one blit flavour: it writes dst from src for length pixels, taking
// each pixel's light level from light.
type blitFunc func(dst, src, light []uint8)

// NewSurface allocates a width×height index buffer.
func NewSurface(width, height int) Surface {
	return Surface{Width: width, Height: height, Pixels: make([]uint8, width*height)}
}

// At returns the index at (x, y).
func (s Surface) At(x, y int) uint8 { return s.Pixels[y*s.Width+x] }

// Set writes the index at (x, y).
func (s Surface) Set(x, y int, index uint8) { s.Pixels[y*s.Width+x] = index }

// DrawDirect copies a sprite onto the surface unchanged, index 0 included
// (BlitDirect, blit_impl.hpp:28-42, reached from ClxDraw, clx_render.cpp:589).
func (s Surface) DrawDirect(clip Clip) { s.draw(clip, blitDirect) }

// DrawWithMap copies a sprite onto the surface through a colour map
// (BlitWithMap, blit_impl.hpp:52-66; ClxDrawTRN, clx_render.cpp:594).
func (s Surface) DrawWithMap(clip Clip, trn assets.TRN) { s.draw(clip, mapBlit(trn)) }

// DrawBlended blends a sprite 50% into what is already on the surface
// (BlitBlended, blit_impl.hpp:99-113; ClxDrawBlended).
func (s Surface) DrawBlended(clip Clip, lut BlendLUT) { s.draw(clip, blendedBlit(lut)) }

// DrawBlendedWithMap applies a colour map to the sprite, then blends it
// (BlitBlendedWithMap, blit_impl.hpp:115-129; ClxDrawBlendedTRN).
func (s Surface) DrawBlendedWithMap(clip Clip, trn assets.TRN, lut BlendLUT) {
	s.draw(clip, blendedMapBlit(trn, lut))
}

// DrawLightmap writes a sprite through the per-pixel light tables
// (BlitWithLightmap, blit_impl.hpp:78-92; ClxDrawWithLightmap). Both the sprite
// index and the result stay in palette-index space: lightTables[level][index] is
// another palette index (Lightmap::adjustColor, light_render.hpp:30-33).
func (s Surface) DrawLightmap(clip Clip, tables *[numShades][paletteSize]uint8) {
	s.draw(clip, lightmapBlit(tables))
}

// DrawBlendedWithLightmap writes a lit sprite blended into what is already on the
// surface: each pixel is lit through its light level's table, then blended with the
// destination through the LUT (BlitPixelsBlendedWithLightmap, blit_impl.hpp:153-176;
// ClxDrawLightBlended).
func (s Surface) DrawBlendedWithLightmap(
	clip Clip,
	tables *[numShades][paletteSize]uint8,
	lut BlendLUT,
) {
	s.draw(clip, blendedLightmapBlit(tables, lut))
}

// ExpandSprite expands the surface through a light variant row for the palette
// pass, keying index 0 out: sprite RLE never emits its holes, so index 0 reaching
// the pass is a hole (docs/research/dg-graphics.md G4 keying trap).
func (s Surface) ExpandSprite(variants *Variants, shade int) Frame {
	return s.expand(variants, shade, keyZero)
}

// ExpandTile expands the surface through a light variant row for the palette pass
// without keying: dungeon micro-tiles write every pixel of their shape, index 0
// included, and 0 there is legitimate black (same G4 trap, other side).
func (s Surface) ExpandTile(variants *Variants, shade int) Frame {
	return s.expand(variants, shade, keepZero)
}

// draw positions a sprite on the surface and applies the flavour row by row
// (DoRender, clx_render.cpp:230-260; one byte per pixel in an index buffer,
// clx_render.cpp:44).
func (s Surface) draw(clip Clip, blit blitFunc) {
	visible := s.visible(clip)
	for y := visible.y; y < visible.y+visible.height; y++ {
		offset := y * clip.Width
		src := clip.Pixels[offset+visible.x : offset+visible.x+visible.width]
		blit(s.row(visible.x+clip.X, y+clip.Y, visible.width), src, lightRow(clip, y, visible))
	}
}

// visible clips a clip to the surface on all four sides (ClipX,
// clx_render.cpp:63-73, plus the height clamp in DoRender).
func (s Surface) visible(clip Clip) visibleRect {
	rows := len(clip.Pixels) / max(clip.Width, 1)

	left, right := max(0, -clip.X), min(clip.Width, s.Width-clip.X)
	top, bottom := max(0, -clip.Y), min(clip.Height, s.Height-clip.Y, rows)

	return visibleRect{
		x:      left,
		y:      top,
		width:  max(0, right-left),
		height: max(0, bottom-top),
	}
}

// expand maps every index to a colour through the variant row for shade.
func (s Surface) expand(variants *Variants, shade int, keying keying) Frame {
	if shade < 0 || shade >= numShades {
		shade = fullyLitShade
	}

	frame := Frame{Width: s.Width, Height: s.Height, Pixels: make([]RGBA, len(s.Pixels))}

	for i, index := range s.Pixels {
		if keying == keyZero && index == 0 {
			frame.Pixels[i] = transparent

			continue
		}

		color := variants[shade][index]
		frame.Pixels[i] = RGBA{R: color[0], G: color[1], B: color[2], A: 0xFF}
	}

	return frame
}

// row returns the buffer slice covering x..x+length on row y; y below the
// surface clips to the top row, the DoRenderBackwards overrun guard
// (clx_render.cpp:113 dst >= dstBegin).
func (s Surface) row(x, y, length int) []uint8 {
	offset := max(y, 0)*s.Width + x

	return s.Pixels[offset : offset+length]
}

// lightRow returns the light levels covering the visible part of a sprite row, or
// nil for a fully lit draw.
func lightRow(clip Clip, y int, rect visibleRect) []uint8 {
	offset := y*clip.Width + rect.x
	if len(clip.Light) < offset+rect.width {
		return nil
	}

	return clip.Light[offset : offset+rect.width]
}

// blitDirect copies pixels unchanged (BlitPixelsDirect, blit_impl.hpp:24-28).
func blitDirect(dst, src, _ []uint8) { copy(dst, src) }

// mapBlit applies a TRN to every copied pixel (BlitPixelsWithMap,
// blit_impl.hpp:47-51).
func mapBlit(trn assets.TRN) blitFunc {
	return func(dst, src, _ []uint8) {
		for i, index := range src {
			dst[i] = trn[index]
		}
	}
}

// blendedBlit blends src into dst through the LUT (BlitPixelsBlended,
// blit_impl.hpp:102-107): the lookup is indexed by destination first, then source.
func blendedBlit(lut BlendLUT) blitFunc {
	return func(dst, src, _ []uint8) {
		for i, index := range src {
			dst[i] = lut[dst[i]][index]
		}
	}
}

// blendedMapBlit applies the TRN first, then blends (BlitPixelsBlendedWithMap,
// blit_impl.hpp:119-122).
func blendedMapBlit(trn assets.TRN, lut BlendLUT) blitFunc {
	return func(dst, src, _ []uint8) {
		for i, index := range src {
			dst[i] = lut[dst[i]][trn[index]]
		}
	}
}

// blendedLightmapBlit lights each pixel, then blends it with the destination
// (BlitPixelsBlendedWithLightmap, blit_impl.hpp:153-176).
func blendedLightmapBlit(tables *[numShades][paletteSize]uint8, lut BlendLUT) blitFunc {
	return func(dst, src, light []uint8) {
		for i, index := range src {
			dst[i] = lut[dst[i]][tables[lightLevel(light, i)][index]]
		}
	}
}

// lightmapBlit rewrites each index through its light level's table
// (BlitPixelsWithLightmap, blit_impl.hpp:78-85).
func lightmapBlit(tables *[numShades][paletteSize]uint8) blitFunc {
	return func(dst, src, light []uint8) {
		for i, index := range src {
			dst[i] = tables[lightLevel(light, i)][index]
		}
	}
}

// lightLevel returns pixel i's light level, fully lit when the draw carries none.
func lightLevel(light []uint8, i int) int {
	if light == nil {
		return fullyLitShade
	}

	return int(light[i])
}
