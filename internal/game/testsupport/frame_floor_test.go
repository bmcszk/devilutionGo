package testsupport_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// TestSmokeFrameReadableFloor pins the readable-floor rendering: walkable and
// blocked cells render in distinct shades and cell borders carry the grid-line
// shade. The camera offset maps world tiles onto frame pixels.
func TestSmokeFrameReadableFloor(t *testing.T) {
	// given
	hero := world.Tile{X: 56, Y: 56}
	blocked := world.Tile{X: 58, Y: 58}
	walkable := world.Tile{X: 54, Y: 54}

	t.Setenv("SMOKE_ARTIFACT_DIR", t.TempDir())

	shades, err := testsupport.SmokeShades()
	if err != nil {
		t.Fatalf("smoke shades: %v", err)
	}

	state := world.NewState()
	state.Grid().SOLData = append(state.Grid().SOLData, world.TileSolid)
	state.Grid().DPiece[blocked.Y][blocked.X] = len(state.Grid().SOLData) - 1

	// when
	path, err := testsupport.ComposeFramePNG(hero, state.Grid(), nil, testsupport.FrameBefore)
	if err != nil {
		t.Fatalf("write frame png: %v", err)
	}

	// then
	img := decodePNG(t, path)

	if got := pixelAt(img, hero, blocked); got != shades["blocked"] {
		t.Fatalf("blocked cell %v color %v, want blocked shade %v", blocked, got, shades["blocked"])
	}

	if got := pixelAt(img, hero, walkable); got != shades["walkable"] {
		t.Fatalf("walkable cell %v color %v, want walkable shade %v", walkable, got, shades["walkable"])
	}

	if got := borderAt(img, hero, walkable); got != shades["gridline"] {
		t.Fatalf("cell %v border color %v, want grid-line shade %v", walkable, got, shades["gridline"])
	}
}

// decodePNG opens and decodes the artifact at path.
func decodePNG(t *testing.T, path string) image.Image {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}

	return img
}

// pixelAt samples the center pixel of the world tile through the camera offset
// the compositor anchored on hero.
func pixelAt(img image.Image, hero, tile world.Tile) color.Color {
	cam := testsupport.CameraOffset(hero)

	return img.At((tile.X-cam.X)*64+32, (tile.Y-cam.Y)*64+32)
}

// borderAt samples the top-left corner pixel of the world tile's cell.
func borderAt(img image.Image, hero, tile world.Tile) color.Color {
	cam := testsupport.CameraOffset(hero)

	return img.At((tile.X-cam.X)*64, (tile.Y-cam.Y)*64)
}
