package demo_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/demo"
)

// Reference: diasurgical/DevilutionX Source/engine/demomode.cpp:157-187 — the
// settings block sits between the save number and the first record — and
// :589-596, where a tag at or above MinCustomEvent is recorded without payload.

func TestReadSettingsParsesBlock(t *testing.T) {
	// given a settings block as upstream writes it: LE16 width, LE16 height,
	// 17 option bytes, 6 potion counters (demomode.cpp:162-185)
	block := make([]byte, 0, 27)

	block = binary.LittleEndian.AppendUint16(block, 768)
	block = binary.LittleEndian.AppendUint16(block, 480)

	options := make([]byte, 17)
	options[7] = 1 // autoEquipWeapons, the byte the committed tape sets
	block = append(block, options...)
	block = append(block, 1, 2, 3, 4, 5, 6)

	// when it is read
	settings, err := demo.ReadSettings(bytes.NewReader(block))
	if err != nil {
		t.Fatalf("ReadSettings: %v", err)
	}

	// then the surface, the options and the potions come back verbatim
	if settings.Width != 768 || settings.Height != 480 {
		t.Fatalf("surface = %dx%d, want 768x480", settings.Width, settings.Height)
	}

	if settings.Options[7] != 1 {
		t.Fatalf("option 7 = %d, want 1 (autoEquipWeapons)", settings.Options[7])
	}

	if settings.Potions != [6]byte{1, 2, 3, 4, 5, 6} {
		t.Fatalf("potions = %v, want 1..6", settings.Potions)
	}
}

func TestPlaybackUpstreamFramingSkipsSettings(t *testing.T) {
	// given an upstream stream: the 5-byte header, the settings block and one
	// tick record at offset 32, exactly like the committed demo tape
	stream := upstreamStream(t, demo.TickRecord{Progress: 0x80})

	// when it replays with the upstream framing
	spy := &playbackSpy{}

	if err := demo.NewPlaybackFramed(bytes.NewReader(stream), spy, demo.UpstreamFraming).Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// then the settings block produced no records: only the tick was delivered
	if len(spy.calls) != 1 || spy.calls[0] != "tick" {
		t.Fatalf("calls = %v, want exactly one tick", spy.calls)
	}
}

func TestPlaybackCustomEventDoesNotBreakTheStream(t *testing.T) {
	// given a recording that carries an SDL custom event (tag 64) before a tick
	stream := customEventStream(t)

	// when it replays into a plain consumer
	spy := &playbackSpy{}

	if err := demo.NewPlayback(bytes.NewReader(stream), spy).Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// then the custom event is not an input and only the tick arrived
	if len(spy.calls) != 1 || spy.calls[0] != "tick" {
		t.Fatalf("calls = %v, want exactly one tick", spy.calls)
	}
}

// customEventStream encodes a custom event followed by a tick, both with
// payloads, in this engine's framing.
func customEventStream(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer

	if err := demo.WriteHeader(&buf, 0); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}

	w := demo.NewWriter(&buf)
	records := []demo.Record{
		demo.CustomRecord{Type: demo.MinCustomEvent, Progress: 17},
		demo.TickRecord{Progress: 0x80},
	}

	for _, rec := range records {
		if err := w.WriteRecord(rec); err != nil {
			t.Fatalf("WriteRecord %T: %v", rec, err)
		}
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	return buf.Bytes()
}

// upstreamStream builds a devilutionX-framed stream: version, save number, the
// settings block, then the given records.
func upstreamStream(t *testing.T, records ...demo.Record) []byte {
	t.Helper()

	var buf bytes.Buffer

	header := []byte{demo.Version, 0, 0, 0, 0}
	if _, err := buf.Write(header); err != nil {
		t.Fatalf("write header: %v", err)
	}

	settings := make([]byte, 27)
	binary.LittleEndian.PutUint16(settings[0:], 768)
	binary.LittleEndian.PutUint16(settings[2:], 480)

	if _, err := buf.Write(settings); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	w := demo.NewWriter(&buf)
	for _, rec := range records {
		if err := w.WriteRecord(rec); err != nil {
			t.Fatalf("WriteRecord %T: %v", rec, err)
		}
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	return buf.Bytes()
}
