// Package missiles_test pins AddElementalArrow (missiles.cpp:1735-1775).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// elementalArrowCases holds the hand-computed rows: base speed 32, the
// level/4 (Rogue) or level/8 (Warrior, Bard) bonus, the hellfire attack-speed
// flags, and the non-hellfire -1 for those three classes (cpp:1743-1762).
// Velocity is the 16.16 vector UpdateMissileVelocity builds for the
// (tile - dst) displacement; the midir row shifts a same-tile shot by South =
// (1, 1) and so flies due south at 32.
var elementalArrowCases = []elementalArrowCase{
	{
		name:  "non-hellfire rogue adds level/4 then loses the class point",
		start: [2]int{0, 0}, dst: [2]int{10, 10}, caster: 0, source: 0,
		class: missiles.HeroClassRogue, level: 12,
		wantVel: [2]int32{0, 1114112}, wantAnim: 1,
	},
	{
		name:  "hellfire warrior adds level/8 plus quick and faster attack",
		start: [2]int{0, 0}, dst: [2]int{10, 10}, caster: 0, source: 0,
		class: missiles.HeroClassWarrior, level: 20, hellfire: true,
		flags:   missiles.ItemSpecialEffectQuickAttack | missiles.ItemSpecialEffectFasterAttack,
		wantVel: [2]int32{0, 1277952}, wantAnim: 1,
	},
	{
		name:  "sorcerer keeps the base speed",
		start: [2]int{0, 0}, dst: [2]int{10, 10}, caster: 0, source: 0,
		class: missiles.HeroClassSorcerer, level: 30,
		wantVel: [2]int32{0, 1048576}, wantAnim: 1,
	},
	{
		name:  "player caster skips the class chain",
		start: [2]int{0, 0}, dst: [2]int{10, 10}, caster: int(missiles.TargetPlayers),
		source: 0, class: missiles.HeroClassRogue, level: 12,
		wantVel: [2]int32{0, 1048576}, wantAnim: 1,
	},
	{
		name:  "same-tile shot shifts by midir",
		start: [2]int{5, 5}, dst: [2]int{5, 5}, midir: missiles.South,
		caster: int(missiles.TargetPlayers), source: 0,
		wantVel: [2]int32{0, 1048576}, wantAnim: 1,
	},
}

// elementalArrowCase pins one AddElementalArrow row.
type elementalArrowCase struct {
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
	wantVel  [2]int32
	wantAnim int
}

// TestAddElementalArrow pins the velocity chain, the direction-16 frame, the
// 256-tick duration, the var1/var2 launch tile and the 5-radius light.
func TestAddElementalArrow(t *testing.T) {
	// given
	cases := elementalArrowCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// given
			fx := newElementalArrowFixture(t, tc)

			// when
			missiles.AddElementalArrow(fx.world, fx.missile, fx.param)

			// then
			assertElementalArrow(t, fx, tc)
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// newElementalArrowFixture seeds the AddMissile fields AddElementalArrow reads
// and wires the scripted player stats plus the light sink.
func newElementalArrowFixture(t *testing.T, tc elementalArrowCase) *addFixture {
	t.Helper()

	light := &lightCall{}
	w := &missiles.HitWorld{
		IsHellfire:       tc.hellfire,
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats {
				return missiles.AddPlayerStats{
					Class:       tc.class,
					Level:       tc.level,
					HasItemFlag: func(flag int) bool { return tc.flags&flag != 0 },
				}
			},
		},
	}
	w.ProcessCallbacks.AddLight = func(x, y int, radius uint8) int {
		*light = lightCall{x: x, y: y, radius: radius}

		return 9
	}

	m := &missiles.Missile{Type: missiles.FireArrow, Caster: tc.caster, Source: tc.source}
	m.SetTraveledFromTile(tc.start[0], tc.start[1])

	p := &missiles.AddParameter{
		Dst:   missiles.TilePosition{X: tc.dst[0], Y: tc.dst[1]},
		Midir: tc.midir,
	}

	return &addFixture{world: w, missile: m, param: p, tape: &rndTape{t: t}, light: light}
}

// assertElementalArrow checks one post-AddElementalArrow state.
func assertElementalArrow(t *testing.T, fx *addFixture, tc elementalArrowCase) {
	t.Helper()

	m := fx.missile
	velocity := m.Position.Velocity
	if velocity.DeltaX != tc.wantVel[0] || velocity.DeltaY != tc.wantVel[1] { // cpp:1764
		t.Fatalf("velocity = (%d, %d), want (%d, %d)",
			velocity.DeltaX, velocity.DeltaY, tc.wantVel[0], tc.wantVel[1])
	}

	if m.AnimFrame != tc.wantAnim { // cpp:1766
		t.Fatalf("AnimFrame = %d, want %d", m.AnimFrame, tc.wantAnim)
	}

	if m.Duration != 256 { // cpp:1767
		t.Fatalf("Duration = %d, want 256", m.Duration)
	}

	start := m.Position.Start
	if m.Var1 != start.X || m.Var2 != start.Y { // cpp:1768-1769
		t.Fatalf("var1/var2 = (%d, %d), want (%d, %d)", m.Var1, m.Var2, start.X, start.Y)
	}

	light := lightCall{x: start.X, y: start.Y, radius: 5}
	if *fx.light != light { // cpp:1770
		t.Fatalf("light = %+v, want %+v", *fx.light, light)
	}
}
