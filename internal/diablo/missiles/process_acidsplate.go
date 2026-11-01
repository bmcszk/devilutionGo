// Package missiles: the acid splat's per-tick body.
package missiles

// ProcessAcidSplate ports ProcessAcidSplate (missiles.cpp:3639-3654): the first
// tick applies the render shift the splat does not carry into the puddle, then
// the splat counts down; the tick reaching 0 spawns the acid puddle under it
// (damage 2 for a level 2+ source monster) and deletes the splat without
// republishing it.
func ProcessAcidSplate(w *HitWorld, m *Missile) {
	if m.Duration == m.AnimLen { // cpp:3641-3644
		m.Position.Tile = shiftTile(m.Position.Tile, South) // Displacement{1, 1}
		m.Position.Offset.DeltaY -= 32
	}

	m.Duration-- // cpp:3645

	if m.Duration == 0 { // cpp:3647-3653
		m.DelFlag = true
		w.ProcessCallbacks.SpawnMissile(m.Position.Tile, TilePosition{}, South, AcidPuddle,
			int(TargetPlayers), m.Source, acidPuddleDamage(w, m.Source), m.SpellLevel, nil)

		return
	}

	PutMissile(w, m)
}

// acidPuddleDamage ports the `Monsters[monst].data().level >= 2 ? 2 : 1`
// damage pick of the splat tail (missiles.cpp:3648).
func acidPuddleDamage(w *HitWorld, source int) int {
	if w.AddSource.monsterStats(source).Level >= 2 {
		return 2
	}

	return 1
}
