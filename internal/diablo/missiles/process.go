// Package missiles: the projectile motion pipeline and the
// ProcessArrow / ProcessGenericProjectile ports — UpdateMissileVelocity,
// PutMissile, UpdateMissilePos, MoveMissilePos, MoveMissile,
// MoveMissileAndCheckMissileCol (Source/missiles.cpp:173-251, 552-663)
// and the two Process fns (missiles.cpp:2932-2958, 2963-3016)
// (devilGo-mwlm slice 3a-beta).
//
// Engine boundary: upstream reads the lighting engine (ChangeLightOffset,
// ChangeLight, AddUnLight — lighting.h:61-65), spawns explosions through
// AddMissile (missiles.cpp:2819) and resolves tile availability through
// IsTileAvailable (tile_properties.cpp:47-59); those systems land with
// later beans, so they are injected as package-level function variables
// below (no-op defaults; tests wire recording spies).
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/world"
)

// NoLight mirrors the NO_LIGHT sentinel (Source/engine/lighting_defs.hpp:9).
const NoLight = -1

// MissilePreFlag ports the engine global (Source/missiles.cpp:74):
// PutMissile raises it when the missile carries _miPreFlag
// (missiles.cpp:209-210); ProcessMissiles clears it per frame
// (missiles.cpp:4214).
var MissilePreFlag bool

// Lighting hooks (lighting.h:61-65). Defaults are no-ops until the
// lighting bean lands.
var (
	// HookChangeLightOffset ports ChangeLightOffset (lighting.h:64,
	// params lid, offsetX, offsetY) — UpdateMissilePos feeds it the
	// missile's light-grid offset (missiles.cpp:222).
	HookChangeLightOffset = func(_ int, _ int8, _ int8) {}
	// HookChangeLight ports ChangeLight (lighting.h:65, params lid,
	// x, y, radius) — ProcessGenericProjectile moves a firebolt's
	// light with the missile (missiles.cpp:3010).
	HookChangeLight = func(_ int, _ int, _ int, _ uint8) {}
	// HookAddUnLight ports AddUnLight (lighting.h:61, param lid) —
	// the light removal of the projectile hit branch
	// (missiles.cpp:3003).
	HookAddUnLight = func(_ int) {}
)

// TODO(devilGo-mwlm slice 3c): real spawn chains — the hook forwards to
// AddMissile (missiles.cpp:2982-3001) once the missile container and the
// explosion Add* bodies exist.
// HookSpawnMissile params: src, dst, dir, spawnID, caster, source,
// parent (the AddMissile argument pack, missiles.cpp:2982-3001).
var HookSpawnMissile = func(_ TilePosition, _ TilePosition, _ Direction,
	_ MissileID, _ int, _ int, _ *Missile) {
}

// TODO(devilGo-mwlm slice 3b): real damage sourcing — the hook resolves
// sourcePlayer/_pIMinDam.._pIMaxDam, sourceMonster min/maxDamage and the
// trap branch once the player/monster tables land.
// HookSourceDamage param: m (the missile; upstream reads
// Players/Monsters/currlevel from its source fields).
var HookSourceDamage = func(_ *Missile) (minDamage, maxDamage int) {
	return 0, 0
}

// HookElementDamage ports the elemental player-damage read of the
// elemental-arrow explosion switch (missiles.cpp:2886-2891):
// _pILMinDam/_pILMaxDam (Lightning) and _pIFMinDam/_pIFMaxDam (Fire) —
// resolved by the injected seam keyed on the element.
var HookElementDamage = func(_ *Missile, _ DamageType) (minDamage, maxDamage int) {
	return 0, 0
}

// HookIsTileAvailable ports IsTileAvailable (Source/tile_properties.cpp:47-59):
// bounds + non-solid + monster occupancy checks, entity state injected.
// HookIsTileAvailable params: sourceMonster, x, y
// (tile_properties.cpp:47-59).
var HookIsTileAvailable = func(_ int, _ int, _ int) bool { return false }

// UpdateMissileVelocity ports UpdateMissileVelocity
// (Source/missiles.cpp:173-186): velocity = worldToNormalScreen(tile -
// destination) * velocityInPixels; a missile already on the destination
// tile stops (velocity = {0, 0}).
func UpdateMissileVelocity(m *Missile, destination TilePosition, velocityInPixels int) {
	m.Position.Velocity = PositionDisplacement{}

	if m.Position.Tile == destination {
		return
	}

	// Get the normalized vector in isometric projection
	// (missiles.cpp:181-182); upstream's Point difference is the
	// Displacement of (tile - destination).
	fixed16NormalVector := WorldToNormalScreen(PositionDisplacement{
		DeltaX: int32(m.Position.Tile.X - destination.X),
		DeltaY: int32(m.Position.Tile.Y - destination.Y),
	})

	// Multiplying by the target velocity gives us a scaled velocity
	// vector (missiles.cpp:185): Displacement * int — component-wise.
	m.Position.Velocity = PositionDisplacement{
		DeltaX: fixed16NormalVector.DeltaX * int32(velocityInPixels),
		DeltaY: fixed16NormalVector.DeltaY * int32(velocityInPixels),
	}
}

// StopMissile ports MissilePosition::StopMissile (Source/missiles.h:49-54):
// zero the velocity and snap the sprite offset to the last renderer
// position when the renderer tile matches (the render fields are written
// by the renderer, scrollrt.cpp:148-158).
func StopMissile(m *Missile) {
	m.Position.Velocity = PositionDisplacement{}
	if m.Position.TileForRendering == m.Position.Tile {
		m.Position.Offset = m.Position.OffsetForRendering
	}
}

// PutMissile ports PutMissile (Source/missiles.cpp:191-211): register the
// missile's tile in the dFlags lookup table (Missile bit, plus the
// FireWall/LightningWall bits by type), raising DelFlag when the tile is
// out of bounds and the MissilePreFlag global for pre-flagged missiles.
func PutMissile(g *world.Grid, m *Missile) {
	position := m.Position.Tile

	if !world.InDungeonBounds(position.X, position.Y) {
		m.DelFlag = true
	}

	if m.DelFlag {
		return
	}

	flags := &g.DFlags[position.Y][position.X]
	*flags |= world.FlagMissile
	if m.Type == FireWall {
		*flags |= world.FlagMissileFire
	}
	if m.Type == LightningWall {
		*flags |= world.FlagMissileLight
	}

	if m.PreFlag {
		MissilePreFlag = true
	}
}

// UpdateMissilePos ports UpdateMissilePos (Source/missiles.cpp:213-224):
// re-derive tile, sprite offset and the light offset from the traveled
// vector.
func UpdateMissilePos(m *Missile) {
	// traveled >> 16: the whole-pixel part (missiles.cpp:214).
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
	ChangeLightOffset(m.Lid, PositionDisplacement{
		DeltaX: absoluteLightOffset.DeltaX - tileOffset.DeltaX*8,
		DeltaY: absoluteLightOffset.DeltaY - tileOffset.DeltaY*8,
	})
}

// ChangeLightOffset dispatches the lighting hook with the port's int8
// offset pair (lighting.h:64 takes DisplacementOf<int8_t>).
func ChangeLightOffset(lid int, offset PositionDisplacement) {
	HookChangeLightOffset(lid, int8(offset.DeltaX), int8(offset.DeltaY))
}

// MoveMissilePos ports MoveMissilePos (Source/missiles.cpp:233-251):
// "Dodgy hack used to correct the position for charging monsters" —
// southward-facing sprites shift one tile south; the move respects tile
// availability via the injected seam.
func MoveMissilePos(m *Missile) {
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
	if HookIsTileAvailable(m.Source, target.X, target.Y) {
		m.Position.Tile = target
		// Displacement(moveDirection).worldToScreen() — upstream
		// builds the direction delta as a Displacement first
		// (displacement.hpp:247-275), so the offset delta is
		// worldToScreen of the tile delta.
		m.Position.Offset.DeltaX += WorldToScreen(PositionDisplacement{
			DeltaX: int32(dx), DeltaY: int32(dy),
		}).DeltaX
		m.Position.Offset.DeltaY += WorldToScreen(PositionDisplacement{
			DeltaX: int32(dx), DeltaY: int32(dy),
		}).DeltaY
	}
}

// directionDeltaOf mirrors DisplacementOf::fromDirection
// (Source/engine/displacement.hpp:247-275): the tile delta per Direction
// enum order (South, SouthWest, West, NorthWest, North, NorthEast, East,
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
// returns (missiles.cpp:565 vs 627) for callers that need it; the port
// keeps upstream's plain bool return.
//
// MoveMissile ports MoveMissile (Source/missiles.cpp:552-628): advance
// traveled by velocity, then walk every tile the missile crossed this
// tick — interpolating through skipped tiles with the 32×16 px increment
// — probing checkTile on each. A failed probe ends the move: with
// ifCheckTileFailsDontMoveToTile the missile never enters the tile (the
// velocity step or increment is reverted and StopMissile freezes it);
// without it the missile lands in the tile and the caller decides.
//
// [lint] revive/nestif: 1:1 upstream port — one flat interpolation loop
// over the per-tile probes (missiles.cpp:552-628); splitting it would
// sever the prevTile/traveled state the loop mutates.
func MoveMissile(m *Missile, checkTile func(TilePosition) bool, //nolint:revive // 1:1 port (missiles.cpp:552-628)
	ifCheckTileFailsDontMoveToTile bool) bool {
	prevTile := m.Position.Tile
	m.Position.Traveled.DeltaX += m.Position.Velocity.DeltaX
	m.Position.Traveled.DeltaY += m.Position.Velocity.DeltaY
	UpdateMissilePos(m)

	var possibleVisitTiles int
	if m.Position.Velocity.DeltaX == 0 || m.Position.Velocity.DeltaY == 0 {
		possibleVisitTiles = PathDistance(prevTile, m.Position.Tile)
	} else {
		possibleVisitTiles = ManhattanDistance(prevTile, m.Position.Tile)
	}

	if possibleVisitTiles == 0 {
		return false
	}

	// Did the missile skip a tile? (missiles.cpp:566-623)
	// [lint] nestif: 1:1 upstream interpolation loop (missiles.cpp:566-623).
	if possibleVisitTiles > 1 { //nolint:nestif // 1:1 port (missiles.cpp:566-623)
		speed := PositionDisplacement{
			DeltaX: absI32(m.Position.Velocity.DeltaX),
			DeltaY: absI32(m.Position.Velocity.DeltaY),
		}
		// float division + Displacement * float: the increment is
		// int32-truncated per component (displacement.hpp:95-100,
		// 311-316; missiles.cpp:571-572).
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
		// multiple of incVelocity (missiles.cpp:576-580): C++ int
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
			// with normal logic (missiles.cpp:588-592)
			if (initialDiff.DeltaX < 0) != (incDiff.DeltaX < 0) {
				break
			}
			if (initialDiff.DeltaY < 0) != (incDiff.DeltaY < 0) {
				break
			}

			// calculate in-between tile (missiles.cpp:595-598)
			pixelsTraveled := PositionDisplacement{
				DeltaX: traveled.DeltaX >> fracBits,
				DeltaY: traveled.DeltaY >> fracBits,
			}
			tileOffset := ScreenToMissile(pixelsTraveled)
			tile := TilePosition{
				X: m.Position.Start.X + int(tileOffset.DeltaX),
				Y: m.Position.Start.Y + int(tileOffset.DeltaY),
			}

			// we haven't quite reached the missile's current position,
			// but we can break early to avoid checking collisions in
			// this tile twice (missiles.cpp:601-604)
			if tile == m.Position.Tile {
				break
			}

			// skip collision logic if the missile is on a corner
			// between tiles (missiles.cpp:607-611)
			if pixelsTraveled.DeltaY%16 == 0 &&
				pixelsTraveled.DeltaX%32 == 0 &&
				absI32(pixelsTraveled.DeltaY/16)%2 != absI32(pixelsTraveled.DeltaX/32)%2 {
				continue
			}

			// don't call checkTile more than once for a tile
			// (missiles.cpp:614-616)
			if prevTile == tile {
				continue
			}

			prevTile = tile

			if !checkTile(tile) {
				m.Position.Traveled = traveled
				if ifCheckTileFailsDontMoveToTile {
					m.Position.Traveled.DeltaX -= incVelocity.DeltaX
					m.Position.Traveled.DeltaY -= incVelocity.DeltaY
					UpdateMissilePos(m)
					StopMissile(m)
				} else {
					UpdateMissilePos(m)
				}
				return true
			}
		}
	}

	if !checkTile(m.Position.Tile) && ifCheckTileFailsDontMoveToTile {
		m.Position.Traveled.DeltaX -= m.Position.Velocity.DeltaX
		m.Position.Traveled.DeltaY -= m.Position.Velocity.DeltaY
		UpdateMissilePos(m)
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
// (Source/tables/misdat.h:194-197): the low 3 bits of the row's flags
// are the DamageType ordinal — the base rows resolve to Physical (Arrow),
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
// (Source/missiles.cpp:636-663): move the missile, then (only when the
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

	tileChanged := MoveMissile(m, checkTile, ifCollidesDontMoveToHitTile)

	tileTargetHash := int16(w.DMonster[m.Position.Tile.Y][m.Position.Tile.X] ^
		w.DPlayer[m.Position.Tile.Y][m.Position.Tile.X])

	// missile didn't change the tile... check that we perform
	// CheckMissileCol only once for any monster/player to avoid
	// multiple hits for slow missiles (missiles.cpp:658-661)
	if !tileChanged && m.LastCollisionTargetHash != tileTargetHash {
		CheckMissileCol(w, m, m.Position.Tile, damageType, minDamage, maxDamage, false, false, nil)
	}

	// remember what target CheckMissileCol was checked against
	// (missiles.cpp:663)
	m.LastCollisionTargetHash = tileTargetHash
}

// ProcessArrow ports ProcessArrow (Source/missiles.cpp:2932-2958): one
// tick of arrow flight — countdown, distance tally, source-scaled damage
// roll (via the seam upstream reads Players/Monsters/currlevel for), the
// move, deletion at duration 0 and the dFlags registration.
func ProcessArrow(w *HitWorld, m *Missile) {
	m.Duration--
	m.Dist++

	// Source damage dispatch (missiles.cpp:2936-2953): player
	// _pIMinDam/_pIMaxDam, monster minDamage/maxDamage, trap
	// currlevel/2*currlevel — resolved by the injected seam.
	mind, maxd := HookSourceDamage(m)

	MoveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), mind, maxd, true, false)
	if m.Duration == 0 {
		m.DelFlag = true
	}
	PutMissile(w.Grid, m)
}

// ProcessGenericProjectile ports ProcessGenericProjectile
// (Source/missiles.cpp:2963-3016): one tick of bolt flight. On expiry
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
			HookSpawnMissile(m.Position.Tile, dst, dir, MagmaBallExplosion,
				m.Caster, m.Source, m)
		case BloodStar:
			HookSpawnMissile(m.Position.Tile, dst, dir, BloodStarExplosion,
				m.Caster, m.Source, m)
		case Acid:
			HookSpawnMissile(m.Position.Tile, dst, dir, AcidSplat,
				m.Caster, m.Source, m)
		case OrangeFlare:
			HookSpawnMissile(m.Position.Tile, dst, dir, OrangeExplosion,
				m.Caster, m.Source, m)
		case BlueFlare:
			HookSpawnMissile(m.Position.Tile, dst, dir, BlueExplosion,
				m.Caster, m.Source, m)
		case RedFlare:
			HookSpawnMissile(m.Position.Tile, dst, dir, RedExplosion,
				m.Caster, m.Source, m)
		case YellowFlare:
			HookSpawnMissile(m.Position.Tile, dst, dir, YellowExplosion,
				m.Caster, m.Source, m)
		case BlueFlare2:
			HookSpawnMissile(m.Position.Tile, dst, dir, BlueExplosion2,
				m.Caster, m.Source, m)
		default:
			// base rows outside the switch never spawn an explosion
			// upstream (missiles.cpp:3001 break).
		}
		if m.Lid != NoLight {
			HookAddUnLight(m.Lid)
		}
		PutMissile(w.Grid, m)
	} else {
		if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
			m.Var1 = m.Position.Tile.X
			m.Var2 = m.Position.Tile.Y
			if m.Lid != NoLight {
				HookChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
			}
		}
		PutMissile(w.Grid, m)
	}
}

// HookAddLight ports AddLight (Source/lighting.h:60, params x, y,
// radius): registers a new light and returns its id — the explosion
// light ramp seeds _mlid from it (missiles.cpp:3622).
var HookAddLight = func(_ int, _ int, _ uint8) int { return NoLight }

// HookGenerateRnd ports GenerateRnd (Source/random.cpp:97-102, param n):
// [0, n) draw — the trap damage rolls and the ChargedBolt steering draw.
var HookGenerateRnd = func(_ int) int { return 0 }

// Currlevel ports the dungeon-level global (Source/gendung.h): the trap
// damage base (missiles.cpp:2861-2862, 2888-2889).
var Currlevel = 0

// HookTransAt ports the tile-transparency lookup (upstream reads
// dTransVal/TransList, gendung.h): true when the tile is transparent.
// Fireball wall-adjust only (missiles.cpp:3115-3131).
var HookTransAt = func(_ int, _ int) bool { return false }

// HookCheckBlock ports CheckBlock (missiles.cpp pathing helper): true
// when the straight line from→to is blocked — the fireball ring only
// probes tiles it can see (missiles.cpp:3105-3111).
var HookCheckBlock = func(_ int, _ int, _ int, _ int) bool { return true }

// leftOf ports Left (Source/engine/direction.hpp:24-28): a left turn on
// the 8-way compass.
func leftOf(facing Direction) Direction {
	return Direction((uint8(facing) + 7) % 8)
}

// rightOf ports Right (direction.hpp:31-35): a right turn.
func rightOf(facing Direction) Direction {
	return Direction((uint8(facing) + 1) % 8)
}

// graphicSetDefaultFrameGroup ports Missile::setDefaultFrameGroup
// (Source/missiles.h:193-195): the frame group drops to 0. The
// setAnimation sheet swap is the AnimType store; animation-count state
// is animation-bean scope (TODO devilGo-mwlm animation slice).
func (m *Missile) setDefaultFrameGroup() {
	m.MFNum = 0
}

// setAnimation ports Missile::setAnimation — the sheet swap by graphic
// id (missiles.h:101 _miAnimType store).
func (m *Missile) setAnimation(g GraphicID) {
	m.AnimType = int(g)
}

// ProcessElementalArrow ports ProcessElementalArrow
// (Source/missiles.cpp:2848-2930): the elemental bows' projectile —
// flies as a physical arrow; at duration 0 it swaps to its elemental
// explosion animation, re-rolls elemental damage through the seam, and
// rides the burn-out frames with the light following the frame.
func ProcessElementalArrow(w *HitWorld, m *Missile) {
	m.Duration--

	animType := GraphicID(m.AnimType)
	if animType != GraphicChargedBolt && animType != GraphicMagmaBallExplosion {
		processElementalArrowFlight(w, m)
	} else {
		// Animation phase (missiles.cpp:2851-2853): the light rides
		// the burn-out frames.
		HookChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(m.AnimFrame+5))
	}

	// Shared tail (missiles.cpp:2921-2929): duration 0 — the explosion
	// burn-out ran out — deletes the missile and removes the light.
	if m.Duration == 0 {
		m.DelFlag = true
		HookAddUnLight(m.Lid)
	}
	PutMissile(w.Grid, m)
}

// elementalTrapDamage ports the trap dice roll of the elemental-arrow
// branches (missiles.cpp:2858-2862, 2886-2891): rnd(10)+1+currlevel and
// rnd(10)+1+currlevel*2.
func elementalTrapDamage() (minDamage, maxDamage int) {
	return HookGenerateRnd(10) + 1 + Currlevel,
		HookGenerateRnd(10) + 1 + Currlevel*2
}

// elementalPack bundles the explosion sheet, element and damage dice of one
// elemental arrow row (language-artifact wrapper; upstream returns 4 locals).
type elementalPack struct {
	graphic        GraphicID
	damageType     DamageType
	minDam, maxDam int
}

// elementalExplosionPack ports the explosion switch of
// ProcessElementalArrow (missiles.cpp:2870-2909): the animation sheet,
// damage element and damage pack per missile type; an unknown type is a
// data-table bug (app_fatal upstream).
func elementalExplosionPack(m *Missile) elementalPack {
	switch m.Type {
	case LightningArrow:
		return elementalPackFor(m, GraphicChargedBolt, DamageLightning)
	case FireArrow:
		return elementalPackFor(m, GraphicMagmaBallExplosion, DamageFire)
	default:
		panic("missiles: ProcessElementalArrow on non-elemental type")
	}
}

// elementalPackFor resolves the damage pack of one elemental row: the
// player's elemental min/max through the seam, or the trap dice
// (missiles.cpp:2871-2888).
func elementalPackFor(m *Missile, eAnim GraphicID, damageType DamageType) elementalPack {
	if m.IsTrap() {
		eMind, eMaxd := elementalTrapDamage()
		return elementalPack{graphic: eAnim, damageType: damageType, minDam: eMind, maxDam: eMaxd}
	}
	eMind, eMaxd := HookElementDamage(m, damageType)
	return elementalPack{graphic: eAnim, damageType: damageType, minDam: eMind, maxDam: eMaxd}
}

// processElementalArrowFlight ports the flight phase body of
// ProcessElementalArrow (missiles.cpp:2855-2918): the physical roll, the
// move, the duration-0 explosion transition and the radius-5 tile
// tracker.
func processElementalArrowFlight(w *HitWorld, m *Missile) {
	m.Dist++
	var mind, maxd int
	if m.IsTrap() {
		mind, maxd = elementalTrapDamage()
	} else {
		mind, maxd = HookSourceDamage(m)
	}
	MoveMissileAndCheckMissileCol(w, m, DamagePhysical, mind, maxd, true, false)

	if m.Duration == 0 {
		// Explosion transition (missiles.cpp:2866-2911).
		m.setDefaultFrameGroup()
		m.Duration = m.AnimLen - 1
		StopMissile(m)
		pack := elementalExplosionPack(m)
		m.setAnimation(pack.graphic)
		// The damage re-probe passes dontDeleteOnCollision = true
		// (missiles.cpp:2911): keepOnCollision.
		CheckMissileCol(w, m, m.Position.Tile, pack.damageType, pack.minDam, pack.maxDam,
			false, true, nil)
		return
	}
	if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
		// Flight tile tracker (missiles.cpp:2913-2918): radius 5.
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		HookChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 5)
	}
}

// fireballImpactRing probes the 8-tile ring on impact (missiles.cpp:3103-3112).
// Returns true if the missile hit something in the ring.
func fireballImpactRing(w *HitWorld, m *Missile, mp TilePosition, damageType DamageType, minDam, maxDam int) bool {
	offsets := [9]Direction{NoDirection, SouthWest, NorthEast, SouthEast,
		East, South, NorthWest, West, North}
	hit := false
	for _, off := range offsets {
		dx, dy := directionDeltaOf(off)
		px, py := mp.X+dx, mp.Y+dy
		if HookCheckBlock(m.Position.Start.X, m.Position.Start.Y, px, py) {
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

// fireballWallAdjust applies the transparency/solid nudges (missiles.cpp:3113-3131).
func fireballWallAdjust(w *HitWorld, m *Missile, mp TilePosition) {
	if !HookTransAt(mp.X, mp.Y) ||
		(m.Position.Velocity.DeltaX < 0 &&
			((HookTransAt(mp.X, mp.Y+1) && w.Grid.IsTileSolid(mp.X, mp.Y+1)) ||
				(HookTransAt(mp.X, mp.Y-1) && w.Grid.IsTileSolid(mp.X, mp.Y-1)))) {
		sdx, sdy := directionDeltaOf(South)
		m.Position.Tile = TilePosition{X: m.Position.Tile.X + sdx, Y: m.Position.Tile.Y + sdy}
		m.Position.Offset.DeltaY -= 32
	}
	if m.Position.Velocity.DeltaY > 0 &&
		((HookTransAt(mp.X+1, mp.Y) && w.Grid.IsTileSolid(mp.X+1, mp.Y)) ||
			(HookTransAt(mp.X-1, mp.Y) && w.Grid.IsTileSolid(mp.X-1, mp.Y))) {
		m.Position.Offset.DeltaY -= 32
	}
	if m.Position.Velocity.DeltaX > 0 &&
		((HookTransAt(mp.X, mp.Y+1) && w.Grid.IsTileSolid(mp.X, mp.Y+1)) ||
			(HookTransAt(mp.X, mp.Y-1) && w.Grid.IsTileSolid(mp.X, mp.Y-1))) {
		m.Position.Offset.DeltaX -= 32
	}
}

// ProcessFireball ports ProcessFireball (Source/missiles.cpp:3086-3163):
// bolt flight, then on impact an 8-tile ring re-probe behind the wall
// transparency adjustment, then the BigExplosion burn-out frames.
func ProcessFireball(w *HitWorld, m *Missile) {
	m.Duration--

	if GraphicID(m.AnimType) == GraphicBigExplosion {
		if m.Duration == 0 {
			m.DelFlag = true
			HookAddUnLight(m.Lid)
		}
		PutMissile(w.Grid, m)
		return
	}

	minDam, maxDam := m.Dam, m.Dam
	if m.Caster != int(TargetMonsters) {
		minDam, maxDam = HookSourceDamage(m)
	}
	damageType := damageTypeFor(m.Type)
	MoveMissileAndCheckMissileCol(w, m, damageType, minDam, maxDam, true, false)

	if m.Duration == 0 {
		mp := m.Position.Tile
		HookChangeLight(m.Lid, mp.X, mp.Y, uint8(m.AnimFrame))
		fireballImpactRing(w, m, mp, damageType, minDam, maxDam)

		fireballWallAdjust(w, m, mp)

		m.setDefaultFrameGroup()
		m.setAnimation(GraphicBigExplosion)
		m.Duration = m.AnimLen - 1
		m.Position.Velocity = PositionDisplacement{}
	} else if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		HookChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
	}
	PutMissile(w.Grid, m)
}

// expLightTable ports ExpLight (Source/missiles.cpp:3613): the explosion
// light radius ramp, 15 entries.
var expLightTable = [15]int{9, 10, 11, 12, 11, 10, 8, 6, 4, 2, 1, 0, 0, 0, 0}

// ProcessMissileExplosion ports ProcessMissileExplosion
// (Source/missiles.cpp:3613-3630): the static explosion animation — the
// light ramps down ExpLight while the frames play, then the light goes.
func ProcessMissileExplosion(w *HitWorld, m *Missile) {
	m.Duration--

	if m.Duration == 0 {
		m.DelFlag = true
		HookAddUnLight(m.Lid)
		return
	}
	if m.Var1 == 0 {
		// First tick: seed the explosion light (missiles.cpp:3621-3622).
		m.Lid = HookAddLight(m.Position.Tile.X, m.Position.Tile.Y, 9)
	} else {
		HookChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(expLightTable[m.Var1]))
	}
	m.Var1++
	PutMissile(w.Grid, m)
}

// bPathTable ports BPath (Source/missiles.cpp:3982): the ChargedBolt
// steering draw table, 16 entries of -1/0/1 (left/straight/right).
var bPathTable = [16]int{-1, 0, 1, -1, 0, 1, -1, -1, 0, 0, 1, 1, 0, 1, -1, 0}

// ProcessChargedBolt ports ProcessChargedBolt
// (Source/missiles.cpp:3979-4018): the weaving bolt — every 16 ticks it
// re-aims along BPath, and a hit freezes it into the Lightning
// burn-out.
func ProcessChargedBolt(w *HitWorld, m *Missile) {
	m.Duration--

	if GraphicID(m.AnimType) != GraphicLightning {
		// Flight phase (missiles.cpp:3983-3998).
		if m.Var3 == 0 {
			// Steering tick: walk BPath (missiles.cpp:3984-3995).
			md := Direction(m.Var2)
			switch bPathTable[m.Mirnd] {
			case -1:
				md = leftOf(md)
			case 1:
				md = rightOf(md)
			default:
				// straight: BPath 0 leaves the heading untouched
				// (missiles.cpp:3991 no-op branch)
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
	}
	MoveMissileAndCheckMissileCol(w, m, damageTypeFor(m.Type), m.Dam, m.Dam, false, false)

	if m.HitFlag {
		// Hit burn-out (missiles.cpp:4003-4009).
		m.Var1 = 8
		m.setDefaultFrameGroup()
		m.Position.Offset = PositionDisplacement{}
		m.Position.Velocity = PositionDisplacement{}
		m.setAnimation(GraphicLightning)
		m.Duration = m.AnimLen
	}
	HookChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, uint8(m.Var1))

	if m.Duration == 0 {
		m.DelFlag = true
		HookAddUnLight(m.Lid)
	}
	PutMissile(w.Grid, m)
}

// ProcessHolyBolt ports ProcessHolyBolt (Source/missiles.cpp:4020-4057):
// bolt flight, then the HolyBoltExplosion burn-out whose light rides the
// animation frames.
func ProcessHolyBolt(w *HitWorld, m *Missile) {
	m.Duration--

	if GraphicID(m.AnimType) != GraphicHolyBoltExplosion {
		processHolyBoltFlight(w, m)
	} else {
		// Animation phase (missiles.cpp:4037-4045): light rides the
		// burn-out frames.
		HookChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
			uint8(m.AnimFrame+7))
		if m.Duration == 0 {
			m.DelFlag = true
			HookAddUnLight(m.Lid)
		}
	}
	PutMissile(w.Grid, m)
}

// processHolyBoltFlight ports the flight phase body of ProcessHolyBolt
// (missiles.cpp:4023-4035): the move, the duration-0 transition into the
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
		HookChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 8)
	}
}
