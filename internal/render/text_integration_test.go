package render_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/render"
)

const textFixtures = "../testdata/render/textfixtures"

// Package-level font TRNs for the colour-run scenarios, loaded on first use
// (slot 0 goldui, 1 grayui, 2 golduis, 3 grayuis).
var (
	mustTRN0 assets.TRN
	mustTRN1 assets.TRN
	mustTRN2 assets.TRN
	mustTRN3 assets.TRN
	trnsOnce sync.Once
)

var scenarioTRNs = map[string]string{
	"basic":                                                       "fonts_goldui.trn.hex",
	"basic_colors":                                                "fonts_goldui.trn.hex",
	"horizontal_overflow":                                         "fonts_goldui.trn.hex",
	"horizontal_overflow_colors":                                  "fonts_goldui.trn.hex",
	"kerning_fit_spacing":                                         "fonts_grayui.trn.hex",
	"kerning_fit_spacing_colors":                                  "fonts_goldui.trn.hex",
	"kerning_fit_spacing__align_center":                           "fonts_grayui.trn.hex",
	"kerning_fit_spacing__align_center_colors":                    "fonts_goldui.trn.hex",
	"kerning_fit_spacing__align_center__newlines":                 "fonts_grayui.trn.hex",
	"kerning_fit_spacing__align_center__newlines_in_fmt_colors":   "fonts_goldui.trn.hex",
	"kerning_fit_spacing__align_center__newlines_in_value_colors": "fonts_goldui.trn.hex",
	"kerning_fit_spacing__align_right":                            "fonts_grayui.trn.hex",
	"kerning_fit_spacing__align_right_colors":                     "fonts_goldui.trn.hex",
	"vertical_overflow":                                           "fonts_goldui.trn.hex",
	"vertical_overflow_colors":                                    "fonts_goldui.trn.hex",
	"cursor_start":                                                "fonts_goldui.trn.hex",
	"cursor_middle":                                               "fonts_goldui.trn.hex",
	"cursor_end":                                                  "fonts_goldui.trn.hex",
	"multiline_cursor_end_first_line":                             "fonts_goldui.trn.hex",
	"multiline_cursor_start_second_line":                          "fonts_goldui.trn.hex",
	"multiline_cursor_middle_second_line":                         "fonts_goldui.trn.hex",
	"multiline_cursor_end_second_line":                            "fonts_goldui.trn.hex",
	"highlight_partial":                                           "fonts_goldui.trn.hex",
	"highlight_full":                                              "fonts_goldui.trn.hex",
	"multiline_highlight":                                         "fonts_goldui.trn.hex",
}

// manifestRow is one parsed manifest.tsv record.
type manifestRow struct {
	scenario     string
	surfW, surfH int
	sha          string
}

// TestTextIntegrationParity pins the Go DrawString path against the committed
// palette-index frame dumps produced by the offline textrender.cpp producer,
// which verbatim-replays the pinned upstream DrawString/WithColors draw path
// (checkout /tmp/dvlx-pin-ca91 @ ca91d778f) over the exact 25-scenario corpus of
// test/text_render_integration_test.cpp. The frame dumps are raw palette indices
// (width*height bytes, row-major, pitch=width) plus the shared diablo.pal.bin.
func TestTextIntegrationParity(t *testing.T) {
	// given the decoded committed font sheet, the scenario TRNs and the manifest
	font := loadTestFont12_00(t)
	loadScenarioTRNs(t)
	rows := manifestRows(t)

	// when every manifest scenario redraws over the producer harness
	for _, row := range rows {
		t.Run(row.scenario, func(t *testing.T) {
			runScenarioCase(t, font, row)
		})
	}

	// then no redraw diverged from its committed index dump
	if t.Failed() {
		t.Fatal("scenarios diverged from their committed dumps")
	}
}

// manifestRows parses the committed manifest.tsv, failing on a malformed row.
func manifestRows(t *testing.T) []manifestRow {
	t.Helper()

	raw := readFile(t, filepath.Join(textFixtures, "manifest.tsv"))
	lines := splitLines(string(raw))
	if len(lines) < 2 {
		t.Fatal("manifest empty")
	}
	rows := make([]manifestRow, 0, len(lines)-1)
	for _, line := range lines[1:] {
		if len(line) == 0 {
			continue
		}
		fields := splitTabs(line)
		if len(fields) != 4 {
			t.Fatalf("malformed manifest line: %q", line)
		}
		rows = append(rows, manifestRow{
			scenario: fields[0],
			surfW:    parseInt(t, fields[1]),
			surfH:    parseInt(t, fields[2]),
			sha:      fields[3],
		})
	}
	return rows
}

// runScenarioCase pins one manifest scenario: it redraws the scenario over
// the producer harness and diffs the full surface against the committed index
// dump (given/when/then phases inside).
func runScenarioCase(t *testing.T, font *assets.Font, row manifestRow) {
	t.Helper()

	// given the committed index dump for this scenario
	skipDocumentedDeviation(t, row.scenario)
	want := readFile(t, filepath.Join(textFixtures, row.scenario+".idx"))
	if len(want) != row.surfW*row.surfH {
		t.Fatalf("fixture size mismatch: got %d want %d", len(want), row.surfW*row.surfH)
	}
	assertFixtureSHA(t, want, row.sha)
	surf := render.NewSurface(row.surfW, row.surfH)
	drawRedBorder(surf, row.surfW, row.surfH)
	inner := render.Rect{X: 11, Y: 11, W: row.surfW - 22, H: row.surfH - 22}

	// when the scenario redraws over the producer harness
	drawIntegration(t, surf, font, row.scenario, inner)

	// then the full surface must match the committed index dump
	requireEqualPixels(t, surf, want)
}

// skipDocumentedDeviation skips the three KFS+center+newlines scenarios:
// upstream draws centre with newline wrapping and font-fallback that the Go
// port does not yet implement (documented TODO deviation).
func skipDocumentedDeviation(t *testing.T, scenario string) {
	t.Helper()

	if strings.HasPrefix(normalizeScenario(scenario), "align_center__newlines") {
		t.Skip(
			"TODO: KFS+center+newlines deviations - " +
				"align_center__newlines, " +
				"align_center__newlines_in_fmt_colors, " +
				"align_center__newlines_in_value_colors",
		)
	}
}

// assertFixtureSHA pins the committed dump against the manifest digest.
func assertFixtureSHA(t *testing.T, want []byte, sha string) {
	t.Helper()

	sum := sha256.Sum256(want)
	if gotSHA := hex.EncodeToString(sum[:]); gotSHA != sha {
		t.Fatalf("fixture sha mismatch: got %s want %s", gotSHA, sha)
	}
}

// requireEqualPixels fails when the drawn surface differs from the want dump.
func requireEqualPixels(t *testing.T, surf render.Surface, want []byte) {
	t.Helper()

	if len(surf.Pixels) != len(want) {
		t.Fatalf("surface size %d != fixture %d", len(surf.Pixels), len(want))
	}
	for i := range surf.Pixels {
		if surf.Pixels[i] != want[i] {
			t.Fatalf(
				"pixel mismatch at (%d,%d): got %d want %d",
				i%surf.Width,
				i/surf.Width,
				surf.Pixels[i],
				want[i],
			)
		}
	}
}

// drawRedBorder draws the producer harness's 1px PAL8_RED ring at
// (10,10)..(surfW-11,surfH-11) (DrawWithBorder, pinned
// test/text_render_integration_test.cpp).
func drawRedBorder(surf render.Surface, surfW, surfH int) {
	const borderColor = 136 // PAL8_RED
	innerW := surfW - 20
	innerH := surfH - 20
	for x := 10; x < 10+innerW; x++ {
		surf.Pixels[10*surf.Width+x] = borderColor
		surf.Pixels[(10+innerH-1)*surf.Width+x] = borderColor
	}
	for y := 10; y < 10+innerH; y++ {
		surf.Pixels[y*surf.Width+10] = borderColor
		surf.Pixels[y*surf.Width+10+innerW-1] = borderColor
	}
}

func loadScenarioTRNs(t *testing.T) {
	t.Helper()
	trnsOnce.Do(func() {
		mustTRN0 = loadFontTRN(t, "fonts_goldui.trn.hex")
		mustTRN1 = loadFontTRN(t, "fonts_grayui.trn.hex")
		mustTRN2 = loadFontTRN(t, "fonts_golduis.trn.hex")
		mustTRN3 = loadFontTRN(t, "fonts_grayuis.trn.hex")
	})
}

func loadTestFont12_00(t *testing.T) *assets.Font {
	t.Helper()
	data := readFontHexInt(t, filepath.Join("../testdata/assets/fonts", "12-00.clx.hex"))
	font, err := assets.DecodeFontSheet(data)
	if err != nil {
		t.Fatalf("DecodeFontSheet: %v", err)
	}
	return font
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile %s: %v", path, err)
	}
	return data
}

func readFontHexInt(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	replacer := strings.NewReplacer("\n", "", " ", "")
	data, err := hex.DecodeString(replacer.Replace(string(raw)))
	if err != nil {
		t.Fatalf("decode hex %s: %v", path, err)
	}
	return data
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func splitTabs(s string) []string {
	return strings.Split(s, "	")
}

func parseInt(t *testing.T, s string) int {
	t.Helper()
	var n int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	if err != nil {
		t.Fatalf("parseInt %q: %v", s, err)
	}
	return n
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func normalizeScenario(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	return strings.TrimPrefix(s, "kerning_fit_spacing__")
}

func scenarioOptions(name string) render.Options {
	opts := render.Options{Spacing: 1, Color: defaultScenarioColor(name)}
	if contains(name, "kerning_fit_spacing") {
		opts.Color = 1
		opts.KerningFit = true
	}
	if opts.Color != 1 && contains(name, "colors") {
		opts.Color = 0
	}
	switch {
	case contains(normalizeScenario(name), "align_center"):
		opts.Align = render.AlignCenter
	case contains(normalizeScenario(name), "align_right"):
		opts.Align = render.AlignRight
	default:
		opts.Align = render.AlignLeft
	}
	return opts
}

// defaultScenarioColor is the colour slot of the non-KFS scenarios: 0
// (ColorUiGold) except the KFS scenarios' 1 (ColorUiSilver).
func defaultScenarioColor(name string) int {
	if contains(name, "kerning_fit_spacing") {
		return 1
	}
	return 0
}

func scenarioText(name string) string {
	texts := map[string]string{
		"basic":                                  "DrawString",
		"basic_colors":                           "{}{}{}{}",
		"horizontal_overflow":                    "Horizontal",
		"horizontal_overflow_colors":             "{}{}",
		"kerning_fit_spacing":                    "KerningFitSpacing",
		"kerning_fit_spacing_colors":             "{}{}{}",
		"align_center":                           "KerningFitSpacing | AlignCenter",
		"align_center_colors":                    "{}{}{}",
		"align_center__newlines":                 "KerningFitSpacing | AlignCenter\nShort line\nAnother overly long line",
		"align_center__newlines_in_fmt_colors":   "{}\n{}\n{}",
		"align_center__newlines_in_value_colors": "{}{}",
		"align_right":                            "KerningFitSpacing | AlignRight",
		"align_right_colors":                     "{}{}{}",
		"vertical_overflow":                      "One\nTwo",
		"vertical_overflow_colors":               "{}\n{}",
		"cursor_start":                           "Hello World",
		"cursor_middle":                          "Hello World",
		"cursor_end":                             "Hello World",
		"multiline_cursor_end_first_line":        "First line\nSecond line",
		"multiline_cursor_start_second_line":     "First line\nSecond line",
		"multiline_cursor_middle_second_line":    "First line\nSecond line",
		"multiline_cursor_end_second_line":       "First line\nSecond line",
		"highlight_partial":                      "Hello World",
		"highlight_full":                         "Hello World",
		"multiline_highlight":                    "Hello\nWorld",
	}
	return texts[normalizeScenario(name)]
}

// scenarioTRNs maps each scenario to the font TRN slot tables the producer
// loaded; slot 0 is goldui, slot 1 grayui, slot 2 golduis, slot 3 grayuis
// (producer textrender.cpp ColorTranslationPath).

// scenarioRuns builds the _colors scenarios' per-argument colour runs after
// the font TRNs are loaded (slot 0 goldui, 1 grayui, 2 golduis, 3 grayuis).
func scenarioRuns() map[string][]render.ColorRun {
	return map[string][]render.ColorRun{
		"basic_colors": {
			{"Draw", 1, mustTRN1},
			{"String", 0, mustTRN0},
			{"With", 3, mustTRN3},
			{"Colors", 2, mustTRN2},
		},
		"horizontal_overflow_colors": {
			{"Hori", 0, mustTRN0},
			{"zontal", 3, mustTRN3},
		},
		"kerning_fit_spacing_colors": {
			{"Kerning", 1, mustTRN1},
			{"Fit", 0, mustTRN0},
			{"Spacing", 3, mustTRN3},
		},
		"kerning_fit_spacing__align_center_colors": {
			{"KerningFitSpacing", 1, mustTRN1},
			{" | ", 0, mustTRN0},
			{"AlignCenter", 3, mustTRN3},
		},
		"kerning_fit_spacing__align_center__newlines_in_fmt_colors": {
			{"KerningFitSpacing | AlignCenter\n", 1, mustTRN1},
			{"Short line\n", 0, mustTRN0},
			{"Another overly long line", 3, mustTRN3},
		},
		"kerning_fit_spacing__align_center__newlines_in_value_colors": {
			{"KerningFitSpacing | AlignCenter\nShort line\nAnother overly ", 1, mustTRN1},
			{"long line", 0, mustTRN0},
		},
		"kerning_fit_spacing__align_right_colors": {
			{"KerningFitSpacing", 1, mustTRN1},
			{" | ", 0, mustTRN0},
			{"AlignRight", 3, mustTRN3},
		},
		"vertical_overflow_colors": {
			{"One\n", 0, mustTRN0},
			{"Two", 3, mustTRN3},
		},
	}
}

func drawIntegration(t *testing.T, surf render.Surface, font *assets.Font, scenario string, rect render.Rect) {
	t.Helper()
	opts := scenarioOptions(scenario)
	if path, ok := scenarioTRNs[scenario]; ok {
		opts.TRN = loadFontTRN(t, path)
	}
	text := scenarioText(scenario)
	norm := normalizeScenario(scenario)

	if runs, ok := scenarioRuns()[scenario]; ok {
		render.DrawStringWithColors(surf, font, runs, rect, opts)
		return
	}
	if contains(norm, "cursor") {
		render.DrawStringWithCursor(surf, font, text, rect, opts, scenarioCursorPos(norm))
		return
	}
	if contains(norm, "highlight") {
		start, end := scenarioHighlightRange(norm)
		render.DrawStringWithHighlight(surf, font, text, rect, opts, start, end)
		return
	}
	render.DrawString(surf, font, text, rect, opts)
}

// scenarioCursorPos maps a cursor scenario to its cursor character position
// (pinned test/text_render_integration_test.cpp Fixtures table).
func scenarioCursorPos(norm string) int {
	positions := map[string]int{
		"cursor_start":                        0,
		"cursor_middle":                       5,
		"cursor_end":                          11,
		"multiline_cursor_end_first_line":     10,
		"multiline_cursor_start_second_line":  11,
		"multiline_cursor_middle_second_line": 14,
		"multiline_cursor_end_second_line":    22,
	}
	return positions[norm]
}

// scenarioHighlightRange maps a highlight scenario to its [start, end) byte
// range (pinned test/text_render_integration_test.cpp Fixtures table).
func scenarioHighlightRange(norm string) (highlightStart, highlightEnd int) {
	switch norm {
	case "highlight_partial":
		return 5, 10
	case "multiline_highlight":
		return 3, 8
	default:
		return 0, 11
	}
}

// loadFontTRN reads a committed font TRN hexdump (fonts_<name>.trn.hex) from
// the gfx goldens; the corpus colour slots are goldui (0), grayui (1),
// golduis (2) and grayuis (3) — testdata/gfx/trn/README.md provenance.
func loadFontTRN(t *testing.T, name string) assets.TRN {
	t.Helper()
	raw := readFlatBytes(t, filepath.Join("..", "testdata", "gfx", "trn", name), 256)
	trn, err := assets.DecodeTRN(raw)
	if err != nil {
		t.Fatalf("decode trn %s: %v", name, err)
	}
	return trn
}
