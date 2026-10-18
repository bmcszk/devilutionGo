// Package missiles: CheckMissileCol (Source/missiles.cpp:479-538) — the
// per-tile missile collision probe, plus its tile-wall helper
// IsMissileBlockedByTile (Source/missiles.cpp:857-870) and the small
// ground the probe needs that is not engine-global state
// (devilGo-mwlm slice 2).
//
// Engine boundary: upstream reads the global entity tables
// (Monsters/Players/dMonster/dPlayer) and calls the hit pipelines
// (MonsterTrapHit/MonsterMHit/Plr2PlrMHit/PlayerMHit). Those land with
// the entity beans, so everything entity-resident is injected as
// callbacks; the occupancy tables themselves (dMonster/dPlayer cells)
// are plain arrays on HitWorld so tests can script them. The wall
// branch is NOT entity-resident — it reads the world grid, which the
// contract routes through internal/world.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/world"
)

// DamageType mirrors enum class DamageType : uint8_t
// (Source/tables/misdat.h:27-33).
type DamageType uint8

// DamageType values (Source/tables/misdat.h:28-32).
const (
	DamagePhysical DamageType = iota
	DamageFire
	DamageLightning
	DamageMagic
	DamageAcid
)

// Caster mirrors enum mienemy_type : uint8_t
// (Source/tables/misdat.h:21-25) — the missile's targeting side
// (_micaster, Source/missiles.h:121).
type Caster uint8

// Caster values (Source/tables/misdat.h:22-24).
const (
	TargetMonsters Caster = iota
	TargetPlayers
	TargetBoth
)

// DeathReason mirrors enum class DeathReason (Source/player.h:159-166).
type DeathReason uint8

// DeathReason values (Source/player.h:160-165).
const (
	// DeathMonsterOrTrap: monster or trap (dungeon).
	DeathMonsterOrTrap DeathReason = iota
	// DeathByPlayer: other player or selfkill (for example firewall).
	DeathByPlayer
	// DeathUnknown: HP is zero but we don't know when or where.
	DeathUnknown
)

// MissileSource mirrors enum class MissileSource (Source/missiles.h:133):
// what fired the missile (sourceType(), Source/missiles.h:164-171).
type MissileSource uint8

// MissileSource values (Source/missiles.h:133).
const (
	SourceTrap MissileSource = iota
	SourceMonster
	SourcePlayer
)

// Direction mirrors enum class Direction : uint8_t
// (Source/engine/direction.hpp:11-20), ordinal order upstream.
type Direction uint8

// Direction values (Source/engine/direction.hpp:12-19).
const (
	South Direction = iota
	SouthWest
	West
	NorthWest
	North
	NorthEast
	East
	SouthEast
	// NoDirection is the sentinel (direction.hpp:19).
	NoDirection
)

// HitCallbacks is the engine boundary CheckMissileCol reads instead of
// the upstream globals (Monsters, Players, PlaySfxLoc, gbIsHellfire
// stays as a plain flag). Entity-hit branches take the injected
// predicates; wire the real MonsterTrapHit/MonsterMHit/Plr2PlrMHit/
// PlayerMHit ports when the entity beans land.
type HitCallbacks struct {
	// MonsterTargeted ports the hit gate on the probed dMonster cell
	// (missiles.cpp:489): without onlyHitWalking — `mid > 0 ||
	// monster.mode == MonsterMode::Petrified` (a walking monster's
	// negative cell never registers); with it —
	// `monster.isWalking() && CheckCanHitOnlyWalking(...)`
	// (CheckCanHitOnlyWalking, Source/missiles.cpp:468-476, needs the
	// monster's ActorPosition, so it stays behind the hook).
	MonsterTargeted func(mid int16, onlyHitWalking *Direction) bool
	// MonsterFactionConflict ports the MonsterTrapHit-branch guard
	// (missiles.cpp:491-497): the monsters are on opposing factions
	// (`monster.isPlayerMinion() != Monsters[missile._misource]
	// .isPlayerMinion()`) or either side is berserked (`MFLAG_BERSERK`,
	// Source/monster.h:53). Entity state — injected.
	MonsterFactionConflict func(targetMonster int, sourceMonster int) bool
	// MonsterTrapHit ports MonsterTrapHit (Source/missiles.cpp:1024):
	// the trap/monster-caster damage roll. trailing keepOnCollision is
	// a port-side probe seam: upstream keeps that flag in
	// CheckMissileCol scope; the port forwards it so tests can pin the
	// re-probe variants (missiles.cpp:2911/3111 keepOnCollision).
	MonsterTrapHit func(minDamage, maxDamage, dist int,
	missileType MissileID, damageType DamageType, shifted, keepOnCollision bool) bool
	// MonsterMHit ports MonsterMHit (Source/monster.cpp): the
	// player-caster damage roll (same trailing probe seam).
	MonsterMHit func(minDamage, maxDamage, dist int,
	missileType MissileID, damageType DamageType, shifted, keepOnCollision bool) bool
	// Plr2PlrMHit ports Plr2PlrMHit (Source/player.cpp): player-vs-player
	// roll; upstream's `&blocked` out param becomes the second return.
	Plr2PlrMHit func(minDamage, maxDamage, dist int,
		missileType MissileID, damageType DamageType, shifted bool) (hit, blocked bool)
	// PlayerMHit ports PlayerMHit (Source/player.cpp): monsterSide is
	// nil for trap-or-self kills (missiles.cpp:522 — the `nullptr`
	// monster argument); `&blocked` becomes the second return.
	PlayerMHit func(monsterSide *int, minDamage, maxDamage, dist int,
		missileType MissileID, damageType DamageType, shifted bool,
		deathReason DeathReason) (hit, blocked bool)
	// BreakObjectMissile ports BreakObjectMissile (Source/objects.cpp):
	// the breakable-object side effect (missiles.cpp:536-538).
	BreakObjectMissile func()
	// PlayHitSound ports the missileData hit-sound tail
	// (missiles.cpp:545-548): upstream fires it when duration hit 0 and
	// the row's hitSound != SfxID::None; audio is engine-owned, so the
	// caller decides via this hook.
	PlayHitSound func()
}

// MovementDistribution mirrors enum class MissileMovementDistribution
// (Source/tables/misdat.h:101-117).
type MovementDistribution uint8

// MovementDistribution values (Source/tables/misdat.h:105-116).
const (
	// MovementDisabled: no movement distribution is calculated —
	// normally the missile doesn't move at all.
	MovementDisabled MovementDistribution = iota
	// MovementBlockable: the missile moves and stops on a hit
	// (for example firebolt).
	MovementBlockable
	// MovementUnblockable: the missile keeps moving through hits
	// (for example flame wave).
	MovementUnblockable
)

// HitWorld is the engine state CheckMissileCol reads per call: the world
// grid (dPiece/dObject/Objects/SOLData — upstream gendung globals) plus
// the entity occupancy cells (dMonster, Source/levels/dun_tile_data.hpp:118;
// dPlayer, dun_tile_data.hpp:117) and the gbIsHellfire flag. 0 = empty
// cell, else the 1-based entity index; dMonster/dPlayer cells are signed
// (negative = walking/auxiliary spawn — PlayerAtPosition player.cpp:2044).
// Cell layout follows the repo grid convention dPiece[dunY][dunX].
type HitWorld struct {
	Grid     *world.Grid
	DMonster [world.MAXDUNY][world.MAXDUNX]int16
	DPlayer  [world.MAXDUNY][world.MAXDUNX]int16
	//	IsHellfire   bool
	//	HitCallbacks HitCallbacks
	IsHellfire bool
	// Rnd mirrors GenerateRnd (random.cpp:97-102) for the hellfire
	// blocked-shot draw PickRandomlyAmong({-1, 1}) (random.hpp:377-384,
	// missiles.cpp:450) — engine state, injected like the callbacks;
	// must be set before CheckMissileCol (tests script the tape).
	Rnd          func(int) int
	HitCallbacks HitCallbacks
}

// PlayerAtPosition ports PlayerAtPosition (Source/player.cpp:2040-2049):
// the 1-based dPlayer cell resolved to a 0-based index; ignoreMovingPlayers
// skips negative (walking) cells.
//
// [lint] revive flag-parameter: ignoreMovingPlayers IS an upstream
// control flag (player.cpp:2044 `playerIndex < 0` skip) — kept 1:1,
// not split.
func (w *HitWorld) PlayerAtPosition(x, y int, ignoreMovingPlayers bool) (index int, ok bool) { //nolint:revive // 1:1 upstream flag (player.cpp:2044)
	if !world.InDungeonBounds(x, y) {
		return 0, false
	}
	playerIndex := w.DPlayer[y][x]
	if playerIndex == 0 || (ignoreMovingPlayers && playerIndex < 0) {
		return 0, false
	}
	return int(absI16(playerIndex)) - 1, true
}

// SourceOf ports Missile::sourceType (Source/missiles.h:164-171).
func (m *Missile) SourceOf() MissileSource {
	if m.Source == -1 {
		return SourceTrap
	}
	if m.Caster == int(TargetPlayers) {
		return SourceMonster
	}
	return SourcePlayer
}

// IsTrap ports Missile::IsTrap (Source/missiles.h:150-152):
// _misource == -1.
func (m *Missile) IsTrap() bool { return m.Source == -1 }

// CheckMissileColResult is the outcome callers need from the probe. The
// C++ version reports through missile fields (duration/_miHitFlag) and
// returns void; the Go port mutates the same fields and additionally
// returns the per-call booleans so tests assert observable outcomes.
type CheckMissileColResult struct {
	// MonsterHit: isMonsterHit (missiles.cpp:506).
	MonsterHit bool
	// PlayerHit: isPlayerHit (missiles.cpp:528).
	PlayerHit bool
	// Blocked: hellfire shield block — triggers the missile rotation
	// instead of deletion (missiles.cpp:529-530).
	Blocked bool
	// WallBlocked: IsMissileBlockedByTile branch (missiles.cpp:534).
	WallBlocked bool
}

// CheckMissileCol ports CheckMissileCol (Source/missiles.cpp:479-538)
// line by line. The upstream parameter pack (missile, damageType,
// minDamage, maxDamage, isDamageShifted, position, dontDeleteOnCollision,
// optional<Direction> onlyHitWalking) is preserved; the bare bools are
// split into named pairs so no flag-param ambiguity survives:
// shifted ↔ isDamageShifted, keepOnCollision ↔ !dontDeleteOnCollision,
// and onlyHitWalking == nil ↔ the empty optional.
//
// [lint] revive cognitive-complexity (25): 1:1 upstream port —
// upstream's CheckMissileCol is one flat function over five branches
// (missiles.cpp:479-538); the body reads branch-for-branch against the
// C++. Only the roll dispatches are extracted below (rollMonsterHit /
// rollPlayerHit, missiles.cpp:491-519), preserving statement order and
// draw order.
func CheckMissileCol(w *HitWorld, missile *Missile, position TilePosition, damageType DamageType, minDamage, maxDamage int, shifted, keepOnCollision bool, onlyHitWalking *Direction) CheckMissileColResult { //nolint:revive // 1:1 port (missiles.cpp:479-538)
	var res CheckMissileColResult

	if !world.InDungeonBounds(position.X, position.Y) {
		return res
	}

	// Monster branch (missiles.cpp:484-504).
	isMonsterHit := false
	mid := w.DMonster[position.Y][position.X]
	if mid != 0 && w.HitCallbacks.MonsterTargeted(mid, onlyHitWalking) {
		monsterIndex := int(absI16(mid)) - 1
		isMonsterHit = w.rollMonsterHit(monsterIndex, missile, damageType,
			minDamage, maxDamage, shifted, keepOnCollision)
	}

	if isMonsterHit {
		if !keepOnCollision {
			missile.Duration = 0
		}
		missile.HitFlag = true
	}

	// Player branch (missiles.cpp:508-532). Upstream passes
	// !onlyHitWalking.has_value() as ignoreMovingPlayers.
	isPlayerHit := false
	blocked := false
	playerIdx, playerPresent := w.PlayerAtPosition(position.X, position.Y, onlyHitWalking == nil)
	if playerPresent {
		isPlayerHit, blocked = w.rollPlayerHit(playerIdx, missile, damageType,
			minDamage, maxDamage, shifted)
	}

	if isPlayerHit {
		if w.IsHellfire && blocked {
			// Hellfire blocked shot: rotate the missile 45° left or
			// right instead of deleting it (missiles.cpp:529-530). The
			// rotation draw comes from w.Rnd: [mwlm2 fix] pass the
			// injected seam through — upstream's
			// RotateBlockedMissile(missile) draws GenerateRnd(2)
			// internally (missiles.cpp:450).
			RotateBlockedMissile(missile, w.Rnd)
		} else if !keepOnCollision {
			missile.Duration = 0
		}
		missile.HitFlag = true
	}

	// Wall / breakable-object branch (missiles.cpp:534-543).
	if IsMissileBlockedByTile(w.Grid, position.X, position.Y) {
		res.WallBlocked = true
		if obj := w.Grid.FindObjectAtPosition(position.X, position.Y); obj != nil && obj.Breakable {
			// BreakObjectMissile(*object) (missiles.cpp:536-538).
			w.HitCallbacks.BreakObjectMissile()
		}

		if !keepOnCollision {
			missile.Duration = 0
		}
		missile.HitFlag = false
	}

	// Hit-sound tail (missiles.cpp:545-548): upstream checks
	// missileData.hitSound != SfxID::None then PlaySfxLoc — audio is
	// engine-owned, so the hook carries the whole decision.
	if missile.Duration == 0 {
		w.HitCallbacks.PlayHitSound()
	}

	res.MonsterHit = isMonsterHit
	res.PlayerHit = isPlayerHit
	res.Blocked = blocked
	return res
}

// rollMonsterHit is the monster-roll dispatch of CheckMissileCol
// (missiles.cpp:491-497): trap, or monster-caster on an opposing
// faction / berserk — the MonsterTrapHit roll; else TARGET_BOTH /
// TARGET_MONSTERS — the MonsterMHit roll. Extracted verbatim
// (statement and draw order preserved).
func (w *HitWorld) rollMonsterHit(monsterIndex int, missile *Missile,
	damageType DamageType, minDamage, maxDamage int, shifted, keepOnCollision bool) bool {
	if missile.IsTrap() ||
		(missile.Caster == int(TargetPlayers) &&
			w.HitCallbacks.MonsterFactionConflict(monsterIndex, missile.Source)) {
		// Trap, or monster-caster on an opposing faction /
		// berserk (missiles.cpp:491-497): the trap damage roll.
		return w.HitCallbacks.MonsterTrapHit(minDamage, maxDamage,
			missile.Dist, missile.Type, damageType, shifted, keepOnCollision)
	}
	if missile.Caster == int(TargetBoth) || missile.Caster == int(TargetMonsters) {
		return w.HitCallbacks.MonsterMHit(minDamage, maxDamage,
			missile.Dist, missile.Type, damageType, shifted, keepOnCollision)
	}
	return false
}

// rollPlayerHit is the player-roll dispatch of CheckMissileCol
// (missiles.cpp:512-519): TARGET_MONSTERS hits other players
// (Plr2PlrMHit) unless self-fired; a monster side rolls PlayerMHit
// with the source monster; TARGET_BOTH/trap rolls PlayerMHit with the
// death reason following the source. Extracted verbatim (statement
// and draw order preserved).
func (w *HitWorld) rollPlayerHit(playerIdx int, missile *Missile,
	damageType DamageType, minDamage, maxDamage int, shifted bool) (hit, blocked bool) {
	if missile.Caster != int(TargetBoth) && !missile.IsTrap() {
		if missile.Caster == int(TargetMonsters) {
			if playerIdx != missile.Source {
				// Plr2PlrMHit (missiles.cpp:514-515).
				return w.HitCallbacks.Plr2PlrMHit(minDamage, maxDamage,
					missile.Dist, missile.Type, damageType, shifted)
			}
			return false, false
		}
		// Monsters[missile._misource] — index only until the
		// monster table lands (missiles.cpp:517-519).
		ms := missile.Source
		return w.HitCallbacks.PlayerMHit(&ms, minDamage, maxDamage,
			missile.Dist, missile.Type, damageType, shifted, DeathMonsterOrTrap)
	}
	// TARGET_BOTH or trap: the death reason follows the source
	// (missiles.cpp:516-518).
	deathReason := DeathMonsterOrTrap
	if missile.SourceOf() == SourcePlayer {
		deathReason = DeathByPlayer
	}
	return w.HitCallbacks.PlayerMHit(nil, minDamage, maxDamage,
		missile.Dist, missile.Type, damageType, shifted, deathReason)
}

// IsMissileBlockedByTile ports IsMissileBlockedByTile
// (Source/missiles.cpp:857-870): out of bounds blocks; the tile's SOL
// entry carrying BlockMissile blocks; a present object blocks unless it
// allows missiles through (_oMissFlag).
func IsMissileBlockedByTile(g *world.Grid, x, y int) bool {
	if !world.InDungeonBounds(x, y) {
		return true
	}

	if g.TileHasAny(x, y, world.TileBlockMissile) {
		return true
	}

	obj := g.FindObjectAtPosition(x, y)
	// _oMissFlag is true if the object allows missiles to pass through
	// so we need to invert the check here (missiles.cpp:866-869).
	return obj != nil && !obj.MissFlag
}

func absI16(v int16) int16 {
	if v < 0 {
		return -v
	}
	return v
}
