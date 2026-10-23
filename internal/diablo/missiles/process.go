// Package missiles: the projectile motion pipeline and the
// ProcessArrow / ProcessGenericProjectile ports — UpdateMissileVelocity
// PutMissile, UpdateMissilePos, MoveMissilePos, MoveMissile
// MoveMissileAndCheckMissileCol
// and the two Process fns
// (devilGo-mwlm slice 3a-beta).
// Engine boundary: upstream reads the lighting engine (ChangeLightOffset
// ChangeLight, AddUnLight, AddLight —, spawns explosions
// through AddMissile, resolves tile availability through
// IsTileAvailable, draws through GenerateRnd and
// reads currlevel/dTransVal; those systems land with later beans, so they
// are injected through the ProcessCallbacks sink carried on HitWorld (the
// HitCallbacks pattern, collision.go — no-op defaults via
// NoopProcessCallbacks; tests wire recording spies).
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// NoLight mirrors the NO_LIGHT sentinel.
const NoLight = -1

// ProcessCallbacks is the engine-boundary sink the process functions
// read instead of the upstream globals (the HitCallbacks pattern
// collision.go): the lighting engine, the explosion
// spawn chain (AddMissile,, the source/elemental damage
// reads (Players/Monsters/currlevel), the tile predicates (IsTileAvailable
// ; dTransVal, gendung.h; CheckBlock) and the
// GenerateRnd draw. Carried on HitWorld; NoopProcessCallbacks returns the
// no-op defaults (tests wire recording spies).
type ProcessCallbacks struct {
	// MissilePreFlag ports the engine global
	// PutMissile raises it when the missile carries _miPreFlag
	// ; ProcessMissiles clears it per frame
	// . Sink state, not a func.
	MissilePreFlag bool
	// ChangeLightOffset ports ChangeLightOffset
	// params lid, offsetX, offsetY) — UpdateMissilePos feeds it the
	// missile's light-grid offset.
	ChangeLightOffset func(lid int, offsetX, offsetY int8)
	// ChangeLight ports ChangeLight, params lid, x, y
	// radius) — the light trackers move a light with the missile
	// et al).
	ChangeLight func(lid, x, y int, radius uint8)
	// AddUnLight ports AddUnLight, param lid) — the light
	// removal of the burn-out branches et al).
	AddUnLight func(lid int)
	// AddLight ports AddLight, params x, y, radius)
	// registers a new light and returns its id — the explosion light
	// ramp seeds _mlid from it.
	AddLight func(x, y int, radius uint8) int
	// SpawnMissile forwards to AddMissile;
	// params: src, dst, dir, spawnID, caster, source, parent (the
	// AddMissile argument pack). TODO(devilGo-mwlm slice 3c): real
	// spawn chains once the missile container and the explosion
	// Add* bodies exist.
	SpawnMissile func(src, dst TilePosition, dir Direction,
		spawnID MissileID, caster, source int, parent *Missile)
	// SourceDamage resolves the source damage pack
	// 2953): player _pIMinDam/_pIMaxDam, monster minDamage/maxDamage
	// trap currlevel dice. TODO(devilGo-mwlm slice 3b): real damage
	// sourcing once the player/monster tables land.
	SourceDamage func(m *Missile) (minDamage, maxDamage int)
	// ElementDamage ports the elemental player-damage read of the
	// elemental-arrow explosion switch
	// _pILMinDam/_pILMaxDam (Lightning) and _pIFMinDam/_pIFMaxDam (Fire).
	ElementDamage func(m *Missile, element DamageType) (minDamage, maxDamage int)
	// IsTileAvailable ports IsTileAvailable
	// bounds + non-solid + monster occupancy checks; params
	// sourceMonster, x, y.
	IsTileAvailable func(sourceMonster, x, y int) bool
	// GenerateRnd ports GenerateRnd, param n)
	// [0, n) draw — the trap damage rolls.
	GenerateRnd func(n int) int
	// Currlevel ports the dungeon-level global
	// the trap damage base.
	Currlevel int
	// TransAt ports the tile-transparency lookup (upstream reads
	// dTransVal/TransList, gendung.h): fireball wall-adjust only
	// .
	TransAt func(x, y int) bool
	// CheckBlock ports CheckBlock (missiles.cpp pathing helper): true
	// when the straight line from→to is blocked — the fireball ring
	// only probes tiles it can see.
	CheckBlock func(fromX, fromY, toX, toY int) bool
}

// NoopProcessCallbacks returns the no-op sink defaults until the
// lighting/damage/tile beans land (the pre-sink hook defaults, 1:1).
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
	}
}

// UpdateMissileVelocity ports UpdateMissileVelocity
// velocity = worldToNormalScreen(tile
// destination) * velocityInPixels; a missile already on the destination
// tile stops (velocity = {0, 0}).
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

// StopMissile ports MissilePosition::StopMissile
// zero the velocity and snap the sprite offset to the last renderer
// position when the renderer tile matches (the render fields are written
// by the renderer,.
func StopMissile(m *Missile) {
	m.Position.Velocity = PositionDisplacement{}
	if m.Position.TileForRendering == m.Position.Tile {
		m.Position.Offset = m.Position.OffsetForRendering
	}
}

// PutMissile ports PutMissile: register the
// missile's tile in the dFlags lookup table (Missile bit, plus the
// FireWall/LightningWall bits by type), raising DelFlag when the tile is
// out of bounds and the sink's MissilePreFlag for pre-flagged missiles.
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

// UpdateMissilePos ports UpdateMissilePos
// re-derive tile, sprite offset and the light offset from the traveled
// vector.
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

// dispatchChangeLightOffset forwards the light-offset update through the
// sink with the port's int8 offset pair takes
// DisplacementOf<int8_t>).
func (w *HitWorld) dispatchChangeLightOffset(lid int, offset PositionDisplacement) {
	w.ProcessCallbacks.ChangeLightOffset(lid, int8(offset.DeltaX), int8(offset.DeltaY))
}

// MoveMissilePos ports MoveMissilePos
// "Dodgy hack used to correct the position for charging monsters" —
// southward-facing sprites shift one tile south; the move respects tile
// availability via the injected seam.
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
		// Displacement(moveDirection).worldToScreen() — upstream
		// builds the direction delta as a Displacement first
		// so the offset delta is
		// worldToScreen of the tile delta.
		offset := WorldToScreen(PositionDisplacement{
			DeltaX: int32(dx), DeltaY: int32(dy),
		})
		m.Position.Offset.DeltaX += offset.DeltaX
		m.Position.Offset.DeltaY += offset.DeltaY
	}
}

// directionDeltaOf mirrors DisplacementOf::fromDirection
// the tile delta per Direction
// enum order (South, SouthWest, West, NorthWest, North, NorthEast, East
// SouthEast).
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

// moveMissileCollisionOutcome distinguishes MoveMissile's two early
// returns vs 627) for callers that need it; the port
// keeps upstream's plain bool return.
// moveMissileCollisionOutcome distinguishes MoveMissile's two early
// returns vs 627) for callers that need it; the port
// keeps upstream's plain bool return.
// MoveMissile ports MoveMissile: advance
// traveled by velocity, then walk every tile the missile crossed this
// tick — interpolating through skipped tiles with the 32×16 px increment
// — probing checkTile on each. A failed probe ends the move: with
// ifCheckTileFailsDontMoveToTile the missile never enters the tile (the
// velocity step or increment is reverted and StopMissile freezes it);
// without it the missile lands in the tile and the caller decides.
// [lint] revive/nestif: 1:1 upstream port — one flat interpolation loop
// over the per-tile probes; splitting it would
// sever the prevTile/traveled state the loop mutates.
func MoveMissile(w *HitWorld, m *Missile, checkTile func(TilePosition) bool, //nolint:revive // 1:1 port (missiles.cpp:552-628)
	ifCheckTileFailsDontMoveToTile bool) bool {
	prevTile := m.Position.Tile
	m.Position.Traveled.DeltaX += m.Position.Velocity.DeltaX
	m.Position.Traveled.DeltaY += m.Position.Velocity.DeltaY
	UpdateMissilePos(w, m)

	var possibleVisitTiles int
	if m.Position.Velocity.DeltaX == 0 || m.Position.Velocity.DeltaY == 0 {
		possibleVisitTiles = PathDistance(prevTile, m.Position.Tile)
	} else {
		possibleVisitTiles = ManhattanDistance(prevTile, m.Position.Tile)
	}

	if possibleVisitTiles == 0 {
		return false
	}

	// Did the missile skip a tile?
	// [lint] nestif: 1:1 upstream interpolation loop.
	if possibleVisitTiles > 1 { //nolint:nestif // 1:1 port (missiles.cpp:566-623)
		speed := PositionDisplacement{
			DeltaX: absI32(m.Position.Velocity.DeltaX),
			DeltaY: absI32(m.Position.Velocity.DeltaY),
		}
		// float division + Displacement * float: the increment is
		// int32-truncated per component
		// 311-316;.
		denominator := float32(speed.DeltaX)
		if 2*speed.DeltaY >= speed.DeltaX {
			denominator = float32(2 * speed.DeltaY)
		}
		factor := float32(32<<16) / denominator
		incVelocity := PositionDisplacement{
			DeltaX: int32(float32(m.Position.Velocity.DeltaX) * factor),
			DeltaY: int32(float32(m.Position.Velocity.DeltaY) * factor),
		}
		traveled := PositionDisplacement{
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
				break
			}
			if (initialDiff.DeltaY < 0) != (incDiff.DeltaY < 0) {
				break
			}

			// calculate in-between tile
			pixelsTraveled := PositionDisplacement{
				DeltaX: traveled.DeltaX >> fracBits,
				DeltaY: traveled.DeltaY >> fracBits,
			}
			tileOffset := ScreenToMissile(pixelsTraveled)
			tile := TilePosition{
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
					UpdateMissilePos(w, m)
					StopMissile(m)
				} else {
					UpdateMissilePos(w, m)
				}
				return true
			}
		}
	}

	if !checkTile(m.Position.Tile) && ifCheckTileFailsDontMoveToTile {
		m.Position.Traveled.DeltaX -= m.Position.Velocity.DeltaX
		m.Position.Traveled.DeltaY -= m.Position.Velocity.DeltaY
		UpdateMissilePos(w, m)
		StopMissile(m)
	}

	return true
}

// absI32 is std::abs for the port's int32 components.
func absI32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// damageTypeFor ports MissileData::damageType()
// the low 3 bits of the row's flags
// are the DamageType ordinal — the base rows resolve to Physical (Arrow)
// Fire (Firebolt/MagmaBall), Magic (BloodStar), Acid (Acid), per the
// on-disk misdat.tsv flags column.
func damageTypeFor(id MissileID) DamageType {
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
// move the missile, then (only when the
// tile did not change) re-probe the current tile at most once per
// occupant — lastCollisionTargetHash remembers the dMonster^dPlayer cell
// so slow missiles do not multi-hit one entity.
func MoveMissileAndCheckMissileCol(w *HitWorld, m *Missile, damageType DamageType, //nolint:revive // 1:1 upstream flags (missiles.cpp:636)
	minDamage, maxDamage int, ignoreStart, ifCollidesDontMoveToHitTile bool) {
	checkTile := func(tile TilePosition) bool {
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

	tileChanged := MoveMissile(w, m, checkTile, ifCollidesDontMoveToHitTile)

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

// ProcessArrow ports ProcessArrow: one
// tick of arrow flight — countdown, distance tally, source-scaled damage
// roll (via the seam upstream reads Players/Monsters/currlevel for), the
// move, deletion at duration 0 and the dFlags registration.
func ProcessArrow(w *HitWorld, m *Missile) {
	m.Duration--
	m.Dist++

	// Source damage dispatch: player
	// _pIMinDam/_pIMaxDam, monster minDamage/maxDamage, trap
	// currlevel/2*currlevel — resolved by the injected seam.
	mind, maxd := w.ProcessCallbacks.SourceDamage(m)

	MoveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), mind, maxd, true, false)
	if m.Duration == 0 {
		m.DelFlag = true
	}
	PutMissile(w, m)
}

// ProcessGenericProjectile ports ProcessGenericProjectile
// one tick of bolt flight. On expiry
// the missile deletes itself, spawns its explosion through the seam and
// removes its light; otherwise it tracks the tile into var1/var2 and
// moves its light along.
func ProcessGenericProjectile(w *HitWorld, m *Missile) {
	m.Duration--

	MoveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), m.Dam, m.Dam, true, true)
	if m.Duration == 0 { //nolint:nestif // 1:1 port (missiles.cpp:2982-3013)
		m.DelFlag = true
		dst := TilePosition{X: 0, Y: 0}
		dir := Direction(m.MFNum)
		switch m.Type {
		case Firebolt, MagmaBall:
			w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, MagmaBallExplosion,
				m.Caster, m.Source, m)
		case BloodStar:
			w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, BloodStarExplosion,
				m.Caster, m.Source, m)
		case Acid:
			w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, AcidSplat,
				m.Caster, m.Source, m)
		case OrangeFlare:
			w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, OrangeExplosion,
				m.Caster, m.Source, m)
		case BlueFlare:
			w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, BlueExplosion,
				m.Caster, m.Source, m)
		case RedFlare:
			w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, RedExplosion,
				m.Caster, m.Source, m)
		case YellowFlare:
			w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, YellowExplosion,
				m.Caster, m.Source, m)
		case BlueFlare2:
			w.ProcessCallbacks.SpawnMissile(m.Position.Tile, dst, dir, BlueExplosion2,
				m.Caster, m.Source, m)
		default:
			// base rows outside the switch never spawn an explosion
			// upstream.
		}
		if m.Lid != NoLight {
			w.ProcessCallbacks.AddUnLight(m.Lid)
		}
		PutMissile(w, m)
	} else {
		if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
			m.Var1 = m.Position.Tile.X
			m.Var2 = m.Position.Tile.Y
			if m.Lid != NoLight {
				w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
			}
		}
		PutMissile(w, m)
	}
}

// leftOf ports Left: a left turn on
// the 8-way compass.
func leftOf(facing Direction) Direction {
	return Direction((uint8(facing) + 7) % 8)
}

// rightOf ports Right: a right turn.
func rightOf(facing Direction) Direction {
	return Direction((uint8(facing) + 1) % 8)
}

// graphicSetDefaultFrameGroup ports Missile::setDefaultFrameGroup
// the frame group drops to 0. The
// setAnimation sheet swap is the AnimType store; animation-count state
// is animation-bean scope (TODO devilGo-mwlm animation slice).
func (m *Missile) setDefaultFrameGroup() {
	m.MFNum = 0
}

// setAnimation ports Missile::setAnimation — the sheet swap by graphic
// id _miAnimType store).
func (m *Missile) setAnimation(g GraphicID) {
	m.AnimType = int(g)
}

// ProcessElementalArrow ports ProcessElementalArrow
// the elemental bows' projectile —
// flies as a physical arrow; at duration 0 it swaps to its elemental
// explosion animation, re-rolls elemental damage through the seam, and
// rides the burn-out frames with the light following the frame.
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

// elementalTrapDamage ports the trap dice roll of the elemental-arrow
// branches: rnd(10)+1+currlevel and
// rnd(10)+1+currlevel*2.
func elementalTrapDamage(w *HitWorld) (minDamage, maxDamage int) {
	return w.ProcessCallbacks.GenerateRnd(10) + 1 + w.ProcessCallbacks.Currlevel,
		w.ProcessCallbacks.GenerateRnd(10) + 1 + w.ProcessCallbacks.Currlevel*2
}

// elementalPack bundles the explosion sheet, element and damage dice of one
// elemental arrow row (language-artifact wrapper; upstream returns 4 locals).
type elementalPack struct {
	graphic        GraphicID
	damageType     DamageType
	minDam, maxDam int
}

// elementalExplosionPack ports the explosion switch of
// ProcessElementalArrow: the animation sheet
// damage element and damage pack per missile type; an unknown type is a
// data-table bug (app_fatal upstream).
func elementalExplosionPack(w *HitWorld, m *Missile) elementalPack {
	switch m.Type {
	case LightningArrow:
		return elementalPackFor(w, m, GraphicChargedBolt, DamageLightning)
	case FireArrow:
		return elementalPackFor(w, m, GraphicMagmaBallExplosion, DamageFire)
	default:
		panic("missiles: ProcessElementalArrow on non-elemental type")
	}
}

// elementalPackFor resolves the damage pack of one elemental row: the
// player's elemental min/max through the seam, or the trap dice
// .
func elementalPackFor(w *HitWorld, m *Missile, eAnim GraphicID, damageType DamageType) elementalPack {
	if m.IsTrap() {
		eMind, eMaxd := elementalTrapDamage(w)
		return elementalPack{graphic: eAnim, damageType: damageType, minDam: eMind, maxDam: eMaxd}
	}
	eMind, eMaxd := w.ProcessCallbacks.ElementDamage(m, damageType)
	return elementalPack{graphic: eAnim, damageType: damageType, minDam: eMind, maxDam: eMaxd}
}

// processElementalArrowFlight ports the flight phase body of
// ProcessElementalArrow: the physical roll, the
// move, the duration-0 explosion transition and the radius-5 tile
// tracker.
func processElementalArrowFlight(w *HitWorld, m *Missile) {
	m.Dist++
	var mind, maxd int
	if m.IsTrap() {
		mind, maxd = elementalTrapDamage(w)
	} else {
		mind, maxd = w.ProcessCallbacks.SourceDamage(m)
	}
	MoveMissileAndCheckMissileCol(w, m, DamagePhysical, mind, maxd, true, false)

	if m.Duration == 0 {
		// Explosion transition.
		m.setDefaultFrameGroup()
		m.Duration = m.AnimLen - 1
		StopMissile(m)
		pack := elementalExplosionPack(w, m)
		m.setAnimation(pack.graphic)
		// The damage re-probe passes dontDeleteOnCollision = true
		// keepOnCollision.
		CheckMissileCol(w, m, m.Position.Tile, pack.damageType, pack.minDam, pack.maxDam,
			false, true, nil)
		return
	}
	if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
		// Flight tile tracker: radius 5.
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 5)
	}
}

// fireballImpactRing probes the 8-tile ring on impact.
// Returns true if the missile hit something in the ring.
func fireballImpactRing(w *HitWorld, m *Missile, mp TilePosition, damageType DamageType, minDam, maxDam int) bool {
	offsets := [9]Direction{NoDirection, SouthWest, NorthEast, SouthEast,
		East, South, NorthWest, West, North}
	hit := false
	for _, off := range offsets {
		dx, dy := directionDeltaOf(off)
		px, py := mp.X+dx, mp.Y+dy
		if w.ProcessCallbacks.CheckBlock(m.Position.Start.X, m.Position.Start.Y, px, py) {
			continue
		}
		res := CheckMissileCol(w, m, TilePosition{X: px, Y: py}, damageType,
			minDam, maxDam, false, true, nil)
		if res.MonsterHit || res.PlayerHit {
			hit = true
		}
	}
	return hit
}

// fireballWallAdjust applies the transparency/solid nudges.
func fireballWallAdjust(w *HitWorld, m *Missile, mp TilePosition) {
	if !w.ProcessCallbacks.TransAt(mp.X, mp.Y) ||
		(m.Position.Velocity.DeltaX < 0 &&
			((w.ProcessCallbacks.TransAt(mp.X, mp.Y+1) && w.Grid.IsTileSolid(mp.X, mp.Y+1)) ||
				(w.ProcessCallbacks.TransAt(mp.X, mp.Y-1) && w.Grid.IsTileSolid(mp.X, mp.Y-1)))) {
		sdx, sdy := directionDeltaOf(South)
		m.Position.Tile = TilePosition{X: m.Position.Tile.X + sdx, Y: m.Position.Tile.Y + sdy}
		m.Position.Offset.DeltaY -= 32
	}
	if m.Position.Velocity.DeltaY > 0 &&
		((w.ProcessCallbacks.TransAt(mp.X+1, mp.Y) && w.Grid.IsTileSolid(mp.X+1, mp.Y)) ||
			(w.ProcessCallbacks.TransAt(mp.X-1, mp.Y) && w.Grid.IsTileSolid(mp.X-1, mp.Y))) {
		m.Position.Offset.DeltaY -= 32
	}
	if m.Position.Velocity.DeltaX > 0 &&
		((w.ProcessCallbacks.TransAt(mp.X, mp.Y+1) && w.Grid.IsTileSolid(mp.X, mp.Y+1)) ||
			(w.ProcessCallbacks.TransAt(mp.X, mp.Y-1) && w.Grid.IsTileSolid(mp.X, mp.Y-1))) {
		m.Position.Offset.DeltaX -= 32
	}
}

// ProcessFireball ports ProcessFireball
// bolt flight, then on impact an 8-tile ring re-probe behind the wall
// transparency adjustment, then the BigExplosion burn-out frames.
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
	damageType := damageTypeFor(m.Type)
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

// expLightTable ports ExpLight: the explosion
// light radius ramp, 15 entries.
var expLightTable = [15]int{9, 10, 11, 12, 11, 10, 8, 6, 4, 2, 1, 0, 0, 0, 0}

// ProcessMissileExplosion ports ProcessMissileExplosion
// the static explosion animation — the
// light ramps down ExpLight while the frames play, then the light goes.
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

// bPathTable ports BPath: the ChargedBolt
// steering draw table, 16 entries of -1/0/1 (left/straight/right).
var bPathTable = [16]int{-1, 0, 1, -1, 0, 1, -1, -1, 0, 0, 1, 1, 0, 1, -1, 0}

// ProcessChargedBolt ports ProcessChargedBolt
// the weaving bolt — every 16 ticks it
// re-aims along BPath, and a hit freezes it into the Lightning
// burn-out, which only counts the duration down (the move, the collision
// probe and the light all sit inside the non-Lightning guard
// upstream,.
func ProcessChargedBolt(w *HitWorld, m *Missile) {
	m.Duration--

	if GraphicID(m.AnimType) != GraphicLightning {
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
			target := TilePosition{
				X: m.Position.Tile.X + mdx,
				Y: m.Position.Tile.Y + mdy,
			}
			UpdateMissileVelocity(m, target, 8)
			m.Var3 = 16
		} else {
			m.Var3--
		}
		MoveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), m.Dam, m.Dam, false, false)

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

	if m.Duration == 0 {
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}
	PutMissile(w, m)
}

// ProcessHolyBolt ports ProcessHolyBolt
// bolt flight, then the HolyBoltExplosion burn-out whose light rides the
// animation frames.
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

// processHolyBoltFlight ports the flight phase body of ProcessHolyBolt
// the move, the duration-0 transition into the
// explosion animation and the radius-8 tile tracker.
func processHolyBoltFlight(w *HitWorld, m *Missile) {
	dam := m.Dam
	MoveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), dam, dam, true, true)
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
