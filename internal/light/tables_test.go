package light_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/light"
)

// fixtureDir holds the committed probe-produced light goldens.
const fixtureDir = "../testdata/gamedata/lighttables"

// readFixture reads a committed golden TSV and decodes its decimal byte
// rows, asserting the byte count against the table's dims header.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(fixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	data, dims, err := decodeTSV(b)
	if err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	// dims is "<cols>x<rows>" from the dims header; the byte count must
	// follow the declared header, not a hardcoded literal.
	if want := dims[0] * dims[1]; len(data) != want {
		t.Fatalf("fixture %s: decoded %d bytes, dims header %dx%d declares %d", name, len(data), dims[0], dims[1], want)
	}
	return data
}

// parseDimsHeader parses a "# dims: <cols>x<rows>" header line, reporting
// whether the line carried a dims declaration.
func parseDimsHeader(line string) ([2]int, bool, error) {
	var dims [2]int
	d, ok := strings.CutPrefix(line, "# dims: ")
	if !ok {
		return dims, false, nil
	}
	parts := strings.Split(d, "x")
	if len(parts) != 2 {
		return dims, true, fmt.Errorf("dims header %q: want <cols>x<rows>", line)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return dims, true, fmt.Errorf("dims header %q: %w", line, err)
		}
		dims[i] = n
	}
	return dims, true, nil
}

// decodeTSV decodes a golden TSV ("#" header lines, then rows of decimal byte
// values, tab-separated) into raw bytes plus the dims header <cols>x<rows>.
func decodeTSV(data []byte) ([]byte, [2]int, error) {
	var (
		raw  []byte
		dims [2]int
	)
	for line := range strings.SplitSeq(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			d, isDims, err := parseDimsHeader(line)
			if err != nil {
				return nil, dims, err
			}
			if isDims {
				dims = d
			}
			continue
		}
		for field := range strings.SplitSeq(line, "	") {
			n, err := strconv.Atoi(field)
			if err != nil {
				return nil, dims, fmt.Errorf("row %q: %w", line, err)
			}
			if n < 0 || n > 255 {
				return nil, dims, fmt.Errorf("row %q: value %d out of byte range", line, n)
			}
			raw = append(raw, byte(n))
		}
	}
	if dims[0] == 0 || dims[1] == 0 {
		return nil, dims, errors.New("missing dims header")
	}
	return raw, dims, nil
}

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
	want := readFixture(t, "falloffs.tsv")
	for _, tc := range []struct {
		name  string
		dtype light.DungeonType
	}{
		{"nest", light.DTYPE_NEST},
		{"crypt", light.DTYPE_CRYPT},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got := flattenFalloffs(tc.dtype)
			// then
			if len(got) != len(want) {
				t.Fatalf("falloffs.tsv size: got %d want %d", len(got), len(want))
			}
			if d := firstDiff(got, want); d >= 0 {
				t.Fatalf("falloffs.tsv mismatch at byte %d: got 0x%02x want 0x%02x", d, got[d], want[d])
			}
		})
	}
	t.Run("standard level type uses the linear falloff", func(t *testing.T) {
		// when
		standard := flattenFalloffs(light.DTYPE_TOWN)
		nest := flattenFalloffs(light.DTYPE_NEST)
		// then
		if firstDiff(standard, nest) < 0 {
			t.Fatal("town falloffs are identical to the nest falloffs, want the linear variant")
		}
	})
}

// TestConeGolden checks the light cone tables against the committed golden.
func TestConeGolden(t *testing.T) {
	// when
	want := readFixture(t, "cone.tsv")
	_, _, cone, _, _ := light.MakeLightTables(light.DTYPE_TOWN)
	got := make([]byte, 0, len(want))
	for i := range cone {
		for j := range cone[i] {
			for k := range cone[i][j] {
				got = append(got, cone[i][j][k][:]...)
			}
		}
	}
	// then
	if len(got) != len(want) {
		t.Fatalf("cone.tsv size: got %d want %d", len(got), len(want))
	}
	if d := firstDiff(got, want); d >= 0 {
		t.Fatalf("cone.tsv mismatch at byte %d: got 0x%02x want 0x%02x", d, got[d], want[d])
	}
}

// TestLightTablesGolden checks the level-specific light tables against the committed goldens.
func TestLightTablesGolden(t *testing.T) {
	cases := []struct {
		name  string
		dtype light.DungeonType
		file  string
	}{
		{"cathedral", light.DTYPE_CATHEDRAL, "cathedral_tables.tsv"},
		{"hell", light.DTYPE_HELL, "hell_tables.tsv"},
		{"nest", light.DTYPE_NEST, "nest_tables.tsv"},
		{"crypt", light.DTYPE_CRYPT, "crypt_tables.tsv"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := readFixture(t, tc.file)
			// when
			tables, _, _, _, _ := light.MakeLightTables(tc.dtype)
			got := make([]byte, 0, len(want))
			for i := range tables {
				got = append(got, tables[i][:]...)
			}
			// then
			if len(got) != len(want) {
				t.Fatalf("%s size: got %d want %d", tc.file, len(got), len(want))
			}
			if d := firstDiff(got, want); d >= 0 {
				t.Fatalf("%s mismatch at byte %d: got 0x%02x want 0x%02x", tc.file, d, got[d], want[d])
			}
		})
	}
}

// TestDecodeTSV pins the decimal TSV decoder used for the committed goldens.
func TestDecodeTSV(t *testing.T) {
	for _, tc := range []struct {
		name string
		tsv  string
		want []byte
	}{
		{
			name: "one row",
			tsv: "# table: t\n# dims: 16x1\n" +
				"# source: devilutionX/Source/engine/light_tables.cpp\n" +
				"0	15	255	3	4	5	6	7	8	9	10	11	12	13	14	15\n",
			want: []byte{0, 15, 255, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			name: "multiple rows",
			tsv:  "# dims: 2x2\n0	1\n2	3\n",
			want: []byte{0, 1, 2, 3},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got, dims, err := decodeTSV([]byte(tc.tsv))
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			// then
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("got %v want %v", got, tc.want)
			}
			if want := dims[0] * dims[1]; len(got) != want {
				t.Fatalf("byte count %d does not match dims header %dx%d", len(got), dims[0], dims[1])
			}
		})
	}
	for _, tc := range []struct {
		name string
		tsv  string
	}{
		{"invalid int", "# dims: 2x1\n0	zz\n"},
		{"value above 255", "# dims: 2x1\n0	256\n"},
		{"missing dims header", "0	1\n"},
		{"malformed dims header", "# dims: two\n0	1\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// when
			_, _, err := decodeTSV([]byte(tc.tsv))
			// then
			if err == nil {
				t.Fatal("want error, got nil")
			}
		})
	}
}

// TestFullyLitFullyDark checks the fully lit / fully dark pointer contract per dungeon type.
func TestFullyLitFullyDark(t *testing.T) {
	t.Run("town", func(t *testing.T) {
		// when
		_, _, _, fullyLit, fullyDark := light.MakeLightTables(light.DTYPE_TOWN)
		// then
		if fullyLit == nil {
			t.Error("town fullyLit is nil, want non-nil")
		}
		if fullyDark == nil {
			t.Error("town fullyDark is nil, want non-nil")
		}
	})
	t.Run("hell fullyLit nil", func(t *testing.T) {
		// when
		_, _, _, fullyLit, _ := light.MakeLightTables(light.DTYPE_HELL)
		// then
		if fullyLit != nil {
			t.Error("hell fullyLit is non-nil, want nil")
		}
	})
	t.Run("nest fullyDark nil", func(t *testing.T) {
		// when
		_, _, _, _, fullyDark := light.MakeLightTables(light.DTYPE_NEST)
		// then
		if fullyDark != nil {
			t.Error("nest fullyDark is non-nil, want nil")
		}
	})
}
