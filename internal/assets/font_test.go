package assets_test

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

const (
	fontFixtures = "../testdata/assets/fonts"
	fontSheet    = "12-05"
	// fontHexPerLine is the producer's hex dump width
	// (internal/testdata/assets/fonts/README.md).
	fontHexPerLine = 16
)

// fontGolden is the parsed <sheet>.sha256 row.
type fontGolden struct {
	digest     string
	pixels     int
	frames     int
	zeroWidth  int
	rowHeights []int
}

func TestDecodeFontSheetMatchesGolden(t *testing.T) {
	// given the committed 12-05 CLX sheet bytes and its decode digest
	golden := readFontGolden(t, fontSheet)
	data := readHexFile(t, filepath.Join(fontFixtures, fontSheet+".clx.hex"))

	// when the sheet is decoded
	font := decodeFontSheetBytes(t, data)

	// then the digest, glyph count and blank-glyph count match the golden
	expectFontDigest(t, font, golden)

	// and the wide and blank glyphs match their per-glyph dumps
	expectFontGlyph(t, font, 0xD0, golden)
	expectFontGlyph(t, font, 0xDE, golden)
	expectFontBlankGlyph(t, font, 0x00)
}

func TestFontLineWidth(t *testing.T) {
	// given a decoded sheet and a two-character line
	font := decodeFontSheetBytes(t, readHexFile(t, filepath.Join(fontFixtures, fontSheet+".clx.hex")))
	const spacing = 1

	// when the line width is measured
	width := font.LineWidth("ab", spacing)

	// then it is the sum of both glyph widths plus one inner spacing
	if want := font.Width('a') + font.Width('b') + spacing; width != want {
		t.Fatalf("LineWidth(ab) = %d, want %d", width, want)
	}

	// and an empty line is zero wide
	if got := font.LineWidth("", spacing); got != 0 {
		t.Fatalf("LineWidth(\"\") = %d, want 0", got)
	}
}

// expectFontDigest asserts the decoded sheet against the committed digest row.
func expectFontDigest(t *testing.T, font *assets.Font, golden fontGolden) {
	t.Helper()

	if got := font.Digest(); got != golden.digest {
		t.Fatalf("digest = %s, want %s", got, golden.digest)
	}
	if got := len(font.Glyphs()); got != golden.frames {
		t.Fatalf("glyph count = %d, want %d", got, golden.frames)
	}
	if got := len(font.Glyphs()); got != assets.FontGlyphCount {
		t.Fatalf("glyph count = %d, want %d", got, assets.FontGlyphCount)
	}

	pixels, zeroWidth := fontExtents(font)
	if pixels != golden.pixels {
		t.Fatalf("decoded pixels = %d, want %d", pixels, golden.pixels)
	}
	if zeroWidth != golden.zeroWidth {
		t.Fatalf("zero-width glyphs = %d, want %d", zeroWidth, golden.zeroWidth)
	}
	if heights := fontRowHeights(font); heights != joinInts(golden.rowHeights) {
		t.Fatalf("glyph row heights = %s, want %s", heights, joinInts(golden.rowHeights))
	}
}

// expectFontGlyph asserts one non-blank glyph against its committed index dump.
func expectFontGlyph(t *testing.T, font *assets.Font, code rune, golden fontGolden) {
	t.Helper()

	glyph := font.Glyph(code)
	if glyph.Width == 0 {
		t.Fatalf("glyph 0x%02x is blank on the sheet", code)
	}

	name := filepath.Join(fontFixtures, fontSheet+".glyph-"+hex.EncodeToString([]byte{byte(code)})+".hex")
	want := readHexDump(t, name)
	got := formatHexDump(glyph.Pixels)
	if got != want {
		t.Fatalf("glyph 0x%02x pixels differ:\n got %s\nwant %s", code, got, want)
	}

	if glyph.Height != golden.rowHeights[0] {
		t.Fatalf("glyph 0x%02x height = %d, want %d", code, glyph.Height, golden.rowHeights[0])
	}
}

// expectFontBlankGlyph asserts a zero-width glyph carries no pixel stream.
func expectFontBlankGlyph(t *testing.T, font *assets.Font, code rune) {
	t.Helper()

	glyph := font.Glyph(code)
	if glyph.Width != 0 {
		t.Fatalf("glyph 0x%02x width = %d, want 0", code, glyph.Width)
	}
	if len(glyph.Pixels) != 0 {
		t.Fatalf("blank glyph 0x%02x carries %d pixels", code, len(glyph.Pixels))
	}
	blankDump := strings.TrimSpace(readHexDump(t, filepath.Join(fontFixtures, fontSheet+".glyph-00.hex")))
	if blankDump != "" {
		t.Fatalf("committed blank-glyph dump = %q, want empty", blankDump)
	}
}

// fontExtents returns the decoded pixel count and the zero-width glyph count.
func fontExtents(font *assets.Font) (pixels, blanks int) {
	for _, glyph := range font.Glyphs() {
		pixels += len(glyph.Pixels)
		if glyph.Width == 0 {
			blanks++
		}
	}
	return pixels, blanks
}

// fontRowHeights returns the sorted, comma-joined glyph heights.
func fontRowHeights(font *assets.Font) string {
	seen := make(map[int]bool)
	heights := make([]int, 0, 1)
	for _, glyph := range font.Glyphs() {
		if !seen[glyph.Height] {
			seen[glyph.Height] = true
			heights = append(heights, glyph.Height)
		}
	}
	return joinInts(sortInts(heights))
}

// formatHexDump formats pixels the way the producer's hexdump does.
func formatHexDump(pixels []byte) string {
	var out strings.Builder
	for i := 0; i < len(pixels); i += fontHexPerLine {
		end := min(i+fontHexPerLine, len(pixels))
		out.WriteString(hex.EncodeToString(pixels[i:end]))
		out.WriteString("\n")
	}
	return out.String()
}

// readHexFile decodes a committed hex dump file into bytes.
func readHexFile(t *testing.T, path string) []byte {
	t.Helper()

	text := readHexDump(t, path)
	joined := strings.NewReplacer("\n", "", " ", "").Replace(text)
	data, err := hex.DecodeString(joined)
	if err != nil {
		t.Fatalf("decode hex %s: %v", path, err)
	}
	return data
}

// readHexDump returns the committed hex dump text verbatim, or fails when the
// fixture is missing: a missing fixture is a corpus defect, never a skip.
func readHexDump(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return string(data)
}

// readFontGolden parses the <sheet>.sha256 row.
func readFontGolden(t *testing.T, sheet string) fontGolden {
	t.Helper()

	line := strings.TrimSpace(readHexDump(t, filepath.Join(fontFixtures, sheet+".sha256")))
	fields := strings.Fields(line)
	if len(fields) != 5 {
		t.Fatalf("golden row %q has %d fields, want 5", line, len(fields))
	}

	return fontGolden{
		digest:     fields[0],
		pixels:     mustAtoi(t, fields[1]),
		frames:     mustAtoi(t, fields[2]),
		zeroWidth:  mustAtoi(t, fields[3]),
		rowHeights: splitInts(t, fields[4]),
	}
}

func decodeFontSheetBytes(t *testing.T, data []byte) *assets.Font {
	t.Helper()

	font, err := assets.DecodeFontSheet(data)
	if err != nil {
		t.Fatalf("DecodeFontSheet: %v", err)
	}
	return font
}

func mustAtoi(t *testing.T, text string) int {
	t.Helper()

	value, err := strconv.Atoi(text)
	if err != nil {
		t.Fatalf("parse %q: %v", text, err)
	}
	return value
}

func splitInts(t *testing.T, text string) []int {
	t.Helper()

	parts := strings.Split(text, ",")
	values := make([]int, 0, len(parts))
	for _, part := range parts {
		values = append(values, mustAtoi(t, part))
	}
	return values
}

func sortInts(values []int) []int {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
	return values
}

func joinInts(values []int) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.Itoa(value)
	}
	return strings.Join(parts, ",")
}
