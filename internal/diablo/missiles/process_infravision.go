// Package missiles: the infravision tick.
package missiles

// ProcessInfravision ports ProcessInfravision (missiles.cpp:3839-3849): the
// flag is raised on every tick and the expiry tick deletes the missile and
// recomputes the player's item values, which is what clears the flag again.
func ProcessInfravision(w *HitWorld, m *Missile) {
	w.ProcessCallbacks.SetInfravision(m.Source) // cpp:3842

	m.Duration-- // cpp:3841

	if m.Duration == 0 { // cpp:3843-3847
		m.DelFlag = true
		w.ProcessCallbacks.RecalcPlayerItems(m.Source)
	}
}
