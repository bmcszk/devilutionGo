// Package missiles: the acid puddle's per-tick body.
package missiles

// acidPuddleEndGroup is AcidPuddleFrame::End (missiles.h:69-72).
const acidPuddleEndGroup = 1

// acidPuddleIdleGroup is AcidPuddleFrame::Idle (missiles.h:69-72).
const acidPuddleIdleGroup = 0

// ProcessAcidPuddle ports ProcessAcidPuddle (missiles.cpp:2295-2320): the
// puddle damages whatever stands on its own tile every tick — the collision
// call may zero the duration, so it is saved and restored (cpp:2300-2306) —
// and the zero tick either deletes the puddle (already past the idle frames)
// or swaps to the End frames and lives another AnimLen (cpp:2307-2314).
func ProcessAcidPuddle(w *HitWorld, m *Missile) {
	m.Duration--

	remaining := m.Duration // cpp:2301

	CheckMissileCol(w, m, m.Position.Tile, MissileColProbe{
		DamageType:     DamageTypeFor(m.Type),
		MinDamage:      m.Dam,
		MaxDamage:      m.Dam,
		Shifted:        true,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	m.Duration = remaining // cpp:2306

	if remaining == 0 { // cpp:2307-2314
		if m.MFNum != acidPuddleIdleGroup {
			m.DelFlag = true
		} else {
			m.setFrameGroup(acidPuddleEndGroup)
			m.Duration = m.AnimLen
		}
	}

	PutMissile(w, m) // cpp:2318
}
