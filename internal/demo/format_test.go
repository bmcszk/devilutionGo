package demo_test

import (
	"bytes"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/demo"
)

// Reference: diasurgical/DevilutionX Source/engine/demomode.cpp (v3.1.0),
// cited inline per assertion. The reader/writer pair must be byte-symmetric:
// record N synthetic events → encode → decode → identical event sequence.

func encodeStream(t *testing.T, records []demo.Record) []byte {
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

func writeReadRecords(t *testing.T, records []demo.Record) []demo.Record {
	t.Helper()
	buf := bytes.NewBuffer(encodeStream(t, records))
	if _, err := demo.ReadHeader(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	r := demo.NewReader(bytes.NewReader(buf.Bytes()[5:]))
	var got []demo.Record
	for {
		rec, err := r.NextRecord()
		if err != nil {
			t.Fatalf("NextRecord: %v", err)
		}
		if rec == nil {
			return got
		}
		got = append(got, rec)
	}
}

func TestRoundtripEventSequence(t *testing.T) {
	records := []demo.Record{
		demo.TickRecord{Progress: 127},
		demo.MouseMotionRecord{X: 320, Y: 240},
		demo.MouseButtonRecord{Button: 1, X: 100, Y: 50, Mod: 2, Down: true},
		demo.MouseButtonRecord{Button: 1, X: 100, Y: 50, Mod: 2},
		demo.MouseWheelRecord{X: -1, Y: 3, Mod: 0x1000},
		demo.KeyRecord{Sym: 0x4000005A, Mod: 1, Down: true},
		demo.KeyRecord{Sym: 0x4000005A, Mod: 1},
		demo.QuitRecord{},
		demo.TickRecord{Progress: 0},
	}
	got := writeReadRecords(t, records)
	if !recordsEqual(got, records) {
		t.Fatalf("roundtrip mismatch:\n got %v\nwant %v", got, records)
	}
}

func TestRoundtripRenderFrame(t *testing.T) {
	records := []demo.Record{
		demo.TickRecord{Progress: 64},
		demo.RenderRecord{Progress: 7},
		demo.RenderRecord{Progress: 0},
	}
	got := writeReadRecords(t, records)
	if !recordsEqual(got, records) {
		t.Fatalf("roundtrip mismatch:\n got %v\nwant %v", got, records)
	}
}

func TestVersionByteRejection(t *testing.T) {
	// demomode.cpp:519-523 rejects files whose version byte exceeds Version=3;
	// a v2 stream must be rejected the same way.
	save, err := demo.ReadHeader(bytes.NewReader([]byte{2, 0, 0, 0, 0}))
	if err == nil {
		t.Fatal("expected error for version 2 header, got nil")
	}
	if save != 0 {
		t.Fatalf("save = %d on error path, want 0", save)
	}
}

func TestEmptyStreamEOF(t *testing.T) {
	stream := []byte{demo.Version, 0, 0, 0, 0, 0}
	if _, err := demo.ReadHeader(bytes.NewReader(stream)); err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	r := demo.NewReader(bytes.NewReader(stream[5:]))
	rec, err := r.NextRecord()
	if err != nil || rec != nil {
		t.Fatalf("expected clean EOF, got rec=%v err=%v", rec, err)
	}
}

func TestStreamLengthMatchesSumOfParts(t *testing.T) {
	records := []demo.Record{
		demo.TickRecord{Progress: 1},
		demo.MouseMotionRecord{X: 1, Y: 2},
		demo.KeyRecord{Sym: 7, Mod: 0, Down: true},
	}
	want := 5 + 2 + 6 + 8
	// Header (demomode.cpp:519-523: version byte + LE32 gSaveNumber) plus
	// per-record payload sizes from demomode.cpp:554-583:
	// tick 2, motion 6, key 8.
	if got := len(encodeStream(t, records)); got != want {
		t.Fatalf("stream length = %d, want %d", got, want)
	}
}

func TestTruncatedStream(t *testing.T) {
	// A payload cut short must surface an error, not a silent partial record
	// (upstream reads without length checks; this port must not).
	var buf bytes.Buffer
	if err := demo.WriteHeader(&buf, 0); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	buf.Write([]byte{9, 0, 0x01})
	_, err := demo.ReadHeader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	r := demo.NewReader(bytes.NewReader(buf.Bytes()[5:]))
	if _, err := r.NextRecord(); err == nil {
		t.Fatal("expected error for truncated event payload, got nil")
	}
}

func recordsEqual(got, want []demo.Record) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
