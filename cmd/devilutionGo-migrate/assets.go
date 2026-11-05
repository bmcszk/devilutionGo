package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

const (
	// assetArchivesEnv names the directories holding the retail archives,
	// separated by os.PathListSeparator — the same lookup the extraction smoke
	// test uses (internal/assets/mpq_smoke_test.go).
	assetArchivesEnv = "DG_ASSET_ARCHIVES"
	// defaultAssetsOut is the assets dir the build embeds (ADR-005).
	defaultAssetsOut = "./assets-out"
	// archiveExtension selects the archives discovery picks up in a configured
	// directory; anything else in the directory is left alone.
	archiveExtension = ".mpq"
	// archivesSeparator separates the two paths in a clash message.
	archivesSeparator = " and "
	// unaddressableArchive is reported for an archive with no record names.
	unaddressableArchive = "skipped archive"
)

// assetsConfig is one `assets` invocation.
type assetsConfig struct {
	archives []string // explicit --mpq files; empty means "discover"
	out      string
	dryRun   bool
	env      string
}

// archiveFlag collects the repeatable --mpq flag.
type archiveFlag []string

// archiveRuns is one pass over the archive set: the report counts, the full
// manifest entry list, and the entries still to be written.
type archiveRuns struct {
	report  migrationReport
	entries []fileEntry
	pending []fileEntry
}

// migrationReport is what one assets run produced, in the terms the operator
// needs: what was written, what the journal let it skip, what failed, and the
// checksum summary that identifies the whole output.
type migrationReport struct {
	written       int
	skipped       int
	summary       string
	failures      []string
	unaddressable []string
	dryRun        bool
}

// runAssetsExtract migrates the configured archives and reports the exit code.
// It is the `assets extract` pass: the nested inventory commands dispatch before
// this point (ADR-019), so a bare `assets` lands here for compatibility.
func runAssetsExtract(args []string, stdout, stderr io.Writer) int {
	cfg, err := parseAssetsFlags(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = fmt.Fprint(stderr, assetsUsage)

			return exitOK
		}

		return exitUsage
	}

	archives, err := resolveArchives(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate: assets: %v\n", err)

		return exitFailure
	}

	report, err := migrateAssets(cfg, archives)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate: assets: %v\n", err)

		return exitFailure
	}

	report.write(stdout, stderr)

	if report.failed() {
		return exitFailure
	}

	return exitOK
}

// parseAssetsFlags reads the assets flags. The flag package writes its own
// diagnostics to stderr, so an error needs no extra message.
func parseAssetsFlags(args []string, stderr io.Writer) (assetsConfig, error) {
	flags := flag.NewFlagSet("devilutionGo-migrate assets extract", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var archives archiveFlag

	flags.Var(&archives, "mpq", "archive to extract (repeatable); overrides "+assetArchivesEnv)

	out := flags.String("out", defaultAssetsOut, "assets dir to write")
	dryRun := flags.Bool("dry-run", false, "report the plan without writing files")

	if err := flags.Parse(args); err != nil {
		return assetsConfig{}, fmt.Errorf("parse flags: %w", err)
	}

	return assetsConfig{
		archives: archives,
		out:      *out,
		dryRun:   *dryRun,
		env:      os.Getenv(assetArchivesEnv),
	}, nil
}

// String renders the collected --mpq values for the flag package.
func (a *archiveFlag) String() string { return strings.Join(*a, string(os.PathListSeparator)) }

// Set appends one --mpq value.
func (a *archiveFlag) Set(value string) error {
	*a = append(*a, value)

	return nil
}

// resolveArchives returns the archive files this run migrates, sorted by path so
// two runs visit them in the same order.
func resolveArchives(cfg assetsConfig) ([]string, error) {
	if len(cfg.archives) > 0 {
		return checkedArchives(cfg.archives)
	}

	archives := make([]string, 0)

	for _, dir := range filepath.SplitList(cfg.env) {
		if dir == "" {
			continue
		}

		found, err := archivesInDir(dir)
		if err != nil {
			return nil, err
		}

		archives = append(archives, found...)
	}

	if len(archives) == 0 {
		return nil, fmt.Errorf("no archives configured: pass --mpq or set %s to the dirs holding the MPQ files",
			assetArchivesEnv)
	}

	return checkedArchives(archives)
}

// checkedArchives verifies every resolved archive is readable and that no two
// share a base name: outputs mirror the archive base name, so a duplicate would
// silently overwrite another archive's tree.
func checkedArchives(archives []string) ([]string, error) {
	slices.Sort(archives)

	seen := make(map[string]string, len(archives))

	for _, archive := range archives {
		if _, err := os.Stat(archive); err != nil {
			return nil, fmt.Errorf("archive: %w", err)
		}

		base := strings.ToLower(filepath.Base(archive))
		if other, clash := seen[base]; clash {
			return nil, fmt.Errorf("archives %s%s%s share the base name %s",
				other, archivesSeparator, archive, base)
		}

		seen[base] = archive
	}

	return archives, nil
}

// archivesInDir lists the archive files directly inside dir, in name order.
func archivesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read archive dir: %w", err)
	}

	archives := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), archiveExtension) {
			continue
		}

		archives = append(archives, filepath.Join(dir, entry.Name()))
	}

	return archives, nil
}

// failed reports whether any record failed to extract.
func (r migrationReport) failed() bool { return len(r.failures) > 0 }

// write prints the run report: failures on stderr, then the counts and the
// summary on stdout.
func (r migrationReport) write(stdout, stderr io.Writer) {
	for _, failure := range r.failures {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate: assets: %s\n", failure)
	}

	for _, archive := range r.unaddressable {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate: assets: %s %s: %s\n",
			unaddressableArchive, archive, errUnaddressable)
	}

	mode := "wrote"
	if r.dryRun {
		mode = "planned"
	}

	_, _ = fmt.Fprintf(stdout, "%s %d file(s), skipped %d already-journaled file(s), %d failure(s)\n",
		mode, r.written, r.skipped, len(r.failures))
	_, _ = fmt.Fprintf(stdout, "summary sha256 %s\n", r.summary)
}

// migrateAssets extracts every archive, writes the outputs and the manifest, and
// checkpoints the work in the journal so a rerun resumes. A dry run computes the
// same report without touching the disk.
func migrateAssets(cfg assetsConfig, archives []string) (migrationReport, error) {
	state, err := loadJournal(cfg.out)
	if err != nil {
		return migrationReport{}, err
	}

	runs, err := runArchives(state, archives)
	if err != nil {
		return migrationReport{}, err
	}

	runs.report.dryRun = cfg.dryRun
	runs.report.summary = summarize(runs.entries).SHA256

	if cfg.dryRun {
		return runs.report, nil
	}

	if err := writeOutputs(cfg, state, archives, runs); err != nil {
		return migrationReport{}, err
	}

	return runs.report, nil
}

// writeOutputs writes the extracted files, then the manifest, then the journal:
// a crash leaves the journal untouched, so the next run redoes that archive
// rather than trusting files it never finished writing.
func writeOutputs(cfg assetsConfig, state *journal, archives []string, runs archiveRuns) error {
	if err := writeEntries(cfg.out, runs.pending); err != nil {
		return err
	}

	if err := writeManifest(cfg.out, buildManifest(archives, state, runs.entries)); err != nil {
		return err
	}

	return saveJournal(cfg.out, state)
}

// runArchives extracts or resumes every archive and returns the report counts,
// the full manifest entry list, and the entries still to be written.
func runArchives(state *journal, archives []string) (archiveRuns, error) {
	runs := archiveRuns{entries: make([]fileEntry, 0), pending: make([]fileEntry, 0)}

	for _, archive := range archives {
		extracted, err := extractOrResume(state, archive)
		if err != nil {
			return archiveRuns{}, err
		}

		if extracted.unaddressable != "" {
			runs.report.unaddressable = append(runs.report.unaddressable, extracted.unaddressable)

			continue
		}

		runs.report.failures = append(runs.report.failures, extracted.failures...)
		runs.entries = append(runs.entries, extracted.state.Files...)

		if extracted.resumed {
			runs.report.skipped += len(extracted.state.Files)

			continue
		}

		runs.report.written += len(extracted.state.Files)
		runs.pending = append(runs.pending, extracted.state.Files...)
	}

	slices.SortFunc(runs.entries, compareEntries)

	return runs, nil
}

// archiveNames returns the names that address archive's records: the archive's
// own (listfile) when it ships one, else the committed inventory (retail
// DIABDAT.MPQ and hellfire.mpq ship none, and MPQ decrypts by name).
func archiveNames(archive *assets.MPQArchive, path string) ([]string, error) {
	names, err := archive.Names()
	if err != nil {
		return nil, fmt.Errorf("%s: listfile: %w", filepath.Base(path), err)
	}

	if len(names) > 0 {
		return names, nil
	}

	inventory, known := assets.NameInventory(path)
	if !known {
		return nil, fmt.Errorf("%s %w", filepath.Base(path), errUnaddressable)
	}

	return inventory, nil
}
