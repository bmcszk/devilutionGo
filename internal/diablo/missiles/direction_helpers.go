// Package missiles: the direction ring helpers shared by the process bodies.
package missiles

func leftOf(facing Direction) Direction {
	return Direction((uint8(facing) + 7) % 8)
}

func rightOf(facing Direction) Direction {
	return Direction((uint8(facing) + 1) % 8)
}

// graphicSetDefaultFrameGroup ports Missile::setDefaultFrameGroup
// (missiles.h:186-189 → setFrameGroupRaw(0), :180-184): the frame group drops
// to 0 and the animation restarts on frame 1.
//
// The setAnimation tail reloads AnimDelay/AnimLen from the row's timing
// columns (reloadAnimTiming); rows without them keep the caller's values.
func (m *Missile) setDefaultFrameGroup() {
	m.setFrameGroup(0)
}

// setFrameGroup ports Missile::setFrameGroup (missiles.h:192-195 →
// setFrameGroupRaw :180-184): the frame group becomes group and the
// animation restarts on frame 1 with the group's own timing.
func (m *Missile) setFrameGroup(group int) {
	m.MFNum = group
	m.AnimFrame = 1
	m.AnimCnt = 0
	m.reloadAnimTiming()
}

// setAnimation ports Missile::setAnimation — the sheet swap by graphic
// id _miAnimType store).
func (m *Missile) setAnimation(g GraphicID) {
	m.AnimType = int(g)
}

func fireballImpactRing(w *HitWorld, m *Missile, mp TilePosition, damageType DamageType, minDam, maxDam int) bool {
	offsets := [9]Direction{
		NoDirection, SouthWest, NorthEast, SouthEast,
		East, South, NorthWest, West, North,
	}
	hit := false
	for _, off := range offsets {
		dx, dy := directionDeltaOf(off)
		px, py := mp.X+dx, mp.Y+dy
		if w.ProcessCallbacks.CheckBlock(m.Position.Start.X, m.Position.Start.Y, px, py) {
			continue
		}
		res := CheckMissileCol(w, m, TilePosition{X: px, Y: py}, MissileColProbe{
			DamageType:     damageType,
			MinDamage:      minDam,
			MaxDamage:      maxDam,
			Shifted:        false,
			Keep:           true,
			OnlyHitWalking: nil,
		})
		if res.MonsterHit || res.PlayerHit {
			hit = true
		}
	}
	return hit
}

func fireballWallAdjust(w *HitWorld, m *Missile, mp TilePosition) {
	fireballWallAdjustNorthward(w, m, mp)
	fireballWallAdjustSouthward(w, m, mp)
	fireballWallAdjustEastward(w, m, mp)
}

// fireballWallAdjustNorthward ports the first wall-shift branch
// (missiles.cpp:3115-3118): a missile whose entry tile is opaque, or moving
// north/west past a solid vertical neighbor of a transparent tile, shifts
// the tile one step south and lifts the offset 32px.
func fireballWallAdjustNorthward(w *HitWorld, m *Missile, mp TilePosition) {
	if !w.ProcessCallbacks.TransAt(mp.X, mp.Y) ||
		(m.Position.Velocity.DeltaX < 0 &&
			((w.ProcessCallbacks.TransAt(mp.X, mp.Y+1) && w.Grid.IsTileSolid(mp.X, mp.Y+1)) ||
				(w.ProcessCallbacks.TransAt(mp.X, mp.Y-1) && w.Grid.IsTileSolid(mp.X, mp.Y-1)))) {
		sdx, sdy := directionDeltaOf(South)
		m.Position.Tile = TilePosition{X: m.Position.Tile.X + sdx, Y: m.Position.Tile.Y + sdy}
		m.Position.Offset.DeltaY -= 32
	}
}

// fireballWallAdjustSouthward ports the dy>0 wall-shift branch
// (missiles.cpp:3119-3123): solid east/west neighbor of a transparent mp
// lifts the offset 32px.
func fireballWallAdjustSouthward(w *HitWorld, m *Missile, mp TilePosition) {
	if m.Position.Velocity.DeltaY > 0 &&
		((w.ProcessCallbacks.TransAt(mp.X+1, mp.Y) && w.Grid.IsTileSolid(mp.X+1, mp.Y)) ||
			(w.ProcessCallbacks.TransAt(mp.X-1, mp.Y) && w.Grid.IsTileSolid(mp.X-1, mp.Y))) {
		m.Position.Offset.DeltaY -= 32
	}
}

// fireballWallAdjustEastward ports the dx>0 wall-shift branch
// (missiles.cpp:3128-3131): solid north/south neighbor of a transparent mp
// shifts the offset x by -32px.
func fireballWallAdjustEastward(w *HitWorld, m *Missile, mp TilePosition) {
	if m.Position.Velocity.DeltaX > 0 &&
		((w.ProcessCallbacks.TransAt(mp.X, mp.Y+1) && w.Grid.IsTileSolid(mp.X, mp.Y+1)) ||
			(w.ProcessCallbacks.TransAt(mp.X, mp.Y-1) && w.Grid.IsTileSolid(mp.X, mp.Y-1))) {
		m.Position.Offset.DeltaX -= 32
	}
}

// chargedBoltFlight ports the pre-burn-out arm of ProcessChargedBolt
// (missiles.cpp:4060-4090): re-steer via the BPath table, move/collide,
// and freeze into the Lightning animation on a hit.
func chargedBoltFlight(w *HitWorld, m *Missile) {
	if m.Var3 == 0 {
		chargedBoltResteer(m)
	} else {
		m.Var3--
	}
	MoveMissileAndCheckMissileCol(w, m, DamageTypeFor(m.Type), m.Dam, m.Dam, false, false)

	if m.HitFlag {
		m.Var1 = 8
		m.setDefaultFrameGroup()
		m.Position.Offset = PositionDisplacement{}
		m.Position.Velocity = PositionDisplacement{}
		m.setAnimation(GraphicLightning)
		m.Duration = m.AnimLen
	}
	w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, uint8(m.Var1))
}

// chargedBoltResteer ports the BPath re-steer (missiles.cpp:4063-4078):
// the bPathTable entry turns the heading one step left/right and the
// velocity is re-aimed 8px toward the neighbor tile.
func chargedBoltResteer(m *Missile) {
	md := Direction(m.Var2)
	switch bPathTable[m.Mirnd] {
	case -1:
		md = leftOf(md)
	case 1:
		md = rightOf(md)
	default:
		// straight: BPath 0 leaves the heading untouched
	}
	m.Mirnd = (m.Mirnd + 1) & 0xF
	mdx, mdy := directionDeltaOf(md)
	target := TilePosition{
		X: m.Position.Tile.X + mdx,
		Y: m.Position.Tile.Y + mdy,
	}
	UpdateMissileVelocity(m, target, 8)
	m.Var3 = 16
}

// chargedBoltBurnOut ports the duration-0 tail of ProcessChargedBolt
// (missiles.cpp:4092-4096): delete and unlight.
func chargedBoltBurnOut(w *HitWorld, m *Missile) {
	if m.Duration == 0 {
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}
}

func processHolyBoltFlight(w *HitWorld, m *Missile) {
	dam := m.Dam
	MoveMissileAndCheckMissileCol(w, m, DamageTypeFor(m.Type), dam, dam, true, true)
	if m.Duration == 0 {
		m.setDefaultFrameGroup()
		m.setAnimation(GraphicHolyBoltExplosion)
		m.Duration = m.AnimLen - 1
		StopMissile(m)
		return
	}
	if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
	}
}
