// Package missiles: the Diablo apocalypse body.
package missiles

// AddDiabloApocalypse ports AddDiabloApocalypse (missiles.cpp:2788-2799):
// every active player with a clear missile line from the launch tile takes one
// DiabloApocalypseBoom aimed South from the {0,0} source tile, and the missile
// then deletes itself.
func AddDiabloApocalypse(w *HitWorld, m *Missile, _ *AddParameter) {
	for _, player := range w.ProcessCallbacks.ActivePlayers() { // cpp:2790-2793
		future := w.ProcessCallbacks.PlayerFutureTile(player)

		if !LineClearMissile(w, m.Position.Start, future) { // cpp:2794-2796
			continue
		}

		w.ProcessCallbacks.SpawnMissile(TilePosition{}, future, South, DiabloApocalypseBoom, // cpp:2797
			m.Caster, m.Source, m.Dam, 0, nil)
	}

	m.DelFlag = true // cpp:2798
}
