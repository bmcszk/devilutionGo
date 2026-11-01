// Package missiles: the Bone Spirit's bodies.
package missiles

// AddBoneSpirit ports AddBoneSpirit (missiles.cpp:2584-2600): the bolt heads
// 16 pixels per tick toward the aim tile (nudged by midir when it is the
// launch tile), lives 256 ticks, remembers the launch tile and the aim tile
// (var1/var2 and var4/var5) and carries a radius-8 light.
func AddBoneSpirit(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:2586-2588
		dx, dy := directionDeltaOf(p.Midir)
		dst = TilePosition{X: dst.X + dx, Y: dst.Y + dy}
	}

	UpdateMissileVelocity(m, dst, 16) // cpp:2589
	setDirection(m, GetDirection(m.Position.Start, dst))
	m.Duration = 256            // cpp:2591
	m.Var1 = m.Position.Start.X // cpp:2592
	m.Var2 = m.Position.Start.Y // cpp:2593
	m.Var4 = dst.X              // cpp:2594
	m.Var5 = dst.Y              // cpp:2595
	m.Lid = w.ProcessCallbacks.AddLight(m.Position.Start.X, m.Position.Start.Y, 8)
}

// ProcessBoneSpirit ports ProcessBoneSpirit (missiles.cpp:3700-3745): while it
// has a direction the bolt flies and probes; on reaching the aim tile it homes
// once — re-reading its damage from the target's life (life >> 7, cpp:3730) —
// or keeps the caster's facing when there is no monster. A bolt with no
// direction is the 7-tick death frame, which only rides the light and deletes
// itself with it.
func ProcessBoneSpirit(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3702

	if Direction(m.MFNum) == NoDirection {
		boneSpiritDeathFrame(w, m)

		return
	}

	MoveMissileAndCheckMissileCol(w, m, DamageTypeFor(m.Type), m.Dam, m.Dam, false, false) // cpp:3714

	tile := m.Position.Tile
	boneSpiritRetarget(w, m, tile)

	if tile != (TilePosition{X: m.Var1, Y: m.Var2}) { // cpp:3739-3743
		m.Var1 = tile.X
		m.Var2 = tile.Y
		w.ProcessCallbacks.ChangeLight(m.Lid, tile.X, tile.Y, 8)
	}

	if m.Duration == 0 { // cpp:3745-3749
		setDirection(m, NoDirection)
		m.Position.Velocity = PositionDisplacement{}
		m.Duration = 7
	}

	PutMissile(w, m)
}

// boneSpiritRetarget ports the one-shot homing step (missiles.cpp:3719-3738).
func boneSpiritRetarget(w *HitWorld, m *Missile, tile TilePosition) {
	if m.Var3 == 0 && tile == (TilePosition{X: m.Var4, Y: m.Var5}) { // cpp:3719-3721
		m.Var3 = 1
	}

	if m.Var3 != 1 { // cpp:3722
		return
	}

	m.Var3 = 2       // cpp:3723
	m.Duration = 255 // cpp:3724

	if target, life := w.ProcessCallbacks.FindClosestMonsterWithLife(tile, 19); // cpp:3726-3734
	target != (TilePosition{}) {
		m.Dam = life >> 7
		setDirection(m, GetDirection(tile, target))
		UpdateMissileVelocity(m, target, 16)

		return
	}

	dir := w.ProcessCallbacks.PlayerDirection(m.Source) // cpp:3735-3738
	dx, dy := directionDeltaOf(dir)
	setDirection(m, dir)
	UpdateMissileVelocity(m, TilePosition{X: tile.X + dx, Y: tile.Y + dy}, 16)
}

// boneSpiritDeathFrame ports the direction-less branch (missiles.cpp:3703-3712).
func boneSpiritDeathFrame(w *HitWorld, m *Missile) {
	w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, uint8(m.AnimFrame))

	if m.Duration == 0 {
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}

	PutMissile(w, m)
}
