package demo_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/demo"
)

// Reference: diasurgical/DevilutionX Source/engine/demomode.cpp (v3.1.0),
// cited inline per assertion. The reader/writer pair must be byte-symmetric:
// record N synthetic events → encode → decode → identical event sequence.

func TestRoundtripEventSequence(t *testing.T) {
	// given
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
	// when
	got := writeReadRecords(t, records)
	// then
	if !slices.Equal(got, records) {
		t.Fatalf("roundtrip mismatch:\n got %v\nwant %v", got, records)
	}
}

func TestRoundtripRenderFrame(t *testing.T) {
	// given
	records := []demo.Record{
		demo.TickRecord{Progress: 64},
		demo.RenderRecord{Progress: 7},
		demo.RenderRecord{Progress: 0},
	}
	// when
	got := writeReadRecords(t, records)
	// then
	if !slices.Equal(got, records) {
		t.Fatalf("roundtrip mismatch:\n got %v\nwant %v", got, records)
	}
}

func TestVersionByteRejection(t *testing.T) {
	// given
	// demomode.cpp:519-523 rejects files whose version byte exceeds demo.Version=3;
	// a v2 stream must be rejected the same way.
	// when
	save, err := demo.ReadHeader(bytes.NewReader([]byte{2, 0, 0, 0, 0}))
	// then
	if err == nil {
		t.Fatal("expected error for version 2 header, got nil")
	}

	if save != 0 {
		t.Fatalf("save = %d on error path, want 0", save)
	}
}

func TestEmptyStreamEOF(t *testing.T) {
	// given
	stream := []byte{demo.Version, 0, 0, 0, 0, 0}
	if _, err := demo.ReadHeader(bytes.NewReader(stream)); err != nil {
		t.Fatalf("readHeader: %v", err)
	}

	r := demo.NewReader(bytes.NewReader(stream[5:]))
	// when
	rec, err := r.NextRecord()
	// then
	if err != nil || rec != nil {
		t.Fatalf("expected clean EOF, got rec=%v err=%v", rec, err)
	}
}

func TestStreamLengthMatchesSumOfParts(t *testing.T) {
	// given
	records := []demo.Record{
		demo.TickRecord{Progress: 1},
		demo.MouseMotionRecord{X: 1, Y: 2},
		demo.KeyRecord{Sym: 7, Mod: 0, Down: true},
	}
	want := 5 + 2 + 6 + 8
	// Header (demomode.cpp:519-523: version byte + LE32 gSaveNumber) plus
	// per-record payload sizes from demomode.cpp:554-583:
	// tick 2, motion 6, key 8.
	// when
	got := len(encodeStream(t, records))
	// then
	if got != want {
		t.Fatalf("stream length = %d, want %d", got, want)
	}
}

func TestTruncatedStream(t *testing.T) {
	// given
	// A payload cut short must surface an error, not a silent partial record
	// (upstream reads without length checks; this port must not).
	var buf bytes.Buffer
	if err := demo.WriteHeader(&buf, 0); err != nil {
		t.Fatalf("writeHeader: %v", err)
	}

	buf.Write([]byte{9, 0, 0x01})

	if _, err := demo.ReadHeader(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("readHeader: %v", err)
	}

	r := demo.NewReader(bytes.NewReader(buf.Bytes()[5:]))
	// when
	_, err := r.NextRecord()
	// then
	if err == nil {
		t.Fatal("expected error for truncated event payload, got nil")
	}
}

func encodeStream(t *testing.T, records []demo.Record) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := demo.WriteHeader(&buf, 0); err != nil {
		t.Fatalf("writeHeader: %v", err)
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
		t.Fatalf("readHeader: %v", err)
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
