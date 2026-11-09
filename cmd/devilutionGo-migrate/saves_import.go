package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// savesJournalVersion guards against reading a journal an older tool wrote,
// the same convention the assets journal uses.
const savesJournalVersion = 1

// dbEntry is one produced database file of a run: where it is and what it
// contains, so a resume can prove the output is still the recorded one.
type dbEntry struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

// saveState is what one save's conversion completed: the save's fingerprint
// and the database file it produced.
type saveState struct {
	Bytes    int64   `json:"bytes"`
	Modified int64   `json:"modified_unix_nano"`
	HeroDB   dbEntry `json:"hero_db"`
}

// savesJournal is the checkpoint file. Records are keyed by save file name;
// the fingerprints say whether the recorded conversions still describe the
// saves on disk, so a save that changed on disk is imported again instead of
// silently reused.
type savesJournal struct {
	Version  int                  `json:"version"`
	Dir      string               `json:"dir"`
	GlobalDB dbEntry              `json:"global_db"`
	Stashes  map[string]saveState `json:"stashes"`
	Saves    map[string]saveState `json:"saves"`
}

// emptySavesJournal returns a journal for a fresh output dir.
func emptySavesJournal(dir string) *savesJournal {
	return &savesJournal{
		Version: savesJournalVersion,
		Dir:     dir,
		Saves:   map[string]saveState{},
		Stashes: map[string]saveState{},
	}
}

// loadSavesJournal reads the checkpoint of an existing output dir. A missing
// file is a first run, not an error; an unreadable or foreign journal is an
// error, because guessing would write store files the journal no longer
// describes.
func loadSavesJournal(out, dir string) (*savesJournal, error) {
	loaded, err := readSavesJournal(out)
	if err != nil {
		return nil, err
	}

	if loaded.Dir != "" && loaded.Dir != dir {
		return emptySavesJournal(dir), nil
	}

	if loaded.Saves == nil {
		loaded.Saves = map[string]saveState{}
	}

	if loaded.Stashes == nil {
		loaded.Stashes = map[string]saveState{}
	}

	loaded.Dir = dir

	return loaded, nil
}

// readSavesJournal reads and parses the checkpoint file: absent means a fresh
// run, foreign content means the operator decides, not the tool.
func readSavesJournal(out string) (*savesJournal, error) {
	content, err := os.ReadFile(filepath.Join(out, savesJournalName))
	if errors.Is(err, os.ErrNotExist) {
		return emptySavesJournal(""), nil
	}

	if err != nil {
		return nil, fmt.Errorf("read journal: %w", err)
	}

	loaded := &savesJournal{}

	if err := json.Unmarshal(content, loaded); err != nil {
		return nil, fmt.Errorf("parse journal: %w", err)
	}

	if loaded.Version != savesJournalVersion {
		return nil, fmt.Errorf("journal version %d, want %d: delete %s to start over",
			loaded.Version, savesJournalVersion, filepath.Join(out, savesJournalName))
	}

	return loaded, nil
}

// saveSavesJournal checkpoints the run; the journal is written last so a crash
// leaves the store files it describes either incomplete on disk or absent from
// the journal, and the next run redoes that save.
func saveSavesJournal(out string, state *savesJournal) error {
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode journal: %w", err)
	}

	if err := writeOutput(filepath.Join(out, savesJournalName), append(content, '\n')); err != nil {
		return fmt.Errorf("write journal: %w", err)
	}

	return nil
}

// resume returns the recorded conversion for a hero save when it still
// describes the file on disk and the produced .db is intact; otherwise it
// reports false and the save is imported again.
func (j *savesJournal) resume(dir, name string) (saveState, bool) {
	recorded, known := j.Saves[name]
	if !known {
		return saveState{}, false
	}

	return recorded, stillDescribes(dir, name, recorded.Bytes, recorded.Modified) && intactRecord(recorded.HeroDB)
}

// resumeStash reports whether a recorded stash conversion is still current.
// Both stash archives share the one global database, so the recorded global
// digest decides, not the per-record output.
func (j *savesJournal) resumeStash(dir, name string) bool {
	recorded, known := j.Stashes[name]

	return known && stillDescribes(dir, name, recorded.Bytes, recorded.Modified) && intactRecord(j.GlobalDB)
}

// stillDescribes says whether the remembered size and mtime still describe the
// save file on disk.
func stillDescribes(dir, name string, bytes, modified int64) bool {
	current, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		return false
	}

	return current.Size() == bytes && current.ModTime().UnixNano() == modified
}

// intactRecord reports whether a produced database file is still the one the
// journal recorded.
func intactRecord(entry dbEntry) bool {
	if entry.SHA256 == "" {
		return false
	}

	body, err := os.ReadFile(filepath.Clean(entry.Path))
	if err != nil || int64(len(body)) != entry.Bytes {
		return false
	}

	return digestDB(body) == entry.SHA256
}

// recordSave checkpoints one converted hero save: the produced file is hashed
// and the source fingerprint stored with it before the next save starts.
func (j *savesJournal) recordSave(dir, name string, entry dbEntry) error {
	current, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("stat %s: %w", name, err)
	}

	j.Saves[name] = saveState{
		Bytes:    current.Size(),
		Modified: current.ModTime().UnixNano(),
		HeroDB:   entry,
	}

	return nil
}

// recordStash checkpoints one converted stash. Updating GlobalDB too keeps the
// resume proof valid even when the run stops before the end-of-run flush:
// the recorded digest is the global file as this import left it.
func (j *savesJournal) recordStash(dir, name string, entry dbEntry) error {
	current, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("stat %s: %w", name, err)
	}

	j.Stashes[name] = saveState{
		Bytes:    current.Size(),
		Modified: current.ModTime().UnixNano(),
		HeroDB:   entry,
	}
	j.GlobalDB = entry

	return nil
}

// record hashes one produced database file into the journal: path, size,
// checksum.
func record(path string) (dbEntry, error) {
	body, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return dbEntry{}, fmt.Errorf("read %s: %w", path, err)
	}

	return dbEntry{Path: path, Bytes: int64(len(body)), SHA256: digestDB(body)}, nil
}

// digestDB is the sha256 of one produced store file.
func digestDB(body []byte) string {
	sum := sha256.Sum256(body)

	return hex.EncodeToString(sum[:])
}

// migrateSaves imports the legacy saves of cfg.dir into the live store:
// per-hero .db files plus global.db under cfg.out, checkpointed after every
// save so a partial run resumes without redoing converted saves (ADR-003,
// ADR-014). A save that does not import is reported and the run continues.
func migrateSaves(ctx context.Context, cfg savesConfig) (savesReport, error) {
	plan, err := classifySaves(cfg.dir)
	if err != nil {
		return savesReport{}, err
	}

	if cfg.dryRun {
		return dryRunReport(cfg, plan), nil
	}

	if err := os.MkdirAll(cfg.out, 0o755); err != nil {
		return savesReport{}, fmt.Errorf("create output dir: %w", err)
	}

	state, err := loadSavesJournal(cfg.out, cfg.dir)
	if err != nil {
		return savesReport{}, err
	}

	report, err := importSaves(ctx, cfg, plan, state)
	if err != nil {
		return savesReport{}, err
	}

	if err := saveSavesJournal(cfg.out, state); err != nil {
		return savesReport{}, err
	}

	return report, nil
}

// dryRunReport sizes the run without touching the store: the save families the
// directory holds and every store file they would land in.
func dryRunReport(cfg savesConfig, plan savesPlan) savesReport {
	report := savesReport{imported: len(plan.heroes) + len(plan.stashes), dryRun: true}

	for _, name := range plan.heroes {
		report.checksums = append(report.checksums, "plan: "+name+" -> "+filepath.Base(heroDBPath(cfg.out, name)))
	}

	for _, name := range plan.stashes {
		report.checksums = append(report.checksums, "plan: "+name+" -> "+defaultGlobalDBName)
	}

	return report
}

// importSaves imports the planned saves one file at a time. Skipped saves are
// re-proven from the journal; converted saves land through persist and are
// checkpointed before the next file is touched.
func importSaves(ctx context.Context, cfg savesConfig, plan savesPlan, state *savesJournal) (savesReport, error) {
	report := savesReport{}

	if err := importHeroSaves(ctx, cfg, plan.heroes, state, &report); err != nil {
		return savesReport{}, err
	}

	if err := importStashSaves(ctx, cfg, plan.stashes, state, &report); err != nil {
		return savesReport{}, err
	}

	entry, err := record(globalDBPath(cfg.out))
	switch {
	case errors.Is(err, os.ErrNotExist):
		// No stash landed and this run never created global.db: the summary has
		// nothing to name.
	case err != nil:
		return savesReport{}, err
	default:
		state.GlobalDB = entry
		report.summary = "sha256 " + entry.SHA256 + " (" + defaultGlobalDBName + ")"
	}

	return report, nil
}

// importHeroSaves walks the hero saves, resuming recorded ones and importing
// the rest; every outcome lands in the report.
func importHeroSaves(
	ctx context.Context, cfg savesConfig, names []string, state *savesJournal, report *savesReport,
) error {
	for _, name := range names {
		if recorded, known := state.resume(cfg.dir, name); known {
			report.skipped++
			report.checksums = appendChecksum(report.checksums, name, recorded.HeroDB)

			continue
		}

		entry, failure := convertHeroSave(ctx, cfg, name)
		if failure != nil {
			report.failures = append(report.failures, *failure)

			continue
		}

		if err := checkpointSave(cfg, state, name, entry); err != nil {
			return err
		}

		report.imported++
		report.checksums = appendChecksum(report.checksums, name, entry)
	}

	return nil
}

// importStashSaves imports the shared stash archives into the global database.
func importStashSaves(
	ctx context.Context, cfg savesConfig, names []string, state *savesJournal, report *savesReport,
) error {
	for _, name := range names {
		if state.resumeStash(cfg.dir, name) {
			report.skipped++

			continue
		}

		entry, failure := convertStash(ctx, cfg, name)
		if failure != nil {
			report.failures = append(report.failures, *failure)

			continue
		}

		if err := checkpointStash(cfg, state, name, entry); err != nil {
			return err
		}

		report.imported++
	}

	return nil
}

// checkpointSave records a hero conversion and flushes the journal file.
func checkpointSave(cfg savesConfig, state *savesJournal, name string, entry dbEntry) error {
	if err := state.recordSave(cfg.dir, name, entry); err != nil {
		return err
	}

	return saveSavesJournal(cfg.out, state)
}

// checkpointStash records a stash conversion and flushes the journal file.
func checkpointStash(cfg savesConfig, state *savesJournal, name string, entry dbEntry) error {
	if err := state.recordStash(cfg.dir, name, entry); err != nil {
		return err
	}

	return saveSavesJournal(cfg.out, state)
}

// appendChecksum formats one per-file checksum line of the report.
func appendChecksum(lines []string, name string, entry dbEntry) []string {
	return append(lines, name+" -> "+filepath.Base(entry.Path)+" sha256 "+entry.SHA256)
}

// convertHeroSave imports one hero save into its per-hero .db file.
func convertHeroSave(ctx context.Context, cfg savesConfig, name string) (dbEntry, *string) {
	prefix, slot, _ := persist.ParseHeroSaveName(name)
	target := heroDBPath(cfg.out, name)

	store, err := persist.OpenSQLiteStore(ctx, target)
	if err != nil {
		return dbEntry{}, failureText(name, err)
	}

	defer func() { _ = store.Close() }()

	if err := persist.ImportSV(ctx, store, filepath.Join(cfg.dir, name), modeForPrefix(prefix), slot); err != nil {
		// The fresh target holds no hero row: remove it, so a corrupted save
		// leaves no store file that looks migrated.
		_ = os.Remove(filepath.Clean(target))

		return dbEntry{}, failureText(name, err)
	}

	entry, err := record(target)
	if err != nil {
		return dbEntry{}, failureText(name, err)
	}

	return entry, nil
}

// convertStash imports a shared stash archive into the global database.
func convertStash(ctx context.Context, cfg savesConfig, name string) (dbEntry, *string) {
	target := globalDBPath(cfg.out)
	mode := persist.GameMode{Spawn: name == stashSaveNames[1]}

	store, err := persist.OpenSQLiteStore(ctx, target)
	if err != nil {
		return dbEntry{}, failureText(name, err)
	}

	defer func() { _ = store.Close() }()

	if err := persist.ImportStash(ctx, store, filepath.Join(cfg.dir, name), mode, persist.GlobalStashSlot); err != nil {
		return dbEntry{}, failureText(name, err)
	}

	entry, err := record(target)
	if err != nil {
		return dbEntry{}, failureText(name, err)
	}

	return entry, nil
}

// failureText turns one rejection into the report string form.
func failureText(name string, err error) *string {
	text := fmt.Sprintf("%s: %v", name, err)

	return &text
}

// modeForPrefix maps a GetSavePath prefix back to its save mode (pfile.cpp:81-88):
// persist's copy is unexported, and internal/persist is outside this lane's
// ownership, so this is the 1:1 duplicate.
func modeForPrefix(prefix string) persist.GameMode {
	switch prefix {
	case "share_":
		return persist.GameMode{Spawn: true, Multiplayer: true}
	case "spawn_":
		return persist.GameMode{Spawn: true}
	case "multi_":
		return persist.GameMode{Multiplayer: true}
	default:
		return persist.GameMode{}
	}
}
