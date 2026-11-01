// Package missiles: the Lightning row's per-tick body.
package missiles

// ProcessLightning ports ProcessLightning (missiles.cpp:3373-3386): the bolt
// burns a tick, collides only once it left its start tile, a hit refunds the
// tick, and 0 deletes it and drops the light.
func ProcessLightning(w *HitWorld, m *Missile) {
	m.Duration--                             // cpp:3375
	kept := m.Duration                       // cpp:3376
	if m.Position.Tile != m.Position.Start { // cpp:3377-3378
		CheckMissileCol(w, m, m.Position.Tile, MissileColProbe{
			DamageType:     DamageTypeFor(m.Type),
			MinDamage:      m.Dam,
			MaxDamage:      m.Dam,
			Shifted:        true,
			Keep:           false,
			OnlyHitWalking: nil,
		})
	}

	if m.HitFlag { // cpp:3379-3380
		m.Duration = kept
	}

	if m.Duration == 0 { // cpp:3381-3384
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}

	PutMissile(w, m)
}
