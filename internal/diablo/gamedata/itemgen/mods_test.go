package itemgen_test

import (
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestApplyItemPowerSignedDeltas(t *testing.T) {
	// given
	mod := itemgen.ItemModifiers{MaxDur: 100, Durability: 100}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0, 0}}

	// when
	for _, power := range []string{"TOHIT", "TOHIT_CURSE", "ATTRIBS", "ATTRIBS_CURSE"} {
		affix := itemgen.Affix{Power: power, Value1: 5, Value2: 5}
		if _, err := itemgen.ApplyItemPower(&mod, affix, itemgen.SpellEnv{}, rnd.rnd); err != nil {
			t.Fatalf("ApplyItemPower(%s): %v", power, err)
		}
	}

	// then
	if mod.ToHit != 0 {
		t.Fatalf("ToHit = %d, want 0 (add then subtract)", mod.ToHit)
	}
	if mod.Str != 0 || mod.Mag != 0 || mod.Dex != 0 || mod.Vit != 0 {
		t.Fatalf("attribs = %+v, want zero after the curse", mod)
	}
}

func TestApplyItemPowerSpecialWrites(t *testing.T) {
	// given
	mod := itemgen.ItemModifiers{MaxDur: 100, Durability: 100}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0}}

	// when
	errs := applyPowers(t, &mod, []itemgen.Affix{
		{Power: "ALLRES", Value1: 10, Value2: 10},
		{Power: "DUR", Value1: 50, Value2: 50},
		{Power: "INDESTRUCTIBLE"},
	}, rnd.rnd)

	// then
	if err := errors.Join(errs...); err != nil {
		t.Fatalf("ApplyItemPower: %v", err)
	}

	if mod.FireRes != 10 || mod.LightRes != 10 || mod.MagicRes != 10 {
		t.Fatalf("resists = %d/%d/%d, want 10 each", mod.FireRes, mod.LightRes, mod.MagicRes)
	}
	if mod.MaxDur != itemgen.DurIndestructible || mod.Durability != itemgen.DurIndestructible {
		t.Fatalf("dur = %d/%d, want 255 after INDESTRUCTIBLE", mod.MaxDur, mod.Durability)
	}
}

func TestApplyItemPowerFlags(t *testing.T) {
	// given
	mod := itemgen.ItemModifiers{}
	rnd := &tapeRnd{t: t, values: []int32{0, 0}}

	// when
	if _, err := itemgen.ApplyItemPower(&mod, itemgen.Affix{Power: "THORNS"}, itemgen.SpellEnv{}, rnd.rnd); err != nil {
		t.Fatalf("THORNS: %v", err)
	}
	if _, err := itemgen.ApplyItemPower(&mod, itemgen.Affix{Power: "KNOCKBACK"}, itemgen.SpellEnv{}, rnd.rnd); err != nil {
		t.Fatalf("KNOCKBACK: %v", err)
	}

	// then
	if mod.Flags&itemgen.FlagThorns == 0 || mod.Flags&itemgen.FlagKnockback == 0 {
		t.Fatalf("flags = %#x, want Thorns and Knockback set", mod.Flags)
	}
	if mod.Flags&itemgen.FlagNoMana != 0 {
		t.Fatalf("flags = %#x, want NoMana clear", mod.Flags)
	}
}

// applyPowers runs a power list against mod and returns every error.
func applyPowers(t *testing.T, mod *itemgen.ItemModifiers, powers []itemgen.Affix, rnd func(int32) int32) []error {
	t.Helper()

	errs := make([]error, 0, len(powers))
	for _, affix := range powers {
		_, err := itemgen.ApplyItemPower(mod, affix, itemgen.SpellEnv{}, rnd)
		errs = append(errs, err)
	}

	return errs
}
