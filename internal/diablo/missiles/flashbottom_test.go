// Package missiles_test pins AddFlashBottom (missiles.cpp:2118-2136) and
// ProcessFlashBottom (cpp:3420-3448).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// flashBottomAddCases pin the three caster branches: a player cast rolls
// GenerateRndSum(20, level + 1) plus level + 1, scaled and halved
// (cpp:2121-2127); a monster cast takes the monster's level for the current
// difficulty times 2 — base 3 + 15 for nightmare = 18, doubled to 36
// (monster.h:409-413, missiles.cpp:2129-2130); a trap takes currlevel / 2.
var flashBottomCases = []flashBottomCase{
	{
		name: "monster cast shields and probes six tiles", duration: 3,
		caster: 0, source: 0, wantDuration: 2, wantShield: 1, wantProbes: 6,
	},
	{
		name: "player cast never touches the shield", duration: 3,
		caster: 1, source: 0, wantDuration: 2, wantProbes: 6,
	},
	{
		name: "last tick drops the shield", duration: 1, caster: 0, source: 0,
		wantShield: 1, wantProbes: 6, wantDel: true, wantUnshield: true,
	},
}

var flashBottomAddCases = []flashBottomAddCase{
	{
		name: "player cast rolls level plus one dice", caster: 0, source: 0, level: 2,
		rnd: []int{4, 5, 6}, wantDam: 27,
	},
	{
		name:   "monster cast doubles the level for the difficulty",
		caster: 2, source: 1, monsterLevel: 3, difficulty: 1, wantDam: 36,
	},
	{
		name: "trap takes half the level", caster: 2, source: -1, currlevel: 9, wantDam: 4,
	},
}

// flashBottomAddCase pins one AddFlashBottom row.
type flashBottomAddCase struct {
	name         string
	caster       int
	source       int
	level        int
	monsterLevel int
	difficulty   int
	currlevel    int
	rnd          []int
	wantDam      int
}

// flashBottomCases pin the ticks: the caster is shielded for the flash, six
// offsets are probed (NorthWest, None, SouthEast, West, SouthWest, South,
// cpp:3430-3438), and the last tick deletes and drops the shield.
// flashBottomCase pins one ProcessFlashBottom tick.
type flashBottomCase struct {
	name         string
	duration     int
	caster       int
	source       int
	wantDuration int
	wantShield   int
	wantProbes   int
	wantUnshield bool
	wantDel      bool
}

// flashBottomSpies records the invincibility writes of one tick.
type flashBottomSpies struct {
	sets   int
	unsets int
}

// TestAddFlashBottom pins the three damage branches and the 19-tick flare.
func TestAddFlashBottom(t *testing.T) {
	// given
	cases := flashBottomAddCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			tape := &rndTape{t: t, tape: tc.rnd}
			w := flashBottomWorld(&flashBottomSpies{}, tape, tc)
			m := missiles.Missile{Type: missiles.FlashBottom, Caster: tc.caster, Source: tc.source}

			// when
			missiles.AddFlashBottom(w, &m, &missiles.AddParameter{})
			tape.assertConsumed()

			// then
			assertFlashBottomAdd(t, m, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessFlashBottom pins the shield bracket, the six-tile probe sweep and
// the delete tail.
func TestProcessFlashBottom(t *testing.T) {
	// given
	cases := flashBottomCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			spies := &flashBottomSpies{}
			w := flashBottomWorld(spies, &rndTape{t: t}, flashBottomAddCase{})
			probes := flashBottomProbes(w)

			// when
			m := missiles.Missile{
				Type: missiles.FlashBottom, Duration: tc.duration, Caster: tc.caster,
				Source: tc.source, Dam: 12,
			}
			m.SetTraveledFromTile(4, 4)
			missiles.ProcessFlashBottom(w, &m)

			// then
			assertFlashBottomTick(t, m, spies, probes(), tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// assertFlashBottomAdd checks the branch damage and the 19-tick flare.
func assertFlashBottomAdd(t *testing.T, m missiles.Missile, tc flashBottomAddCase) {
	t.Helper()

	if m.Dam != tc.wantDam { // cpp:2123-2132
		t.Fatalf("Dam = %d, want %d", m.Dam, tc.wantDam)
	}

	if m.Duration != 19 { // cpp:2135
		t.Fatalf("Duration = %d, want 19", m.Duration)
	}
}

// assertFlashBottomTick checks the state, the probe count and the shield
// writes of one tick.
func assertFlashBottomTick(t *testing.T, m missiles.Missile, spies *flashBottomSpies,
	probes int, tc flashBottomCase,
) {
	t.Helper()

	if m.Duration != tc.wantDuration || m.DelFlag != tc.wantDel {
		t.Fatalf("duration/del = %d/%v, want %d/%v",
			m.Duration, m.DelFlag, tc.wantDuration, tc.wantDel)
	}

	if probes != tc.wantProbes {
		t.Fatalf("collision probes = %d, want %d", probes, tc.wantProbes)
	}

	if spies.sets != tc.wantShield || (spies.unsets != 0) != tc.wantUnshield {
		t.Fatalf("shield sets/unsets = %d/%d, want %d/unshield %v",
			spies.sets, spies.unsets, tc.wantShield, tc.wantUnshield)
	}
}

// flashBottomWorld wires the shield sink, the draw tape and the stat reads.
func flashBottomWorld(spies *flashBottomSpies, tape *rndTape, tc flashBottomAddCase) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.Currlevel = tc.currlevel
	cb.Difficulty = tc.difficulty
	cb.SetInvincible = func(_ int, invincible bool) {
		if invincible {
			spies.sets++
		} else {
			spies.unsets++
		}
	}

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{Level: tc.level}
			},
			MonsterStats: func(int) missiles.AddMonsterStats {
				return missiles.AddMonsterStats{Level: tc.monsterLevel}
			},
		},
	}
}

// flashBottomProbes counts the collision probes of one tick.
func flashBottomProbes(w *missiles.HitWorld) func() int {
	probes := 0
	for _, tile := range [][2]int{{3, 4}, {4, 4}, {5, 4}, {3, 5}, {4, 5}, {5, 5}} {
		w.DMonster[tile[1]][tile[0]] = 1
	}
	w.HitCallbacks.MonsterTargeted = func(int16, *missiles.Direction) bool {
		probes++

		return false
	}

	return func() int { return probes }
}
