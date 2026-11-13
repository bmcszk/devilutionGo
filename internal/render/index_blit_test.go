package render_test

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// blitTrace is one row of the committed blit-flavour trace.
type blitTrace struct {
	flavour string
	src     []byte
	dst     []byte
	light   []byte
	want    []byte
}

// TestBlitFlavoursMatchUpstreamTrace pins all six blit flavours byte for byte
// against a trace produced by the pinned upstream blit_impl.hpp over the committed
// cathedral light tables and the gendata/pause.trn golden
// (blit_impl.hpp:20-190).
func TestBlitFlavoursMatchUpstreamTrace(t *testing.T) {
	for _, row := range readBlitTrace(t) {
		t.Run(row.flavour, func(t *testing.T) {
			// given
			cases := blitFlavours(t)
			surface := render.NewSurface(len(row.dst), 1)
			copy(surface.Pixels, row.dst)
			clip := render.Clip{X: 0, Y: 0, Width: len(row.src), Height: 1, Pixels: row.src, Light: row.light}

			// when
			cases[row.flavour](t, surface, clip)

			// then
			expectBytesEqual(t, surface.Pixels, row.want)
		})
	}
}

// TestBlitLightmapUsesLightTables pins the lightmap flavour to the level's light
// tables: each destination index is lightTables[level][source index]
// (BlitPixelsWithLightmap, blit_impl.hpp:78-85; Lightmap::adjustColor,
// light_render.hpp:30-33).
func TestBlitLightmapUsesLightTables(t *testing.T) {
	// given
	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)
	surface := render.NewSurface(4, 1)
	source := []byte{7, 7, 7, 7}
	levels := []byte{0, 1, 15, 8}
	clip := render.Clip{X: 0, Y: 0, Width: 4, Height: 1, Pixels: source, Light: levels}

	// when
	surface.DrawLightmap(clip, &tables.Tables)

	// then
	for i, level := range levels {
		if got, want := surface.At(i, 0), tables.Tables[level][source[i]]; got != want {
			t.Fatalf("pixel %d at light %d = %d, want %d", i, level, got, want)
		}
	}
}

// TestBlitDirectKeepsIndexZeroForTiles pins the tile half of the keying trap:
// a direct blit copies index 0 unchanged, so a dungeon micro-tile keeps its
// legitimate black (docs/research/dg-graphics.md G4).
func TestBlitDirectKeepsIndexZeroForTiles(t *testing.T) {
	// given
	surface := render.NewSurface(3, 1)
	surface.Pixels[0], surface.Pixels[1], surface.Pixels[2] = 9, 9, 9
	clip := render.Clip{X: 1, Y: 0, Width: 2, Height: 1, Pixels: []byte{0, 12}}

	// when
	surface.DrawDirect(clip)

	// then
	expectBytesEqual(t, surface.Pixels, []byte{9, 0, 12})
}

// TestExpandSpriteKeysIndexZeroTransparent pins the sprite half of the keying
// trap: a sprite expansion turns index 0 into a transparent pixel
// (docs/research/dg-graphics.md G4, G11).
func TestExpandSpriteKeysIndexZeroTransparent(t *testing.T) {
	// given
	variants := scriptedVariants()
	surface := render.NewSurface(2, 1)
	copy(surface.Pixels, []byte{0, 1})

	// when
	sprite := surface.ExpandSprite(&variants, 0)

	// then
	if got := sprite.Pixels[0]; got != (render.RGBA{}) {
		t.Fatalf("sprite index 0 expanded to %v, want transparent", got)
	}

	if sprite.Pixels[1] != (render.RGBA{R: 4, G: 5, B: 6, A: 0xFF}) {
		t.Fatalf("sprite index 1 expanded to %v, want the palette colour", sprite.Pixels[1])
	}
}

// TestExpandTileKeepsIndexZeroOpaque pins the tile half of the keying trap: a
// tile expansion has no transparent index, so index 0 keeps its palette colour
// (docs/research/dg-graphics.md G4, G11).
func TestExpandTileKeepsIndexZeroOpaque(t *testing.T) {
	// given
	variants := scriptedVariants()
	surface := render.NewSurface(2, 1)
	copy(surface.Pixels, []byte{0, 1})

	// when
	tile := surface.ExpandTile(&variants, 0)

	// then
	if got := tile.Pixels[0]; got != (render.RGBA{R: 1, G: 2, B: 3, A: 0xFF}) {
		t.Fatalf("tile index 0 expanded to %v, want the palette colour opaque", got)
	}

	if tile.Pixels[1] != (render.RGBA{R: 4, G: 5, B: 6, A: 0xFF}) {
		t.Fatalf("tile index 1 expanded to %v, want the palette colour", tile.Pixels[1])
	}
}

// TestBlitClipsToSurface pins the rasteriser's edge clipping: a sprite hanging off
// the right and bottom edges writes only the visible pixels and never panics
// (ClipX + the height clamp, clx_render.cpp:63-73, :230-260).
func TestBlitClipsToSurface(t *testing.T) {
	// given
	surface := render.NewSurface(4, 2)
	clip := render.Clip{X: 2, Y: 1, Width: 4, Height: 3, Pixels: []byte{5, 6, 7, 8, 11, 12, 13, 14, 21, 22, 23, 24}}

	// when
	surface.DrawDirect(clip)

	// then
	expectBytesEqual(t, surface.Pixels, []byte{0, 0, 0, 0, 0, 0, 5, 6})
}

// TestBlitClipsNegativeOffset pins the left/top clip: a sprite drawn at a negative
// coordinate writes only its overlapping part, at the correct source column and
// row (SkipRestOfLineWithOverrun, clx_render.cpp:80-95).
func TestBlitClipsNegativeOffset(t *testing.T) {
	// given
	surface := render.NewSurface(2, 2)
	clip := render.Clip{X: -1, Y: -1, Width: 3, Height: 3, Pixels: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}}

	// when
	surface.DrawDirect(clip)

	// then
	expectBytesEqual(t, surface.Pixels, []byte{5, 6, 8, 9})
}

// TestBlitRowStrideFollowsTheSprite pins the source stride to the sprite width, not
// the surface width: with a sprite taller than the surface and a negative vertical
// offset, each destination row must read its own sprite row (the source index is
// row*clip.Width+column, never row*surface.Width, clx_render.cpp:80-95).
func TestBlitRowStrideFollowsTheSprite(t *testing.T) {
	// given
	surface := render.NewSurface(2, 2)
	sprite := []byte{1, 2, 3, 4, 5, 6}
	clip := render.Clip{X: 0, Y: -1, Width: 2, Height: 3, Pixels: sprite}

	// when
	surface.DrawDirect(clip)

	// then
	expectBytesEqual(t, surface.Pixels, []byte{3, 4, 5, 6})
}

// TestBlitLightmapFollowsTheVisibleColumn pins the light buffer to the sprite, not
// the surface: with a negative offset the light level of sprite column 1 drives
// surface column 0 (lightRow offset, clx_render.cpp:230-260).
func TestBlitLightmapFollowsTheVisibleColumn(t *testing.T) {
	// given
	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)
	surface := render.NewSurface(1, 1)
	clip := render.Clip{
		X: -1, Y: 0, Width: 2, Height: 1,
		Pixels: []byte{7, 7},
		Light:  []byte{0, 15},
	}

	// when
	surface.DrawLightmap(clip, &tables.Tables)

	// then
	if got, want := surface.At(0, 0), tables.Tables[15][7]; got != want {
		t.Fatalf("pixel = %d, want %d (light level of sprite column 1)", got, want)
	}
}

// TestBlitOverwritesRatherThanBlendsDirectly pins the flush-draw law: a second
// direct draw replaces the pixel instead of blending with it, which is what makes
// the index buffer a plain raster target (BlitFillDirect, blit_impl.hpp:20-22).
func TestBlitOverwritesRatherThanBlendsDirectly(t *testing.T) {
	// given
	surface := render.NewSurface(2, 1)
	surface.Pixels[0], surface.Pixels[1] = 40, 41
	clip := render.Clip{X: 0, Y: 0, Width: 2, Height: 1, Pixels: []byte{1, 2}}

	// when
	surface.DrawDirect(clip)

	// then
	expectBytesEqual(t, surface.Pixels, []byte{1, 2})
}

// blitFlavours maps a trace flavour name to the Go blit it exercises.
func blitFlavours(t *testing.T) map[string]func(*testing.T, render.Surface, render.Clip) {
	t.Helper()

	pal := loadPalette(t, "diablo.pal.hex")
	lut := render.BuildBlendingLUT(pal, -1, -1)
	trn := loadTRN(t, "gendata_pause.trn.hex")
	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)

	return map[string]func(*testing.T, render.Surface, render.Clip){
		"direct":  func(_ *testing.T, s render.Surface, c render.Clip) { s.DrawDirect(c) },
		"withmap": func(_ *testing.T, s render.Surface, c render.Clip) { s.DrawWithMap(c, trn) },
		"blended": func(_ *testing.T, s render.Surface, c render.Clip) { s.DrawBlended(c, lut) },
		"blended_withmap": func(_ *testing.T, s render.Surface, c render.Clip) {
			s.DrawBlendedWithMap(c, trn, lut)
		},
		"lightmap": func(_ *testing.T, s render.Surface, c render.Clip) {
			s.DrawLightmap(c, &tables.Tables)
		},
		"blended_lightmap": func(_ *testing.T, s render.Surface, c render.Clip) {
			s.DrawBlendedWithLightmap(c, &tables.Tables, lut)
		},
	}
}

// readBlitTrace decodes the committed flavour trace.
func readBlitTrace(t *testing.T) []blitTrace {
	t.Helper()

	path := filepath.Join(fixtureDir, "blit", "blit_trace.tsv")

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	var rows []blitTrace
	for line := range strings.SplitSeq(string(body), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rows = append(rows, parseBlitRow(t, path, line))
	}

	if len(rows) == 0 {
		t.Fatalf("fixture %s holds no rows", path)
	}

	return rows
}

// parseBlitRow decodes the seven columns of one trace row.
func parseBlitRow(t *testing.T, path, line string) blitTrace {
	t.Helper()

	fields := strings.Split(line, "	")
	if len(fields) != 7 {
		t.Fatalf("fixture %s: %d fields in %q, want 7", path, len(fields), line)
	}

	return blitTrace{
		flavour: fields[0],
		src:     decodeOptionalColumn(t, path, fields[2]),
		dst:     decodeColumn(t, path, fields[3]),
		light:   decodeOptionalColumn(t, path, fields[4]),
		want:    decodeColumn(t, path, fields[6]),
	}
}

// decodeColumn decodes a required hex column.
func decodeColumn(t *testing.T, path, value string) []byte {
	t.Helper()

	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("fixture %s: hex %q: %v", path, value, err)
	}

	return decoded
}

// decodeOptionalColumn decodes a hex column, mapping the fixture's "-" to nil.
func decodeOptionalColumn(t *testing.T, path, value string) []byte {
	t.Helper()

	if value == "-" {
		return nil
	}

	return decodeColumn(t, path, value)
}

// scriptedVariants builds every shade from a scripted palette, so an expansion
// result names the index it came from.
func scriptedVariants() render.Variants {
	pal := assets.Palette{{1, 2, 3}, {4, 5, 6}}
	variants := render.Variants{}

	for shade := range variants {
		for index := range variants[shade] {
			variants[shade][index] = render.RGB(pal[index])
		}
	}

	return variants
}

// loadTRN reads a committed .trn hexdump from the gfx goldens.
func loadTRN(t *testing.T, name string) assets.TRN {
	t.Helper()

	raw := readFlatBytes(t, filepath.Join("..", "testdata", "gfx", "trn", name), 256)

	trn, err := assets.DecodeTRN(raw)
	if err != nil {
		t.Fatalf("decode trn %s: %v", name, err)
	}

	return trn
}
