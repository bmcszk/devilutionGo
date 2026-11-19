// The M1 scripted-session test (devilGo-ftbu AC-1): one headless walk — boot,
// create warrior, town, dungeon descend, kill ten, loot, save, load — driven
// from the landed surfaces and frozen into a canonical StateHash at each of the
// nine stages, equal within one run and across two independent runs.
package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// statehashRunEnv is the env var naming the dir the first green run writes its
// per-stage hash list to.
const statehashRunEnv = "M1_LANE_DIR"

// lastRun is the most recent scripted walk, the run the harness writes out.
var lastRun *scriptedRun

// stepHash is one frozen stage: its name and its canonical StateHash.
type stepHash struct {
	Step    string `json:"step"`
	HashHex string `json:"hashHex"`
}

// scriptedRun is one walk: the sim it drove, the stage names and hashes it froze,
// and the recorded tick-log fold.
type scriptedRun struct {
	sim    *sessionSim
	steps  []string
	hashes []uint64
	fold   uint64
}

// TestM1ScriptedSessionRunReplay pins the within-run half of AC-1: two fresh
// sims at the same seed walk the same tape and freeze identical per-stage hashes.
func TestM1ScriptedSessionRunReplay(t *testing.T) {
	// given the scripted walk on a fresh sim
	bootstrap := runScriptedWalk(t, scenarioSeed)

	// when the same tape walks a second fresh sim at the same seed, on its own
	// grid and its own store file
	replay := runScriptedWalk(t, scenarioSeed)

	// then the walks froze the same stages in the same order
	if !slices.Equal(replay.steps, bootstrap.steps) {
		t.Fatalf("walk steps %v, want %v", replay.steps, bootstrap.steps)
	}

	// and every stage's replay hash equals the bootstrap run's
	for i, want := range bootstrap.hashes {
		if got := replay.hashes[i]; got != want {
			t.Fatalf("stage %s: replay StateHash = %x, want the bootstrap run's %x",
				bootstrap.steps[i], got, want)
		}
	}

	// and the two runs folded the same recorded tick-log trail
	if replay.fold != bootstrap.fold {
		t.Fatalf("replay tick-log fold = %x, want the bootstrap run's %x", replay.fold, bootstrap.fold)
	}
}

// TestM1ScriptedSessionNineStages pins the walk's stage list: the nine stages of
// AC-1, in order, no stage more.
func TestM1ScriptedSessionNineStages(t *testing.T) {
	// given the walk on a fresh sim
	run := runScriptedWalk(t, scenarioSeed)

	// when the walk's frozen stages are read back
	steps := run.steps

	// then the walk froze exactly the contract's stages, in order
	if !slices.Equal(steps, stepOrder) {
		t.Fatalf("walk steps %v, want %v", steps, stepOrder)
	}
}

// TestM1ScriptedSessionStagesAdvance pins that each stage is a real advance: no
// stage freezes a hash an earlier stage already froze, so a stage that claims
// movement, a kill or a roundtrip cannot be a no-op.
func TestM1ScriptedSessionStagesAdvance(t *testing.T) {
	// given the walk on a fresh sim
	run := runScriptedWalk(t, scenarioSeed)

	// when the frozen hashes are compared stage by stage
	seen := make(map[uint64]string, len(run.steps))

	// then no stage repeats an earlier stage's hash
	for i, step := range run.steps {
		if prior, repeated := seen[run.hashes[i]]; repeated {
			t.Fatalf("stage %s froze StateHash %x already frozen at stage %s", step, run.hashes[i], prior)
		}

		seen[run.hashes[i]] = step
	}
}

// TestM1ScriptedSessionKillsAreKernelDriven pins that the ten kills come out of
// the kernel: the recorded log carries ten MonsterDied events. A walk that
// deletes monsters from the test side records none.
func TestM1ScriptedSessionKillsAreKernelDriven(t *testing.T) {
	// given the scripted walk recorded into its tick log
	run := runScriptedWalk(t, scenarioSeed)

	// when the recorded log is read back
	log, err := run.sim.store.LoadTickLog(context.Background())
	if err != nil {
		t.Fatalf("LoadTickLog: %v", err)
	}

	// then every kill is a kernel event and no monster is left standing
	if died := countEvents(log, kernel.KindMonsterDied); died != monsterCount {
		t.Fatalf("tick log carries %d MonsterDied events, want %d", died, monsterCount)
	}

	if live := run.sim.walk.liveMonsters(); live != 0 {
		t.Fatalf("%d monsters still live after the walk, want 0", live)
	}
}

// TestM1ScriptedSessionReplayTrail pins the replay half of AC-1: the recorded
// walk drives a fresh kernel of the same handlers through persist.Replay, and
// the reproduced checksum trail folds to the recorded run's fold.
func TestM1ScriptedSessionReplayTrail(t *testing.T) {
	// given the scripted walk recorded into its tick log
	run := runScriptedWalk(t, scenarioSeed)

	// when the log replays on a fresh kernel at the same seed
	result, err := replayTrail(t, run)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	// then no tick diverged
	if result.Mismatch != nil {
		t.Fatalf("replay diverged at tick %d: %x != %x",
			result.Mismatch.Tick, result.Mismatch.Left, result.Mismatch.Right)
	}

	// and the replayed final fold matches the recorded run's
	if result.Final != run.fold {
		t.Fatalf("replay final fold = %x, want the recorded run's %x", result.Final, run.fold)
	}
}

// TestMain writes the first green run's per-stage hash list under the lane dir
// when the env names one, and propagates the suite's own failure.
func TestMain(m *testing.M) {
	m.Run()

	if dir := os.Getenv(statehashRunEnv); dir != "" {
		if err := writeStateHashRun(dir, lastRun); err != nil {
			fmt.Fprintf(os.Stderr, "statehash run: %v\n", err)
		}
	}
}

// TestM1ScriptedSessionDescendAndLootAreKernelVerbs pins that the walk's two
// remaining transitions are real verbs: the recorded log carries exactly one
// Descended and one Looted event, so the walk state cannot move without the
// kernel announcing it.
func TestM1ScriptedSessionDescendAndLootAreKernelVerbs(t *testing.T) {
	// given the scripted walk recorded into its tick log
	run := runScriptedWalk(t, scenarioSeed)

	// when the recorded log is read back
	log := recordedLog(t, run.sim)

	// then both transitions appear exactly once, and no stage manufactured them
	if descended := countEvents(log, kernel.KindDescended); descended != 1 {
		t.Fatalf("tick log carries %d Descended events, want 1", descended)
	}

	if looted := countEvents(log, kernel.KindLooted); looted != 1 {
		t.Fatalf("tick log carries %d Looted events, want 1", looted)
	}
}

// TestM1ScriptedSessionCreateIsAKernelVerb pins that the walk's hero comes out of
// the kernel: the recorded log carries exactly one CreatePack command and exactly
// one CreatedPack event naming the class and the spawn tile, so the harness
// cannot hand itself a warrior.
func TestM1ScriptedSessionCreateIsAKernelVerb(t *testing.T) {
	// given the scripted walk recorded into its tick log
	run := runScriptedWalk(t, scenarioSeed)

	// when the recorded log is read back
	log := recordedLog(t, run.sim)

	// then the create went in as one command
	if submitted := countCommands(log, kernel.KindCreatePack); submitted != 1 {
		t.Fatalf("tick log carries %d CreatePack commands, want 1", submitted)
	}

	// and the kernel announced the hero it raised
	assertCreatedPackEvent(t, log, warriorClass, spawnTile)
}

// TestM1ScriptedSessionSilentCreateEmitsNothing is the no-submit half of the
// silent-create pin: a session that never submits CreatePack raises no hero and
// records no event, so the empty session stays empty.
func TestM1ScriptedSessionSilentCreateEmitsNothing(t *testing.T) {
	// given a fresh session that has run one silent tick
	sim := newSessionSim(t, scenarioSeed)
	sim.descend(1)

	before := len(sim.kernel.Trace())

	// when a turn passes with no CreatePack submitted
	sim.session.Tick(sim.passed + 1)

	// then the turn emitted nothing and raised no hero
	if emitted := len(sim.kernel.Trace()) - before; emitted != 0 {
		t.Fatalf("a turn with no CreatePack emitted %d events, want 0 (no change, no event)", emitted)
	}

	if _, _, raised := sim.grid.RaisedPack(); raised {
		t.Fatal("the session raised a hero with no CreatePack command, want none")
	}

	if log := recordedLog(t, sim); countEvents(log, kernel.KindCreatedPack) != 0 {
		t.Fatalf("tick log carries %d CreatedPack events for a silent create, want 0",
			countEvents(log, kernel.KindCreatedPack))
	}
}

// TestM1ScriptedSessionSilentTicksEmitNothing is the zero-events regression pin,
// extended to the new verbs: a loot with nothing on the floor and a descend onto
// the level already loaded both change nothing, so both emit nothing.
func TestM1ScriptedSessionSilentTicksEmitNothing(t *testing.T) {
	// given a session already standing on the loaded dungeon level, having looted
	sim := newSessionSim(t, scenarioSeed)
	sim.descend(1)
	sim.clickWalk(2, sim.levels.Item())
	sim.loot(3)

	before := len(sim.kernel.Trace())

	// when the player asks for what already happened: the same level, an empty tile
	sim.descend(4)
	sim.loot(5)

	// then neither tick emitted anything
	if emitted := len(sim.kernel.Trace()) - before; emitted != 0 {
		t.Fatalf("the reload and the empty loot emitted %d events, want 0 (no change, no event)", emitted)
	}
}

// TestM1ScriptedSessionDungeonRunsTheCommittedLayout pins that the dungeon stage
// is the parity-true cathedral layout, not a synthetic grid: the walk's grid
// holds the fixture's own solid and walkable tiles, and every monster stands on a
// tile the layout itself calls walkable.
func TestM1ScriptedSessionDungeonRunsTheCommittedLayout(t *testing.T) {
	// given the scripted walk
	run := runScriptedWalk(t, scenarioSeed)

	// when the walk's grid is classified tile by tile
	solid, walkable := countGridClasses(run.sim.grid.Grid())

	// then the layout is a real dungeon and the ring came out of its floors
	if solid == 0 || walkable == 0 {
		t.Fatalf("dungeon grid holds %d solid / %d walkable tiles, want the committed layout's both", solid, walkable)
	}

	assertRingOnWalkableTiles(t, run)
}

// TestM1ScriptedSessionStateHashFoldsTheLayout pins that the level layout is part
// of the frozen state: changing one piece of the loaded grid must change the
// StateHash, so a walk on a different level cannot freeze the same hash.
func TestM1ScriptedSessionStateHashFoldsTheLayout(t *testing.T) {
	// given a walk that has descended onto the real layout
	run := runScriptedWalk(t, scenarioSeed)

	before := testsupport.StateHash(walkAdapter{state: run.sim.walk})

	// when one piece of the loaded grid changes class
	grid := run.sim.grid.Grid()
	grid.SOLData[grid.DPiece[run.sim.levels.layout.Entry.Y][run.sim.levels.layout.Entry.X]] |= world.TileSolid
	after := testsupport.StateHash(walkAdapter{state: run.sim.walk})

	// then the hash moved with it
	if before == after {
		t.Fatalf("StateHash %x survived a layout change, want the layout folded into it", before)
	}
}

// countGridClasses counts the loaded grid's tiles by classification.
func countGridClasses(grid *world.Grid) (solid, walkable int) {
	for y := range world.MAXDUNY {
		for x := range world.MAXDUNX {
			if grid.IsTileSolid(x, y) {
				solid++

				continue
			}

			walkable++
		}
	}

	return solid, walkable
}

// assertRingOnWalkableTiles pins that every monster of the kill ring stands on a
// tile the loaded layout calls walkable.
func assertRingOnWalkableTiles(t *testing.T, run *scriptedRun) {
	t.Helper()

	ring := run.sim.levels.ring
	if len(ring) != monsterCount {
		t.Fatalf("kill ring holds %d tiles, want %d", len(ring), monsterCount)
	}

	grid := run.sim.grid.Grid()

	for _, tile := range ring {
		if !grid.IsTileWalkable(tile.X, tile.Y) {
			t.Fatalf("monster tile %v is not walkable on the loaded layout, want every kill on a real floor", tile)
		}
	}
}

// runScriptedWalk drives the nine-stage walk on a fresh sim at the seed and
// freezes a StateHash at every stage.
func runScriptedWalk(t *testing.T, seed uint32) *scriptedRun {
	t.Helper()

	run := &scriptedRun{sim: newSessionSim(t, seed)}

	run.freeze(stepBoot)
	run.createWarrior()
	run.freeze(stepCreate)
	run.enterTown()
	run.freeze(stepTownEnter)
	run.descendDungeon()
	run.freeze(stepDungeonEnter)
	run.killFirst()
	run.freeze(stepKillFirst)
	run.killRemaining()
	run.freeze(stepKillLast)
	run.loot()
	run.freeze(stepLoot)
	run.save(t)
	run.freeze(stepSave)
	run.load(t)
	run.freeze(stepLoad)

	run.fold = foldRecordedTrail(t, run.sim)
	lastRun = run

	return run
}

// freeze hashes the walk's canonical state at a stage.
func (r *scriptedRun) freeze(stage string) {
	r.steps = append(r.steps, stage)
	r.hashes = append(r.hashes, testsupport.StateHash(walkAdapter{state: r.sim.walk}))
}

// createWarrior submits the create verb: the gate lets it through, the world
// slice raises the hero from the class's starting pack, and the walk's warrior
// arrives only with the KindCreatedPack event.
func (r *scriptedRun) createWarrior() {
	r.sim.createPack(r.nextTick(), warriorClass, spawnTile)
}

// enterTown clicks the spawn's neighbour tile: the walk resolves the click to a
// MouseWalk command, the world slice steps the hero, and the kernel announces it.
func (r *scriptedRun) enterTown() {
	r.sim.clickWalk(r.nextTick(), nextSpawnTile)
}

// descendDungeon submits the descend verb: the gate lets it through, the world
// slice loads the committed cathedral layout, and the walk's dungeon state moves
// only when the KindDescended event arrives.
func (r *scriptedRun) descendDungeon() {
	r.sim.descend(r.nextTick())
}

// killFirst kills the monster the ring starts with; freeze at kill#1 names that
// one kill.
func (r *scriptedRun) killFirst() {
	r.killNextMonster()
}

// killRemaining kills the other nine monsters.
func (r *scriptedRun) killRemaining() {
	for range monsterCount - 1 {
		r.killNextMonster()
	}
}

// killNextMonster clicks the first live monster until a swing lands: the landed
// combat pipeline rolls a hit percent, so a kill costs one or more attack ticks
// and every tick is a recorded pass.
func (r *scriptedRun) killNextMonster() {
	tiles := r.sim.walk.monsterTiles()
	if len(tiles) == 0 {
		return
	}

	target := tiles[0]

	for range attackAttemptLimit {
		r.sim.clickAttack(r.nextTick(), target)

		if !r.sim.walk.monsters[target] {
			return
		}
	}

	r.sim.t.Fatalf("monster on tile %v survived %d attack ticks", target, attackAttemptLimit)
}

// loot steps onto the level's floor item and then submits the loot verb: the
// world slice takes the item off the player's own tile and announces it, and the
// walk puts the item into the warrior's pack only from that event.
func (r *scriptedRun) loot() {
	r.sim.clickWalk(r.nextTick(), r.sim.levels.Item())
	r.sim.loot(r.nextTick())
}

// save writes the session blocks through the store and stamps the survived-save
// flag the canonical hash carries.
func (r *scriptedRun) save(t *testing.T) {
	t.Helper()

	r.sim.saveSession(context.Background(), t)
	r.sim.walk.saved = true
}

// load reopens the store file and loads every block back.
func (r *scriptedRun) load(t *testing.T) {
	t.Helper()

	r.sim.loadSession(context.Background(), t)
}

// nextTick counts the tape forward one tick per decision.
func (r *scriptedRun) nextTick() uint32 {
	return r.sim.passed + 1
}

// foldRecordedTrail flushes the writer and folds the recorded checksum trail.
func foldRecordedTrail(t *testing.T, sim *sessionSim) uint64 {
	t.Helper()

	trail, err := recordedTrail(context.Background(), sim)
	if err != nil {
		t.Fatalf("recorded trail: %v", err)
	}

	return persist.FoldChecksums(trail)
}

// recordedTrail flushes the tick-log writer and projects the stored log into the
// checksum trail the fold consumes.
func recordedTrail(ctx context.Context, sim *sessionSim) ([]persist.TickChecksum, error) {
	if err := sim.writer.Flush(ctx); err != nil {
		return nil, fmt.Errorf("flush tick log: %w", err)
	}

	log, err := sim.store.LoadTickLog(ctx)
	if err != nil {
		return nil, fmt.Errorf("load tick log: %w", err)
	}

	trail := make([]persist.TickChecksum, 0, len(log))
	for _, record := range log {
		trail = append(trail, persist.TickChecksum{Tick: record.Tick, Hash: record.Checksum})
	}

	return trail, nil
}

// replayTrail drains the run's log through a fresh kernel of the same handlers,
// the same seed and the same world slice.
func replayTrail(t *testing.T, run *scriptedRun) (persist.ReplayResult, error) {
	t.Helper()

	ctx := context.Background()

	if err := run.sim.writer.Flush(ctx); err != nil {
		return persist.ReplayResult{}, fmt.Errorf("flush tick log: %w", err)
	}

	fresh := run.sim.freshKernel()

	result, err := persist.Replay(ctx, run.sim.store, fresh, nil)
	if err != nil {
		return persist.ReplayResult{}, fmt.Errorf("replay trail: %w", err)
	}

	return result, nil
}

// writeStateHashRun writes the run's per-stage hash list under dir. A write
// failure is reported, not dropped: a missing statehash-run.json otherwise
// surfaces only as a comparator gap far from its cause.
func writeStateHashRun(dir string, run *scriptedRun) error {
	if run == nil {
		return nil
	}

	encoded, err := json.MarshalIndent(statehashList(run), "", "  ")
	if err != nil {
		return fmt.Errorf("encode statehash run: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "statehash-run.json"), encoded, 0o644); err != nil {
		return fmt.Errorf("write statehash run: %w", err)
	}

	return nil
}

// statehashList projects the run's per-stage hashes.
func statehashList(run *scriptedRun) []stepHash {
	stages := make([]stepHash, 0, len(run.steps))
	for i, step := range run.steps {
		stages = append(stages, stepHash{Step: step, HashHex: fmt.Sprintf("%x", run.hashes[i])})
	}

	return stages
}

// recordedLog flushes the sim's tick-log writer and reads the recorded log back.
func recordedLog(t *testing.T, sim *sessionSim) []persist.TickRecord {
	t.Helper()

	ctx := context.Background()
	if err := sim.writer.Flush(ctx); err != nil {
		t.Fatalf("flush tick log: %v", err)
	}

	log, err := sim.store.LoadTickLog(ctx)
	if err != nil {
		t.Fatalf("LoadTickLog: %v", err)
	}

	return log
}

// countEvents counts the events of one kind across a recorded tick log.
func countEvents(log []persist.TickRecord, kind kernel.Kind) int {
	return len(eventsOfKind(log, kind))
}

// countCommands counts the commands of one kind across a recorded tick log.
func countCommands(log []persist.TickRecord, kind kernel.Kind) int {
	count := 0
	for _, record := range log {
		for _, command := range record.Commands {
			if command.Kind == kind {
				count++
			}
		}
	}

	return count
}

// eventsOfKind collects the events of one kind across a recorded tick log, in
// recording order.
func eventsOfKind(log []persist.TickRecord, kind kernel.Kind) []kernel.Event {
	events := make([]kernel.Event, 0, len(log))
	for _, record := range log {
		for _, event := range record.Events {
			if event.Kind == kind {
				events = append(events, event)
			}
		}
	}

	return events
}

// assertCreatedPackEvent pins the create event count and payload.
func assertCreatedPackEvent(t *testing.T, log []persist.TickRecord, class int, tile world.Tile) {
	t.Helper()

	events := eventsOfKind(log, kernel.KindCreatedPack)
	if len(events) != 1 {
		t.Fatalf("tick log carries %d CreatedPack events, want 1", len(events))
	}

	event := events[0]
	if event.A != int32(class) || event.B != int32(tile.X) || event.C != int32(tile.Y) {
		t.Fatalf("CreatedPack payload = (%d,%d,%d), want (%d,%d,%d)",
			event.A, event.B, event.C, class, tile.X, tile.Y)
	}
}
