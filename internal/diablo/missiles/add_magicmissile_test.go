// Package missiles_test pins AddGenericMagicMissile — the BloodStar row's
// addFn (misdat.tsv row 24) — from missiles.cpp:2285-2325: the fixed
// velocity 16, the succubus-family animation override, the 16-direction
// frame group and the player damage formula.
package missiles_test

import (
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

const monstdatPath = "../../../internal/testdata/gamedata/monstdat/monstdat.tsv"

// addMagicMissileCases holds the hand-computed AddGenericMagicMissile rows.
var addMagicMissileCases = []addMagicMissileCase{
	{
		name: "player damage formula", script: addScript{
			dst: [2]int{10, 10}, caster: 0, source: 0, spllvl: 2, magic: 16, light: 42,
		},
		wantDam: 12, wantAnimType: 22,
	},
	{
		name: "succubus overrides the animation", script: addScript{
			dst: [2]int{10, 10}, caster: 1, source: 3, monsterType: 101, monsterMin: 2,
			monsterMax: 5, light: 42, rnd: []int{2},
		},
		wantDam: 4, wantAnimType: 22,
	},
	{
		name: "snow witch uses the blue sheet", script: addScript{
			dst: [2]int{10, 10}, caster: 1, source: 3, monsterType: 102, monsterMin: 2,
			monsterMax: 5, light: 42, rnd: []int{0},
		},
		wantDam: 2, wantAnimType: 40,
	},
	{
		name: "hell spawn uses the red sheet", script: addScript{
			dst: [2]int{10, 10}, caster: 1, source: 3, monsterType: 103, monsterMin: 2,
			monsterMax: 5, light: 42, rnd: []int{3},
		},
		wantDam: 5, wantAnimType: 44,
	},
	{
		name: "trap rolls the trap dice", script: addScript{
			dst: [2]int{10, 10}, caster: 0, source: -1, currlevel: 5, light: 42, rnd: []int{3},
		},
		wantDam: 8, wantAnimType: 22,
	},
	{
		name: "same tile shot shifts by midir", script: addScript{
			start: [2]int{5, 5}, dst: [2]int{5, 5}, midir: missiles.South, caster: 0, source: 0,
			spllvl: 0, magic: 0, light: 42,
		},
		wantDam: 0, wantAnimType: 22,
	},
}

// addMagicMissileCase is one AddGenericMagicMissile row.
type addMagicMissileCase struct {
	name         string
	script       addScript
	wantDam      int
	wantAnimType int
}

func TestAddGenericMagicMissile(t *testing.T) {
	for _, tt := range addMagicMissileCases {
		t.Run(tt.name, func(t *testing.T) {
			// given
			fx := newAddFixture(t, tt.script)
			w, m, p := fx.world, fx.missile, fx.param
			// AddMissile seeds _miAnimType from the misdat graphic
			// (missiles.cpp:2824); the BloodStar row's is BloodStar.
			m.Type = missiles.BloodStar
			m.AnimType = int(missiles.GraphicBloodStar)

			// when
			missiles.AddGenericMagicMissile(w, m, p)
			// then
			fx.tape.assertConsumed()

			assertMagicMissileState(t, m, tt)
		})
	}
}

// TestMonsterTypeConsts pins the four BloodStar monster-type ordinals to the
// committed monstdat.tsv row order (enum _monster_id, tables/monstdat.h:262-265).
func TestMonsterTypeConsts(t *testing.T) {
	// given
	rows := monsterRowIndexes(t)

	// when
	got := map[string]int{
		"MT_SUCCUBUS": rows["MT_SUCCUBUS"],
		"MT_SNOWWICH": rows["MT_SNOWWICH"],
		"MT_HLSPWN":   rows["MT_HLSPWN"],
		"MT_SOLBRNR":  rows["MT_SOLBRNR"],
	}

	// then
	for name, want := range map[string]int{
		"MT_SUCCUBUS": missiles.MonsterTypeSuccubus,
		"MT_SNOWWICH": missiles.MonsterTypeSnowWich,
		"MT_HLSPWN":   missiles.MonsterTypeHellSpawn,
		"MT_SOLBRNR":  missiles.MonsterTypeSolbrner,
	} {
		if got, ok := got[name]; !ok || got != want || got == 0 {
			t.Fatalf("%s ordinal = %d (found=%v), want %d", name, got, ok, want)
		}
	}
}

// monsterRowIndexes maps each _monster_id column value to its data row index,
// which is the enum ordinal (monstdat.cpp GetMonsterData indexes by ordinal).
func monsterRowIndexes(t *testing.T) map[string]int {
	t.Helper()

	data, err := os.ReadFile(monstdatPath)
	if err != nil {
		t.Fatalf("read monstdat fixture: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	rows := make(map[string]int, len(lines)-1)

	for i, line := range lines[1:] {
		rows[strings.SplitN(line, "\t", 2)[0]] = i
	}

	return rows
}

// assertMagicMissileState checks the post-AddGenericMagicMissile motion,
// animation sheet, light and damage fields (missiles.cpp:2290-2325).
func assertMagicMissileState(t *testing.T, m *missiles.Missile, tt addMagicMissileCase) {
	t.Helper()

	velocity := m.Position.Velocity
	if velocity.DeltaX != 0 || velocity.DeltaY != 524288 { // cpp:2290, speed 16
		t.Fatalf("velocity = (%d, %d), want (0, 524288)", velocity.DeltaX, velocity.DeltaY)
	}

	if m.AnimType != tt.wantAnimType { // cpp:2295-2306, misdat graphic default
		t.Fatalf("animType = %d, want %d", m.AnimType, tt.wantAnimType)
	}

	if m.MFNum != 0 { // cpp:2308-2310, dst (10,10) is South
		t.Fatalf("frame group = %d, want 0", m.MFNum)
	}

	if m.Duration != 256 { // cpp:2291
		t.Fatalf("duration = %d, want 256", m.Duration)
	}

	if m.Lid != 42 { // cpp:2294
		t.Fatalf("lid = %d, want 42", m.Lid)
	}

	if m.Dam != tt.wantDam { // cpp:2312-2325
		t.Fatalf("dam = %d, want %d", m.Dam, tt.wantDam)
	}
}
