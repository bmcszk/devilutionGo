// Package missiles_test pins AddMagmaBall (missiles.cpp:1896-1922): the
// 3-tick velocity lead applied to traveled, the UpdateMissilePos snap, and
// the hellfire whole-pixel duration branch.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// addMagmaBallCases holds the hand-computed AddMagmaBall rows: velocity is
// 16.16 for the (tile - dst) normal vector, traveled grows by 3*velocity,
// and UpdateMissilePos snaps the tile from traveled>>16 through
// ScreenToMissile (missiles.cpp:1899-1901).
var addMagmaBallCases = []addMagmaBallCase{
	{
		name: "base game always runs 256 ticks", script: addScript{
			dst: [2]int{10, 10}, caster: 1, source: 0, dam: 1, light: 42,
		},
		wantVel: [2]int32{0, 524288}, wantTraveled: [2]int32{0, 1572864},
		wantTile: [2]int{1, 1}, wantOffset: [2]int32{0, -8}, wantDuration: 256,
	},
	{
		name: "hellfire zero velocity runs 1 tick", script: addScript{
			dst: [2]int{0, 0}, caster: 1, source: 0, dam: 1, light: 42, hellfire: true,
		},
		wantVel: [2]int32{0, 0}, wantTraveled: [2]int32{0, 0},
		wantTile: [2]int{0, 0}, wantOffset: [2]int32{0, 0}, wantDuration: 1,
	},
	{
		name: "hellfire fractional velocity runs 256 ticks", script: addScript{
			dst: [2]int{5, 0}, caster: 1, source: 0, dam: 1, light: 42, hellfire: true,
		},
		wantVel: [2]int32{741440, 370720}, wantTraveled: [2]int32{2224320, 1112160},
		wantTile: [2]int{1, 0}, wantOffset: [2]int32{1, 0}, wantDuration: 256,
	},
}

// addMagmaBallCase is one AddMagmaBall row.
type addMagmaBallCase struct {
	name         string
	script       addScript
	wantVel      [2]int32
	wantTraveled [2]int32
	wantTile     [2]int
	wantOffset   [2]int32
	wantDuration int
}

func TestAddMagmaBall(t *testing.T) {
	for _, tt := range addMagmaBallCases {
		t.Run(tt.name, func(t *testing.T) {
			// given
			w, m, p := newAddFixture(t, tt.script)

			// when
			missiles.AddMagmaBall(w, m, p)

			// then
			assertMagmaBallState(t, m, tt)
		})
	}
}

// assertMagmaBallState checks the post-AddMagmaBall motion and timing state
// (missiles.cpp:1898-1922).
func assertMagmaBallState(t *testing.T, m *missiles.Missile, tt addMagmaBallCase) {
	t.Helper()

	assertMagmaBallMotion(t, m, tt)
	assertMagmaBallPosition(t, m, tt)
	assertMagmaBallTiming(t, m, tt)
}

// assertMagmaBallMotion checks the velocity and the 3-tick traveled lead.
func assertMagmaBallMotion(t *testing.T, m *missiles.Missile, tt addMagmaBallCase) {
	t.Helper()

	velocity, traveled := m.Position.Velocity, m.Position.Traveled
	if velocity.DeltaX != tt.wantVel[0] || velocity.DeltaY != tt.wantVel[1] {
		t.Errorf("velocity = (%d, %d), want (%d, %d)", velocity.DeltaX, velocity.DeltaY,
			tt.wantVel[0], tt.wantVel[1])
	}

	if traveled.DeltaX != tt.wantTraveled[0] || traveled.DeltaY != tt.wantTraveled[1] {
		t.Errorf("traveled = (%d, %d), want (%d, %d)", traveled.DeltaX, traveled.DeltaY,
			tt.wantTraveled[0], tt.wantTraveled[1])
	}
}

// assertMagmaBallPosition checks the snap tile and the pixel offset.
func assertMagmaBallPosition(t *testing.T, m *missiles.Missile, tt addMagmaBallCase) {
	t.Helper()

	if m.Position.Tile.X != tt.wantTile[0] || m.Position.Tile.Y != tt.wantTile[1] {
		t.Errorf("tile = (%d, %d), want (%d, %d)", m.Position.Tile.X, m.Position.Tile.Y,
			tt.wantTile[0], tt.wantTile[1])
	}

	offset := m.Position.Offset
	if offset.DeltaX != tt.wantOffset[0] || offset.DeltaY != tt.wantOffset[1] {
		t.Errorf("offset = (%d, %d), want (%d, %d)", offset.DeltaX, offset.DeltaY,
			tt.wantOffset[0], tt.wantOffset[1])
	}
}

// assertMagmaBallTiming checks the duration, light and damage fields.
func assertMagmaBallTiming(t *testing.T, m *missiles.Missile, tt addMagmaBallCase) {
	t.Helper()

	if m.Duration != tt.wantDuration { // cpp:1902-1905
		t.Fatalf("duration = %d, want %d", m.Duration, tt.wantDuration)
	}

	if m.Lid != 42 { // cpp:1908
		t.Fatalf("lid = %d, want 42", m.Lid)
	}

	if m.Dam != 1 { // cpp:1909-1921, pre-rolled damage is kept
		t.Fatalf("dam = %d, want 1", m.Dam)
	}
}
