// Package missiles: the projectile motion pipeline and the
// ProcessArrow / ProcessGenericProjectile ports — UpdateMissileVelocity,
// PutMissile, UpdateMissilePos, MoveMissilePos, MoveMissile,
// MoveMissileAndCheckMissileCol and the two Process fns
// (devilGo-mwlm slice 3a-beta).
// Engine boundary: upstream reads the lighting engine (ChangeLightOffset,
// ChangeLight, AddUnLight, AddLight — missiles.cpp:223, 691), spawns explosions
// through AddMissile, resolves tile availability through
// IsTileAvailable, draws through GenerateRnd and
// reads currlevel/dTransVal; those systems land with later beans, so they
// are injected through the ProcessCallbacks sink carried on HitWorld (the
// hitCallbacks pattern, collision.go — no-op defaults via
// NoopProcessCallbacks; tests wire recording spies).
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// NoLight mirrors the NO_LIGHT sentinel.
const NoLight = -1

// expLightTable ports ExpLight: the explosion
// light radius ramp, 15 entries.
var expLightTable = [15]int{9, 10, 11, 12, 11, 10, 8, 6, 4, 2, 1, 0, 0, 0, 0}

// bPathTable ports BPath: the ChargedBolt
// steering draw table, 16 entries of -1/0/1 (left/straight/right).
var bPathTable = [16]int{-1, 0, 1, -1, 0, 1, -1, -1, 0, 0, 1, 1, 0, 1, -1, 0}

// ProcessCallbacks is the engine-boundary sink the process functions
// read instead of the upstream globals (the hitCallbacks pattern,
// collision.go): the lighting engine, the explosion
// spawn chain (AddMissile, missiles.cpp:2823), the source/elemental damage
// reads (Players/Monsters/currlevel), the tile predicates (IsTileAvailable,
// tile_properties.cpp; dTransVal, gendung.h; CheckBlock, missiles.cpp:138) and the
// GenerateRnd draw. Carried on HitWorld; NoopProcessCallbacks returns the
// no-op defaults (tests wire recording spies).
type ProcessCallbacks struct {
	// MissilePreFlag ports the engine global MissilePreFlag
	// (missiles.cpp:74): PutMissile raises it when the missile carries
	// _miPreFlag (missiles.cpp:209-211); ProcessMissiles clears it per frame
	// (missiles.cpp:4211). Sink state, not a func.
	MissilePreFlag bool
	// ChangeLightOffset ports ChangeLightOffset
	// (lighting.cpp; called from missiles.cpp:223) — UpdateMissilePos feeds it
	// the missile's light-grid offset.
	ChangeLightOffset func(lid int, offsetX, offsetY int8)
	// ChangeLight ports ChangeLight (lighting.cpp, lid, x, y,
	// radius) — the light trackers move a light with the missile
	// (e.g. the firewall/wall updates at missiles.cpp:3983-4012).
	ChangeLight func(lid, x, y int, radius uint8)
	// AddUnLight ports AddUnLight (lighting.cpp, param lid) — the light
	// removal of the burn-out branches (e.g. missiles.cpp:2999-3001).
	AddUnLight func(lid int)
	// AddLight ports AddLight (lighting.cpp, params x, y, radius):
	// registers a new light and returns its id — the explosion light
	// ramp seeds _mlid from it (missiles.cpp:691, 1613 et al.).
	AddLight func(x, y int, radius uint8) int
	// SpawnMissile forwards to AddMissile (missiles.cpp:2823;
	// params: src, dst, dir, spawnID, caster, source, parent — the
	// AddMissile argument pack). Real spawn chains land with the
	// missile container and the explosion Add* bodies
	// (devilGo-mwlm 3c).
	SpawnMissile func(src, dst TilePosition, dir Direction,
		spawnID MissileID, caster, source int, parent *Missile)
	// SourceDamage resolves the source damage pack (cpp:2947-2953):
	// player _pIMinDam/_pIMaxDam, monster minDamage/maxDamage,
	// trap currlevel dice. Real damage sourcing lands with the
	// player/monster tables (devilGo-mwlm 3b).
	SourceDamage func(m *Missile) (minDamage, maxDamage int)
	// ElementDamage ports the elemental player-damage read of the
	// elemental-arrow explosion switch (cpp:2944-2953):
	// _pILMinDam/_pILMaxDam (Lightning) and _pIFMinDam/_pIFMaxDam (Fire).
	ElementDamage func(m *Missile, element DamageType) (minDamage, maxDamage int)
	// IsTileAvailable ports IsTileAvailable (tile_properties.cpp):
	// bounds + non-solid + monster occupancy checks; params
	// sourceMonster, x, y.
	IsTileAvailable func(sourceMonster, x, y int) bool
	// GenerateRnd ports GenerateRnd (random.cpp, param n): the
	// [0, n) draw — the trap damage rolls (cpp:2858-2862).
	GenerateRnd func(n int) int
	// Currlevel ports the dungeon-level global (missiles.cpp:74 area
	// engine globals): the trap damage base (cpp:2858-2862).
	Currlevel int
	// TransAt ports the tile-transparency lookup (upstream reads
	// dTransVal/TransList, gendung.h): fireball wall-adjust only
	// (cpp:3103-3112).
	TransAt func(x, y int) bool
	// CheckBlock ports CheckBlock (missiles.cpp pathing helper): true
	// when the straight line from→to is blocked — the fireball ring
	// only probes tiles it can see.
	CheckBlock func(fromX, fromY, toX, toY int) bool
	// Fatal receives the ported app_fatal trips of the process fns
	// (missiles.cpp:2920): the handler signature is the upstream
	// behavior-table one and cannot return an error, so the trip is
	// reported here and the handler stops. The cmd layer presents it.
	Fatal func(fatal *appfat.FatalError)
}

// moveMissileCollisionOutcome distinguishes MoveMissile's two early
// returns (missiles.cpp:625-627) for callers that need it; the port
// keeps upstream's plain bool return.
// MoveMissile ports MoveMissile (missiles.cpp:552-628): advance
// traveled by velocity, then walk every tile the missile crossed this
// tick — interpolating through skipped tiles with the 32×16 px increment
// — probing checkTile on each. A failed probe ends the move: with
// dontMove the missile never enters the tile (the velocity step or
// increment is reverted and StopMissile freezes it); without it the
// missile lands in the tile and the caller decides. dontMove is the
// upstream `moveit` bool flag of the 1:1 signature.
type DontMoveOnFailedProbe bool

// skipStartTile is the upstream `ignoreStart` flag of MoveMissileAndCheckMissileCol
// (missiles.cpp:638): true skips collision check on the starting tile.
type skipStartTile bool

// ifCheckTileFailsDontMoveToTile is the upstream `moveit` flag of the
// interpolation loop (missiles.cpp:585-628): true reverts the increment
// step and freezes the missile.
type ifCheckTileFailsDontMoveToTile bool

// NoopProcessCallbacks returns the default no-op callback set.
func NoopProcessCallbacks() ProcessCallbacks {
	return ProcessCallbacks{
		ChangeLightOffset: func(int, int8, int8) {},
		ChangeLight:       func(int, int, int, uint8) {},
		AddUnLight:        func(int) {},
		AddLight:          func(int, int, uint8) int { return NoLight },
		SpawnMissile: func(TilePosition, TilePosition, Direction,
			MissileID, int, int, *Missile) {
		},
		SourceDamage:    func(*Missile) (int, int) { return 0, 0 },
		ElementDamage:   func(*Missile, DamageType) (int, int) { return 0, 0 },
		IsTileAvailable: func(int, int, int) bool { return false },
		GenerateRnd:     func(int) int { return 0 },
		TransAt:         func(int, int) bool { return false },
		CheckBlock:      func(int, int, int, int) bool { return true },
		Fatal:           func(*appfat.FatalError) {},
	}
}

func UpdateMissileVelocity(m *Missile, destination TilePosition, velocityInPixels int) {
	m.Position.Velocity = PositionDisplacement{}

	if m.Position.Tile == destination {
		return
	}

	// WorldToNormalScreen's point difference is the
	// displacement of (tile - destination).
	fixed16NormalVector := WorldToNormalScreen(PositionDisplacement{
		DeltaX: int32(m.Position.Tile.X - destination.X),
		DeltaY: int32(m.Position.Tile.Y - destination.Y),
	})

	m.Position.Velocity = PositionDisplacement{
		DeltaX: fixed16NormalVector.DeltaX * int32(velocityInPixels),
		DeltaY: fixed16NormalVector.DeltaY * int32(velocityInPixels),
	}
}

// StopMissile ports MissilePosition::StopMissile (missiles.h: zero the
// velocity and snap the sprite offset to the last renderer
// position when the renderer tile matches (the render fields are written
// by the renderer).
func StopMissile(m *Missile) {
	m.Position.Velocity = PositionDisplacement{}
	if m.Position.TileForRendering == m.Position.Tile {
		m.Position.Offset = m.Position.OffsetForRendering
	}
}

func PutMissile(w *HitWorld, m *Missile) {
	position := m.Position.Tile

	if !world.InDungeonBounds(position.X, position.Y) {
		m.DelFlag = true
	}

	if m.DelFlag {
		return
	}

	flags := &w.Grid.DFlags[position.Y][position.X]
	*flags |= world.FlagMissile
	if m.Type == FireWall {
		*flags |= world.FlagMissileFire
	}
	if m.Type == LightningWall {
		*flags |= world.FlagMissileLight
	}

	if m.PreFlag {
		w.ProcessCallbacks.MissilePreFlag = true
	}
}

func UpdateMissilePos(w *HitWorld, m *Missile) {
	// traveled >> 16: the whole-pixel part.
	pixelsTravelled := PositionDisplacement{
		DeltaX: m.Position.Traveled.DeltaX >> fracBits,
		DeltaY: m.Position.Traveled.DeltaY >> fracBits,
	}

	tileOffset := ScreenToMissile(pixelsTravelled)
	m.Position.Tile = TilePosition{
		X: m.Position.Start.X + int(tileOffset.DeltaX),
		Y: m.Position.Start.Y + int(tileOffset.DeltaY),
	}

	m.Position.Offset = PositionDisplacement{
		DeltaX: pixelsTravelled.DeltaX + WorldToScreen(tileOffset).DeltaX,
		DeltaY: pixelsTravelled.DeltaY + WorldToScreen(tileOffset).DeltaY,
	}

	absoluteLightOffset := ScreenToLight(pixelsTravelled)
	w.dispatchChangeLightOffset(m.Lid, PositionDisplacement{
		DeltaX: absoluteLightOffset.DeltaX - tileOffset.DeltaX*8,
		DeltaY: absoluteLightOffset.DeltaY - tileOffset.DeltaY*8,
	})
}

func MoveMissilePos(w *HitWorld, m *Missile) {
	var moveDirection Direction
	switch Direction(m.MFNum) {
	case East:
		moveDirection = SouthEast
	case West:
		moveDirection = SouthWest
	case South, SouthWest, SouthEast:
		moveDirection = South
	default:
		return
	}

	dx, dy := directionDeltaOf(moveDirection)
	target := TilePosition{X: m.Position.Tile.X + dx, Y: m.Position.Tile.Y + dy}
	if w.ProcessCallbacks.IsTileAvailable(m.Source, target.X, target.Y) {
		m.Position.Tile = target
		// Displacement(moveDirection).WorldToScreen() — upstream
		// builds the direction delta as a Displacement first
		// (missiles.cpp:243-252) so the offset delta is
		// WorldToScreen of the tile delta.
		offset := WorldToScreen(PositionDisplacement{
			DeltaX: int32(dx), DeltaY: int32(dy),
		})
		m.Position.Offset.DeltaX += offset.DeltaX
		m.Position.Offset.DeltaY += offset.DeltaY
	}
}

func MoveMissile(w *HitWorld, m *Missile, checkTile func(TilePosition) bool,
	dontMove DontMoveOnFailedProbe,
) bool {
	prevTile := m.Position.Tile
	m.Position.Traveled.DeltaX += m.Position.Velocity.DeltaX
	m.Position.Traveled.DeltaY += m.Position.Velocity.DeltaY
	UpdateMissilePos(w, m)

	possibleVisitTiles := possibleVisitTileCount(prevTile, m.Position.Tile, m.Position.Velocity)
	if possibleVisitTiles == 0 {
		return false
	}

	// Did the missile skip a tile?
	if possibleVisitTiles > 1 { // 1:1 port (missiles.cpp:566-623)
		if hit := interpolateMissilePath(w, m, &prevTile, checkTile, ifCheckTileFailsDontMoveToTile(dontMove)); hit {
			return true
		}
	}

	if !checkTile(m.Position.Tile) && bool(dontMove) {
		m.Position.Traveled.DeltaX -= m.Position.Velocity.DeltaX
		m.Position.Traveled.DeltaY -= m.Position.Velocity.DeltaY
		UpdateMissilePos(w, m)
		StopMissile(m)
	}

	return true
}

func DamageTypeFor(id MissileID) DamageType {
	switch id {
	case Firebolt, MagmaBall, Fireball:
		return DamageFire
	case BloodStar:
		return DamageMagic
	case Acid:
		return DamageAcid
	case LightningArrow:
		return DamageLightning
	default:
		return DamagePhysical
	}
}

// MoveMissileAndCheckMissileCol ports MoveMissileAndCheckMissileCol
// (missiles.cpp:636-701); ignoreStart and ifCollidesDontMoveToHitTile are the
// two upstream bool flags, kept as named parameters of the 1:1 signature.
func MoveMissileAndCheckMissileCol(w *HitWorld, m *Missile, damageType DamageType,
	minDamage, maxDamage int, skipStart skipStartTile, dontMove DontMoveOnFailedProbe,
) {
	checkTile := missileCollisionProbe(w, m, damageType, minDamage, maxDamage, skipStart)
	tileChanged := MoveMissile(w, m, checkTile, dontMove)
	if !tileChanged {
		slowMissileRecheck(w, m, damageType, minDamage, maxDamage)
	}
	// remember what target CheckMissileCol was checked against
	m.LastCollisionTargetHash = currentTargetHash(w, m)
	// remember what target CheckMissileCol was checked against
	m.LastCollisionTargetHash = currentTargetHash(w, m)
}

func ProcessArrow(w *HitWorld, m *Missile) {
	m.Duration--
	m.Dist++

	// Source damage dispatch: player
	// _pIMinDam/_pIMaxDam, monster minDamage/maxDamage, trap
	// currlevel/2*currlevel — resolved by the injected seam.
	mind, maxd := w.ProcessCallbacks.SourceDamage(m)

	MoveMissileAndCheckMissileCol(w, m, DamageTypeFor(m.Type), mind, maxd, true, false)
	if m.Duration == 0 {
		m.DelFlag = true
	}
	PutMissile(w, m)
}

func ProcessGenericProjectile(w *HitWorld, m *Missile) {
	m.Duration--

	MoveMissileAndCheckMissileCol(w, m, DamageTypeFor(m.Type), m.Dam, m.Dam, true, true)
	if m.Duration == 0 { // 1:1 port (missiles.cpp:2982-3013)
		genericProjectileImpact(w, m)
		return
	}
	if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		if m.Lid != NoLight {
			w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
		}
	}
	PutMissile(w, m)
}

func ProcessElementalArrow(w *HitWorld, m *Missile) {
	m.Duration--

	animType := GraphicID(m.AnimType)
	if animType != GraphicChargedBolt && animType != GraphicMagmaBallExplosion {
		processElementalArrowFlight(w, m)
	} else {
		// Animation phase: the light rides
		// the burn-out frames.
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(m.AnimFrame+5))
	}

	// Shared tail: duration 0 — the explosion
	// burn-out ran out — deletes the missile and removes the light.
	if m.Duration == 0 {
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}
	PutMissile(w, m)
}

func ProcessFireball(w *HitWorld, m *Missile) {
	m.Duration--

	if GraphicID(m.AnimType) == GraphicBigExplosion {
		if m.Duration == 0 {
			m.DelFlag = true
			w.ProcessCallbacks.AddUnLight(m.Lid)
		}
		PutMissile(w, m)
		return
	}

	minDam, maxDam := m.Dam, m.Dam
	if m.Caster != int(TargetMonsters) {
		minDam, maxDam = w.ProcessCallbacks.SourceDamage(m)
	}
	damageType := DamageTypeFor(m.Type)
	MoveMissileAndCheckMissileCol(w, m, damageType, minDam, maxDam, true, false)

	if m.Duration == 0 {
		mp := m.Position.Tile
		w.ProcessCallbacks.ChangeLight(m.Lid, mp.X, mp.Y, uint8(m.AnimFrame))
		fireballImpactRing(w, m, mp, damageType, minDam, maxDam)

		fireballWallAdjust(w, m, mp)

		m.setDefaultFrameGroup()
		m.setAnimation(GraphicBigExplosion)
		m.Duration = m.AnimLen - 1
		m.Position.Velocity = PositionDisplacement{}
	} else if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
	}
	PutMissile(w, m)
}

func ProcessMissileExplosion(w *HitWorld, m *Missile) {
	m.Duration--

	if m.Duration == 0 {
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
		return
	}
	if m.Var1 == 0 {
		// First tick: seed the explosion light.
		m.Lid = w.ProcessCallbacks.AddLight(m.Position.Tile.X, m.Position.Tile.Y, 9)
	} else {
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(expLightTable[m.Var1]))
	}
	m.Var1++
	PutMissile(w, m)
}

func ProcessChargedBolt(w *HitWorld, m *Missile) {
	m.Duration--

	if GraphicID(m.AnimType) != GraphicLightning {
		chargedBoltFlight(w, m)
	}

	chargedBoltBurnOut(w, m)
	PutMissile(w, m)
}

func ProcessHolyBolt(w *HitWorld, m *Missile) {
	m.Duration--

	if GraphicID(m.AnimType) != GraphicHolyBoltExplosion {
		processHolyBoltFlight(w, m)
	} else {
		// Animation phase: light rides the
		// burn-out frames.
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(m.AnimFrame+7))
		if m.Duration == 0 {
			m.DelFlag = true
			w.ProcessCallbacks.AddUnLight(m.Lid)
		}
	}
	PutMissile(w, m)
}

// dispatchChangeLightOffset forwards the light-offset update through the
// sink with the port's int8 offset pair (missiles.cpp:223 takes
// DisplacementOf<int8_t>).
func (w *HitWorld) dispatchChangeLightOffset(lid int, offset PositionDisplacement) {
	w.ProcessCallbacks.ChangeLightOffset(lid, int8(offset.DeltaX), int8(offset.DeltaY))
}

func directionDeltaOf(d Direction) (deltaX, deltaY int) {
	switch d {
	case South:
		return 1, 1
	case SouthWest:
		return 0, 1
	case West:
		return -1, 1
	case NorthWest:
		return -1, 0
	case North:
		return -1, -1
	case NorthEast:
		return 0, -1
	case East:
		return 1, -1
	case SouthEast:
		return 1, 0
	default: // NoDirection
		return 0, 0
	}
}

// interpolateMissilePath ports the skipped-tile loop of MoveMissile
// (missiles.cpp:566-623): walk in 32<<16 increments toward the missile's new
// position, probing checkTile on each intermediate tile. On a failed probe
// the traveled step is reverted per ifCheckTileFailsDontMoveToTile, the
// missile is frozen, and hit=true tells the caller to skip the final-tile
// check (the upstream return path).
func interpolateMissilePath(w *HitWorld, m *Missile, prevTile *TilePosition,
	checkTile func(TilePosition) bool, revertOnFailed ifCheckTileFailsDontMoveToTile,
) (hit bool) {
	incVelocity := missileIncrement(m)
	traveled := alignedTraveledStart(m, incVelocity)
	// upstream loop is do-while(true) (missiles.cpp:584): the body runs at
	// least once — the aligned start may be the zero vector (first tick),
	// which must still step.
	for {
		var done bool
		traveled, done = interpolationStep(w, m, traveled, incVelocity, prevTile,
			checkTile, revertOnFailed)
		if done {
			return true
		}
		if traveled == (PositionDisplacement{}) {
			return false
		}
	}
}

// interpolationStep advances one increment of the interpolation loop and,
// for a probeable new tile, runs the collision probe. It reports
// done=true when a failed probe ended the move (missiles.cpp:585-628).
func interpolationStep(w *HitWorld, m *Missile, traveled, incVelocity PositionDisplacement,
	prevTile *TilePosition, checkTile func(TilePosition) bool, revertOnFailed ifCheckTileFailsDontMoveToTile,
) (PositionDisplacement, bool) {
	traveled = steppedTraveled(m, traveled, incVelocity)
	tile, skip := intermediateTile(m, traveled)
	if tile == m.Position.Tile {
		return PositionDisplacement{}, false
	}
	if skip || tile == *prevTile {
		return traveled, false
	}
	*prevTile = tile
	if !checkTile(tile) {
		missileProbeFailed(w, m, traveled, incVelocity, revertOnFailed)
		return traveled, true
	}
	return traveled, false
}

// alignedTraveledStart ports the traveled-vector alignment
// (missiles.cpp:576-583): the loop starts from the traveled vector minus
// the velocity step, aligned to the next smallest multiple of the
// increment (C++ int division truncates toward zero).
func alignedTraveledStart(m *Missile, incVelocity PositionDisplacement) PositionDisplacement {
	traveled := PositionDisplacement{
		DeltaX: m.Position.Traveled.DeltaX - m.Position.Velocity.DeltaX,
		DeltaY: m.Position.Traveled.DeltaY - m.Position.Velocity.DeltaY,
	}
	return alignTraveledToIncrement(traveled, incVelocity)
}

// alignTraveledToIncrement ports the traveled-vector alignment
// (missiles.cpp:578-583): C++ int division truncates toward zero, so the
// start lands on the next smallest multiple of the increment.
func alignTraveledToIncrement(traveled, incVelocity PositionDisplacement) PositionDisplacement {
	if incVelocity.DeltaY != 0 {
		traveled.DeltaY = (traveled.DeltaY / incVelocity.DeltaY) * incVelocity.DeltaY
	}
	if incVelocity.DeltaX != 0 {
		traveled.DeltaX = (traveled.DeltaX / incVelocity.DeltaX) * incVelocity.DeltaX
	}
	return traveled
}

// steppedTraveled advances traveled by one increment and reports the
// reached-current-position sentinel (zero value) once the step crosses the
// sign of the diff on either axis (missiles.cpp:585-603) — upstream
// resumes normal logic at that point.
func steppedTraveled(m *Missile, traveled, incVelocity PositionDisplacement) PositionDisplacement {
	initialDiff := PositionDisplacement{
		DeltaX: m.Position.Traveled.DeltaX - traveled.DeltaX,
		DeltaY: m.Position.Traveled.DeltaY - traveled.DeltaY,
	}
	traveled.DeltaX += incVelocity.DeltaX
	traveled.DeltaY += incVelocity.DeltaY
	incDiff := PositionDisplacement{
		DeltaX: m.Position.Traveled.DeltaX - traveled.DeltaX,
		DeltaY: m.Position.Traveled.DeltaY - traveled.DeltaY,
	}
	// we are at the original calculated position => resume
	// with normal logic
	if (initialDiff.DeltaX < 0) != (incDiff.DeltaX < 0) {
		return PositionDisplacement{}
	}
	if (initialDiff.DeltaY < 0) != (incDiff.DeltaY < 0) {
		return PositionDisplacement{}
	}
	return traveled
}

// intermediateTile converts the stepped traveled vector into the
// in-between tile (missiles.cpp:605-609) and reports whether the corner
// guard (missiles.cpp:615-619) skips collision logic for this step.
func intermediateTile(m *Missile, traveled PositionDisplacement) (TilePosition, bool) {
	pixelsTraveled := PositionDisplacement{
		DeltaX: traveled.DeltaX >> fracBits,
		DeltaY: traveled.DeltaY >> fracBits,
	}
	tileOffset := ScreenToMissile(pixelsTraveled)
	tile := TilePosition{
		X: m.Position.Start.X + int(tileOffset.DeltaX),
		Y: m.Position.Start.Y + int(tileOffset.DeltaY),
	}
	// skip collision logic if the missile is on a corner
	// between tiles
	if pixelsTraveled.DeltaY%16 == 0 &&
		pixelsTraveled.DeltaX%32 == 0 &&
		absI32(pixelsTraveled.DeltaY/16)%2 != absI32(pixelsTraveled.DeltaX/32)%2 {
		return tile, true
	}
	return tile, false
}

// missileIncrement ports the 32<<16 increment computation of MoveMissile
// (missiles.cpp:570-571): the increment is int32-truncated per component —
// denominator = (2*speed.dy >= speed.dx) ? 2*speed.dy : speed.dx;
// incVelocity = velocity * ((32 << 16) / denominator).
func missileIncrement(m *Missile) PositionDisplacement {
	speed := PositionDisplacement{
		DeltaX: absI32(m.Position.Velocity.DeltaX),
		DeltaY: absI32(m.Position.Velocity.DeltaY),
	}
	denominator := float32(speed.DeltaX)
	if 2*speed.DeltaY >= speed.DeltaX {
		denominator = float32(2 * speed.DeltaY)
	}
	factor := float32(32<<16) / denominator
	return PositionDisplacement{
		DeltaX: int32(float32(m.Position.Velocity.DeltaX) * factor),
		DeltaY: int32(float32(m.Position.Velocity.DeltaY) * factor),
	}
}

// missileProbeFailed applies the failed probe of the interpolation loop:
// with ifCheckTileFailsDontMoveToTile the increment step is reverted and
// the missile frozen; otherwise the missile lands in the tile.
func missileProbeFailed(w *HitWorld, m *Missile, traveled, incVelocity PositionDisplacement,
	revertOnFailed ifCheckTileFailsDontMoveToTile,
) {
	m.Position.Traveled = traveled
	if revertOnFailed {
		m.Position.Traveled.DeltaX -= incVelocity.DeltaX
		m.Position.Traveled.DeltaY -= incVelocity.DeltaY
		UpdateMissilePos(w, m)
		StopMissile(m)
		return
	}
	UpdateMissilePos(w, m)
}

// possibleVisitTileCount ports the tile-count estimate of MoveMissile
// (missiles.cpp:563-565): axis-aligned motion counts path distance,
// diagonal motion the Manhattan distance.
func possibleVisitTileCount(prevTile, tile TilePosition, velocity PositionDisplacement) int {
	if velocity.DeltaX == 0 || velocity.DeltaY == 0 {
		return PathDistance(prevTile, tile)
	}
	return ManhattanDistance(prevTile, tile)
}

func absI32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// currentTargetHash ports the monster^player tile hash of
// MoveMissileAndCheckMissileCol (missiles.cpp:684-686).
func currentTargetHash(w *HitWorld, m *Missile) int16 {
	return int16(w.DMonster[m.Position.Tile.Y][m.Position.Tile.X] ^
		w.DPlayer[m.Position.Tile.Y][m.Position.Tile.X])
}

// missileCollisionProbe builds the per-tile probe of
// MoveMissileAndCheckMissileCol (missiles.cpp:638-657): the start tile is
// skipped when ignoreStart, a live missile keeps flying, and a blockable
// hit or blocked tile stops the move.
func missileCollisionProbe(w *HitWorld, m *Missile, damageType DamageType,
	minDamage, maxDamage int, ignoreStart skipStartTile,
) func(TilePosition) bool {
	return func(tile TilePosition) bool {
		if ignoreStart && m.Position.Start == tile {
			return true
		}

		CheckMissileCol(w, m, tile, MissileColProbe{
			DamageType:     damageType,
			MinDamage:      minDamage,
			MaxDamage:      maxDamage,
			Shifted:        false,
			Keep:           false,
			OnlyHitWalking: nil,
		})

		return missileMayKeepFlying(w, m, tile)
	}
}

// missileMayKeepFlying ports the keep-flying decision of the
// CheckMissileCol probe (missiles.cpp:649-657): a live missile continues;
// a blockable hit or a missile-blocked tile stops the move.
func missileMayKeepFlying(w *HitWorld, m *Missile, tile TilePosition) bool {
	if m.Duration != 0 {
		return true
	}
	if m.HitFlag && m.MovementDistribution == MovementBlockable {
		return false
	}
	return !IsMissileBlockedByTile(w.Grid, tile.X, tile.Y)
}

// slowMissileRecheck ports the tail of MoveMissileAndCheckMissileCol
// (missiles.cpp:690-700): when the missile didn't change the tile, probe
// the current tile once per target hash to avoid multiple hits for slow
// missiles.
func slowMissileRecheck(w *HitWorld, m *Missile, damageType DamageType, minDamage, maxDamage int) {
	tileTargetHash := currentTargetHash(w, m)

	// check that we perform CheckMissileCol only once for any
	// monster/player to avoid multiple hits for slow missiles
	if m.LastCollisionTargetHash != tileTargetHash {
		CheckMissileCol(w, m, m.Position.Tile, MissileColProbe{
			DamageType:     damageType,
			MinDamage:      minDamage,
			MaxDamage:      maxDamage,
			Shifted:        false,
			Keep:           false,
			OnlyHitWalking: nil,
		})
	}
}

// projectileExplosionOf maps each projectile row to its explosion missile
// (missiles.cpp:2986-3008); base rows outside the table never spawn an
// explosion upstream.
func projectileExplosionOf(id MissileID) (MissileID, bool) {
	switch id {
	case Firebolt, MagmaBall:
		return MagmaBallExplosion, true
	case BloodStar:
		return BloodStarExplosion, true
	case Acid:
		return AcidSplat, true
	case OrangeFlare:
		return OrangeExplosion, true
	case BlueFlare:
		return BlueExplosion, true
	case RedFlare:
		return RedExplosion, true
	case YellowFlare:
		return YellowExplosion, true
	case BlueFlare2:
		return BlueExplosion2, true
	default:
		return 0, false
	}
}

// genericProjectileImpact ports the duration-0 branch of
// ProcessGenericProjectile (missiles.cpp:2982-3013): delete the impact
// missile, spawn its explosion row and remove the flight light.
func genericProjectileImpact(w *HitWorld, m *Missile) {
	m.DelFlag = true
	if explosion, ok := projectileExplosionOf(m.Type); ok {
		dst := TilePosition{X: 0, Y: 0}
		dir := Direction(m.MFNum)
		w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, explosion,
			m.Caster, m.Source, m)
	}
	if m.Lid != NoLight {
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}
	PutMissile(w, m)
}

func leftOf(facing Direction) Direction {
	return Direction((uint8(facing) + 7) % 8)
}

func rightOf(facing Direction) Direction {
	return Direction((uint8(facing) + 1) % 8)
}

// graphicSetDefaultFrameGroup ports Missile::setDefaultFrameGroup
// the frame group drops to 0. The
// setAnimation sheet swap is the AnimType store; animation-count state
// animation-count state lands with the animation slice
// (devilGo-mwlm).
func (m *Missile) setDefaultFrameGroup() {
	m.MFNum = 0
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
