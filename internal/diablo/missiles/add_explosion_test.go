// Package missiles_test pins AddMissileExplosion (missiles.cpp:2020-2047).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// explosionCases holds the hand-computed rows of AddMissileExplosion: the
// parent's tile, start, offset and traveled are copied wholesale
// (cpp:2043-2046) and duration becomes the row's animation length
// (cpp:2047); the succubus tiers swap the sheet first (cpp:2024-2036).
var explosionCases = []explosionCase{
	{
		name: "player caster keeps the row sheet and copies the parent",
		script: addScript{
			caster: 0, source: 0, animLen: 5, rnd: nil,
		},
		parent: missiles.Missile{}, wantAnim: 0, wantDuration: 5,
	},
	{
		name: "succubus keeps the plain blood-star explosion sheet",
		script: addScript{
			caster: int(missiles.TargetPlayers), source: 1, monsterType: 101,
			animLen: 7, rnd: nil,
		},
		parent: missiles.Missile{}, wantAnim: 23, wantDuration: 7,
	},
	{
		name: "snow witch uses the blue-tier sheet",
		script: addScript{
			caster: int(missiles.TargetPlayers), source: 1, monsterType: 102,
			animLen: 7, rnd: nil,
		},
		parent: missiles.Missile{}, wantAnim: 41, wantDuration: 7,
	},
	{
		name: "hell spawn uses the red-tier sheet",
		script: addScript{
			caster: int(missiles.TargetPlayers), source: 1, monsterType: 103,
			animLen: 7, rnd: nil,
		},
		parent: missiles.Missile{}, wantAnim: 45, wantDuration: 7,
	},
	{
		name: "solbrner uses the yellow-tier sheet",
		script: addScript{
			caster: int(missiles.TargetPlayers), source: 1, monsterType: 104,
			animLen: 7, rnd: nil,
		},
		parent: missiles.Missile{}, wantAnim: 43, wantDuration: 7,
	},
}

// explosionCase pins one AddMissileExplosion row.
type explosionCase struct {
	name         string
	script       addScript
	parent       missiles.Missile
	wantAnim     missiles.GraphicID
	wantDuration int
}

// TestAddMissileExplosion pins the parent copy and the monster-sheet switch.
func TestAddMissileExplosion(t *testing.T) {
	// given
	cases := explosionCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newAddFixture(t, tc.script)
			parent := parentMissile()
			fx.param.Parent = &parent

			// when
			missiles.AddMissileExplosion(fx.world, fx.missile, fx.param)

			// then
			assertExplosion(t, fx.missile, parent, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// parentMissile is the detonating missile: distinct tile, start, offset and
// traveled so a partial copy is visible.
func parentMissile() missiles.Missile {
	parent := missiles.Missile{}
	parent.Position.Tile = missiles.TilePosition{X: 11, Y: 12}
	parent.Position.Start = missiles.TilePosition{X: 3, Y: 4}
	parent.Position.Offset = missiles.PositionDisplacement{DeltaX: 5, DeltaY: 6}
	parent.Position.Traveled = missiles.PositionDisplacement{DeltaX: 700, DeltaY: 800}

	return parent
}

// assertExplosion checks one post-AddMissileExplosion row.
func assertExplosion(t *testing.T, m *missiles.Missile, parent missiles.Missile, tc explosionCase) {
	t.Helper()

	if missiles.GraphicID(m.AnimType) != tc.wantAnim {
		t.Fatalf("AnimType = %d, want %d", m.AnimType, tc.wantAnim)
	}

	if m.Position.Tile != parent.Position.Tile {
		t.Fatalf("tile = %v, want the parent's %v", m.Position.Tile, parent.Position.Tile)
	}

	if m.Position.Start != parent.Position.Start {
		t.Fatalf("start = %v, want the parent's %v", m.Position.Start, parent.Position.Start)
	}

	if m.Position.Offset != parent.Position.Offset {
		t.Fatalf("offset = %v, want the parent's %v", m.Position.Offset, parent.Position.Offset)
	}

	if m.Position.Traveled != parent.Position.Traveled {
		t.Fatalf("traveled = %v, want the parent's %v", m.Position.Traveled, parent.Position.Traveled)
	}

	if m.Duration != tc.wantDuration {
		t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
	}
}
