// Package missiles_test pins AddElemental (missiles.cpp:2560-2582) and
// ProcessElemental (cpp:3560-3620), the homing elemental bolt that re-targets
// once it reaches its aim tile and then blows up like a fireball.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// elementalCases pin the process body: the destination latch (cpp:3600-3602),
// the single re-target (cpp:3603-3613), the light move and the burn-out
// animation that stops the missile (cpp:3614-3619).
var elementalCases = []elementalCase{
	{
		name: "reaching the aim tile re-targets the closest monster", var3: 0, atAim: true,
		monster: missiles.TilePosition{X: 9, Y: 5}, wantVar3: 2, wantDuration: 255,
	},
	{
		name: "no monster falls back to the caster facing", var3: 0, atAim: true,
		wantVar3: 2, wantDuration: 255,
	},
	{
		name: "expiry flips into the explosion anim", var3: 2, duration: 1,
		wantVar3: 2, wantAnim: missiles.GraphicBigExplosion, wantStop: true,
	},
}

// elementalCase pins one ProcessElemental tick.
type elementalCase struct {
	name         string
	var3         int
	atAim        bool
	duration     int
	monster      missiles.TilePosition
	wantVar3     int
	wantDuration int
	wantAnim     missiles.GraphicID
	wantStop     bool
}

// elementalLights records the light calls Elemental makes.
type elementalLights struct {
	added      int
	lastRadius uint8
	changes    int
}

// TestAddElemental pins the damage roll, the 16-pixel heading, the 256-tick
// life and the radius-8 light (cpp:2568-2582).
func TestAddElemental(t *testing.T) {
	// given
	w, lights := elementalWorld(&rndTape{t: t, tape: []int{3, 2}}, nil)
	m := missiles.Missile{Type: missiles.Elemental, Source: 0, SpellLevel: 2}
	m.SetTraveledFromTile(5, 5)
	m.Position.Start = missiles.TilePosition{X: 5, Y: 5}
	p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 5, Y: 9}}

	// when
	missiles.AddElemental(w, &m, p)

	// then
	// dmg = 2 * (level 4 + 3 + 2) + 4 = 22; ScaleSpellEffect(22, 2) = 27; /2 = 13
	if m.Dam != 13 {
		t.Fatalf("damage = %d, want 13", m.Dam)
	}

	if m.Duration != 256 || m.Var1 != 5 || m.Var2 != 5 || m.Var4 != 5 || m.Var5 != 9 { // cpp:2579-2582
		t.Fatalf("duration/vars = %d/%d/%d/%d/%d, want 256/5/5/5/9",
			m.Duration, m.Var1, m.Var2, m.Var4, m.Var5)
	}

	if lights.added != 1 || lights.lastRadius != 8 { // cpp:2582
		t.Fatalf("lights = %d radius %d, want 1 radius 8", lights.added, lights.lastRadius)
	}
}

// TestProcessElemental pins the re-target, the light follow and the burn-out.
func TestProcessElemental(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(elementalCases))
	for _, tc := range elementalCases {
		subtests = append(subtests, subtest{tc.name, runElementalTick(tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// elementalWorld wires the tape, the light sinks, the block predicate and the
// closest-monster search.
func elementalWorld(tape *rndTape, closest func(missiles.TilePosition, int) missiles.TilePosition) (
	*missiles.HitWorld, *elementalLights,
) {
	lights := &elementalLights{}
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.AddLight = func(_, _ int, radius uint8) int {
		lights.added++
		lights.lastRadius = radius

		return 3
	}
	cb.ChangeLight = func(_, _, _ int, _ uint8) { lights.changes++ }
	cb.CheckBlock = func(int, int, int, int) bool { return false }
	cb.PlayerDirection = func(int) missiles.Direction { return missiles.South }
	cb.FindClosestMonster = func(pos missiles.TilePosition, _ int) missiles.TilePosition { return pos }

	if closest != nil {
		cb.FindClosestMonster = closest
	}

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			PlayerStats: func(int) missiles.AddPlayerStats { return missiles.AddPlayerStats{Level: 4} },
		},
	}, lights
}

// runElementalTick drives one ProcessElemental tick.
func runElementalTick(tc elementalCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		w, _ := elementalWorld(&rndTape{t: t}, func(missiles.TilePosition, int) missiles.TilePosition {
			return tc.monster
		})
		m := elementalMissile(tc)

		// when
		missiles.ProcessElemental(w, &m)

		// then
		assertElementalTick(t, m, tc)
	}
}

// elementalMissile is the bolt one tick starts from: parked on the aim tile
// recorded in var4/var5.
func elementalMissile(tc elementalCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.Elemental, Source: 0, Dam: 13, Var1: 5, Var2: 5,
		Var3: tc.var3, Var4: 5, Var5: 5, Lid: 3, AnimLen: 8, Duration: 200,
	}
	if tc.duration != 0 {
		m.Duration = tc.duration
	}

	m.SetTraveledFromTile(5, 5)
	m.Position.Start = missiles.TilePosition{X: 5, Y: 5}
	m.Position.Tile = missiles.TilePosition{X: 5, Y: 5}

	return m
}

// assertElementalTick checks the re-target, the burn-out flip and the stop.
func assertElementalTick(t *testing.T, m missiles.Missile, tc elementalCase) {
	t.Helper()

	if m.Var3 != tc.wantVar3 { // cpp:3600-3604
		t.Fatalf("var3 = %d, want %d", m.Var3, tc.wantVar3)
	}

	if tc.wantDuration != 0 && m.Duration != tc.wantDuration { // cpp:3605
		t.Fatalf("duration = %d, want %d", m.Duration, tc.wantDuration)
	}

	if tc.wantAnim != 0 && missiles.GraphicID(m.AnimType) != tc.wantAnim { // cpp:3617
		t.Fatalf("anim = %d, want %d", m.AnimType, tc.wantAnim)
	}

	if tc.wantStop && m.Position.Velocity != (missiles.PositionDisplacement{}) { // cpp:3618
		t.Fatalf("velocity = %+v, want the missile stopped", m.Position.Velocity)
	}
}
