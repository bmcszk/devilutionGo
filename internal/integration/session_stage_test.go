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
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
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
)

// The walk's tiles. The landed world slice boots the hero at the zero tile and
// steps one grid step per click (world.StepReach), so every click target is the
// tile adjacent to the hero: spawn is the boot tile, then one diagonal town step,
// then one diagonal dungeon step, then one cardinal step onto the loot.
var (
	spawnTile     = world.Tile{X: 0, Y: 0}
	nextSpawnTile = world.Tile{X: 1, Y: 1}
	dungeonTile   = world.Tile{X: 2, Y: 2}
	lootTile      = world.Tile{X: 2, Y: 3}
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

// sessionSim is one fresh simulation: the kernel with its gate and streams, the
// landed world slice and movement handler, the test-owned combat handler, the
// input-bridge session that submits at the tick boundary, and a tick-log writer
// over a fresh SQLite store file.
type sessionSim struct {
	walk    *walkState
	grid    *world.State
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
// level kind the walk is on, the live monster tiles, the warrior pack the persist
// roundtrip carries, and the two roundtrip flags.
type walkState struct {
	// tile is the hero tile the last walk command reached.
	tile world.Tile
	// dungeon is false in town, true on Cathedral L1.
	dungeon bool
	// monsters is the live scripted monster tile set.
	monsters map[world.Tile]bool
	// warrior is the packed pack.h image the persist roundtrip carries.
	warrior persist.PlayerPack
	// saved is set once the session blocks are written through the store.
	saved bool
	// loaded is set once the hero pack was loaded back from the store.
	loaded bool
}

// newWalkState starts the session at the spawn: the world booted, no warrior,
// and the dungeon's monster ring already standing. The ring is boot state, not a
// stage mutation, so a replay kernel seeded from the same walk starts with the
// same monsters and reproduces every kill from the recorded commands alone.
func newWalkState() *walkState {
	state := &walkState{tile: spawnTile, monsters: make(map[world.Tile]bool, monsterCount)}

	for i := range monsterCount {
		state.monsters[monsterRingTile(i)] = true
	}

	return state
}

// descend moves the session onto Cathedral L1.
func (s *walkState) descend() {
	s.dungeon = true
}

// monsterRingTile is the tile of the i-th scripted monster: the ring the kills
// clear one monster at a time.
func monsterRingTile(i int) world.Tile {
	return world.Tile{X: dungeonTile.X + 2 + i, Y: dungeonTile.Y + 2}
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
		{Name: "monster-tiles", Bytes: encodeMonsterTiles(a.state.monsters)},
	}
}

// modeWord is the canonical mode word: class, level kind, survived save,
// survived load.
func modeWord(s *walkState) []byte {
	return []byte{s.warrior.Class, flagByte(s.dungeon), flagByte(s.saved), flagByte(s.loaded)}
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

// flagByte is the canonical 1/0 image byte of a flag. The value is hash data, so
// it is looked up rather than branched on: no bool ever steers control flow here.
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
// fresh grid and its own store file, so two walks in one process cannot share
// state.
func newSessionSim(t *testing.T, seed uint32) *sessionSim {
	t.Helper()

	sim := &sessionSim{walk: newWalkState(), seed: seed, t: t}
	sim.grid = scriptedGrid()
	sim.streams = kernel.NewStreams(seed, streamCombat)
	sim.kernel = kernel.New(world.NewGate(sim.grid, 0), sim.streams)
	world.RegisterMovement(sim.kernel, sim.grid)
	sim.kernel.Register(kernel.KindMouseAttack, sim.attackHandler())

	sim.session = inputbridge.NewSession(sim.kernel)
	sim.path = filepath.Join(t.TempDir(), fmt.Sprintf("hero-%x.db", seed))
	sim.store = mustOpenStore(t, sim.path)
	sim.writer = sim.store.NewTickLogWriter(persist.TickLogOptions{SnapshotEvery: -1, Batch: 1})

	return sim
}

// freshKernel builds a second kernel over a fresh scripted world slice, wired
// exactly like the recording sim's: the same handlers, the same seed. Replay
// determinism means the same handlers, seed and commands reproduce the same
// checksums (ADR-015).
func (m *sessionSim) freshKernel() *kernel.Kernel {
	fresh := &sessionSim{walk: newWalkState(), grid: scriptedGrid()}
	fresh.streams = kernel.NewStreams(m.seed, streamCombat)
	fresh.kernel = kernel.New(world.NewGate(fresh.grid, 0), fresh.streams)
	world.RegisterMovement(fresh.kernel, fresh.grid)
	fresh.kernel.Register(kernel.KindMouseAttack, fresh.attackHandler())

	return fresh.kernel
}

// scriptedGrid builds the world slice on an all-walkable grid.
func scriptedGrid() *world.State {
	grid := world.NewState()

	for i := range grid.Grid().SOLData {
		grid.Grid().SOLData[i] = world.TileNone
	}

	return grid
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

// warriorPack builds the packed warrior the persist roundtrip carries.
func warriorPack(at world.Tile) persist.PlayerPack {
	warrior := persist.PlayerPack{Class: 0, Level: 1, Px: uint8(at.X), Py: uint8(at.Y)}
	copy(warrior.Name[:], heroNameText)

	warrior.BaseStr, warrior.BaseMag, warrior.BaseDex, warrior.BaseVit = 30, 10, 20, 25
	warrior.HPBase, warrior.MaxHP, warrior.ManaBase, warrior.MaxMana = 96, 128, 8, 8

	return warrior
}

// heroName reads the packed name field.
func heroName(warrior persist.PlayerPack) string {
	name := slices.Clone(warrior.Name[:])
	for len(name) > 0 && name[len(name)-1] == 0 {
		name = name[:len(name)-1]
	}

	return string(name)
}
