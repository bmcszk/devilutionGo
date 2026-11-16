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

// smokePaletteRelPath is the palette fixture relative to the repo's internal/
// root; the lookup walks up from the cwd so any package (not only internal/game
// tests) can compose a frame.
const smokePaletteRelPath = "testdata/render/pal/l1_1.pal.hex"

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
	smokeIndexFloor    = 0xC0
	smokeIndexHero     = 0x40
	smokeIndexTarget   = 0x30
	smokeIndexBlocked  = 0x14
	smokeIndexGridLine = 0x15
)

const (
	// FrameBefore is frame-1-before.png: the world before the click.
	FrameBefore FrameKind = iota
	// FrameAfter is frame-2-after.png: the world after the click.
	FrameAfter
)

// FrameKind selects the story artifact a compose writes: the before frame
// (hero at the start tile, no target marker) or the after frame (hero moved,
// click target marked).
type FrameKind int

// SmokeShades expands the artifact's palette roles through the committed l1_1
// palette: the RGBA colors a composed frame uses for the hero marker, the
// click target, walkable floor, blocked tiles, and the cell grid lines. Tests
// pin artifacts against these exact values.
func SmokeShades() (map[string]color.RGBA, error) {
	pal, err := loadSmokePalette()
	if err != nil {
		return nil, fmt.Errorf("testsupport: smoke shades: %w", err)
	}

	expand := func(idx uint8) color.RGBA {
		return color.RGBA{R: pal[idx][0], G: pal[idx][1], B: pal[idx][2], A: 255}
	}

	return map[string]color.RGBA{
		"hero":     expand(smokeIndexHero),
		"target":   expand(smokeIndexTarget),
		"walkable": expand(smokeIndexFloor),
		"blocked":  expand(smokeIndexBlocked),
		"gridline": expand(smokeIndexGridLine),
	}, nil
}

// ComposeFramePNG paints the smoke world (walkable floor, blocked tiles darker,
// the hero tile distinct, 1px grid lines) into an index surface, expands it
// through the committed l1_1 palette fixture, and encodes a PNG into
// SMOKE_ARTIFACT_DIR (or the OS temp dir when unset). A nil grid renders
// everything walkable. It returns the written path. Test-only tooling, not
// simulation code.
func ComposeFramePNG(hero world.Tile, grid *world.Grid, target *world.Tile, kind FrameKind) (string, error) {
	surf := render.NewSurface(smokeFrameTilesW*smokeTilePixels, smokeFrameTilesH*smokeTilePixels)
	composeFrame(surf, hero, grid)

	if kind == FrameAfter && target != nil {
		stampTarget(surf, hero, *target)
	}

	pal, err := loadSmokePalette()
	if err != nil {
		return "", fmt.Errorf("testsupport: smoke frame: %w", err)
	}

	rgb := expandPalette(pal, surf)

	outDir := os.Getenv("SMOKE_ARTIFACT_DIR")
	if outDir == "" {
		outDir = os.TempDir()
	}

	path := filepath.Join(outDir, frameName(kind))
	if err := writePNG(path, rgb, surf.Width, surf.Height); err != nil {
		return "", fmt.Errorf("testsupport: smoke frame: %w", err)
	}

	_, _ = fmt.Fprintln(os.Stderr, "smoke frame:", path)

	return path, nil
}

// CameraOffset is the world-to-camera translation for the top-left frame tile:
// center the hero on the viewport, then clamp so the viewport never leaves the
// world bounds (the hero at a world edge renders fully inside the frame).
func CameraOffset(hero world.Tile) world.Tile {
	maxOffX := world.MAXDUNX - smokeFrameTilesW
	maxOffY := world.MAXDUNY - smokeFrameTilesH

	offX := min(max(hero.X-smokeFrameTilesW/2, 0), maxOffX)
	offY := min(max(hero.Y-smokeFrameTilesH/2, 0), maxOffY)

	return world.Tile{X: offX, Y: offY}
}

// composeFrame paints the world grid through the camera anchored on the hero:
// walkable floor, blocked tiles darker, the hero tile distinct. Smoke artifact,
// not parity.
func composeFrame(surf render.Surface, hero world.Tile, grid *world.Grid) {
	cam := CameraOffset(hero)

	for y := range smokeFrameTilesH {
		for x := range smokeFrameTilesW {
			worldTile := world.Tile{X: x + cam.X, Y: y + cam.Y}
			fillTile(surf, world.Tile{X: x, Y: y}, hero.Sub(cam), worldTile, grid)
		}
	}
}

// findInternalTestData walks up from the cwd until the repo's internal/ root
// holds the path, returning its absolute location.
func findInternalTestData(rel string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("testsupport: smoke frame: %w", err)
	}

	for {
		candidate := filepath.Join(dir, "internal", rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("testsupport: smoke frame: %s not found above %s", rel, dir)
		}

		dir = parent
	}
}

// fillTile paints one smoke tile square: the hero tile distinct from floor.
// The hero marker keeps its cell's grid line on the bottom and right, so a
// world-edge spawn still shows the line on the frame border instead of the
// marker clipping into it.
func fillTile(surf render.Surface, tile, hero, worldTile world.Tile, grid *world.Grid) {
	for py := range smokeTilePixels {
		for px := range smokeTilePixels {
			surf.Set(tile.X*smokeTilePixels+px, tile.Y*smokeTilePixels+py, uint8(tileShade(tile, hero, worldTile, grid, px, py)))
		}
	}
}

// tileShade picks the palette index for one pixel of the tile: blocked tiles
// darker than walkable floor, the hero marker over its cell, and a 1px top/left
// grid line elsewhere (the hero keeps its bottom/right line so a world-edge
// spawn never clips the marker into the frame border).
func tileShade(tile, hero, worldTile world.Tile, grid *world.Grid, px, py int) uint8 {
	if tile == hero {
		if px < smokeTilePixels-1 && py < smokeTilePixels-1 {
			return smokeIndexHero
		}

		return smokeIndexGridLine
	}

	if isGridLine(px, py) {
		return smokeIndexGridLine
	}

	if gridBlocked(grid, worldTile) {
		return smokeIndexBlocked
	}

	return smokeIndexFloor
}

// gridBlocked reports whether the tile's SOL entry marks it solid; tiles
// outside the world bounds or a nil grid (bare compositor callers) block
// nothing.
func gridBlocked(grid *world.Grid, tile world.Tile) bool {
	if grid == nil || !world.InDungeonBounds(tile.X, tile.Y) {
		return false
	}

	return grid.IsTileSolid(tile.X, tile.Y)
}

// isGridLine reports whether the pixel is one of the cell's 1px top/left grid
// lines.
func isGridLine(px, py int) bool { return px == 0 || py == 0 }

// frameName is the artifact filename for the frame kind.
func frameName(kind FrameKind) string {
	if kind == FrameBefore {
		return "frame-1-before.png"
	}

	return "frame-2-after.png"
}

// stampTarget paints the click target tile in the target shade, under the
// grid lines, wherever it lands inside the viewport.
func stampTarget(surf render.Surface, hero, target world.Tile) {
	cam := CameraOffset(hero)
	rel := target.Sub(cam)

	if rel.X < 0 || rel.X >= smokeFrameTilesW || rel.Y < 0 || rel.Y >= smokeFrameTilesH {
		return
	}

	for py := 1; py < smokeTilePixels; py++ {
		for px := 1; px < smokeTilePixels; px++ {
			surf.Set(rel.X*smokeTilePixels+px, rel.Y*smokeTilePixels+py, uint8(smokeIndexTarget))
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
	path, err := findInternalTestData(smokePaletteRelPath)
	if err != nil {
		return assets.Palette{}, err
	}

	body, err := os.ReadFile(path)
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
