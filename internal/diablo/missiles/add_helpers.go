// Package missiles: the private helpers of the add-time (Add*) bodies.
package missiles

// wholePixelMask is upstream's 0xFFFF0000 velocity mask
// (missiles.cpp:1902), kept as the signed int32 it masks with.
const wholePixelMask int32 = -65536

// SetAnimation ports Missile::setAnimation (missiles.cpp:76-107): the sheet
// id and the frame counters reset.
//
// Deviation: the per-direction AnimDelay/AnimLen columns
// (missiles.cpp:96-97) need the missile sprite table — not ported yet, so the
// caller's values are kept.
func SetAnimation(m *Missile, graphic GraphicID) {
	m.AnimType = int(graphic)
	m.AnimFrame = 1
	m.AnimCnt = 0
}

// addMissileLight ports the AddLight(start, 8) call of the projectile add
// bodies (missiles.cpp:1878, 1908).
func addMissileLight(w *HitWorld, m *Missile) int {
	return w.ProcessCallbacks.AddLight(m.Position.Start.X, m.Position.Start.Y, 8)
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

// spellLevelVelocity is std::min(_mispllvl * 2, 47) (missiles.cpp:1871).
func spellLevelVelocity(spellLevel int) int {
	if spellLevel*2 > 47 {
		return 47
	}

	return spellLevel * 2
}

// projectileDamage resolves the `if (_midam == 0)` block shared by the
// projectile add bodies (missiles.cpp:1879-1892, 1909-1921).
func projectileDamage(w *HitWorld, m *Missile) int {
	if m.Dam != 0 { // cpp:1879
		return m.Dam
	}

	switch m.SourceOf() {
	case SourcePlayer: // cpp:1882-1885
		player := w.AddSource.playerStats(m.Source)

		return w.ProcessCallbacks.GenerateRnd(10) + (player.Magic / 8) + m.SpellLevel + 1
	case SourceMonster: // cpp:1887-1888
		return projectileMonsterDamage(w, m)
	case SourceTrap: // cpp:1889-1890
		return projectileTrapDamage(w)
	default:
		return 0
	}
}

// projectileMonsterDamage ports ProjectileMonsterDamage
// (missiles.cpp:260-264): RandomIntBetween(monster.minDamage, monster.maxDamage).
func projectileMonsterDamage(w *HitWorld, m *Missile) int {
	monster := w.AddSource.monsterStats(m.Source)

	return randomIntBetween(w, monster.MinDamage, monster.MaxDamage)
}

// projectileTrapDamage ports ProjectileTrapDamage (missiles.cpp:266-269):
// currlevel + GenerateRnd(2 * currlevel).
func projectileTrapDamage(w *HitWorld) int {
	currlevel := w.ProcessCallbacks.Currlevel

	return currlevel + w.ProcessCallbacks.GenerateRnd(2*currlevel)
}

// magmaBallDuration ports the hellfire whole-pixel duration branch
// (missiles.cpp:1902-1905): in hellfire a missile whose velocity has no
// whole-pixel part — including the zero velocity of a same-tile spawn — lands
// immediately.
func magmaBallDuration(w *HitWorld, m *Missile) int {
	if !w.IsHellfire ||
		m.Position.Velocity.DeltaX&wholePixelMask != 0 ||
		m.Position.Velocity.DeltaY&wholePixelMask != 0 {
		return 256
	}

	return 1
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

// monsterStats reads one monster index, zero-valued when the sink is unset.
func (s AddSource) monsterStats(index int) AddMonsterStats {
	if s.MonsterStats == nil {
		return AddMonsterStats{}
	}

	return s.MonsterStats(index)
}

// hasItemFlag answers HasAnyOf(_pIFlags, flag); an unset query means no flags.
func (s AddPlayerStats) hasItemFlag(flag int) bool {
	if s.HasItemFlag == nil {
		return false
	}

	return s.HasItemFlag(flag)
}

// applyMagicMissileSheet ports the succubus-family setAnimation switch of
// AddGenericMagicMissile (missiles.cpp:2295-2306). Upstream uses four
// independent `if`s over mutually exclusive monster types, so the switch is
// equivalent.
func applyMagicMissileSheet(w *HitWorld, m *Missile) {
	monster := w.AddSource.monsterStats(m.Source)

	switch monster.Type {
	case MonsterTypeSuccubus:
		SetAnimation(m, GraphicBloodStar)
	case MonsterTypeSnowWich:
		SetAnimation(m, GraphicBloodStarBlue)
	case MonsterTypeHellSpawn:
		SetAnimation(m, GraphicBloodStarRed)
	case MonsterTypeSolbrner:
		SetAnimation(m, GraphicBloodStarYellow)
	default: // every other monster keeps the misdat graphic
	}
}

// magicMissileDamage resolves the `if (_midam == 0)` block of
// AddGenericMagicMissile (missiles.cpp:2312-2325): the player branch has no
// roll — 3 * spellLevel - magic / 8 + magic / 2.
func magicMissileDamage(w *HitWorld, m *Missile) int {
	if m.Dam != 0 {
		return m.Dam
	}

	switch m.SourceOf() {
	case SourcePlayer: // cpp:2315-2319
		player := w.AddSource.playerStats(m.Source)

		return 3*m.SpellLevel - (player.Magic / 8) + (player.Magic / 2)
	case SourceMonster: // cpp:2321-2322
		return projectileMonsterDamage(w, m)
	case SourceTrap: // cpp:2323-2324
		return projectileTrapDamage(w)
	default:
		return 0
	}
}

// acidDuration ports the duration branch of AddAcid
// (missiles.cpp:2332-2335): 5 * (monster intelligence + 4), or a single tick
// in hellfire when the velocity has no whole-pixel part.
func acidDuration(w *HitWorld, m *Missile) int {
	if !w.IsHellfire ||
		m.Position.Velocity.DeltaX&wholePixelMask != 0 ||
		m.Position.Velocity.DeltaY&wholePixelMask != 0 {
		monster := w.AddSource.monsterStats(m.Source)

		return 5 * (monster.Intelligence + 4)
	}

	return 1
}

// acidDamage resolves the `if (_midam == 0)` block of AddAcid
// (missiles.cpp:2339-2350): the player branch is a no-op upstream.
func acidDamage(w *HitWorld, m *Missile) int {
	if m.Dam != 0 {
		return m.Dam
	}

	switch m.SourceOf() {
	case SourceMonster: // cpp:2343-2344
		return projectileMonsterDamage(w, m)
	case SourceTrap: // cpp:2345-2346
		return projectileTrapDamage(w)
	case SourcePlayer: // cpp:2341-2342 — "not typically created by Players"
		return m.Dam
	default:
		return 0
	}
}
