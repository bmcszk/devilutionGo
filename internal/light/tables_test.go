package light_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/light"
)

// fixtureDir holds the committed probe-produced light goldens.
const fixtureDir = "../testdata/gamedata/lighttables"

// hexColumnWidth is the width of the hex column of an xxd dump line: 16
// bytes grouped in pairs (8 groups of 4 hex chars plus 7 separating spaces).
const hexColumnWidth = 39

// readFixture reads a committed golden and decodes its xxd text dump.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(fixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	raw, err := decodeHexDump(b)
	if err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return raw
}

// decodeHexDump decodes an xxd text dump ("00000000: 0f0f 0f0f ...  ....")
// back into the raw bytes it holds.
func decodeHexDump(data []byte) ([]byte, error) {
	var raw []byte
	for line := range strings.SplitSeq(string(data), "\n") {
		_, after, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		after = strings.TrimLeft(after, " ")
		if len(after) > hexColumnWidth {
			after = after[:hexColumnWidth] // drop the trailing ASCII column
		}
		decoded, err := hex.DecodeString(strings.ReplaceAll(after, " ", ""))
		if err != nil {
			return nil, fmt.Errorf("line %q: %w", line, err)
		}
		raw = append(raw, decoded...)
	}
	return raw, nil
}

// firstDiff returns the index of the first differing byte, -1 when equal.
func firstDiff(got, want []byte) int {
	n := len(got)
	if len(want) < n {
		n = len(want)
	}
	for i := 0; i < n; i++ {
		if got[i] != want[i] {
			return i
		}
	}
	if len(got) != len(want) {
		return n
	}
	return -1
}

// flattenFalloffs returns the falloff tables as a flat byte slice.
func flattenFalloffs(dtype light.DungeonType) []byte {
	_, falloffs, _, _, _ := light.MakeLightTables(dtype)
	out := make([]byte, 0, len(falloffs)*128)
	for i := range falloffs {
		out = append(out, falloffs[i][:]...)
	}
	return out
}

// TestFalloffsGolden checks the light falloff tables against the committed golden.
//
// Upstream keeps one global LightFalloffs array that MakeLightTable() regenerates per
// level type (devilutionX Source/engine/light_tables.cpp:79-101): nest/crypt use the
// quadratic falloff with over exposure, every other level type uses the linear falloff.
// The committed golden holds the nest/crypt variant (identical for both types).
func TestFalloffsGolden(t *testing.T) {
	want := readFixture(t, "falloffs.hex")
	for _, tc := range []struct {
		name  string
		dtype light.DungeonType
	}{
		{"nest", light.DTYPE_NEST},
		{"crypt", light.DTYPE_CRYPT},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := flattenFalloffs(tc.dtype)
			if len(got) != len(want) {
				t.Fatalf("falloffs.hex size: got %d want %d", len(got), len(want))
			}
			if d := firstDiff(got, want); d >= 0 {
				t.Fatalf("falloffs.hex mismatch at byte %d: got 0x%02x want 0x%02x", d, got[d], want[d])
			}
		})
	}
	t.Run("standard level type uses the linear falloff", func(t *testing.T) {
		standard := flattenFalloffs(light.DTYPE_TOWN)
		nest := flattenFalloffs(light.DTYPE_NEST)
		if firstDiff(standard, nest) < 0 {
			t.Fatal("town falloffs are identical to the nest falloffs, want the linear variant")
		}
	})
}

// TestConeGolden checks the light cone tables against the committed golden.
func TestConeGolden(t *testing.T) {
	want := readFixture(t, "cone.hex")
	_, _, cone, _, _ := light.MakeLightTables(light.DTYPE_TOWN)
	got := make([]byte, 0, len(want))
	for i := range cone {
		for j := range cone[i] {
			for k := range cone[i][j] {
				got = append(got, cone[i][j][k][:]...)
			}
		}
	}
	if len(got) != len(want) {
		t.Fatalf("cone.hex size: got %d want %d", len(got), len(want))
	}
	if d := firstDiff(got, want); d >= 0 {
		t.Fatalf("cone.hex mismatch at byte %d: got 0x%02x want 0x%02x", d, got[d], want[d])
	}
}

// TestLightTablesGolden checks the level-specific light tables against the committed goldens.
func TestLightTablesGolden(t *testing.T) {
	cases := []struct {
		name  string
		dtype light.DungeonType
		file  string
	}{
		{"cathedral", light.DTYPE_CATHEDRAL, "cathedral_tables.hex"},
		{"hell", light.DTYPE_HELL, "hell_tables.hex"},
		{"nest", light.DTYPE_NEST, "nest_tables.hex"},
		{"crypt", light.DTYPE_CRYPT, "crypt_tables.hex"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := readFixture(t, tc.file)
			tables, _, _, _, _ := light.MakeLightTables(tc.dtype)
			got := make([]byte, 0, len(want))
			for i := range tables {
				got = append(got, tables[i][:]...)
			}
			if len(got) != len(want) {
				t.Fatalf("%s size: got %d want %d", tc.file, len(got), len(want))
			}
			if d := firstDiff(got, want); d >= 0 {
				t.Fatalf("%s mismatch at byte %d: got 0x%02x want 0x%02x", tc.file, d, got[d], want[d])
			}
		})
	}
}

// TestDecodeHexDump pins the xxd text decoder used for the committed goldens.
func TestDecodeHexDump(t *testing.T) {
	for _, tc := range []struct {
		name string
		dump string
		want []byte
	}{
		{
			name: "full line",
			dump: "00000000: 0f0f 0f0f 0f0f 0f0f 0f0f 0f0f 0f0f 0f0f  ................\n",
			want: bytes.Repeat([]byte{0x0f}, 16),
		},
		{
			name: "partial last line",
			dump: "00000000: 3d01 0041 42                             =..AB\n",
			want: []byte{0x3d, 0x01, 0x00, 0x41, 0x42},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeHexDump([]byte(tc.dump))
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
	t.Run("invalid hex", func(t *testing.T) {
		if _, err := decodeHexDump([]byte("00000000: zzzz\n")); err == nil {
			t.Fatal("want error for invalid hex, got nil")
		}
	})
}

// TestFullyLitFullyDark checks the fully lit / fully dark pointer contract per dungeon type.
func TestFullyLitFullyDark(t *testing.T) {
	t.Run("town", func(t *testing.T) {
		_, _, _, fullyLit, fullyDark := light.MakeLightTables(light.DTYPE_TOWN)
		if fullyLit == nil {
			t.Error("town fullyLit is nil, want non-nil")
		}
		if fullyDark == nil {
			t.Error("town fullyDark is nil, want non-nil")
		}
	})
	t.Run("hell fullyLit nil", func(t *testing.T) {
		_, _, _, fullyLit, _ := light.MakeLightTables(light.DTYPE_HELL)
		if fullyLit != nil {
			t.Error("hell fullyLit is non-nil, want nil")
		}
	})
	t.Run("nest fullyDark nil", func(t *testing.T) {
		_, _, _, _, fullyDark := light.MakeLightTables(light.DTYPE_NEST)
		if fullyDark != nil {
			t.Error("nest fullyDark is non-nil, want nil")
		}
	})
}
