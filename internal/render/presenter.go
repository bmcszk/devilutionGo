package render

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// fullFade is the fade level of an undimmed frame: ApplyFadeLevel's 256
	// leaves every component unchanged (palette.cpp:140-147).
	fullFade = 256
	// quadVertices is the vertex count of the presenting quad.
	quadVertices = 4
)

var (
	// quadIndices are the two triangles of the presenting quad.
	quadIndices = []uint16{0, 1, 2, 1, 2, 3}

	// paletteShader compiles the embedded program once per process. The source is
	// a build-time constant, so a compile failure is a packaging error, not a
	// runtime condition (same stance as lightvariants.go's unknown-rotation panic).
	paletteShader = sync.OnceValue(func() *PaletteShader {
		shader, err := NewPaletteShader()
		if err != nil {
			panic(fmt.Sprintf("render: %v", err))
		}

		return shader
	})
)

// Presenter renders a palette-index surface as expanded RGBA with one
// DrawTrianglesShader call: the palette-lookup pass at the end of the frame
// (docs/research/dg-graphics.md G10 D2).
type Presenter struct {
	shader   *PaletteShader
	fade     int
	variants *ebiten.Image
	index    *ebiten.Image
	shade    *ebiten.Image
	target   *ebiten.Image

	width, height int
}

// PresenterOption configures a Presenter.
type PresenterOption func(*Presenter)

// NewPresenter builds the presentation pass for one palette state.
//
// The palette variants are uploaded once here: a palette cycle re-derives them,
// so the owner rebuilds the Presenter on a cycle, exactly as upstream swaps the
// system palette (SystemPaletteUpdated, palette.cpp:157-186).
func NewPresenter(state PaletteState, options ...PresenterOption) *Presenter {
	presenter := &Presenter{fade: fullFade}
	for _, option := range options {
		option(presenter)
	}

	presenter.shader = paletteShader()
	presenter.variants = ebiten.NewImage(variantTextureWidth, variantTextureHeight)
	presenter.variants.WritePixels(PaletteVariantsTexture(&state.Variants, presenter.fade))

	return presenter
}

// UniformShades returns a light-level buffer filled with shade, the per-pixel
// form of a uniformly lit draw (ClxDrawLight's lightTableIndex, scrollrt.cpp:252).
func UniformShades(width, height int, shade uint8) []uint8 {
	shades := make([]uint8, width*height)
	for i := range shades {
		shades[i] = shade
	}

	return shades
}

// WithFadeLevel dims the presented frame by fadeLevel/256, the arithmetic of
// ApplyFadeLevel (palette.cpp:140-147). 256 is unmodified, 0 is black.
func WithFadeLevel(fadeLevel int) PresenterOption {
	return func(p *Presenter) {
		p.fade = fadeLevel
	}
}

// Frame presents an index surface and returns the expanded frame, reading the
// result back from the render target.
//
// shades carries one light level per surface pixel (Lightmap::adjustColor,
// light_render.hpp:14-33); nil means a fully lit surface. This is the readback
// half of the pass, used by the parity tests; the game loop calls Draw.
func (p *Presenter) Frame(surface Surface, shades []uint8, keying Keying) Frame {
	p.Draw(surface, shades, keying)

	frame := Frame{Width: surface.Width, Height: surface.Height, Pixels: make([]RGBA, surface.Width*surface.Height)}
	raw := make([]byte, len(frame.Pixels)*textureComponents)
	p.target.ReadPixels(raw)

	for i := range frame.Pixels {
		offset := i * textureComponents
		frame.Pixels[i] = RGBA{R: raw[offset], G: raw[offset+1], B: raw[offset+2], A: raw[offset+3]}
	}

	return frame
}

// Draw runs the palette-lookup pass for one surface into the presenter's render
// target: one DrawTrianglesShader call per frame, the whole world surface in one
// batch (docs/research/dg-graphics.md G10 D2).
func (p *Presenter) Draw(surface Surface, shades []uint8, keying Keying) {
	p.resize(surface.Width, surface.Height)

	p.index.WritePixels(IndexTexture(surface.Pixels))
	p.shade.WritePixels(ShadeTexture(surface.Width, surface.Height, shades))
	p.target.Clear()

	p.target.DrawTrianglesShader(p.quad(), quadIndices, p.shader.program, &ebiten.DrawTrianglesShaderOptions{
		Uniforms: map[string]any{"Keying": keying.KeyingUniform()},
		Images:   [4]*ebiten.Image{p.index, p.shade, p.variants, nil},
	})
}

// Target returns the render target holding the last presented frame.
func (p *Presenter) Target() *ebiten.Image {
	return p.target
}

// String identifies the pass in the ebitenginedebug command log, so a draw-count
// capture can attribute the final pass to this package.
func (p *Presenter) String() string {
	return fmt.Sprintf("palette-lookup %dx%d fade=%d", p.width, p.height, p.fade)
}

// resize allocates the per-surface images once per surface size.
func (p *Presenter) resize(width, height int) {
	if p.width == width && p.height == height && p.target != nil {
		return
	}

	p.width, p.height = width, height
	p.index = ebiten.NewImage(width, height)
	p.shade = ebiten.NewImage(width, height)
	p.target = ebiten.NewImage(width, height)
}

// quad returns the four vertices covering the whole surface in source and
// destination space alike, so the shader samples one index texel per pixel.
func (p *Presenter) quad() []ebiten.Vertex {
	width, height := float32(p.width), float32(p.height)

	vertices := make([]ebiten.Vertex, quadVertices)
	for i, corner := range [quadVertices][2]float32{{0, 0}, {width, 0}, {0, height}, {width, height}} {
		vertices[i] = ebiten.Vertex{
			DstX: corner[0], DstY: corner[1],
			SrcX: corner[0], SrcY: corner[1],
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		}
	}

	return vertices
}
