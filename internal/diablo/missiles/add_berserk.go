// Package missiles: the Hellfire Berserk add body.
package missiles

// Berserk constants (monster.h:48,54; tables/monstdat.h:74,77; monster.h:37-58).
const (
	mflagGolem                = 1 << 5
	mflagBerserk              = 1 << 11
	resistMagic               = 1 << 0
	immuneMagic               = 1 << 3
	monsterModeFadeInBerserk  = 8
	monsterModeFadeOutBerserk = 9
	monsterModeChargeBerserk  = 14
	monsterAIDiablo           = 27
	spellBerserk              = 45
	berserkSearchRadius       = 5
)

// AddBerserk ports AddBerserk (missiles.cpp:1295-1344): the row always deletes
// itself and starts out fizzled; a trap caster stops there. Otherwise the
// closest eligible monster within 5 of the aim — not a minion, not already
// berserked, not unique or Diablo, not fading or charging, not magic immune,
// and passing the resist-magic coin flip — gets the berserk and golem flags,
// one d10-scaled damage bump per damage field, and a nest-sized light.
func AddBerserk(w *HitWorld, m *Missile, p *AddParameter) {
	m.DelFlag = true      // cpp:1297
	p.SpellFizzled = true // cpp:1298

	if m.SourceOf() == SourceTrap { // cpp:1300-1301
		return
	}

	spot, found := findBerserkTarget(w, p.Dst)
	if !found {
		return
	}

	monster := berserkMonsterAt(w, spot)
	level := w.ProcessCallbacks.PlayerSpellLevel(m.Source, spellBerserk) // cpp:1332

	w.ProcessCallbacks.SetMonsterFlags(monster, // cpp:1333
		w.ProcessCallbacks.MonsterFlags(monster)|mflagBerserk|mflagGolem)
	raiseBerserkDamage(w, monster, level) // cpp:1334-1337

	radius := uint8(3) // cpp:1340
	if w.ProcessCallbacks.IsNestLevel() {
		radius = 9
	}

	light := w.ProcessCallbacks.AddLight(spot.X, spot.Y, radius) // cpp:1341
	w.ProcessCallbacks.SetMonsterLight(monster, light)

	p.SpellFizzled = false // cpp:1342
}

// findBerserkTarget ports the FindClosestValidPosition call of AddBerserk
// (missiles.cpp:1303-1329).
func findBerserkTarget(w *HitWorld, dst TilePosition) (TilePosition, bool) {
	return findTrappedTileAt(w, dst, berserkSearchRadius, func(x, y int) bool {
		return berserkEligible(w, int(w.DMonster[y][x]))
	})
}

// berserkEligible is the monster predicate of AddBerserk (missiles.cpp:1304-1327).
func berserkEligible(w *HitWorld, cell int) bool {
	idx := int(absI16(int16(cell))) - 1
	if idx < 0 { // cpp:1309-1310
		return false
	}

	if w.ProcessCallbacks.MonsterIsPlayerMinion(idx) { // cpp:1312
		return false
	}

	if w.ProcessCallbacks.MonsterFlags(idx)&mflagBerserk != 0 { // cpp:1313-1314
		return false
	}

	if w.ProcessCallbacks.MonsterIsUnique(idx) || // cpp:1316-1317
		w.ProcessCallbacks.MonsterAI(idx) == monsterAIDiablo {
		return false
	}

	if berserkModeBlocked(w, idx) { // cpp:1318-1319
		return false
	}

	return !berserkResistBlocked(w, idx) // cpp:1320-1323
}

// berserkModeBlocked reports the fade/charge modes AddBerserk rejects
// (missiles.cpp:1318-1319).
func berserkModeBlocked(w *HitWorld, idx int) bool {
	switch w.ProcessCallbacks.MonsterMode(idx) {
	case monsterModeFadeInBerserk, monsterModeFadeOutBerserk, monsterModeChargeBerserk:
		return true
	default:
		return false
	}
}

// berserkResistBlocked is the magic-resistance gate of AddBerserk
// (missiles.cpp:1320-1323), including the resist-magic coin flip.
func berserkResistBlocked(w *HitWorld, idx int) bool {
	resistance := w.ProcessCallbacks.MonsterResistance(idx)
	if resistance&immuneMagic != 0 {
		return true
	}

	return resistance&resistMagic != 0 && (resistance&resistMagic != 1 || !flipCoin(w))
}

// raiseBerserkDamage ports the four damage writes of AddBerserk
// (missiles.cpp:1334-1337): each field takes its own d10 draw, scaled by 1.2
// and raised by the spell level.
func raiseBerserkDamage(w *HitWorld, monster, level int) {
	damage := w.ProcessCallbacks.MonsterDamage(monster)

	damage.Min = berserkDamageField(w, damage.Min, level)
	damage.Max = berserkDamageField(w, damage.Max, level)
	damage.MinSpecial = berserkDamageField(w, damage.MinSpecial, level)
	damage.MaxSpecial = berserkDamageField(w, damage.MaxSpecial, level)

	w.ProcessCallbacks.SetMonsterDamage(monster, damage)
}

// berserkDamageField scales one damage field (missiles.cpp:1334-1337).
func berserkDamageField(w *HitWorld, value, level int) int {
	return (w.ProcessCallbacks.GenerateRnd(10)+120)*value/100 + level
}

// berserkMonsterAt resolves the monster index standing on spot.
func berserkMonsterAt(w *HitWorld, spot TilePosition) int {
	return int(absI16(w.DMonster[spot.Y][spot.X])) - 1
}

// flipCoin ports FlipCoin (engine/random.cpp): a two-way draw, true on zero.
func flipCoin(w *HitWorld) bool {
	return w.ProcessCallbacks.GenerateRnd(2) == 0
}
