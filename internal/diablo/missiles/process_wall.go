// Package missiles: the Fire Wall / Flame Wave control bodies.
package missiles

import "github.com/bmcszk/devilutionGo/internal/engine"

// wallCursors is the pair of arm cursors plus whether each arm may advance.
type wallCursors struct {
	left    TilePosition
	right   TilePosition
	leftOK  bool
	rightOK bool
}

// AddWallControl ports AddWallControl (missiles.cpp:2350-2370): the control
// parks a spread tile within 5 of the aim — not the launch tile, placeable and
// with a clear missile line — records it as both arm cursors, faces the arms
// sideways from the cast direction (Left/Left and Right/Right, cpp:2365-2366)
// and lives 7 ticks.
func AddWallControl(w *HitWorld, m *Missile, p *AddParameter) {
	spread, ok := findWallSpread(w, m, p.Dst)
	if !ok { // cpp:2357-2361
		m.DelFlag = true
		p.SpellFizzled = true

		return
	}

	m.DelFlag = false
	m.Var1 = spread.X // cpp:2363-2364
	m.Var2 = spread.Y
	m.Var5 = spread.X // cpp:2365-2366
	m.Var6 = spread.Y
	m.Var3 = int(Left(Left(p.Midir)))   // cpp:2367
	m.Var4 = int(Right(Right(p.Midir))) // cpp:2368
	m.Duration = 7                      // cpp:2369
}

// ProcessWallControl ports ProcessWallControl (missiles.cpp:3770-3845): every
// tick but the last plants wall segments. The seed tick checks and plants the
// centre segment, then steps both arm cursors out; later ticks grow both arms
// with TryGrowWall, each arm stopping on its own (var7 for the right arm, the
// limit flag for the left) while the control keeps ticking.
func ProcessWallControl(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3772

	if m.Duration == 0 { // cpp:3773-3776
		m.DelFlag = true

		return
	}

	spawnID, damage := wallSegment(w, m)

	leftPosition := TilePosition{X: m.Var1, Y: m.Var2}
	rightPosition := TilePosition{X: m.Var5, Y: m.Var6}
	leftDirection := Direction(m.Var3)
	rightDirection := Direction(m.Var4)

	cursors := wallCursors{left: leftPosition, right: rightPosition}
	if leftPosition == rightPosition { // cpp:3806-3816
		if !seedWall(w, m, spawnID, leftPosition, leftDirection, damage) {
			m.DelFlag = true

			return
		}

		cursors.left, cursors.leftOK = MoveWallToNextPosition(leftPosition, leftDirection)
		cursors.right, cursors.rightOK = MoveWallToNextPosition(rightPosition, rightDirection)
	} else { // cpp:3817-3828
		cursors = grownWallCursors(w, m, spawnID, damage, cursors)
	}

	advanceWallArms(m, cursors)
}

// AddFlameWaveControl ports AddFlameWaveControl (missiles.cpp:2559-2566): the
// aim tile, a 1-tick life and frame 4 of the control's sheet.
func AddFlameWaveControl(_ *HitWorld, m *Missile, p *AddParameter) {
	m.Var1 = p.Dst.X
	m.Var2 = p.Dst.Y
	m.Duration = 1
	m.AnimFrame = 4
}

// ProcessFlameWaveControl ports ProcessFlameWaveControl (missiles.cpp:3875-3905):
// the control plants one Flame Wave segment one tile along the line from the
// caster to the aim tile and then grows (spellLevel / 2) + 2 segments to each
// side, both arms stopping at the first blocked tile; the 1-tick life then
// deletes the control.
func ProcessFlameWaveControl(w *HitWorld, m *Missile) {
	id := m.Source
	pdir := w.ProcessCallbacks.PlayerDirection(id)
	src := m.Position.Tile
	sd := GetDirection(src, TilePosition{X: m.Var1, Y: m.Var2})
	dx, dy := directionDeltaOf(sd)
	start := TilePosition{X: src.X + dx, Y: src.Y + dy}

	if CanPlaceWall(w, start) { // cpp:3882
		PlaceWall(w, id, FlameWave, start, pdir, m.SpellLevel, 0)
		flameWaveArms(w, m, start, sd, pdir)
	}

	m.Duration-- // cpp:3901-3903
	if m.Duration == 0 {
		m.DelFlag = true
	}
}

// flameWaveArms ports the two growth loops (missiles.cpp:3886-3899).
func flameWaveArms(w *HitWorld, m *Missile, start TilePosition, sd, pdir Direction) {
	segments := (m.SpellLevel / 2) + 2

	growArm(w, m, start, Left(Left(sd)), pdir, segments)
	growArm(w, m, start, Right(Right(sd)), pdir, segments)
}

// growArm walks one arm out from start, stopping at the first blocked tile.
func growArm(w *HitWorld, m *Missile, start TilePosition, dir, pdir Direction, segments int) {
	cursor := start

	for range segments {
		dx, dy := directionDeltaOf(dir)
		cursor = TilePosition{X: cursor.X + dx, Y: cursor.Y + dy}

		if !TryGrowWall(w, m.Source, FlameWave, cursor, dir, pdir, m.SpellLevel, 0) {
			return
		}
	}
}

// seedWall ports the seed-tile half of ProcessWallControl (missiles.cpp:3807-3816):
// the centre segment is planted only on a placeable tile.
func seedWall(w *HitWorld, m *Missile, spawnID MissileID, position TilePosition,
	direction Direction, damage int,
) bool {
	if !CanPlaceWall(w, position) {
		return false
	}

	PlaceWall(w, m.Source, spawnID, position, direction, m.SpellLevel, damage)

	return true
}

// grownWallCursors ports the grow-tile half (missiles.cpp:3817-3828): each arm
// steps out only when its own stop flag is clear and the wall took.
func grownWallCursors(w *HitWorld, m *Missile, spawnID MissileID, damage int, cursors wallCursors) wallCursors {
	if m.LimitReached == 0 && TryGrowWall(w, m.Source, spawnID, cursors.left,
		Direction(m.Var3), South, m.SpellLevel, damage) {
		cursors.left, cursors.leftOK = MoveWallToNextPosition(cursors.left, Direction(m.Var3))
	}

	if m.Var7 == 0 && TryGrowWall(w, m.Source, spawnID, cursors.right,
		Direction(m.Var4), South, m.SpellLevel, damage) {
		cursors.right, cursors.rightOK = MoveWallToNextPosition(cursors.right, Direction(m.Var4))
	}

	return cursors
}

// wallSegment resolves the wall missile and damage packet for the control's
// type (missiles.cpp:3784-3804).
func wallSegment(w *HitWorld, m *Missile) (MissileID, int) {
	if m.IsTrap() { // cpp:3789-3799: traps deal no level-scaled damage
		return FireWall, 16 * generateRndSum(w, 10, 2)
	}

	if m.Type == FireWallControl {
		return FireWall, 0
	}

	level := w.AddSource.playerStats(m.Source).Level

	return LightningWall, 16 * (generateRndSum(w, 10, 2) + level + 2)
}

// advanceWallArms ports the cursor write-back and the per-arm stop flags
// (missiles.cpp:3830-3842).
func advanceWallArms(m *Missile, cursors wallCursors) {
	if cursors.leftOK {
		m.Var1 = cursors.left.X
		m.Var2 = cursors.left.Y
	} else {
		m.LimitReached = 1
	}

	if cursors.rightOK {
		m.Var5 = cursors.right.X
		m.Var6 = cursors.right.Y
	} else {
		m.Var7 = 1
	}
}

// findWallSpread ports the FindClosestValidPosition predicate of the add body
// (missiles.cpp:2352-2356).
func findWallSpread(w *HitWorld, m *Missile, dst TilePosition) (TilePosition, bool) {
	start := m.Position.Start

	posOk := func(candidate engine.Point) bool {
		target := TilePosition{X: candidate.X, Y: candidate.Y}

		return start != target && CanPlaceWall(w, target) && LineClearMissile(w, start, target)
	}

	spot, found := engine.FindClosestValidPosition(posOk, engine.Point{X: dst.X, Y: dst.Y}, 0, 5)
	if !found {
		return TilePosition{}, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, true
}
