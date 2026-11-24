// The M1 scripted-session stage (devilGo-ftbu AC-1): the harness that drives one
// headless walk — boot, create, town, dungeon, ten kills, loot, save, load —
// from the landed surfaces. The kernel, gate, streams, world slice, input
// bridge, combat pipeline and persist store all stay untouched: this file is
// glue that submits their own commands and reads their own state back.
package integration_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/combat"
	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	dtest "github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// The dungeon level and dungeon_type the walk descends to: cathedral level 1
// (gendung.cpp:326-344 puts levels 1-4 on DTYPE_CATHEDRAL).
const (
	dungeonLevel     = 1
	dungeonLevelType = 1
)

// The walk's nine stages, in walk order (devilGo-ftbu AC-1).
const (
	stepBoot         = "boot"
	stepCreate       = "create"
	stepTownEnter    = "town-enter"
	stepDungeonEnter = "dungeon-enter"
	stepKillFirst    = "kill#1"
	stepKillLast     = "kill#10"
	stepLoot         = "loot"
	stepSave         = "save"
	stepLoad         = "load"
)

const (
	// streamCombat is the RNG stream the combat handler draws from.
	streamCombat = "combat"
	// warriorClass is the class ordinal of the warrior row in classdat.tsv, the
	// word a CreatePack command carries.
	warriorClass = 0
	// monsterCount is how many monsters the dungeon holds and how many the walk
	// must kill through the kernel.
	monsterCount = 10
	// monsterHitPoints is every scripted monster's health in 1/64 fixed point.
	monsterHitPoints int64 = 1 << 6
	// attackAttemptLimit bounds one monster's kill loop; a run that cannot kill
	// in this many swings fails instead of hanging.
	attackAttemptLimit = 1000
	// heroNameText is the warrior name the session's pack carries.
	heroNameText = "M1Harness"
	// saveSlot is the hero slot the walk saves into and loads back; slot 0 is
	// the single-player slot the store keys the hero row by (pfile.h:24).
	saveSlot = 0
	// walkClicks is how many click-to-walk decisions the walk submits: stepping
	// into town and stepping onto the floor item. Every one of them reaches the
	// world as the bridge's own KindMouseWalk row.
	walkClicks = 2
	// lootItemSeed is the seed of the item the descend loads, and lootGold the
	// gold the walk picks up with it: both are boot state of the scripted level,
	// so a replay kernel starts from the same floor.
	lootItemSeed uint32 = 1234
	lootGold     int32  = 100
)

// The walk's tiles. The landed world slice steps one grid step per click
// (world.StepReach), so every click target is adjacent to the hero: the boot tile,
// then one diagonal town step. The dungeon tiles come from the loaded level
// itself (levelEntry + the monster tiles the level's own walkable cells yield).
var (
	spawnTile     = world.Tile{X: 0, Y: 0}
	nextSpawnTile = world.Tile{X: 1, Y: 1}
)

// scenarioSeed is the seed the AC-1 pins walk from: one fixed value, so the
// stage-order and event pins read one walk. The AC-2 matrix walks every seed of
// committedSeeds instead (seedparity_test.go).
var scenarioSeed = uint32(2588)

// stepOrder is the frozen stage order the walk must reach, in order.
var stepOrder = []string{
	stepBoot, stepCreate, stepTownEnter, stepDungeonEnter,
	stepKillFirst, stepKillLast, stepLoot, stepSave, stepLoad,
}

// warriorCombat is the scripted warrior's melee input words for the landed combat
// pipeline: the pipeline takes explicit inputs (combat.Ref) by design, so the walk
// supplies a sword-arm the roll order can act on. ToHitBonus lifts the clamped hit
// percent off its floor and the damage range is wide enough to be lethal in 1/64
// fixed point against monsterHitPoints.
var warriorCombat = combat.Ref{
	Class:      combat.Warrior,
	Level:      1,
	ToHitBonus: 90,
	MinDam:     8,
	MaxDam:     16,
}

// warriorStartingPack is the pack the create verb raises the walk's hero from.
var warriorStartingPack = world.StartingPack{
	Class: warriorClass, BaseStr: 30, BaseMag: 10, BaseDex: 20, BaseVit: 25,
	HPBase: 96, MaxHPBase: 128, ManaBase: 8, MaxManaBase: 8,
}

// levelImages is the sim's LevelSource: the committed cathedral layout plus the
// kill ring the loaded grid yields. The image is loaded once per sim, so a
// replay kernel is served the same bytes the recording kernel was.
type levelImages struct {
	layout *dtest.L1Layout
	ring   []world.Tile
	item   world.Tile
}

// scriptedPacks is the sim's PackSource: the warrior's starting pack. The base
// stats are the warrior row of the committed class table
// (internal/testdata/assets/txtdata/upstream/classes/warrior/attributes.tsv:
// baseStr 30, baseMag 10, baseDex 20, baseVit 25); the vitals are the scripted
// walk's own, because the class-data calculateBaseLife / calculateBaseMana path
// (player.cpp:1976-1988) is not landed yet.
type scriptedPacks struct{}

// eventBus is the walk's consumer of the kernel's events: a kernel.Publisher, so
// it runs after the pass has settled and cannot change what the pass decided
// (ADR-018 law 3). It mirrors the announced consequences into the walk's own
// state: a Descended event moves the session onto the level, a Looted event puts
// the item into the warrior's pack.
type eventBus struct {
	sim *sessionSim
}

// sessionSim is one fresh simulation: the kernel with its gate and streams, the
// landed world slice and movement handler, the test-owned combat handler, the
// input-bridge session that submits at the tick boundary, and a tick-log writer
// over a fresh SQLite store file.
type sessionSim struct {
	walk    *walkState
	grid    *world.State
	levels  *levelImages
	packs   scriptedPacks
	kernel  *kernel.Kernel
	streams *kernel.Streams
	session *inputbridge.Session
	store   *persist.SQLiteStore
	writer  *persist.TickLogWriter
	// passed counts the finished passes, so the tick tape stays strictly
	// increasing across the walk.
	passed uint32
	// recordErr carries the last tick-log write failure the walk asserts on.
	recordErr error
	// seed is the sim's stream seed, so a replay kernel draws the same sequence.
	seed uint32
	// t is the owning test, so a bounded loop can fail instead of hanging.
	t *testing.T
}

// walkAdapter implements testsupport.HashableState over walkState in a fixed
// field order: identity, position, vitals, mode, occupancy.
type walkAdapter struct {
	state *walkState
}

// walkState is the session state the walk freezes: where the hero stands, which
// level the walk is on, the live monster tiles, the warrior pack the persist
// roundtrip carries, and the two roundtrip flags.
type walkState struct {
	// tile is the hero tile the last walk command left the world slice on.
	tile world.Tile
	// dungeon is false in town, true once the descend verb loaded Cathedral L1.
	// Only the KindDescended handler sets it.
	dungeon bool
	// level and levelType are the loaded level's identity, empty in town.
	level, levelType int
	// monsters is the live scripted monster tile set.
	monsters map[world.Tile]bool
	// warrior is the packed pack.h image the persist roundtrip carries.
	warrior persist.PlayerPack
	// packs is the class source the create event's class word is resolved
	// through, so the walk raises the same hero the kernel raised.
	packs scriptedPacks
	// grid is the world slice the walk drives; the descend handler is the only
	// code that changes its layout.
	grid *world.State
	// events is the walk's own view of the kernel's event log: the survive
	// flags the canonical hash folds are read from it, so the harness stamps
	// nothing.
	events eventLog
}

// slotArchive implements world.SessionArchive over the persist store: the save
// verb's write goes out as the three stored blocks (the packed hero row, the
// game heading and the level block), and the load verb's read brings them back.
// It is the diablo-side fill of the world seam, so the world slice never sees a
// database handle (ADR-008).
type slotArchive struct {
	store *persist.SQLiteStore
}

// frozenLevel is the level seed decoded back out of one level block.
type frozenLevel struct{ LevelSeed uint32 }

// eventLog is the walk's view of the kernel's event trace, folded into the two
// survive flags the mode word carries. The flags therefore cannot be set unless
// the kernel announced the change.
type eventLog struct {
	// saved counts the SavedSlot events the kernel emitted.
	saved int
	// loaded counts the LoadedSlot events the kernel emitted.
	loaded int
}

// survivedSave reports whether the session's save went through the kernel.
func (l eventLog) survivedSave() bool { return l.saved > 0 }

// survivedLoad reports whether the session's load went through the kernel.
func (l eventLog) survivedLoad() bool { return l.loaded > 0 }

// StartingPack implements world.PackSource: the warrior class and no other.
func (scriptedPacks) StartingPack(class int) (world.StartingPack, bool) {
	if class != warriorClass {
		return world.StartingPack{}, false
	}

	return warriorStartingPack, true
}

// newWalkState starts the session in town at the spawn with no warrior and no
// monsters: the dungeon's monsters only exist once the descend verb has loaded
// the level, so their tiles come from the loaded grid.
func newWalkState() *walkState {
	return &walkState{tile: spawnTile, monsters: make(map[world.Tile]bool, monsterCount)}
}

// created records the hero the create event announced: the walk's warrior pack
// comes from the class source the event's class word names, and its tile from the
// event's packed spawn words. Only the KindCreatedPack handler calls it, so the
// walk's hero cannot appear without a kernel event.
func (s *walkState) created(event kernel.Event, grid *world.State) {
	pack, ok := s.packs.StartingPack(int(event.A))
	if !ok {
		return
	}

	s.warrior = heroPack(pack)
	s.tile = world.Tile{X: int(event.B), Y: int(event.C)}
	s.grid = grid
}

// entered records the level the descend event announced and seeds the monster
// ring on the level's own walkable tiles. Only the KindDescended handler calls
// it, so the walk's dungeon state cannot move without a kernel event.
func (s *walkState) entered(level, levelType int, hero world.Tile, ring []world.Tile) {
	s.dungeon = true
	s.level, s.levelType = level, levelType
	s.tile = hero

	for _, tile := range ring {
		s.monsters[tile] = true
	}
}

// restored records the session the load event announced: the level the event's
// B word names and the hero tile its C word packs. Only the KindLoadedSlot
// handler calls it, so the walk's restored state cannot move without a kernel
// event.
func (s *walkState) restored(event kernel.Event) {
	s.events.loaded++
	s.dungeon = true
	s.level = int(event.B)
	s.levelType = int(event.C >> 16)
	s.tile = world.Tile{X: int(event.C&0xFFFF) % world.MAXDUNX, Y: int(event.C&0xFFFF) / world.MAXDUNX}

	// The restored level is the same dungeon the save froze, so it holds no live
	// monster: the kills the save outlived stay dead.
	s.monsters = make(map[world.Tile]bool, monsterCount)
}

// snapshot copies the walk state a save freezes, so the roundtrip pin compares
// the loaded walk with what the save wrote instead of with itself.
func (s *walkState) snapshot() walkSnapshot {
	return walkSnapshot{
		tile: s.tile, dungeon: s.dungeon, level: s.level, levelType: s.levelType, warrior: s.warrior,
	}
}

// monsterTiles returns the live monster tiles in tile order, so no map iteration
// decides which monster the next attack targets.
func (s *walkState) monsterTiles() []world.Tile {
	tiles := make([]world.Tile, 0, len(s.monsters))
	for tile := range s.monsters {
		tiles = append(tiles, tile)
	}

	slices.SortFunc(tiles, func(a, b world.Tile) int { return tileKey(a) - tileKey(b) })

	return tiles
}

// liveMonsters counts the monsters still standing.
func (s *walkState) liveMonsters() int { return len(s.monsters) }

// kill removes the monster standing on tile; only the combat handler calls it.
func (s *walkState) kill(tile world.Tile) { delete(s.monsters, tile) }

// tileKey folds a tile into one sortable word.
func tileKey(t world.Tile) int { return t.Y*world.MAXDUNX + t.X }

// CanonicalFields returns the player-visible state in a fixed order with
// fixed-width encodings: no map iteration, no clock, no pointer reaches it.
func (a walkAdapter) CanonicalFields() []testsupport.HashField {
	return []testsupport.HashField{
		{Name: "identity", Bytes: []byte{a.state.warrior.Class, a.state.warrior.Level}},
		{Name: "position", Bytes: encodeTile(a.state.tile)},
		{Name: "vitals", Bytes: encodeHP(a.state.warrior.HPBase)},
		{Name: "mode", Bytes: modeWord(a.state)},
		{Name: "level", Bytes: levelWord(a.state)},
		{Name: "layout", Bytes: a.state.layout()},
		{Name: "monster-tiles", Bytes: encodeMonsterTiles(a.state.monsters)},
	}
}

// modeWord is the canonical mode word: class, level kind, survived save,
// survived load.
func modeWord(s *walkState) []byte {
	return []byte{
		s.warrior.Class, flagByte(s.dungeon), flagByte(s.events.survivedSave()),
		flagByte(s.events.survivedLoad()),
	}
}

// levelWord is the canonical level word: the loaded level's identity. A town or
// unloaded session freezes two zeroes.
func levelWord(s *walkState) []byte {
	buf := make([]byte, 2)
	buf[0], buf[1] = byte(s.level), byte(s.levelType)

	return buf
}

// encodeTile is the fixed-width image of a tile.
func encodeTile(t world.Tile) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(t.Y)<<8|uint16(t.X))

	return buf
}

// encodeHP is the fixed-width image of a health word.
func encodeHP(hp int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(hp))

	return buf
}

// flagByte is the canonical 1/0 image byte of a flag. The invariant is the
// fixed one-byte width the hash is folded over, not the absent branch: picking
// 0/1 from a bool is equally deterministic.
func flagByte(set bool) byte { return map[bool]byte{false: 0, true: 1}[set] }

// encodeMonsterTiles folds the live monster tiles in tile order.
func encodeMonsterTiles(monsters map[world.Tile]bool) []byte {
	keys := make([]int, 0, len(monsters))
	for tile := range monsters {
		keys = append(keys, tileKey(tile))
	}

	slices.Sort(keys)

	buf := make([]byte, 2*len(keys))
	for i, key := range keys {
		buf[2*i] = byte(key >> 8)
		buf[2*i+1] = byte(key)
	}

	return buf
}

// newSessionSim boots one fresh sim at the seed. Every walk runs on its own
// grid and its own store file, so two walks in one process cannot share state.
// The grid starts empty: the walk's dungeon stage loads the committed cathedral
// layout through the Descend verb.
func newSessionSim(t *testing.T, seed uint32) *sessionSim {
	t.Helper()

	sim := &sessionSim{walk: newWalkState(), seed: seed, t: t, levels: newLevelImages(t, seed)}
	sim.grid = world.NewState()
	sim.streams = kernel.NewStreams(seed, streamCombat)
	sim.kernel = kernel.New(world.NewSessionGate(sim.grid, sim.grid, 0), sim.streams)
	world.RegisterMovement(sim.kernel, sim.grid)
	world.RegisterSessionVerbs(sim.kernel, sim.grid, sim.levels)
	world.RegisterCreatePack(sim.kernel, sim.grid, sim.packs)
	sim.kernel.Register(kernel.KindMouseAttack, sim.attackHandler())
	sim.kernel.SetPublisher(&eventBus{sim: sim})

	sim.session = inputbridge.NewSession(sim.kernel)
	sim.store = mustOpenStore(t, filepath.Join(t.TempDir(), fmt.Sprintf("hero-%x.db", seed)))
	sim.writer = sim.store.NewTickLogWriter(persist.TickLogOptions{SnapshotEvery: -1, Batch: 1})
	world.RegisterSaveLoad(sim.kernel, sim.grid, slotArchive{store: sim.store}, sim.levels)

	return sim
}

// freshKernel builds a second kernel over a fresh world slice, wired exactly like
// the recording sim's: the same handlers, the same seed, the same level images.
// Replay determinism means the same handlers, seed and commands reproduce the
// same checksums (ADR-015).
func (m *sessionSim) freshKernel() *kernel.Kernel {
	fresh := &sessionSim{walk: newWalkState(), grid: world.NewState(), levels: m.levels}
	fresh.streams = kernel.NewStreams(m.seed, streamCombat)
	fresh.kernel = kernel.New(world.NewSessionGate(fresh.grid, fresh.grid, 0), fresh.streams)
	world.RegisterMovement(fresh.kernel, fresh.grid)
	world.RegisterSessionVerbs(fresh.kernel, fresh.grid, fresh.levels)
	world.RegisterCreatePack(fresh.kernel, fresh.grid, fresh.packs)
	fresh.kernel.Register(kernel.KindMouseAttack, fresh.attackHandler())
	fresh.kernel.SetPublisher(&eventBus{sim: fresh})

	// The replayed kernel writes through the same archive the recording kernel
	// used: a replayed load therefore restores the slot the recorded save
	// wrote, and the emitted events match the recorded ones.
	world.RegisterSaveLoad(fresh.kernel, fresh.grid, slotArchive{store: m.store}, m.levels)

	return fresh.kernel
}

// attackHandler is the test-owned combat handler: a MouseAttack command carries
// the monster id in its C word (kind_of_kind: A the ranged flag, B the verb, C
// the id target), the handler resolves it to a tile, rolls the landed combat
// pipeline over the combat stream and, on a lethal blow, kills the monster and
// announces MonsterDied (ADR-018 law 2).
func (m *sessionSim) attackHandler() kernel.Handler {
	return func(c kernel.Command, streams *kernel.Streams) []kernel.Event {
		tile, ok := m.monsterTile(int(c.C))
		if !ok {
			return nil
		}

		hit, damage := combat.PlrHitMonstCore(warriorCombat, combatRnd(streams))
		if !hit || damage < monsterHitPoints {
			return nil
		}

		m.walk.kill(tile)

		return []kernel.Event{{Kind: kernel.KindMonsterDied, A: c.C, B: int32(tile.X), C: int32(tile.Y)}}
	}
}

// monsterTile resolves a scripted monster id back to its live tile.
func (m *sessionSim) monsterTile(id int) (world.Tile, bool) {
	for tile := range m.walk.monsters {
		if monsterIDFor(tile) == id {
			return tile, true
		}
	}

	return world.Tile{}, false
}

// combatRnd adapts the kernel stream to combat's GenerateRnd signature.
func combatRnd(streams *kernel.Streams) func(int) int {
	return func(n int) int {
		drawn, ok := streams.Draw(streamCombat)
		if !ok {
			return 0
		}

		return int(drawn) % n
	}
}

// monsterIDFor is the scripted monster id: the tile's folded key, the word the
// attack command carries and the handler resolves back.
func monsterIDFor(tile world.Tile) int { return tileKey(tile) }

// mustOpenStore opens a SQLite store or fails the test.
func mustOpenStore(t *testing.T, path string) *persist.SQLiteStore {
	t.Helper()

	store, err := persist.OpenSQLiteStore(context.Background(), path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	return store
}

// clickWalk submits one click-to-walk decision for tile and runs the tick. The
// landed input bridge owns the decision rows (LeftMouseCmd), and the world slice
// owns the walk: the KindMouseWalk command the bridge builds reaches the grid
// through the kernel's own pass, exactly as it does for a real click.
func (m *sessionSim) clickWalk(tick uint32, tile world.Tile) {
	decision := controls.MousePathInput{
		Cursor: controls.CursorHand, Town: !m.walk.dungeon, Item: -1, Monster: -1, Player: -1, Object: -1,
	}

	m.submitDecision(tick, decision, tile)
}

// clickAttack submits one attack decision for the monster standing on tile and
// runs the tick. The mouse path resolves the monster id itself, so the id the
// combat handler receives is the bridge's, not the test's.
func (m *sessionSim) clickAttack(tick uint32, tile world.Tile) {
	decision := controls.MousePathInput{
		Cursor: controls.CursorHand, Town: false, Item: -1, Monster: monsterIDFor(tile), Player: -1, Object: -1,
	}

	m.submitDecision(tick, decision, tile)
}

// submitDecision resolves one click through the bridge, queues the command its
// verb row means, and closes the tick.
func (m *sessionSim) submitDecision(tick uint32, decision controls.MousePathInput, tile world.Tile) {
	command, ok := inputbridge.MousePathCommand(decision, controls.MousePathTarget{X: tile.X, Y: tile.Y})
	if !ok {
		m.t.Fatalf("click on tile %v produced no command", tile)
	}

	m.session.SubmitDecision(command)
	m.runTick(tick)
}

// runTick closes the tick boundary, reads the hero tile back out of the landed
// world slice, and records the finished pass into the tick log, so the recorded
// trail is the kernel's own record of the walk.
func (m *sessionSim) runTick(tick uint32) {
	m.session.Tick(tick)

	m.walk.tile = m.grid.Hero()

	record := persist.TickRecord{
		Tick:     tick,
		Commands: m.kernel.TraceCommandsAt(tick),
		Events:   m.kernel.TraceEventsAt(tick),
		Checksum: m.kernel.TraceChecksumAt(tick),
	}

	if err := m.writer.Record(context.Background(), record); err != nil {
		m.recordErr = err
	}

	m.passed++
}

// SaveSlot writes one session record as the store's three blocks.
func (a slotArchive) SaveSlot(rec world.SessionRecord) error {
	if err := a.saveHero(rec); err != nil {
		return err
	}

	if err := a.saveGame(rec); err != nil {
		return err
	}

	return a.saveLevel(rec)
}

// LoadSlot reads the three blocks back, reporting ErrNoSlot when no hero row
// holds the slot.
func (a slotArchive) LoadSlot(slot int) (world.SessionRecord, error) {
	ctx := context.Background()

	hero, err := a.store.LoadHero(ctx, slot)
	if err != nil {
		return world.SessionRecord{}, world.ErrNoSlot
	}

	game, err := a.store.LoadGame(ctx, slot, string(persist.KindSingle))
	if err != nil {
		return world.SessionRecord{}, fmt.Errorf("load game block: %w", err)
	}

	header, _, err := persist.DecodeGameHeader(game)
	if err != nil {
		return world.SessionRecord{}, fmt.Errorf("decode game header: %w", err)
	}

	level, err := a.store.LoadLevel(ctx, slot, int(header.CurrLevel))
	if err != nil {
		return world.SessionRecord{}, fmt.Errorf("load level block: %w", err)
	}

	frozen, err := decodeLevelBlock(level)
	if err != nil {
		return world.SessionRecord{}, err
	}

	pack, err := decodeHeroPack(hero)
	if err != nil {
		return world.SessionRecord{}, err
	}

	return world.SessionRecord{
		Slot:      slot,
		LevelSeed: frozen.LevelSeed,
		Level:     int(header.CurrLevel),
		LevelType: int(header.LevelType),
		Loaded:    header.CurrLevel != 0,
		Hero:      world.Tile{X: int(header.ViewX), Y: int(header.ViewY)},
		Pack:      pack,
	}, nil
}

// saveHero writes the record's packed hero row.
func (a slotArchive) saveHero(rec world.SessionRecord) error {
	hero := persist.HeroRecord{
		Slot: rec.Slot, Name: heroNameText, ClassID: rec.Pack.Class,
		Pack: heroPack(rec.Pack).Encode(), Spawn: true,
	}

	if err := a.store.SaveHero(context.Background(), hero); err != nil {
		return fmt.Errorf("save hero: %w", err)
	}

	return nil
}

// saveGame writes the record's game-heading block.
func (a slotArchive) saveGame(rec world.SessionRecord) error {
	header, err := saveGameHeader(rec).Encode()
	if err != nil {
		return fmt.Errorf("encode game header: %w", err)
	}

	if err := a.store.SaveGame(context.Background(), rec.Slot, string(persist.KindSingle), header); err != nil {
		return fmt.Errorf("save game: %w", err)
	}

	return nil
}

// saveLevel writes the record's level block.
func (a slotArchive) saveLevel(rec world.SessionRecord) error {
	data := encodeLevelBlock(rec.LevelSeed, rec.Hero)
	if err := a.store.SaveLevel(context.Background(), rec.Slot, rec.Level, data); err != nil {
		return fmt.Errorf("save level: %w", err)
	}

	return nil
}

// saveGameHeader is the GameHeader a session record freezes: the record's level
// identity and the hero tile a load drops the hero back onto.
func saveGameHeader(rec world.SessionRecord) persist.GameHeader {
	levelType, currLevel := uint32(0), uint32(0)
	if rec.Loaded {
		levelType, currLevel = uint32(rec.LevelType), uint32(rec.Level)
	}

	return persist.GameHeader{
		Spawn: true, CurrLevel: currLevel, LevelType: levelType,
		ViewX: int32(rec.Hero.X), ViewY: int32(rec.Hero.Y),
	}
}

// encodeLevelBlock writes the record's level block: the upstream counter header,
// the grids the hero lit, then the level seed this slice carries. Upstream keeps
// the seed in the save game's own level-seed array (loadsave.cpp:1907-1918); the
// scripts table column follows at the next schema version, so the block keeps it
// here instead of inventing a column in a lane outside this contract's scope.
func encodeLevelBlock(seed uint32, hero world.Tile) []byte {
	header := persist.LevelHeader{ActiveItems: 1}
	grids := persist.TileGrids{}

	index := hero.Y*world.MAXDUNX + hero.X
	grids.Lights[index] = 15
	grids.Automap[hero.Y%persist.AutomapHeight*persist.AutomapWidth+hero.X%persist.AutomapWidth] = 1

	block := append(header.Encode(), grids.Encode()...)

	return binary.LittleEndian.AppendUint32(block, seed)
}

// decodeLevelBlock reads the counter header and the seed tail of one block.
func decodeLevelBlock(data []byte) (frozenLevel, error) {
	if len(data) < persist.LevelHeaderSize+4 {
		return frozenLevel{}, fmt.Errorf("level block = %d bytes, want at least %d",
			len(data), persist.LevelHeaderSize+4)
	}

	if _, err := persist.DecodeLevelHeader(data); err != nil {
		return frozenLevel{}, fmt.Errorf("decode level header: %w", err)
	}

	return frozenLevel{LevelSeed: binary.LittleEndian.Uint32(data[len(data)-4:])}, nil
}

// decodeHeroPack rebuilds the starting pack a stored hero row was raised from.
func decodeHeroPack(hero persist.HeroRecord) (world.StartingPack, error) {
	decoded, err := persist.DecodePlayerPack(hero.Pack)
	if err != nil {
		return world.StartingPack{}, fmt.Errorf("decode hero pack: %w", err)
	}

	return startingPackFrom(decoded), nil
}

// startingPackFrom projects a decoded hero row back onto the pack the create
// verb raised it from.
func startingPackFrom(hero persist.PlayerPack) world.StartingPack {
	return world.StartingPack{
		Class:   int(hero.Class),
		BaseStr: int(hero.BaseStr), BaseMag: int(hero.BaseMag),
		BaseDex: int(hero.BaseDex), BaseVit: int(hero.BaseVit),
		HPBase: int(hero.HPBase), MaxHPBase: int(hero.MaxHP),
		ManaBase: int(hero.ManaBase), MaxManaBase: int(hero.MaxMana),
	}
}

// heroPack builds the packed hero the persist roundtrip carries out of the pack
// the create verb raised.
func heroPack(starting world.StartingPack) persist.PlayerPack {
	hero := persist.PlayerPack{Class: uint8(starting.Class), Level: 1}
	copy(hero.Name[:], heroNameText)

	hero.BaseStr = uint8(starting.BaseStr)
	hero.BaseMag = uint8(starting.BaseMag)
	hero.BaseDex = uint8(starting.BaseDex)
	hero.BaseVit = uint8(starting.BaseVit)

	hero.HPBase = int32(starting.HPBase)
	hero.MaxHP = int32(starting.MaxHPBase)
	hero.ManaBase = int32(starting.ManaBase)
	hero.MaxMana = int32(starting.MaxManaBase)

	return hero
}

// newLevelImages loads the seed's own committed parity-true cathedral layout and
// derives the monster ring and the floor item from its own walkable tiles. The
// seed picks the corpus variant, so every walked seed loads the dungeon that
// seed generates.
func newLevelImages(t *testing.T, seed uint32) *levelImages {
	t.Helper()

	layout := dtest.LoadL1Layout(t, dungeonLevel, seed)
	item := floorItemTile(layout.Image, layout.Entry)

	// The item the level starts with lies on a walkable tile one step from the
	// entry, so the walk can step onto it and take it (InitItems,
	// diablo.cpp:3155 leaves the level's own drops on the floor).
	layout.Image.Items = []world.FloorItem{{Index: 0, Tile: item}}

	return &levelImages{layout: layout, ring: killRing(layout.Image, layout.Entry), item: item}
}

// LevelImage implements world.LevelSource: the walk's cathedral level and no
// other. The entry tile is the level's own ViewPosition, which LoadLevel then
// re-resolves through the SyncInitPlrPos window (player.cpp:3235-3271).
func (l *levelImages) LevelImage(level, levelType int) (*world.LevelLayout, world.Tile, bool) {
	if level != dungeonLevel || levelType != dungeonLevelType {
		return nil, world.Tile{}, false
	}

	return l.layout.Image, l.layout.Entry, true
}

// Ring reports the monster tiles the loaded level yields.
func (l *levelImages) Ring() []world.Tile { return l.ring }

// Item reports the tile the level's floor item lies on.
func (l *levelImages) Item() world.Tile { return l.item }

// floorItemTile picks the first walkable cardinal neighbour of entry, in a fixed
// direction order, so the walk has a real tile to step onto and loot.
func floorItemTile(image *world.LevelLayout, entry world.Tile) world.Tile {
	for _, delta := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
		tile := world.Tile{X: entry.X + delta[0], Y: entry.Y + delta[1]}
		if world.SOLTileProperties(image.SOL[image.Piece[tile.X*world.MAXDUNY+tile.Y]])&world.TileSolid == 0 {
			return tile
		}
	}

	return entry
}

// killRing picks monsterCount walkable tiles from the real layout: the tiles are
// collected in tile order, so no map iteration decides which monster is where,
// and the picks start at the level entry, so the walk meets them along the
// dungeon it actually loaded.
func killRing(image *world.LevelLayout, entry world.Tile) []world.Tile {
	candidates := walkableTiles(image)

	start := 0
	for i, tile := range candidates {
		if tile == entry {
			start = i

			break
		}
	}

	ring := make([]world.Tile, 0, monsterCount)
	for i := range monsterCount {
		ring = append(ring, candidates[(start+i)%len(candidates)])
	}

	return ring
}

// walkableTiles lists every non-solid tile of the image in tile order, using the
// same classification the grid applies (world.SOLTileProperties over the piece's
// SOL entry).
func walkableTiles(image *world.LevelLayout) []world.Tile {
	tiles := make([]world.Tile, 0, world.MAXDUNX*world.MAXDUNY)

	for y := range world.MAXDUNY {
		for x := range world.MAXDUNX {
			tile := world.Tile{X: x, Y: y}
			if world.SOLTileProperties(image.SOL[image.Piece[x*world.MAXDUNY+y]])&world.TileSolid == 0 {
				tiles = append(tiles, tile)
			}
		}
	}

	return tiles
}

// layout freezes the loaded grid's layout: the piece layer plus its SOL
// classification, folded by the grid itself. An unloaded session freezes zero.
func (s *walkState) layout() []byte {
	if s.grid == nil {
		return make([]byte, 8)
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, s.grid.Grid().LayoutDigest())

	return buf
}

// createPack submits the create verb for one class on one spawn tile and runs
// the tick. The walk's warrior comes from the KindCreatedPack handler alone, so a
// dropped or failed create leaves the session with no warrior.
func (m *sessionSim) createPack(tick uint32, class int, tile world.Tile) {
	m.session.SubmitDecision(kernel.Command{
		Kind: kernel.KindCreatePack, A: int32(class), B: int32(tile.X), C: int32(tile.Y),
	})
	m.runTick(tick)
}

// saveSlot submits the save verb for one slot and runs the tick. The walk's
// survived-save flag comes from the SavedSlot event alone, so a dropped or
// failed save leaves the flag clear.
func (m *sessionSim) save(tick uint32) {
	m.session.SubmitDecision(kernel.Command{Kind: kernel.KindSaveSlot, A: saveSlot, B: int32(m.seed)})
	m.runTick(tick)
}

// load submits the load verb for the same slot and runs the tick. The seed word
// is the sim's own, so a load only restores the slot the same seed saved.
func (m *sessionSim) load(tick uint32) {
	m.session.SubmitDecision(kernel.Command{Kind: kernel.KindLoadSlot, A: saveSlot, B: int32(m.seed)})
	m.runTick(tick)
}

// descend submits the descend verb for the walk's dungeon level and runs the
// tick. The walk's dungeon state is written by the KindDescended handler alone,
// so a dropped or failed descend leaves the session in town.
func (m *sessionSim) descend(tick uint32) {
	m.session.SubmitDecision(kernel.Command{Kind: kernel.KindDescend, A: dungeonLevel, B: dungeonLevelType})
	m.runTick(tick)
}

// loot submits the loot verb for the player's own tile and runs the tick.
func (m *sessionSim) loot(tick uint32) {
	m.session.SubmitDecision(kernel.Command{Kind: kernel.KindLoot})
	m.runTick(tick)
}

// Publish applies one finished pass's events to the walk state, in trace order.
func (b *eventBus) Publish(events []kernel.Event) {
	for _, event := range events {
		b.apply(event)
	}
}

// apply routes one event to the walk-state change it announces.
func (b *eventBus) apply(event kernel.Event) {
	switch event.Kind {
	case kernel.KindCreatedPack:
		b.sim.walk.created(event, b.sim.grid)
	case kernel.KindDescended:
		b.sim.walk.entered(int(event.A), int(event.B), b.sim.grid.Hero(), b.sim.levels.Ring())
	case kernel.KindLooted:
		b.sim.walk.takeItem(int(event.A))
	case kernel.KindSavedSlot:
		b.sim.walk.events.saved++
	case kernel.KindLoadedSlot:
		b.sim.walk.restored(event)
	default:
	}
}

// takeItem puts one looted item into the warrior's pack: the pack half of the
// pickup (InvGetItem hands the item to the inventory, inv.cpp:1685-1700).
func (s *walkState) takeItem(index int) {
	warrior := &s.warrior
	warrior.NumInv = 1
	warrior.InvList[0] = persist.ItemPack{Idx: uint16(index), Seed: lootItemSeed}
	warrior.InvGrid[0] = 1
	warrior.Gold += lootGold
}
