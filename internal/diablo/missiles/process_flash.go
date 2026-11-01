// Package missiles: the flash rows' per-tick bodies.
package missiles

// flashTopOffsets is the probe sweep of ProcessFlashTop (missiles.cpp:3462-3466,
// identical to ProcessFlashBottom's first three entries).
var flashTopOffsets = [3]Direction{North, NorthEast, East}

// ProcessFlashTop ports ProcessFlashTop (missiles.cpp:3454-3471): the flash
// shields its caster for as long as it burns, sweeps North/NorthEast/East, and
// the last tick deletes it and drops the shield.
func ProcessFlashTop(w *HitWorld, m *Missile) {
	m.setFlashShield(w, true) // cpp:3456-3459

	m.Duration-- // cpp:3460

	for _, offset := range flashTopOffsets { // cpp:3462-3466
		CheckMissileCol(w, m, shiftTile(m.Position.Tile, offset), MissileColProbe{
			DamageType:     DamageTypeFor(m.Type),
			MinDamage:      m.Dam,
			MaxDamage:      m.Dam,
			Shifted:        true,
			Keep:           true,
			OnlyHitWalking: nil,
		})
	}

	if m.Duration == 0 { // cpp:3467-3471
		m.DelFlag = true
		m.setFlashShield(w, false)
	}

	PutMissile(w, m)
}

// setFlashShield ports the `_pInvincible` bracket of the flash rows
// (missiles.cpp:3456-3459, 3466-3470): only a monster-cast, non-trap flash
// touches the caster's flag.
func (m *Missile) setFlashShield(w *HitWorld, invincible bool) {
	if m.Caster != int(TargetMonsters) || m.IsTrap() {
		return
	}

	w.ProcessCallbacks.SetInvincible(m.Source, invincible)
}
