package light_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/light"
)

// fixtureDir holds the committed probe-produced light goldens.
const fixtureDir = "../../test/fixtures/generated/lighttables"

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(fixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return b
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
	want := readFixture(t, "falloffs.bin")
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
				t.Fatalf("falloffs.bin size: got %d want %d", len(got), len(want))
			}
			if d := firstDiff(got, want); d >= 0 {
				t.Fatalf("falloffs.bin mismatch at byte %d: got 0x%02x want 0x%02x", d, got[d], want[d])
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
	want := readFixture(t, "cone.bin")
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
		t.Fatalf("cone.bin size: got %d want %d", len(got), len(want))
	}
	if d := firstDiff(got, want); d >= 0 {
		t.Fatalf("cone.bin mismatch at byte %d: got 0x%02x want 0x%02x", d, got[d], want[d])
	}
}

// TestLightTablesGolden checks the level-specific light tables against the committed goldens.
func TestLightTablesGolden(t *testing.T) {
	cases := []struct {
		name  string
		dtype light.DungeonType
		file  string
	}{
		{"cathedral", light.DTYPE_CATHEDRAL, "cathedral_tables.bin"},
		{"hell", light.DTYPE_HELL, "hell_tables.bin"},
		{"nest", light.DTYPE_NEST, "nest_tables.bin"},
		{"crypt", light.DTYPE_CRYPT, "crypt_tables.bin"},
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
