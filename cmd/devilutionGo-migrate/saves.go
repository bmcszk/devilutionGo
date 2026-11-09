package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

const (
	// defaultGlobalDBName is the shared stash database the import writes
	// (ADR-003: one global database for the shared stash).
	defaultGlobalDBName = "global.db"
	// heroDBSuffix is the per-hero live save extension.
	heroDBSuffix = ".db"
	// savesJournalName is the checkpoint file inside the output dir; a name of
	// its own, so the saves tree can share a directory with the assets tree.
	savesJournalName = ".migrate-saves-state.json"
)

const savesUsage = `saves flags: --dir <dir> --out <dir> [--dry-run]
  --dir    legacy save directory holding the .sv saves and the stash archives
  --out    output dir for the per-hero .db files plus global.db
           (default: the --dir itself; no settings lane yet)
`

// stashSaveNames are the shared stash archives GetStashSavePath produces
// (Source/pfile.cpp:94-104; persist isStashSaveName).
var stashSaveNames = []string{"stash.sv", "stash_spawn.sv"}

// savesConfig is one `saves` invocation.
type savesConfig struct {
	dir    string
	out    string
	dryRun bool
}

// savesPlan is the classified content of a legacy save directory, in the name
// order os.ReadDir already returns so a resume run walks identically.
type savesPlan struct {
	heroes  []string // hero saves: single_/spawn_/multi_/share_ with a slot
	stashes []string // shared stash archives
}

// savesReport is what one saves run produced, in the terms the operator needs:
// what was imported, what the journal skipped, what failed, and the per-file
// checksums that identify the store files.
type savesReport struct {
	imported  int
	skipped   int
	failures  []string
	checksums []string
	summary   string
	dryRun    bool
}

// failed reports whether any save was rejected.
func (r savesReport) failed() bool { return len(r.failures) > 0 }

// write prints the failures, the pass counts and the checksum summary.
func (r savesReport) write(stdout, stderr io.Writer) {
	for _, failure := range r.failures {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate: saves: %s\n", failure)
	}

	mode := "migrated"
	if r.dryRun {
		mode = "planned"
	}

	_, _ = fmt.Fprintf(stdout, "%s %d save(s), skipped %d already-migrated, %d failure(s)\n",
		mode, r.imported, r.skipped, len(r.failures))

	for _, line := range r.checksums {
		_, _ = fmt.Fprintln(stdout, line)
	}

	if r.summary != "" {
		_, _ = fmt.Fprintf(stdout, "summary %s\n", r.summary)
	}
}

// runSavesCommand migrates the legacy saves of one directory into the live
// store: per-hero .db files plus global.db (ADR-014 onto ADR-003).
func runSavesCommand(args []string, stdout, stderr io.Writer) int {
	cfg, err := parseSavesFlags(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = fmt.Fprint(stderr, savesUsage)

			return exitOK
		}

		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate: saves: %v\n", err)

		return exitUsage
	}

	report, err := migrateSaves(context.Background(), cfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "devilutionGo-migrate: saves: %v\n", err)

		return exitFailure
	}

	report.write(stdout, stderr)

	if report.failed() {
		return exitFailure
	}

	return exitOK
}

// parseSavesFlags reads the saves flags. The flag package writes its own
// diagnostics to stderr, so a bad flag needs no extra message.
func parseSavesFlags(args []string, stderr io.Writer) (savesConfig, error) {
	flags := flag.NewFlagSet("devilutionGo-migrate saves", flag.ContinueOnError)
	flags.SetOutput(stderr)

	dir := flags.String("dir", "", "legacy save directory holding the .sv saves and the stash archives")
	out := flags.String("out", "", "output dir for the per-hero .db files plus global.db")
	dryRun := flags.Bool("dry-run", false, "report the plan without writing files")

	if err := flags.Parse(args); err != nil {
		return savesConfig{}, fmt.Errorf("parse saves flags: %w", err)
	}

	cfg := savesConfig{dir: *dir, out: *out, dryRun: *dryRun}

	if cfg.dir == "" {
		return cfg, errors.New("--dir is required")
	}

	if cfg.out == "" {
		cfg.out = cfg.dir
	}

	return cfg, nil
}

// classifySaves splits the legacy directory into hero saves and stash archives.
// Anything else in the directory (info files, foreign extensions) is not a
// save and is left alone.
func classifySaves(dir string) (savesPlan, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return savesPlan{}, fmt.Errorf("read save directory %s: %w", dir, err)
	}

	plan := savesPlan{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		switch {
		case isHeroSave(entry.Name()):
			plan.heroes = append(plan.heroes, entry.Name())
		case slices.Contains(stashSaveNames, entry.Name()):
			plan.stashes = append(plan.stashes, entry.Name())
		default:
			// info files and foreign extensions are not saves and stay
		}
	}

	return plan, nil
}

// isHeroSave reports whether a file name is the GetSavePath form of a hero
// save: a GetSavePath prefix, a slot and the .sv extension (pfile.cpp:81-92).
func isHeroSave(name string) bool {
	_, _, ok := persist.ParseHeroSaveName(name)

	return ok
}

// heroDBPath is the per-hero live save a legacy hero save becomes: the save
// name carries over, so single_0.sv lands beside its source as single_0.db.
func heroDBPath(out, name string) string {
	return filepath.Join(out, name[:len(name)-len(".sv")]+heroDBSuffix)
}

// globalDBPath is the shared stash database of an output dir.
func globalDBPath(out string) string {
	return filepath.Join(out, defaultGlobalDBName)
}
