// Package missiles: the projectile motion pipeline and the
// ProcessArrow / ProcessGenericProjectile ports — updateMissileVelocity,
// PutMissile, updateMissilePos, moveMissilePos, moveMissile,
// moveMissileAndCheckMissileCol and the two Process fns
// (devilGo-mwlm slice 3a-beta).
// Engine boundary: upstream reads the lighting engine (ChangeLightOffset,
// ChangeLight, AddUnLight, AddLight — missiles.cpp:223, 691), spawns explosions
// through AddMissile, resolves tile availability through
// IsTileAvailable, draws through GenerateRnd and
// reads currlevel/dTransVal; those systems land with later beans, so they
// are injected through the processCallbacks sink carried on hitWorld (the
// hitCallbacks pattern, collision.go — no-op defaults via
// noopProcessCallbacks; tests wire recording spies).
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// NoLight mirrors the NO_LIGHT sentinel.
const NoLight = -1

// processCallbacks is the engine-boundary sink the process functions
// read instead of the upstream globals (the hitCallbacks pattern,
// collision.go): the lighting engine, the explosion
// spawn chain (AddMissile, missiles.cpp:2823), the source/elemental damage
// reads (Players/Monsters/currlevel), the tile predicates (IsTileAvailable,
// tile_properties.cpp; dTransVal, gendung.h; CheckBlock, missiles.cpp:138) and the
// GenerateRnd draw. Carried on hitWorld; noopProcessCallbacks returns the
// no-op defaults (tests wire recording spies).
type processCallbacks struct {
	// MissilePreFlag ports the engine global MissilePreFlag
	// (missiles.cpp:74): PutMissile raises it when the missile carries
	// _miPreFlag (missiles.cpp:209-211); ProcessMissiles clears it per frame
	// (missiles.cpp:4211). Sink state, not a func.
	MissilePreFlag bool
	// ChangeLightOffset ports ChangeLightOffset
	// (lighting.cpp; called from missiles.cpp:223) — updateMissilePos feeds it
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
	SpawnMissile func(src, dst tilePosition, dir Direction,
		spawnID missileID, caster, source int, parent *Missile)
	// SourceDamage resolves the source damage pack (cpp:2947-2953):
	// player _pIMinDam/_pIMaxDam, monster minDamage/maxDamage,
	// trap currlevel dice. Real damage sourcing lands with the
	// player/monster tables (devilGo-mwlm 3b).
	SourceDamage func(m *Missile) (minDamage, maxDamage int)
	// ElementDamage ports the elemental player-damage read of the
	// elemental-arrow explosion switch (cpp:2944-2953):
	// _pILMinDam/_pILMaxDam (Lightning) and _pIFMinDam/_pIFMaxDam (Fire).
	ElementDamage func(m *Missile, element damageType) (minDamage, maxDamage int)
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
}

func noopProcessCallbacks() processCallbacks {
	return processCallbacks{
		ChangeLightOffset: func(int, int8, int8) {},
		ChangeLight:       func(int, int, int, uint8) {},
		AddUnLight:        func(int) {},
		AddLight:          func(int, int, uint8) int { return NoLight },
		SpawnMissile: func(tilePosition, tilePosition, Direction,
			missileID, int, int, *Missile) {
		},
		SourceDamage:    func(*Missile) (int, int) { return 0, 0 },
		ElementDamage:   func(*Missile, damageType) (int, int) { return 0, 0 },
		IsTileAvailable: func(int, int, int) bool { return false },
		GenerateRnd:     func(int) int { return 0 },
		TransAt:         func(int, int) bool { return false },
		CheckBlock:      func(int, int, int, int) bool { return true },
	}
}

func updateMissileVelocity(m *Missile, destination tilePosition, velocityInPixels int) {
	m.Position.Velocity = positionDisplacement{}

	if m.Position.Tile == destination {
		return
	}

	// worldToNormalScreen's point difference is the
	// displacement of (tile - destination).
	fixed16NormalVector := worldToNormalScreen(positionDisplacement{
		DeltaX: int32(m.Position.Tile.X - destination.X),
		DeltaY: int32(m.Position.Tile.Y - destination.Y),
	})

	m.Position.Velocity = positionDisplacement{
		DeltaX: fixed16NormalVector.DeltaX * int32(velocityInPixels),
		DeltaY: fixed16NormalVector.DeltaY * int32(velocityInPixels),
	}
}

// stopMissile ports MissilePosition::stopMissile (missiles.h: zero the
// velocity and snap the sprite offset to the last renderer
// position when the renderer tile matches (the render fields are written
// by the renderer).
func stopMissile(m *Missile) {
	m.Position.Velocity = positionDisplacement{}
	if m.Position.TileForRendering == m.Position.Tile {
		m.Position.Offset = m.Position.OffsetForRendering
	}
}

func PutMissile(w *hitWorld, m *Missile) {
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
		w.processCallbacks.MissilePreFlag = true
	}
}

func updateMissilePos(w *hitWorld, m *Missile) {
	// traveled >> 16: the whole-pixel part.
	pixelsTravelled := positionDisplacement{
		DeltaX: m.Position.Traveled.DeltaX >> fracBits,
		DeltaY: m.Position.Traveled.DeltaY >> fracBits,
	}

	tileOffset := screenToMissile(pixelsTravelled)
	m.Position.Tile = tilePosition{
		X: m.Position.Start.X + int(tileOffset.DeltaX),
		Y: m.Position.Start.Y + int(tileOffset.DeltaY),
	}

	m.Position.Offset = positionDisplacement{
		DeltaX: pixelsTravelled.DeltaX + worldToScreen(tileOffset).DeltaX,
		DeltaY: pixelsTravelled.DeltaY + worldToScreen(tileOffset).DeltaY,
	}

	absoluteLightOffset := screenToLight(pixelsTravelled)
	w.dispatchChangeLightOffset(m.Lid, positionDisplacement{
		DeltaX: absoluteLightOffset.DeltaX - tileOffset.DeltaX*8,
		DeltaY: absoluteLightOffset.DeltaY - tileOffset.DeltaY*8,
	})
}

// dispatchChangeLightOffset forwards the light-offset update through the
// sink with the port's int8 offset pair (missiles.cpp:223 takes
// DisplacementOf<int8_t>).
func (w *hitWorld) dispatchChangeLightOffset(lid int, offset positionDisplacement) {
	w.processCallbacks.ChangeLightOffset(lid, int8(offset.DeltaX), int8(offset.DeltaY))
}

func moveMissilePos(w *hitWorld, m *Missile) {
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
	target := tilePosition{X: m.Position.Tile.X + dx, Y: m.Position.Tile.Y + dy}
	if w.processCallbacks.IsTileAvailable(m.Source, target.X, target.Y) {
		m.Position.Tile = target
		// Displacement(moveDirection).worldToScreen() — upstream
		// builds the direction delta as a Displacement first
		// (missiles.cpp:243-252) so the offset delta is
		// worldToScreen of the tile delta.
		offset := worldToScreen(positionDisplacement{
			DeltaX: int32(dx), DeltaY: int32(dy),
		})
		m.Position.Offset.DeltaX += offset.DeltaX
		m.Position.Offset.DeltaY += offset.DeltaY
	}
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

// moveMissileCollisionOutcome distinguishes moveMissile's two early
// returns (missiles.cpp:625-627) for callers that need it; the port
// keeps upstream's plain bool return.
// moveMissile ports moveMissile: advance
// traveled by velocity, then walk every tile the missile crossed this
// tick — interpolating through skipped tiles with the 32×16 px increment
// — probing checkTile on each. A failed probe ends the move: with
// ifCheckTileFailsDontMoveToTile the missile never enters the tile (the
// velocity step or increment is reverted and stopMissile freezes it);
// without it the missile lands in the tile and the caller decides.
// [lint] revive/nestif: 1:1 upstream port — one flat interpolation loop
// over the per-tile probes; splitting it would
// sever the prevTile/traveled state the loop mutates.
func moveMissile(w *hitWorld, m *Missile, checkTile func(tilePosition) bool, // TODO remove nolint:revive // 1:1 port (missiles.cpp:552-628)
	ifCheckTileFailsDontMoveToTile bool) bool {
	prevTile := m.Position.Tile
	m.Position.Traveled.DeltaX += m.Position.Velocity.DeltaX
	m.Position.Traveled.DeltaY += m.Position.Velocity.DeltaY
	updateMissilePos(w, m)

	var possibleVisitTiles int
	if m.Position.Velocity.DeltaX == 0 || m.Position.Velocity.DeltaY == 0 {
		possibleVisitTiles = pathDistance(prevTile, m.Position.Tile)
	} else {
		possibleVisitTiles = manhattanDistance(prevTile, m.Position.Tile)
	}

	if possibleVisitTiles == 0 {
		return false
	}

	// Did the missile skip a tile?
	// [lint] nestif: 1:1 upstream interpolation loop.
	if possibleVisitTiles > 1 { // TODO remove nolint:nestif // 1:1 port (missiles.cpp:566-623)
		speed := positionDisplacement{
			DeltaX: absI32(m.Position.Velocity.DeltaX),
			DeltaY: absI32(m.Position.Velocity.DeltaY),
		}
		// float division + Displacement * float: the increment is
		// int32-truncated per component (missiles.cpp:570-571:
		// denominator = (2*speed.dy >= speed.dx) ? 2*speed.dy : speed.dx;
		// incVelocity = velocity * ((32 << 16) / denominator)).
		denominator := float32(speed.DeltaX)
		if 2*speed.DeltaY >= speed.DeltaX {
			denominator = float32(2 * speed.DeltaY)
		}
		factor := float32(32<<16) / denominator
		incVelocity := positionDisplacement{
			DeltaX: int32(float32(m.Position.Velocity.DeltaX) * factor),
			DeltaY: int32(float32(m.Position.Velocity.DeltaY) * factor),
		}
		traveled := positionDisplacement{
			DeltaX: m.Position.Traveled.DeltaX - m.Position.Velocity.DeltaX,
			DeltaY: m.Position.Traveled.DeltaY - m.Position.Velocity.DeltaY,
		}
		// Adjust the traveled vector to start on the next smallest
		// multiple of incVelocity: C++ int
		// division truncates toward zero.
		if incVelocity.DeltaY != 0 {
			traveled.DeltaY = (traveled.DeltaY / incVelocity.DeltaY) * incVelocity.DeltaY
		}
		if incVelocity.DeltaX != 0 {
			traveled.DeltaX = (traveled.DeltaX / incVelocity.DeltaX) * incVelocity.DeltaX
		}
		for {
			initialDiff := positionDisplacement{
				DeltaX: m.Position.Traveled.DeltaX - traveled.DeltaX,
				DeltaY: m.Position.Traveled.DeltaY - traveled.DeltaY,
			}
			traveled.DeltaX += incVelocity.DeltaX
			traveled.DeltaY += incVelocity.DeltaY
			incDiff := positionDisplacement{
				DeltaX: m.Position.Traveled.DeltaX - traveled.DeltaX,
				DeltaY: m.Position.Traveled.DeltaY - traveled.DeltaY,
			}

			// we are at the original calculated position => resume
			// with normal logic
			if (initialDiff.DeltaX < 0) != (incDiff.DeltaX < 0) {
				break
			}
			if (initialDiff.DeltaY < 0) != (incDiff.DeltaY < 0) {
				break
			}

			// calculate in-between tile
			pixelsTraveled := positionDisplacement{
				DeltaX: traveled.DeltaX >> fracBits,
				DeltaY: traveled.DeltaY >> fracBits,
			}
			tileOffset := screenToMissile(pixelsTraveled)
			tile := tilePosition{
				X: m.Position.Start.X + int(tileOffset.DeltaX),
				Y: m.Position.Start.Y + int(tileOffset.DeltaY),
			}

			// we haven't quite reached the missile's current position
			// but we can break early to avoid checking collisions in
			// this tile twice
			if tile == m.Position.Tile {
				break
			}

			// skip collision logic if the missile is on a corner
			// between tiles
			if pixelsTraveled.DeltaY%16 == 0 &&
				pixelsTraveled.DeltaX%32 == 0 &&
				absI32(pixelsTraveled.DeltaY/16)%2 != absI32(pixelsTraveled.DeltaX/32)%2 {
				continue
			}

			// don't call checkTile more than once for a tile
			if prevTile == tile {
				continue
			}

			prevTile = tile

			if !checkTile(tile) {
				m.Position.Traveled = traveled
				if ifCheckTileFailsDontMoveToTile {
					m.Position.Traveled.DeltaX -= incVelocity.DeltaX
					m.Position.Traveled.DeltaY -= incVelocity.DeltaY
					updateMissilePos(w, m)
					stopMissile(m)
				} else {
					updateMissilePos(w, m)
				}
				return true
			}
		}
	}

	if !checkTile(m.Position.Tile) && ifCheckTileFailsDontMoveToTile {
		m.Position.Traveled.DeltaX -= m.Position.Velocity.DeltaX
		m.Position.Traveled.DeltaY -= m.Position.Velocity.DeltaY
		updateMissilePos(w, m)
		stopMissile(m)
	}

	return true
}

func absI32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func damageTypeFor(id missileID) damageType {
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

func moveMissileAndCheckMissileCol(w *hitWorld, m *Missile, damageType damageType, // TODO remove nolint:revive // 1:1 upstream flags (missiles.cpp:636)
	minDamage, maxDamage int, ignoreStart, ifCollidesDontMoveToHitTile bool) {
	checkTile := func(tile tilePosition) bool {
		if ignoreStart && m.Position.Start == tile {
			return true
		}

		CheckMissileCol(w, m, tile, damageType, minDamage, maxDamage, false, false, nil)

		if m.Duration != 0 {
			return true
		}

		if m.HitFlag && m.MovementDistribution == MovementBlockable {
			return false
		}

		return !IsMissileBlockedByTile(w.Grid, tile.X, tile.Y)
	}

	tileChanged := moveMissile(w, m, checkTile, ifCollidesDontMoveToHitTile)

	tileTargetHash := int16(w.DMonster[m.Position.Tile.Y][m.Position.Tile.X] ^
		w.DPlayer[m.Position.Tile.Y][m.Position.Tile.X])

	// missile didn't change the tile... check that we perform
	// CheckMissileCol only once for any monster/player to avoid
	// multiple hits for slow missiles
	if !tileChanged && m.LastCollisionTargetHash != tileTargetHash {
		CheckMissileCol(w, m, m.Position.Tile, damageType, minDamage, maxDamage, false, false, nil)
	}

	// remember what target CheckMissileCol was checked against
	m.LastCollisionTargetHash = tileTargetHash
}

func ProcessArrow(w *hitWorld, m *Missile) {
	m.Duration--
	m.Dist++

	// Source damage dispatch: player
	// _pIMinDam/_pIMaxDam, monster minDamage/maxDamage, trap
	// currlevel/2*currlevel — resolved by the injected seam.
	mind, maxd := w.processCallbacks.SourceDamage(m)

	moveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), mind, maxd, true, false)
	if m.Duration == 0 {
		m.DelFlag = true
	}
	PutMissile(w, m)
}

func ProcessGenericProjectile(w *hitWorld, m *Missile) {
	m.Duration--

	moveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), m.Dam, m.Dam, true, true)
	if m.Duration == 0 { // TODO remove nolint:nestif // 1:1 port (missiles.cpp:2982-3013)
		m.DelFlag = true
		dst := tilePosition{X: 0, Y: 0}
		dir := Direction(m.MFNum)
		switch m.Type {
		case Firebolt, MagmaBall:
			w.processCallbacks.SpawnMissile(m.Position.Tile, dst, dir, MagmaBallExplosion,
				m.caster, m.Source, m)
		case BloodStar:
			w.processCallbacks.SpawnMissile(m.Position.Tile, dst, dir, BloodStarExplosion,
				m.caster, m.Source, m)
		case Acid:
			w.processCallbacks.SpawnMissile(m.Position.Tile, dst, dir, AcidSplat,
				m.caster, m.Source, m)
		case OrangeFlare:
			w.processCallbacks.SpawnMissile(m.Position.Tile, dst, dir, OrangeExplosion,
				m.caster, m.Source, m)
		case BlueFlare:
			w.processCallbacks.SpawnMissile(m.Position.Tile, dst, dir, BlueExplosion,
				m.caster, m.Source, m)
		case RedFlare:
			w.processCallbacks.SpawnMissile(m.Position.Tile, dst, dir, RedExplosion,
				m.caster, m.Source, m)
		case YellowFlare:
			w.processCallbacks.SpawnMissile(m.Position.Tile, dst, dir, YellowExplosion,
				m.caster, m.Source, m)
		case BlueFlare2:
			w.processCallbacks.SpawnMissile(m.Position.Tile, dst, dir, BlueExplosion2,
				m.caster, m.Source, m)
		default:
			// base rows outside the switch never spawn an explosion
			// upstream.
		}
		if m.Lid != NoLight {
			w.processCallbacks.AddUnLight(m.Lid)
		}
		PutMissile(w, m)
	} else {
		if m.Position.Tile != (tilePosition{X: m.Var1, Y: m.Var2}) {
			m.Var1 = m.Position.Tile.X
			m.Var2 = m.Position.Tile.Y
			if m.Lid != NoLight {
				w.processCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
			}
		}
		PutMissile(w, m)
	}
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
func (m *Missile) setAnimation(g graphicID) {
	m.AnimType = int(g)
}

func ProcessElementalArrow(w *hitWorld, m *Missile) {
	m.Duration--

	animType := graphicID(m.AnimType)
	if animType != GraphicChargedBolt && animType != GraphicMagmaBallExplosion {
		processElementalArrowFlight(w, m)
	} else {
		// Animation phase: the light rides
		// the burn-out frames.
		w.processCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(m.AnimFrame+5))
	}

	// Shared tail: duration 0 — the explosion
	// burn-out ran out — deletes the missile and removes the light.
	if m.Duration == 0 {
		m.DelFlag = true
		w.processCallbacks.AddUnLight(m.Lid)
	}
	PutMissile(w, m)
}

func elementalTrapDamage(w *hitWorld) (minDamage, maxDamage int) {
	return w.processCallbacks.GenerateRnd(10) + 1 + w.processCallbacks.Currlevel,
		w.processCallbacks.GenerateRnd(10) + 1 + w.processCallbacks.Currlevel*2
}

// elementalPack bundles the explosion sheet, element and damage dice of one
// elemental arrow row (language-artifact wrapper; upstream returns 4 locals).
type elementalPack struct {
	graphic        graphicID
	damageType     damageType
	minDam, maxDam int
}

func elementalExplosionPack(w *hitWorld, m *Missile) elementalPack {
	switch m.Type {
	case LightningArrow:
		return elementalPackFor(w, m, GraphicChargedBolt, DamageLightning)
	case FireArrow:
		return elementalPackFor(w, m, GraphicMagmaBallExplosion, DamageFire)
	default:
		panic("missiles: ProcessElementalArrow on non-elemental type")
	}
}

func elementalPackFor(w *hitWorld, m *Missile, eAnim graphicID, damageType damageType) elementalPack {
	if m.IsTrap() {
		eMind, eMaxd := elementalTrapDamage(w)
		return elementalPack{graphic: eAnim, damageType: damageType, minDam: eMind, maxDam: eMaxd}
	}
	eMind, eMaxd := w.processCallbacks.ElementDamage(m, damageType)
	return elementalPack{graphic: eAnim, damageType: damageType, minDam: eMind, maxDam: eMaxd}
}

func processElementalArrowFlight(w *hitWorld, m *Missile) {
	m.Dist++
	var mind, maxd int
	if m.IsTrap() {
		mind, maxd = elementalTrapDamage(w)
	} else {
		mind, maxd = w.processCallbacks.SourceDamage(m)
	}
	moveMissileAndCheckMissileCol(w, m, DamagePhysical, mind, maxd, true, false)

	if m.Duration == 0 {
		// Explosion transition.
		m.setDefaultFrameGroup()
		m.Duration = m.AnimLen - 1
		stopMissile(m)
		pack := elementalExplosionPack(w, m)
		m.setAnimation(pack.graphic)
		// The damage re-probe passes dontDeleteOnCollision = true
		// keepOnCollision.
		CheckMissileCol(w, m, m.Position.Tile, pack.damageType, pack.minDam, pack.maxDam,
			false, true, nil)
		return
	}
	if m.Position.Tile != (tilePosition{X: m.Var1, Y: m.Var2}) {
		// Flight tile tracker: radius 5.
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		w.processCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 5)
	}
}

func fireballImpactRing(w *hitWorld, m *Missile, mp tilePosition, damageType damageType, minDam, maxDam int) bool {
	offsets := [9]Direction{NoDirection, SouthWest, NorthEast, SouthEast,
		East, South, NorthWest, West, North}
	hit := false
	for _, off := range offsets {
		dx, dy := directionDeltaOf(off)
		px, py := mp.X+dx, mp.Y+dy
		if w.processCallbacks.CheckBlock(m.Position.Start.X, m.Position.Start.Y, px, py) {
			continue
		}
		res := CheckMissileCol(w, m, tilePosition{X: px, Y: py}, damageType,
			minDam, maxDam, false, true, nil)
		if res.MonsterHit || res.PlayerHit {
			hit = true
		}
	}
	return hit
}

func fireballWallAdjust(w *hitWorld, m *Missile, mp tilePosition) {
	if !w.processCallbacks.TransAt(mp.X, mp.Y) ||
		(m.Position.Velocity.DeltaX < 0 &&
			((w.processCallbacks.TransAt(mp.X, mp.Y+1) && w.Grid.IsTileSolid(mp.X, mp.Y+1)) ||
				(w.processCallbacks.TransAt(mp.X, mp.Y-1) && w.Grid.IsTileSolid(mp.X, mp.Y-1)))) {
		sdx, sdy := directionDeltaOf(South)
		m.Position.Tile = tilePosition{X: m.Position.Tile.X + sdx, Y: m.Position.Tile.Y + sdy}
		m.Position.Offset.DeltaY -= 32
	}
	if m.Position.Velocity.DeltaY > 0 &&
		((w.processCallbacks.TransAt(mp.X+1, mp.Y) && w.Grid.IsTileSolid(mp.X+1, mp.Y)) ||
			(w.processCallbacks.TransAt(mp.X-1, mp.Y) && w.Grid.IsTileSolid(mp.X-1, mp.Y))) {
		m.Position.Offset.DeltaY -= 32
	}
	if m.Position.Velocity.DeltaX > 0 &&
		((w.processCallbacks.TransAt(mp.X, mp.Y+1) && w.Grid.IsTileSolid(mp.X, mp.Y+1)) ||
			(w.processCallbacks.TransAt(mp.X, mp.Y-1) && w.Grid.IsTileSolid(mp.X, mp.Y-1))) {
		m.Position.Offset.DeltaX -= 32
	}
}

func ProcessFireball(w *hitWorld, m *Missile) {
	m.Duration--

	if graphicID(m.AnimType) == GraphicBigExplosion {
		if m.Duration == 0 {
			m.DelFlag = true
			w.processCallbacks.AddUnLight(m.Lid)
		}
		PutMissile(w, m)
		return
	}

	minDam, maxDam := m.Dam, m.Dam
	if m.caster != int(targetMonsters) {
		minDam, maxDam = w.processCallbacks.SourceDamage(m)
	}
	damageType := damageTypeFor(m.Type)
	moveMissileAndCheckMissileCol(w, m, damageType, minDam, maxDam, true, false)

	if m.Duration == 0 {
		mp := m.Position.Tile
		w.processCallbacks.ChangeLight(m.Lid, mp.X, mp.Y, uint8(m.AnimFrame))
		fireballImpactRing(w, m, mp, damageType, minDam, maxDam)

		fireballWallAdjust(w, m, mp)

		m.setDefaultFrameGroup()
		m.setAnimation(GraphicBigExplosion)
		m.Duration = m.AnimLen - 1
		m.Position.Velocity = positionDisplacement{}
	} else if m.Position.Tile != (tilePosition{X: m.Var1, Y: m.Var2}) {
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		w.processCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
	}
	PutMissile(w, m)
}

// expLightTable ports ExpLight: the explosion
// light radius ramp, 15 entries.
var expLightTable = [15]int{9, 10, 11, 12, 11, 10, 8, 6, 4, 2, 1, 0, 0, 0, 0}

func ProcessMissileExplosion(w *hitWorld, m *Missile) {
	m.Duration--

	if m.Duration == 0 {
		m.DelFlag = true
		w.processCallbacks.AddUnLight(m.Lid)
		return
	}
	if m.Var1 == 0 {
		// First tick: seed the explosion light.
		m.Lid = w.processCallbacks.AddLight(m.Position.Tile.X, m.Position.Tile.Y, 9)
	} else {
		w.processCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(expLightTable[m.Var1]))
	}
	m.Var1++
	PutMissile(w, m)
}

// bPathTable ports BPath: the ChargedBolt
// steering draw table, 16 entries of -1/0/1 (left/straight/right).
var bPathTable = [16]int{-1, 0, 1, -1, 0, 1, -1, -1, 0, 0, 1, 1, 0, 1, -1, 0}

func ProcessChargedBolt(w *hitWorld, m *Missile) {
	m.Duration--

	if graphicID(m.AnimType) != GraphicLightning {
		if m.Var3 == 0 {
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
			target := tilePosition{
				X: m.Position.Tile.X + mdx,
				Y: m.Position.Tile.Y + mdy,
			}
			updateMissileVelocity(m, target, 8)
			m.Var3 = 16
		} else {
			m.Var3--
		}
		moveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), m.Dam, m.Dam, false, false)

		if m.HitFlag {
			m.Var1 = 8
			m.setDefaultFrameGroup()
			m.Position.Offset = positionDisplacement{}
			m.Position.Velocity = positionDisplacement{}
			m.setAnimation(GraphicLightning)
			m.Duration = m.AnimLen
		}
		w.processCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, uint8(m.Var1))
	}

	if m.Duration == 0 {
		m.DelFlag = true
		w.processCallbacks.AddUnLight(m.Lid)
	}
	PutMissile(w, m)
}

func ProcessHolyBolt(w *hitWorld, m *Missile) {
	m.Duration--

	if graphicID(m.AnimType) != GraphicHolyBoltExplosion {
		processHolyBoltFlight(w, m)
	} else {
		// Animation phase: light rides the
		// burn-out frames.
		w.processCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(m.AnimFrame+7))
		if m.Duration == 0 {
			m.DelFlag = true
			w.processCallbacks.AddUnLight(m.Lid)
		}
	}
	PutMissile(w, m)
}

func processHolyBoltFlight(w *hitWorld, m *Missile) {
	dam := m.Dam
	moveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), dam, dam, true, true)
	if m.Duration == 0 {
		m.setDefaultFrameGroup()
		m.setAnimation(GraphicHolyBoltExplosion)
		m.Duration = m.AnimLen - 1
		stopMissile(m)
		return
	}
	if m.Position.Tile != (tilePosition{X: m.Var1, Y: m.Var2}) {
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		w.processCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
	}
}
