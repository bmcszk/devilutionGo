package persist_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// archiveSeed is the level seed the roundtrip freezes: a value no zeroed blob
// could carry, so the tail read is discriminating.
const archiveSeed = uint32(0x00C0FFEE)

// archiveWarrior is the pack the archive tests raise: the scripted warrior's
// frozen starting pack (internal/integration/session_stage_test.go).
var archiveWarrior = world.StartingPack{
	Class: 0, BaseStr: 30, BaseMag: 10, BaseDex: 20, BaseVit: 25,
	HPBase: 96, MaxHPBase: 128, ManaBase: 8, MaxManaBase: 8,
}

// archivePacks is the test PackSource the archive serves packs through: the
// warrior class and no other, mirroring the boot source the load path uses.
type archivePacks struct{}

// StartingPack implements world.PackSource.
func (archivePacks) StartingPack(class int) (world.StartingPack, bool) {
	if class != 0 {
		return world.StartingPack{}, false
	}

	return archiveWarrior, true
}

// TestSessionArchiveRoundtripMatchesRecord drives the production archive: a
// saved record survives save → close → reopen → load field for field, with
// the hero row's own name and creation time intact (the create flow wrote
// the row; the archive must not clobber them).
func TestSessionArchiveRoundtripMatchesRecord(t *testing.T) {
	// given a set with the hero row the create flow writes and the
	// archive over it
	dir := t.TempDir()
	set := newRoundtripSet(t, dir)

	archive := persist.NewSessionArchive(set, archivePacks{})

	// when the record freezes into slot 3
	rec := world.SessionRecord{
		Slot: 3, LevelSeed: archiveSeed,
		Level: 0, LevelType: 0, Loaded: false,
		Hero: world.Tile{X: 75, Y: 68},
		Pack: archiveWarrior,
	}

	if err := archive.SaveSlot(rec); err != nil {
		t.Fatalf("SaveSlot: %v", err)
	}

	row, err := set.LoadHero(context.Background(), 3)
	if err != nil {
		t.Fatalf("LoadHero: %v", err)
	}

	// and the slot loads back (nothing stays open — the set opens per call)
	got, err := persist.NewSessionArchive(set, archivePacks{}).LoadSlot(3)
	if err != nil {
		t.Fatalf("LoadSlot: %v", err)
	}

	// then the record round-trips field for field
	want := rec
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded record = %+v, want the frozen %+v", got, want)
	}

	// and the hero row kept the name the create flow wrote
	if row.Name != "Round Hero" {
		t.Fatalf("hero name after SaveSlot = %q, want Round Hero (LoadGame keeps the row's own name)", row.Name)
	}
}

// TestSessionArchiveEmptySlotIsErrNoSlot pins the empty-slot error: a slot
// with no hero row loads as world.ErrNoSlot, not a store failure.
func TestSessionArchiveEmptySlotIsErrNoSlot(t *testing.T) {
	archive := persist.NewSessionArchive(newArchiveSet(t), archivePacks{})

	// when the empty slot is read back
	_, err := archive.LoadSlot(7)

	// then the error is the empty-slot sentinel
	if err == nil {
		t.Fatal("LoadSlot(7) = nil error, want world.ErrNoSlot")
	}

	if !errors.Is(err, world.ErrNoSlot) {
		t.Fatalf("LoadSlot(7) error = %v, want errors.Is(err, world.ErrNoSlot)", err)
	}
}

// TestSessionArchiveHasGameFollowsTheGameRow pins the existence check: a slot
// with a frozen session reports HasGame, a slot without does not.
func TestSessionArchiveHasGameFollowsTheGameRow(t *testing.T) {
	set := newArchiveSet(t)
	archive := persist.NewSessionArchive(set, archivePacks{})

	ctx := context.Background()

	hero := persist.HeroRecord{
		Slot: 0, Name: "Has Game", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}

	if err := set.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	has, err := archive.HasGame(ctx, 0, string(persist.KindSingle))
	if err != nil {
		t.Fatalf("HasGame before save: %v", err)
	}

	if has {
		t.Fatal("HasGame(slot 0) = true before any SaveSlot, want false")
	}

	// when the session freezes into the slot
	frozen := world.SessionRecord{
		Slot: 0, LevelSeed: archiveSeed, Hero: world.Tile{X: 75, Y: 68}, Pack: archiveWarrior,
	}

	if err := archive.SaveSlot(frozen); err != nil {
		t.Fatalf("SaveSlot: %v", err)
	}

	// then the game row exists
	has, err = archive.HasGame(ctx, 0, string(persist.KindSingle))
	if err != nil {
		t.Fatalf("HasGame after save: %v", err)
	}

	if !has {
		t.Fatal("HasGame(slot 0) = false after SaveSlot, want true")
	}
}

// TestSessionArchiveRoutesToTheSlotsOwnFile pins the per-hero routing law: a
// record saved into slot 3 lands in single_3.db — and NEVER in single_0.db —
// because the archive opens the slot's own file per operation
// (pfile.cpp:609-612 OpenSaveArchive @ca91d778f).
func TestSessionArchiveRoutesToTheSlotsOwnFile(t *testing.T) {
	dir := t.TempDir()
	set := persist.NewHeroStoreSet(dir, heroSetNS)

	ctx := context.Background()

	// given the create flow's hero row on slot 3, written through the set
	hero := persist.HeroRecord{
		Slot: 3, Name: "Slot Three", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}

	if err := set.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero slot 3: %v", err)
	}

	archive := persist.NewSessionArchive(set, archivePacks{})

	// when the record freezes into slot 3
	rec := world.SessionRecord{
		Slot: 3, LevelSeed: archiveSeed, Level: 0, LevelType: 0, Loaded: false,
		Hero: world.Tile{X: 75, Y: 68},
		Pack: archiveWarrior,
	}

	if err := archive.SaveSlot(rec); err != nil {
		t.Fatalf("SaveSlot: %v", err)
	}

	// then the slot's own file holds the frozen session
	if _, err := os.Stat(filepath.Join(dir, "single_3.db")); err != nil {
		t.Fatalf("single_3.db missing after SaveSlot: %v", err)
	}

	// and slot 0's file was never created
	if _, err := os.Stat(filepath.Join(dir, "single_0.db")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("single_0.db stat = %v, want not-exist (the save must not leak into slot 0)", err)
	}
}

// newRoundtripSet builds the roundtrip's per-hero store set in dir and writes
// the create flow's hero row — the row the save must keep the name of.
func newRoundtripSet(t *testing.T, dir string) *persist.HeroStoreSet {
	t.Helper()

	set := persist.NewHeroStoreSet(dir, persist.SaveNamespace{})

	hero := persist.HeroRecord{
		Slot: 3, Name: "Round Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}

	if err := set.SaveHero(context.Background(), hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	return set
}

// newArchiveSet builds the per-hero store set the archive tests run against.
func newArchiveSet(t *testing.T) *persist.HeroStoreSet {
	t.Helper()

	return persist.NewHeroStoreSet(t.TempDir(), persist.SaveNamespace{})
}
