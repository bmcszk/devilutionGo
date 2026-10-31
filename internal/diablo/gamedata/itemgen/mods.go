package itemgen

// Modifier flag values follow items.h ItemSpecialEffect; only the bits
// SaveItemPower sets are listed (items.cpp:712-928).
const (
	FlagFireDamage ModifierFlags = 1 << iota
	FlagLightningDamage
	FlagMultipleArrows
	FlagFireArrows
	FlagLightningArrows
	FlagThorns
	FlagNoMana
	FlagHalfTrapDamage
	FlagKnockback
	FlagTripleDemonDamage
	FlagZeroResistance
	FlagStealMana3
	FlagStealMana5
	FlagStealLife3
	FlagStealLife5
	FlagDoppelganger
)

// ModifierFlags is the ItemSpecialEffect bit set SaveItemPower ORs in.
type ModifierFlags uint32

// ItemModifiers accumulates the item fields SaveItemPower writes
// (items.cpp:701-928), so the switch can be ported without an Item entity.
type ItemModifiers struct {
	ToHit      int
	Dam        int
	AC         int
	PLAC       int
	FireRes    int
	LightRes   int
	MagicRes   int
	Str        int
	Mag        int
	Dex        int
	Vit        int
	GetHit     int
	HP         int
	Mana       int
	MaxDur     int
	Durability int
	Light      int
	SplLvlAdd  int
	Flags      ModifierFlags
}

// ApplyItemPower ports the item-mutating switch of SaveItemPower
// (items.cpp:701-928) onto ItemModifiers, returning the same rolled value
// SaveItemPower reports. Powers the switch does not name leave mod untouched.
func ApplyItemPower(mod *ItemModifiers, affix Affix, rnd func(int32) int32) (AffixBonus, error) {
	r := RndPL(affix.Value1, affix.Value2, rnd)
	applySigned(mod, affix.Power, r)

	return AffixBonus{Value: r}, nil
}

// applySigned ports the +/- stat deltas of SaveItemPower (items.cpp:701-928),
// split by the item field group each branch writes.
func applySigned(mod *ItemModifiers, power string, r int) {
	if applyCombatDelta(mod, power, r) || applyACDelta(mod, power, r) {
		return
	}
	if applyResistDelta(mod, power, r) || applyAttributeDelta(mod, power, r) {
		return
	}
	if applyAttribsDelta(mod, power, r) || applyVitalDelta(mod, power, r) {
		return
	}
	applySpecial(mod, power, r)
}

// applyCombatDelta covers the ToHit and damage branches (:704-728).
func applyCombatDelta(mod *ItemModifiers, power string, r int) bool {
	switch power {
	case "TOHIT":
		mod.ToHit += r
	case "TOHIT_CURSE":
		mod.ToHit -= r
	case "DAMP":
		mod.Dam += r
	case "DAMP_CURSE":
		mod.Dam -= r
	default:
		return false
	}

	return true
}

// applyACDelta covers the armour-class and spell-level group (:730-758).
func applyACDelta(mod *ItemModifiers, power string, r int) bool {
	switch power {
	case "ACP":
		mod.PLAC += r
	case "ACP_CURSE":
		mod.PLAC -= r
	case "SETAC":
		mod.AC = r
	case "AC_CURSE":
		mod.AC -= r
	case "SORC", "SPLLVLADD":
		mod.SplLvlAdd = r
	default:
		return false
	}

	return true
}

// applyResistDelta covers the single-resistance branches (:744-752).
func applyResistDelta(mod *ItemModifiers, power string, r int) bool {
	switch power {
	case "FIRERES":
		mod.FireRes += r
	case "LIGHTRES": //nolint:misspell // upstream IPL_LIGHTRES spelling
		mod.LightRes += r
	case "MAGICRES":
		mod.MagicRes += r
	default:
		return false
	}

	return true
}

// applyAttributeDelta covers the single-stat branches (:792-812).
func applyAttributeDelta(mod *ItemModifiers, power string, r int) bool {
	switch power {
	case "STR":
		mod.Str += r
	case "STR_CURSE":
		mod.Str -= r
	case "MAG":
		mod.Mag += r
	case "MAG_CURSE":
		mod.Mag -= r
	case "DEX":
		mod.Dex += r
	case "DEX_CURSE":
		mod.Dex -= r
	case "VIT":
		mod.Vit += r
	case "VIT_CURSE":
		mod.Vit -= r
	default:
		return false
	}

	return true
}

// applyAttribsDelta is the four-stat pair (:814-824).
func applyAttribsDelta(mod *ItemModifiers, power string, r int) bool {
	switch power {
	case "ATTRIBS":
		mod.Str += r
		mod.Mag += r
		mod.Dex += r
		mod.Vit += r
	case "ATTRIBS_CURSE":
		mod.Str -= r
		mod.Mag -= r
		mod.Dex -= r
		mod.Vit -= r
	default:
		return false
	}

	return true
}

// applyVitalDelta covers the damage-on-hit, life and mana branches (:826-846).
func applyVitalDelta(mod *ItemModifiers, power string, r int) bool {
	switch power {
	case "GETHIT_CURSE":
		mod.GetHit += r
	case "GETHIT":
		mod.GetHit -= r
	case "LIFE":
		mod.HP += r
	case "LIFE_CURSE":
		mod.HP -= r
	case "MANA":
		mod.Mana += r
	case "MANA_CURSE":
		mod.Mana -= r
	default:
		return false
	}

	return true
}

// applySpecial covers the powers that write several fields or set flags
// (items.cpp:756-928).
func applySpecial(mod *ItemModifiers, power string, r int) {
	switch power {
	case "ALLRES":
		mod.FireRes = max(mod.FireRes+r, 0)
		mod.LightRes = max(mod.LightRes+r, 0)
		mod.MagicRes = max(mod.MagicRes+r, 0)
	case "DUR":
		bonus := r * mod.MaxDur / 100
		mod.MaxDur += bonus
		mod.Durability += bonus
	case "DUR_CURSE":
		mod.MaxDur = max(mod.MaxDur-r*mod.MaxDur/100, 1)
		mod.Durability = mod.MaxDur
	case "INDESTRUCTIBLE":
		mod.MaxDur = DurIndestructible
		mod.Durability = DurIndestructible
	case "LIGHT":
		mod.Light += r
	case "LIGHT_CURSE":
		mod.Light -= r
	default:
		applyFlagPower(mod, power)
	}
}

// applyFlagPower is the flag-only tail of the switch (:848-928).
func applyFlagPower(mod *ItemModifiers, power string) {
	flags := map[string]ModifierFlags{
		"MULT_ARROWS":  FlagMultipleArrows,
		"THORNS":       FlagThorns,
		"NOMANA":       FlagNoMana,
		"ABSHALFTRAP":  FlagHalfTrapDamage,
		"KNOCKBACK":    FlagKnockback,
		"3XDAMVDEM":    FlagTripleDemonDamage,
		"ALLRESZERO":   FlagZeroResistance,
		"DOPPELGANGER": FlagDoppelganger,
	}
	mod.Flags |= flags[power]
}
