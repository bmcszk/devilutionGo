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

// scenarioSeed is the fixed simulation seed every walk starts from. It is the
// only entropy in the harness: two walks at this seed must freeze equal hashes.
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
	path    string
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
	// saved is set once the session blocks are written through the store.
	saved bool
	// loaded is set once the hero pack was loaded back from the store.
	loaded bool
}

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
	return []byte{s.warrior.Class, flagByte(s.dungeon), flagByte(s.saved), flagByte(s.loaded)}
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

	sim := &sessionSim{walk: newWalkState(), seed: seed, t: t, levels: newLevelImages(t)}
	sim.grid = world.NewState()
	sim.streams = kernel.NewStreams(seed, streamCombat)
	sim.kernel = kernel.New(world.NewSessionGate(sim.grid, sim.grid, 0), sim.streams)
	world.RegisterMovement(sim.kernel, sim.grid)
	world.RegisterSessionVerbs(sim.kernel, sim.grid, sim.levels)
	world.RegisterCreatePack(sim.kernel, sim.grid, sim.packs)
	sim.kernel.Register(kernel.KindMouseAttack, sim.attackHandler())
	sim.kernel.SetPublisher(&eventBus{sim: sim})

	sim.session = inputbridge.NewSession(sim.kernel)
	sim.path = filepath.Join(t.TempDir(), fmt.Sprintf("hero-%x.db", seed))
	sim.store = mustOpenStore(t, sim.path)
	sim.writer = sim.store.NewTickLogWriter(persist.TickLogOptions{SnapshotEvery: -1, Batch: 1})

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
// landed input bridge owns the decision rows (LeftMouseCmd); the harness asks it
// for the command and submits that command's movement verb, because the landed
// world slice registers its handler on KindPlayerMove and no landed surface maps
// the bridge's KindMouseWalk row onto it yet.
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

	m.session.SubmitDecision(movementVerb(command))
	m.runTick(tick)
}

// movementVerb maps the bridge's walk row onto the world slice's movement verb.
// Every other row submits unchanged, so the gate and the trace see the bridge's
// own command.
func movementVerb(command kernel.Command) kernel.Command {
	if command.Kind != kernel.KindMouseWalk {
		return command
	}

	return kernel.Command{Kind: kernel.KindPlayerMove, A: command.A, B: command.B}
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

// saveSession writes the session's blocks through the persist store: the packed
// warrior row, the game heading and the level block.
func (m *sessionSim) saveSession(ctx context.Context, t *testing.T) {
	t.Helper()

	m.saveHero(ctx, t)
	m.saveGame(ctx, t)
	m.saveLevel(ctx, t)
}

// saveHero writes the packed warrior row.
func (m *sessionSim) saveHero(ctx context.Context, t *testing.T) {
	t.Helper()

	hero := persist.HeroRecord{
		Slot: 0, Name: heroName(m.walk.warrior), ClassID: int(m.walk.warrior.Class),
		Pack: m.walk.warrior.Encode(), Spawn: true,
	}

	if err := m.store.SaveHero(ctx, hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}
}

// saveGame writes the game-heading blob.
func (m *sessionSim) saveGame(ctx context.Context, t *testing.T) {
	t.Helper()

	header, err := m.header().Encode()
	if err != nil {
		t.Fatalf("GameHeader.Encode: %v", err)
	}

	if err := m.store.SaveGame(ctx, 0, "single", header); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}
}

// saveLevel writes the level block.
func (m *sessionSim) saveLevel(ctx context.Context, t *testing.T) {
	t.Helper()

	if err := m.store.SaveLevel(ctx, 0, 1, encodeLevel(m.walk)); err != nil {
		t.Fatalf("SaveLevel: %v", err)
	}
}

// loadSession reopens the store file and loads every block back: the hero pack,
// the game heading and the level block. A block that does not decode fails the
// test, so the load half proves the save half.
func (m *sessionSim) loadSession(ctx context.Context, t *testing.T) {
	t.Helper()

	reopened, err := persist.OpenSQLiteStore(ctx, m.path)
	if err != nil {
		t.Fatalf("OpenSQLiteStore: %v", err)
	}

	defer func() {
		if err := reopened.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	m.loadHero(ctx, t, reopened)
	m.loadGame(ctx, t, reopened)
	m.loadLevel(ctx, t, reopened)

	m.walk.loaded = true
}

// loadHero loads the hero pack back and makes it the walk's warrior.
func (m *sessionSim) loadHero(ctx context.Context, t *testing.T, store *persist.SQLiteStore) {
	t.Helper()

	hero, err := store.LoadHero(ctx, 0)
	if err != nil {
		t.Fatalf("LoadHero: %v", err)
	}

	decoded, err := persist.DecodePlayerPack(hero.Pack)
	if err != nil {
		t.Fatalf("DecodePlayerPack: %v", err)
	}

	m.walk.warrior = decoded
}

// loadGame loads the game heading back and checks its own magic decoded.
func (*sessionSim) loadGame(ctx context.Context, t *testing.T, store *persist.SQLiteStore) {
	t.Helper()

	blob, err := store.LoadGame(ctx, 0, "single")
	if err != nil {
		t.Fatalf("LoadGame: %v", err)
	}

	if _, _, err := persist.DecodeGameHeader(blob); err != nil {
		t.Fatalf("DecodeGameHeader: %v", err)
	}
}

// loadLevel loads the level block back and checks its grids decoded.
func (*sessionSim) loadLevel(ctx context.Context, t *testing.T, store *persist.SQLiteStore) {
	t.Helper()

	blob, err := store.LoadLevel(ctx, 0, 1)
	if err != nil {
		t.Fatalf("LoadLevel: %v", err)
	}

	if len(blob) < persist.LevelHeaderSize {
		t.Fatalf("level block = %d bytes, want at least %d", len(blob), persist.LevelHeaderSize)
	}

	if _, err := persist.DecodeLevelHeader(blob); err != nil {
		t.Fatalf("DecodeLevelHeader: %v", err)
	}
}

// header is the session's GameHeader: the walk's level kind and view position.
func (m *sessionSim) header() persist.GameHeader {
	levelType, currLevel := uint32(0), uint32(0)
	if m.walk.dungeon {
		levelType, currLevel = 1, 1
	}

	return persist.GameHeader{
		Spawn:       true,
		CurrLevel:   currLevel,
		LevelType:   levelType,
		ViewX:       int32(m.walk.tile.X),
		ViewY:       int32(m.walk.tile.Y),
		ActiveItems: int32(m.walk.warrior.NumInv),
	}
}

// encodeLevel encodes the walk's level block: the counter header and the grids
// the walk lit.
func encodeLevel(walk *walkState) []byte {
	header := persist.LevelHeader{ActiveItems: int32(walk.warrior.NumInv)}
	grids := persist.TileGrids{}

	index := walk.tile.Y*world.MAXDUNX + walk.tile.X
	grids.Lights[index] = 15
	grids.Automap[walk.tile.Y%persist.AutomapHeight*persist.AutomapWidth+walk.tile.X%persist.AutomapWidth] = 1

	return append(header.Encode(), grids.Encode()...)
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

// heroName reads the packed name field.
func heroName(warrior persist.PlayerPack) string {
	name := slices.Clone(warrior.Name[:])
	for len(name) > 0 && name[len(name)-1] == 0 {
		name = name[:len(name)-1]
	}

	return string(name)
}

// newLevelImages loads the committed parity-true cathedral layout and derives the
// monster ring and the floor item from its own walkable tiles.
func newLevelImages(t *testing.T) *levelImages {
	t.Helper()

	layout := dtest.LoadL1Layout(t, dungeonLevel, scenarioSeed)
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
