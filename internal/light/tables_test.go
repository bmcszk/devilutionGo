// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package light

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// fixtureDir holds the committed probe-produced light goldens.
const fixtureDir = "../testdata/gamedata/lighttables"

func TestFalloffsGolden(t *testing.T) {
	want := readFixture(t, "falloffs.tsv")
	for _, tc := range []struct {
		name  string
		dtype DungeonType
	}{
		{"nest", DTYPE_NEST},
		{"crypt", DTYPE_CRYPT},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expectFalloffsGolden(t, tc.dtype, want)
		})
	}
	t.Run("standard level type uses the linear falloff", func(t *testing.T) {
		expectStandardFalloffDiffers(t)
	})
}

func expectFalloffsGolden(t *testing.T, dtype DungeonType, want []byte) {
	t.Helper()
	got := flattenFalloffs(dtype)
	if len(got) != len(want) {
		t.Fatalf("falloffs.tsv size: got %d want %d", len(got), len(want))
	}
	if d := firstDiff(got, want); d >= 0 {
		t.Fatalf("falloffs.tsv mismatch at byte %d: got 0x%02x want 0x%02x", d, got[d], want[d])
	}
}

// Town falloffs use the linear variant, not the nest/crypt quadratic one.
func expectStandardFalloffDiffers(t *testing.T) {
	t.Helper()
	standard := flattenFalloffs(DTYPE_TOWN)
	nest := flattenFalloffs(DTYPE_NEST)
	if firstDiff(standard, nest) < 0 {
		t.Fatal("town falloffs are identical to the nest falloffs, want the linear variant")
	}
}

func TestConeGolden(t *testing.T) {
	want := readFixture(t, "cone.tsv")
	_, _, cone, _, _ := makeLightTables(DTYPE_TOWN)
	expectBytesGolden(t, "cone.tsv", flattenCone(cone), want)
}

func flattenCone(cone [8][8][16][16]uint8) []byte {
	got := make([]byte, 0, 256*256*8)
	for i := range cone {
		for j := range cone[i] {
			for k := range cone[i][j] {
				got = append(got, cone[i][j][k][:]...)
			}
		}
	}
	return got
}

func expectBytesGolden(t *testing.T, name string, got, want []byte) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s size: got %d want %d", name, len(got), len(want))
	}
	if d := firstDiff(got, want); d >= 0 {
		t.Fatalf("%s mismatch at byte %d: got 0x%02x want 0x%02x", name, d, got[d], want[d])
	}
}

func TestLightTablesGolden(t *testing.T) {
	cases := []struct {
		name  string
		dtype DungeonType
		file  string
	}{
		{"cathedral", DTYPE_CATHEDRAL, "cathedral_tables.tsv"},
		{"hell", DTYPE_HELL, "hell_tables.tsv"},
		{"nest", DTYPE_NEST, "nest_tables.tsv"},
		{"crypt", DTYPE_CRYPT, "crypt_tables.tsv"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := readFixture(t, tc.file)
			tables, _, _, _, _ := makeLightTables(tc.dtype)
			got := make([]byte, 0, len(want))
			for i := range tables {
				got = append(got, tables[i][:]...)
			}
			expectBytesGolden(t, tc.file, got, want)
		})
	}
}

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
			expectDecodeTSVOK(t, tc.tsv, tc.want)
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
			expectDecodeTSVErr(t, tc.tsv)
		})
	}
}

func expectDecodeTSVOK(t *testing.T, tsv string, want []byte) {
	t.Helper()
	got, dims, err := decodeTSV([]byte(tsv))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if want := dims[0] * dims[1]; len(got) != want {
		t.Fatalf("byte count %d does not match dims header %dx%d", len(got), dims[0], dims[1])
	}
}

func expectDecodeTSVErr(t *testing.T, tsv string) {
	t.Helper()
	if _, _, err := decodeTSV([]byte(tsv)); err == nil {
		t.Fatal("want error, got nil")
	}
}

func TestFullyLitFullyDark(t *testing.T) {
	t.Run("town", func(t *testing.T) {
		expectTownFullyLitAndDark(t)
	})
	t.Run("hell fullyLit nil", func(t *testing.T) {
		if _, _, _, fullyLit, _ := makeLightTables(DTYPE_HELL); fullyLit != nil {
			t.Error("hell fullyLit is non-nil, want nil")
		}
	})
	t.Run("nest fullyDark nil", func(t *testing.T) {
		if _, _, _, _, fullyDark := makeLightTables(DTYPE_NEST); fullyDark != nil {
			t.Error("nest fullyDark is non-nil, want nil")
		}
	})
}

func expectTownFullyLitAndDark(t *testing.T) {
	t.Helper()
	_, _, _, fullyLit, fullyDark := makeLightTables(DTYPE_TOWN)
	if fullyLit == nil {
		t.Error("town fullyLit is nil, want non-nil")
	}
	if fullyDark == nil {
		t.Error("town fullyDark is nil, want non-nil")
	}
}

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

func decodeTSV(data []byte) ([]byte, [2]int, error) {
	raw, dims, err := decodeTSVLines(string(data))
	if err != nil {
		return nil, dims, err
	}
	if dims[0] == 0 || dims[1] == 0 {
		return nil, dims, errors.New("missing dims header")
	}
	return raw, dims, nil
}

func decodeTSVLines(data string) ([]byte, [2]int, error) {
	var (
		raw  []byte
		dims [2]int
		err  error
	)
	for line := range strings.SplitSeq(data, "\n") {
		raw, dims, err = appendTSVLine(raw, dims, line)
		if err != nil {
			return nil, dims, err
		}
	}
	return raw, dims, nil
}

// Blank lines skipped; "#" comment lines may update dims; data rows append decoded bytes.
func appendTSVLine(raw []byte, dims [2]int, line string) ([]byte, [2]int, error) {
	trimmed := strings.TrimSpace(line)
	switch {
	case trimmed == "":
		return raw, dims, nil
	case strings.HasPrefix(trimmed, "#"):
		d, err := applyDimsHeader(dims, line)
		return raw, d, err
	default:
		row, err := parseTSVRow(line)
		if err != nil {
			return raw, dims, err
		}
		return append(raw, row...), dims, nil
	}
}

func applyDimsHeader(dims [2]int, line string) ([2]int, error) {
	d, isDims, err := parseDimsHeader(line)
	if err != nil {
		return dims, err
	}
	if isDims {
		dims = d
	}
	return dims, nil
}

func parseTSVRow(line string) ([]byte, error) {
	var raw []byte
	for field := range strings.SplitSeq(line, "\t") {
		n, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("row %q: %w", line, err)
		}
		if n < 0 || n > 255 {
			return nil, fmt.Errorf("row %q: value %d out of byte range", line, n)
		}
		raw = append(raw, byte(n))
	}
	return raw, nil
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

func flattenFalloffs(dtype DungeonType) []byte {
	_, falloffs, _, _, _ := makeLightTables(dtype)
	out := make([]byte, 0, len(falloffs)*128)
	for i := range falloffs {
		out = append(out, falloffs[i][:]...)
	}
	return out
}
