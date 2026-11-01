// Package missiles: the elemental bolt's bodies.
package missiles

// AddElemental ports AddElemental (missiles.cpp:2560-2582): damage is
// ScaleSpellEffect(2 * (level + GenerateRndSum(10, 2)) + 4, spell level) / 2
// (cpp:2570-2572), the bolt heads 16 pixels per tick toward the aim tile
// (nudged by midir when it is the launch tile), lives 256 ticks, remembers the
// launch tile and the aim tile and carries a radius-8 light (cpp:2574-2582).
func AddElemental(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:2562-2564
		dx, dy := directionDeltaOf(p.Midir)
		dst = TilePosition{X: dst.X + dx, Y: dst.Y + dy}
	}

	level := w.AddSource.playerStats(m.Source).Level
	dmg := 2*(level+generateRndSum(w, 10, 2)) + 4 // cpp:2570
	m.Dam = scaleSpellEffect(dmg, m.SpellLevel) / 2

	UpdateMissileVelocity(m, dst, 16) // cpp:2574
	setDirection(m, GetDirection(m.Position.Start, dst))
	m.Duration = 256            // cpp:2577
	m.Var1 = m.Position.Start.X // cpp:2578
	m.Var2 = m.Position.Start.Y // cpp:2579
	m.Var4 = dst.X              // cpp:2580
	m.Var5 = dst.Y              // cpp:2581
	m.Lid = w.ProcessCallbacks.AddLight(m.Position.Start.X, m.Position.Start.Y, 8)
}

// ProcessElemental ports ProcessElemental (missiles.cpp:3560-3620): the bolt
// flies until it reaches the aim tile recorded in var4/var5, re-targets the
// closest monster within 19 tiles once — or keeps the caster's facing when
// there is none — and then runs its 255-tick second leg; when the life runs
// out it flips into the BigExplosion animation, stops dead and damages the
// nine tiles around itself. The explosion frames then run until they expire
// and delete the missile with its light.
func ProcessElemental(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3562

	if GraphicID(m.AnimType) == GraphicBigExplosion {
		elementalExplosionTick(w, m)

		return
	}

	damageType := DamageTypeFor(m.Type)
	MoveMissileAndCheckMissileCol(w, m, damageType, m.Dam, m.Dam, false, false) // cpp:3592

	elementalRetarget(w, m)

	if tile := m.Position.Tile; tile != (TilePosition{X: m.Var1, Y: m.Var2}) { // cpp:3614-3617
		m.Var1 = tile.X
		m.Var2 = tile.Y
		w.ProcessCallbacks.ChangeLight(m.Lid, tile.X, tile.Y, 8)
	}

	if m.Duration == 0 { // cpp:3619-3623
		m.setDefaultFrameGroup()
		m.setAnimation(GraphicBigExplosion)
		m.Duration = m.AnimLen - 1
		m.Position.Velocity = PositionDisplacement{}
	}

	PutMissile(w, m) // cpp:3625
}

// elementalRetarget ports the one-shot homing step (missiles.cpp:3600-3613).
func elementalRetarget(w *HitWorld, m *Missile) {
	tile := m.Position.Tile

	if m.Var3 == 0 && tile == (TilePosition{X: m.Var4, Y: m.Var5}) { // cpp:3600-3602
		m.Var3 = 1
	}

	if m.Var3 != 1 { // cpp:3603
		return
	}

	m.Var3 = 2       // cpp:3604
	m.Duration = 255 // cpp:3605

	if target := w.ProcessCallbacks.FindClosestMonster(tile, 19); target != (TilePosition{}) { // cpp:3606-3610
		setDirection(m, GetDirection(tile, target))
		UpdateMissileVelocity(m, target, 16)

		return
	}

	dir := w.ProcessCallbacks.PlayerDirection(m.Source) // cpp:3611-3613
	dx, dy := directionDeltaOf(dir)
	setDirection(m, dir)
	UpdateMissileVelocity(m, TilePosition{X: tile.X + dx, Y: tile.Y + dy}, 16)
}

// elementalExplosionTick ports the BigExplosion frames (missiles.cpp:3564-3587):
// the light rides the frame, the nine neighbours of the start tile take the
// damage unless a wall stands between them, and the last frame deletes the
// missile with its light.
func elementalExplosionTick(w *HitWorld, m *Missile) {
	w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, uint8(m.AnimFrame))

	startPoint := m.Position.Start
	if m.Var3 == 2 { // cpp:3570
		startPoint = TilePosition{X: m.Var4, Y: m.Var5}
	}

	offsets := [9]Direction{NoDirection, SouthWest, NorthEast, SouthEast, East, South, NorthWest, West, North}
	for _, off := range offsets { // cpp:3572-3585
		dx, dy := directionDeltaOf(off)
		px, py := m.Position.Tile.X+dx, m.Position.Tile.Y+dy

		if w.ProcessCallbacks.CheckBlock(startPoint.X, startPoint.Y, px, py) {
			continue
		}

		CheckMissileCol(w, m, TilePosition{X: px, Y: py}, MissileColProbe{
			DamageType: DamageTypeFor(m.Type), MinDamage: m.Dam, MaxDamage: m.Dam,
			Shifted: true, Keep: true, OnlyHitWalking: nil,
		})
	}

	if m.Duration == 0 { // cpp:3587-3590
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}

	PutMissile(w, m)
}
