//go:build windowed

package scene_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// TestScreenshotCaptureWriter tests the screenshot writer headlessly:
// it verifies that the capture function produces a valid PNG without
// requiring a live display (devilGo-1awt).
func TestScreenshotCaptureWriter(t *testing.T) {
	// given a composed town scene
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES")
	}

	composed, err := scene.ComposeTownFrame(dirs)
	if err != nil {
		t.Fatalf("ComposeTownFrame: %v", err)
	}

	// when we capture the frame through the presenter's target
	// (simulating the scenePresenter.capture logic)
	presenter := render.NewPresenter(composed.State)
	presenter.Draw(composed.Surface, nil, render.KeyTile)

	// capture the target image
	raw := make([]byte, 4*composed.Surface.Width*composed.Surface.Height)
	presenter.Target().ReadPixels(raw)

	img := image.NewRGBA(image.Rect(0, 0, composed.Surface.Width, composed.Surface.Height))
	for y := range composed.Surface.Height {
		for x := range composed.Surface.Width {
			offset := (y*composed.Surface.Width + x) * 4
			r := raw[offset]
			g := raw[offset+1]
			b := raw[offset+2]
			a := raw[offset+3]
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	// then the PNG encodes and decodes correctly
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}

	// decode back to verify
	decoded, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}

	if decoded.Bounds() != img.Bounds() {
		t.Fatalf("decoded bounds %v != original %v", decoded.Bounds(), img.Bounds())
	}

	// write artifact for visual check
	artifactDir := os.Getenv("DG_SCENE_ARTIFACT_DIR")
	if artifactDir == "" {
		artifactDir = t.TempDir()
	}
	artifactPath := filepath.Join(artifactDir, "screenshot_test.png")
	if err := os.WriteFile(artifactPath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
}
