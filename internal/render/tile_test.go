package render_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render"
)

// The goldens are the committed dump of the offline C++ dumper
// (internal/testdata/render/producer/palrender.cpp, upstream ca91d778f):
// every tile type's fabricated frame rendered through the vendored
// RenderTileFrame for each valid mask and light flavour, plus the black tile.
// Section format: `---- <tile> <mask> <light>` then the dump region rows
// (x [256, 383], y [144, 256]) as "v:count" runs, background 0xAA.

const (
	goldenRegionWidth  = 128
	goldenRegionHeight = 113
	goldenBackground   = 0xAA
	goldenRegionX0     = 256 // DumpX0 = TileX - 64
	goldenRegionY0     = 144 // DumpY0 = TileY - 96
)

// goldenCase is one (tile, frame, mask, fabricated light levels) parity row.
type goldenCase struct {
	tile   render.TileType
	src    []byte
	height int
	mask   render.MaskType
	levels []uint8
}

// triangleFrames carries the fabricated triangle and trapezoid payloads.
type triangleFrames struct {
	Left      []byte
	Right     []byte
	LeftTrap  []byte
	RightTrap []byte
}

// goldenAccumulator builds one dump's pixel region until the next header
// (dun_render_benchmark.cpp section format of the committed TSV).
type goldenAccumulator struct {
	name string
	body []uint8
}

// TestRenderTileGolden renders every dumped (tile, mask, light) combination
// and byte-compares the result with the upstream dumper's region.
func TestRenderTileGolden(t *testing.T) {
	// given: the fabrications and one case row per committed dump section
	base := render.TileLighting{Tables: lightTablesFixture(), Blend: greyRampBlendLUT()}

	for _, c := range goldenCaseSet() {
		for _, level := range c.levels {
			// when
			surface := render.NewSurface(640, 480)
			fillBackground(surface)
			name := casename(c, level)
			lighting := base
			lighting.Level = level
			render.RenderTileFrame(surface, render.PixelPos{X: 320, Y: 240}, c.tile, c.src, c.height, c.mask, lighting)

			// then
			compareGolden(t, name, regionPixels(surface))
		}
	}
}

// TestWorldDrawBlackTileGolden pins the black diamond dump.
func TestWorldDrawBlackTileGolden(t *testing.T) {
	// given
	surface := render.NewSurface(640, 480)
	fillBackground(surface)

	// when
	render.WorldDrawBlackTile(surface, 320, 240)

	// then
	compareGolden(t, "BlackTile", regionPixels(surface))
}

// TestRenderTileGoldenCoversEverySection pins goldenCaseSet + the three light
// flavours against the committed TSV: every dumped section (the 48 tile/mask
// rows plus BlackTile) is asserted by the suite, so an unconsumed section —
// hand-editable truth — fails here before it can drift.
func TestRenderTileGoldenCoversEverySection(t *testing.T) {
	// given: the names the golden run consumes and the committed section names
	asserted := make(map[string]bool)
	for _, c := range goldenCaseSet() {
		for _, level := range c.levels {
			asserted[casename(c, level)] = true
		}
	}
	asserted["BlackTile"] = true

	// when: the committed TSV's section headers are read
	sections := goldenSectionNames(t)

	// then: the sets match exactly
	if len(asserted) != len(sections) {
		t.Fatalf("goldenCaseSet asserts %d sections, TSV dumps %d", len(asserted), len(sections))
	}
	for name := range asserted {
		if !sections[name] {
			t.Fatalf("goldenCaseSet asserts %q, which the committed TSV does not dump", name)
		}
	}
}

// TestWorldDrawBlackTileClippedSkips pins the per-half clip guards of
// world_draw_black_tile (dun_render.cpp:1102, :1109): a half that is not
// fully on the surface draws nothing instead of writing off-column or
// off-row pixels, while a fully visible half still renders.
func TestWorldDrawBlackTileClippedSkips(t *testing.T) {
	t.Run("partially off left half is skipped, right half draws", func(t *testing.T) {
		// given
		surface := render.NewSurface(640, 480)
		fillBackground(surface)

		// when: the left triangle's widest rows start left of column 0
		sx, sy := -10, 240
		render.WorldDrawBlackTile(surface, sx, sy)

		// then: the right half's footprint is the only painted area — the
		// skipped left half would otherwise wrap its negative columns into
		// row 224's tail
		rightX0 := sx + render.DunFrameWidth
		expectBackgroundOutside(t, surface, rightX0, sy-render.TriangleHeight+1, rightX0+render.DunFrameWidth, sy+1)
		for _, pt := range []struct{ x, y int }{{rightX0, sy}, {rightX0 + 6, sy - 15}} {
			if got := surface.Pixels[pt.y*surface.Width+pt.x]; got != 0 {
				t.Fatalf("right-half pixel (%d,%d) = %d, want black", pt.x, pt.y, got)
			}
		}
	})

	t.Run("fully off left half is skipped", func(t *testing.T) {
		// given
		surface := render.NewSurface(640, 480)
		fillBackground(surface)

		// when: the left triangle sits fully left of column 0 and the right
		// one starts before it too
		render.WorldDrawBlackTile(surface, -70, 240)

		// then: no pixel changed anywhere
		expectBackgroundOutside(t, surface, 0, 0, 0, 0)
	})

	t.Run("vertically clipped tile draws nothing without panicking", func(t *testing.T) {
		// given
		surface := render.NewSurface(640, 480)
		fillBackground(surface)

		// when: the diamond's upper rows would index above row 0
		render.WorldDrawBlackTile(surface, 320, 15)

		// then: the call returns and no pixel changed
		expectBackgroundOutside(t, surface, 0, 0, 0, 0)
	})
}

// squareFrame mirrors the C++ producer's BuildSquare: unique value per byte.
func squareFrame() []byte {
	square := make([]byte, 1024)

	for i := range square {
		square[i] = byte((i + 1) & 0xFF)
	}

	return square
}

// casename formats the dump section's header words for a case.
func casename(c goldenCase, level uint8) string {
	return tileName(c.tile) + " " + maskName(c.mask) + " " + lightName(level)
}

// tileName maps a tile to the dump section's tile word.
func tileName(tile render.TileType) string {
	switch tile {
	case render.TileSquare:
		return "Square"
	case render.TileTransparentSquare:
		return "TransparentSquare"
	case render.TileLeftTriangle:
		return "LeftTriangle"
	case render.TileRightTriangle:
		return "RightTriangle"
	case render.TileLeftTrapezoid:
		return "LeftTrapezoid"
	case render.TileRightTrapezoid:
		return "RightTrapezoid"
	default:
		return "?"
	}
}

// maskName maps a mask to the dump section's mask word.
func maskName(mask render.MaskType) string {
	switch mask {
	case render.MaskSolid:
		return "Solid"
	case render.MaskTransparent:
		return "Transparent"
	case render.MaskRight:
		return "Right"
	case render.MaskLeft:
		return "Left"
	default:
		return "?"
	}
}

// lightName maps a fabricated level to the C++ producer's name.
func lightName(level uint8) string {
	switch level {
	case 0:
		return "FullyLit"
	case 15:
		return "FullyDark"
	default:
		return "PartiallyLit"
	}
}

// goldenCaseSet builds the parity rows' fabrications.
func goldenCaseSet() []goldenCase {
	frames := triangleFrameFixtures()
	transparent := transparentSquareFixture()
	all := []uint8{0, 5, 15} // FullyLit, PartiallyLit, FullyDark

	cases := []goldenCase{
		{tile: render.TileSquare, src: squareFrame(), height: 32, mask: render.MaskSolid, levels: all},
		{tile: render.TileSquare, src: squareFrame(), height: 32, mask: render.MaskTransparent, levels: all},
		{tile: render.TileTransparentSquare, src: transparent, height: 32, mask: render.MaskSolid, levels: all},
		{tile: render.TileTransparentSquare, src: transparent, height: 32, mask: render.MaskTransparent, levels: all},
		{tile: render.TileTransparentSquare, src: transparent, height: 32, mask: render.MaskLeft, levels: all},
		{tile: render.TileTransparentSquare, src: transparent, height: 32, mask: render.MaskRight, levels: all},
		{tile: render.TileLeftTriangle, src: frames.Left, height: 31, mask: render.MaskSolid, levels: all},
		{tile: render.TileLeftTriangle, src: frames.Left, height: 31, mask: render.MaskTransparent, levels: all},
		{tile: render.TileRightTriangle, src: frames.Right, height: 31, mask: render.MaskSolid, levels: all},
		{tile: render.TileRightTriangle, src: frames.Right, height: 31, mask: render.MaskTransparent, levels: all},
		{tile: render.TileLeftTrapezoid, src: frames.LeftTrap, height: 32, mask: render.MaskSolid, levels: all},
		{tile: render.TileLeftTrapezoid, src: frames.LeftTrap, height: 32, mask: render.MaskTransparent, levels: all},
		{tile: render.TileLeftTrapezoid, src: frames.LeftTrap, height: 32, mask: render.MaskLeft, levels: all},
		{tile: render.TileRightTrapezoid, src: frames.RightTrap, height: 32, mask: render.MaskSolid, levels: all},
		{
			tile:   render.TileRightTrapezoid,
			src:    frames.RightTrap,
			height: 32,
			mask:   render.MaskTransparent,
			levels: all,
		},
		{tile: render.TileRightTrapezoid, src: frames.RightTrap, height: 32, mask: render.MaskRight, levels: all},
	}

	return cases
}

// decodeGoldenRow expands one "v:count" run row.
func decodeGoldenRow(row string) []uint8 {
	pixels := make([]uint8, 0, goldenRegionWidth)

	for _, field := range strings.Fields(row) {
		pixels = appendEncodedRun(pixels, field)
	}

	return pixels
}

// appendEncodedRun appends one "v" or "v:count" field's pixels.
func appendEncodedRun(pixels []uint8, field string) []uint8 {
	index, count, err := splitRun(field)
	if err != nil {
		return pixels
	}

	for range count {
		pixels = append(pixels, uint8(index))
	}

	return pixels
}

// splitRun reads one "v" or "v:count" field.
func splitRun(field string) (index, count int, err error) {
	at := strings.IndexByte(field, ':')
	if at < 0 {
		var parseErr error
		index, parseErr = strconv.Atoi(field)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("golden run value %q: %w", field, parseErr)
		}

		return index, 1, nil
	}

	var valueErr error
	index, valueErr = strconv.Atoi(field[:at])
	if valueErr != nil {
		return 0, 0, fmt.Errorf("golden run value %q: %w", field, valueErr)
	}

	var countErr error
	count, countErr = strconv.Atoi(field[at+1:])
	if countErr != nil {
		return 0, 0, fmt.Errorf("golden run count %q: %w", field, countErr)
	}

	return index, count, nil
}

// goldenSections parses the committed TSV into one decoded region per section.
func goldenSections(t *testing.T) map[string][]uint8 {
	t.Helper()

	body := readGoldensTSV(t)

	sections := make(map[string][]uint8)
	acc := goldenAccumulator{}

	for _, line := range strings.Split(body, "\n") {
		acc = acc.onLine(sections, line)
	}
	acc.store(sections)

	return sections
}

// goldenSectionNames reads the committed TSV's section headers, the dumped
// set the golden run must cover exactly.
func goldenSectionNames(t *testing.T) map[string]bool {
	t.Helper()

	names := make(map[string]bool)
	for _, line := range strings.Split(readGoldensTSV(t), "\n") {
		if strings.HasPrefix(line, "---- ") {
			names[strings.TrimSpace(strings.TrimPrefix(line, "---- "))] = true
		}
	}

	return names
}

// readGoldensTSV loads the committed golden dump or fails.
func readGoldensTSV(t *testing.T) string {
	t.Helper()

	body, err := os.ReadFile(filepath.Join(fixtureDir, "tiles", "render", "rtf_goldens.tsv"))
	if err != nil {
		t.Fatalf("read goldens: %v", err)
	}

	return string(body)
}

// compareGolden asserts one rendered case against its dump section.
func compareGolden(t *testing.T, name string, got []uint8) {
	t.Helper()

	want, ok := goldenSections(t)[name]
	if !ok {
		t.Fatalf("golden %s missing from the committed TSV", name)
	}
	if mismatches := countMismatches(want, got); mismatches > 0 {
		for _, where := range firstMismatches(want, got, 4) {
			t.Errorf("%s %s", name, where)
		}
		t.Errorf("%s: %d mismatched pixels", name, mismatches)
	}
}

// countMismatches counts differing pixels.
func countMismatches(want, got []uint8) int {
	n := 0

	for i := range want {
		if want[i] != got[i] {
			n++
		}
	}

	return n
}

// firstMismatches prints the first few differing pixels' coordinates.
func firstMismatches(want, got []uint8, limit int) []string {
	out := []string{}

	for i := 0; i < len(want) && len(out) < limit; i++ {
		if want[i] != got[i] {
			out = append(out, fmt.Sprintf("at (%d,%d) = %d, want %d", i%goldenRegionWidth, i/goldenRegionWidth, got[i], want[i]))
		}
	}

	return out
}

// regionPixels lifts the dump region out of a target surface.
func regionPixels(surface render.Surface) []uint8 {
	pixels := make([]uint8, 0, goldenRegionWidth*goldenRegionHeight)

	for y := goldenRegionY0; y < goldenRegionY0+goldenRegionHeight; y++ {
		start := y*surface.Width + goldenRegionX0
		pixels = append(pixels, surface.Pixels[start:start+goldenRegionWidth]...)
	}

	return pixels
}

// fillBackground paints the golden background index over the whole index
// buffer.
func fillBackground(surface render.Surface) {
	for i := range surface.Pixels {
		surface.Pixels[i] = goldenBackground
	}
}

// expectBackgroundOutside walks every pixel of the surface, skipping the
// rectangle [x0,x1)×[y0,y1) and failing when any visited pixel is not the
// golden background index. Skip reporting stays in one predicate so the
// walk itself reads flat.
func expectBackgroundOutside(t *testing.T, surface render.Surface, x0, y0, x1, y1 int) {
	t.Helper()

	for y := range surface.Height {
		for x := range surface.Width {
			if backgrounded(t, surface, x, y, x0, y0, x1, y1) {
				continue
			}
		}
	}
}

// backgrounded reports the mismatch for one pixel, skipping pixels inside
// the given rectangle; outside the rectangle any non-background index fails
// the test.
func backgrounded(t *testing.T, surface render.Surface, x, y, x0, y0, x1, y1 int) bool {
	t.Helper()

	if x >= x0 && x < x1 && y >= y0 && y < y1 {
		return true
	}
	if got := surface.Pixels[y*surface.Width+x]; got != goldenBackground {
		t.Fatalf("pixel (%d,%d) = %d, want background %d", x, y, got, goldenBackground)
	}

	return false
}

// onLine folds one TSV line into the accumulator; section returns hand the
// finished region to the map.
func (a goldenAccumulator) onLine(sections map[string][]uint8, line string) goldenAccumulator {
	switch {
	case strings.HasPrefix(line, "---- "):
		a.store(sections)
		name := strings.TrimSpace(strings.TrimPrefix(line, "---- "))

		return goldenAccumulator{name: name, body: make([]uint8, 0, goldenRegionWidth*goldenRegionHeight)}
	case line == "" || strings.HasPrefix(line, "#") || a.name == "":
		return a
	default:
		a.body = append(a.body, decodeGoldenRow(line)...)

		return a
	}
}

// store records the accumulator when it holds a complete region.
func (a goldenAccumulator) store(sections map[string][]uint8) {
	if a.name != "" && len(a.body) == goldenRegionWidth*goldenRegionHeight {
		sections[a.name] = a.body
	}
}

// triangleFrameFixtures mirrors the C++ producer's FillTriangle plus the
// trapezoid rect rows.
func triangleFrameFixtures() triangleFrames {
	base := triangleBase()
	rect := trapezoidRect()
	lower := base[:272] // the 272-byte trapezoid base (16 rows, 2+4+..+32)

	return triangleFrames{
		Left:      slices.Concat(base, nil),
		Right:     slices.Concat(base, nil),
		LeftTrap:  slices.Concat(lower, rect),
		RightTrap: slices.Concat(lower, rect),
	}
}

// triangleBase builds the 512-byte re-encoded triangle payload.
func triangleBase() []byte {
	frame := []byte{}
	width := 2

	for row := range 16 {
		frame = appendTriangleRow(frame, row, width)
		width += 2
	}

	for row := range 15 {
		frame = appendTriangleRow(frame, row, 30-2*row)
	}

	return frame
}

// appendTriangleRow appends one triangle row of the given width.
func appendTriangleRow(frame []byte, row, width int) []byte {
	for j := range width {
		frame = append(frame, byte(1+((row*4+j)&0x7F)))
	}

	return frame
}

// trapezoidRect builds the trapezoid's 16 rect rows.
func trapezoidRect() []byte {
	rect := make([]byte, 0, 16*32)

	for r := range 16 {
		for x := range 32 {
			rect = append(rect, byte(1+((r*6+x)&0x7F)))
		}
	}

	return rect
}

// transparentSquareFixture mirrors the C++ producer's BuildTransparentSquare:
// 32 rows bottom-up, alternating solid runs of 3 and transparent runs of 4,
// values 1 + ((r*7 + x) & 0x7F), rows ending exactly at the 32nd column.
func transparentSquareFixture() []byte {
	stream := []byte{}

	for r := range 32 {
		stream = appendVirtualRow(stream, r)
	}

	return stream
}

// appendVirtualRow appends row r's fabricated TransparentSquare runs.
func appendVirtualRow(stream []byte, r int) []byte {
	x := 0

	for x < 32 {
		// the producer's fixed solid = 3; the 3+4 pattern never reaches the
		// tail clamp upstream applies per pixel (palrender.cpp:113-118)
		solid := 3
		stream = append(stream, byte(solid))
		for j := range solid {
			stream = append(stream, byte(1+((r*7+x+j)&0x7F)))
		}
		x += solid
		trans := 32 - x
		if trans > 4 {
			trans = 4
		}
		if trans > 0 {
			stream = append(stream, byte(-int8(trans)))
		}
		x += trans
	}

	return stream
}

// lightTablesFixture mirrors the C++ producer: table k maps index i to
// (i + k*13) & 0xFF.
func lightTablesFixture() *[render.NumShades][render.PaletteSize]uint8 {
	tables := new([render.NumShades][render.PaletteSize]uint8)

	for k := range tables {
		for i := range tables[k] {
			tables[k][i] = byte((i + k*13) & 0xFF)
		}
	}

	return tables
}

// greyRampBlendLUT mirrors the C++ producer's blend: a grey ramp palette with
// all 256 entries keeps the 50% blend at the exact midpoint.
func greyRampBlendLUT() render.BlendLUT {
	var lut render.BlendLUT

	for d := range lut {
		for s := range lut[d] {
			lut[d][s] = uint8((d + s) / 2)
		}
	}

	return lut
}
