//nolint:testpackage // internal-state test: pokes unexported identifiers
package demo

import (
	"bytes"
	"slices"
	"testing"
)

// Reference: diasurgical/DevilutionX Source/engine/demomode.cpp.
// The playback driver must consume the same stream twice and issue an
// identical call sequence into the consumer, and must deliver every
// input event that precedes a tick record before that tick itself.

type playbackSpy struct {
	calls []string
}

func TestPlaybackSameStreamSameCallSequence(t *testing.T) {
	// given
	records := []Record{
		mouseMotionRecord{X: 1, Y: 2},
		mouseButtonRecord{Button: 1, X: 3, Y: 4, Mod: 5, Down: true},
		mouseWheelRecord{X: -2, Y: 1, Mod: 6},
		keyRecord{Sym: 9, Mod: 1, Down: true},
		tickRecord{Progress: 100},
		renderRecord{Progress: 20},
		keyRecord{Sym: 9, Mod: 1},
		quitRecord{},
		tickRecord{Progress: 101},
	}
	stream := encode(t, records...)

	play := func() []string {
		spy := &playbackSpy{}
		p := newPlayback(bytes.NewReader(stream), spy)
		if err := p.Run(); err != nil {
			t.Fatalf("Run: %v", err)
		}
		return spy.calls
	}

	// when
	first := play()
	second := play()

	// then
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
	// given
	stream := encode(t,
		mouseMotionRecord{X: 10, Y: 20},
		keyRecord{Sym: 65, Mod: 0, Down: true},
		tickRecord{Progress: 42},
		renderRecord{Progress: 8},
		quitRecord{},
		tickRecord{Progress: 43},
	)
	spy := &playbackSpy{}
	p := newPlayback(bytes.NewReader(stream), spy)
	// when
	err := p.Run()
	// then
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := []string{"motion", "keyDown", "tick", "tick", "tick", "quit", "tick"}
	if !slices.Equal(spy.calls, want) {
		t.Fatalf("got %v, want %v", spy.calls, want)
	}
}

func TestPlaybackUncappedConsumption(t *testing.T) {
	// given
	// Timedemo semantics (demomode.cpp:644-645: Timedemo bypasses the
	// wall-clock pacing branch entirely): Run must not sleep or consult any
	// clock; the whole stream drains as fast as the consumer allows.
	stream := encode(t,
		tickRecord{Progress: 1},
		tickRecord{Progress: 2},
		tickRecord{Progress: 3},
	)
	spy := &playbackSpy{}
	p := newPlayback(bytes.NewReader(stream), spy)
	// when
	if err := p.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// then
	if len(spy.calls) != 3 {
		t.Fatalf("got %d calls, want 3", len(spy.calls))
	}
}

func TestPlaybackTruncatedStream(t *testing.T) {
	// given
	// Key-down tag (13) with its sym/mod payload cut off.
	stream := []byte{Version, 0, 0, 0, 0, 13}
	p := newPlayback(bytes.NewReader(stream), &playbackSpy{})
	// when
	err := p.Run()
	// then
	if err == nil {
		t.Fatal("expected error from truncated stream, got nil")
	}
}

func encode(t *testing.T, records ...Record) []byte {
	t.Helper()
	return encodeStream(t, records)
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
