// Package missiles_test pins the container append + field seeding of
// AddMissile (missiles.cpp:2801-2855) and the frame-group pick it drives
// (missiles.cpp:2834-2837).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// frameGroupCase pins one frame-group pick (missiles.cpp:2834-2837).
type frameGroupCase struct {
	name    string
	graphic missiles.GraphicID
	midir   missiles.Direction
	want    int
}

// TestAddMissileSeedsRow: the pack lands in the row, the addFn runs, and the
// position block starts on the source tile.
func TestAddMissileSeedsRow(t *testing.T) {
	// given
	list := &missiles.MissileList{}
	args := missiles.AddMissileArgs{
		Src:        missiles.TilePosition{X: 12, Y: 20},
		Dst:        missiles.TilePosition{X: 12, Y: 30},
		Midir:      missiles.South,
		Type:       missiles.Arrow,
		Caster:     int(missiles.TargetMonsters),
		Source:     0,
		Dam:        7,
		SpellLevel: 3,
		Graphic:    missiles.GraphicArrow,
	}

	// when
	m := list.AddMissile(args)

	// then
	if m == nil {
		t.Fatal("AddMissile returned nil for a live addFn")
	}

	if list.Len() != 1 {
		t.Fatalf("container rows = %d, want 1", list.Len())
	}

	assertSeededIdentity(t, m, args)
	assertSeededAnimation(t, m, args)
}

// TestAddMissileFrameGroup: sheets with fewer than 8 frame groups keep frame
// group 0, the rest follow midir (missiles.cpp:2834-2837). The rows use an unwired id so the
// stub addFn leaves the pick in place — every ported addFn re-runs
// setDirection on its own (missiles.cpp:1775, 2331).
func TestAddMissileFrameGroup(t *testing.T) {
	// given
	cases := []frameGroupCase{
		{"arrow sheet keeps group 0", missiles.GraphicArrow, missiles.South, 0},
		{"bloodstar sheet keeps group 0", missiles.GraphicBloodStar, missiles.West, 0},
		{"magmaball 8-group sheet follows midir", missiles.GraphicMagmaBall, missiles.West, int(missiles.West)},
		{"fireball 16-group sheet follows midir", missiles.GraphicFireball, missiles.North, int(missiles.North)},
	}

	// when
	missilesByCase := make([]*missiles.Missile, 0, len(cases))
	for _, tc := range cases {
		missilesByCase = append(missilesByCase, addHealingMissile(tc))
	}

	// then
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertFrameGroup(t, missilesByCase[i], tc)
		})
	}
}

// addHealingMissile appends one unwired-id row through AddMissile.
func addHealingMissile(tc frameGroupCase) *missiles.Missile {
	list := &missiles.MissileList{}

	return list.AddMissile(missiles.AddMissileArgs{
		Src:     missiles.TilePosition{X: 5, Y: 5},
		Dst:     missiles.TilePosition{X: 5, Y: 9},
		Midir:   tc.midir,
		Type:    missiles.Healing,
		Caster:  int(missiles.TargetMonsters),
		Source:  0,
		Graphic: tc.graphic,
	})
}

// assertFrameGroup checks the picked frame group and the frame restart.
func assertFrameGroup(t *testing.T, m *missiles.Missile, tc frameGroupCase) {
	t.Helper()

	if m == nil {
		t.Fatal("AddMissile returned nil")
	}

	if m.MFNum != tc.want {
		t.Fatalf("MFNum = %d, want %d", m.MFNum, tc.want)
	}

	if m.AnimFrame != 1 {
		t.Fatalf("AnimFrame = %d, want 1", m.AnimFrame)
	}

	if m.AnimCnt != 0 {
		t.Fatalf("AnimCnt = %d, want 0", m.AnimCnt)
	}
}

// assertSeededIdentity checks the identity and source seeding of the row
// (missiles.cpp:2813-2818).
func assertSeededIdentity(t *testing.T, m *missiles.Missile, args missiles.AddMissileArgs) {
	t.Helper()

	if m.Type != args.Type {
		t.Fatalf("Type = %d, want %d", m.Type, args.Type)
	}

	if m.Caster != args.Caster {
		t.Fatalf("Caster = %d, want %d", m.Caster, args.Caster)
	}

	if m.Source != args.Source {
		t.Fatalf("Source = %d, want %d", m.Source, args.Source)
	}

	if m.SpellLevel != args.SpellLevel {
		t.Fatalf("SpellLevel = %d, want %d", m.SpellLevel, args.SpellLevel)
	}
}

// assertSeededAnimation checks the position and animation seeding
// (missiles.cpp:2819-2828) plus the arrow's 256-tick duration.
func assertSeededAnimation(t *testing.T, m *missiles.Missile, args missiles.AddMissileArgs) {
	t.Helper()

	if m.Position.Tile != args.Src || m.Position.Start != args.Src {
		t.Fatalf("position = %+v, want tile and start on %+v", m.Position, args.Src)
	}

	if m.AnimAdd != 1 {
		t.Fatalf("AnimAdd = %d, want 1", m.AnimAdd)
	}

	if m.AnimType != int(args.Graphic) {
		t.Fatalf("AnimType = %d, want %d", m.AnimType, int(args.Graphic))
	}

	if m.LastCollisionTargetHash != 0 {
		t.Fatalf("LastCollisionTargetHash = %d, want 0", m.LastCollisionTargetHash)
	}

	if m.Duration != 256 {
		t.Fatalf("Duration = %d, want the arrow's 256 ticks", m.Duration)
	}
}
