// Package missiles: the add-time (Add*) bodies of the misdat table rows.
//
// The C++ Add* functions read the Players/Monsters globals, the lighting
// engine and the RNG (missiles.cpp:1772-2800). The players/monsters tables
// are not ported yet and entities.Player imports this package, so the reads
// are injected through AddSource (field of HitWorld) and the ProcessCallbacks
// sink — the hitCallbacks pattern of collision.go.
package missiles

// Item special-effect bits the add bodies query (tables/itemdat.h:386-414).
const (
	ItemSpecialEffectRandomArrowVelocity = 1 << 2
	ItemSpecialEffectQuickAttack         = 1 << 17
	ItemSpecialEffectFastAttack          = 1 << 18
	ItemSpecialEffectFasterAttack        = 1 << 19
	ItemSpecialEffectFastestAttack       = 1 << 20
)

// Hero classes the arrow velocity table branches on
// (tables/playerdat.hpp:21-26, enum class HeroClass).
const (
	HeroClassWarrior   = 0
	HeroClassRogue     = 1
	HeroClassSorcerer  = 2
	HeroClassMonk      = 3
	HeroClassBard      = 4
	HeroClassBarbarian = 5
)

// arrowAttackSpeedFlags is the hellfire attack-speed flag set of AddArrow,
// in upstream test order (missiles.cpp:1791-1792).
var arrowAttackSpeedFlags = []int{
	ItemSpecialEffectQuickAttack,
	ItemSpecialEffectFastAttack,
	ItemSpecialEffectFasterAttack,
	ItemSpecialEffectFastestAttack,
}

// arrowAttackSpeedBonus is the velocity increment per attack-speed flag
// (missiles.cpp:1791-1792: ++, += 2, += 4, += 8).
var arrowAttackSpeedBonus = map[int]int{
	ItemSpecialEffectQuickAttack:   1,
	ItemSpecialEffectFastAttack:    2,
	ItemSpecialEffectFasterAttack:  4,
	ItemSpecialEffectFastestAttack: 8,
}

// AddParameter ports AddMissileParameter (missiles.cpp:2795-2800): the pack
// AddMissile hands to the row's addFn, plus the spellFizzled output flag.
type AddParameter struct {
	// Dst is the target tile the motion is aimed at (parameter.dst).
	Dst TilePosition
	// Midir is the missile direction (parameter.midir).
	Midir Direction
	// Parent is the missile that spawned this one (parameter.parent).
	Parent *Missile
	// SpellFizzled tells AddMissile to drop the missile (parameter.spellFizzled).
	SpellFizzled bool
}

// AddPlayerStats is the add-time player read of one player index.
type AddPlayerStats struct {
	// Class is _pClass (tables/playerdat.hpp HeroClass).
	Class int
	// Level is getCharacterLevel().
	Level int
	// Magic is _pMagic.
	Magic int
	// HasItemFlag answers HasAnyOf(_pIFlags, flag) (tables/itemdat.h:386).
	HasItemFlag func(flag int) bool
}

// AddMonsterStats is the add-time monster read of one monster index.
type AddMonsterStats struct {
	// Type is type().type (tables/monstdat.h:160+).
	Type int
	// Intelligence is intelligence (monster.h:258).
	Intelligence int
	// MinDamage/MaxDamage are the projectile damage range (missiles.cpp:260-264).
	MinDamage int
	MaxDamage int
}

// AddSource is the injected replacement for the Players/Monsters globals the
// add bodies read.
type AddSource struct {
	PlayerStats  func(index int) AddPlayerStats
	MonsterStats func(index int) AddMonsterStats
}

// NoopAddSource returns the no-op defaults (zero stats, no item flags).
func NoopAddSource() AddSource {
	return AddSource{
		PlayerStats:  func(int) AddPlayerStats { return AddPlayerStats{} },
		MonsterStats: func(int) AddMonsterStats { return AddMonsterStats{} },
	}
}

// AddArrow ports AddArrow (missiles.cpp:1772-1797): velocity from the class
// and level table (plus the hellfire attack-speed flags), the direction-16
// frame, and the 256-tick duration.
func AddArrow(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:1775-1777
		dst = shiftTile(dst, p.Midir)
	}

	av := 32
	if m.Caster == int(TargetMonsters) { // cpp:1779
		av = arrowVelocity(w, m, av)
	}

	UpdateMissileVelocity(m, dst, av)                            // cpp:1793
	m.AnimFrame = int(GetDirection16(m.Position.Start, dst)) + 1 // cpp:1794
	m.Duration = 256                                             // cpp:1795
}

// arrowVelocity applies the TARGET_MONSTERS velocity chain
// (missiles.cpp:1780-1792): the RandomArrowVelocity roll, the class level
// bonus and — in hellfire only — the attack-speed item flags.
func arrowVelocity(w *HitWorld, m *Missile, av int) int {
	player := w.AddSource.playerStats(m.Source)

	if player.hasItemFlag(ItemSpecialEffectRandomArrowVelocity) { // cpp:1782-1784
		av = randomIntBetween(w, 16, 47)
	}

	switch player.Class {
	case HeroClassRogue: // cpp:1785-1786
		av += (player.Level - 1) / 4
	case HeroClassWarrior, HeroClassBard: // cpp:1787-1788
		av += (player.Level - 1) / 8
	default: // Sorcerer/Monk/Barbarian get no arrow bonus
	}

	if !w.IsHellfire { // cpp:1790-1792
		return av
	}

	// cpp:1791-1792 — four independent `if`s, so the sum is equivalent.
	for _, flag := range arrowAttackSpeedFlags {
		if player.hasItemFlag(flag) {
			av += arrowAttackSpeedBonus[flag]
		}
	}

	return av
}

// shiftTile applies `tile += dir` (missiles.cpp:1776) through the Direction
// delta table of the motion pipeline.
func shiftTile(tile TilePosition, dir Direction) TilePosition {
	deltaX, deltaY := directionDeltaOf(dir)

	return TilePosition{X: tile.X + deltaX, Y: tile.Y + deltaY}
}

// randomIntLessThan ports RandomIntLessThan (engine/random.hpp:391-394):
// std::max(GenerateRnd(v), 0).
func randomIntLessThan(w *HitWorld, v int) int {
	drawn := w.ProcessCallbacks.GenerateRnd(v)
	if drawn < 0 {
		return 0
	}

	return drawn
}

// randomIntBetween ports RandomIntBetween (engine/random.hpp:403-406):
// RandomIntLessThan(max-min+1) + min.
func randomIntBetween(w *HitWorld, minValue, maxValue int) int {
	return randomIntLessThan(w, maxValue-minValue+1) + minValue
}

// playerStats reads one player index, zero-valued when the sink is unset.
func (s AddSource) playerStats(index int) AddPlayerStats {
	if s.PlayerStats == nil {
		return AddPlayerStats{}
	}

	return s.PlayerStats(index)
}

// hasItemFlag answers HasAnyOf(_pIFlags, flag); an unset query means no flags.
func (s AddPlayerStats) hasItemFlag(flag int) bool {
	if s.HasItemFlag == nil {
		return false
	}

	return s.HasItemFlag(flag)
}
