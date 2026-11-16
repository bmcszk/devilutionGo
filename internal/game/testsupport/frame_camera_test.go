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

// TestSmokeFrameHeroClampedNearWorldEdge pins the camera clamp: a hero spawned
// near the world edge must render fully inside the frame, touching no frame
// border row or column.
func TestSmokeFrameHeroClampedNearWorldEdge(t *testing.T) {
	tests := []struct {
		name string
		hero world.Tile
	}{
		{name: "deep off-viewport spawn", hero: world.Tile{X: 100, Y: 100}},
		{name: "last world tile", hero: world.Tile{X: world.MAXDUNX - 1, Y: world.MAXDUNY - 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			t.Setenv("SMOKE_ARTIFACT_DIR", t.TempDir())

			// when
			path, err := testsupport.ComposeFramePNG(tt.hero, nil, nil, testsupport.FrameBefore)
			if err != nil {
				t.Fatalf("write frame png: %v", err)
			}

			// then
			assertHeroRectInsideFrame(t, path)
		})
	}
}

// assertHeroRectInsideFrame decodes the artifact and checks the hero-shaded
// pixels form a rect that stays off every frame border row and column. The
// camera offset decides where the hero lands, so the scan is color-driven.
func assertHeroRectInsideFrame(t *testing.T, path string) {
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

	shades, err := testsupport.SmokeShades()
	if err != nil {
		t.Fatalf("smoke shades: %v", err)
	}

	heroColor := shades["hero"]
	bounds := img.Bounds()

	if x, y := borderPixel(img, heroColor); x >= 0 {
		t.Fatalf("hero shade at (%d, %d) touches a frame edge %v", x, y, bounds)
	}
}

// borderPixel returns the first pixel of shade sitting on a frame border row
// or column; (-1, -1) when the shade stays strictly inside.
func borderPixel(img image.Image, shade color.RGBA) (x, y int) {
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		if x := borderPixelInRow(img, shade, y, bounds); x >= 0 {
			return x, y
		}
	}

	return -1, -1
}

// borderPixelInRow returns the first x in row y where shade sits on a border
// column, or on the row's own border line; -1 when the row is clean.
func borderPixelInRow(img image.Image, shade color.RGBA, y int, bounds image.Rectangle) int {
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		if img.At(x, y) != shade {
			continue
		}

		onBorder := x == bounds.Min.X || x == bounds.Max.X-1 || y == bounds.Min.Y || y == bounds.Max.Y-1
		if onBorder {
			return x
		}
	}

	return -1
}
