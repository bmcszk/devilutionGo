// Package missiles: the inferno flame jet's per-tick body.
package missiles

// ProcessInferno ports ProcessInferno (missiles.cpp:3937-3960): the jet burns
// on its own tile every tick, refunds its duration when the burn lands a kill
// (cpp:3342-3343), parks its frame at 20 once the growth phase ends
// (cpp:3344-3345), dims its light along the mirrored frame ramp
// (cpp:3346-3351) and only republishes while the growth phase is over
// (cpp:3359-3360).
func ProcessInferno(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3337
	m.Var2--     // cpp:3338

	saved := m.Duration

	CheckMissileCol(w, m, m.Position.Tile, MissileColProbe{ // cpp:3339-3340
		DamageType:     DamageTypeFor(m.Type),
		MinDamage:      m.Dam,
		MaxDamage:      m.Dam,
		Shifted:        true,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	if m.Duration == 0 && m.HitFlag { // cpp:3342-3343
		m.Duration = saved
	}

	if m.Var2 == 0 { // cpp:3344-3345
		m.AnimFrame = 20
	}

	if m.Var2 <= 0 { // cpp:3346-3351
		mirrorInfernoLight(w, m)
	}

	if m.Duration == 0 { // cpp:3352-3355
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}

	if m.Var2 <= 0 { // cpp:3359-3360
		PutMissile(w, m)
	}
}

// mirrorInfernoLight ports the frame-to-radius mirror of the burn-out phase
// (missiles.cpp:3347-3350): frames above 11 fold back down (24 - frame) so the
// light dims symmetrically.
func mirrorInfernoLight(w *HitWorld, m *Missile) {
	radius := m.AnimFrame
	if radius > 11 {
		radius = 24 - radius
	}

	w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, uint8(radius))
}
