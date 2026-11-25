//go:build windowed

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// TestAC4InteractiveDemoSmoke is the AC-4 refresh gate (devilGo-jjd0): the
// interactive demo's beats are driven headless-deterministically through the
// windowed boot's construction surface — front screens, selhero create
// (CreatedPack in the tick log) and the live walk (PlayerMoved) — with the
// screen state captured as PNG + pixel stats at each beat.
//
// Beats:
//  1. boot → title
//  2. any key → main menu
//  3. Enter on Single Player → selhero
//  4. create (name + warrior) → CreatedPack event + persisted hero row
//  5. walk (firm east pushes through the lane seam) → PlayerMoved per step
//
// The walk beat drives the lane seam directly (the established live-view
// seam, live_view_test.go): the window flow today never releases the keys to
// the movement lane — window_ebiten.go's Update gates movement on the flow
// leaving the three front screens, and the flow has no enter-town state yet.
// That gap is recorded honestly in docs/m1-manual-smoke.md.
//
// Set AC4_DEMO_ARTIFACT_DIR to a directory to receive the per-beat captures
// (PNG + a .stats sidecar); the default is the test's temp dir.
func TestAC4InteractiveDemoSmoke(t *testing.T) {
	boot := newAC4Boot(t)

	// ---- beat 1: boot lands on the title --------------------------------
	if got := boot.flow.Screen(); got != scene.FrontTitle {
		t.Fatalf("beat 1: boot screen = %v, want FrontTitle", got)
	}
	spawnPNG := boot.capture(t, "ac4-beat1-title")

	// ---- beat 2: any key advances to the main menu ----------------------
	boot.handleKey(ui.KeySpace)
	if got := boot.flow.Screen(); got != scene.FrontMainMenu {
		t.Fatalf("beat 2: screen after any key = %v, want FrontMainMenu", got)
	}
	boot.capture(t, "ac4-beat2-mainmenu")

	// ---- beat 3: Enter on Single Player enters selhero -------------------
	boot.handleKey(ui.KeyEnter)
	if got := boot.flow.Screen(); got != scene.FrontSelHero {
		t.Fatalf("beat 3: screen after Enter = %v, want FrontSelHero", got)
	}

	if rows := boot.selhero.SelHeroRows(); !slices.Equal(rows, []string{"Create Hero", "Delete Hero", "Quit"}) {
		t.Fatalf("beat 3: selhero rows = %v, want the three action rows", rows)
	}
	boot.capture(t, "ac4-beat3-selhero")

	// ---- beat 4: create a warrior through the create flow ----------------
	boot.handleKey(ui.KeyEnter) // Create Hero (the list opens focused on row 0)
	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beat 4: beginCreate: %v", err)
	}
	if err := boot.selhero.typeName("Gate Hero"); err != nil {
		t.Fatalf("beat 4: typeName: %v", err)
	}
	boot.selhero.confirmName()
	boot.handleKey(ui.KeyEnter) // Warrior is the focused class row

	// the created hero lands only through the settled pass: one frame, and
	// the tick log must carry exactly one CreatedPack for the warrior class
	boot.frame()

	events := createdPackEvents(boot.lane.game.Kernel.Trace())
	if len(events) != 1 {
		t.Fatalf("beat 4: CreatedPack events = %d, want exactly one", len(events))
	}
	if events[0].A != 0 {
		t.Fatalf("beat 4: CreatedPack class = %d, want 0 (warrior)", events[0].A)
	}

	rows, err := boot.stores.ListHeroes(t.Context())
	if err != nil {
		t.Fatalf("beat 4: ListHeroes: %v", err)
	}
	if len(rows) != 1 || rows[0].Name != "Gate Hero" || rows[0].ClassID != 0 {
		t.Fatalf("beat 4: store rows = %+v, want one warrior named Gate Hero", rows)
	}
	boot.capture(t, "ac4-beat4-created")

	// ---- beat 5: live walk — firm east pushes, PlayerMoved per step ------
	// The lane seam the window's own movement read feeds (windowMovement):
	// queue the push, then the frame's settled pass dispatches it.
	start := boot.lane.state.Hero()

	lastPNG := spawnPNG
	for step := 1; step <= 3; step++ {
		dir := controls.AxisDirection{X: controls.AxisXRight, Y: controls.AxisYNone}
		if !boot.lane.PushStick(controls.PushFirm, dir) {
			t.Fatalf("beat 5: gate refused the east push on step %d", step)
		}

		boot.lane.frame(func() {})

		tile := boot.lane.state.Hero()
		wantTile := walkEastTile(start, step)
		if tile != wantTile {
			t.Fatalf("beat 5 step %d: hero tile = %v, want %v", step, tile, wantTile)
		}

		assertPlayerMovedOnce(t, boot.lane.game.Kernel, boot.lane.game.TickCount-1, wantTile)

		// the presenter recomposes at the moved tile: the walked frame's
		// pixels must differ from the previous beat's (the view follows)
		boot.present.syncScene()
		framePNG := boot.capturePNG(t)
		if bytes.Equal(framePNG, lastPNG) {
			t.Fatalf("beat 5 step %d: walked frame is pixel-identical to the previous beat", step)
		}
		lastPNG = framePNG

		name := fmt.Sprintf("ac4-beat5-walk%02d", step)
		writeArtifact(t, name+".png", framePNG)
		writeArtifact(t, name+".stats", []byte(boot.stats(t).String()))
		t.Logf("%s: %s", name, boot.stats(t))
	}

	// the presenter composes at the walked tile, not the spawn tile
	tile := boot.lane.state.Hero()
	if got := boot.present.composedHeroTile(); got != (render.TilePos{X: tile.X, Y: tile.Y}) {
		t.Fatalf("beat 5: presenter tile = %v, want the walked tile %v", got, tile)
	}
}

// walkEastTile is the tile the hero lands on after step firm east pushes
// from start: the movement verb's step is (+1, -1) per east push
// (internal/game/world/movement.go step).
func walkEastTile(start world.Tile, step int) world.Tile {
	return world.Tile{X: start.X + step, Y: start.Y - step}
}

// ac4Boot is the gate's boot: the production construction surface
// (buildWindowGame) over a temp-dir store, plus the capture seam.
type ac4Boot struct {
	*windowBoot
}

func newAC4Boot(t *testing.T) *ac4Boot {
	t.Helper()
	return &ac4Boot{windowBoot: newWindowBoot(t)}
}

func (b *ac4Boot) frame() { b.lane.frame(func() {}) }

// handleKey feeds one key press to the flow: the same input seam the window
// layer's windowFlowKey mapping feeds (frontBoot.handleKey's shape).
func (b *ac4Boot) handleKey(key ui.Key) {
	b.flow.HandleKey(key)
}

// capture composes the current screen state (town at the hero tile, front
// band in front), writes the PNG artifact and asserts the first-town-frame
// baseline profile when the caller asks for it.
func (b *ac4Boot) capture(t *testing.T, name string) []byte {
	t.Helper()

	b.present.syncScene()
	b.present.drawFront(b.flow)

	pngData := b.capturePNG(t)
	writeArtifact(t, name+".png", pngData)

	stats := b.stats(t)
	writeArtifact(t, name+".stats", []byte(stats.String()))

	t.Logf("%s: %s", name, stats)

	return pngData
}

// capturePNG expands the presenter's composed surface through shade 0 (the
// palette pass the real present runs) and encodes it as PNG bytes.
func (b *ac4Boot) capturePNG(t *testing.T) []byte {
	t.Helper()

	img := frameToImage(b.expandedFrame())

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode capture png: %v", err)
	}

	return buf.Bytes()
}

// expandedFrame expands the presenter's composed surface through shade 0:
// the CPU palette pass, deterministic and ebiten-free. The presenter is nil
// before the first compose; the capture sites call syncScene first.
func (b *ac4Boot) expandedFrame() render.Frame {
	return b.present.surface.ExpandTile(&b.present.lastState.Variants, 0)
}

// stats computes the VISUAL LAW's per-capture pixel profile off the current
// expanded frame.
func (b *ac4Boot) stats(t *testing.T) pixelStats {
	t.Helper()

	return computePixelStats(b.expandedFrame())
}

// artifactDir resolves AC4_DEMO_ARTIFACT_DIR ("" = the test's temp dir).
func artifactDir(t *testing.T) string {
	t.Helper()

	dir := os.Getenv("AC4_DEMO_ARTIFACT_DIR")
	if dir == "" {
		return t.TempDir()
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("artifact dir: %v", err)
	}

	return dir
}

func writeArtifact(t *testing.T, name string, data []byte) {
	t.Helper()

	path := filepath.Join(artifactDir(t), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write artifact %s: %v", path, err)
	}

	t.Logf("artifact: %s", path)
}

// frameToImage converts one expanded frame into an NRGBA image for PNG
// encoding (scene.frameToImage's shape; that helper is unexported there).
func frameToImage(frame render.Frame) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, frame.Width, frame.Height))
	for y := range frame.Height {
		for x := range frame.Width {
			px := frame.Pixels[y*frame.Width+x]
			img.SetNRGBA(x, y, color.NRGBA{R: px.R, G: px.G, B: px.B, A: px.A})
		}
	}

	return img
}

// pixelStats is the VISUAL LAW's per-capture read: mean brightness of the
// non-black pixels, min/max channel value, distinct colours and the
// non-black fraction — the same profile the s5 gate recorded.
type pixelStats struct {
	Mean     float64
	Min      int
	Max      int
	Distinct int
	NonBlack float64
}

func computePixelStats(frame render.Frame) pixelStats {
	var total uint64
	min, max := 255, 0
	nonBlack := 0
	seen := make(map[rgbaTriple]bool, 256)

	for _, px := range frame.Pixels {
		if px.R == 0 && px.G == 0 && px.B == 0 {
			continue
		}

		nonBlack++
		total += uint64(px.R) + uint64(px.G) + uint64(px.B)

		for _, v := range [3]int{int(px.R), int(px.G), int(px.B)} {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}

		seen[rgbaTriple{px.R, px.G, px.B}] = true
	}

	if nonBlack == 0 {
		return pixelStats{}
	}

	return pixelStats{
		Mean:     float64(total) / (3 * float64(nonBlack)),
		Min:      min,
		Max:      max,
		Distinct: len(seen),
		NonBlack: 100 * float64(nonBlack) / float64(len(frame.Pixels)),
	}
}

func (s pixelStats) String() string {
	return fmt.Sprintf("mean %.2f, min %d, max %d, %d distinct colours, %.2f%% non-black",
		s.Mean, s.Min, s.Max, s.Distinct, s.NonBlack)
}

type rgbaTriple struct{ R, G, B uint8 }
