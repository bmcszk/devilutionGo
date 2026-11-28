// Package audit is the data-completeness gate over the committed TSV corpora:
// it walks both trees, proves every .tsv is a declared part of the closed
// world, and checks each registered table's row count and first/last parsed
// row against the file and the committed manifest.
package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// Corpus roots walked by the audit, relative to the repo root.
const (
	// CorpusGamedata is the game corpus: base rows plus the hf override rows.
	CorpusGamedata = "internal/testdata/gamedata"
	// CorpusUpstream is the mirrored upstream txtdata corpus plus its probes.
	CorpusUpstream = "internal/testdata/assets/txtdata"
)

// Dispositions reported for one committed .tsv file.
const (
	// StatusRegistered means a Go parser reads the file; rows and keys are checked.
	StatusRegistered = "REGISTERED"
	// StatusUnported means an upstream game table has no Go parser yet.
	StatusUnported = "UNPORTED"
	// StatusGolden means the TSV is a code-produced expectation, not an input.
	StatusGolden = "GOLDEN"
	// StatusProbe means the TSV feeds a loader probe, not the game.
	StatusProbe = "PROBE"
)

// Parser returns the identity of every data row the Go table holds, in file order.
type Parser func(data []byte) ([]string, error)

// Family declares one glob of committed .tsv files: the manifest table they
// belong to and the parser that must agree with them.
type Family struct {
	Corpus  string // corpus root the glob is relative to
	Glob    string // path.Match pattern, slash separated
	Table   string
	Variant string
	KeyCol  int    // identity column; the zero value is the file's first column
	Parse   Parser // nil => declared but not an input table (Status says why)
	Status  string // empty => derived from Parse
	Note    string
}

// Row is the audit's line for one committed file.
type Row struct {
	Corpus      string
	File        string
	Table       string
	Variant     string
	Status      string
	RowsTSV     int // data rows measured on disk
	RowsGo      int // rows the parser produced; -1 without a parser
	WantRows    int // manifest expectation; -1 without a manifest entry
	FirstRowSHA string
	LastRowSHA  string
	Note        string
}

// Report is one audit run: a line per committed file plus every problem found.
type Report struct {
	Rows     []Row
	Problems []string
}

// Entry is one measured expectation for one committed file.
type Entry struct {
	Corpus      string `json:"corpus"`
	File        string `json:"file"`
	Table       string `json:"table"`
	Variant     string `json:"variant"`
	WantRows    int    `json:"wantRows"`
	FirstRowSHA string `json:"firstRowSha"`
	LastRowSHA  string `json:"lastRowSha"`
}

// Manifest is the measured expectation set, with the command that produced it.
type Manifest struct {
	Measure string  `json:"measureCommand"`
	Entries []Entry `json:"entries"`
}

// Audit checks both committed corpora under repoRoot against the manifest.
func Audit(repoRoot, manifestPath string) (Report, error) {
	return Run(repoRoot, manifestPath, Families())
}

// Run is Audit with an explicit registry, so a test can script the cases the
// committed corpus cannot contain.
func Run(repoRoot, manifestPath string, families []Family) (Report, error) {
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return Report{}, err
	}

	report := Report{Problems: globProblems(families)}
	declared := make(map[int]bool, len(families))
	seen := make(map[string]bool, len(families))

	for _, corpus := range corpora() {
		files, err := listTSV(filepath.Join(repoRoot, filepath.FromSlash(corpus)))
		if err != nil {
			return Report{}, err
		}

		for _, file := range files {
			collect(&report, repoRoot, corpus, file, families, declared, seen, manifest)
		}
	}

	report.Problems = append(report.Problems, missingFiles(families, declared)...)
	report.Problems = append(report.Problems, strayEntries(manifest, seen)...)
	slices.Sort(report.Problems)

	return report, nil
}

// OK reports whether the audit found no problem.
func (r Report) OK() bool { return len(r.Problems) == 0 }

// Text renders the report as a tab-separated, diffable table: one header line,
// one line per committed file, then one "! " line per problem.
func (r Report) Text() string {
	lines := make([]string, 0, 1+len(r.Rows)+len(r.Problems))
	lines = append(lines, "corpus\tfile\ttable\tvariant\tstatus\trows_tsv\trows_go\twant_rows"+
		"\tfirst_row_sha\tlast_row_sha\tnote")

	for _, row := range r.Rows {
		lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s\t%s\t%s",
			row.Corpus, row.File, row.Table, row.Variant, row.Status,
			row.RowsTSV, row.RowsGo, row.WantRows, row.FirstRowSHA, row.LastRowSHA, row.Note))
	}

	for _, problem := range r.Problems {
		lines = append(lines, "! "+problem)
	}

	return strings.Join(lines, "\n") + "\n"
}

// Residuals are the parts of the data-completeness story this gate does not
// cover, each with the bean that owns it.
func Residuals() []string {
	return []string{
		"AiProc/AI-constant registry counts: behaviour, not a table -- PlaceGroup and its " +
			"constants landed in internal/diablo/gamedata/monsterspawn (v791 data half); the " +
			"AI registry counts stay open, owned by \"Monster pack placement + behaviour constants\" " +
			"(devilGo-v791)",
		"hf-delta expectations (which hf table adds or overrides which row): " +
			"owned by \"Hellfire delta audit\" (devilGo-cciu); this gate asserts hf row " +
			"counts and keys against the base parsers only",
		"lighttables/*.tsv are code-produced goldens (internal/light), so they are " +
			"enumerated for the closed world but carry no parser expectation",
	}
}

// LoadManifest decodes the manifest at path; read and decode failures carry the path.
func LoadManifest(manifestPath string) (Manifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest %q: %w", manifestPath, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest %q: %w", manifestPath, err)
	}

	return manifest, nil
}

// corpora are the roots Run walks, in report order.
func corpora() []string { return []string{CorpusGamedata, CorpusUpstream} }

// entry finds the expectation for one corpus file.
func (m Manifest) entry(corpus, file string) (Entry, bool) {
	for _, e := range m.Entries {
		if e.Corpus == corpus && e.File == file {
			return e, true
		}
	}

	return Entry{}, false
}

// collect audits one committed file: registry match, then parser, then manifest.
func collect(
	report *Report, repoRoot, corpus, file string,
	families []Family, declared map[int]bool, seen map[string]bool, manifest Manifest,
) {
	idx, family, problems := match(families, corpus, file)
	if len(problems) > 0 {
		report.Problems = append(report.Problems, problems...)
		return
	}

	declared[idx] = true
	seen[corpus+"/"+file] = true

	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(corpus), filepath.FromSlash(file)))
	if err != nil {
		report.Problems = append(report.Problems, fmt.Sprintf("%s: %v", label(family, corpus, file), err))
		return
	}

	raw := dataRows(data)
	row := Row{
		Corpus: corpus, File: file,
		Table: family.Table, Variant: family.Variant,
		Status:  statusOf(family),
		RowsTSV: len(raw), RowsGo: -1, WantRows: -1,
		Note: family.Note,
	}
	if len(raw) > 0 {
		row.FirstRowSHA = rowSHA(raw[0])
		row.LastRowSHA = rowSHA(raw[len(raw)-1])
	}

	if family.Parse != nil {
		checkParsed(report, &row, family, raw, data)
	}

	checkManifest(report, &row, manifest)
	report.Rows = append(report.Rows, row)
}

// checkParsed compares the parsed Go rows with the file they were read from.
func checkParsed(report *Report, row *Row, family Family, raw [][]string, data []byte) {
	name := label(family, row.Corpus, row.File)

	keys, err := family.Parse(data)
	if err != nil {
		report.Problems = append(report.Problems, fmt.Sprintf("%s: parse: %v", name, err))
		return
	}

	row.RowsGo = len(keys)

	if row.RowsGo != row.RowsTSV {
		report.Problems = append(report.Problems,
			fmt.Sprintf("%s: parsed %d rows, file has %d", name, row.RowsGo, row.RowsTSV))

		return
	}

	if len(keys) == 0 {
		report.Problems = append(report.Problems, fmt.Sprintf("%s: parsed no rows", name))
		return
	}

	last := len(keys) - 1

	if got, want := keys[0], cell(raw[0], family.KeyCol); got != want {
		report.Problems = append(report.Problems,
			fmt.Sprintf("%s: first parsed row = %q, file column %d = %q", name, got, family.KeyCol, want))
	}

	if got, want := keys[last], cell(raw[last], family.KeyCol); got != want {
		report.Problems = append(report.Problems,
			fmt.Sprintf("%s: last parsed row = %q, file column %d = %q", name, got, family.KeyCol, want))
	}
}

// checkManifest compares the measured file with its committed expectation.
func checkManifest(report *Report, row *Row, manifest Manifest) {
	name := label(Family{Table: row.Table, Variant: row.Variant}, row.Corpus, row.File)

	entry, ok := manifest.entry(row.Corpus, row.File)
	if !ok {
		if row.Status == StatusRegistered {
			report.Problems = append(report.Problems, fmt.Sprintf("%s: registered table has no manifest entry", name))
		}

		return
	}

	row.WantRows = entry.WantRows

	if entry.WantRows != row.RowsTSV {
		report.Problems = append(report.Problems,
			fmt.Sprintf("%s: manifest wantRows = %d, file has %d", name, entry.WantRows, row.RowsTSV))
	}

	if entry.FirstRowSHA != row.FirstRowSHA {
		report.Problems = append(report.Problems,
			fmt.Sprintf("%s: manifest firstRowSha = %s, file has %s", name, entry.FirstRowSHA, row.FirstRowSHA))
	}

	if entry.LastRowSHA != row.LastRowSHA {
		report.Problems = append(report.Problems,
			fmt.Sprintf("%s: manifest lastRowSha = %s, file has %s", name, entry.LastRowSHA, row.LastRowSHA))
	}
}

// match finds the one family that declares a file.
func match(families []Family, corpus, file string) (int, Family, []string) {
	idx, matched := -1, 0

	var family Family

	for i, candidate := range families {
		if !declares(candidate, corpus, file) {
			continue
		}

		idx, family = i, candidate
		matched++
	}

	switch matched {
	case 0:
		return 0, Family{}, []string{fmt.Sprintf("unregistered file: %s/%s", corpus, file)}
	case 1:
		return idx, family, nil
	default:
		return 0, Family{}, []string{fmt.Sprintf("file declared by %d families: %s/%s", matched, corpus, file)}
	}
}

// declares reports whether family's glob claims corpus/file.
func declares(family Family, corpus, file string) bool {
	if family.Corpus != corpus {
		return false
	}

	ok, err := path.Match(family.Glob, file)

	return err == nil && ok
}

// globProblems reports every glob that does not compile, so a typo cannot
// silently shrink the closed world.
func globProblems(families []Family) []string {
	var problems []string

	for _, family := range families {
		if _, err := path.Match(family.Glob, ""); err != nil {
			problems = append(problems,
				fmt.Sprintf("%s: bad glob %q: %v", family.Table, family.Glob, err))
		}
	}

	return problems
}

// missingFiles reports every declaration with no file behind it.
func missingFiles(families []Family, declared map[int]bool) []string {
	var problems []string

	for i, family := range families {
		if declared[i] {
			continue
		}

		problems = append(problems,
			fmt.Sprintf("%s/%s: declared but no file matches %s/%s",
				family.Table, family.Variant, family.Corpus, family.Glob))
	}

	return problems
}

// strayEntries reports every manifest entry with no file behind it.
func strayEntries(manifest Manifest, seen map[string]bool) []string {
	var problems []string

	for _, entry := range manifest.Entries {
		if seen[entry.Corpus+"/"+entry.File] {
			continue
		}

		problems = append(problems,
			fmt.Sprintf("%s/%s: manifest entry has no file at %s/%s",
				entry.Table, entry.Variant, entry.Corpus, entry.File))
	}

	return problems
}

// listTSV returns every .tsv under root as a sorted slash path.
func listTSV(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(walked string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %q: %w", walked, walkErr)
		}

		rel, ok, err := tsvRel(root, walked, entry)
		if err != nil {
			return err
		}

		if ok {
			files = append(files, rel)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %q: %w", root, err)
	}

	slices.Sort(files)

	return files, nil
}

// tsvRel reports the corpus-relative path of one walked .tsv entry.
func tsvRel(root, walked string, entry fs.DirEntry) (string, bool, error) {
	if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tsv") {
		return "", false, nil
	}

	rel, err := filepath.Rel(root, walked)
	if err != nil {
		return "", false, fmt.Errorf("relative path for %q: %w", walked, err)
	}

	return filepath.ToSlash(rel), true, nil
}

// dataRows splits a TSV file into data rows: the header is dropped, a trailing
// CR is trimmed (hf tables are CRLF) and blank lines are skipped.
func dataRows(data []byte) [][]string {
	lines := strings.Split(string(data), "\n")

	rows := make([][]string, 0, len(lines))
	for i, line := range lines {
		if i == 0 {
			continue
		}

		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}

		rows = append(rows, strings.Split(line, "\t"))
	}

	return rows
}

// cell returns column i of row, or "" when the row is short.
func cell(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}

	return row[i]
}

// rowSHA is the identity of one raw data row: sha256 over its tab-joined cells.
func rowSHA(row []string) string {
	sum := sha256.Sum256([]byte(strings.Join(row, "\t")))

	return hex.EncodeToString(sum[:])
}

// label is the problem prefix: table/variant plus the offending path.
func label(family Family, corpus, file string) string {
	return fmt.Sprintf("%s/%s %s/%s", family.Table, family.Variant, corpus, file)
}

// statusOf derives a family's disposition from its declaration.
func statusOf(family Family) string {
	if family.Status != "" {
		return family.Status
	}

	if family.Parse != nil {
		return StatusRegistered
	}

	return StatusUnported
}
