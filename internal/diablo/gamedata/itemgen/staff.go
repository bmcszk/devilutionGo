package itemgen

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// Spell ids the staff path compares against (tables/spelldat.h:30-69).
const (
	spellFirebolt    = 1
	spellResurrect   = 32
	spellTelekinesis = 33
	spellHealOther   = 34
	spellBloodStar   = 35
	spellNull        = 0
)

// staffPrefixCoinDenom is GetStaffPrefix's FlipCoin(10) (items.cpp:1098).
const staffPrefixCoinDenom = 10

// staffAffixCoinDenom is GetStaffSpell's vanilla-only FlipCoin(4) (:1239).
const staffAffixCoinDenom = 4

// spawnCappedStaffLevel is the gbIsSpawn staff power cap (items.cpp:1248).
const spawnCappedStaffLevel = 10

// spawnExcludedStaffSpells are the GetSpellStaffLevel -1 spells under
// gbIsSpawn (spells.cpp:326-337).
var spawnExcludedStaffSpells = [...]string{
	"StoneCurse", "Guardian", "Golem", "Apocalypse", "Elemental", "BloodStar", "BoneSpirit",
}

// SpellEnv is the global game-mode state the staff path reads.
type SpellEnv struct {
	Hellfire    bool
	Multiplayer bool
	Spawn       bool
}

// StaffRoll reports the GetStaffSpell outcome (items.cpp:1237-1284).
type StaffRoll struct {
	// AffixFallback is the vanilla-only FlipCoin(4) branch: the staff takes
	// ordinary staff affixes instead of a spell (:1239-1242).
	AffixFallback bool
	Spell         int
	Charges       int
	MinMag        int
	Value         int // charges*staffCost()/5 added to _ivalue (:1271-1275)
	MinLevel      int // GetItemPower minlvl (:1240 or lvl/2)
	MaxLevel      int // GetItemPower maxlvl (:1240 or lvl)
}

// StaffAffix reports the GetStaffPower outcome (items.cpp:1286-1308).
type StaffAffix struct {
	Prefix    Affix
	HasPrefix bool
	Bonus     AffixBonus
	Values    AffixValues
}

// GetStaffSpell ports items.cpp:1236-1284. spells must be the Null-first
// spelldat table, so spell ids index it directly.
func GetStaffSpell(spells []txtdata.SpellData, lvl int, env SpellEnv, rnd func(int32) int32) (StaffRoll, error) {
	if !env.Hellfire && flipCoin(rnd, staffAffixCoinDenom) {
		return StaffRoll{AffixFallback: true, MinLevel: lvl / 2, MaxLevel: lvl}, nil
	}

	roll, err := rollStaffSpell(spells, lvl, env, rnd)
	if err != nil {
		return StaffRoll{}, fmt.Errorf("staff spell at level %d: %w", lvl, err)
	}

	return roll, nil
}

// GetSpellStaffLevel ports spells.cpp:324-349.
func GetSpellStaffLevel(spells []txtdata.SpellData, s int, env SpellEnv) int {
	if env.Spawn && isSpawnExcludedStaffSpell(spells, s) {
		return -1
	}
	if s < 0 || s >= len(spells) {
		return -1
	}

	return spells[s].StaffLevel
}

// GetStaffPrefix ports items.cpp:1096-1101: a staff prefix is drawn unless the
// FlipCoin(10) fails on a staff that need not be good.
//
//nolint:revive // onlygood is upstream's GetStaffPrefix parameter, not control coupling.
func GetStaffPrefix(prefixes []Affix, maxlvl int, onlygood bool, rnd func(int32) int32) (Affix, bool) {
	if !flipCoin(rnd, staffPrefixCoinDenom) && !onlygood {
		return Affix{}, false
	}

	return SelectAffix(prefixes, AffixStaff, 0, maxlvl, onlygood, AlignAny, false, rnd)
}

// GetStaffPower ports the prefix half of GetStaffPower (items.cpp:1288-1297).
// The name/value bookkeeping and the final CalcItemValue are the caller's.
func GetStaffPower(
	prefixes []Affix, maxlvl int, onlygood bool, hero HeroStats, rnd func(int32) int32,
) (StaffAffix, error) {
	prefix, ok := GetStaffPrefix(prefixes, maxlvl, onlygood, rnd)
	if !ok {
		return StaffAffix{}, nil
	}

	out := StaffAffix{Prefix: prefix, HasPrefix: true}
	bonus, err := SaveItemAffix(&out.Values, prefix, hero, rnd)
	if err != nil {
		return StaffAffix{}, fmt.Errorf("staff prefix %q: %w", prefix.Name, err)
	}
	out.Bonus = bonus

	return out, nil
}

// walkStaffSpells ports the rv-decrement walk (items.cpp:1251-1257).
func walkStaffSpells(spells []txtdata.SpellData, l, rv int, env SpellEnv) int {
	best := spellNull
	for s := spellFirebolt; rv > 0; s = nextStaffSpell(spells, s, env) {
		if level := GetSpellStaffLevel(spells, s, env); level != -1 && l >= level {
			rv--
			best = s
		}
	}

	return best
}

// nextStaffSpell ports the s++ / single-player skip / wrap tail (items.cpp:1258-1265).
func nextStaffSpell(spells []txtdata.SpellData, s int, env SpellEnv) int {
	s++
	if !env.Multiplayer && s == spellResurrect {
		return spellTelekinesis
	}
	if !env.Multiplayer && s == spellHealOther {
		return spellBloodStar
	}
	if s == len(spells) {
		return spellFirebolt
	}

	return s
}

// rollStaffSpell picks the spell id and its charges (items.cpp:1245-1284).
func rollStaffSpell(spells []txtdata.SpellData, lvl int, env SpellEnv, rnd func(int32) int32) (StaffRoll, error) {
	l := max(lvl/2, 1)
	rv := int(rnd(int32(len(spells)))) + 1

	if env.Spawn && lvl > spawnCappedStaffLevel {
		lvl = spawnCappedStaffLevel
	}

	spell := walkStaffSpells(spells, l, rv, env)
	data, err := spellAt(spells, spell)
	if err != nil {
		return StaffRoll{}, err
	}

	// maxc = sStaffMax - minc + 1 (items.cpp:1269-1272).
	charges := data.StaffMin + int(rnd(int32(data.StaffMax-data.StaffMin+1)))

	return StaffRoll{
		Spell:    spell,
		Charges:  charges,
		MinMag:   data.MinInt,
		Value:    staffValue(data, charges),
		MinLevel: 0,
		MaxLevel: lvl,
	}, nil
}

// isSpawnExcludedStaffSpell matches the gbIsSpawn skip list by spell name.
func isSpawnExcludedStaffSpell(spells []txtdata.SpellData, s int) bool {
	if s < 0 || s >= len(spells) {
		return false
	}
	for _, name := range spawnExcludedStaffSpells {
		if spells[s].Name == name {
			return true
		}
	}

	return false
}

// staffValue ports _iCharges * staffCost() / 5 (items.cpp:1274-1275), where
// staffCost() = staffCost10 * 10 (tables/spelldat.h:251-254).
func staffValue(data txtdata.SpellData, charges int) int {
	return charges * data.StaffCost10 * 10 / 5
}

// spellAt returns the spell row for an id, guarding the table bounds.
func spellAt(spells []txtdata.SpellData, id int) (txtdata.SpellData, error) {
	if id < 0 || id >= len(spells) {
		return txtdata.SpellData{}, fmt.Errorf("spell id %d outside the table of %d", id, len(spells))
	}

	return spells[id], nil
}
