package render

import (
	_ "embed"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// KeySprite turns a drawn index 0 into a transparent pixel
	// (BlitDirect's keying rule, clx_render.cpp:589).
	KeySprite Keying = iota
	// KeyTile keeps a drawn index 0 opaque, the tile rule (same trap).
	KeyTile
)

const (
	// variantTextureWidth is the variant texture width: one texel per palette index.
	variantTextureWidth = paletteSize
	// variantTextureHeight is the variant texture height: one texel row per light
	// shade, the 16 palette variants of light_render.hpp:14-33.
	variantTextureHeight = numShades
	// textureComponents is the byte count of one RGBA texel in an upload buffer.
	textureComponents = 4
)

// paletteShaderSource is the embedded Kage program of the presentation pass.
//
//go:embed shader.kage
var paletteShaderSource []byte

// Keying is what a drawn palette index 0 means in the presentation pass
// (docs/research/dg-graphics.md G4): sprites key it out, tiles draw it as the
// shade row's black.
type Keying int

// PaletteShader is the compiled palette-lookup program: one fragment pass that
// reads a palette index and a light shade per pixel and returns the expanded
// colour (ClxDrawLight, scrollrt.cpp:252).
type PaletteShader struct {
	program *ebiten.Shader
}

// KeyingUniform is the shader uniform value selecting keying: the Kage program
// compares it against 1 for sprite keying.
func (k Keying) KeyingUniform() int {
	if k == KeySprite {
		return 1
	}

	return 0
}

// NewPaletteShader compiles the embedded shader.kage program.
//
// Compiling needs no graphics device, so this works before the game loop starts.
// There is no upstream palette shader to port: the program is written against
// the Kage pixel-unit rules (docs/research/dg-graphics.md G10 D2).
func NewPaletteShader() (*PaletteShader, error) {
	program, err := ebiten.NewShader(paletteShaderSource)
	if err != nil {
		return nil, fmt.Errorf("render: compile palette shader: %w", err)
	}

	return &PaletteShader{program: program}, nil
}

// Source returns the Kage source the program was compiled from, so a test can
// pin the pass's coordinate and lookup law.
func (*PaletteShader) Source() []byte {
	source := make([]byte, len(paletteShaderSource))
	copy(source, paletteShaderSource)

	return source
}

// PaletteVariantsTexture lays the 16 variant rows out as the RGBA texel buffer
// the shader samples: texel (index, shade) is the palette colour that index
// maps to under shade.
//
// fadeLevel dims every colour by fadeval/256, the arithmetic of ApplyFadeLevel
// (palette.cpp:140-147), so a fade never touches the per-pixel shader path.
func PaletteVariantsTexture(variants *Variants, fadeLevel int) []byte {
	texture := make([]byte, variantTextureWidth*variantTextureHeight*textureComponents)

	for shade := range numShades {
		for index := range paletteSize {
			offset := (shade*variantTextureWidth + index) * textureComponents
			color := variants[shade][index]

			texture[offset] = fadeComponent(color[0], fadeLevel)
			texture[offset+1] = fadeComponent(color[1], fadeLevel)
			texture[offset+2] = fadeComponent(color[2], fadeLevel)
			texture[offset+3] = 0xFF
		}
	}

	return texture
}

// IndexTexture lays an index buffer out as the RGBA texel buffer the shader
// reads its palette index from: the index rides in the red channel, the other
// channels are unused.
func IndexTexture(pixels []uint8) []byte {
	texture := make([]byte, len(pixels)*textureComponents)
	for i, index := range pixels {
		texture[i*textureComponents] = index
		texture[i*textureComponents+3] = 0xFF
	}

	return texture
}

// ShadeTexture lays a per-pixel light-level buffer out as the RGBA texel buffer
// the shader reads its shade from. A nil light buffer means a fully lit
// surface, so every texel carries shade 0 (dLight 0 = bright,
// light_tables.cpp:19-70).
func ShadeTexture(width, height int, shades []uint8) []byte {
	texture := make([]byte, width*height*textureComponents)
	for i := range width * height {
		level := uint8(fullyLitShade)
		if shades != nil {
			level = shades[i]
		}

		texture[i*textureComponents] = level
		texture[i*textureComponents+3] = 0xFF
	}

	return texture
}

// fadeComponent dims one colour component by fadeval/256
// (ApplyFadeLevel, palette.cpp:140-147).
func fadeComponent(component uint8, fadeLevel int) uint8 {
	return uint8(fadeLevel * int(component) / fullFade)
}
