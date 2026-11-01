// Package missiles_test pins AddAcid (missiles.cpp:2328-2352): the fixed
// velocity 16 with no midir nudge, the intelligence-scaled duration, the
// NO_LIGHT sentinel and the closing PutMissile.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// addAcidCases holds the hand-computed AddAcid rows. The fixture's monster
// intelligence is 5, so the base duration is 5*(5+4) = 45 (cpp:2332).
var addAcidCases = []addAcidCase{
	{
		name: "monster acid lives 5*(intelligence+4) ticks", script: addScript{
			dst: [2]int{10, 10}, caster: 1, source: 3, monsterMin: 2, monsterMax: 5,
			rnd: []int{2},
		},
		wantVel: [2]int32{0, 524288}, wantDuration: 45, wantDam: 4,
	},
	{
		name: "hellfire zero velocity lasts 1 tick", script: addScript{
			dst: [2]int{0, 0}, caster: 1, source: 3, dam: 7, hellfire: true,
		},
		wantVel: [2]int32{0, 0}, wantDuration: 1, wantDam: 7, wantMFNum: 1,
	},
	{
		name: "trap rolls the trap dice", script: addScript{
			dst: [2]int{10, 10}, caster: 0, source: -1, currlevel: 5, rnd: []int{3},
		},
		wantVel: [2]int32{0, 524288}, wantDuration: 45, wantDam: 8,
	},
	{
		name: "player branch leaves damage untouched", script: addScript{
			dst: [2]int{10, 10}, caster: 0, source: 0,
		},
		wantVel: [2]int32{0, 524288}, wantDuration: 45, wantDam: 0,
	},
}

// addAcidCase is one AddAcid row.
type addAcidCase struct {
	name         string
	script       addScript
	wantVel      [2]int32
	wantDuration int
	wantDam      int
	wantMFNum    int
}

func TestAddAcid(t *testing.T) {
	for _, tt := range addAcidCases {
		t.Run(tt.name, func(t *testing.T) {
			// given
			w, m, p := newAddFixture(t, tt.script)

			// when
			missiles.AddAcid(w, m, p)

			// then
			assertAcidState(t, w, m, tt)
		})
	}
}

// assertAcidState checks the post-AddAcid motion, timing and damage fields
// plus the PutMissile flag write (missiles.cpp:2330-2351).
func assertAcidState(t *testing.T, w *missiles.HitWorld, m *missiles.Missile, tt addAcidCase) {
	t.Helper()

	assertAcidMotion(t, m, tt)
	assertAcidFields(t, w, m, tt)
}

// assertAcidMotion checks the velocity, frame group and duration.
func assertAcidMotion(t *testing.T, m *missiles.Missile, tt addAcidCase) {
	t.Helper()

	velocity := m.Position.Velocity
	if velocity.DeltaX != tt.wantVel[0] || velocity.DeltaY != tt.wantVel[1] {
		t.Fatalf("velocity = (%d, %d), want (%d, %d)", velocity.DeltaX, velocity.DeltaY,
			tt.wantVel[0], tt.wantVel[1])
	}

	if m.MFNum != tt.wantMFNum { // cpp:2331, no midir nudge
		t.Fatalf("frame group = %d, want %d", m.MFNum, tt.wantMFNum)
	}

	if m.Duration != tt.wantDuration { // cpp:2332-2335
		t.Fatalf("duration = %d, want %d", m.Duration, tt.wantDuration)
	}
}

// assertAcidFields checks the light, the scratch registers, damage and the
// PutMissile flag write.
func assertAcidFields(t *testing.T, w *missiles.HitWorld, m *missiles.Missile, tt addAcidCase) {
	t.Helper()

	if m.Lid != missiles.NoLight { // cpp:2336
		t.Fatalf("lid = %d, want NO_LIGHT", m.Lid)
	}

	if m.Var1 != m.Position.Start.X || m.Var2 != m.Position.Start.Y { // cpp:2337-2338
		t.Fatalf("var1/var2 = (%d, %d), want start tile (%d, %d)", m.Var1, m.Var2,
			m.Position.Start.X, m.Position.Start.Y)
	}

	if m.Dam != tt.wantDam { // cpp:2339-2350
		t.Fatalf("dam = %d, want %d", m.Dam, tt.wantDam)
	}

	flags := w.Grid.DFlags[m.Position.Tile.Y][m.Position.Tile.X]
	if flags&world.FlagMissile == 0 { // cpp:2351 PutMissile
		t.Fatalf("missile flag not set at (%d, %d)", m.Position.Tile.X, m.Position.Tile.Y)
	}
}
