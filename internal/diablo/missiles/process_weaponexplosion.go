// Package missiles: the barbarian weapon explosion's per-tick body.
package missiles

// weaponExplosionExpLight is the radius curve of the ten frames
// (missiles.cpp:3575).
var weaponExplosionExpLight = [10]uint8{9, 10, 11, 12, 11, 10, 8, 6, 4, 2}

// ProcessWeaponExplosion ports ProcessWeaponExplosion (missiles.cpp:3573-3600):
// the first tick raises a radius-9 light, the following ticks dim it along the
// ExpLight curve indexed by var1, and the last tick deletes the explosion and
// drops the light. var1 grows one per tick, so a life longer than the curve
// indexes past it (upstream reads out of bounds there; unreachable with the
// AnimLen - 1 lifetimes the Add body hands out: 1..16 frames map to var1 0..15).
func ProcessWeaponExplosion(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3577

	if m.Var1 == 0 { // cpp:3578-3580
		m.Lid = w.ProcessCallbacks.AddLight(m.Position.Tile.X, m.Position.Tile.Y, 9)
	} else if m.Duration != 0 { // cpp:3581-3584
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			weaponExplosionExpLight[m.Var1%len(weaponExplosionExpLight)])
	}

	m.Var1++ // cpp:3585

	if m.Duration == 0 { // cpp:3586-3592
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)

		return
	}

	PutMissile(w, m) // cpp:3594
}
