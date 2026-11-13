package render_test

import (
	"bytes"
	"encoding/hex"
	"image"
	"runtime"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// scaleGoldenMainSource is the throwaway program the scale-golden test builds:
// it presents the 640x480 parity surface at full fade, scales the presented
// frame into a 1280x960 window target through Presenter.ScaleToWindow and
// dumps both frames as hex. It runs in its own process: the loop is
// single-shot per process and the scaled readback needs a live window-size
// target.
const scaleGoldenMainSource = `package main

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

type scaleProbe struct {
	surface render.Surface
	state   render.PaletteState
	pres    *render.Presenter
	window  *ebiten.Image
	frame   int
}

func (p *scaleProbe) hex(target *ebiten.Image, w, h int) string {
	raw := make([]byte, w*h*4)
	target.ReadPixels(raw)

	return hex.EncodeToString(raw)
}

func (p *scaleProbe) Update() error {
	p.frame++
	if p.frame < 3 {
		return nil
	}

	fmt.Println("LOGICAL", p.hex(p.pres.Target(), 640, 480))

	p.pres.ScaleToWindow(p.window, 1280, 960)
	fmt.Println("SCALED", p.hex(p.window, 1280, 960))

	return ebiten.Termination
}

func (p *scaleProbe) Draw(screen *ebiten.Image) {
	// present every frame so the presenter target exists by the readback frame
	p.pres.Draw(p.surface, nil, render.KeyTile)
	p.pres.ScaleToWindow(p.window, 1280, 960)

	screen.Fill(color.RGBA{})
}

func (p *scaleProbe) Layout(_, _ int) (int, int) { return 1280, 960 }

func loadScalePalette() assets.Palette {
	body, err := os.ReadFile(filepath.Join("internal", "testdata", "render", "pal", "l1_1.pal.hex"))
	if err != nil {
		panic(err)
	}

	flat, err := hex.DecodeString(strings.Join(strings.Fields(string(body)), ""))
	if err != nil {
		panic(err)
	}

	pal, err := assets.LoadPaletteFromBytes(flat)
	if err != nil {
		panic(err)
	}

	return *pal
}

func main() {
	tables := light.MakeLightTables(light.DTYPE_CATHEDRAL)
	state := render.NewPaletteState(loadScalePalette(), tables, assets.BlendingFor(0))

	surface := render.NewSurface(640, 480)
	for i := range surface.Pixels {
		surface.Pixels[i] = uint8((i*13 + 29) % 256)
	}

	probe := &scaleProbe{
		surface: surface,
		state:   state,
		pres:    render.NewPresenter(state),
		window:  ebiten.NewImage(1280, 960),
	}

	if err := ebiten.RunGame(probe); err != nil && err != ebiten.Termination {
		fmt.Println("scale probe error:", err)
	}
}
`

// TestScaleToWindowLetterboxGeometry pins the letterbox model: at factor 2 the
// 640x480 logical frame occupies the centred 1280x960 rectangle of the window
// (SDL_LOGICAL_PRESENTATION_INTEGER_SCALE, display.cpp:780-811), and windows
// below the logical surface floor to factor 1 (display.cpp:291).
func TestScaleToWindowLetterboxGeometry(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		windowW, windowH int
		wantFactor       int
		wantRect         image.Rectangle
	}{
		{"1280x960 window", 1280, 960, 2, image.Rect(0, 0, 1280, 960)},
		{"1920x1080 window", 1920, 1080, 2, image.Rect(320, 60, 1600, 1020)},
		{"320x240 sub-logical window", 320, 240, 1, image.Rect(-160, -120, 480, 360)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// when the window geometry is computed
			factor := render.IntegerScaleFactor(tc.windowW, tc.windowH, render.WorldLogicalWidth, render.WorldLogicalHeight)
			rect := render.ScaledFrameRect(tc.windowW, tc.windowH, render.WorldLogicalWidth, render.WorldLogicalHeight)

			// then the factor and centred rect match the letterbox model
			expectScaleGeometry(t, factor, rect, tc.wantFactor, tc.wantRect)
		})
	}
}

// TestScaleToWindowGolden2x renders a known 640x480 pattern, scales it into a
// 1280x960 window target and dumps both the logical and the scaled frame as
// hex, so a failing capture is diffable byte-for-byte against the CPU oracle
// (the golden hook of the resolution AC).
func TestScaleToWindowGolden2x(t *testing.T) {
	// given
	if !gpuAvailable(t) {
		t.Skipf("window-scaling golden needs a GPU-capable run loop; %s CI has none", runtime.GOOS)
	}

	// when
	out := runCapture(t, scaleGoldenMainSource, "scale golden")

	logical, scaled := captureHexFrames(t, out)

	// then the scaled dump is the exact 2x expansion of the logical dump
	want := scaleExpansion(t, logical, 2)

	if scaled != want {
		t.Fatalf("scaled %x..., want %x...", scaled[:64], want[:64])
	}
}

// expectScaleGeometry asserts one computed window geometry.
func expectScaleGeometry(t *testing.T, factor int, rect image.Rectangle, wantFactor int, wantRect image.Rectangle) {
	t.Helper()

	if factor != wantFactor {
		t.Fatalf("factor = %d, want %d", factor, wantFactor)
	}

	if rect != wantRect {
		t.Fatalf("rect = %v, want %v", rect, wantRect)
	}
}

// expectScaleToWindow2xByteExact pins the integer window-scaling contract at
// factor 2 (display.cpp:121-126, :780-811): a known 2x2 index pattern presented
// at full fade and scaled into a 4x4 window target expands each logical pixel
// into an exact 2x2 block, with the nearest filter and no fractional scaling.
// Runs inside the shared loop because the scale draw is a GPU operation.
func expectScaleToWindow2xByteExact(t *testing.T) {
	t.Helper()

	pal := loadPalette(t, paletteFixture)
	state := render.NewPaletteState(pal, light.MakeLightTables(light.DTYPE_CATHEDRAL), assets.BlendingFor(0))

	surface := render.NewSurface(2, 2)
	surface.Set(0, 0, 7)
	surface.Set(1, 0, 13)
	surface.Set(0, 1, 0)
	surface.Set(1, 1, 42)

	full := render.NewPresenter(state)
	frame := full.Frame(surface, nil, render.KeyTile)

	// a 4x4 window over the 2x2 logical frame: factor 2, no letterbox
	window := ebiten.NewImage(4, 4)
	full.ScaleToWindow(window, 4, 4)

	// when the window pixels are read back
	raw := make([]byte, 4*4*4)
	window.ReadPixels(raw)

	// then every logical pixel is an exact 2x2 block of its own colour
	want := expandFrameRows(frame.Pixels, 2, 2, 2)

	if !bytes.Equal(raw, want) {
		t.Fatalf("2x integer scale diverges:\ngot  %x\nwant %x", raw, want)
	}
}

// captureHexFrames pulls the LOGICAL/SCALED hex payloads from a scale capture.
func captureHexFrames(t *testing.T, out string) (logical, scaled string) {
	t.Helper()

	for line := range strings.SplitSeq(out, "\n") {
		logical = cutCaptureField(logical, line, "LOGICAL ")
		scaled = cutCaptureField(scaled, line, "SCALED ")
	}

	if logical == "" || scaled == "" {
		t.Fatalf("scale capture lost its LOGICAL/SCALED dumps:\n%s", out)
	}

	return logical, scaled
}

// cutCaptureField keeps current when the line does not carry the field tag.
func cutCaptureField(current, line, tag string) string {
	rest, ok := strings.CutPrefix(line, tag)
	if !ok {
		return current
	}

	return strings.TrimSpace(rest)
}

// scaleExpansion builds the byte-exact integer-factor expansion of a raw RGBA
// dump: every logical pixel becomes an exact factor x factor block.
func scaleExpansion(t *testing.T, hexFrame string, factor int) string {
	t.Helper()

	raw, err := hex.DecodeString(hexFrame)
	if err != nil {
		t.Fatalf("decode logical frame: %v", err)
	}

	const (
		logicalW = render.WorldLogicalWidth
		logicalH = render.WorldLogicalHeight
		bpp      = 4
	)

	if len(raw) != logicalW*logicalH*bpp {
		t.Fatalf("logical frame is %d bytes, want %d", len(raw), logicalW*logicalH*bpp)
	}

	var out strings.Builder
	out.Grow(len(hexFrame) * factor * factor)

	for row := range logicalH {
		writeExpandedRow(&out, raw, row, factor)
	}

	return out.String()
}

// writeExpandedRow appends one logical row duplicated factor times, each pixel
// repeated factor times across the row.
func writeExpandedRow(out *strings.Builder, raw []byte, row, factor int) {
	for range factor {
		writeExpandedRowOnce(out, raw, row, factor)
	}
}

// writeExpandedRowOnce appends one pass over the logical row.
func writeExpandedRowOnce(out *strings.Builder, raw []byte, row, factor int) {
	const (
		logicalW = render.WorldLogicalWidth
		bpp      = 4
	)

	for col := range logicalW {
		writeExpandedTexel(out, raw, row, col, factor)
	}
}

// writeExpandedTexel appends one texel repeated factor times.
func writeExpandedTexel(out *strings.Builder, raw []byte, row, col, factor int) {
	const (
		logicalW = render.WorldLogicalWidth
		bpp      = 4
	)

	texel := raw[(row*logicalW+col)*bpp : (row*logicalW+col)*bpp+bpp]

	for range factor {
		out.WriteString(hex.EncodeToString(texel))
	}
}

// expandFrameRows expands a row-major RGBA frame by an integer factor: every
// logical pixel becomes an exact factor x factor block.
func expandFrameRows(pixels []render.RGBA, width, height, factor int) []byte {
	out := make([]byte, 0, width*height*factor*factor*4)

	for row := range height {
		for range factor {
			out = append(out, expandPixelRow(pixels, row, width, factor)...)
		}
	}

	return out
}

// expandPixelRow returns one logical row with every pixel repeated factor
// times across the row.
func expandPixelRow(pixels []render.RGBA, row, width, factor int) []byte {
	out := make([]byte, 0, width*factor*4)

	for col := range width {
		for range factor {
			pixel := pixels[row*width+col]
			out = append(out, pixel.R, pixel.G, pixel.B, pixel.A)
		}
	}

	return out
}
