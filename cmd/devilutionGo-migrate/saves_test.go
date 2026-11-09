package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// savedHero is one synthetic save the fixtures write.
type savedHero struct {
	name    string
	file    string
	slot    int
	mode    persist.GameMode
	pack    persist.PlayerPack
	cuLevel uint32
	lvType  uint32
	hasGame bool
}

func TestSavesImportsLegacySavesFieldCorrect(t *testing.T) {
	// given a legacy dir with one save per family, one with a game stream
	dir := t.TempDir()
	writeLegacySaves(t, dir)
	out := filepath.Join(t.TempDir(), "out")

	// when the saves subcommand runs
	code, stdout, stderr := runSaves(t, dir, out)

	// then every save migrated with a checksum line and nothing failed
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "migrated 4 save(s)") {
		t.Errorf("stdout = %q, want four migrations", stdout)
	}

	// and every family reads back field-correct through the persist codecs
	assertRetailHero(t, out)
	assertMultiSave(t, out)
	assertGlobalStash(t, out)
}

// TestSavesImportsBothStashFamilies checks the saves subcommand keeps the
// stash family pair: stash.sv and stash_spawn.sv land in separate rows of the
// global database instead of overwriting each other (pfile.cpp:96-104).
func TestSavesImportsBothStashFamilies(t *testing.T) {
	// given a legacy dir holding both stash families with distinct content
	dir := t.TempDir()
	writeStashFamily(t, dir, "stash.sv", persist.GameMode{}, 77, 1)
	writeStashFamily(t, dir, "stash_spawn.sv", persist.GameMode{Spawn: true}, 777, 2)
	out := filepath.Join(t.TempDir(), "out")

	// when the saves subcommand runs
	code, stdout, stderr := runSaves(t, dir, out)

	// then both stashes migrated and nothing failed
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "migrated 2 save(s)") {
		t.Errorf("stdout = %q, want both stashes migrated", stdout)
	}

	// and each family reads back from its own slot
	ctx := context.Background()

	global, err := persist.OpenSQLiteStore(ctx, filepath.Join(out, defaultGlobalDBName))
	if err != nil {
		t.Fatalf("open global store: %v", err)
	}

	t.Cleanup(func() { _ = global.Close() })

	assertStashSlot(ctx, t, global, persist.GlobalStashSlot, 77, 1)
	assertStashSlot(ctx, t, global, persist.SpawnStashSlot, 777, 2)
}

func TestSavesRunSkipsConvertedSavesOnRerun(t *testing.T) {
	// given a legacy dir fully migrated once
	dir := t.TempDir()
	writeLegacySaves(t, dir)
	out := filepath.Join(t.TempDir(), "out")

	runSaves(t, dir, out)

	stores := []string{"single_0.db", "multi_5.db", "spawn_2.db"}
	first := make([][]byte, 0, len(stores))
	for _, name := range stores {
		first = append(first, importedBytes(t, filepath.Join(out, name)))
	}

	// when the run repeats against the untouched directory
	code, stdout, stderr := runSaves(t, dir, out)

	// then nothing was rewritten and everything came from the journal
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "migrated 0 save(s), skipped 4 already-migrated") {
		t.Errorf("stdout = %q, want a full skip", stdout)
	}

	for i, name := range stores {
		if got := importedBytes(t, filepath.Join(out, name)); !bytes.Equal(got, first[i]) {
			t.Errorf("%s was rewritten on resume", name)
		}
	}
}

func TestSavesRunReimportsChangedSave(t *testing.T) {
	// given a legacy dir fully migrated once; one save replaced on disk afterwards
	dir := t.TempDir()
	writeLegacySaves(t, dir)
	out := filepath.Join(t.TempDir(), "out")

	runSaves(t, dir, out)

	turned := goldenSaves()[1]
	turned.name = "SecondHero"
	turned.pack = spPack("SecondHero", 3, 9)

	writeHeroSave(t, dir, turned)

	// when the run resumes
	code, stdout, stderr := runSaves(t, dir, out)

	// then only the changed save was reimported
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "migrated 1 save(s), skipped 3 already-migrated") {
		t.Errorf("stdout = %q, want one reimport", stdout)
	}

	// and the row carries the new hero
	_, pack := mustHero(t, out, 5)
	if want := strings.TrimRight(string(pack.Name[:]), "\x00"); want != "SecondHero" {
		t.Errorf("name = %q, want SecondHero", want)
	}
}

func TestSavesResumesInterruptedRun(t *testing.T) {
	// given a run that converted the first hero save and stopped midway
	dir := t.TempDir()
	writeLegacySaves(t, dir)
	out := filepath.Join(t.TempDir(), "out")

	interruptedJournal(t, dir, out, goldenSaves()[0])

	// when the interrupted run is resumed as the subcommand sees it
	code, stdout, stderr := runSaves(t, dir, out)

	// then the recorded save is skipped with its recorded checksum
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "migrated 3 save(s), skipped 1 already-migrated") {
		t.Errorf("stdout = %q, want one resume skip", stdout)
	}

	if !strings.Contains(stdout, "single_0.sv -> single_0.db") {
		t.Errorf("stdout = %q, want the resumed checksum line", stdout)
	}
}

func TestSavesDryRunPlansWithoutWriting(t *testing.T) {
	// given a legacy dir and an output dir that does not exist yet
	dir := t.TempDir()
	writeLegacySaves(t, dir)
	out := filepath.Join(t.TempDir(), "missing")

	// when a dry run is asked for
	code, stdout, stderr := runSaves(t, dir, out, "--dry-run")

	// then the plan lists every save and its destination without writing
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	for _, want := range []string{
		"planned 4 save(s)",
		"plan: single_0.sv -> single_0.db",
		"plan: multi_5.sv -> multi_5.db",
		"plan: spawn_2.sv -> spawn_2.db",
		"plan: stash.sv -> " + defaultGlobalDBName,
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout = %q, want %q", stdout, want)
		}
	}

	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("dry run created %s", out)
	}
}

func TestSavesReportsBadSaveAndContinues(t *testing.T) {
	// given a legacy dir holding one valid set and one corrupted save
	dir := t.TempDir()
	writeLegacySaves(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "share_2.sv"), []byte("not an mpq"), 0o644); err != nil {
		t.Fatalf("write bad save: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out")

	// when the subcommand runs
	code, stdout, stderr := runSaves(t, dir, out)

	// then the valid saves landed, the corrupt one is named and the run exits 1
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr, "share_2.sv") {
		t.Errorf("stderr = %q, want it to name the rejected save", stderr)
	}

	if !strings.Contains(stdout, "migrated 4 save(s)") {
		t.Errorf("stdout = %q, want the unaffected saves to land", stdout)
	}

	if _, err := os.Stat(filepath.Join(out, "share_2.db")); !errors.Is(err, os.ErrNotExist) {
		t.Error("corrupt save left a store file behind")
	}
}

func TestSavesReportsHowToRecoverFromATornJournal(t *testing.T) {
	// given a legacy dir plus an output dir whose journal is torn mid-write
	dir := t.TempDir()
	writeLegacySaves(t, dir)
	out := filepath.Join(t.TempDir(), "out")

	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}

	if err := os.WriteFile(filepath.Join(out, savesJournalName), []byte(`{"version":1,"dir":"`), 0o644); err != nil {
		t.Fatalf("write torn journal: %v", err)
	}

	// when the subcommand runs
	code, _, stderr := runSaves(t, dir, out)

	// then the parse failure carries the delete-to-start-over recovery the
	// version-mismatch branch already gives
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr, "parse journal") {
		t.Errorf("stderr = %q, want the parse failure", stderr)
	}

	if !strings.Contains(stderr, "delete "+filepath.Join(out, savesJournalName)+" to start over") {
		t.Errorf("stderr = %q, want the delete-to-recover hint", stderr)
	}
}

func TestSavesKeepsTheOldStoreWhenACorruptReimportFails(t *testing.T) {
	// given a legacy dir fully migrated once; one hero save corrupted afterwards
	dir := t.TempDir()
	writeLegacySaves(t, dir)
	out := filepath.Join(t.TempDir(), "out")

	runSaves(t, dir, out)

	good := importedBytes(t, filepath.Join(out, "single_0.db"))

	if err := os.WriteFile(filepath.Join(dir, "single_0.sv"), []byte("not an mpq"), 0o644); err != nil {
		t.Fatalf("corrupt save: %v", err)
	}

	// when the run reimports the changed save
	code, stdout, stderr := runSaves(t, dir, out)

	// then the failure is named, the run exits 1 and the previously migrated
	// store is still on disk unchanged
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr, "single_0.sv") {
		t.Errorf("stderr = %q, want it to name the rejected save", stderr)
	}

	if !strings.Contains(stdout, "migrated 0 save(s), skipped 3 already-migrated, 1 failure(s)") {
		t.Errorf("stdout = %q, want the rest of the run unaffected", stdout)
	}

	if now := importedBytes(t, filepath.Join(out, "single_0.db")); !bytes.Equal(now, good) {
		t.Error("the failed reimport deleted the previously migrated store")
	}
}

func TestSavesFailedStashImportLeavesNoResidueOnAFreshOutDir(t *testing.T) {
	// given a legacy dir whose stash archive is corrupt
	dir := t.TempDir()
	writeLegacySaves(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "stash.sv"), []byte("not an mpq"), 0o644); err != nil {
		t.Fatalf("corrupt stash: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out")

	// when the subcommand runs
	code, stdout, stderr := runSaves(t, dir, out)

	// then the failure is named, the run exits 1 and no empty global.db
	// residue is left behind on the fresh out dir
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr, "stash.sv") {
		t.Errorf("stderr = %q, want it to name the rejected stash archive", stderr)
	}

	if !strings.Contains(stdout, "migrated 3 save(s)") {
		t.Errorf("stdout = %q, want the hero saves to land", stdout)
	}

	if _, err := os.Stat(filepath.Join(out, defaultGlobalDBName)); !errors.Is(err, os.ErrNotExist) {
		t.Error("the failed stash import left a global.db residue behind")
	}
}

func TestSavesIgnoresForeignFiles(t *testing.T) {
	// given a legacy dir holding non-save files beside one save set
	dir := t.TempDir()
	writeLegacySaves(t, dir)

	keep := filepath.Join(dir, "readme.txt")
	if err := os.WriteFile(keep, []byte("not a save"), 0o644); err != nil {
		t.Fatalf("write foreign file: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out")

	// when the subcommand runs
	code, stdout, stderr := runSaves(t, dir, out)

	// then only the saves migrated and the foreign file exists untouched
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "migrated 4 save(s)") {
		t.Errorf("stdout = %q, want the saves only", stdout)
	}

	if body, err := os.ReadFile(keep); err != nil || string(body) != "not a save" {
		t.Errorf("foreign file was touched: %v", err)
	}
}

func TestSavesFailsOnMissingLegacyDir(t *testing.T) {
	// given a --dir that does not exist
	dir := filepath.Join(t.TempDir(), "no-such-saves")

	// when the subcommand runs
	code, _, stderr := runSaves(t, dir, t.TempDir())

	// then the missing directory is reported and the run exits 1
	if code != exitFailure {
		t.Fatalf("exit code = %d, want %d", code, exitFailure)
	}

	if !strings.Contains(stderr, "read save directory") {
		t.Errorf("stderr = %q, want the directory error", stderr)
	}
}

func TestSavesWithoutDirPrintsUsage(t *testing.T) {
	// when the subcommand runs without --dir
	code, _, stderr := runSaves(t, "", t.TempDir())

	// then it is a usage error that names the missing flag
	if code != exitUsage {
		t.Fatalf("exit code = %d, want %d", code, exitUsage)
	}

	if !strings.Contains(stderr, "--dir is required") {
		t.Errorf("stderr = %q, want the --dir requirement", stderr)
	}
}

func TestSavesHelpStatesTheCoveredFamilies(t *testing.T) {
	// when -h prints the saves help
	code, _, stderr := runSaves(t, t.TempDir(), t.TempDir(), "-h")

	// then the help names every imported family exactly and says .sp stays
	// untouched, so an operator with .sp files is not promised an import
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d", code, exitOK)
	}

	for _, want := range []string{
		"single_/spawn_/multi_/share_ .sv hero saves",
		"stash.sv",
		"stash_spawn.sv",
		".sp files are left untouched",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr = %q, want %q", stderr, want)
		}
	}
}

func TestSavesOutDefaultsToTheLegacyDir(t *testing.T) {
	// given a legacy dir with one save and no --out
	dir := t.TempDir()
	writeHeroSave(t, dir, goldenSaves()[0])

	// when the subcommand runs
	code, stdout, stderr := runSaves(t, dir, "")

	// then the store landed beside the legacy save
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}

	if !strings.Contains(stdout, "migrated 1 save(s)") {
		t.Errorf("stdout = %q, want one migration", stdout)
	}

	if _, err := os.Stat(filepath.Join(dir, "single_0.db")); err != nil {
		t.Errorf("store missing beside the legacy save: %v", err)
	}
}

// spPack builds a synthetic pack whose fields the read-back assertions use;
// they vary per save so the migration cannot pass two saves with one row.
func spPack(name string, class, level int) persist.PlayerPack {
	var pack persist.PlayerPack

	pack.Class = uint8(class)
	pack.Level = uint8(level)
	pack.BaseStr = 30
	pack.BaseMag = 10
	pack.BaseDex = 25
	pack.BaseVit = 20
	pack.Exp = uint32(1000*level + class)
	pack.Gold = 400
	pack.HPBase = 96
	pack.MaxHP = 128
	pack = devilItems(pack, class+level)

	copy(pack.Name[:], name)

	return pack
}

// devilItems lays one equipped weapon with a per-save identity into the body
// slot, the field the read-back test pivots on. Every other slot is empty.
func devilItems(pack persist.PlayerPack, seed int) persist.PlayerPack {
	for i := range pack.InvBody {
		pack.InvBody[i].Idx = persist.PackedItemNone
	}

	for i := range pack.InvList {
		pack.InvList[i].Idx = persist.PackedItemNone
	}

	for i := range pack.SpdList {
		pack.SpdList[i].Idx = persist.PackedItemNone
	}

	pack.InvBody[0] = persist.ItemPack{
		Seed:       uint32(seed),
		CreateInfo: uint16(seed),
		Idx:        3,
		BId:        4,
		BDur:       5,
		BMDur:      6,
		BCh:        7,
		BMCh:       8,
		Value:      9,
		Buff:       10,
	}

	return pack
}

// goldenSaves is the synthetic save directory the tests migrate: one save per
// save family, one with a game stream, one without.
func goldenSaves() []savedHero {
	return []savedHero{
		{
			name: "Conan", file: "single_0.sv", slot: 0, mode: persist.GameMode{}, pack: spPack("Conan", 0, 12),
		},
		{
			name: "SmokeHero", file: "multi_5.sv", slot: 5,
			mode:    persist.GameMode{Multiplayer: true},
			pack:    spPack("SmokeHero", 1, 7),
			cuLevel: 3, lvType: 2, hasGame: true,
		},
		{
			name: "ShareHero", file: "spawn_2.sv", slot: 2, mode: persist.GameMode{Spawn: true},
			pack: spPack("ShareHero", 2, 5),
		},
	}
}

// writeLegacySaves fills dir with the synthetic saves through the persist
// codec writers. Repo zero-dependency policy: no committed save binaries —
// MPQ containers are not byte-stable (upstream records the same,
// test/writehero_test.cpp:415-419).
func writeLegacySaves(t *testing.T, dir string) {
	t.Helper()

	for _, hero := range goldenSaves() {
		writeHeroSave(t, dir, hero)
	}

	writeLegacyStash(t, dir, "stash.sv", persist.GameMode{})
}

// writeHeroSave writes one synthetic hero save archive.
func writeHeroSave(t *testing.T, dir string, hero savedHero) {
	t.Helper()

	var game []byte

	if hero.hasGame {
		header, err := persist.GameHeader{
			CurrLevel: hero.cuLevel, LevelType: hero.lvType, ViewX: 30, ViewY: 25,
		}.Encode()
		if err != nil {
			t.Fatalf("encode game header: %v", err)
		}

		game = header
	}

	archive := persist.LegacySVStore{Path: filepath.Join(dir, hero.file), Mode: hero.mode}
	row := persist.HeroRecord{Slot: hero.slot, Name: hero.name, Pack: hero.pack.Encode()}

	if err := archive.Write(row, game); err != nil {
		t.Fatalf("write %s: %v", hero.file, err)
	}
}

// writeStashFamily writes one stash archive with distinct content, so a test
// can prove the two families did not overwrite each other.
func writeStashFamily(t *testing.T, dir, name string, mode persist.GameMode, gold, page uint32) {
	t.Helper()

	stash := persist.StashSave{
		Gold:        gold,
		Pages:       []persist.StashPage{{Index: 0}},
		Items:       [][]byte{make([]byte, persist.ItemSizeFor(mode))},
		CurrentPage: page,
	}

	archive := persist.LegacySVStore{Path: filepath.Join(dir, name), Mode: mode}
	if err := archive.WriteStash(stash, persist.ItemSizeFor(mode)); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// assertStashSlot reads one stash family row and checks its content.
func assertStashSlot(
	ctx context.Context, t *testing.T, global *persist.SQLiteStore, slot int, wantGold, wantPage uint32,
) {
	t.Helper()

	record, err := global.LoadStash(ctx, slot)
	if err != nil {
		t.Fatalf("load stash slot %d: %v", slot, err)
	}

	// Retail item size for both families: ItemSizeFor branches on Hellfire
	// only, so spawn shares Diablo's size.
	stash, err := persist.DecodeStashSave(record.Blob, persist.ItemSizeFor(persist.GameMode{}))
	if err != nil {
		t.Fatalf("decode stash slot %d: %v", slot, err)
	}

	if stash.Gold != wantGold || stash.CurrentPage != wantPage {
		t.Errorf("stash slot %d = gold %d page %d, want gold %d page %d",
			slot, stash.Gold, stash.CurrentPage, wantGold, wantPage)
	}
}

// writeLegacyStash writes one shared stash archive.
func writeLegacyStash(t *testing.T, dir, name string, mode persist.GameMode) {
	t.Helper()

	archive := persist.LegacySVStore{Path: filepath.Join(dir, name), Mode: mode}

	if err := archive.WriteStash(goldenStash(mode), persist.ItemSizeFor(mode)); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// goldenStash is the synthetic shared stash fixture.
func goldenStash(mode persist.GameMode) persist.StashSave {
	return persist.StashSave{
		Gold:        77,
		Pages:       []persist.StashPage{{Index: 0}},
		Items:       [][]byte{make([]byte, persist.ItemSizeFor(mode))},
		CurrentPage: 1,
	}
}

// runSaves runs the saves subcommand against one fixture dir.
func runSaves(t *testing.T, dir, out string, extra ...string) (code int, stdout, stderr string) {
	t.Helper()

	var outs, errs bytes.Buffer

	args := append([]string{"saves", "--dir", dir, "--out", out}, extra...)
	code = run(args, &outs, &errs)

	return code, outs.String(), errs.String()
}

// importedBytes returns an imported store file's content for the no-rewrite
// assertions.
func importedBytes(t *testing.T, path string) []byte {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return body
}

// mustHero decodes the pack of the hero row a migration produced.
func mustHero(t *testing.T, dir string, slot int) (persist.HeroRecord, persist.PlayerPack) {
	t.Helper()

	_, hero := openHeroStore(t, filepath.Join(dir, heroStoreName(slot)), slot)

	pack, err := persist.DecodePlayerPack(hero.Pack)
	if err != nil {
		t.Fatalf("decode pack: %v", err)
	}

	return hero, pack
}

// openHeroStore opens one produced store and reads the hero row; the store is
// closed when the test ends.
func openHeroStore(t *testing.T, path string, slot int) (*persist.SQLiteStore, persist.HeroRecord) {
	t.Helper()

	ctx := context.Background()

	store, err := persist.OpenSQLiteStore(ctx, path)
	if err != nil {
		t.Fatalf("open store %s: %v", path, err)
	}

	t.Cleanup(func() { _ = store.Close() })

	hero, err := store.LoadHero(ctx, slot)
	if err != nil {
		t.Fatalf("load hero slot %d: %v", slot, err)
	}

	return store, hero
}

// heroStoreName is the per-hero live file a legacy hero save becomes.
func heroStoreName(slot int) string {
	for _, hero := range goldenSaves() {
		if hero.slot == slot {
			return strings.TrimSuffix(hero.file, ".sv") + ".db"
		}
	}

	return ""
}

// interruptedJournal records the checkpoint of a run that converted exactly
// one hero save and then stopped: store file, journal entry, journal file.
func interruptedJournal(t *testing.T, dir, out string, hero savedHero) {
	t.Helper()

	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}

	target := filepath.Join(out, strings.TrimSuffix(hero.file, ".sv")+".db")

	ctx := context.Background()

	store, err := persist.OpenSQLiteStore(ctx, target)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	if err := persist.ImportSV(ctx, store, filepath.Join(dir, hero.file), hero.mode, hero.slot); err != nil {
		t.Fatalf("import: %v", err)
	}

	_ = store.Close()

	entry, err := record(target)
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	state, err := loadSavesJournal(out, dir)
	if err != nil {
		t.Fatalf("load journal: %v", err)
	}

	if err := state.recordSave(dir, hero.file, entry); err != nil {
		t.Fatalf("record save: %v", err)
	}

	if err := saveSavesJournal(out, state); err != nil {
		t.Fatalf("save journal: %v", err)
	}
}

// assertRetailHero reads back the single-player save through the codecs: pack
// fields, the equipped weapon, the row flags, and no game row beside it.
func assertRetailHero(t *testing.T, out string) {
	t.Helper()

	store, hero := openHeroStore(t, filepath.Join(out, heroStoreName(0)), 0)
	pack, err := persist.DecodePlayerPack(hero.Pack)
	if err != nil {
		t.Fatalf("decode pack: %v", err)
	}

	fields := []struct {
		label string
		got   any
		want  any
	}{
		{"name", strings.TrimRight(string(pack.Name[:]), "\x00"), "Conan"},
		{"class", pack.Class, uint8(0)},
		{"row class", hero.ClassID, 0},
		{"level", pack.Level, uint8(12)},
		{"exp", pack.Exp, uint32(12000)},
		{"gold", pack.Gold, int32(400)},
		{"hp base", pack.HPBase, int32(96)},
		{"max hp", pack.MaxHP, int32(128)},
		{"weapon idx", pack.InvBody[0].Idx, uint16(3)},
		{"weapon seed", pack.InvBody[0].Seed, uint32(12)},
		{"row name", hero.Name, "Conan"},
		{"row slot", hero.Slot, 0},
		{"multiplayer row", hero.Multiplayer, false},
		{"hellfire row", hero.Hellfire, false},
	}

	for _, field := range fields {
		if !reflect.DeepEqual(field.got, field.want) {
			t.Errorf("%s = %v, want %v", field.label, field.got, field.want)
		}
	}

	ctx := context.Background()

	if _, err := store.LoadGame(ctx, 0, string(persist.KindSingle)); !errors.Is(err, persist.ErrHeroNotFound) {
		t.Errorf("load game without stream = %v, want ErrHeroNotFound", err)
	}
}

// assertMultiSave reads back the multiplayer save: row flags, the spawn row
// class, and its game stream with the right level data.
func assertMultiSave(t *testing.T, out string) {
	t.Helper()

	store, hero := openHeroStore(t, filepath.Join(out, heroStoreName(5)), 5)
	if !hero.Multiplayer {
		t.Errorf("multi hero row flags = %+v, want multiplayer", hero)
	}

	// the spawn save is the class-2 fixture; its row must carry the class
	spawnStore, spawnHero := openHeroStore(t, filepath.Join(out, heroStoreName(2)), 2)
	if spawnHero.ClassID != 2 {
		t.Errorf("spawn row ClassID = %d, want 2", spawnHero.ClassID)
	}

	if err := spawnStore.Close(); err != nil {
		t.Errorf("close spawn store: %v", err)
	}

	ctx := context.Background()

	game, err := store.LoadGame(ctx, 5, string(persist.KindFor(persist.GameMode{Multiplayer: true})))
	if err != nil {
		t.Fatalf("load game: %v", err)
	}

	header, magic, err := persist.DecodeGameHeader(game)
	if err != nil {
		t.Fatalf("decode game header: %v", err)
	}

	if header.CurrLevel != 3 || header.LevelType != 2 || magic != persist.MagicRetail {
		t.Errorf("game = level %d type %d magic %s, want level 3 type 2", header.CurrLevel, header.LevelType, magic)
	}
}

// assertGlobalStash reads back the shared stash from the global database.
func assertGlobalStash(t *testing.T, out string) {
	t.Helper()

	ctx := context.Background()

	global, err := persist.OpenSQLiteStore(ctx, filepath.Join(out, defaultGlobalDBName))
	if err != nil {
		t.Fatalf("open global store: %v", err)
	}

	t.Cleanup(func() { _ = global.Close() })

	record, err := global.LoadStash(ctx, persist.GlobalStashSlot)
	if err != nil {
		t.Fatalf("load stash: %v", err)
	}

	stash, err := persist.DecodeStashSave(record.Blob, persist.ItemSizeFor(persist.GameMode{}))
	if err != nil {
		t.Fatalf("decode stash: %v", err)
	}

	if stash.Gold != 77 || stash.CurrentPage != 1 {
		t.Errorf("stash = gold %d page %d, want gold 77 page 1", stash.Gold, stash.CurrentPage)
	}
}
