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

// TestScreenshotCaptureWriter pins the headless half of the capture seam: the
// composed town frame expands through the palette, encodes to PNG and decodes
// back at the same bounds with real content. The pixel-reading presenter path
// (Target().ReadPixels) cannot be pinned here — ebiten only allows ReadPixels
// inside RunGame (ui.go:164-167), so the windowed tag alone cannot satisfy the
// guard, and production capture reads the screen Image inside draw instead
// (~/.hermes/skills/dev/devilutiongo-lane-ops/references/visual-verification.md).
func TestScreenshotCaptureWriter(t *testing.T) {
	// given a composed town scene from the retail archives
	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES")
	}

	composed, err := scene.ComposeTownFrame(dirs)
	if err != nil {
		t.Fatalf("ComposeTownFrame: %v", err)
	}

	// when the frame expands through the palette and round-trips as PNG
	img := frameImage(composed.Surface.ExpandTile(&composed.State.Variants, 0))

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}

	decoded, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}

	// then the decode keeps the frame bounds and carries a non-flat frame
	if decoded.Bounds() != img.Bounds() {
		t.Fatalf("decoded bounds %v != expanded frame %v", decoded.Bounds(), img.Bounds())
	}

	if got := scene.DistinctColors(composed); got < 64 {
		t.Fatalf("composed frame has %d distinct palette colours, want >= 64 (a real town frame is richly coloured)", got)
	}

	// the plain compose carries no boot brightness, so the gate is coverage, not
	// tone (the mean-brightness gate lives in TestWarriorStandFrameOverTown)
	if _, nonBlack := scene.MeanBrightness(composed); nonBlack < 0.5 {
		t.Fatalf("only %.1f%% of the frame is non-black, want >= 50%% (a blank capture)", nonBlack*100)
	}

	// and the artifact lands for the parent's visual check
	artifactDir := os.Getenv("DG_SCENE_ARTIFACT_DIR")
	if artifactDir == "" {
		artifactDir = t.TempDir()
	}

	artifactPath := filepath.Join(artifactDir, "screenshot_test.png")
	if err := os.WriteFile(artifactPath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
}

// frameImage lays the expanded frame into an encoder-ready image.
func frameImage(frame render.Frame) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, frame.Width, frame.Height))

	for y := range frame.Height {
		for x := range frame.Width {
			c := frame.Pixels[y*frame.Width+x]
			img.SetNRGBA(x, y, color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A})
		}
	}

	return img
}
