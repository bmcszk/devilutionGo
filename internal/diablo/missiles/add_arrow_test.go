// Package missiles_test pins AddArrow (missiles.cpp:1772-1797): the
// velocity boost chain, the midir nudge on a same-tile shot, the 16-step
// animation frame and the 256-tick duration.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// addArrowCases holds the hand-computed AddArrow rows. Velocity expectations
// are the 16.16 fixed-point vector UpdateMissileVelocity produces for the
// (tile - dst) displacement (missiles.cpp:173-185); the (0,0)->(5,0) row uses
// the normalized diagonal 46340/23170 (magnitude sqrt(50)).
var addArrowCases = []addArrowCase{
	{
		name: "sorcerer level 1 keeps base velocity", start: [2]int{0, 0}, dst: [2]int{10, 10},
		caster: 0, source: 0, class: 2, level: 1, wantVel: [2]int32{0, 1048576}, wantAnim: 1,
	},
	{
		name: "rogue level 5 adds 1", start: [2]int{0, 0}, dst: [2]int{10, 10},
		caster: 0, source: 0, class: 1, level: 5, wantVel: [2]int32{0, 1081344}, wantAnim: 1,
	},
	{
		name: "warrior level 9 adds 1", start: [2]int{0, 0}, dst: [2]int{10, 10},
		caster: 0, source: 0, class: 0, level: 9, wantVel: [2]int32{0, 1081344}, wantAnim: 1,
	},
	{
		name: "bard level 9 adds 1", start: [2]int{0, 0}, dst: [2]int{10, 10},
		caster: 0, source: 0, class: 4, level: 9, wantVel: [2]int32{0, 1081344}, wantAnim: 1,
	},
	{
		name: "random arrow velocity replaces base", start: [2]int{0, 0}, dst: [2]int{10, 10},
		caster: 0, source: 0, class: 2, level: 1, flags: missiles.ItemSpecialEffectRandomArrowVelocity,
		rnd: []int{0}, wantVel: [2]int32{0, 524288}, wantAnim: 1,
	},
	{
		name: "hellfire attack flags add 1+2+4+8", start: [2]int{0, 0}, dst: [2]int{10, 10},
		caster: 0, source: 0, class: 2, level: 1, hellfire: true,
		flags: missiles.ItemSpecialEffectQuickAttack | missiles.ItemSpecialEffectFastAttack |
			missiles.ItemSpecialEffectFasterAttack | missiles.ItemSpecialEffectFastestAttack,
		wantVel: [2]int32{0, 1540096}, wantAnim: 1,
	},
	{
		name: "attack flags ignored without hellfire", start: [2]int{0, 0}, dst: [2]int{10, 10},
		caster: 0, source: 0, class: 2, level: 1,
		flags:   missiles.ItemSpecialEffectQuickAttack | missiles.ItemSpecialEffectFastAttack,
		wantVel: [2]int32{0, 1048576}, wantAnim: 1,
	},
	{
		name: "same tile shot shifts by midir", start: [2]int{5, 5}, dst: [2]int{5, 5},
		midir: missiles.South, caster: 0, source: 0, class: 2, level: 1,
		wantVel: [2]int32{0, 1048576}, wantAnim: 1,
	},
	{
		name: "diagonal destination", start: [2]int{0, 0}, dst: [2]int{5, 0},
		caster: 0, source: 0, class: 2, level: 1, wantVel: [2]int32{1482880, 741440}, wantAnim: 15,
	},
	{
		name: "monster caster skips the player read", start: [2]int{0, 0}, dst: [2]int{10, 10},
		caster: 1, source: 0, class: 1, level: 99, wantVel: [2]int32{0, 1048576}, wantAnim: 1,
	},
}

// addArrowCase is one AddArrow row: the AddMissile field seeding plus the
// scripted player stats and RNG tape.
type addArrowCase struct {
	name     string
	start    [2]int
	dst      [2]int
	midir    missiles.Direction
	caster   int
	source   int
	class    int
	level    int
	flags    int
	hellfire bool
	rnd      []int
	wantVel  [2]int32
	wantAnim int
}

func TestAddArrow(t *testing.T) {
	for _, tt := range addArrowCases {
		t.Run(tt.name, func(t *testing.T) {
			// given
			fx := addArrowFixture(t, tt)
			w, m, p := fx.world, fx.missile, fx.param

			// when
			missiles.AddArrow(w, m, p)

			// then
			fx.tape.assertConsumed()
			assertArrowState(t, m, tt)
		})
	}
}

// assertArrowState checks the post-AddArrow velocity, frame and duration.
func assertArrowState(t *testing.T, m *missiles.Missile, tt addArrowCase) {
	t.Helper()

	velocity := m.Position.Velocity
	if velocity.DeltaX != tt.wantVel[0] || velocity.DeltaY != tt.wantVel[1] {
		t.Fatalf("velocity = (%d, %d), want (%d, %d)", velocity.DeltaX, velocity.DeltaY,
			tt.wantVel[0], tt.wantVel[1])
	}

	if m.AnimFrame != tt.wantAnim {
		t.Fatalf("animFrame = %d, want %d", m.AnimFrame, tt.wantAnim)
	}

	if m.Duration != 256 { // cpp:1795
		t.Fatalf("duration = %d, want 256", m.Duration)
	}

	if m.Position.Tile != m.Position.Start {
		t.Fatalf("AddArrow moved the tile to %v", m.Position.Tile)
	}
}

// addArrowFixture rebuilds the AddMissile field seeding AddArrow reads
// (missiles.cpp:2810-2838: tile/start = src, caster, source) and wires the
// scripted player stats plus the RNG tape.
func addArrowFixture(t *testing.T, tt addArrowCase) *addFixture {
	t.Helper()

	w := &missiles.HitWorld{
		IsHellfire:       tt.hellfire,
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{
					Class:       tt.class,
					Level:       tt.level,
					HasItemFlag: func(flag int) bool { return tt.flags&flag != 0 },
				}
			},
		},
	}
	tape := &rndTape{t: t, tape: tt.rnd}
	w.ProcessCallbacks.GenerateRnd = tape.draw

	m := &missiles.Missile{Type: missiles.Arrow, Caster: tt.caster, Source: tt.source}
	m.SetTraveledFromTile(tt.start[0], tt.start[1])

	p := &missiles.AddParameter{
		Dst:   missiles.TilePosition{X: tt.dst[0], Y: tt.dst[1]},
		Midir: tt.midir,
	}

	return &addFixture{world: w, missile: m, param: p, tape: tape}
}
