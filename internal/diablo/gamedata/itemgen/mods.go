package itemgen

// Modifier flag values are the ItemSpecialEffect bits SaveItemPower sets
// (tables/itemdat.h:389-414); the unused bits keep their upstream gaps.
const (
	FlagFireArrows         ModifierFlags = 1 << 3
	FlagFireDamage         ModifierFlags = 1 << 4
	FlagLightningDamage    ModifierFlags = 1 << 5
	FlagMultipleArrows     ModifierFlags = 1 << 9
	FlagKnockback          ModifierFlags = 1 << 11
	FlagStealMana3         ModifierFlags = 1 << 13
	FlagStealMana5         ModifierFlags = 1 << 14
	FlagStealLife3         ModifierFlags = 1 << 15
	FlagStealLife5         ModifierFlags = 1 << 16
	FlagQuickAttack        ModifierFlags = 1 << 17
	FlagFastAttack         ModifierFlags = 1 << 18
	FlagFasterAttack       ModifierFlags = 1 << 19
	FlagFastestAttack      ModifierFlags = 1 << 20
	FlagFastHitRecovery    ModifierFlags = 1 << 21
	FlagFasterHitRecovery  ModifierFlags = 1 << 22
	FlagFastestHitRecovery ModifierFlags = 1 << 23
	FlagFastBlock          ModifierFlags = 1 << 24
	FlagLightningArrows    ModifierFlags = 1 << 25
	FlagThorns             ModifierFlags = 1 << 26
	FlagNoMana             ModifierFlags = 1 << 27
	FlagHalfTrapDamage     ModifierFlags = 1 << 28
	FlagTripleDemonDamage  ModifierFlags = 1 << 30
	FlagZeroResistance     ModifierFlags = 1 << 31
)

// DamAcFlagDoppelganger is ItemSpecialEffectHf::Doppelganger (items.h:423),
// a bit in the separate dwDamAcFlags word rather than _iFlags.
const DamAcFlagDoppelganger uint8 = 1 << 4

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
	// Charges and MaxCharges are the staff counter the CHARGES and SPELL
	// powers write (:764-771).
	Charges    int
	MaxCharges int
	// Elemental damage bands (:772-800).
	FMinDam int
	FMaxDam int
	LMinDam int
	LMaxDam int
	// DamMod is _iPLDamMod (DAMMOD, :946) and EnAc is _iPLEnAc (TARGAC, :955).
	DamMod int
	EnAc   int
	// DamAcFlags is the dwDamAcFlags word (ItemSpecialEffectHf).
	DamAcFlags uint8
}

// ApplyItemPower ports the item-mutating switch of SaveItemPower
// (items.cpp:701-928) onto ItemModifiers, returning the same rolled value
// SaveItemPower reports. Powers the switch does not name leave mod untouched.
func ApplyItemPower(mod *ItemModifiers, affix Affix, env SpellEnv, rnd func(int32) int32) (AffixBonus, error) {
	r := RndPL(affix.Value1, affix.Value2, rnd)

	if applyTablePower(mod, affix, r, env) {
		return AffixBonus{Value: r}, nil
	}
	applySigned(mod, affix.Power, r, affix.Value1)

	return AffixBonus{Value: r}, nil
}

// applySigned ports the +/- stat deltas of SaveItemPower (items.cpp:701-928),
// split by the item field group each branch writes.
func applySigned(mod *ItemModifiers, power string, r, param int) {
	if applyCombatDelta(mod, power, r) || applyACDelta(mod, power, r) {
		return
	}
	if applyResistDelta(mod, power, r) || applyAttributeDelta(mod, power, r) {
		return
	}
	if applyAttribsDelta(mod, power, r) || applyVitalDelta(mod, power, r) {
		return
	}
	applySpecial(mod, power, r, param)
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
func applySpecial(mod *ItemModifiers, power string, r, param int) {
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
		// Upstream adds power.param1 here, not the RndPL roll
		// (items.cpp:864-866); the two differ only when param1 != param2.
		mod.Light += param
	case "LIGHT_CURSE":
		mod.Light -= param
	default:
		applyFlagPower(mod, power)
	}
}

// applyFlagPower is the flag-only tail of the switch (:848-928).
func applyFlagPower(mod *ItemModifiers, power string) {
	flags := map[string]ModifierFlags{
		"MULT_ARROWS": FlagMultipleArrows,
		"THORNS":      FlagThorns,
		"NOMANA":      FlagNoMana,
		"ABSHALFTRAP": FlagHalfTrapDamage,
		"KNOCKBACK":   FlagKnockback,
		"3XDAMVDEM":   FlagTripleDemonDamage,
		"ALLRESZERO":  FlagZeroResistance,
	}
	mod.Flags |= flags[power]
}

// applyTablePower covers the powers whose effect depends on the affix's raw
// param1/param2 rather than the RndPL roll, so they bypass the signed-delta
// chain (items.cpp:764-955).
func applyTablePower(mod *ItemModifiers, affix Affix, r int, env SpellEnv) bool {
	switch affix.Power {
	case "CHARGES":
		mod.Charges *= affix.Value1
		mod.MaxCharges = mod.Charges
	case "FIREDAM", "LIGHTDAM", "FIRE_ARROWS", "LIGHT_ARROWS":
		applyElemental(mod, affix)
	case "STEALMANA", "STEALLIFE":
		applyLeech(mod, affix)
	case "FASTATTACK", "FASTRECOVER", "FASTBLOCK", "DAMMOD", "TARGAC", "DOPPELGANGER":
		applySpeedPower(mod, affix, r, env)
	default:
		return false
	}

	return true
}

// applyElemental sets the elemental damage bands and clears the other element
// (:772-800).
func applyElemental(mod *ItemModifiers, affix Affix) {
	switch affix.Power {
	case "FIREDAM":
		mod.Flags |= FlagFireDamage
		mod.Flags &^= FlagLightningDamage
		mod.FMinDam, mod.FMaxDam = affix.Value1, affix.Value2
		mod.LMinDam, mod.LMaxDam = 0, 0
	case "LIGHTDAM":
		mod.Flags |= FlagLightningDamage
		mod.Flags &^= FlagFireDamage
		mod.LMinDam, mod.LMaxDam = affix.Value1, affix.Value2
		mod.FMinDam, mod.FMaxDam = 0, 0
	case "FIRE_ARROWS":
		mod.Flags |= FlagFireArrows | FlagFireDamage
		mod.Flags &^= FlagLightningArrows
		mod.FMinDam, mod.FMaxDam = affix.Value1, affix.Value2
		mod.LMinDam, mod.LMaxDam = 0, 0
	case "LIGHT_ARROWS":
		mod.Flags |= FlagLightningArrows | FlagLightningDamage
		mod.Flags &^= FlagFireArrows
		mod.LMinDam, mod.LMaxDam = affix.Value1, affix.Value2
		mod.FMinDam, mod.FMaxDam = 0, 0
	default:
		// applyTablePower only routes the four elemental powers here.
	}
}

// applyLeech maps the mana/life steal parameter onto its flag (:920-936).
func applyLeech(mod *ItemModifiers, affix Affix) {
	if affix.Power == "STEALMANA" {
		switch affix.Value1 {
		case 3:
			mod.Flags |= FlagStealMana3
		case 5:
			mod.Flags |= FlagStealMana5
		default:
			// param1 values other than 3/5 set no flag (:921-926).
		}

		return
	}
	switch affix.Value1 {
	case 3:
		mod.Flags |= FlagStealLife3
	case 5:
		mod.Flags |= FlagStealLife5
	default:
		// param1 values other than 3/5 set no flag (:930-936).
	}
}

// applySpeedPower covers the attack/recovery/block flags, DAMMOD and TARGAC
// (:946-956, :718-721).
func applySpeedPower(mod *ItemModifiers, affix Affix, r int, env SpellEnv) {
	switch affix.Power {
	case "FASTATTACK":
		mod.Flags |= attackSpeedFlag(affix.Value1)
	case "FASTRECOVER":
		mod.Flags |= recoveryFlag(affix.Value1)
	case "FASTBLOCK":
		mod.Flags |= FlagFastBlock
	case "DAMMOD":
		mod.DamMod += r
	case "TARGAC":
		if env.Hellfire {
			mod.EnAc = affix.Value1
		} else {
			mod.EnAc += r
		}
	case "DOPPELGANGER":
		mod.DamAcFlags |= DamAcFlagDoppelganger
	default:
		// applyTablePower only routes the six powers above here.
	}
}

// attackSpeedFlag maps param1 1..4 onto the attack-speed bits (:926-936).
func attackSpeedFlag(param1 int) ModifierFlags {
	flags := [...]ModifierFlags{FlagQuickAttack, FlagFastAttack, FlagFasterAttack, FlagFastestAttack}
	if param1 < 1 || param1 > len(flags) {
		return 0
	}

	return flags[param1-1]
}

// recoveryFlag maps param1 1..3 onto the hit-recovery bits (:937-945).
func recoveryFlag(param1 int) ModifierFlags {
	flags := [...]ModifierFlags{FlagFastHitRecovery, FlagFasterHitRecovery, FlagFastestHitRecovery}
	if param1 < 1 || param1 > len(flags) {
		return 0
	}

	return flags[param1-1]
}
