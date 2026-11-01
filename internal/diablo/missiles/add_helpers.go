// Package missiles: the private helpers of the add-time (Add*) bodies.
package missiles

// wholePixelMask is upstream's 0xFFFF0000 velocity mask
// (missiles.cpp:1902), kept as the signed int32 it masks with.
const wholePixelMask int32 = -65536

// AddFireball ports AddFireball (missiles.cpp:1970-1996): speed 16 plus twice
// the spell level (capped at 34) for player casters, the level-and-2d10 damage
// scaled by the spell level, the 16-step frame group, light 8 on the launch
// tile and the start tile in var1/var2.
func AddFireball(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:1973-1975
		dst = shiftTile(dst, p.Midir)
	}

	sp := 16
	if m.Caster == int(TargetMonsters) { // cpp:1977-1984
		sp += min(m.SpellLevel*2, 34)
		m.Dam = fireballDamage(w, m)
	}

	UpdateMissileVelocity(m, dst, sp)                        // cpp:1985
	SetDirection16(m, GetDirection16(m.Position.Start, dst)) // cpp:1986
	m.Duration = 256                                         // cpp:1987
	m.Var1 = m.Position.Start.X                              // cpp:1988
	m.Var2 = m.Position.Start.Y                              // cpp:1989
	m.Lid = addMissileLight(w, m)                            // cpp:1990
}

// AddMissileExplosion ports AddMissileExplosion (missiles.cpp:2020-2047): the
// succubus-family sheet swap for monster casters, then the parent's position
// copied wholesale and the animation length as duration.
//
// Deviation: upstream asserts pParent != nullptr (cpp:2040); the port returns
// early instead, because a nil parent here is a caller bug the engine cannot
// recover from mid-frame.
func AddMissileExplosion(w *HitWorld, m *Missile, p *AddParameter) {
	applyExplosionSheet(w, m)

	if p.Parent == nil {
		return
	}

	m.Position.Tile = p.Parent.Position.Tile
	m.Position.Start = p.Parent.Position.Start
	m.Position.Offset = p.Parent.Position.Offset
	m.Position.Traveled = p.Parent.Position.Traveled
	m.Duration = m.AnimLen
}

// AddNovaBall ports AddNovaBall (missiles.cpp:1944-1957): speed 16, the
// RandomIntBetween(1, 8) start frame, duration 255, and var1/var2 anchored on
// the caster's own tile — or the launch tile for a sourceless missile.
func AddNovaBall(w *HitWorld, m *Missile, p *AddParameter) {
	UpdateMissileVelocity(m, p.Dst, 16)     // cpp:1946
	m.AnimFrame = randomIntBetween(w, 1, 8) // cpp:1947
	m.Duration = 255                        // cpp:1948
	position := novaBallAnchor(w, m)
	m.Var1 = position.X // cpp:1955
	m.Var2 = position.Y // cpp:1956
}

// AddLightning ports AddLightning (missiles.cpp:2001-2018): the launch tile
// snaps to the target, the spawned bolt inherits the parent's pixel position,
// the start frame is RandomIntBetween(1, 8), the lifetime depends on who cast
// it, and the light rides the pre-snap tile.
func AddLightning(w *HitWorld, m *Missile, p *AddParameter) {
	m.Position.Start = p.Dst // cpp:2003

	syncPositionWithParent(m, p) // cpp:2005

	m.AnimFrame = randomIntBetween(w, 1, 8) // cpp:2007

	if m.Caster == int(TargetPlayers) || m.IsTrap() { // cpp:2009-2014
		if m.IsTrap() || w.AddSource.monsterStats(m.Source).Type == MonsterTypeFamiliar {
			m.Duration = 8
		} else {
			m.Duration = 10
		}
	} else {
		m.Duration = m.SpellLevel/2 + 6 // cpp:2015
	}

	m.Lid = w.ProcessCallbacks.AddLight(m.Position.Tile.X, m.Position.Tile.Y, 4) // cpp:2017
}

// AddFireWall ports AddFireWall (missiles.cpp:1954-1969): the shifted
// GenerateRndSum(10, 2) damage plus the source level (currlevel for a trap),
// speed 16, the spell-level duration chain and the var1 burn-out anchor.
func AddFireWall(w *HitWorld, m *Missile, p *AddParameter) {
	m.Dam = generateRndSum(w, 10, 2) + 2 // cpp:1956
	if m.Source >= 0 {                   // cpp:1957
		m.Dam += w.AddSource.playerStats(m.Source).Level
	} else {
		m.Dam += w.ProcessCallbacks.Currlevel
	}

	m.Dam <<= 3 // cpp:1958

	UpdateMissileVelocity(m, p.Dst, 16) // cpp:1960

	m.Duration = 10 // cpp:1962
	if m.SpellLevel > 0 {
		m.Duration *= m.SpellLevel + 1 // cpp:1964
	}

	if m.Caster == int(TargetPlayers) || m.Source < 0 { // cpp:1965
		m.Duration += w.ProcessCallbacks.Currlevel
	}

	m.Duration *= 16                // cpp:1967
	m.Var1 = m.Duration - m.AnimLen // cpp:1968
}

// AddElementalArrow ports AddElementalArrow (missiles.cpp:1735-1775): the
// midir nudge on a same-tile shot, the class velocity chain with its hellfire
// attack-speed branch, the direction-16 frame group, the 256-tick duration and
// the 5-radius light.
func AddElementalArrow(w *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == dst { // cpp:1738-1740
		dst = shiftTile(dst, p.Midir)
	}

	av := 32
	if m.Caster == int(TargetMonsters) { // cpp:1742
		av = elementalArrowVelocity(w, m, av)
	}

	UpdateMissileVelocity(m, dst, av)                                              // cpp:1764
	SetDirection16(m, GetDirection16(m.Position.Start, dst))                       // cpp:1766
	m.Duration = 256                                                               // cpp:1767
	m.Var1 = m.Position.Start.X                                                    // cpp:1768
	m.Var2 = m.Position.Start.Y                                                    // cpp:1769
	m.Lid = w.ProcessCallbacks.AddLight(m.Position.Start.X, m.Position.Start.Y, 5) // cpp:1770
}

// AddApocalypseBoom ports AddApocalypseBoom (missiles.cpp:2455-2460): the boom
// lands on the target tile and lives for the row's animation length.
func AddApocalypseBoom(_ *HitWorld, m *Missile, p *AddParameter) {
	m.Position.Tile = p.Dst  // cpp:2457
	m.Position.Start = p.Dst // cpp:2458
	m.Duration = m.AnimLen   // cpp:2459
}

// AddFlameWave ports AddFlameWave (missiles.cpp:2170-2179): the level-scaled
// damage roll, speed 16, the 255-tick duration and the render adjustment that
// pushes the tile one step south and the sprite 32 pixels up.
func AddFlameWave(w *HitWorld, m *Missile, p *AddParameter) {
	m.Dam = randomIntLessThan(w, 10) + w.AddSource.playerStats(m.Source).Level + 1 // cpp:2172
	UpdateMissileVelocity(m, p.Dst, 16)                                            // cpp:2173
	m.Duration = 255                                                               // cpp:2174
	m.Position.Tile = shiftTile(m.Position.Tile, South)                            // cpp:2177
	m.Position.Offset.DeltaY -= 32                                                 // cpp:2178
}

// AddFlashTop ports AddFlashTop (missiles.cpp:2138-2152): the level-scaled
// damage roll of a monster-cast flash, the currlevel/2 trap damage, the
// pre-flag that draws it over walls, and the 19-tick flare.
func AddFlashTop(w *HitWorld, m *Missile, _ *AddParameter) {
	if m.Caster == int(TargetMonsters) { // cpp:2140
		if !m.IsTrap() {
			dmg := w.AddSource.playerStats(m.Source).Level + 1 // cpp:2142
			dmg += generateRndSum(w, 20, dmg)                  // cpp:2143
			m.Dam = scaleSpellEffect(dmg, m.SpellLevel)        // cpp:2144
			m.Dam += m.Dam / 2                                 // cpp:2145
		} else {
			m.Dam = w.ProcessCallbacks.Currlevel / 2 // cpp:2147
		}
	}

	m.PreFlag = true // cpp:2150
	m.Duration = 19  // cpp:2151
}

// AddResurrectBeam ports AddResurrectBeam (missiles.cpp:2749-2754): the beam
// straddles its destination tile and lives for the Resurrect graphic's frame-0
// animation length.
func AddResurrectBeam(_ *HitWorld, m *Missile, p *AddParameter) {
	m.Position.Tile = p.Dst   // cpp:2751
	m.Position.Start = p.Dst  // cpp:2752
	m.Duration = animLenOf(m) // cpp:2753
}

// AddInfravision ports AddInfravision (missiles.cpp:2554-2557): the effect
// lasts ScaleSpellEffect(1584, spllvl) ticks.
func AddInfravision(_ *HitWorld, m *Missile, _ *AddParameter) {
	m.Duration = scaleSpellEffect(1584, m.SpellLevel) // cpp:2556
}

// AddNova ports AddNova (missiles.cpp:2567-2581): the aim tile is stashed in
// var1/var2 for the render, the damage is a level-scaled six-dice roll for a
// player cast and a currlevel/2 three-dice roll for a trap, and the nova lives
// one tick.
func AddNova(w *HitWorld, m *Missile, p *AddParameter) {
	m.Var1 = p.Dst.X // cpp:2569
	m.Var2 = p.Dst.Y // cpp:2570

	if !m.IsTrap() {
		dmg := generateRndSum(w, 6, 5) + // cpp:2574
			w.AddSource.playerStats(m.Source).Level + 5
		m.Dam = scaleSpellEffect(dmg/2, m.SpellLevel) // cpp:2575
	} else {
		m.Dam = w.ProcessCallbacks.Currlevel/2 + generateRndSum(w, 3, 3) // cpp:2577
	}

	m.Duration = 1 // cpp:2580
}

// AddRedPortal ports AddRedPortal (missiles.cpp:2695-2701): the portal lives
// 100 ticks and flips to its idle frames AnimLen ticks before the end, which is
// the var1 threshold ProcessRedPortal compares against (cpp:4106).
func AddRedPortal(w *HitWorld, m *Missile, _ *AddParameter) {
	m.Duration = 100         // cpp:2697
	m.Var1 = 100 - m.AnimLen // cpp:2698
	PutMissile(w, m)         // cpp:2699
}

// AddAcidPuddle ports AddAcidPuddle (missiles.cpp:2355-2361): the puddle lives
// GenerateRnd(15) plus 40 per point of the source monster's intelligence + 1
// (cpp:2358), and it renders both over walls and over its own light.
func AddAcidPuddle(w *HitWorld, m *Missile, _ *AddParameter) {
	m.LightFlag = true // cpp:2357

	intelligence := w.AddSource.monsterStats(m.Source).Intelligence // cpp:2358
	m.Duration = w.ProcessCallbacks.GenerateRnd(15) + 40*(intelligence+1)

	m.PreFlag = true // cpp:2360
}

// animLenOf is the port's GetMissileSpriteData(id).animLen(0) read
// (missiles.cpp:2753): the row's frame-0 length, 0 when the caller supplied
// no timing table.
func animLenOf(m *Missile) int {
	if len(m.AnimLengths) == 0 {
		return 0
	}

	return m.AnimLengths[0]
}

// elementalArrowVelocity ports the TARGET_MONSTERS velocity chain of
// AddElementalArrow (missiles.cpp:1743-1762): the level/4 (Rogue) or level/8
// (Warrior, Bard) bonus, the hellfire attack-speed flags, and the non-hellfire
// -1 for those same three classes.
func elementalArrowVelocity(w *HitWorld, m *Missile, av int) int {
	player := w.AddSource.playerStats(m.Source)
	av += elementalClassBonus(player)

	if !w.IsHellfire { // cpp:1758-1761
		return elementalNonHellfireBonus(player.Class, av)
	}

	for _, flag := range arrowAttackSpeedFlags { // cpp:1749-1757
		if player.hasItemFlag(flag) {
			av += arrowAttackSpeedBonus[flag]
		}
	}

	return av
}

// elementalClassBonus is the level-scaled half of the chain
// (missiles.cpp:1744-1747).
func elementalClassBonus(player AddPlayerStats) int {
	switch player.Class {
	case HeroClassRogue:
		return player.Level / 4
	case HeroClassWarrior, HeroClassBard:
		return player.Level / 8
	default: // Sorcerer/Monk/Barbarian get no elemental-arrow bonus
		return 0
	}
}

// elementalNonHellfireBonus is the non-hellfire tail (missiles.cpp:1758-1761):
// Rogue, Warrior and Bard lose one point of the base speed.
func elementalNonHellfireBonus(class, av int) int {
	if class == HeroClassRogue || class == HeroClassWarrior || class == HeroClassBard {
		return av - 1
	}

	return av
}

// novaBallAnchor ports the anchor ternary of AddNovaBall (missiles.cpp:1949):
// a sourceless missile uses its own launch tile, otherwise the source player's
// current tile.
func novaBallAnchor(w *HitWorld, m *Missile) TilePosition {
	if m.Source < 0 {
		return m.Position.Start
	}

	return w.AddSource.playerStats(m.Source).Tile
}

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
		m.setAnimation(GraphicBloodStar)
	case MonsterTypeSnowWich:
		m.setAnimation(GraphicBloodStarBlue)
	case MonsterTypeHellSpawn:
		m.setAnimation(GraphicBloodStarRed)
	case MonsterTypeSolbrner:
		m.setAnimation(GraphicBloodStarYellow)
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

// applyExplosionSheet ports the monster-sheet swap of AddMissileExplosion
// (missiles.cpp:2022-2037): four mutually exclusive monster types, each with
// its own blood-star explosion sheet; other monsters keep the misdat graphic.
func applyExplosionSheet(w *HitWorld, m *Missile) {
	if m.Caster == int(TargetMonsters) || m.Source < 0 {
		return
	}

	switch w.AddSource.monsterStats(m.Source).Type {
	case MonsterTypeSuccubus:
		m.setAnimation(GraphicBloodStarExplosion)
	case MonsterTypeSnowWich:
		m.setAnimation(GraphicBloodStarBlueExplosion)
	case MonsterTypeHellSpawn:
		m.setAnimation(GraphicBloodStarRedExplosion)
	case MonsterTypeSolbrner:
		m.setAnimation(GraphicBloodStarYellowExplosion)
	default: // every other monster keeps the misdat graphic
	}
}

// scaleSpellEffect ports ScaleSpellEffect (missiles.cpp:119-126): each spell
// level adds an eighth of the running base, integer-truncated per step.
func scaleSpellEffect(base, spellLevel int) int {
	for range spellLevel {
		base += base / 8
	}

	return base
}

// generateRndSum ports GenerateRndSum (missiles.cpp:128-136): iterations
// draws of GenerateRnd(range), summed in draw order.
func generateRndSum(w *HitWorld, rng, iterations int) int {
	value := 0
	for range iterations {
		value += randomIntLessThan(w, rng)
	}

	return value
}

// fireballDamage ports the AddFireball damage term (missiles.cpp:1980-1983):
// 2*(level + 2d10) + 4, then scaled by the spell level.
func fireballDamage(w *HitWorld, m *Missile) int {
	player := w.AddSource.playerStats(m.Source)
	dmg := 2*(player.Level+generateRndSum(w, 10, 2)) + 4

	return scaleSpellEffect(dmg, m.SpellLevel)
}

// chargedBoltDamage ports the AddChargedBolt damage term (missiles.cpp:2701):
// a magic-quarter roll for player casters, a flat 15 for monsters. A quarter
// of 0..3 magic makes the range 0, and GenerateRnd returns 0 without advancing
// the seed for non-positive limits (Source/engine/random.cpp:98-99).
func chargedBoltDamage(w *HitWorld, m *Missile) int {
	if m.Caster != int(TargetMonsters) {
		return 15
	}

	magic := w.AddSource.playerStats(m.Source).Magic

	return randomIntLessThan(w, magic/4) + 1
}

// holyBoltDamage ports the AddHolyBolt damage roll (missiles.cpp:2734):
// GenerateRnd(10) + getCharacterLevel() + 9.
func holyBoltDamage(w *HitWorld, m *Missile) int {
	player := w.AddSource.playerStats(m.Source)

	return randomIntLessThan(w, 10) + player.Level + 9
}

// syncPositionWithParent ports SyncPositionWithParent (missiles.cpp:792-800):
// a child missile takes over the parent's pixel offset and traveled vector.
func syncPositionWithParent(m *Missile, p *AddParameter) {
	if p.Parent == nil {
		return
	}

	m.Position.Offset = p.Parent.Position.Offset
	m.Position.Traveled = p.Parent.Position.Traveled
}
