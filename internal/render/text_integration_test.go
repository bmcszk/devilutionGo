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

func loadScenarioTRNs(t *testing.T) {
	t.Helper()
	trnsOnce.Do(func() {
		mustTRN0 = loadFontTRN(t, "fonts_goldui.trn.hex")
		mustTRN1 = loadFontTRN(t, "fonts_grayui.trn.hex")
		mustTRN2 = loadFontTRN(t, "fonts_golduis.trn.hex")
		mustTRN3 = loadFontTRN(t, "fonts_grayuis.trn.hex")
	})
}

// TestTextIntegrationParity pins the Go DrawString path against the committed
// palette-index frame dumps produced by the offline textrender.cpp producer,
// which verbatim-replays the pinned upstream DrawString/WithColors draw path
// (checkout /tmp/dvlx-pin-ca91 @ ca91d778f) over the exact 25-scenario corpus of
// test/text_render_integration_test.cpp. The frame dumps are raw palette indices
// (width*height bytes, row-major, pitch=width) plus the shared diablo.pal.bin.
func TestTextIntegrationParity(t *testing.T) {
	font := loadTestFont12_00(t)

	manifestPath := filepath.Join(textFixtures, "manifest.tsv")
	raw := readFile(t, manifestPath)
	lines := splitLines(string(raw))
	if len(lines) < 2 {
		t.Fatalf("manifest empty")
	}
	for _, line := range lines[1:] {
		if len(line) == 0 {
			continue
		}
		fields := splitTabs(line)
		if len(fields) != 4 {
			t.Fatalf("malformed manifest line: %q", line)
		}
		scenario, widthStr, heightStr, sha := fields[0], fields[1], fields[2], fields[3]
		// Manifest width/height are FULL surface dimensions (inner+20)x(inner+20)
		// with 1px border at (10,10)..(w+9,h+9) where w,h are inner dimensions
		surfW := parseInt(t, widthStr)
		surfH := parseInt(t, heightStr)

		loadScenarioTRNs(t)

		t.Run(scenario, func(t *testing.T) {
			// The three KFS+center+newlines scenarios are documented deviations
			// (upstream draws centre with newline wrapping and font-fallback that the
			// Go port does not yet implement). Skipped with TODO until resolved.
			norm := normalizeScenario(scenario)
			if strings.HasPrefix(norm, "align_center__newlines") {
				t.Skip(
					"TODO: KFS+center+newlines deviations - " +
						"align_center__newlines, " +
						"align_center__newlines_in_fmt_colors, " +
						"align_center__newlines_in_value_colors",
				)
			}
			idxPath := filepath.Join(textFixtures, scenario+".idx")
			wantIdx := readFile(t, idxPath)
			if len(wantIdx) != surfW*surfH {
				t.Fatalf("fixture size mismatch: got %d want %d", len(wantIdx), surfW*surfH)
			}
			sum := sha256.Sum256(wantIdx)
			gotSHA := hex.EncodeToString(sum[:])
			if gotSHA != sha {
				t.Fatalf("fixture sha mismatch: got %s want %s", gotSHA, sha)
			}

			surf := render.NewSurface(surfW, surfH)
			opts := scenarioOptions(scenario)

			// Draw the 1px PAL8_RED border (producer DrawWithBorder)
			// Border rectangle: top-left (10,10), size (innerW, innerH) = (surfW-20, surfH-20)
			// Top/bottom at y=10 and y=10+innerH-1 = surfH-11
			// Left/right at x=10 and x=10+innerW-1 = surfW-11
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

			// Inner rect: inset 1px from border = (11, 11) with size (innerW-2, innerH-2)
			inner := render.Rect{X: 11, Y: 11, W: innerW - 2, H: innerH - 2}

			drawIntegration(surf, font, scenario, inner, opts, t)

			// Compare full surface
			if len(surf.Pixels) != len(wantIdx) {
				t.Fatalf("surface size %d != fixture %d", len(surf.Pixels), len(wantIdx))
			}
			for i := range surf.Pixels {
				if surf.Pixels[i] != wantIdx[i] {
					x := i % surf.Width
					y := i / surf.Width
					t.Fatalf(
						"pixel mismatch at (%d,%d): got %d want %d",
						x,
						y,
						surf.Pixels[i],
						wantIdx[i],
					)
				}
			}
		})
	}
}

func loadTestFont12_00(t *testing.T) *assets.Font {
	t.Helper()
	data := readFontHexInt(t, filepath.Join("../testdata/assets/fonts", "12-05.clx.hex"))
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
	hasOrig := func(s string) bool { return contains(name, s) }
	has := func(s string) bool { return contains(normalizeScenario(name), s) }
	hasKFS := func(s string) bool { return contains(name, s) }

	opts := render.Options{Spacing: 1}
	if hasOrig("basic") || hasOrig("horizontal") || hasOrig("vertical") || hasOrig("cursor") || hasOrig("highlight") ||
		hasOrig("multiline") {
		opts.Color = 0
	}
	if hasKFS("kerning_fit_spacing") {
		opts.Color = 1
		opts.KerningFit = true
	}
	if has("align_center") {
		opts.Align = render.AlignCenter
	}
	if has("align_right") {
		opts.Align = render.AlignRight
	}
	if has("colors") {
		opts.Color = 0
	}
	return opts
}

func scenarioText(name string) string {
	norm := normalizeScenario(name)
	switch norm {
	case "basic":
		return "DrawString"
	case "basic_colors":
		return "{}{}{}{}"
	case "horizontal_overflow":
		return "Horizontal"
	case "horizontal_overflow_colors":
		return "{}{}"
	case "kerning_fit_spacing":
		return "KerningFitSpacing"
	case "kerning_fit_spacing_colors":
		return "{}{}{}"
	case "align_center":
		return "KerningFitSpacing | AlignCenter"
	case "align_center_colors":
		return "{}{}{}"
	case "align_center__newlines":
		return "KerningFitSpacing | AlignCenter\nShort line\nAnother overly long line"
	case "align_center__newlines_in_fmt_colors":
		return "{}\n{}\n{}"
	case "align_center__newlines_in_value_colors":
		return "{}{}"
	case "align_right":
		return "KerningFitSpacing | AlignRight"
	case "align_right_colors":
		return "{}{}{}"
	case "vertical_overflow":
		return "One\nTwo"
	case "vertical_overflow_colors":
		return "{}\n{}"
	case "cursor_start", "cursor_middle", "cursor_end":
		return "Hello World"
	case "multiline_cursor_end_first_line",
		"multiline_cursor_start_second_line",
		"multiline_cursor_middle_second_line",
		"multiline_cursor_end_second_line":
		return "First line\nSecond line"
	case "highlight_partial", "highlight_full":
		return "Hello World"
	case "multiline_highlight":
		return "Hello\nWorld"
	default:
		panic("unknown scenario: " + norm)
	}
}

// scenarioTRNs maps each scenario to the font TRN slot tables the producer
// loaded; slot 0 is goldui, slot 1 grayui, slot 2 golduis, slot 3 grayuis
// (producer textrender.cpp ColorTranslationPath).
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

// scenarioRuns builds the _colors scenarios' per-argument colour runs after
// the font TRNs are loaded (slot 0 goldui, 1 grayui, 2 golduis, 3 grayuis).
func scenarioRuns() map[string][]render.ColorRun {
	return map[string][]render.ColorRun{
		"basic_colors": {{"Draw", 1, mustTRN1},
			{"String", 0, mustTRN0},
			{"With", 3, mustTRN3},
			{"Colors", 2, mustTRN2}},
		"horizontal_overflow_colors": {
			{"Hori", 0, mustTRN0},
			{"zontal", 3, mustTRN3}},
		"kerning_fit_spacing_colors": {
			{"Kerning", 1, mustTRN1},
			{"Fit", 0, mustTRN0},
			{"Spacing", 3, mustTRN3}},
		"kerning_fit_spacing__align_center_colors": {
			{"KerningFitSpacing", 1, mustTRN1},
			{" | ", 0, mustTRN0},
			{"AlignCenter", 3, mustTRN3}},
		"kerning_fit_spacing__align_center__newlines_in_fmt_colors": {
			{"KerningFitSpacing | AlignCenter\n", 1, mustTRN1},
			{"Short line\n", 0, mustTRN0},
			{"Another overly long line", 3, mustTRN3}},
		"kerning_fit_spacing__align_center__newlines_in_value_colors": {
			{"KerningFitSpacing | AlignCenter\nShort line\nAnother overly ", 1, mustTRN1},
			{"long line", 0, mustTRN0}},
		"kerning_fit_spacing__align_right_colors": {
			{"KerningFitSpacing", 1, mustTRN1},
			{" | ", 0, mustTRN0},
			{"AlignRight", 3, mustTRN3}},
		"vertical_overflow_colors": {
			{"One\n", 0, mustTRN0},
			{"Two", 3, mustTRN3}},
	}
}

func drawIntegration(
	surf render.Surface,
	font *assets.Font,
	scenario string,
	rect render.Rect,
	opts render.Options,
	t *testing.T,
) {
	t.Helper()
	text := scenarioText(scenario)
	norm := normalizeScenario(scenario)
	if path, ok := scenarioTRNs[scenario]; ok {
		opts.TRN = loadFontTRN(t, path)
	}
	has := func(s string) bool { return contains(norm, s) }

	if runs, ok := scenarioRuns()[scenario]; ok {
		render.DrawStringWithColors(surf, font, runs, rect, opts)
		return
	}
	if has("cursor") {
		var cursorPos int
		switch norm {
		case "cursor_start":
			cursorPos = 0
		case "cursor_middle":
			cursorPos = 5
		case "cursor_end":
			cursorPos = 11
		case "multiline_cursor_end_first_line":
			cursorPos = 10
		case "multiline_cursor_start_second_line":
			cursorPos = 11
		case "multiline_cursor_middle_second_line":
			cursorPos = 14
		case "multiline_cursor_end_second_line":
			cursorPos = 22
		}
		render.DrawStringWithCursor(surf, font, text, rect, opts, cursorPos)
		return
	}
	if has("highlight") {
		var highlightStart, highlightEnd int
		switch norm {
		case "highlight_partial":
			highlightStart, highlightEnd = 5, 10
		case "highlight_full":
			highlightStart, highlightEnd = 0, 11
		case "multiline_highlight":
			highlightStart, highlightEnd = 3, 8
		}
		render.DrawStringWithHighlight(surf, font, text, rect, opts, highlightStart, highlightEnd)
		return
	}
	render.DrawString(surf, font, text, rect, opts)
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
