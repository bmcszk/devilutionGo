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
		writeStateHashRun(dir, lastRun)
	}
}

// countEvents counts the events of one kind across a recorded tick log.
func countEvents(log []persist.TickRecord, kind kernel.Kind) int {
	count := 0
	for _, record := range log {
		for _, event := range record.Events {
			if event.Kind == kind {
				count++
			}
		}
	}

	return count
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

// createWarrior raises the warrior from the walk's starting pack.
func (r *scriptedRun) createWarrior() {
	r.sim.walk.warrior = warriorPack(spawnTile)
}

// enterTown clicks the spawn's neighbour tile: the walk resolves the click to a
// MouseWalk command, the world slice steps the hero, and the kernel announces it.
func (r *scriptedRun) enterTown() {
	r.sim.clickWalk(r.nextTick(), nextSpawnTile)
}

// descendDungeon clicks the dungeon pedestal, then moves the session onto
// Cathedral L1 and seeds the monster ring. The level change is a level-load, not
// a movement command: the landed surfaces carry no descend verb yet.
func (r *scriptedRun) descendDungeon() {
	r.sim.clickWalk(r.nextTick(), dungeonTile)
	r.sim.walk.descend()
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

// loot walks onto the loot tile and takes the drop into the warrior's pack.
func (r *scriptedRun) loot() {
	r.sim.clickWalk(r.nextTick(), lootTile)

	warrior := &r.sim.walk.warrior
	warrior.NumInv = 1
	warrior.InvList[0] = persist.ItemPack{Idx: 42, Seed: 1234}
	warrior.InvGrid[0] = 1
	warrior.Gold += 100
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

// writeStateHashRun writes the run's per-stage hash list under dir.
func writeStateHashRun(dir string, run *scriptedRun) {
	if run == nil {
		return
	}

	encoded, err := json.MarshalIndent(statehashList(run), "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(filepath.Join(dir, "statehash-run.json"), encoded, 0o644)
}

// statehashList projects the run's per-stage hashes.
func statehashList(run *scriptedRun) []stepHash {
	stages := make([]stepHash, 0, len(run.steps))
	for i, step := range run.steps {
		stages = append(stages, stepHash{Step: step, HashHex: fmt.Sprintf("%x", run.hashes[i])})
	}

	return stages
}
