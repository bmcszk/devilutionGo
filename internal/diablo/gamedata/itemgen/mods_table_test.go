package itemgen_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestApplyItemPowerCharges(t *testing.T) {
	// given
	mod := itemgen.ItemModifiers{Charges: 5}

	// when
	err := applyRealAffix(t, &mod, "Plentiful") // CHARGES 2 2
	// then
	if err != nil {
		t.Fatalf("ApplyItemPower: %v", err)
	}
	if mod.Charges != 10 || mod.MaxCharges != 10 {
		t.Fatalf("charges = %d/%d, want 10/10 (param1 multiplier)", mod.Charges, mod.MaxCharges)
	}
}

func TestApplyItemPowerElementalDamage(t *testing.T) {
	tests := []struct {
		name      string
		affix     string
		wantSet   itemgen.ModifierFlags
		wantClear itemgen.ModifierFlags
		minDam    int
		maxDam    int
	}{
		{
			name: "fire", affix: "Flaming", wantSet: itemgen.FlagFireDamage,
			wantClear: itemgen.FlagLightningDamage, minDam: 1, maxDam: 10,
		},
		{
			name: "lightning", affix: "Lightning", wantSet: itemgen.FlagLightningDamage,
			wantClear: itemgen.FlagFireDamage, minDam: 2, maxDam: 20,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			mod := itemgen.ItemModifiers{Flags: itemgen.FlagFireDamage | itemgen.FlagLightningDamage}

			// when
			err := applyRealAffix(t, &mod, tc.affix)
			// then
			if err != nil {
				t.Fatalf("ApplyItemPower: %v", err)
			}
			assertElementalPower(t, mod, tc)
		})
	}
}

func TestApplyItemPowerArrows(t *testing.T) {
	tests := []struct {
		name      string
		affix     string
		wantSet   itemgen.ModifierFlags
		wantClear itemgen.ModifierFlags
	}{
		{
			name: "fire arrows", affix: "flame",
			wantSet: itemgen.FlagFireArrows | itemgen.FlagFireDamage, wantClear: itemgen.FlagLightningArrows,
		},
		{
			name: "lightning arrows", affix: "shock",
			wantSet: itemgen.FlagLightningArrows | itemgen.FlagLightningDamage, wantClear: itemgen.FlagFireArrows,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			mod := itemgen.ItemModifiers{Flags: itemgen.FlagFireArrows | itemgen.FlagLightningArrows}

			// when
			err := applyRealAffix(t, &mod, tc.affix)
			// then
			if err != nil {
				t.Fatalf("ApplyItemPower: %v", err)
			}
			assertFlags(t, mod.Flags, tc.wantSet, tc.wantClear)
		})
	}
}

func TestApplyItemPowerLeech(t *testing.T) {
	tests := []struct {
		name  string
		affix string
		want  itemgen.ModifierFlags
	}{
		{name: "mana 3", affix: "the bat", want: itemgen.FlagStealMana3},
		{name: "life 3", affix: "the leech", want: itemgen.FlagStealLife3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			var mod itemgen.ItemModifiers

			// when
			err := applyRealAffix(t, &mod, tc.affix)
			// then
			if err != nil {
				t.Fatalf("ApplyItemPower: %v", err)
			}
			assertFlags(t, mod.Flags, tc.want, mod.Flags&^tc.want)
		})
	}
}

func TestApplyItemPowerHasteAndDamageMod(t *testing.T) {
	tests := []struct {
		name      string
		affix     string
		wantFlags itemgen.ModifierFlags
		wantDam   int
	}{
		{name: "quick attack", affix: "readiness", wantFlags: itemgen.FlagQuickAttack},
		{name: "fast recovery", affix: "balance", wantFlags: itemgen.FlagFastHitRecovery},
		{name: "fast block", affix: "blocking", wantFlags: itemgen.FlagFastBlock},
		{name: "dam mod", affix: "quality", wantDam: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			var mod itemgen.ItemModifiers

			// when
			err := applyRealAffix(t, &mod, tc.affix)
			// then
			if err != nil {
				t.Fatalf("ApplyItemPower: %v", err)
			}
			assertHastePower(t, mod, tc.wantFlags, tc.wantDam)
		})
	}
}

func TestApplyItemPowerTargacEnv(t *testing.T) {
	tests := []struct {
		name string
		env  itemgen.SpellEnv
		draw int32
		want int
	}{
		{name: "hellfire sets param1", env: itemgen.SpellEnv{Hellfire: true}, draw: 5, want: 2},
		{name: "vanilla adds the roll", env: itemgen.SpellEnv{}, draw: 4, want: 6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			var mod itemgen.ItemModifiers
			rnd := &tapeRnd{t: t, values: []int32{tc.draw}}

			// when
			_, err := itemgen.ApplyItemPower(&mod, realAffix(t, "piercing"), tc.env, rnd.rnd)
			// then
			if err != nil {
				t.Fatalf("ApplyItemPower: %v", err)
			}
			assertEnAc(t, mod, tc.want)
		})
	}
}

func TestApplyItemPowerDoppelganger(t *testing.T) {
	// given
	var mod itemgen.ItemModifiers
	rnd := &tapeRnd{t: t, values: []int32{0}}
	affix := itemgen.Affix{Power: "DOPPELGANGER", Value1: 1, Value2: 1}

	// when
	_, err := itemgen.ApplyItemPower(&mod, affix, itemgen.SpellEnv{Hellfire: true}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("ApplyItemPower: %v", err)
	}
	if mod.DamAcFlags&itemgen.DamAcFlagDoppelganger == 0 {
		t.Fatalf("damAcFlags = %#x, want the ItemSpecialEffectHf Doppelganger bit", mod.DamAcFlags)
	}
}

// applyRealAffix looks the affix up in the committed tables and applies it.
func applyRealAffix(t *testing.T, mod *itemgen.ItemModifiers, name string) error {
	t.Helper()

	rnd := &tapeRnd{t: t, values: []int32{0}}
	_, err := itemgen.ApplyItemPower(mod, realAffix(t, name), itemgen.SpellEnv{}, rnd.rnd)
	if err != nil {
		return fmt.Errorf("applyRealAffix(%s): %w", name, err)
	}

	return nil
}

// realAffix finds the named affix in the committed prefix and suffix tables.
func realAffix(t *testing.T, name string) itemgen.Affix {
	t.Helper()

	for _, affix := range append(realPrefixes(t), realSuffixes(t)...) {
		if affix.Name == name {
			return affix
		}
	}
	t.Fatalf("affix %q not in the committed tables", name)

	return itemgen.Affix{}
}

// elementalBands is the fire/lightning pair set by an elemental power: the
// rolled band lands on its own element and the other side zeroes out.
func elementalBands(set itemgen.ModifierFlags, minDam, maxDam int) [4]int {
	if set == itemgen.FlagFireDamage {
		return [4]int{minDam, maxDam, 0, 0}
	}

	return [4]int{0, 0, minDam, maxDam}
}

// assertFlags checks one ItemSpecialEffect bit is present and another absent.
func assertFlags(t *testing.T, got, set, clears itemgen.ModifierFlags) {
	t.Helper()

	if got&set == 0 {
		t.Fatalf("flags = %#x, want %#x set", uint32(got), uint32(set))
	}
	if got&clears != 0 {
		t.Fatalf("flags = %#x, want %#x cleared", uint32(got), uint32(clears))
	}
}

// assertElementalPower checks both the element flag flip and the damage bands.
func assertElementalPower(t *testing.T, mod itemgen.ItemModifiers, tc struct {
	name      string
	affix     string
	wantSet   itemgen.ModifierFlags
	wantClear itemgen.ModifierFlags
	minDam    int
	maxDam    int
},
) {
	t.Helper()

	assertFlags(t, mod.Flags, tc.wantSet, tc.wantClear)
	bands := [4]int{mod.FMinDam, mod.FMaxDam, mod.LMinDam, mod.LMaxDam}
	if want := elementalBands(tc.wantSet, tc.minDam, tc.maxDam); bands != want {
		t.Fatalf("bands = %v, want %v", bands, want)
	}
}

func assertHastePower(t *testing.T, mod itemgen.ItemModifiers, wantFlags itemgen.ModifierFlags, wantDam int) {
	t.Helper()

	if mod.Flags != wantFlags || mod.DamMod != wantDam {
		t.Fatalf("flags = %#x damMod = %d, want %#x / %d",
			uint32(mod.Flags), mod.DamMod, uint32(wantFlags), wantDam)
	}
}

func assertEnAc(t *testing.T, mod itemgen.ItemModifiers, want int) {
	t.Helper()

	if mod.EnAc != want {
		t.Fatalf("enAc = %d, want %d", mod.EnAc, want)
	}
}
