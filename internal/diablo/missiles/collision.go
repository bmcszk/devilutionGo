// Package missiles: CheckMissileCol — the
// per-tile missile collision probe, plus its tile-wall helper
// IsMissileBlockedByTile and the small
// ground the probe needs that is not engine-global state
// (devilGo-mwlm slice 2).
// Engine boundary: upstream reads the global entity tables
// (Monsters/Players/dMonster/dPlayer) and calls the hit pipelines
// (MonsterTrapHit/MonsterMHit/Plr2PlrMHit/PlayerMHit). Those land with
// the entity beans, so everything entity-resident is injected as
// callbacks; the occupancy tables themselves (dMonster/dPlayer cells)
// are plain arrays on hitWorld so tests can script them. The wall
// branch is NOT entity-resident — it reads the world grid, which the
// contract routes through internal/game/world.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// damageType mirrors upstream enum class damageType : uint8_t.
type damageType uint8

const (
	DamagePhysical damageType = iota
	DamageFire
	DamageLightning
	DamageMagic
	DamageAcid
)

// Caster mirrors upstream enum mienemy_type : uint8_t — the missile's
// targeting side (_micaster).
type Caster uint8

const (
	TargetMonsters Caster = iota
	TargetPlayers
	TargetBoth
)

// deathReason mirrors enum class deathReason.
type deathReason uint8

const (
	// DeathMonsterOrTrap: monster or trap (dungeon).
	DeathMonsterOrTrap deathReason = iota
	// DeathByPlayer: other player or selfkill (for example firewall).
	DeathByPlayer
	// DeathUnknown: HP is zero but we don't know when or where.
	DeathUnknown
)

// missileSource mirrors upstream enum class missileSource — what fired
// the missile (sourceType()).
type missileSource uint8

const (
	SourceTrap missileSource = iota
	SourceMonster
	SourcePlayer
)

// Direction mirrors upstream enum class Direction : uint8_t, ordinal
// order preserved.
type Direction uint8

// Direction values (ordinal order upstream).
const (
	South Direction = iota
	SouthWest
	West
	NorthWest
	North
	NorthEast
	East
	SouthEast
	NoDirection
)

// HitCallbacks is the engine boundary CheckMissileCol reads instead of
// the upstream globals (Monsters, Players, PlaySfxLoc, gbIsHellfire
// stays as a plain flag). Entity-hit branches take the injected
// predicates; wire the real MonsterTrapHit/MonsterMHit/Plr2PlrMHit/
// PlayerMHit ports when the entity beans land.
type HitCallbacks struct {
	// MonsterTargeted ports the hit gate on the probed dMonster cell
	// without onlyHitWalking — `mid > 0 ||
	// monster.mode == MonsterMode::Petrified` (a walking monster's
	// negative cell never registers); with it —
	// `monster.isWalking() && CheckCanHitOnlyWalking(...)`
	// (CheckCanHitOnlyWalking,, needs the
	// monster's ActorPosition, so it stays behind the hook).
	MonsterTargeted func(mid int16, onlyHitWalking *Direction) bool
	// MonsterFactionConflict ports the MonsterTrapHit-branch guard
	// the monsters are on opposing factions
	// (`monster.isPlayerMinion() != Monsters[missile._misource]
	// .isPlayerMinion()`) or either side is berserked (`MFLAG_BERSERK`
	// . Entity state — injected.
	MonsterFactionConflict func(targetMonster int, sourceMonster int) bool
	// MonsterTrapHit ports MonsterTrapHit
	// the trap/monster-caster damage roll. trailing keepOnCollision is
	// a port-side probe seam: upstream keeps that flag in
	// CheckMissileCol scope; the port forwards it so tests can pin the
	// re-probe variants/3111 keepOnCollision).
	MonsterTrapHit func(minDamage, maxDamage, dist int,
		missileType missileID, damageType damageType, shifted, keepOnCollision bool) bool
	// MonsterMHit ports MonsterMHit: the
	// player-caster damage roll (same trailing probe seam).
	MonsterMHit func(minDamage, maxDamage, dist int,
		missileType missileID, damageType damageType, shifted, keepOnCollision bool) bool
	// Plr2PlrMHit ports Plr2PlrMHit: player-vs-player
	// roll; upstream's `&blocked` out param becomes the second return.
	Plr2PlrMHit func(minDamage, maxDamage, dist int,
		missileType missileID, damageType damageType, shifted bool) (hit, blocked bool)
	// PlayerMHit ports PlayerMHit: monsterSide is
	// nil for trap-or-self kills — the `nullptr`
	// monster argument); `&blocked` becomes the second return.
	PlayerMHit func(monsterSide *int, minDamage, maxDamage, dist int,
		missileType missileID, damageType damageType, shifted bool,
		deathReason deathReason) (hit, blocked bool)
	// BreakObjectMissile ports BreakObjectMissile
	// the breakable-object side effect.
	BreakObjectMissile func()
	// PlayHitSound ports the missileData hit-sound tail
	// upstream fires it when duration hit 0 and
	// the row's hitSound != SfxID::None; audio is engine-owned, so the
	// caller decides via this hook.
	PlayHitSound func()
}

// MovementDistribution mirrors upstream enum class
// MissileMovementDistribution.
type MovementDistribution uint8

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

// hitWorld is the engine state CheckMissileCol reads per call: the world
// grid (dPiece/dObject/Objects/SOLData — upstream gendung globals) plus
// the entity occupancy cells (dMonster, dPlayer) and the gbIsHellfire
// flag. 0 = empty cell, else the 1-based entity index; dMonster/dPlayer
// cells are signed (negative = walking/auxiliary spawn).
// Cell layout follows the repo grid convention dPiece[dunY][dunX].
type hitWorld struct {
	Grid       *world.Grid
	DMonster   [world.MAXDUNY][world.MAXDUNX]int16
	DPlayer    [world.MAXDUNY][world.MAXDUNX]int16
	IsHellfire bool
	// Rnd mirrors GenerateRnd for the hellfire
	// blocked-shot draw PickRandomlyAmong({-1, 1})
	// — engine state, injected like the callbacks;
	// must be set before CheckMissileCol (tests script the tape).
	Rnd          func(int) int
	HitCallbacks HitCallbacks
	// processCallbacks is the engine-boundary sink of the motion
	// pipeline and the Process fns (process.go) — the HitCallbacks
	// pattern applied to the process-side globals (lighting engine
	// explosion spawns, damage reads, tile predicates, RNG).
	processCallbacks processCallbacks
}

// PlayerAtPosition ports PlayerAtPosition
// the 1-based dPlayer cell resolved to a 0-based index; ignoreMovingPlayers
// skips negative (walking) cells.
// [lint] revive flag-parameter: ignoreMovingPlayers IS an upstream
// control flag `playerIndex < 0` skip) — kept 1:1
// not split.
func (w *hitWorld) PlayerAtPosition(x, y int, ignoreMovingPlayers bool) (index int, ok bool) { //nolint:revive // 1:1 upstream flag (player.cpp:2044)
	if !world.InDungeonBounds(x, y) {
		return 0, false
	}
	playerIndex := w.DPlayer[y][x]
	if playerIndex == 0 || (ignoreMovingPlayers && playerIndex < 0) {
		return 0, false
	}
	return int(absI16(playerIndex)) - 1, true
}

// SourceOf ports Missile::sourceType — trap when Source == -1.
func (m *Missile) SourceOf() missileSource {
	if m.Source == -1 {
		return SourceTrap
	}
	if m.caster == int(TargetPlayers) {
		return SourceMonster
	}
	return SourcePlayer
}

// IsTrap ports Missile::IsTrap (_misource == -1).
func (m *Missile) IsTrap() bool { return m.Source == -1 }

// checkMissileColResult is the outcome callers need from the probe. The
// C++ version reports through missile fields (duration/_miHitFlag); the
// Go port mutates the same fields and additionally returns the per-call
// booleans so tests assert observable outcomes.
type checkMissileColResult struct {
	MonsterHit bool
	PlayerHit  bool
	// Blocked: hellfire shield block — triggers the missile rotation
	// instead of deletion.
	Blocked bool
	// WallBlocked: the IsMissileBlockedByTile branch
	// (missiles.cpp:536-544) — sets HitFlag = false and breaks objects.
	WallBlocked bool
}

// CheckMissileCol ports CheckMissileCol
// line by line. The upstream parameter pack (missile, damageType
// minDamage, maxDamage, isDamageShifted, position, dontDeleteOnCollision
// optional<Direction> onlyHitWalking) is preserved; the bare bools are
// split into named pairs so no flag-param ambiguity survives
// shifted ↔ isDamageShifted, keepOnCollision ↔ !dontDeleteOnCollision
// and onlyHitWalking == nil ↔ the empty optional.
// [lint] revive cognitive-complexity (25): 1:1 upstream port —
// upstream's CheckMissileCol is one flat function over five branches
// ; the body reads branch-for-branch against the
// C++. Only the roll dispatches are extracted below (rollMonsterHit /
// rollPlayerHit,, preserving statement order and
// draw order.
func CheckMissileCol(w *hitWorld, missile *Missile, position tilePosition, damageType damageType, minDamage, maxDamage int, shifted, keepOnCollision bool, onlyHitWalking *Direction) checkMissileColResult { //nolint:revive // 1:1 port (missiles.cpp:479-538)
	var res checkMissileColResult

	if !world.InDungeonBounds(position.X, position.Y) {
		return res
	}

	// Monster branch.
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

	// Player branch. Upstream passes
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
			// right instead of deleting it. The
			// rotation draw comes from w.Rnd: [mwlm2 fix] pass the
			// injected seam through — upstream's
			// rotateBlockedMissile(missile) draws GenerateRnd(2)
			// internally.
			rotateBlockedMissile(missile, w.Rnd)
		} else if !keepOnCollision {
			missile.Duration = 0
		}
		missile.HitFlag = true
	}

	// Wall / breakable-object branch.
	if IsMissileBlockedByTile(w.Grid, position.X, position.Y) {
		res.WallBlocked = true
		if obj := w.Grid.FindObjectAtPosition(position.X, position.Y); obj != nil && obj.Breakable {
			w.HitCallbacks.BreakObjectMissile()
		}

		if !keepOnCollision {
			missile.Duration = 0
		}
		missile.HitFlag = false
	}

	// Hit-sound tail: upstream checks
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
// trap, or monster-caster on an opposing
// faction / berserk — the MonsterTrapHit roll; else TARGET_BOTH /
// TARGET_MONSTERS — the MonsterMHit roll. Extracted verbatim
// (statement and draw order preserved).
func (w *hitWorld) rollMonsterHit(monsterIndex int, missile *Missile,
	damageType damageType, minDamage, maxDamage int, shifted, keepOnCollision bool) bool {
	if missile.IsTrap() ||
		(missile.caster == int(TargetPlayers) &&
			w.HitCallbacks.MonsterFactionConflict(monsterIndex, missile.Source)) {
		// Trap, or monster-caster on an opposing faction /
		// berserk: the trap damage roll.
		return w.HitCallbacks.MonsterTrapHit(minDamage, maxDamage,
			missile.Dist, missile.Type, damageType, shifted, keepOnCollision)
	}
	if missile.caster == int(TargetBoth) || missile.caster == int(TargetMonsters) {
		return w.HitCallbacks.MonsterMHit(minDamage, maxDamage,
			missile.Dist, missile.Type, damageType, shifted, keepOnCollision)
	}
	return false
}

// rollPlayerHit is the player-roll dispatch of CheckMissileCol
// TARGET_MONSTERS hits other players
// (Plr2PlrMHit) unless self-fired; a monster side rolls PlayerMHit
// with the source monster; TARGET_BOTH/trap rolls PlayerMHit with the
// death reason following the source. Extracted verbatim (statement
// and draw order preserved).
func (w *hitWorld) rollPlayerHit(playerIdx int, missile *Missile,
	damageType damageType, minDamage, maxDamage int, shifted bool) (hit, blocked bool) {
	if missile.caster != int(TargetBoth) && !missile.IsTrap() {
		if missile.caster == int(TargetMonsters) {
			if playerIdx != missile.Source {
				// Plr2PlrMHit.
				return w.HitCallbacks.Plr2PlrMHit(minDamage, maxDamage,
					missile.Dist, missile.Type, damageType, shifted)
			}
			return false, false
		}
		// Monsters[missile._misource] — index only until the
		// monster table lands.
		ms := missile.Source
		return w.HitCallbacks.PlayerMHit(&ms, minDamage, maxDamage,
			missile.Dist, missile.Type, damageType, shifted, DeathMonsterOrTrap)
	}
	// TARGET_BOTH or trap: the death reason follows the source
	// .
	deathReason := DeathMonsterOrTrap
	if missile.SourceOf() == SourcePlayer {
		deathReason = DeathByPlayer
	}
	return w.HitCallbacks.PlayerMHit(nil, minDamage, maxDamage,
		missile.Dist, missile.Type, damageType, shifted, deathReason)
}

// IsMissileBlockedByTile ports IsMissileBlockedByTile
// out of bounds blocks; the tile's SOL
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
	// so we need to invert the check here.
	return obj != nil && !obj.MissFlag
}

func absI16(v int16) int16 {
	if v < 0 {
		return -v
	}
	return v
}
