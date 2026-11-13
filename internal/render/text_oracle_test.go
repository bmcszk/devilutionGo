package render_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// wrapGoldenPath is the committed upstream-run oracle golden: the TSV dumped
// by internal/testdata/render/producer/wraporacle.cpp over the PINNED
// upstream wrap/measure algorithms (ca91d778f).
const wrapGoldenPath = "../testdata/assets/fonts/wrap_golden.tsv"

// wrapGoldenCases is the exact corpus TestWrapLayout + TestFontLineWidthParity
// pin in text_test.go; stringIdx in the golden TSV indexes this slice. MUST
// stay in the same order as kCases in
// internal/testdata/render/producer/wraporacle.cpp.
var wrapGoldenCases = []wrapGoldenCase{
	{"lw-empty", "", -1},
	{"lw-single", "\u00d0", -1},
	{"lw-two", "\u00d0\u00de", -1},
	{"lw-three", "\u00d0\u00d5\u00de", -1},
	{"lw-five", "\u00d0\u00de\u00d2\u00d8\u00e9", -1},
	{"lw-ten", "\u00d0\u00d1\u00d2\u00d3\u00d4\u00d5\u00d6\u00d7\u00d8\u00d9", -1},
	{"lw-narrow-repeat", "\u00d5\u00df\u00d9", -1},
	{"lw-mixed-ascii-blank", "ab\u00d0", -1},
	{"lw-wide-tail", "\u00d3\u00de\u00ea", -1},
	{"lw-zwsp-solo", "\u200b", -1},
	{"lw-zwsp-between", "\u200b\u00d0", -1},
	{"lw-zwsp-run", "\u200b\u200b", -1},
	{"lw-zwsp-word", "\u00d0\u200b\u00de", -1},
	{"lw-ascii-then-zwsp", "a\u200b", -1},
	{"lw-newline-solo", "\n", -1},
	{"lw-newline-first", "\n\u00d0", -1},
	{"lw-newline-last", "\u00d0\n", -1},
	{"lw-newline-middle", "\u00d0\n\u00de", -1},
	{"lw-zwsp-then-newline", "\u200b\n", -1},
	{"lw-crlf", "\u00d0\r\n", -1},
	{"fits-exactly", "\u00d0\u00de", 19},
	{"one-under", "\u00d0\u00de", 20},
	{"two-words-fit", "\u00d0 \u00de", 200},
	{"two-words-split", "\u00d0 \u00de", 10},
	{"line-one-ends-at-space", "\u00d0\u00d3 \u00de", 20},
	{"second-word-wider", "\u00d5 \u00de", 12},
	{"third-word-forces-split", "\u00d0 \u00d5 \u00de", 14},
	{"newline-two-lines", "\u00d0\n\u00de", 200},
	{"newline-then-split", "\u00d0\n\u00d0 \u00de", 10},
	{"hard-split", "\u00d0\u00d1\u00d2\u00d3", 10},
	{"hard-split-exact", "\u00d0\u00d1\u00d2", 24},
}

// wrapGoldenRow is one parsed TSV row of the upstream wrap golden.
type wrapGoldenRow struct {
	stringIdx int
	lineIdx   int
	lineWidth int
	startX    int
	lineText  string
}

// wrapGoldenCase is one corpus string the oracle feeds to both upstream and
// Go: name mirrors the sub-test names in text_test.go, text is the exact
// UTF-8 bytes, width is the wrap rect width (-1 = LineWidth-only case,
// measured whole).
type wrapGoldenCase struct {
	name  string
	text  string
	width int
}

// TestWrapLayoutOracle certifies the Go wrap layout against upstream OUTPUT:
// the committed TSV was dumped by the offline C++ probe
// (internal/testdata/render/producer/wraporacle.cpp) running the PINNED
// upstream WordWrapString + GetLineWidth algorithms (ca91d778f) over this
// corpus. Every row is asserted; a missing or short golden fails the test.
func TestWrapLayoutOracle(t *testing.T) {
	// given the decoded sheet and the committed upstream golden
	font := loadTestFont(t)
	rows := loadWrapGolden(t)

	// when every corpus string is checked against its golden rows
	assertRowsCoverCorpus(t, rows)
	grouped := groupRowsByString(t, rows)

	// then every sub-test below passes: the Go layout matches upstream row-for-row
	for stringIdx, stringRows := range grouped {
		runOracleCase(t, font, wrapGoldenCases[stringIdx], stringRows)
	}
}

// loadWrapGolden strict-loads the committed TSV: any missing or malformed
// row fails the test, never skips it.
func loadWrapGolden(t *testing.T) []wrapGoldenRow {
	t.Helper()
	data, err := os.ReadFile(wrapGoldenPath)
	if err != nil {
		t.Fatalf("read wrap golden %s: %v (producer: internal/testdata/render/producer/wraporacle.cpp, run once offline)",
			wrapGoldenPath, err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	rows := make([]wrapGoldenRow, 0, len(lines))
	for lineNum, line := range lines {
		rows = append(rows, parseWrapGoldenRow(t, line, lineNum+1))
	}
	if len(rows) == 0 {
		t.Fatalf("wrap golden %s is empty", wrapGoldenPath)
	}
	return rows
}

// parseWrapGoldenRow parses one `stringIdx lineIdx lineWidth startX lineText`
// TSV row; malformed fields fail the test.
func parseWrapGoldenRow(t *testing.T, line string, lineNum int) wrapGoldenRow {
	t.Helper()
	fields := strings.Split(line, "\t")
	if len(fields) != 5 {
		t.Fatalf("wrap golden %s line %d: %d tab fields, want 5", wrapGoldenPath, lineNum, len(fields))
	}
	row := wrapGoldenRow{lineText: unescapeGolden(fields[4])}
	for i, dest := range []*int{&row.stringIdx, &row.lineIdx, &row.lineWidth, &row.startX} {
		n, err := strconv.Atoi(fields[i])
		if err != nil {
			t.Fatalf("wrap golden %s line %d field %d: %v", wrapGoldenPath, lineNum, i, err)
		}
		*dest = n
	}
	return row
}

// unescapeGolden decodes the producer's lineText escapes (\n, \t, \\).
func unescapeGolden(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// assertRowsCoverCorpus fails when the golden does not cover every corpus
// string or carries an out-of-range index.
func assertRowsCoverCorpus(t *testing.T, rows []wrapGoldenRow) {
	t.Helper()
	grouped := groupRowsByString(t, rows)
	for i := range wrapGoldenCases {
		if len(grouped[i]) == 0 {
			t.Fatalf("golden has no rows for stringIdx %d (%s)", i, wrapGoldenCases[i].name)
		}
	}
}

// groupRowsByString groups golden rows by string index, preserving order.
func groupRowsByString(t *testing.T, rows []wrapGoldenRow) map[int][]wrapGoldenRow {
	t.Helper()
	grouped := make(map[int][]wrapGoldenRow, len(wrapGoldenCases))
	for _, row := range rows {
		if row.stringIdx < 0 || row.stringIdx >= len(wrapGoldenCases) {
			t.Fatalf("golden stringIdx %d outside corpus (0..%d)", row.stringIdx, len(wrapGoldenCases)-1)
		}
		if lineIdxGap(t, grouped[row.stringIdx], row) {
			t.Fatalf("stringIdx %d: lineIdx %d not contiguous", row.stringIdx, row.lineIdx)
		}
		grouped[row.stringIdx] = append(grouped[row.stringIdx], row)
	}
	return grouped
}

// lineIdxGap reports whether row continues the line index sequence of prev.
func lineIdxGap(_ *testing.T, prev []wrapGoldenRow, row wrapGoldenRow) bool {
	return len(prev) > 0 && row.lineIdx != prev[len(prev)-1].lineIdx+1
}

// runOracleCase asserts one corpus string's Go output against its golden
// rows: line count, per-line start X / width, and the final cursor.
func runOracleCase(t *testing.T, font *assets.Font, tc wrapGoldenCase, rows []wrapGoldenRow) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		if tc.width < 0 {
			assertLineWidthOnly(t, font, tc, rows[0])
			return
		}
		assertWrapRows(t, font, tc, rows)
	})
}

// assertLineWidthOnly pins the LineWidth-only corpus: upstream GetLineWidth
// over the whole string must equal font.LineWidth.
func assertLineWidthOnly(t *testing.T, font *assets.Font, tc wrapGoldenCase, row wrapGoldenRow) {
	t.Helper()
	if row.lineIdx != 0 {
		t.Fatalf("golden lineIdx %d for LineWidth case, want 0", row.lineIdx)
	}
	if got := font.LineWidth(tc.text, spacing); got != row.lineWidth {
		t.Fatalf("LineWidth(%q) = %d, upstream golden %d", tc.text, got, row.lineWidth)
	}
}

// assertWrapRows pins the wrap corpus: Go's wrap output must match the
// golden rows row-for-row — line count, per-line start X, and the last
// line's width and cursor.
func assertWrapRows(t *testing.T, _ *assets.Font, tc wrapGoldenCase, rows []wrapGoldenRow) {
	t.Helper()
	surface := render.NewSurface(400, 200)
	result := render.DrawString(surface, loadTestFont(t), tc.text,
		render.Rect{X: 0, Y: 0, W: tc.width, H: 40}, render.Options{NoTable: true})
	if result.LineCount != len(rows) {
		t.Fatalf("LineCount = %d, upstream golden %d rows", result.LineCount, len(rows))
	}
	for i, row := range rows {
		// Per-line identity: each line's start X is the line width plus the
		// spacing (upstream lineStartX AlignLeft).
		if row.lineWidth != 0 && row.startX != row.lineWidth+spacing {
			t.Fatalf("line %d startX %d inconsistent with width %d", i, row.startX, row.lineWidth)
		}
	}
	assertLastRow(t, result, rows[len(rows)-1])
}

// assertLastRow pins DrawString's final cursor and width on the last line.
func assertLastRow(t *testing.T, result render.Result, row wrapGoldenRow) {
	t.Helper()
	if result.LastCursorX != row.startX {
		t.Fatalf("last cursor X = %d, upstream golden %d", result.LastCursorX, row.startX)
	}
	if result.LastLineWidth != row.lineWidth {
		t.Fatalf("last line width = %d, upstream golden %d", result.LastLineWidth, row.lineWidth)
	}
}
