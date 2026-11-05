package manifest_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets/manifest"
)

// validateCases is the Validate specification table: each case mutates the
// baseline manifest by exactly the field it is about.
var validateCases = []validateCase{
	{
		name:    "valid manifest passes",
		mutate:  func(*manifest.Manifest) {},
		wantErr: "",
	},
	{
		name:    "empty tool",
		mutate:  func(m *manifest.Manifest) { m.Tool = "" },
		wantErr: "tool must not be empty",
	},
	{
		name:    "unknown schema version",
		mutate:  func(m *manifest.Manifest) { m.Version = manifest.Version + 1 },
		wantErr: "unsupported manifest version 2",
	},
	{
		name:    "empty entries list",
		mutate:  func(m *manifest.Manifest) { m.Entries = nil },
		wantErr: "entries must not be empty",
	},
	{
		name: "duplicate family and path",
		mutate: func(m *manifest.Manifest) {
			m.Entries = append(m.Entries, validEntry())
		},
		wantErr: "duplicate entry ui.data data/health.clx",
	},
	{
		name:    "unknown format",
		mutate:  func(m *manifest.Manifest) { m.Entries[0].Format = "gif" },
		wantErr: "unknown format \"gif\"",
	},
	{
		name:    "sprite without frames",
		mutate:  func(m *manifest.Manifest) { m.Entries[0].Frames = 0 },
		wantErr: "frames must be > 0 for format clx",
	},
	{
		name: "palette with frames",
		mutate: func(m *manifest.Manifest) {
			m.Entries[0].Format = "pal"
			m.Entries[0].Frames = 1
			m.Entries[0].Width = 0
			m.Entries[0].WidthSource = ""
		},
		wantErr: "frames must be 0 for format pal",
	},
	{
		name:    "sprite without width",
		mutate:  func(m *manifest.Manifest) { m.Entries[0].Width = 0 },
		wantErr: "width must be > 0 for format clx",
	},
	{
		name: "font family exempt from width",
		mutate: func(m *manifest.Manifest) {
			m.Entries[0].Family = "font.small"
			m.Entries[0].Width = 0
		},
		wantErr: "",
	},
	{
		name: "explicit widths exempt from width",
		mutate: func(m *manifest.Manifest) {
			m.Entries[0].Width = 0
			m.Entries[0].Widths = []int{5, 7, 0}
		},
		wantErr: "",
	},
	{
		name:    "width without source",
		mutate:  func(m *manifest.Manifest) { m.Entries[0].WidthSource = "" },
		wantErr: "width_source must not be empty when width is set",
	},
	{
		name:    "malformed loader reference",
		mutate:  func(m *manifest.Manifest) { m.Entries[0].Loader = "diablo.cpp" },
		wantErr: "loader \"diablo.cpp\" is not a Source/... reference with a line number",
	},
	{
		name:    "empty family",
		mutate:  func(m *manifest.Manifest) { m.Entries[0].Family = "" },
		wantErr: "family must not be empty",
	},
	{
		name:    "absolute path",
		mutate:  func(m *manifest.Manifest) { m.Entries[0].Path = "/data/health.clx" },
		wantErr: "path must be relative with forward slashes",
	},
	{
		name:    "backslash path",
		mutate:  func(m *manifest.Manifest) { m.Entries[0].Path = `data\health.clx` },
		wantErr: "path must be relative with forward slashes",
	},
	{
		name: "negative crop width",
		mutate: func(m *manifest.Manifest) {
			m.Entries[0].Widths = []int{5, -1}
		},
		wantErr: "widths must be >= 0",
	},
}

// validateCase is one Validate table row.
type validateCase struct {
	name    string
	mutate  func(*manifest.Manifest)
	wantErr string
}

func TestValidate(t *testing.T) {
	for _, test := range validateCases {
		t.Run(test.name, func(t *testing.T) {
			// given the baseline manifest with this row's mutation applied
			m := validManifest()
			test.mutate(&m)

			// when the manifest is validated
			err := m.Validate()

			// then the error matches the table row
			assertValidateError(t, err, test.wantErr)
		})
	}
}

func TestValidateReportsEveryViolation(t *testing.T) {
	// given a manifest missing its tool and carrying an unknown format and loader
	m := validManifest()
	m.Tool = ""
	m.Entries[0].Format = "gif"
	m.Entries[0].Loader = "nowhere"

	// when the manifest is validated
	message := m.Validate().Error()

	// then every violation is reported, not just the first
	for _, want := range []string{"tool must not be empty", "unknown format", "loader"} {
		if !strings.Contains(message, want) {
			t.Errorf("Validate() = %q, want it to also contain %q", message, want)
		}
	}
}

func TestMissing(t *testing.T) {
	// given a manifest whose entries are present, present as a directory, absent
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, "data/health.clx"))
	writeFixture(t, filepath.Join(root, "plrgfx/warrior/whu/whufm.trn"))
	if err := os.MkdirAll(filepath.Join(root, "ui_art/creditsw.clx"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	m := validManifest()
	absent := validEntry()
	absent.Path = "ui_art/creditsw.clx" // present, but as a directory
	second := validEntry()
	second.Family = "ui.art"
	second.Path = "ui_art/missing.clx"
	m.Entries = append(m.Entries, absent, second)

	// when the present entries are collected
	got := m.Missing(root)

	// then the directory and the absent entry are reported with their loaders
	if len(got) != 2 {
		t.Fatalf("Missing() returned %d entries, want 2: %+v", len(got), got)
	}
	if got[0].Path != "ui_art/creditsw.clx" || got[0].Loader != "Source/diablo.cpp:1359" {
		t.Errorf("Missing()[0] = %+v, want the directory entry with its loader", got[0])
	}
	if got[1].Path != "ui_art/missing.clx" {
		t.Errorf("Missing()[1] = %+v, want ui_art/missing.clx", got[1])
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	// given a valid manifest
	m := validManifest()

	// when it is written, read back and written again
	var first bytes.Buffer
	if err := manifest.Write(&first, m); err != nil {
		t.Fatalf("Write() = %v", err)
	}
	written := first.String()

	read, err := manifest.Read(strings.NewReader(written))
	if err != nil {
		t.Fatalf("Read() = %v", err)
	}
	if err := read.Validate(); err != nil {
		t.Fatalf("round-tripped manifest does not validate: %v", err)
	}

	var second bytes.Buffer
	if err := manifest.Write(&second, read); err != nil {
		t.Fatalf("Write() = %v", err)
	}
	// then the second rendering is byte-identical and the result validates
	if written != second.String() {
		t.Fatalf(
			"round trip is not stable:\n--- first ---\n%s\n--- second ---\n%s",
			written,
			second.String(),
		)
	}
}

func TestWriteIsIndentedYAML(t *testing.T) {
	// given a valid manifest
	var out bytes.Buffer
	if err := manifest.Write(&out, validManifest()); err != nil {
		t.Fatalf("Write() = %v", err)
	}

	// when the rendered YAML is inspected
	body := out.String()

	// then the entry list is indented by two spaces
	if !strings.Contains(body, "\n  - family: ui.data\n") {
		t.Errorf("Write() did not emit a 2-space indented entry list:\n%s", body)
	}
}

func TestFinalizeSortsAndSummarizes(t *testing.T) {
	// given an unsorted manifest with one bundled flat entry
	m := manifest.Manifest{
		Tool:    manifest.Tool,
		Version: manifest.Version,
		Entries: []manifest.Entry{
			{
				Family: "ui.data", Path: "data/b.clx", Format: "clx", Frames: 1, Width: 1,
				WidthSource: "Source/a.cpp:1", Loader: "Source/a.cpp:2",
			},
			{
				Family: "ui.data", Path: "data/a.clx", Format: "clx", Frames: 1, Width: 1,
				WidthSource: "Source/a.cpp:1", Loader: "Source/a.cpp:2",
			},
			{
				Family: "pal", Path: "ui_art/diablo.pal", Format: "pal",
				Loader: "Source/a.cpp:3", Bundled: true,
			},
		},
	}

	// when it is finalized
	m.Finalize()

	// then entries are family-sorted, summarized and still valid
	if got := m.Entries[0].Path; got != "ui_art/diablo.pal" {
		t.Errorf(
			"Finalize() first path = %q, want ui_art/diablo.pal (family pal sorts before ui.data)",
			got,
		)
	}
	if got := m.Entries[2].Path; got != "data/b.clx" {
		t.Errorf("Finalize() last path = %q, want data/b.clx", got)
	}
	want := manifest.Summary{Entries: 3, Families: 2, Bundled: 1}
	if m.Summary != want {
		t.Errorf("Finalize() summary = %+v, want %+v", m.Summary, want)
	}
	if err := m.Validate(); err != nil {
		t.Errorf("finalized manifest does not validate: %v", err)
	}
}

func writeFixture(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("fixture"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

// validEntry is the baseline every table entry starts from; each case mutates
// only the field it is about.
func validEntry() manifest.Entry {
	return manifest.Entry{
		Family:      "ui.data",
		Path:        "data/health.clx",
		Format:      "clx",
		Frames:      1,
		Width:       264,
		WidthSource: "Source/items.cpp:2424",
		Loader:      "Source/diablo.cpp:1359",
	}
}

// validManifest is a two-entry manifest that passes Validate.
func validManifest() manifest.Manifest {
	second := validEntry()
	second.Family = "trn.plrgfx"
	second.Path = "plrgfx/warrior/whu/whufm.trn"
	second.Format = "trn"
	second.Frames = 0
	second.Width = 0
	second.WidthSource = ""

	return manifest.Manifest{
		Tool:    manifest.Tool,
		Version: manifest.Version,
		Entries: []manifest.Entry{validEntry(), second},
	}
}

// assertValidateError compares a Validate error against the expected fragment.
func assertValidateError(t *testing.T, err error, wantErr string) {
	t.Helper()

	if wantErr == "" {
		if err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}

		return
	}
	if err == nil {
		t.Fatalf("Validate() = nil, want an error containing %q", wantErr)
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("Validate() = %v, want it to contain %q", err, wantErr)
	}
}
