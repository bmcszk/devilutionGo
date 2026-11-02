// Package missiles: the Hellfire Reflect add body.
package missiles

// maxReflections is std::numeric_limits<uint16_t>::max(), the ceiling
// AddReflect tests before it adds (missiles.cpp:1292).
const maxReflections = 65535

// AddReflect ports AddReflect (missiles.cpp:1278-1297): the row always deletes
// itself; a player caster adds spell level (2 when unlevelled) times its
// character level to wReflections — dropping the gain instead of overflowing
// the uint16 counter — and the local player's total is sent to the other
// clients.
func AddReflect(w *HitWorld, m *Missile, _ *AddParameter) {
	m.DelFlag = true // cpp:1280

	if m.SourceOf() != SourcePlayer { // cpp:1282-1283
		return
	}

	level := w.AddSource.playerStats(m.Source).Level
	add := 2 * level // cpp:1287

	if m.SpellLevel != 0 { // cpp:1287
		add = m.SpellLevel * level
	}

	reflections := w.ProcessCallbacks.PlayerReflections(m.Source)
	if reflections+add >= maxReflections { // cpp:1288-1290
		add = 0
	}

	reflections += add
	w.ProcessCallbacks.SetPlayerReflections(m.Source, reflections) // cpp:1294

	if w.ProcessCallbacks.IsMyPlayer(m.Source) { // cpp:1295-1296
		w.ProcessCallbacks.SendSetReflect(reflections)
	}
}
