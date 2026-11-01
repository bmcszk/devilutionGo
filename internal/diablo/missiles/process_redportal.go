// Package missiles: the red portal's per-tick body.
package missiles

// redPortalIdleGroup is RedPortalFrame::Idle (missiles.h:84-87).
const redPortalIdleGroup = 1

// levelTypeTown is DTYPE_TOWN (levels/gendung_defs.hpp:16).
const levelTypeTown = 0

// redPortalExpLight is the expLight ramp of ProcessRedPortal
// (missiles.cpp:4102): 1..15 then a flat 15 tail.
var redPortalExpLight = [17]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 15, 15}

// ProcessRedPortal ports ProcessRedPortal (missiles.cpp:4100-4125): the
// countdown stops at 1, the portal flips to the idle frames when the duration
// reaches var1, the light ramps one entry per tick outside town, and the last
// tick deletes the portal and removes its light.
func ProcessRedPortal(w *HitWorld, m *Missile) {
	if m.Duration > 1 { // cpp:4104-4105
		m.Duration--
	}

	if m.Duration == m.Var1 { // cpp:4106-4107
		m.setFrameGroup(redPortalIdleGroup)
	}

	if w.ProcessCallbacks.LevelType != levelTypeTown && m.MFNum != redPortalIdleGroup && m.Duration != 0 {
		sweepRedPortalLight(w, m) // cpp:4109-4114
	}

	if m.Duration == 0 { // cpp:4115-4118
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}

	PutMissile(w, m) // cpp:4120
}

// sweepRedPortalLight ports the ramp body (missiles.cpp:4110-4113): the first
// step registers the light at radius 1, every step moves it to the next
// expLight entry.
func sweepRedPortalLight(w *HitWorld, m *Missile) {
	if m.Var2 == 0 {
		m.Lid = w.ProcessCallbacks.AddLight(m.Position.Tile.X, m.Position.Tile.Y, 1)
	}

	w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
		uint8(redPortalExpLight[minInt(m.Var2, len(redPortalExpLight)-1)]))
	m.Var2++
}

// minInt is the clamped index guard of the expLight read; upstream's var2 never
// leaves the array (the portal dies at 17 ramp steps).
func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}
