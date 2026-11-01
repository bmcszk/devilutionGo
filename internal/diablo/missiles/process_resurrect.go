// Package missiles: the resurrect beam's per-tick body.
package missiles

// ProcessResurrectBeam ports ProcessResurrectBeam (missiles.cpp:4164-4170): a
// pure countdown that deletes the beam on the tick reaching 0.
func ProcessResurrectBeam(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:4166

	if m.Duration == 0 { // cpp:4167-4168
		m.DelFlag = true
	}

	PutMissile(w, m) // cpp:4169
}
