package game_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/game"
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
