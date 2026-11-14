package game_test

import (
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// TestSmokeStoryFrames pins the causal-chain artifact: the smoke run writes a
// before frame (hero at the start tile), an after frame (hero at the clicked
// tile, click target marked), and the target color appears only in the after
// frame.
func TestSmokeStoryFrames(t *testing.T) {
	// given
	// a driver (the just smoke recipe) presets SMOKE_ARTIFACT_DIR; only a bare
	// go test run falls back to a temp dir, so the recipe's env survives.
	outDir, preset := os.LookupEnv("SMOKE_ARTIFACT_DIR")
	if !preset {
		outDir = t.TempDir()
		t.Setenv("SMOKE_ARTIFACT_DIR", outDir)
	}

	h, err := game.NewSmokeHarness(slog.New(newCaptureHandler()))
	if err != nil {
		t.Fatalf("new smoke harness: %v", err)
	}
	defer func() { _ = h.Close() }()

	h.RunTicks(10)

	start := h.State().Hero()
	target := world.Tile{X: start.X + 1, Y: start.Y}

	shades, err := testsupport.SmokeShades()
	if err != nil {
		t.Fatalf("smoke shades: %v", err)
	}

	// when
	before, err := testsupport.ComposeFramePNG(start, h.State().Grid(), &target, testsupport.FrameBefore)
	if err != nil {
		t.Fatalf("write before frame: %v", err)
	}

	h.Click(int(target.X)*64+32, int(target.Y)*64+32)
	h.RunTicks(10)

	after, err := testsupport.ComposeFramePNG(h.State().Hero(), h.State().Grid(), &target, testsupport.FrameAfter)
	if err != nil {
		t.Fatalf("write after frame: %v", err)
	}

	// then
	assertStoryFrames(t, outDir, before, after, shades)
}

// assertStoryFrames pins the story artifact contract: both files under outDir,
// the hero moved between them, the target shade only over the clicked cell in
// the after frame.
func assertStoryFrames(t *testing.T, outDir, before, after string, shades map[string]color.RGBA) {
	t.Helper()

	if filepath.Dir(before) != outDir || filepath.Dir(after) != outDir {
		t.Fatalf("artifacts at %s / %s, want both under %s", before, after, outDir)
	}

	beforeImg := decodePNG(t, before)
	afterImg := decodePNG(t, after)

	heroBefore := shadeRectCenter(t, beforeImg, shades["hero"])
	heroAfter := shadeRectCenter(t, afterImg, shades["hero"])
	if heroBefore == heroAfter {
		t.Fatalf("hero marker at the same frame position %v in both frames, want it to move", heroBefore)
	}

	if got := pixelCenter(afterImg, heroAfter); got != shades["target"] {
		t.Fatalf("clicked cell center color %v, want target shade %v", got, shades["target"])
	}

	if got := tileShadeCount(afterImg, shades["target"]); got != 63*63 {
		t.Fatalf("target shade covers %d px, want the clicked cell's 63x63 interior", got)
	}

	if tileShadeCount(beforeImg, shades["target"]) != 0 {
		t.Fatal("before frame contains the target shade, want none")
	}

	assertArtifact(t, outDir, "frame-1-before.png")
	assertArtifact(t, outDir, "frame-2-after.png")
}

// assertArtifact fails when name is missing under outDir.
func assertArtifact(t *testing.T, outDir, name string) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
		t.Fatalf("%s missing: %v", name, err)
	}
}

// shadeRectCenter returns the center of the rect the shade fills, or fails.
func shadeRectCenter(t *testing.T, img image.Image, shade color.Color) image.Point {
	t.Helper()

	bounds := img.Bounds()
	minX, minY, maxX, maxY := bounds.Dx(), bounds.Dy(), -1, -1

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.At(x, y) != shade {
				continue
			}

			minX, maxX = min(minX, x), max(maxX, x)
			minY, maxY = min(minY, y), max(maxY, y)
		}
	}

	if maxX < 0 {
		t.Fatalf("shade %v absent from frame", shade)
	}

	return image.Point{X: (minX + maxX) / 2, Y: (minY + maxY) / 2}
}

// pixelCenter samples the middle pixel of the rect the shade fills.
func pixelCenter(img image.Image, center image.Point) color.Color {
	return img.At(center.X, center.Y)
}

// tileShadeCount counts the pixels painted in the shade.
func tileShadeCount(img image.Image, shade color.Color) int {
	bounds := img.Bounds()
	count := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.At(x, y) == shade {
				count++
			}
		}
	}

	return count
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
