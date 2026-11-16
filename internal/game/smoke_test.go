package game_test

import (
	"context"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// smokeTickDelay is gnTickDelay, diablo.cpp:146 (50ms, 20 ticks per second);
// the unexported constant is not visible from the external test package.
const smokeTickDelay = 50 * time.Millisecond

// captureHandler is a slog handler hook that records every record it sees so a
// test can assert on the structured log output.
type captureHandler struct {
	records []slog.Record
	attrs   map[string]any
}

func TestSmokeBootLogsTicks(t *testing.T) {
	// given
	hook := newCaptureHandler()
	logger := slog.New(hook)

	h, err := game.NewSmokeHarness(logger)
	if err != nil {
		t.Fatalf("new smoke harness: %v", err)
	}
	defer func() { _ = h.Close() }()

	// when
	for range 60 {
		h.Clock().NowValue += smokeTickDelay
		h.Game().Update(smokeTickDelay)
	}

	ticks := h.Shutdown()

	// then
	assertBootLine(t, hook)
	assertPassLines(t, hook, 60)
	assertShutdownLine(t, hook, ticks)
}

func TestSmokeInjectedClickMovesHero(t *testing.T) {
	// given
	hook := newCaptureHandler()

	h, err := game.NewSmokeHarness(slog.New(hook))
	if err != nil {
		t.Fatalf("new smoke harness: %v", err)
	}
	defer func() { _ = h.Close() }()

	h.RunTicks(10)

	// the fixed smoke camera: one 64px tile per grid tile, world at origin.
	start := h.State().Hero()
	target := world.Tile{X: start.X + 1, Y: start.Y}

	// when
	h.Click(int(target.X)*64+32, int(target.Y)*64+32)

	h.RunTicks(10)

	// then
	if got := h.State().Hero(); got != target {
		t.Fatalf("hero at %v, want %v", got, target)
	}

	assertClickLogged(t, hook, "PlayerMove", "accepted")
}

func TestSmokeInjectedClickOnUnwalkableRefused(t *testing.T) {
	// given
	hook := newCaptureHandler()

	h, err := game.NewSmokeHarness(slog.New(hook))
	if err != nil {
		t.Fatalf("new smoke harness: %v", err)
	}
	defer func() { _ = h.Close() }()

	h.RunTicks(10)

	start := h.State().Hero()
	target := world.Tile{X: start.X, Y: start.Y + 1}

	grid := h.State().Grid()
	grid.DPiece[target.Y][target.X] = 0
	grid.SOLData[0] = world.TileSolid

	// when
	h.Click(int(target.X)*64+32, int(target.Y)*64+32)

	h.RunTicks(10)

	// then
	if got := h.State().Hero(); got != start {
		t.Fatalf("hero moved to %v, want unchanged %v", got, start)
	}

	assertClickLogged(t, hook, "PlayerMove", "refused")
}

func TestSmokeFramePNGArtifact(t *testing.T) {
	// given
	// a driver (the just smoke recipe) presets SMOKE_ARTIFACT_DIR; only a bare
	// go test run falls back to a temp dir, so the recipe's env survives.
	outDir, preset := os.LookupEnv("SMOKE_ARTIFACT_DIR")
	if !preset {
		outDir = t.TempDir()
		t.Setenv("SMOKE_ARTIFACT_DIR", outDir)
	}

	hook := newCaptureHandler()

	h, err := game.NewSmokeHarness(slog.New(hook))
	if err != nil {
		t.Fatalf("new smoke harness: %v", err)
	}
	defer func() { _ = h.Close() }()

	h.RunTicks(10)

	start := h.State().Hero()
	target := world.Tile{X: start.X + 1, Y: start.Y}
	h.Click(int(target.X)*64+32, int(target.Y)*64+32)
	h.RunTicks(10)

	// when
	path, err := testsupport.ComposeFramePNG(h.State().Hero(), h.State().Grid(), nil, testsupport.FrameBefore)
	if err != nil {
		t.Fatalf("write frame png: %v", err)
	}

	if filepath.Dir(path) != outDir {
		t.Fatalf("artifact at %s, want under %s", path, outDir)
	}

	// then
	assertFramePNG(t, path, target)
}

// assertFramePNG decodes the artifact and pins its dimensions and palette use.
func assertFramePNG(t *testing.T, path string, hero world.Tile) {
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

	const frameW, frameH = 20 * 64, 15 * 64
	if got := img.Bounds(); got.Dx() != frameW || got.Dy() != frameH {
		t.Fatalf("png bounds %v, want %dx%d", got, frameW, frameH)
	}

	floor := img.At(0, 0)
	heroPixel := img.At(hero.X*64+32, hero.Y*64+32)
	if floor == heroPixel {
		t.Fatalf("floor and hero share color %v, want distinct", floor)
	}
}

// assertClickLogged checks a click line names the command kind and gate verdict.
func assertClickLogged(t *testing.T, hook *captureHandler, kind, verdict string) {
	t.Helper()

	for _, r := range hook.records {
		attrs := recordAttrs(r)

		if !isClickLine(r.Message) {
			continue
		}

		if attrs["kind"] == kind && attrs["verdict"] == verdict {
			return
		}
	}

	t.Fatalf("no click log line kind=%s verdict=%s", kind, verdict)
}

// isClickLine reports whether the message is a click verdict line.
func isClickLine(msg string) bool {
	return msg == "click submitted" || msg == "click verdict"
}

// recordAttrs flattens one record's attributes into a map.
func recordAttrs(r slog.Record) map[string]any {
	attrs := map[string]any{}
	r.Attrs(func(a slog.Attr) bool { attrs[a.Key] = a.Value.Any(); return true })

	return attrs
}

// assertBootLine checks the boot line's wiring fields and timestamp.
func assertBootLine(t *testing.T, hook *captureHandler) {
	t.Helper()

	boot := recordJSON(t, hook, "smoke boot")
	if boot["kernel"] != "wired" || boot["movement"] != "wired" {
		t.Fatalf("boot line missing wiring fields: %v", boot)
	}

	if boot["tick"] == nil {
		t.Fatalf("boot line missing tick field: %v", boot)
	}

	if boot["time"] == "" {
		t.Fatalf("boot line missing timestamp: %v", boot)
	}
}

// assertPassLines checks the per-tick kernel pass summaries.
func assertPassLines(t *testing.T, hook *captureHandler, want int) {
	t.Helper()

	passLines := 0

	for _, r := range hook.records {
		if r.Message != "kernel pass" {
			continue
		}

		passLines++

		if r.Time.IsZero() {
			t.Fatal("pass line missing timestamp")
		}
	}

	if passLines < want {
		t.Fatalf("got %d per-tick pass lines, want >= %d", passLines, want)
	}
}

// assertShutdownLine checks the shutdown line names the settled tick count.
func assertShutdownLine(t *testing.T, hook *captureHandler, ticks uint32) {
	t.Helper()

	shutdown := recordJSON(t, hook, "smoke shutdown")
	if ticks == 0 || shutdown["ticks"] == nil {
		t.Fatalf("shutdown line missing ticks field: %v (ticks=%d)", shutdown, ticks)
	}
}

func newCaptureHandler() *captureHandler { return &captureHandler{attrs: map[string]any{}} }

func (*captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	r.Attrs(func(a slog.Attr) bool { h.attrs[a.Key] = a.Value.Any(); return true })
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

func (h *captureHandler) find(msg string) (slog.Record, bool) {
	for _, r := range h.records {
		if r.Message == msg {
			return r, true
		}
	}

	return slog.Record{}, false
}

// recordJSON renders one record's message and attrs as a field map, the shape
// the test asserts on.
func recordJSON(t *testing.T, h *captureHandler, msg string) map[string]any {
	t.Helper()

	r, ok := h.find(msg)
	if !ok {
		t.Fatalf("no log line %q among %d records", msg, len(h.records))
	}

	out := map[string]any{"time": r.Time.Format(time.RFC3339), "msg": msg}
	r.Attrs(func(a slog.Attr) bool { out[a.Key] = a.Value.Any(); return true })

	return out
}
