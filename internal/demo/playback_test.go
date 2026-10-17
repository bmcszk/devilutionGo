package demo_test

import (
	"bytes"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/demo"
)

// Reference: diasurgical/DevilutionX Source/engine/demomode.cpp.
// The playback driver must consume the same stream twice and issue an
// identical call sequence into the consumer, and must deliver every
// input event that precedes a tick record before that tick itself.

type playbackSpy struct {
	calls []string
}

func (s *playbackSpy) Tick(_ uint8) {
	s.calls = append(s.calls, "tick")
}

func (s *playbackSpy) MouseMotion(_, _ uint16) {
	s.calls = append(s.calls, "motion")
}

func (s *playbackSpy) MouseButton(_ uint8, _, _, _ uint16, down bool) {
	if down {
		s.calls = append(s.calls, "buttonDown")
		return
	}
	s.calls = append(s.calls, "buttonUp")
}

func (s *playbackSpy) MouseWheel(_, _ int16, _ uint16) {
	s.calls = append(s.calls, "wheel")
}

func (s *playbackSpy) Key(_ uint32, _ uint16, down bool) {
	if down {
		s.calls = append(s.calls, "keyDown")
		return
	}
	s.calls = append(s.calls, "keyUp")
}

func (s *playbackSpy) Quit() {
	s.calls = append(s.calls, "quit")
}

func encode(t *testing.T, records ...demo.Record) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := demo.WriteHeader(&buf, 0); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	w := demo.NewWriter(&buf)
	for i := range records {
		if err := w.WriteRecord(records[i]); err != nil {
			t.Fatalf("WriteRecord %d: %v", i, err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.Bytes()
}

func TestPlaybackSameStreamSameCallSequence(t *testing.T) {
	records := []demo.Record{
		demo.MouseMotionRecord{X: 1, Y: 2},
		demo.MouseButtonRecord{Button: 1, X: 3, Y: 4, Mod: 5, Down: true},
		demo.MouseWheelRecord{X: -2, Y: 1, Mod: 6},
		demo.KeyRecord{Sym: 9, Mod: 1, Down: true},
		demo.TickRecord{Progress: 100},
		demo.RenderRecord{Progress: 20},
		demo.KeyRecord{Sym: 9, Mod: 1},
		demo.QuitRecord{},
		demo.TickRecord{Progress: 101},
	}
	stream := encode(t, records...)

	play := func() []string {
		spy := &playbackSpy{}
		p := demo.NewPlayback(bytes.NewReader(stream), spy)
		if err := p.Run(); err != nil {
			t.Fatalf("Run: %v", err)
		}
		return spy.calls
	}

	first := play()
	second := play()

	if len(first) == 0 {
		t.Fatal("playback produced no consumer calls")
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("call %d: %q != %q", i, first[i], second[i])
		}
	}
}

func TestPlaybackOrderingAndProgress(t *testing.T) {
	stream := encode(t,
		demo.MouseMotionRecord{X: 10, Y: 20},
		demo.KeyRecord{Sym: 65, Mod: 0, Down: true},
		demo.TickRecord{Progress: 42},
		demo.RenderRecord{Progress: 8},
		demo.QuitRecord{},
		demo.TickRecord{Progress: 43},
	)
	spy := &playbackSpy{}
	p := demo.NewPlayback(bytes.NewReader(stream), spy)
	if err := p.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"motion", "keyDown", "tick", "tick", "tick", "quit", "tick"}
	if len(spy.calls) != len(want) {
		t.Fatalf("got %v, want %v", spy.calls, want)
	}
	for i := range want {
		if spy.calls[i] != want[i] {
			t.Fatalf("call %d: got %q, want %q", i, spy.calls[i], want[i])
		}
	}
}

func TestPlaybackUncappedConsumption(t *testing.T) {
	// Timedemo semantics (demomode.cpp:644-645: Timedemo bypasses the
	// wall-clock pacing branch entirely): Run must not sleep or consult any
	// clock; the whole stream drains as fast as the consumer allows.
	stream := encode(t,
		demo.TickRecord{Progress: 1},
		demo.TickRecord{Progress: 2},
		demo.TickRecord{Progress: 3},
	)
	spy := &playbackSpy{}
	p := demo.NewPlayback(bytes.NewReader(stream), spy)
	if err := p.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(spy.calls) != 3 {
		t.Fatalf("got %d calls, want 3", len(spy.calls))
	}
}

func TestPlaybackTruncatedStream(t *testing.T) {
	// Key-down tag (13) with its sym/mod payload cut off.
	stream := []byte{demo.Version, 0, 0, 0, 0, 13}
	p := demo.NewPlayback(bytes.NewReader(stream), &playbackSpy{})
	if err := p.Run(); err == nil {
		t.Fatal("expected error from truncated stream, got nil")
	}
}
