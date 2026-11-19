package world_test

// Pins for the session verbs (devilGo-sj5a): descend loads a level image and
// announces it, loot takes the floor item under the player. Both own their
// state through the handler only (ADR-018 laws 2 and 3).

import (
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// The cathedral level and dungeon_type the fixtures use (gendung.cpp:326-344:
// level 1 is DTYPE_CATHEDRAL, ordinal 1 in gendung_defs.hpp:15-26).
const (
	cathLevel = 1
	cathType  = 1
)

// spawnTile is the tile the scripted session's first level entry lands on.
var spawnTile = world.Tile{X: 40, Y: 40}

// warriorPack is the scripted warrior's starting pack.
var warriorPack = world.StartingPack{
	Class: 0, BaseStr: 30, BaseMag: 10, BaseDex: 20, BaseVit: 25,
	HPBase: 96, MaxHPBase: 128, ManaBase: 8, MaxManaBase: 8,
}

// scriptedLevels is the LevelSource a test scripts: one level image per level
// number, plus the entry tile the load starts from.
var errArchiveRefused = errors.New("scripted archive refused the write")

type scriptedLevels struct {
	images map[int]*world.LevelLayout
	spawn  world.Tile
}

// scriptedPacks is the PackSource a test scripts: one starting pack per class
// ordinal.
type scriptedPacks struct {
	packs map[int]world.StartingPack
}

// scriptedArchive is the SessionArchive a test scripts: one record per slot,
// plus the errors it is told to return, so the seam's own contract is provable
// without a database.
type scriptedArchive struct {
	records map[int]world.SessionRecord
	saveErr error
}

func (a *scriptedArchive) SaveSlot(rec world.SessionRecord) error {
	if a.saveErr != nil {
		return a.saveErr
	}

	if a.records == nil {
		a.records = make(map[int]world.SessionRecord)
	}

	a.records[rec.Slot] = rec

	return nil
}

func (a *scriptedArchive) LoadSlot(slot int) (world.SessionRecord, error) {
	rec, ok := a.records[slot]
	if !ok {
		return world.SessionRecord{}, world.ErrNoSlot
	}

	return rec, nil
}

func (s scriptedLevels) LevelImage(level, _ int) (*world.LevelLayout, world.Tile, bool) {
	image, ok := s.images[level]

	return image, s.spawn, ok
}

func (s scriptedPacks) StartingPack(class int) (world.StartingPack, bool) {
	pack, ok := s.packs[class]

	return pack, ok
}

func TestDescend_LoadsTheLevelAndAnnouncesIt(t *testing.T) {
	// given
	state := world.NewState()
	source := scriptedLevels{
		images: map[int]*world.LevelLayout{cathLevel: openLayout()},
		spawn:  world.Tile{X: 40, Y: 40},
	}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, source)

	// when
	accepted := k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	events := k.Pass(0)

	// then the gate passed it, one event announced it, and the slice moved
	if !accepted {
		t.Fatal("gate dropped the first descend, want it dispatched")
	}

	assertDescendedEvent(t, events, cathLevel, cathType)
	assertLevelState(t, state, source.spawn, cathLevel, cathType)
}

func TestDescend_FoldsTheLayoutIntoTheGrid(t *testing.T) {
	// given
	state := world.NewState()
	source := scriptedLevels{images: map[int]*world.LevelLayout{cathLevel: solidLayout(world.Tile{X: 10, Y: 10})}}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, source)

	// when
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	k.Pass(0)

	// then the piece layer and its SOL classification both landed
	if !state.Grid().IsTileSolid(10, 10) {
		t.Fatal("descend did not load the level's solid tile, want the layout folded into the grid")
	}

	if !state.Grid().IsTileWalkable(11, 11) {
		t.Fatal("descend marked a walkable tile solid, want the layout's own classification")
	}
}

func TestDescend_UnknownLevelChangesNothing(t *testing.T) {
	// given
	state := world.NewState()
	source := scriptedLevels{images: map[int]*world.LevelLayout{}}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, source)

	// when
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	events := k.Pass(0)

	// then
	if len(events) != 0 {
		t.Fatalf("unloadable level emitted %d events, want 0 (no change, no event)", len(events))
	}

	if _, _, ok := state.Level(); ok {
		t.Fatal("unloadable level recorded a loaded level, want the session untouched")
	}
}

func TestDescend_GateDropsAReloadOfTheSameLevel(t *testing.T) {
	// given
	state := world.NewState()
	source := scriptedLevels{images: map[int]*world.LevelLayout{cathLevel: openLayout()}}
	gate := world.NewSessionGate(state, state, 0)

	k := kernel.New(gate, nil)
	world.RegisterSessionVerbs(k, state, source)

	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	k.Pass(0)

	// when the same level is asked for again
	rule := gate.Reject(kernel.Command{Tick: 1, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	accepted := k.Submit(kernel.Command{Tick: 1, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	events := k.Pass(1)

	// then
	if rule != world.RuleSameLevel {
		t.Fatalf("reload judged %q, want %q", rule, world.RuleSameLevel)
	}

	if accepted || len(events) != 0 {
		t.Fatalf("reload accepted=%v events=%d, want the gate to drop it before dispatch", accepted, len(events))
	}
}

func TestLoot_TakesTheFloorItemAndAnnouncesIt(t *testing.T) {
	// given a session standing on a level whose entry tile holds an item
	const itemIndex = 7

	tile := world.Tile{X: 20, Y: 21}
	layout := openLayout()
	layout.Items = []world.FloorItem{{Index: itemIndex, Tile: tile}}

	state := world.NewState()

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, scriptedLevels{
		images: map[int]*world.LevelLayout{cathLevel: layout}, spawn: tile,
	})

	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	k.Pass(0)

	// when
	k.Submit(kernel.Command{Tick: 1, Kind: kernel.KindLoot})
	events := k.Pass(1)

	// then
	if len(events) != 1 || events[0].Kind != kernel.KindLooted {
		t.Fatalf("events = %v, want one %q", events, kernel.KindLooted)
	}

	if events[0].A != itemIndex || events[0].B != int32(tile.X) || events[0].C != int32(tile.Y) {
		t.Fatalf("event payload = (%d,%d,%d), want (%d,%d,%d)",
			events[0].A, events[0].B, events[0].C, itemIndex, tile.X, tile.Y)
	}

	if got := state.Grid().DItem[tile.Y][tile.X]; got != 0 {
		t.Fatalf("dItem at the looted tile = %d, want 0 (the item left the floor)", got)
	}
}

func TestLoot_EmptyTileChangesNothing(t *testing.T) {
	// given
	state := world.NewState()

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, scriptedLevels{images: map[int]*world.LevelLayout{cathLevel: openLayout()}})

	// when
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindLoot})
	events := k.Pass(0)

	// then
	if len(events) != 0 {
		t.Fatalf("loot on an empty tile emitted %d events, want 0 (no change, no event)", len(events))
	}
}

func TestCreatePack_RaisesTheHeroAndAnnouncesIt(t *testing.T) {
	// given
	const class = 0

	tile := world.Tile{X: 40, Y: 41}
	state := world.NewState()

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterCreatePack(k, state, scriptedPacks{packs: map[int]world.StartingPack{class: warriorPack}})

	// when
	command := kernel.Command{Tick: 0, Kind: kernel.KindCreatePack, A: class, B: int32(tile.X), C: int32(tile.Y)}
	accepted := k.Submit(command)
	events := k.Pass(0)

	// then the gate passed it, one event announced it, and the slice moved
	if !accepted {
		t.Fatal("gate dropped the create, want it dispatched")
	}

	if len(events) != 1 || events[0].Kind != kernel.KindCreatedPack {
		t.Fatalf("events = %v, want one %q", events, kernel.KindCreatedPack)
	}

	if events[0].A != class || events[0].B != int32(tile.X) || events[0].C != int32(tile.Y) {
		t.Fatalf("event payload = (%d,%d,%d), want (%d,%d,%d)",
			events[0].A, events[0].B, events[0].C, class, tile.X, tile.Y)
	}

	raised, hero, ok := state.RaisedPack()
	if !ok || raised != warriorPack || hero != tile {
		t.Fatalf("raised pack = (%+v,%v,%v), want (%+v,%v,true)", raised, hero, ok, warriorPack, tile)
	}
}

func TestCreatePack_UnknownClassChangesNothing(t *testing.T) {
	// given a source with no pack for the class the command names
	state := world.NewState()

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterCreatePack(k, state, scriptedPacks{packs: map[int]world.StartingPack{}})

	// when
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindCreatePack, A: 0})
	events := k.Pass(0)

	// then
	if len(events) != 0 {
		t.Fatalf("a class with no pack emitted %d events, want 0 (no change, no event)", len(events))
	}

	if _, _, ok := state.RaisedPack(); ok {
		t.Fatal("a class with no pack raised a hero, want the session untouched")
	}
}

func TestSaveSlot_FreezesTheSessionAndAnnouncesIt(t *testing.T) {
	// given a session that has loaded a level and raised a hero on it
	const (
		slot = 3
		seed = 2588
	)

	state, source := loadedSession(t)

	// when
	archive := &scriptedArchive{}

	k := kernel.New(world.NewSessionGate(state, state, 0), nil)
	world.RegisterSaveLoad(k, state, archive, source)

	accepted := k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindSaveSlot, A: slot, B: seed})
	events := k.Pass(0)

	// then the gate passed it, one event announced it and the record froze the session
	if !accepted {
		t.Fatal("gate dropped the save, want it dispatched")
	}

	if len(events) != 1 || events[0].Kind != kernel.KindSavedSlot {
		t.Fatalf("events = %v, want one %q", events, kernel.KindSavedSlot)
	}

	if events[0].A != slot || events[0].B != cathLevel || events[0].C != seed {
		t.Fatalf("event payload = (%d,%d,%d), want (%d,%d,%d)",
			events[0].A, events[0].B, events[0].C, slot, cathLevel, seed)
	}

	assertFrozenRecord(t, archive.records[slot], slot, seed, state.Hero())
}

func TestSaveSlot_NoHeroChangesNothing(t *testing.T) {
	// given a session that never raised a hero
	state := world.NewState()

	// when
	archive := &scriptedArchive{}

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSaveLoad(k, state, archive, scriptedLevels{})

	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindSaveSlot, A: 0, B: 2588})
	events := k.Pass(0)

	// then the handler stayed silent and wrote nothing
	if len(events) != 0 {
		t.Fatalf("a save with no hero emitted %d events, want 0 (no change, no event)", len(events))
	}

	if len(archive.records) != 0 {
		t.Fatalf("a save with no hero wrote %d records, want 0", len(archive.records))
	}
}

func TestSaveSlot_WriteFailureChangesNothing(t *testing.T) {
	// given an archive that refuses the write
	state, source := loadedSession(t)
	archive := &scriptedArchive{saveErr: errArchiveRefused}

	// when
	k := kernel.New(world.NewSessionGate(state, state, 0), nil)
	world.RegisterSaveLoad(k, state, archive, source)

	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindSaveSlot, A: 0, B: 2588})
	events := k.Pass(0)

	// then the failed write announced nothing
	if len(events) != 0 {
		t.Fatalf("a refused save emitted %d events, want 0 (no change, no event)", len(events))
	}
}

func TestLoadSlot_RestoresTheSavedSessionAndAnnouncesIt(t *testing.T) {
	// given a session saved into a slot
	const (
		slot = 3
		seed = 2588
	)

	saved, source := loadedSession(t)
	archive := &scriptedArchive{}

	k := kernel.New(world.NewSessionGate(saved, saved, 0), nil)
	world.RegisterSaveLoad(k, saved, archive, source)
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindSaveSlot, A: slot, B: seed})
	k.Pass(0)

	frozen := saved.Grid().LayoutDigest()
	hero := saved.Hero()

	// when the same slot loads into a fresh session that walked away from it
	fresh := world.NewState()
	freshKernel := kernel.New(world.NewSessionGate(fresh, fresh, 0), nil)
	world.RegisterSaveLoad(freshKernel, fresh, archive, source)

	freshKernel.Submit(kernel.Command{Tick: 0, Kind: kernel.KindLoadSlot, A: slot, B: seed})
	events := freshKernel.Pass(0)

	// then one event announced the restore and the session equals the saved one
	if len(events) != 1 || events[0].Kind != kernel.KindLoadedSlot {
		t.Fatalf("events = %v, want one %q", events, kernel.KindLoadedSlot)
	}

	wantTileWord := int32(cathType)<<16 | int32(hero.Y*world.MAXDUNX+hero.X)
	if events[0].A != slot || events[0].B != cathLevel || events[0].C != wantTileWord {
		t.Fatalf("event payload = (%d,%d,%d), want (%d,%d,%d)",
			events[0].A, events[0].B, events[0].C, slot, cathLevel, wantTileWord)
	}

	assertRestoredSession(t, fresh, hero, frozen)
}

func TestLoadSlot_EmptySlotChangesNothing(t *testing.T) {
	// given a fresh session and an archive holding nothing
	state := world.NewState()

	// when
	k := kernel.New(world.NewSessionGate(state, state, 0), nil)
	world.RegisterSaveLoad(k, state, &scriptedArchive{}, scriptedLevels{})

	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindLoadSlot, A: 0, B: 2588})
	events := k.Pass(0)

	// then
	if len(events) != 0 {
		t.Fatalf("a load from an empty slot emitted %d events, want 0 (no change, no event)", len(events))
	}

	if _, _, ok := state.RaisedPack(); ok {
		t.Fatal("a load from an empty slot raised a hero, want the session untouched")
	}
}

func TestLoadSlot_ForeignLevelSeedChangesNothing(t *testing.T) {
	// given a slot written from one level seed
	saved, source := loadedSession(t)
	archive := &scriptedArchive{}

	k := kernel.New(world.NewSessionGate(saved, saved, 0), nil)
	world.RegisterSaveLoad(k, saved, archive, source)
	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindSaveSlot, A: 0, B: 2588})
	k.Pass(0)

	// when the load asks for it at another seed
	fresh := world.NewState()
	freshKernel := kernel.New(world.NewSessionGate(fresh, fresh, 0), nil)
	world.RegisterSaveLoad(freshKernel, fresh, archive, source)

	freshKernel.Submit(kernel.Command{Tick: 0, Kind: kernel.KindLoadSlot, A: 0, B: 4000})
	events := freshKernel.Pass(0)

	// then the slot names a different dungeon, so nothing was restored
	if len(events) != 0 {
		t.Fatalf("a load at a foreign seed emitted %d events, want 0 (no change, no event)", len(events))
	}

	if _, _, ok := fresh.Level(); ok {
		t.Fatal("a load at a foreign seed restored a level, want the session untouched")
	}
}

// assertRestoredSession pins that the loaded slice equals the saved one: the
// same layout digest, the same hero tile, the same level identity and the same
// raised pack.
func assertRestoredSession(t *testing.T, fresh *world.State, hero world.Tile, layoutDigest uint64) {
	t.Helper()

	if got := fresh.Grid().LayoutDigest(); got != layoutDigest {
		t.Fatalf("restored layout digest = %x, want the saved %x", got, layoutDigest)
	}

	if fresh.Hero() != hero {
		t.Fatalf("restored hero at %v, want the saved %v", fresh.Hero(), hero)
	}

	if level, kind, ok := fresh.Level(); !ok || level != cathLevel || kind != cathType {
		t.Fatalf("restored level state = (%d,%d,%v), want (%d,%d,true)", level, kind, ok, cathLevel, cathType)
	}

	if pack, tile, ok := fresh.RaisedPack(); !ok || pack != warriorPack || tile != hero {
		t.Fatalf("restored pack = (%+v,%v,%v), want (%+v,%v,true)", pack, tile, ok, warriorPack, hero)
	}
}

// openLayout is a level image whose whole 112x112 grid is walkable.
func openLayout() *world.LevelLayout {
	layout := &world.LevelLayout{
		Piece: make([]int, world.MAXDUNX*world.MAXDUNY),
		SOL:   make([]byte, 1),
	}

	return layout
}

// solidLayout is openLayout with one tile made solid.
func solidLayout(tile world.Tile) *world.LevelLayout {
	layout := openLayout()
	layout.Piece[tile.X*world.MAXDUNY+tile.Y] = 1
	layout.SOL = []byte{0, 0x01}

	return layout
}

// assertDescendedEvent pins the announce event's kind and payload.
func assertDescendedEvent(t *testing.T, events []kernel.Event, level, levelType int) {
	t.Helper()

	if len(events) != 1 || events[0].Kind != kernel.KindDescended {
		t.Fatalf("events = %v, want one %q", events, kernel.KindDescended)
	}

	if events[0].A != int32(level) || events[0].B != int32(levelType) {
		t.Fatalf("event payload = (%d,%d), want (%d,%d)", events[0].A, events[0].B, level, levelType)
	}
}

// assertLevelState pins where the load put the player and what it recorded.
func assertLevelState(t *testing.T, state *world.State, hero world.Tile, level, levelType int) {
	t.Helper()

	if got := state.Hero(); got != hero {
		t.Fatalf("hero at %v after the descend, want the level's entry %v", got, hero)
	}

	if gotLevel, kind, ok := state.Level(); !ok || gotLevel != level || kind != levelType {
		t.Fatalf("level state = (%d,%d,%v), want (%d,%d,true)", gotLevel, kind, ok, level, levelType)
	}
}

// loadedSession builds a session standing on the cathedral level with the
// warrior raised, plus the level source that served it.
func loadedSession(t *testing.T) (*world.State, scriptedLevels) {
	t.Helper()

	source := scriptedLevels{images: map[int]*world.LevelLayout{cathLevel: openLayout()}, spawn: spawnTile}
	state := world.NewState()

	k := kernel.New(world.NewGate(state, 0), nil)
	world.RegisterSessionVerbs(k, state, source)
	world.RegisterCreatePack(k, state, scriptedPacks{packs: map[int]world.StartingPack{warriorPack.Class: warriorPack}})

	k.Submit(kernel.Command{Tick: 0, Kind: kernel.KindDescend, A: cathLevel, B: cathType})
	k.Pass(0)
	k.Submit(
		kernel.Command{
			Tick: 1,
			Kind: kernel.KindCreatePack,
			A:    int32(warriorPack.Class),
			B:    int32(spawnTile.X),
			C:    int32(spawnTile.Y),
		},
	)
	k.Pass(1)

	return state, source
}

// assertFrozenRecord pins the record a save wrote.
func assertFrozenRecord(t *testing.T, rec world.SessionRecord, slot int, seed uint32, hero world.Tile) {
	t.Helper()

	if rec.Slot != slot || rec.LevelSeed != seed {
		t.Fatalf("record slot/seed = (%d,%d), want (%d,%d)", rec.Slot, rec.LevelSeed, slot, seed)
	}

	if rec.Level != cathLevel || rec.LevelType != cathType || !rec.Loaded {
		t.Fatalf("record level = (%d,%d,%v), want (%d,%d,true)",
			rec.Level, rec.LevelType, rec.Loaded, cathLevel, cathType)
	}

	if rec.Hero != hero || rec.Pack != warriorPack {
		t.Fatalf("record hero/pack = (%v,%+v), want (%v,%+v)", rec.Hero, rec.Pack, hero, warriorPack)
	}
}
