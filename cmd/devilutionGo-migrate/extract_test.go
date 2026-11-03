package main

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// renderCase is one record and the manifest metadata it must produce.
type renderCase struct {
	name       string
	payload    []byte
	family     string
	format     string
	outputName string
	rows       int
}

func TestRenderClassifiesEachRecord(t *testing.T) {
	// given one record per format the tool knows
	cases := []renderCase{
		{"txtdata\\items.tsv", tsvFixture(), familyTable, "csv", "txtdata/items.csv", 2},
		{"data\\health.clx", clxFixture(t), familySprite, "clx", "data/health.clx", 0},
		{"fonts\\12-05.clx", fontSheetFixture(t), familyFont, "clx", "fonts/12-05.clx", 0},
		{"ui_art\\title.pcx", pcxFixture(t), familySprite, "pcx", "ui_art/title.pcx", 0},
		{"ui_art\\diablo.pal", palFixture(t), familyPalette, "pal", "ui_art/diablo.pal", 0},
		{"gendata\\pause.trn", trnFixture(t), familyTRN, "trn", "gendata/pause.trn", 0},
		{"levels\\l1.cel", []byte("cel!"), familySprite, "cel", "levels/l1.cel", 0},
		{"plrgfx\\warrior\\whu.cel", []byte("cel!"), familySprite, "cel", "plrgfx/warrior/whu.cel", 0},
		{"sfx\\items\\titlemov.wav", []byte("RIFF"), familyRaw, "wav", "sfx/items/titlemov.wav", 0},
	}

	// when every record is rendered
	rendered := renderAll(t, cases)

	// then each is classified by its format
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) { assertRendered(t, rendered[testCase.name], testCase) })
	}
}

func TestRenderWritesTablesAsCSV(t *testing.T) {
	// when a table with a quoted value is rendered
	got, err := render("txtdata\\items.tsv", []byte("id\tname\n7\t\"quoted\"\n"))
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	// then the CSV file is the output and carries the source values
	if got.outputName != "txtdata/items.csv" {
		t.Fatalf("output name = %q", got.outputName)
	}

	if want := "id,name\n7,\"\"\"quoted\"\"\"\n"; string(got.payload) != want {
		t.Errorf("csv = %q, want %q", got.payload, want)
	}
}

func TestRenderKeepsUndecodableFormatsByteIdentical(t *testing.T) {
	// given CEL/CL2 records, whose frame dimensions live in caller tables
	// upstream (cel_to_clx.cpp:84-86, cl2_to_clx.cpp:66), not in the container
	payload := []byte{0x00, 0x01, 0x02, 0x03}
	names := []string{"levels\\l1.cel", "plrgfx\\whu.cl2"}

	// when each is rendered
	outputs := renderAll(t, undecodableCases(names, payload))

	// then the bytes pass through and no frame layout is claimed
	for _, name := range names {
		t.Run(name, func(t *testing.T) { assertUndecoded(t, outputs[name], payload) })
	}
}

func TestRenderRejectsCorruptPayload(t *testing.T) {
	// given records that claim a format but carry broken bytes
	cases := map[string]struct {
		name    string
		payload []byte
		want    string
	}{
		"clx":     {"data\\health.clx", []byte{1, 2, 3}, "decode clx"},
		"pcx":     {"ui_art\\title.pcx", []byte("PCX"), "decode pcx"},
		"palette": {"ui_art\\diablo.pal", []byte{1, 2, 3}, "decode palette"},
		"trn":     {"gendata\\pause.trn", []byte{1, 2, 3}, "decode trn"},
		"font":    {"fonts\\12-05.clx", []byte{1, 2, 3}, "decode font sheet"},
	}

	// when each record is rendered
	_, failures := renderFailures(cases)

	// then the decoder's refusal is the failure
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) { assertRejected(t, name, failures[name], testCase.want) })
	}
}

func TestIsBookkeepingRecord(t *testing.T) {
	// given archive metadata records next to real assets
	cases := map[string]bool{
		`(listfile)`:       true,
		`(attributes)`:     true,
		`(signature)`:      true,
		`ui_art\title.pcx`: false,
		`data\(paren).clx`: false,
	}

	// when every record name is classified
	classified := classify(cases, isBookkeepingRecord)

	// then only the parenthesised metadata records are skipped
	for name, want := range cases {
		if classified[name] != want {
			t.Errorf("isBookkeepingRecord(%q) = %v, want %v", name, classified[name], want)
		}
	}
}

func TestIsFontSheet(t *testing.T) {
	// given CLX records from the fonts directory and from anywhere else
	cases := map[string]bool{
		"fonts/12-05.clx":   true,
		"Fonts/12-05.clx":   true,
		"data/health.clx":   false,
		"gendata/fonts.clx": false,
	}

	// when every record is classified
	classified := classify(cases, isFontSheet)

	// then only the fonts directory counts as a font sheet
	for name, want := range cases {
		if classified[name] != want {
			t.Errorf("isFontSheet(%q) = %v, want %v", name, classified[name], want)
		}
	}
}

func TestRecordOutputPathRejectsEscapingNames(t *testing.T) {
	// given record names an archive could carry that climb out of its tree
	cases := []string{
		`..\..\evil.txt`,
		`..\evil.txt`,
		`a\..\..\evil.txt`,
		`/etc/passwd`,
		`\\evil.txt`,
		`..`,
		`.`,
		``,
	}

	// when each name is placed inside the archive tree
	failures := placeAll(t, cases)

	// then every one of them is refused instead of written outside --out
	for _, name := range cases {
		t.Run(name, func(t *testing.T) { assertEscapingNameRefused(t, failures[name], name) })
	}
}

func TestRecordOutputPathKeepsNamesInsideTheTree(t *testing.T) {
	// given names that stay inside the archive's tree, nested or not
	cases := map[string]string{
		`ui_art\title.pcx`: "spawn/ui_art/title.pcx",
		`a\b\c.txt`:        "spawn/a/b/c.txt",
		`a\\b.txt`:         "spawn/a/b.txt",
		`deep\a\..\b.txt`:  "spawn/deep/b.txt",
	}

	// when each name is placed inside the archive tree
	placed := placeInside(t, cases)

	// then each lands at the cleaned path under the archive stem
	for name, want := range cases {
		t.Run(name, func(t *testing.T) { assertConfinedNameKept(t, placed[name], want) })
	}
}

func TestClxLayoutReportsUniformFrameSize(t *testing.T) {
	// given a list whose frames all share one size
	list := &assets.ClxSpriteList{
		Frames:     []assets.ClxFrame{{Width: 96, Height: 96}, {Width: 96, Height: 96}},
		GroupSizes: []int{2},
	}

	// when the layout is summarized
	layout := clxLayout(list)

	// then the count and the shared size are recorded
	if layout.Count != 2 || layout.Width != 96 || layout.Height != 96 {
		t.Errorf("layout = %+v, want 2 frames of 96x96", *layout)
	}

	if len(layout.Groups) != 1 || layout.Groups[0] != 2 {
		t.Errorf("groups = %v, want [2]", layout.Groups)
	}
}

func TestClxLayoutDropsMixedFrameSizes(t *testing.T) {
	// given a list whose frames differ in size
	list := &assets.ClxSpriteList{Frames: []assets.ClxFrame{
		{Width: 96, Height: 96},
		{Width: 64, Height: 32},
	}}

	// when the layout is summarized
	layout := clxLayout(list)

	// then only the count survives: a single size would be a lie
	if layout.Count != 2 || layout.Width != 0 || layout.Height != 0 {
		t.Errorf("layout = %+v, want a count without a size", *layout)
	}
}

// assertRendered checks one rendered record against its expectation.
func assertRendered(t *testing.T, got rendered, want renderCase) {
	t.Helper()

	if got.family != want.family || got.format != want.format {
		t.Errorf("family/format = %s/%s, want %s/%s", got.family, got.format, want.family, want.format)
	}

	if got.outputName != want.outputName {
		t.Errorf("output name = %q, want %q", got.outputName, want.outputName)
	}

	if got.rows != want.rows {
		t.Errorf("rows = %d, want %d", got.rows, want.rows)
	}
}

// assertUndecoded checks that a record passed through untouched.
func assertUndecoded(t *testing.T, got rendered, payload []byte) {
	t.Helper()

	if string(got.payload) != string(payload) {
		t.Errorf("payload = %v, want the record bytes", got.payload)
	}

	if got.frames != nil {
		t.Errorf("frames = %+v, want none", *got.frames)
	}
}

// assertRejected checks that a corrupt record failed with the named cause.
func assertRejected(t *testing.T, name string, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("render accepted a corrupt record %s", name)
	}

	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to name %q", err, want)
	}
}

// renderAll renders every case and fails the test on a decode error.
func renderAll(t *testing.T, cases []renderCase) map[string]rendered {
	t.Helper()

	outputs := make(map[string]rendered, len(cases))

	for _, testCase := range cases {
		got, err := render(testCase.name, testCase.payload)
		if err != nil {
			t.Fatalf("render %s: %v", testCase.name, err)
		}

		outputs[testCase.name] = got
	}

	return outputs
}

// undecodableCases builds one case per record name.
func undecodableCases(names []string, payload []byte) []renderCase {
	cases := make([]renderCase, 0, len(names))

	for _, name := range names {
		cases = append(cases, renderCase{name: name, payload: payload})
	}

	return cases
}

// renderFailures renders every case and returns the errors by case name.
func renderFailures(cases map[string]struct {
	name    string
	payload []byte
	want    string
},
) (map[string]rendered, map[string]error) {
	rendered := make(map[string]rendered, len(cases))
	failures := make(map[string]error, len(cases))

	for key, testCase := range cases {
		got, err := render(testCase.name, testCase.payload)
		rendered[key] = got
		failures[key] = err
	}

	return rendered, failures
}

// classify applies one name classifier to every case.
func classify(cases map[string]bool, classifier func(string) bool) map[string]bool {
	classified := make(map[string]bool, len(cases))

	for name := range cases {
		classified[name] = classifier(name)
	}

	return classified
}

// placeAll places every record name and returns the resulting errors by name.
func placeAll(t *testing.T, names []string) map[string]error {
	t.Helper()

	failures := make(map[string]error, len(names))

	for _, name := range names {
		_, err := recordOutputPath("spawn", name)
		failures[name] = err
	}

	return failures
}

// placeInside places every record name and returns the resulting path by name.
func placeInside(t *testing.T, cases map[string]string) map[string]string {
	t.Helper()

	placed := make(map[string]string, len(cases))

	for name := range cases {
		got, err := recordOutputPath("spawn", name)
		if err != nil {
			t.Fatalf("recordOutputPath(%q): %v", name, err)
		}

		placed[name] = got
	}

	return placed
}

// assertEscapingNameRefused checks that a record name climbing out of the tree
// is refused.
func assertEscapingNameRefused(t *testing.T, err error, name string) {
	t.Helper()

	if err == nil {
		t.Fatalf("recordOutputPath accepted %q", name)
	}

	if !strings.Contains(err.Error(), "escapes the archive tree") {
		t.Errorf("error = %q, want the escape reason", err)
	}
}

// assertConfinedNameKept checks that a name inside the tree lands under the stem.
func assertConfinedNameKept(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("path = %q, want %q", got, want)
	}
}
