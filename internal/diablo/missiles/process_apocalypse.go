// Package missiles: the ApocalypseBoom row's per-tick body.
package missiles

// ProcessApocalypseBoom ports ProcessApocalypseBoom (missiles.cpp:3722-3732):
// the boom probes its own tile once, latches var1 on the first hit and dies
// when its duration runs out.
func ProcessApocalypseBoom(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3724

	if m.Var1 == 0 { // cpp:3725-3726
		CheckMissileCol(w, m, m.Position.Tile, MissileColProbe{
			DamageType:     DamageTypeFor(m.Type),
			MinDamage:      m.Dam,
			MaxDamage:      m.Dam,
			Shifted:        false,
			Keep:           true,
			OnlyHitWalking: nil,
		})
	}

	if m.HitFlag { // cpp:3727-3728
		m.Var1 = 1
	}

	if m.Duration == 0 { // cpp:3729-3730
		m.DelFlag = true
	}

	PutMissile(w, m)
}
