// Package missiles_test pins AddBoneSpirit (missiles.cpp:2584-2600) and
// ProcessBoneSpirit (cpp:3700-3745), the homing bolt whose damage becomes the
// target's life and whose burn-out frame dies after seven ticks.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// boneSpiritCases pin the process body: the re-target that resets the damage
// from the target's life (cpp:3729-3734), the caster-facing fallback
// (cpp:3735-3738), the burn-out handover to NoDirection (cpp:3742-3745) and
// the 7-tick death frame (cpp:3703-3710).
var boneSpiritCases = []boneSpiritCase{
	{
		name: "reaching the aim tile homes and re-damages", atAim: true,
		monster: missiles.TilePosition{X: 9, Y: 5}, monsterLife: 2048,
		wantDam: 16, wantVar3: 2, wantDuration: 255,
	},
	{name: "no monster keeps the caster facing", atAim: true, wantVar3: 2, wantDuration: 255, wantDam: 40},
	{
		name: "expiry drops the direction and starts the death frame", burnOut: true,
		wantVar3: 2, wantDuration: 7, wantNoDirection: true, wantStop: true,
	},
	{name: "the death frame deletes the bolt", dead: true, wantVar3: 2, wantDel: true, wantUnlights: 1},
}

// boneSpiritCase pins one ProcessBoneSpirit tick.
type boneSpiritCase struct {
	name            string
	atAim           bool
	burnOut         bool
	dead            bool
	monster         missiles.TilePosition
	monsterLife     int
	wantDam         int
	wantVar3        int
	wantDuration    int
	wantNoDirection bool
	wantStop        bool
	wantDel         bool
	wantUnlights    int
}

// boneSpiritLights records the light traffic.
type boneSpiritLights struct {
	added      int
	lastRadius uint8
	unlights   int
}

// TestAddBoneSpirit pins the 16-pixel heading, the 256-tick life, the aim tile
// and the radius-8 light (cpp:2590-2599).
func TestAddBoneSpirit(t *testing.T) {
	// given
	w, lights := boneSpiritWorld(&rndTape{t: t}, nil)
	m := missiles.Missile{Type: missiles.BoneSpirit, Source: 0}
	m.SetTraveledFromTile(5, 5)
	m.Position.Start = missiles.TilePosition{X: 5, Y: 5}

	// when
	missiles.AddBoneSpirit(w, &m, &missiles.AddParameter{Dst: missiles.TilePosition{X: 5, Y: 9}})

	// then
	if m.Duration != 256 || m.Var1 != 5 || m.Var2 != 5 || m.Var4 != 5 || m.Var5 != 9 {
		t.Fatalf("duration/vars = %d/%d/%d/%d/%d, want 256/5/5/5/9",
			m.Duration, m.Var1, m.Var2, m.Var4, m.Var5)
	}

	if lights.added != 1 || lights.lastRadius != 8 {
		t.Fatalf("lights = %d radius %d, want 1 radius 8", lights.added, lights.lastRadius)
	}
}

// TestProcessBoneSpirit pins the homing and the death frame.
func TestProcessBoneSpirit(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(boneSpiritCases))
	for _, tc := range boneSpiritCases {
		subtests = append(subtests, subtest{tc.name, runBoneSpiritTick(tc)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// boneSpiritWorld wires the tape, the light sinks, the closest-monster search
// and the monster life read.
func boneSpiritWorld(
	tape *rndTape,
	closest func(missiles.TilePosition, int) (missiles.TilePosition, int),
) (*missiles.HitWorld, *boneSpiritLights) {
	lights := &boneSpiritLights{}
	cb := missiles.NoopProcessCallbacks()
	cb.GenerateRnd = tape.draw
	cb.AddLight = func(_, _ int, radius uint8) int {
		lights.added++
		lights.lastRadius = radius

		return 3
	}
	cb.ChangeLight = func(int, int, int, uint8) {}
	cb.AddUnLight = func(int) { lights.unlights++ }
	cb.CheckBlock = func(int, int, int, int) bool { return false }
	cb.PlayerDirection = func(int) missiles.Direction { return missiles.South }
	cb.FindClosestMonsterWithLife = closest

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}, lights
}

// runBoneSpiritTick drives one ProcessBoneSpirit tick.
func runBoneSpiritTick(tc boneSpiritCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		closest := func(missiles.TilePosition, int) (missiles.TilePosition, int) {
			return tc.monster, tc.monsterLife
		}
		if tc.monster == (missiles.TilePosition{}) {
			closest = func(missiles.TilePosition, int) (missiles.TilePosition, int) {
				return missiles.TilePosition{}, 0
			}
		}

		w, lights := boneSpiritWorld(&rndTape{t: t}, closest)
		m := boneSpiritMissile(tc)

		// when
		missiles.ProcessBoneSpirit(w, &m)

		// then
		assertBoneSpiritTick(t, m, lights, tc)
	}
}

// boneSpiritMissile is the bolt one tick starts from.
func boneSpiritMissile(tc boneSpiritCase) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.BoneSpirit, Source: 0, Dam: 40, Var1: 5, Var2: 5,
		Var3: 2, Var4: 5, Var5: 5, Lid: 3, Duration: 200,
	}
	if tc.atAim {
		m.Var3 = 0
	}

	m.SetTraveledFromTile(5, 5)
	m.Position.Start = missiles.TilePosition{X: 5, Y: 5}
	m.Position.Tile = missiles.TilePosition{X: 5, Y: 5}
	if tc.burnOut {
		m.Duration = 1
	}

	if tc.dead {
		m.MFNum = int(missiles.NoDirection)
		m.Duration = 1
	}

	return m
}

// assertBoneSpiritTick checks the re-target, the damage and the death frame.
func assertBoneSpiritTick(t *testing.T, m missiles.Missile, lights *boneSpiritLights, tc boneSpiritCase) {
	t.Helper()

	if m.Var3 != tc.wantVar3 || m.Duration != tc.wantDuration {
		t.Fatalf("var3/duration = %d/%d, want %d/%d", m.Var3, m.Duration, tc.wantVar3, tc.wantDuration)
	}

	if tc.wantDam != 0 && m.Dam != tc.wantDam { // cpp:3730
		t.Fatalf("damage = %d, want %d", m.Dam, tc.wantDam)
	}

	assertBoneSpiritTail(t, m, lights, tc)
}

// assertBoneSpiritTail checks the burn-out handover and the death-frame delete.
func assertBoneSpiritTail(t *testing.T, m missiles.Missile, lights *boneSpiritLights, tc boneSpiritCase) {
	t.Helper()

	if tc.wantNoDirection && m.MFNum != int(missiles.NoDirection) { // cpp:3742
		t.Fatalf("direction = %d, want NoDirection", m.MFNum)
	}

	if tc.wantStop && m.Position.Velocity != (missiles.PositionDisplacement{}) { // cpp:3743
		t.Fatalf("velocity = %+v, want the bolt stopped", m.Position.Velocity)
	}

	if m.DelFlag != tc.wantDel || lights.unlights != tc.wantUnlights { // cpp:3708-3710
		t.Fatalf("del/unlights = %v/%d, want %v/%d", m.DelFlag, lights.unlights, tc.wantDel, tc.wantUnlights)
	}
}
