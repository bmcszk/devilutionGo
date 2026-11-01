// Package missiles: the level-1 sheet flash's per-tick body.
package missiles

// flashBottomOffsets is the six-tile probe sweep of ProcessFlashBottom
// (missiles.cpp:3430-3438).
var flashBottomOffsets = [6]Direction{NorthWest, NoDirection, SouthEast, West, SouthWest, South}

// ProcessFlashBottom ports ProcessFlashBottom (missiles.cpp:3420-3448): the
// flash shields its caster while it burns, sweeps the six diagonal-and-straight
// neighbours, and the last tick deletes it and drops the shield.
func ProcessFlashBottom(w *HitWorld, m *Missile) {
	m.setFlashShield(w, true) // cpp:3422-3425

	m.Duration-- // cpp:3426

	for _, offset := range flashBottomOffsets { // cpp:3430-3438
		CheckMissileCol(w, m, shiftTile(m.Position.Tile, offset), MissileColProbe{
			DamageType:     DamageTypeFor(m.Type),
			MinDamage:      m.Dam,
			MaxDamage:      m.Dam,
			Shifted:        true,
			Keep:           true,
			OnlyHitWalking: nil,
		})
	}

	if m.Duration == 0 { // cpp:3439-3447
		m.DelFlag = true
		m.setFlashShield(w, false)
	}

	PutMissile(w, m) // cpp:3448
}
