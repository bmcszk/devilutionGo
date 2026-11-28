package monsterspawn_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/monsterspawn"
)

func TestPlaceGroupScriptedTapes(t *testing.T) {
	t.Parallel()

	// given the committed scripted tapes, each the exact roll sequence of one
	// hand-derived call
	cases := readTapes(t)

	if len(cases) == 0 {
		t.Fatal("committed tape fixture is empty")
	}

	// when every case runs on its scripted source
	placements := runTapes(t, cases)

	// then each placement matches its row
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assertTapeCase(t, tc, placements[i])
		})
	}
}

func TestPlaceGroupScatterKeepsTheVanillaDeltaXBug(t *testing.T) {
	t.Parallel()

	// given a pack whose scatter direction is West, where the two deltas differ
	// (deltaX -1, deltaY +1)
	req := monsterspawn.GroupRequest{Num: 2, Total: 200, Base: monsterspawn.Stats{HitPoints: 10, MaxHitPoints: 10}}

	// when it runs on a tape whose scatter rolls West twice
	placement := monsterspawn.PlaceGroup(openWorld(), scriptedRnd(t, []int32{0, 1, 2, 2, 2, 3, 2, 2}), req)

	// then the second monster sits one tile up-left, not one up-right: the
	// vanilla step adds the X delta to BOTH axes (monster.cpp:341). A build
	// that "fixes" it to deltaY lands on (15,18) and fails here.
	if len(placement.Tiles) != 2 {
		t.Fatalf("placed = %d monsters, want 2", len(placement.Tiles))
	}

	if got := placement.Tiles[1]; got.X != 15 || got.Y != 16 {
		t.Fatalf("second monster at (%d,%d), want (15,16): the yp += deltaX bug", got.X, got.Y)
	}
}

func TestPlaceGroupBudgetClampTruncates(t *testing.T) {
	t.Parallel()

	// given a level holding 199 of its 200 monsters and a pack of five
	req := monsterspawn.GroupRequest{Num: 5, Active: 199, Total: 200, Base: monsterspawn.Stats{}}

	// when the pack is placed
	placement := monsterspawn.PlaceGroup(openWorld(), scriptedRnd(t, []int32{0, 1, 2, 0, 0}), req)

	// then the clamp truncated it to total - active = 1 monster
	if len(placement.Tiles) != 1 {
		t.Fatalf("placed = %d monsters, want 1 (totalmonsters - ActiveMonsterCount)", len(placement.Tiles))
	}

	if placement.Rolls != 5 {
		t.Fatalf("rolls = %d, want 5", placement.Rolls)
	}
}

func TestPlaceGroupBudgetBoundaryPlacesNothing(t *testing.T) {
	t.Parallel()

	// given a level already at its monster budget
	req := monsterspawn.GroupRequest{Num: 5, Active: 200, Total: 200, Base: monsterspawn.Stats{}}

	// when the pack is placed
	placement := monsterspawn.PlaceGroup(openWorld(), scriptedRnd(t, []int32{0, 1}), req)

	// then nothing is placed and the anchor roll still happened: the upstream
	// size_t wrap at total < active is NOT reproduced, so no panic either
	if len(placement.Tiles) != 0 {
		t.Fatalf("placed = %d monsters, want 0", len(placement.Tiles))
	}

	if placement.Rolls != 2 {
		t.Fatalf("rolls = %d, want the 2 anchor rolls", placement.Rolls)
	}
}

func TestPlaceGroupLeashRejectsAtRadiusFour(t *testing.T) {
	t.Parallel()

	// given a leashed pack whose anchor is blocked, and a world where the only
	// other placeable tile sits four steps away — exactly on the leash limit
	world := tileWorld([][2]int{{17, 18}, {21, 22}})
	req := monsterspawn.GroupRequest{
		Num:     1,
		Total:   200,
		Base:    monsterspawn.Stats{MaxHitPoints: 10, HitPoints: 10, Gargoyle: true},
		Leader:  &monsterspawn.Leader{X: 16, Y: 17, Intelligence: 7},
		Leashed: true,
	}

	// when the scatter walks four steps out and four steps back
	script := []int32{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 4, 4, 4, 4, 4, 4, 4, 2, 2, 2}
	placement := monsterspawn.PlaceGroup(world, scriptedRnd(t, script), req)

	// then the tile on the limit was refused (its leash delta is 4) and the
	// walk came back to the in-leash tile
	if len(placement.Tiles) != 1 {
		t.Fatalf("placed = %d monsters, want 1", len(placement.Tiles))
	}

	if got := placement.Tiles[0]; got.X != 17 || got.Y != 18 {
		t.Fatalf("placed at (%d,%d), want the in-leash (17,18)", got.X, got.Y)
	}
}

func TestPlaceGroupRewindRetriesTheWholeAttempt(t *testing.T) {
	t.Parallel()

	// given a world with room for exactly one monster (the 16,16 anchor tile)
	// where the pack asks for two: every attempt places one and then exhausts
	// its scatter budget
	world := tileWorld([][2]int{{16, 16}})
	req := monsterspawn.GroupRequest{Num: 2, Total: 200, Base: monsterspawn.Stats{}}

	// when the call runs out of retries
	placement := monsterspawn.PlaceGroup(world, constRnd(0), req)

	// then all ten attempts ran, each drawing a 2-roll anchor, one direction
	// roll and 101 scatter steps of 2 rolls (monster.cpp:312-341)
	if len(placement.Tiles) != 1 {
		t.Fatalf("placed = %d monsters, want 1", len(placement.Tiles))
	}

	if placement.Rolls != 2050 {
		t.Fatalf("rolls = %d, want 10 attempts x 205 rolls", placement.Rolls)
	}

	// ...and the nine retries unplaced the monster the previous attempt left
	if placement.Rewinds != 9 {
		t.Fatalf("rewinds = %d, want 9", placement.Rewinds)
	}
}

func TestIsMonsterAvailableOverCommittedTable(t *testing.T) {
	t.Parallel()

	// given every committed monstdat row (base and Hellfire tables)
	rows := committedMonsters(t)

	// when each row's rule is gated in both modes at three levels
	mismatches := gateMismatches(t, rows, []int{1, 5, 16})

	// then no row disagrees with its availability class and level band
	if len(mismatches) > 0 {
		t.Fatalf("gate mismatches:\n%s", strings.Join(mismatches, "\n"))
	}
}

func TestIsMonsterAvailableSharewareExcludesRetail(t *testing.T) {
	t.Parallel()

	// given every committed row with its availability class
	rows := committedMonsters(t)

	// when the gate runs for retail-class monsters in shareware
	passed := gatePasses(t, rows, monsterspawn.AvailabilityRetail, monsterspawn.Shareware)

	// then none of them pass at their own minimum level
	if len(passed) > 0 {
		t.Fatalf("retail monsters passed the shareware gate: %v", passed)
	}
}

func TestIsMonsterAvailableNeverClassNeverSpawns(t *testing.T) {
	t.Parallel()

	// given every committed row with its availability class
	rows := committedMonsters(t)

	// when the gate runs for never-class monsters in a retail game
	passed := gatePasses(t, rows, monsterspawn.AvailabilityNever, monsterspawn.Retail)

	// then none of them pass at their own minimum level
	if len(passed) > 0 {
		t.Fatalf("never-class monsters passed the retail gate: %v", passed)
	}
}
