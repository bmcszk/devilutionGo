package itemgen

import (
	"fmt"
)

// toHitBonusLevels are CalculateToHitBonus' accepted levels (items.cpp:673-690).
var toHitBonusLevels = [...]struct{ level, lo, hi int }{
	{-50, 6, 10},
	{-25, 1, 5},
	{20, 1, 5},
	{36, 6, 10},
	{51, 11, 15},
	{66, 16, 20},
	{81, 21, 30},
	{96, 31, 40},
	{111, 41, 50},
	{126, 51, 75},
	{151, 76, 100},
}

// noBonusItemTypes are the GetItemBonus cases that draw no affix (items.cpp:1339-1342).
var noBonusItemTypes = [...]string{"None", "Misc", "Gold"}

// AffixValues ports the Item value accumulators _iVAdd1/_iVMult1/_iVAdd2/_iVMult2.
type AffixValues struct {
	Add1  int
	Mult1 int
	Add2  int
	Mult2 int
}

// AffixBonus is what SaveItemPower reports: the power's rolled value, plus the
// to-hit it adds to _iPLToHit for the TOHIT_DAMP variants (items.cpp:701-728).
type AffixBonus struct {
	Value int
	ToHit int
}

// ItemPowerRequest is the GetItemPower input: the base item identity plus the
// affix roll window.
type ItemPowerRequest struct {
	BaseName  string
	BaseValue int
	MinLevel  int
	MaxLevel  int
	Types     AffixItemType
	OnlyGood  bool
}

// ItemPowerResult is the composed GetItemPower outcome (items.cpp:1210-1234):
// the affixes drawn, their value slots, the magical name and the final value.
type ItemPowerResult struct {
	Magic  bool
	Name   string
	Value  int
	Values AffixValues
	Prefix *Affix
	Suffix *Affix
}

// BonusPlan is the GetItemBonus decision (items.cpp:1309-1344): which affix
// table the item type draws from, or the staff spell path instead.
type BonusPlan struct {
	MinLevel   int
	MaxLevel   int
	Flags      AffixItemType
	StaffSpell bool
	NoBonus    bool
}

// Store routes an affix value into slot 1, or slot 2 when slot 1 is taken
// (items.cpp:1055-1060).
func (v *AffixValues) Store(value, multVal int) {
	if v.Add1 != 0 || v.Mult1 != 0 {
		v.Add2 = value
		v.Mult2 = multVal

		return
	}
	v.Add1 = value
	v.Mult1 = multVal
}

// SaveItemAffix ports items.cpp:1048-1061: roll the power, scale it through
// PLVal, then route it into the first or second value slot.
func SaveItemAffix(values *AffixValues, affix Affix, rnd func(int32) int32) (AffixBonus, error) {
	bonus, err := SaveItemPower(affix, rnd)
	if err != nil {
		return AffixBonus{}, fmt.Errorf("save affix %q: %w", affix.Name, err)
	}

	value := PLVal(bonus.Value, affix.Value1, affix.Value2, affix.MinVal, affix.MaxVal)
	values.Store(value, affix.MultVal)

	return bonus, nil
}

// CalcItemValue ports items.cpp:600-610: the affix multipliers scale the base
// value, additive affixes then apply, and the result never drops below 1.
func CalcItemValue(values AffixValues, baseValue int) int {
	v := values.Mult1 + values.Mult2
	if v > 0 {
		v *= baseValue
	}
	if v < 0 {
		v = baseValue / v
	}
	v = values.Add1 + values.Add2 + v

	return max(v, 1)
}

// SaveItemPower ports the value part of SaveItemPower (items.cpp:701-728).
func SaveItemPower(affix Affix, rnd func(int32) int32) (AffixBonus, error) {
	r := RndPL(affix.Value1, affix.Value2, rnd)

	switch affix.Power {
	case "TOHIT_DAMP", "DOPPELGANGER":
		// DOPPELGANGER falls through into TOHIT_DAMP (items.cpp:718-721).
		r = RndPL(affix.Value1, affix.Value2, rnd)
		toHit, err := CalculateToHitBonus(affix.Value1, rnd)
		if err != nil {
			return AffixBonus{}, fmt.Errorf("power %q: %w", affix.Power, err)
		}

		return AffixBonus{Value: r, ToHit: toHit}, nil
	case "TOHIT_DAMP_CURSE":
		toHit, err := CalculateToHitBonus(-affix.Value1, rnd)
		if err != nil {
			return AffixBonus{}, fmt.Errorf("power %q: %w", affix.Power, err)
		}

		return AffixBonus{Value: r, ToHit: toHit}, nil
	}

	return AffixBonus{Value: r}, nil
}

// CalculateToHitBonus ports items.cpp:671-692. Upstream app_fatal's on an
// unknown level; the error is the library-safe equivalent.
func CalculateToHitBonus(level int, rnd func(int32) int32) (int, error) {
	for _, entry := range toHitBonusLevels {
		if entry.level != level {
			continue
		}
		bonus := RndPL(entry.lo, entry.hi, rnd)
		if entry.level < 0 {
			bonus = -bonus
		}

		return bonus, nil
	}

	return 0, fmt.Errorf("unknown to hit bonus level %d", level)
}

// RndPL ports RndPL (items.cpp:666-669).
func RndPL(param1, param2 int, rnd func(int32) int32) int {
	return param1 + int(rnd(int32(param2-param1+1)))
}

// PLVal ports PLVal (items.cpp:1039-1046): the rolled power value scaled onto
// the affix's [minVal, maxVal] band.
func PLVal(pv, p1, p2, minv, maxv int) int {
	if p1 == p2 {
		return minv
	}
	if minv == maxv {
		return minv
	}

	return minv + ((maxv - minv) * (100 * (pv - p1) / (p2 - p1)) / 100)
}

// GenerateMagicItemName ports items.cpp:1156-1177 (the untranslated format).
func GenerateMagicItemName(baseName string, prefix, suffix *Affix) string {
	switch {
	case prefix != nil && suffix != nil:
		return fmt.Sprintf("%s %s of %s", prefix.Name, baseName, suffix.Name)
	case prefix != nil:
		return fmt.Sprintf("%s %s", prefix.Name, baseName)
	case suffix != nil:
		return fmt.Sprintf("%s of %s", baseName, suffix.Name)
	}

	return baseName
}

// PlanItemBonus ports the item-type switch of GetItemBonus (items.cpp:1309-1344),
// including the min(minlvl, 25) cap at :1311.
//
//nolint:revive // allowspells is upstream's GetItemBonus parameter, not control coupling.
func PlanItemBonus(itemType string, minlvl, maxlvl int, allowspells bool) (BonusPlan, error) {
	plan := BonusPlan{MinLevel: min(minlvl, 25), MaxLevel: maxlvl}

	for _, none := range noBonusItemTypes {
		if itemType == none {
			plan.NoBonus = true

			return plan, nil
		}
	}

	flags, ok := bonusFlags(itemType)
	if !ok {
		return BonusPlan{}, fmt.Errorf("unknown item type %q", itemType)
	}
	plan.Flags = flags

	if flags == AffixStaff && allowspells {
		plan.StaffSpell = true
	}

	return plan, nil
}

// GetItemPower ports items.cpp:1210-1234: roll each affix inside its own
// callback, so the prefix pick and its SaveItemAffix rolls run before the suffix
// pick. The name and re-price follow once both callbacks returned.
func GetItemPower(
	prefixes, suffixes []Affix,
	req ItemPowerRequest,
	rnd func(int32) int32,
) (ItemPowerResult, error) {
	var (
		out     ItemPowerResult
		saveErr error
	)
	save := func(values *AffixValues, affix Affix) {
		if _, err := SaveItemAffix(values, affix, rnd); err != nil && saveErr == nil {
			saveErr = fmt.Errorf("affix %q: %w", affix.Name, err)
		}
	}
	walkItemPower(affixWalk{
		Prefixes: prefixes, Suffixes: suffixes,
		MinLevel: req.MinLevel, MaxLevel: req.MaxLevel, Types: req.Types, OnlyGood: req.OnlyGood,
		Rnd: rnd,
		OnPrefix: func(prefix Affix) {
			out.Magic = true
			out.Prefix = &prefix
			save(&out.Values, prefix)
		},
		OnSuffix: func(suffix Affix) {
			out.Magic = true
			out.Suffix = &suffix
			save(&out.Values, suffix)
		},
	})
	if saveErr != nil {
		return ItemPowerResult{}, saveErr
	}

	out.Name = GenerateMagicItemName(req.BaseName, out.Prefix, out.Suffix)
	out.Value = req.BaseValue
	if out.Magic {
		out.Value = CalcItemValue(out.Values, req.BaseValue)
	}

	return out, nil
}

// bonusFlags maps an itemdat itemType cell to its affix type (items.cpp:1314-1341).
func bonusFlags(itemType string) (AffixItemType, bool) {
	switch itemType {
	case "Sword", "Axe", "Mace":
		return AffixWeapon, true
	case "Bow":
		return AffixBow, true
	case "Shield":
		return AffixShield, true
	case "LightArmor", "Helm", "MediumArmor", "HeavyArmor":
		return AffixArmor, true
	case "Staff":
		return AffixStaff, true
	case "Ring", "Amulet":
		return AffixMisc, true
	}

	return 0, false
}
