// Package testsupport holds the helpers shared by the parity tests: the
// golden-grid diff and the canonical state hash.
package testsupport

import (
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// smokeFrameTilesW and smokeFrameTilesH size the smoke frame: a 20×15 tile
// viewport at 64px per tile.
const (
	smokeFrameTilesW = 20
	smokeFrameTilesH = 15
	smokeTilePixels  = 64
)

// smoke palette indexes: floor and hero get distinct fixed indexes expanded
// through the committed l1_1 palette.
const (
	smokeIndexFloor = 0xC0
	smokeIndexHero  = 0x40
)

// ComposeFramePNG paints the smoke world (floor everywhere, the hero tile
// distinct) into an index surface, expands it through the committed l1_1
// palette fixture, and encodes a PNG into SMOKE_ARTIFACT_DIR (or the OS temp
// dir when unset). It returns the written path. Test-only tooling, not
// simulation code.
func ComposeFramePNG(hero world.Tile) (string, error) {
	surf := render.NewSurface(smokeFrameTilesW*smokeTilePixels, smokeFrameTilesH*smokeTilePixels)
	composeFrame(surf, hero)

	pal, err := loadSmokePalette()
	if err != nil {
		return "", fmt.Errorf("testsupport: smoke frame: %w", err)
	}

	rgb := expandPalette(pal, surf)

	outDir := os.Getenv("SMOKE_ARTIFACT_DIR")
	if outDir == "" {
		outDir = os.TempDir()
	}

	path := filepath.Join(outDir, "frame.png")
	if err := writePNG(path, rgb, surf.Width, surf.Height); err != nil {
		return "", fmt.Errorf("testsupport: smoke frame: %w", err)
	}

	_, _ = fmt.Fprintln(os.Stderr, "smoke frame:", path)

	return path, nil
}

// composeFrame paints the world grid: floor everywhere, the hero tile
// distinct. Smoke artifact, not parity.
func composeFrame(surf render.Surface, hero world.Tile) {
	for y := range smokeFrameTilesH {
		for x := range smokeFrameTilesW {
			fillTile(surf, world.Tile{X: x, Y: y}, hero)
		}
	}
}

// fillTile paints one smoke tile square: the hero tile distinct from floor.
func fillTile(surf render.Surface, tile, hero world.Tile) {
	index := smokeIndexFloor
	if tile == hero {
		index = smokeIndexHero
	}

	for py := range smokeTilePixels {
		for px := range smokeTilePixels {
			surf.Set(tile.X*smokeTilePixels+px, tile.Y*smokeTilePixels+py, uint8(index))
		}
	}
}

// expandPalette turns the index surface into RGBA through the palette.
func expandPalette(pal assets.Palette, surf render.Surface) []color.RGBA {
	out := make([]color.RGBA, surf.Width*surf.Height)

	for i, idx := range surf.Pixels {
		out[i] = color.RGBA{R: pal[idx][0], G: pal[idx][1], B: pal[idx][2], A: 255}
	}

	return out
}

// loadSmokePalette decodes the committed l1_1 hex dump.
func loadSmokePalette() (assets.Palette, error) {
	body, err := os.ReadFile(filepath.Join("..", "testdata", "render", "pal", "l1_1.pal.hex"))
	if err != nil {
		return assets.Palette{}, fmt.Errorf("read palette fixture: %w", err)
	}

	decoded, err := hex.DecodeString(strings.Join(strings.Fields(string(body)), ""))
	if err != nil {
		return assets.Palette{}, fmt.Errorf("decode palette fixture: %w", err)
	}

	pal, err := assets.LoadPaletteFromBytes(decoded)
	if err != nil {
		return assets.Palette{}, fmt.Errorf("load palette: %w", err)
	}

	return *pal, nil
}

// writePNG encodes the RGBA rows into path.
func writePNG(path string, rgb []color.RGBA, width, height int) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			img.SetRGBA(x, y, rgb[y*width+x])
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	return nil
}
