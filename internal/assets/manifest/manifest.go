// Package manifest models the machine-checked graphics asset inventory: one
// entry per file the renderer asks for, with the C++ loader call site that asks
// for it and the source of its frame dimensions.
//
// The package is engine-layer only: it knows generic invariants (a sprite has
// frames and a width, a palette does not) and nothing about Diablo families.
// Family enumeration lives in cmd/dgo-assets, which may import internal/diablo.
package manifest

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Version is the schema version this package writes and reads. A manifest with
// any other version is refused rather than guessed at.
const Version = 1

// Tool is the writer name recorded in every manifest this package produces.
const Tool = "dgo-assets"

// loaderPattern matches a C++ loader reference the way the sources spell it:
// `Source/dir/file.cpp:1234`.
var loaderPattern = regexp.MustCompile(`^Source/[A-Za-z0-9_./-]+\.(cpp|hpp|h):[0-9]+$`)

// spriteFormats are the formats whose entries carry frames; every other format
// is a flat file (one palette, one translation table, one tile record).
var spriteFormats = map[string]bool{
	"cel": true,
	"cl2": true,
	"clx": true,
	"pcx": true,
}

// flatFormats are the formats that must not carry a frame count.
var flatFormats = map[string]bool{
	"pal": true,
	"trn": true,
	"til": true,
	"min": true,
	"sol": true,
	"dun": true,
	"amp": true,
}

// Entry is one graphics file the renderer loads.
type Entry struct {
	// Family groups entries by the subsystem that loads them, e.g. "ui.data".
	Family string `yaml:"family"`
	// Path is the archive-relative path the loader builds, forward slashes only.
	Path string `yaml:"path"`
	// Format is the on-disk container the path resolves to.
	Format string `yaml:"format"`
	// Bundled is true when the file ships in the devilutionX asset archive
	// rather than coming from a retail MPQ.
	Bundled bool `yaml:"bundled"`
	// Frames is the sprite-list length; zero for flat formats.
	Frames int `yaml:"frames,omitempty"`
	// Width is the frame width in pixels; zero for flat formats and fonts.
	Width int `yaml:"width,omitempty"`
	// WidthSource names the C++ site the width comes from, e.g.
	// "Source/items.cpp:2424" or a `-widths.txt` file reference.
	WidthSource string `yaml:"width_source,omitempty"`
	// Widths holds per-frame crop widths when the source is a widths file.
	Widths []int `yaml:"widths,omitempty"`
	// Loader is the C++ call site that requests this file.
	Loader string `yaml:"loader"`
}

// Missing is one manifest entry whose path is absent from the asset root.
type Missing struct {
	Path   string
	Loader string
}

// Summary is the roll-up a reader checks first: how many entries, families and
// bundled files the manifest describes.
type Summary struct {
	Entries  int `yaml:"entries"`
	Families int `yaml:"families"`
	Bundled  int `yaml:"bundled"`
}

// Manifest is the whole inventory.
type Manifest struct {
	Tool    string  `yaml:"tool"`
	Version int     `yaml:"version"`
	Entries []Entry `yaml:"entries"`
	Summary Summary `yaml:"summary"`
}

// Finalize puts the manifest in its canonical order and recomputes Summary.
// It is idempotent, so a producer may call it before every write.
func (m *Manifest) Finalize() {
	slices.SortFunc(m.Entries, func(a, b Entry) int {
		if byFamily := strings.Compare(a.Family, b.Family); byFamily != 0 {
			return byFamily
		}
		return strings.Compare(a.Path, b.Path)
	})

	families := make(map[string]struct{}, len(m.Entries))
	bundled := 0
	for _, entry := range m.Entries {
		families[entry.Family] = struct{}{}
		if entry.Bundled {
			bundled++
		}
	}

	m.Summary = Summary{Entries: len(m.Entries), Families: len(families), Bundled: bundled}
}

// Validate reports every invariant violation in one error, so a producer sees
// the whole list instead of fixing one entry per run.
func (m Manifest) Validate() error {
	problems := make([]error, 0, len(m.Entries)+3)

	if m.Tool == "" {
		problems = append(problems, errors.New("manifest: tool must not be empty"))
	}
	if m.Version != Version {
		problems = append(problems, fmt.Errorf("manifest: unsupported manifest version %d, want %d", m.Version, Version))
	}
	if len(m.Entries) == 0 {
		problems = append(problems, errors.New("manifest: entries must not be empty"))
	}

	seen := make(map[string]struct{}, len(m.Entries))
	for _, entry := range m.Entries {
		problems = append(problems, entry.validate(seen)...)
	}

	return errors.Join(problems...)
}

// Missing lists the entries whose path is absent under root. A path that exists
// as a directory is reported as missing: loaders read files, not directories.
func (m Manifest) Missing(root string) []Missing {
	var missing []Missing
	for _, entry := range m.Entries {
		if isFile(filepath.Join(root, filepath.FromSlash(entry.Path))) {
			continue
		}
		missing = append(missing, Missing{Path: entry.Path, Loader: entry.Loader})
	}

	return missing
}

// Write emits the manifest as 2-space indented YAML.
func Write(w io.Writer, m Manifest) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	return nil
}

// Read parses a manifest written by Write. It does not validate: the caller
// decides whether a broken manifest is fatal.
func Read(r io.Reader) (Manifest, error) {
	var m Manifest
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&m); err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}

	return m, nil
}

// validate returns every problem with this entry, seen carries the
// (family, path) keys claimed by earlier entries.
func (e Entry) validate(seen map[string]struct{}) []error {
	problems := make([]error, 0, 5)
	where := e.describe()

	problems = append(problems, e.validateIdentity(seen)...)
	problems = append(problems, e.validateFormat(where)...)
	problems = append(problems, e.validateDimensions(where)...)
	problems = append(problems, e.validateLoader()...)

	if e.Width > 0 && e.WidthSource == "" {
		problems = append(problems, fmt.Errorf("%s: width_source must not be empty when width is set", where))
	}
	for _, width := range e.Widths {
		if width < 0 {
			problems = append(problems, fmt.Errorf("%s: widths must be >= 0, got %d", where, width))
			break
		}
	}

	return problems
}

// validateIdentity checks the family/path pair and claims it in seen.
func (e Entry) validateIdentity(seen map[string]struct{}) []error {
	var problems []error

	if e.Family == "" {
		problems = append(problems, fmt.Errorf("%s: family must not be empty", e.describe()))
	}
	if !isRelativeSlashPath(e.Path) {
		problems = append(problems, fmt.Errorf("%s: path must be relative with forward slashes", e.describe()))
	}

	key := e.Family + " " + e.Path
	if _, claimed := seen[key]; claimed {
		problems = append(problems, fmt.Errorf("%s: duplicate entry %s", e.describe(), key))
	}
	seen[key] = struct{}{}

	return problems
}

// validateFormat checks the container name and that a frame count matches it.
func (e Entry) validateFormat(where string) []error {
	switch {
	case spriteFormats[e.Format]:
		if e.Frames <= 0 {
			return []error{fmt.Errorf("%s: frames must be > 0 for format %s", where, e.Format)}
		}
	case flatFormats[e.Format]:
		if e.Frames != 0 {
			return []error{fmt.Errorf("%s: frames must be 0 for format %s", where, e.Format)}
		}
	default:
		return []error{fmt.Errorf("%s: unknown format %q", where, e.Format)}
	}

	return nil
}

// validateDimensions checks that a sprite carries a width, either inline or as
// per-frame crop widths. Font sheets store their own glyph widths.
func (e Entry) validateDimensions(where string) []error {
	if !spriteFormats[e.Format] || e.Width > 0 || len(e.Widths) > 0 {
		return nil
	}
	if strings.HasPrefix(e.Family, "font.") {
		return nil
	}

	return []error{fmt.Errorf("%s: width must be > 0 for format %s (or supply widths)", where, e.Format)}
}

// validateLoader checks the C++ call-site reference format.
func (e Entry) validateLoader() []error {
	if loaderPattern.MatchString(e.Loader) {
		return nil
	}

	return []error{fmt.Errorf("%s: loader %q is not a Source/... reference with a line number", e.describe(), e.Loader)}
}

// describe names the entry in error messages.
func (e Entry) describe() string {
	return fmt.Sprintf("entry %s (%s)", e.Path, e.Family)
}

// isRelativeSlashPath reports whether path is archive-relative and uses forward
// slashes only, which the manifest schema requires of every entry.
func isRelativeSlashPath(path string) bool {
	if path == "" || strings.HasPrefix(path, "/") || strings.Contains(path, `\`) {
		return false
	}

	return filepath.Clean(path) == path
}

// isFile reports whether path exists and is a regular file.
func isFile(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.Mode().IsRegular()
}
