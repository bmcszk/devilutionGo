// The M1 seed-parity gate (devilGo-451a, AC-2): the scripted descend/loot walk
// runs on every committed DRLG seed, each on its own seed's dungeon, and the
// per-seed nine-stage StateHash trail is frozen as the parity expectation.
// tools/probe/seeds.txt (devilGo-vsx3) is the provenance of the seed list: it is
// the one list the fixture generator ran with, so every seed here carries a
// committed l1-main and l1-dpiece corpus under internal/testdata/drlg.
package integration_test

import (
	"fmt"
	"slices"
	"testing"

	dtest "github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// committedSeeds is the AC-2 corpus, in the order tools/probe/seeds.txt pins it.
var committedSeeds = []uint32{
	2588, 743271966, 1383137027, 844660068, 19770182, 1522546307, 125121312, 1511478689,
}

// pinnedSeedTrails is the AC-2 expectation table: the walk's own values, measured
// on the committed corpus and then frozen. The three stages before the descend
// carry one value for every seed — no draw is taken and no level is loaded yet —
// and everything from dungeon-enter on is per seed, because the seed picks the
// layout and every draw after it.
var pinnedSeedTrails = []pinnedTrail{
	{2588, 0xa649c25d04554b50, []uint64{
		0x4e645f013c56620d, 0x4f9b29731d9af0d5, 0x3bb0f5f6ee811ecf,
		0xb43a9dbddea642e5, 0x5ee08a006f1d7554, 0x2238a0e0bba6b242,
		0x5c24612269a886b9, 0x9969b049a43b11a6, 0x16a7958ec7ea644f,
	}},
	{743271966, 0x710f84bad15a4b10, []uint64{
		0x4e645f013c56620d, 0x4f9b29731d9af0d5, 0x3bb0f5f6ee811ecf,
		0xa771dcc3c0b653b1, 0xe69ecfd5fd8b6882, 0x7d682e53d5b7bf8a,
		0x8293059efde4b889, 0xf2b32e7056cc43fe, 0xf7aad8777b3ba0ab,
	}},
	{1383137027, 0xda9eae7046a7f2d3, []uint64{
		0x4e645f013c56620d, 0x4f9b29731d9af0d5, 0x3bb0f5f6ee811ecf,
		0xe6b1098ea24a3423, 0x4f34bc9ae6b7b398, 0x335fd1f951b304c8,
		0xa9bc15da3a0e0cd3, 0xf97ea96c5ec95954, 0x8a98e7169d7ec965,
	}},
	{844660068, 0xb525c24896866ee6, []uint64{
		0x4e645f013c56620d, 0x4f9b29731d9af0d5, 0x3bb0f5f6ee811ecf,
		0x7155a223adc04b77, 0x5837a00607e540ec, 0x4d890d4b673b9cf5,
		0x39ed72e47b78b69a, 0x76bb04bbf427a549, 0x66d1b3d4a1ea598c,
	}},
	{19770182, 0x9bb8a566daca26be, []uint64{
		0x4e645f013c56620d, 0x4f9b29731d9af0d5, 0x3bb0f5f6ee811ecf,
		0x503afb549995327a, 0x6baff2731a4b01fd, 0x75b0b3a51d6c660e,
		0xc57b446b814d4545, 0x09ee665adb5393be, 0xa1a2e0ed36148977,
	}},
	{1522546307, 0x7c486faf98d1de46, []uint64{
		0x4e645f013c56620d, 0x4f9b29731d9af0d5, 0x3bb0f5f6ee811ecf,
		0x3c6f8f9e296448fd, 0x2e2a35899bf2468e, 0xc1ed3aeb6b347cc3,
		0x2a3fcad2e8137188, 0xd33c42a00f52bd8f, 0x15724e510558c0d6,
	}},
	{125121312, 0x963b8e0053535830, []uint64{
		0x4e645f013c56620d, 0x4f9b29731d9af0d5, 0x3bb0f5f6ee811ecf,
		0xe0b1cdc2a221523f, 0xd610afb78a07ba5f, 0x064074e5a78b5e2f,
		0x12e740f4ed544f08, 0x1e394d661fef0fcf, 0x2ede104d3d2fcafe,
	}},
	{1511478689, 0x66a9f3162c8829e0, []uint64{
		0x4e645f013c56620d, 0x4f9b29731d9af0d5, 0x3bb0f5f6ee811ecf,
		0xf33599d3cc580321, 0xc3bbf94988e32d42, 0xa4a06e01b626a599,
		0xcd835d05f41943f2, 0x3104d29cf97230bd, 0x468d357f7dca210c,
	}},
}

// pinnedTrail is one seed's frozen walk result: the StageHash of each of the nine
// stages in stepOrder, and the fold of the walk's recorded tick-log trail.
type pinnedTrail struct {
	seed   uint32
	fold   uint64
	stages []uint64
}

// TestM1SeedParityWalksEveryCommittedSeed pins that each committed seed's walk
// reaches all nine stages and does it on the dungeon that seed generates: the
// descended grid's layout digest equals the digest the seed's committed fixture
// yields. The fixture is the independent oracle, so a walk that loads one seed's
// dungeon eight times fails seven cases by name.
func TestM1SeedParityWalksEveryCommittedSeed(t *testing.T) {
	for _, seed := range committedSeeds {
		t.Run(seedSubtestName(seed), func(t *testing.T) {
			// given the scripted walk on a fresh sim at this seed
			run := runScriptedWalk(t, seed)

			// when the walk's frozen stages and its loaded grid are read back
			steps, digest := run.steps, run.sim.grid.Grid().LayoutDigest()

			// then the walk completed every stage of the contract
			assertWalkReachedEveryStage(t, steps)

			// and the dungeon stage ran this seed's own committed layout
			assertWalkRunsSeedLayout(t, seed, digest)
		})
	}
}

// TestM1SeedParityFreezesThePerSeedTrails pins the AC-2 expectation itself: each
// seed's nine stage hashes and its recorded trail fold equal the frozen values.
// The values differ per seed from the descend on, which is the point — one pin
// per seed, so a drifted layout, kill order or combat draw names its seed.
func TestM1SeedParityFreezesThePerSeedTrails(t *testing.T) {
	for _, pinned := range pinnedSeedTrails {
		t.Run(seedSubtestName(pinned.seed), func(t *testing.T) {
			// given the scripted walk on a fresh sim at the pinned seed
			run := runScriptedWalk(t, pinned.seed)

			// when the walk's frozen trail is read back
			steps, hashes, fold := run.steps, run.hashes, run.fold

			// then every stage froze the pinned hash
			assertPinnedStages(t, pinned, steps, hashes)

			// and the recorded trail folded to the pinned word
			if fold != pinned.fold {
				t.Fatalf("seed %d tick-log fold = %016x, want the pinned %016x",
					pinned.seed, fold, pinned.fold)
			}

			t.Logf("seed %d fold %016x stages %v", pinned.seed, fold, hexStages(hashes))
		})
	}
}

// TestM1SeedParityTrailsDifferPerSeed pins the non-vacuity of the table: the eight
// pinned walks must not be one dungeon under eight names. The tables are read, not
// re-walked, because the values are the fixture.
func TestM1SeedParityTrailsDifferPerSeed(t *testing.T) {
	// given the pinned table
	if len(pinnedSeedTrails) != len(committedSeeds) {
		t.Fatalf("pinned table holds %d seeds, want the corpus's %d",
			len(pinnedSeedTrails), len(committedSeeds))
	}

	// when the descended stages are compared seed by seed
	seen := make(map[uint64]uint32, len(pinnedSeedTrails))

	// then no two seeds share a dungeon-enter hash
	for _, pinned := range pinnedSeedTrails {
		entry := pinned.stages[stepIndex(stepDungeonEnter)]
		if prior, repeated := seen[entry]; repeated {
			t.Fatalf("seeds %d and %d froze the same dungeon-enter hash %016x, want each seed's own",
				prior, pinned.seed, entry)
		}

		seen[entry] = pinned.seed
	}
}

// assertWalkReachedEveryStage pins that the walk froze the contract's nine stages
// in order.
func assertWalkReachedEveryStage(t *testing.T, steps []string) {
	t.Helper()

	if len(steps) != len(stepOrder) {
		t.Fatalf("walk froze %d stages, want %d", len(steps), len(stepOrder))
	}

	for i, step := range stepOrder {
		if steps[i] != step {
			t.Fatalf("walk stage %d = %q, want %q", i, steps[i], step)
		}
	}
}

// assertWalkRunsSeedLayout pins that the walk's dungeon grid is the layout the
// seed's committed fixture yields: the fixture loads onto a fresh state through
// the same LoadLevel the descend verb runs, then the two digests are compared.
func assertWalkRunsSeedLayout(t *testing.T, seed uint32, digest uint64) {
	t.Helper()

	if want := fixtureLayoutDigest(t, seed); digest != want {
		t.Fatalf("seed %d walk layout = %016x, want its own committed fixture's %016x", seed, digest, want)
	}
}

// assertPinnedStages pins the walk's stage hashes against one pinned trail.
func assertPinnedStages(t *testing.T, pinned pinnedTrail, steps []string, hashes []uint64) {
	t.Helper()

	if len(pinned.stages) != len(stepOrder) {
		t.Fatalf("seed %d pins %d stages, want %d", pinned.seed, len(pinned.stages), len(stepOrder))
	}

	if !slices.Equal(steps, stepOrder) {
		t.Fatalf("seed %d walked stages %v, want %v", pinned.seed, steps, stepOrder)
	}

	for i, want := range pinned.stages {
		if got := hashes[i]; got != want {
			t.Fatalf("seed %d stage %s StateHash = %016x, want the pinned %016x",
				pinned.seed, stepOrder[i], got, want)
		}
	}
}

// fixtureLayoutDigest folds one seed's committed cathedral layout into the digest
// the walk's grid must carry.
func fixtureLayoutDigest(t *testing.T, seed uint32) uint64 {
	t.Helper()

	layout := dtest.LoadL1Layout(t, dungeonLevel, seed)

	state := world.NewState()
	if _, ok := state.LoadLevel(layout.Image, layout.Entry, dungeonLevel, dungeonLevelType); !ok {
		t.Fatalf("seed %d fixture layout does not load: no spawn near %v", seed, layout.Entry)
	}

	return state.Grid().LayoutDigest()
}

// hexStages renders the walk's stage hashes for the test log.
func hexStages(hashes []uint64) []string {
	rendered := make([]string, len(hashes))
	for i, hash := range hashes {
		rendered[i] = fmt.Sprintf("%016x", hash)
	}

	return rendered
}

// stepIndex reports a stage's position in stepOrder.
func stepIndex(step string) int {
	for i, name := range stepOrder {
		if name == step {
			return i
		}
	}

	return -1
}

// seedSubtestName is one seed's subtest name: the seed in decimal, so a failing
// case names itself in the test output matrix.
func seedSubtestName(seed uint32) string {
	return fmt.Sprintf("seed-%d", seed)
}
