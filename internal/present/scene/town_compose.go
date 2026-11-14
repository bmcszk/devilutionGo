package scene

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// view geometry of the composed frame: the tile counts CalcViewportGeometry
// derives for the 640x480 world surface (scrollrt.cpp:1675-1721).
const (
	viewRows    = 18
	viewColumns = 11
	viewOriginX = -32
	viewOriginY = 160
)

// heroTile is the tile the hero stands on for the first scene slice:
// ViewPosition of ENTRY_MAIN (75, 68) mapped to the dPiece grid the world
// walk reads (town.cpp:363-365; the player entity lands with devilGo-s7x0).
var heroTile = render.TilePos{X: 75, Y: 68}

// WritePNG writes the composed frame's palette-expanded PNG artifact.
func WritePNG(composed TownScene, path string) error {
	shades := render.UniformShades(composed.Surface.Width, composed.Surface.Height, 0)
	frame := composed.Surface.ExpandTile(&composed.State.Variants, 0)
	_ = shades

	data, err := frameToPNGBytes(frame)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir artifact dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

// DistinctColors counts the distinct palette indices of the composed
// surface; a real town frame is richly coloured, a flat one is not.
func DistinctColors(composed TownScene) int {
	seen := make(map[uint8]bool, 256)
	for _, index := range composed.Surface.Pixels {
		seen[index] = true
	}

	return len(seen)
}

// HasDiamondEdge checks the tile-diamond structure: over the lower half of
// the frame some column must carry a stepwise colour boundary whose row
// neighbours shift one pixel — the isometric diamond silhouette.
func HasDiamondEdge(composed TownScene) bool {
	for x := range composed.Surface.Width {
		if diamondEdgeAt(composed.Surface, x) {
			return true
		}
	}

	return false
}

// townWorld builds the render.World adapter and the frame's view geometry
// from the composed grids (scrollrt.cpp:784-953 reads the dPiece/dLight/
// dSpecial/TransList globals directly; the adapter injects them).
func townWorld(grid levels.PieceGrid, micros render.DungeonMicros, sol render.SOLData) (render.World, render.View) {
	view := &townWorldView{grid: grid, micros: micros, sol: sol, light: townLight()}

	start, origin := composeView(heroTile)

	return view.world(), render.View{
		Start:       start,
		Origin:      origin,
		Rows:        viewRows,
		Columns:     viewColumns,
		ScreenWidth: render.WorldLogicalWidth,
		MicroTiles:  16,
	}
}

// townLight extracts the grids' dLight layer; CreateTown runs after
// InitLight, whose town default is 0 = fully lit (gendung.cpp:301-307,
// light.go InitLight).
func townLight() [levels.MAXDUNX][levels.MAXDUNY]uint8 {
	return [levels.MAXDUNX][levels.MAXDUNY]uint8{}
}

// composeView computes the first tile of the view and its pixel origin from
// the hero tile, the CalcFirstTilePosition model for a standing player:
// render from one row up at the tile above-left, origin offset by the
// viewport geometry (scrollrt.cpp:1150-1234 CalcFirstTilePosition +
// CalcViewportGeometry).
func composeView(hero render.TilePos) (render.TilePos, render.PixelPos) {
	heroScreen := render.WorldToScreen(hero.X, hero.Y)
	// the hero tile lands at the screen centre bottom; the walk starts at
	// the tile above the viewport's centre (scrollrt.cpp:1679-1693).
	start := render.TilePos{X: hero.X - (viewColumns+1)/2 + 1, Y: hero.Y - (viewRows+1)/2 + 1}
	origin := render.PixelPos{
		X: viewOriginX - heroScreen.X%render.TileWidth,
		Y: viewOriginY,
	}

	return start, origin
}

// renderTownFrame runs the three world passes into a 640x480 index surface
// (scrollrt.cpp:1281-1283 DrawFloor/DrawTileContent/DrawOOB).
func renderTownFrame(
	world render.World,
	view render.View,
	cel []byte,
	state render.PaletteState,
	town *townWorldView,
) render.Surface {
	r := &townRenderer{
		surface: render.NewSurface(render.WorldLogicalWidth, render.WorldLogicalHeight),
		view:    town,
		cel:     cel,
		state:   state,
		tables:  light.MakeLightTables(light.DTYPE_TOWN),
	}

	render.DrawWorld(r, world, view)

	return r.surface
}

// frameToPNGBytes serialises one Frame as PNG.
func frameToPNGBytes(frame render.Frame) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, frameToImage(frame)); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}

	return buf.Bytes(), nil
}

// diamondEdgeAt reports whether column x carries a diagonal boundary: a run
// of consecutive rows where the colour-change neighbour walks one pixel per
// row.
func diamondEdgeAt(surface render.Surface, x int) bool {
	steps := 0
	for y := render.TriangleHeight; y+1 < surface.Height; y++ {
		if boundaryAt(surface, x, y) == boundaryAt(surface, x+1, y+1) && boundaryAt(surface, x, y) > 0 {
			steps++
			if steps >= render.TriangleHeight/2 {
				return true
			}

			continue
		}

		steps = 0
	}

	return false
}

// boundaryAt returns the column of the first horizontal colour change at or
// right of x on row y, 0 when none.
func boundaryAt(surface render.Surface, x, y int) int {
	row := y * surface.Width
	for i := x + 1; i < surface.Width; i++ {
		if surface.Pixels[row+i] != surface.Pixels[row+i-1] {
			return i
		}
	}

	return 0
}

// frameToImage renders one Frame into an NRGBA image (the artifact form).
func frameToImage(frame render.Frame) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, frame.Width, frame.Height))
	for y := range frame.Height {
		for x := range frame.Width {
			c := frame.Pixels[y*frame.Width+x]
			img.SetNRGBA(x, y, nrgbaOf(c))
		}
	}

	return img
}

// nrgbaOf converts one render RGBA pixel.
func nrgbaOf(c render.RGBA) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}
