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
// are plain arrays on HitWorld so tests can script them. The wall
// branch is NOT entity-resident — it reads the world grid, which the
// contract routes through internal/game/world.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

const (
	DamagePhysical DamageType = iota
	DamageFire
	DamageLightning
	DamageMagic
	DamageAcid
)

const (
	TargetMonsters Caster = iota
	TargetPlayers
	TargetBoth
)

const (
	// DeathMonsterOrTrap: monster or trap (dungeon).
	DeathMonsterOrTrap DeathReason = iota
	// DeathByPlayer: other player or selfkill (for example firewall).
	DeathByPlayer
	// DeathUnknown: HP is zero but we don't know when or where.
	DeathUnknown
)

const (
	SourceTrap missileSource = iota
	SourceMonster
	SourcePlayer
)

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

// DamageType mirrors upstream enum class DamageType : uint8_t.
type DamageType uint8

// Caster mirrors upstream enum mienemy_type : uint8_t — the missile's
// targeting side (_micaster).
type Caster uint8

// DeathReason mirrors enum class deathReason.
type DeathReason uint8

// missileSource mirrors upstream enum class missileSource — what fired
// the missile (sourceType()).
type missileSource uint8

// Direction mirrors upstream enum class Direction : uint8_t, ordinal
// order preserved.
type Direction uint8

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
	// `monster.isWalking() && CheckCanHitOnlyWalking(...)`,
	// which needs the
	// monster's ActorPosition, so it stays behind the hook).
	MonsterTargeted func(mid int16, onlyHitWalking *Direction) bool
	// MonsterFactionConflict ports the MonsterTrapHit-branch guard
	// the monsters are on opposing factions
	// (`monster.isPlayerMinion() != Monsters[missile._misource]
	// .isPlayerMinion()`) or either side is berserked (`MFLAG_BERSERK`
	// . Entity state — injected.
	MonsterFactionConflict func(targetMonster, sourceMonster int) bool
	// MonsterTrapHit ports MonsterTrapHit
	// the trap/monster-Caster damage roll. trailing KeepOnCollision is
	// a port-side probe seam: upstream keeps that flag in
	// CheckMissileCol scope; the port forwards it so tests can pin the
	// re-probe variants/3111 KeepOnCollision).
	MonsterTrapHit func(minDamage, maxDamage, dist int,
		missileType MissileID, damageType DamageType, shifted bool, keepOnCollision KeepOnCollision) bool
	// MonsterMHit ports MonsterMHit: the
	// player-Caster damage roll (same trailing probe seam).
	MonsterMHit func(minDamage, maxDamage, dist int,
		missileType MissileID, damageType DamageType, shifted bool, keepOnCollision KeepOnCollision) bool
	// Plr2PlrMHit ports Plr2PlrMHit: player-vs-player
	// roll; upstream's `&blocked` out param becomes the second return.
	Plr2PlrMHit func(minDamage, maxDamage, dist int,
		missileType MissileID, damageType DamageType, shifted bool) (hit, blocked bool)
	// PlayerMHit ports PlayerMHit: monsterSide is
	// nil for trap-or-self kills — the `nullptr`
	// monster argument); `&blocked` becomes the second return.
	PlayerMHit func(monsterSide *int, minDamage, maxDamage, dist int,
		missileType MissileID, damageType DamageType, shifted bool,
		reason DeathReason) (hit, blocked bool)
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

// HitWorld is the engine state CheckMissileCol reads per call: the world
// grid (dPiece/dObject/Objects/SOLData — upstream gendung globals) plus
// the entity occupancy cells (dMonster, dPlayer) and the gbIsHellfire
// flag. 0 = empty cell, else the 1-based entity index; dMonster/dPlayer
// cells are signed (negative = walking/auxiliary spawn).
// Cell layout follows the repo grid convention dPiece[dunY][dunX].
type HitWorld struct {
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
	// ProcessCallbacks is the engine-boundary sink of the motion
	// pipeline and the Process fns (process.go) — the hitCallbacks
	// pattern applied to the process-side globals (lighting engine
	// explosion spawns, damage reads, tile predicates, RNG).
	ProcessCallbacks ProcessCallbacks
}

// KeepOnCollision is the upstream `keep` flag of CheckMissileCol
// (missiles.cpp:479): when true a hit does not zero the missile duration.
type KeepOnCollision bool

// ignoreMovingPlayers is the upstream skip flag of PlayerAtPosition
// (player.cpp:2044): negative (walking) player cells are not targets.
type ignoreMovingPlayers bool

// CheckMissileColResult is the outcome callers need from the probe. The
// C++ version reports through missile fields (duration/_miHitFlag); the
// Go port mutates the same fields and additionally returns the per-call
// booleans so tests assert observable outcomes.
type CheckMissileColResult struct {
	MonsterHit bool
	PlayerHit  bool
	// Blocked: hellfire shield block — triggers the missile rotation
	// instead of deletion.
	Blocked bool
	// WallBlocked: the IsMissileBlockedByTile branch
	// (missiles.cpp:536-544) — sets HitFlag = false and breaks objects.
	WallBlocked bool
}

// MissileColProbe carries the per-probe arguments of CheckMissileCol
// (missiles.cpp:479): the damage pack plus the shifted/KeepOnCollision
// flags and the onlyHitWalking pointer of the 1:1 signature.
type MissileColProbe struct {
	DamageType DamageType
	MinDamage  int
	MaxDamage  int
	Shifted    bool
	Keep       KeepOnCollision
	// OnlyHitWalking is the upstream std::optional<Direction>; nil means
	// "ignore moving players" at the player branch.
	OnlyHitWalking *Direction
}

// PlayerAtPosition ports PlayerAtPosition
// the 1-based dPlayer cell resolved to a 0-based index; ignoreMovingPlayers
// skips negative (walking) cells.
// [lint] revive flag-parameter: ignoreMovingPlayers IS an upstream
// control flag `playerIndex < 0` skip) — kept 1:1
// not split.
// PlayerAtPosition ports PlayerAtPosition (player.cpp:2044);
// ignoreMovingPlayers is the upstream bool flag of the 1:1 signature.
func (w *HitWorld) PlayerAtPosition(x, y int, skipMoving ignoreMovingPlayers) (index int, ok bool) {
	if !world.InDungeonBounds(x, y) {
		return 0, false
	}

	playerIndex := w.DPlayer[y][x]
	if playerIndex == 0 || (bool(skipMoving) && playerIndex < 0) {
		return 0, false
	}

	return int(absI16(playerIndex)) - 1, true
}

// SourceOf ports Missile::sourceType — trap when Source == -1.
func (m *Missile) SourceOf() missileSource {
	if m.Source == -1 {
		return SourceTrap
	}

	if m.Caster == int(TargetPlayers) {
		return SourceMonster
	}

	return SourcePlayer
}

// IsTrap ports Missile::IsTrap (_misource == -1).
func (m *Missile) IsTrap() bool { return m.Source == -1 }

// CheckMissileCol ports CheckMissileCol (missiles.cpp:479-538); the bools
// shifted/KeepOnCollision and the onlyHitWalking pointer are the upstream
// flags of the 1:1 signature.
func CheckMissileCol(
	w *HitWorld,
	missile *Missile,
	position TilePosition,
	probe MissileColProbe,
) CheckMissileColResult {
	var res CheckMissileColResult

	if !world.InDungeonBounds(position.X, position.Y) {
		return res
	}

	res.MonsterHit = w.checkMissileColMonsterBranch(missile, position, probe)

	// Upstream passes !onlyHitWalking.has_value() as ignoreMovingPlayers.
	res.PlayerHit, res.Blocked = w.checkMissileColPlayerBranch(missile, position, probe,
		ignoreMovingPlayers(probe.OnlyHitWalking == nil))

	res.WallBlocked = w.checkMissileColWallBranch(missile, position, probe.Keep)

	// Hit-sound tail: upstream checks
	// missileData.hitSound != SfxID::None then PlaySfxLoc — audio is
	// engine-owned, so the hook carries the whole decision.
	if missile.Duration == 0 {
		w.HitCallbacks.PlayHitSound()
	}

	return res
}

// checkMissileColMonsterBranch ports the dMonster branch of CheckMissileCol
// (missiles.cpp:490-509): roll against the targeted monster and, on a hit,
// zero the duration unless KeepOnCollision and raise _miHitFlag.
func (w *HitWorld) checkMissileColMonsterBranch(missile *Missile, position TilePosition, probe MissileColProbe) bool {
	mid := w.DMonster[position.Y][position.X]
	if mid == 0 || !w.HitCallbacks.MonsterTargeted(mid, probe.OnlyHitWalking) {
		return false
	}

	monsterIndex := int(absI16(mid)) - 1

	isMonsterHit := w.rollMonsterHit(monsterIndex, missile, probe.DamageType,
		probe.MinDamage, probe.MaxDamage, probe.Shifted, probe.Keep)
	if isMonsterHit {
		if !probe.Keep {
			missile.Duration = 0
		}

		missile.HitFlag = true
	}

	return isMonsterHit
}

// checkMissileColPlayerBranch ports the dPlayer branch of CheckMissileCol
// (missiles.cpp:511-529): roll against the player found at the tile and
// apply the Hellfire blocked-shot rotation on a hit.
func (w *HitWorld) checkMissileColPlayerBranch(missile *Missile, position TilePosition, probe MissileColProbe,
	skipMoving ignoreMovingPlayers,
) (hit, blocked bool) {
	playerIdx, playerPresent := w.PlayerAtPosition(position.X, position.Y, skipMoving)
	if !playerPresent {
		return false, false
	}

	isPlayerHit, blocked := w.rollPlayerHit(playerIdx, missile, probe.DamageType,
		probe.MinDamage, probe.MaxDamage, probe.Shifted)
	if isPlayerHit {
		if w.IsHellfire && blocked {
			// Hellfire blocked shot: rotate the missile 45° left or
			// right instead of deleting it. The
			// rotation draw comes from w.Rnd: [mwlm2 fix] pass the
			// injected seam through — upstream's
			// RotateBlockedMissile(missile) draws GenerateRnd(2)
			// internally.
			RotateBlockedMissile(missile, w.Rnd)
		} else if !probe.Keep {
			missile.Duration = 0
		}

		missile.HitFlag = true
	}

	return isPlayerHit, blocked
}

// checkMissileColWallBranch ports the wall / breakable-object branch of
// CheckMissileCol (missiles.cpp:536-544).
func (w *HitWorld) checkMissileColWallBranch(missile *Missile, position TilePosition, keep KeepOnCollision) bool {
	if !IsMissileBlockedByTile(w.Grid, position.X, position.Y) {
		return false
	}

	if obj := w.Grid.FindObjectAtPosition(position.X, position.Y); obj != nil && obj.Breakable {
		w.HitCallbacks.BreakObjectMissile()
	}

	if !keep {
		missile.Duration = 0
	}

	missile.HitFlag = false

	return true
}

// rollMonsterHit is the monster-roll dispatch of CheckMissileCol
// trap, or monster-Caster on an opposing
// faction / berserk — the MonsterTrapHit roll; else TARGET_BOTH /
// TARGET_MONSTERS — the MonsterMHit roll. Extracted verbatim
// (statement and draw order preserved).
func (w *HitWorld) rollMonsterHit(monsterIndex int, missile *Missile,
	damageType DamageType, minDamage, maxDamage int, shifted bool, keepOnCollision KeepOnCollision,
) bool {
	if missile.IsTrap() ||
		(missile.Caster == int(TargetPlayers) &&
			w.HitCallbacks.MonsterFactionConflict(monsterIndex, missile.Source)) {
		// Trap, or monster-Caster on an opposing faction /
		// berserk: the trap damage roll.
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
// TARGET_MONSTERS hits other players
// (Plr2PlrMHit) unless self-fired; a monster side rolls PlayerMHit
// with the source monster; TARGET_BOTH/trap rolls PlayerMHit with the
// death reason following the source. Extracted verbatim (statement
// and draw order preserved).
func (w *HitWorld) rollPlayerHit(playerIdx int, missile *Missile,
	damageType DamageType, minDamage, maxDamage int, shifted bool,
) (hit, blocked bool) {
	if missile.Caster != int(TargetBoth) && !missile.IsTrap() {
		return w.rollPlayerHitByMonsterSource(playerIdx, missile, damageType, minDamage, maxDamage, shifted)
	}
	// TARGET_BOTH or trap: the death reason follows the source
	// .
	reason := DeathMonsterOrTrap
	if missile.SourceOf() == SourcePlayer {
		reason = DeathByPlayer
	}

	return w.HitCallbacks.PlayerMHit(nil, minDamage, maxDamage,
		missile.Dist, missile.Type, damageType, shifted, reason)
}

// rollPlayerHitByMonsterSource ports the Caster != TARGET_BOTH && !trap
// arm of the player branch: monsters hit other players via Plr2PlrMHit
// (unless self-fired), monster sources via PlayerMHit with the source
// monster (statement and draw order preserved).
func (w *HitWorld) rollPlayerHitByMonsterSource(playerIdx int, missile *Missile,
	damageType DamageType, minDamage, maxDamage int, shifted bool,
) (hit, blocked bool) {
	if missile.Caster != int(TargetMonsters) {
		// Monsters[missile._misource] — index only until the
		// monster table lands.
		ms := missile.Source

		return w.HitCallbacks.PlayerMHit(&ms, minDamage, maxDamage,
			missile.Dist, missile.Type, damageType, shifted, DeathMonsterOrTrap)
	}

	if playerIdx != missile.Source {
		// Plr2PlrMHit.
		return w.HitCallbacks.Plr2PlrMHit(minDamage, maxDamage,
			missile.Dist, missile.Type, damageType, shifted)
	}

	return false, false
}

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
