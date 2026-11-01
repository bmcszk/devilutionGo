package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
	"github.com/bmcszk/devilutionGo/internal/engine"
)

// tapeRnd replays scripted GenerateRnd values and records every argument.
type tapeRnd struct {
	t      *testing.T
	values []int32
	args   []int32
}

// bonusCase is one PlanItemBonus expectation.
type bonusCase struct {
	name        string
	itemType    string
	minlvl      int
	allowspell  bool
	wantMin     int
	wantFlags   itemgen.AffixItemType
	wantStaff   bool
	wantNoBonus bool
}

func (r *tapeRnd) rnd(n int32) int32 {
	r.t.Helper()

	r.args = append(r.args, n)
	if len(r.values) == 0 {
		r.t.Fatalf("tapeRnd: draw %d (arg %d) beyond the script", len(r.args), n)
	}
	v := r.values[0]
	r.values = r.values[1:]

	return v
}

func TestSelectAffixPoolWithDuplicates(t *testing.T) {
	// given
	prefixes := realPrefixes(t)

	// when
	pool := itemgen.EligibleAffixes(prefixes, itemgen.AffixWeapon, 1, 6, false, itemgen.AlignAny, false)

	// then
	want := []string{
		"Tin", "Tin", "Brass", "Brass", "Bronze", "Bronze", "Iron", "Iron",
		"Steel", "Steel", "Useless", "Useless", "Bent", "Bent", "Weak", "Weak",
		"Jagged", "Jagged", "Deadly", "Deadly", "Clumsy", "Clumsy", "Dull", "Dull",
		"Sharp", "Sharp", "Fine", "Fine", "Red", "Blue", "White",
	}
	assertNames(t, pool, want)
}

func TestSelectAffixSeededPick(t *testing.T) {
	// given
	prefixes := realPrefixes(t)
	rnd := seededRnd(1) // seed 1: GenerateRnd(31) = 5

	// when
	pick, ok := itemgen.SelectAffix(prefixes, itemgen.AffixWeapon, 1, 6, false, itemgen.AlignAny, false, rnd)

	// then
	if !ok {
		t.Fatal("SelectAffix: empty pool, want a pick")
	}
	if pick.Name != "Bronze" {
		t.Fatalf("pick = %q, want Bronze (pool index 5)", pick.Name)
	}
	if pick.Power != "TOHIT" || pick.Value1 != 1 || pick.Value2 != 5 || pick.MultVal != 2 {
		t.Fatalf("pick = %+v, want the Bronze row", pick)
	}
}

func TestSelectAffixFiltersByTypeAndLevel(t *testing.T) {
	// given
	prefixes := realPrefixes(t)

	// when
	armor := itemgen.EligibleAffixes(prefixes, itemgen.AffixArmor, 1, 6, false, itemgen.AlignAny, false)
	outOfRange := itemgen.EligibleAffixes(prefixes, itemgen.AffixWeapon, 61, 99, false, itemgen.AlignAny, false)

	// then
	if len(armor) == 0 {
		t.Fatal("armor pool is empty, want the level-1..6 armor affixes")
	}
	assertAffixesMatch(t, armor, itemgen.AffixArmor, 1, 6)
	if len(outOfRange) != 0 {
		t.Fatalf("out-of-range pool size = %d, want 0", len(outOfRange))
	}
}

func TestSelectAffixEmptyPool(t *testing.T) {
	// given
	prefixes := realPrefixes(t)

	// when
	pick, ok := itemgen.SelectAffix(prefixes, itemgen.AffixWeapon, 61, 99, false, itemgen.AlignAny, false, seededRnd(1))

	// then
	if ok {
		t.Fatal("SelectAffix ok = true, want false (no eligible affix)")
	}
	if pick.Name != "" {
		t.Fatalf("pick = %+v, want the zero Affix", pick)
	}
}

func TestEligibleAffixesChanceWeightCap(t *testing.T) {
	// given
	affixes := []itemgen.Affix{{Name: "Heavy", Types: itemgen.AffixWeapon, MinLevel: 1, Chance: 300}}

	// when
	pool := itemgen.EligibleAffixes(affixes, itemgen.AffixWeapon, 1, 6, false, itemgen.AlignAny, false)

	// then
	if len(pool) != 256 {
		t.Fatalf("pool size = %d, want the StaticVector cap 256", len(pool))
	}
}

func TestSelectAffixStaffChargesFilter(t *testing.T) {
	// given
	affixes := []itemgen.Affix{
		{Name: "of Charging", Power: "CHARGES", Types: itemgen.AffixStaff, MinLevel: 1, Chance: 1},
		{Name: "of Fire", Power: "FIREDAM", Types: itemgen.AffixStaff, MinLevel: 1, Chance: 1},
	}

	// when
	withCharges := itemgen.EligibleAffixes(affixes, itemgen.AffixStaff, 1, 6, false, itemgen.AlignAny, false)
	withoutCharges := itemgen.EligibleAffixes(affixes, itemgen.AffixStaff, 1, 6, false, itemgen.AlignAny, true)

	// then
	if len(withCharges) != 2 {
		t.Fatalf("charges kept pool = %d, want 2", len(withCharges))
	}
	if len(withoutCharges) != 1 || withoutCharges[0].Name != "of Fire" {
		t.Fatalf("charges excluded pool = %+v, want only of Fire", withoutCharges)
	}
}

func TestGetItemPowerPrefixAndSuffixAffixLock(t *testing.T) {
	// given
	prefixes, suffixes := lockAffixes()

	// when
	locked := runLoadout(t, prefixes, suffixes, []int32{0, 1, 0, 0, 0})
	unlocked := runLoadout(t, prefixes, suffixes, []int32{0, 1, 0, 1, 0})

	// then
	if locked.Prefix == nil || locked.Prefix.Alignment != itemgen.AlignEvil {
		t.Fatalf("prefix = %+v, want the evil Tin", locked.Prefix)
	}
	if locked.Suffix == nil || locked.Suffix.Name != "tears" {
		t.Fatalf("suffix = %+v, want tears (good health filtered by the evil prefix)", locked.Suffix)
	}
	if unlocked.Prefix.Name != "Bronze" || unlocked.Suffix.Name != "health" {
		t.Fatalf("unlocked pair = %q/%q, want Bronze/health (no lock from an Any prefix)",
			unlocked.Prefix.Name, unlocked.Suffix.Name)
	}
}

func TestGetItemPowerPrefixAndSuffixDrawOrder(t *testing.T) {
	// given
	prefixes, suffixes := lockAffixes()
	rnd := &tapeRnd{t: t, values: []int32{0, 1, 0, 0, 0}}

	// when
	itemgen.GetItemPowerPrefixAndSuffix(prefixes, suffixes, 1, 6, itemgen.AffixWeapon, false, rnd.rnd)

	// then
	// prefix coin, suffix coin, onlygood coin, prefix pick, suffix pick.
	assertDraws(t, rnd.args, []int32{4, 3, 3, 2, 2})
}

func TestCalcItemValueBranches(t *testing.T) {
	// given
	cases := []struct {
		name   string
		values itemgen.AffixValues
		base   int
		want   int
	}{
		{"positive multiplier scales the base", itemgen.AffixValues{Mult1: 2}, 1000, 2000},
		{"two multipliers add before scaling", itemgen.AffixValues{Mult1: 2, Mult2: 3}, 1000, 5000},
		{"negative multiplier divides the base", itemgen.AffixValues{Mult1: -3, Add1: 1000}, 1000, 667},
		{"negative result floors at 1", itemgen.AffixValues{Mult1: -3}, 1000, 1},
		{"additive affixes shift the result", itemgen.AffixValues{Add1: 300, Mult1: 2}, 1000, 2300},
		{"no affix slots still floor at 1", itemgen.AffixValues{}, 500, 1},
	}

	// when
	got := make([]int, len(cases))
	for i, tc := range cases {
		got[i] = itemgen.CalcItemValue(tc.values, tc.base)
	}

	// then
	for i, tc := range cases {
		if got[i] != tc.want {
			t.Fatalf("%s: CalcItemValue(%+v, %d) = %d, want %d", tc.name, tc.values, tc.base, got[i], tc.want)
		}
	}
}

func TestPLValBands(t *testing.T) {
	// when
	sameRoll := itemgen.PLVal(10, 6, 6, 100, 500)
	fixedBand := itemgen.PLVal(9, 1, 10, 700, 700)
	scaled := itemgen.PLVal(10, 1, 20, 200, 1000)

	// then
	if sameRoll != 100 {
		t.Fatalf("PLVal on a single-valued power = %d, want minVal 100", sameRoll)
	}
	if fixedBand != 700 {
		t.Fatalf("PLVal on a fixed band = %d, want 700", fixedBand)
	}
	if scaled != 576 {
		t.Fatalf("PLVal(10 of 1..20 onto 200..1000) = %d, want 576", scaled)
	}
}

func TestSaveItemAffixSlots(t *testing.T) {
	// given
	var values itemgen.AffixValues
	first := itemgen.Affix{Name: "Bronze", Power: "TOHIT", Value1: 1, Value2: 5, MinVal: 100, MaxVal: 500, MultVal: 2}
	second := itemgen.Affix{Name: "quality", Power: "DAMMOD", Value1: 1, Value2: 2, MinVal: 200, MaxVal: 300, MultVal: 3}
	rnd := seededRnd(7)

	// when
	saved := saveBoth(t, &values, first, second, rnd)

	// then
	if saved.Add1 < 100 || saved.Add1 > 500 || saved.Mult1 != 2 {
		t.Fatalf("slot 1 = %+v, want the Bronze value in Add1/Mult1", saved)
	}
	if values.Add1 != saved.Add1 || values.Mult1 != 2 {
		t.Fatalf("slot 1 changed to %+v, want it left alone", values)
	}
	if values.Add2 < 200 || values.Add2 > 300 || values.Mult2 != 3 {
		t.Fatalf("slot 2 = %+v, want the quality value in Add2/Mult2", values)
	}
}

func TestGenerateMagicItemName(t *testing.T) {
	// given
	prefix := itemgen.Affix{Name: "Bronze"}
	suffix := itemgen.Affix{Name: "quality"}

	// when
	both := itemgen.GenerateMagicItemName("Sword", &prefix, &suffix)
	onlyPrefix := itemgen.GenerateMagicItemName("Sword", &prefix, nil)
	onlySuffix := itemgen.GenerateMagicItemName("Sword", nil, &suffix)
	plain := itemgen.GenerateMagicItemName("Sword", nil, nil)

	// then
	if both != "Bronze Sword of quality" {
		t.Fatalf("both = %q, want Bronze Sword of quality", both)
	}
	if onlyPrefix != "Bronze Sword" {
		t.Fatalf("prefix only = %q, want Bronze Sword", onlyPrefix)
	}
	if onlySuffix != "Sword of quality" {
		t.Fatalf("suffix only = %q, want Sword of quality", onlySuffix)
	}
	if plain != "Sword" {
		t.Fatalf("plain = %q, want Sword", plain)
	}
}

func TestPlanItemBonusMinLevelCap(t *testing.T) {
	// given
	cases := []bonusCase{
		{"sword caps minlvl at 25", "Sword", 40, false, 25, itemgen.AffixWeapon, false, false},
		{"sword below the cap is kept", "Sword", 12, false, 12, itemgen.AffixWeapon, false, false},
		{"bow uses the bow table", "Bow", 5, false, 5, itemgen.AffixBow, false, false},
		{"helm uses the armor table", "Helm", 5, false, 5, itemgen.AffixArmor, false, false},
		{"ring uses the misc table", "Ring", 5, false, 5, itemgen.AffixMisc, false, false},
		{"staff with spells takes the staff spell path", "Staff", 9, true, 9, itemgen.AffixStaff, true, false},
		{"staff without spells uses staff affixes", "Staff", 9, false, 9, itemgen.AffixStaff, false, false},
		{"gold draws no bonus", "Gold", 9, true, 9, 0, false, true},
	}

	// when
	plans := make([]itemgen.BonusPlan, len(cases))
	for i, tc := range cases {
		plan, err := itemgen.PlanItemBonus(tc.itemType, tc.minlvl, 40, tc.allowspell)
		if err != nil {
			t.Fatalf("%s: PlanItemBonus: %v", tc.name, err)
		}
		plans[i] = plan
	}

	// then
	for i, tc := range cases {
		assertPlan(t, tc, plans[i])
	}
}

func TestPlanItemBonusUnknownType(t *testing.T) {
	// when
	plan, err := itemgen.PlanItemBonus("Excalibur", 1, 5, false)

	// then
	if err == nil {
		t.Fatal("PlanItemBonus: err = nil, want an unknown-type error")
	}
	if plan.Flags != 0 {
		t.Fatalf("plan = %+v, want the zero plan", plan)
	}
}

func TestCalculateToHitBonus(t *testing.T) {
	// given
	cases := []struct {
		name    string
		level   int
		draw    int32
		want    int
		wantErr bool
	}{
		{"negative level returns the negated roll", -25, 0, -1, false},
		{"level 51 rolls 11..15", 51, 4, 15, false},
		{"level 151 rolls 76..100", 151, 24, 100, false},
		{"unknown level is an error", 42, 0, 0, true},
	}

	// when
	got, errs := allToHitBonuses(t, cases)

	// then
	for i, tc := range cases {
		if tc.wantErr != (errs[i] != nil) {
			t.Fatalf("%s: err = %v, wantErr %v", tc.name, errs[i], tc.wantErr)
		}
		if !tc.wantErr && got[i] != tc.want {
			t.Fatalf("%s: CalculateToHitBonus = %d, want %d", tc.name, got[i], tc.want)
		}
	}
}

func TestSaveItemPowerToHitDampDrawsTwice(t *testing.T) {
	// given
	affix := itemgen.Affix{Name: "Sharp", Power: "TOHIT_DAMP", Value1: 20, Value2: 35}
	rnd := &tapeRnd{t: t, values: []int32{3, 8, 2}}

	// when
	bonus, err := itemgen.SaveItemPower(affix, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("SaveItemPower: %v", err)
	}
	if bonus.Value != 28 {
		t.Fatalf("Value = %d, want 28 (second RndPL(20,35) draw = 8)", bonus.Value)
	}
	if bonus.ToHit != 3 {
		t.Fatalf("ToHit = %d, want 3 (RndPL(1,5) = 2 at level 20)", bonus.ToHit)
	}
	assertDraws(t, rnd.args, []int32{16, 16, 5})
}

// lockAffixes is a two-prefix / three-suffix table with one evil prefix and
// one good suffix, so the affix lock is observable.
// TestAffixCoinDenominators pins the three coin draws the slot allocator makes
// before any affix pick: FlipCoin(4) for the prefix, FlipCoin(3) for the suffix
// and FlipCoin(3) for the onlygood upgrade (items.cpp:1181-1191). When both
// coins miss, the retry adds a FlipCoin(2) instead (:1183).
func TestAffixCoinDenominators(t *testing.T) {
	tests := []struct {
		name   string
		values []int32
		want   []int32
	}{
		{name: "no retry", values: []int32{0, 1, 0, 0, 0}, want: []int32{4, 3, 3, 2, 2}},
		{name: "retry coin", values: []int32{1, 0, 0, 0, 0}, want: []int32{4, 3, 2, 3, 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			prefixes, suffixes := lockAffixes()
			rnd := &tapeRnd{t: t, values: tc.values}

			// when
			itemgen.GetItemPowerPrefixAndSuffix(prefixes, suffixes, 1, 6, itemgen.AffixWeapon, false, rnd.rnd)

			// then
			assertDraws(t, rnd.args, tc.want)
		})
	}
}

func lockAffixes() (prefixes, suffixes []itemgen.Affix) {
	prefixes = []itemgen.Affix{
		{
			Name: "Tin", Power: "TOHIT_CURSE", Value1: 6, Value2: 10, MinLevel: 3,
			Types: itemgen.AffixWeapon, Alignment: itemgen.AlignEvil, Chance: 1, MultVal: -3,
		},
		{
			Name: "Bronze", Power: "TOHIT", Value1: 1, Value2: 5, MinLevel: 1,
			Types: itemgen.AffixWeapon, Alignment: itemgen.AlignAny, Chance: 1, Useful: true, MultVal: 2,
		},
	}
	suffixes = []itemgen.Affix{
		{
			Name: "health", Power: "GETHIT", Value1: 1, Value2: 1, MinLevel: 2,
			Types: itemgen.AffixWeapon, Alignment: itemgen.AlignGood, Chance: 1, Useful: true,
		},
		{
			Name: "tears", Power: "GETHIT_CURSE", Value1: 1, Value2: 1, MinLevel: 2,
			Types: itemgen.AffixWeapon, Alignment: itemgen.AlignEvil, Chance: 1,
		},
		{
			Name: "strength", Power: "STR", Value1: 1, Value2: 5, MinLevel: 1,
			Types: itemgen.AffixWeapon, Alignment: itemgen.AlignAny, Chance: 1, Useful: true,
		},
	}

	return prefixes, suffixes
}

func runLoadout(t *testing.T, prefixes, suffixes []itemgen.Affix, values []int32) itemgen.AffixLoadout {
	t.Helper()

	rnd := &tapeRnd{t: t, values: values}

	return itemgen.GetItemPowerPrefixAndSuffix(prefixes, suffixes, 1, 6, itemgen.AffixWeapon, false, rnd.rnd)
}

func saveBoth(
	t *testing.T,
	values *itemgen.AffixValues,
	first, second itemgen.Affix,
	rnd func(int32) int32,
) itemgen.AffixValues {
	t.Helper()

	if _, err := itemgen.SaveItemAffix(values, first, rnd); err != nil {
		t.Fatalf("SaveItemAffix first: %v", err)
	}
	slot1 := *values
	if _, err := itemgen.SaveItemAffix(values, second, rnd); err != nil {
		t.Fatalf("SaveItemAffix second: %v", err)
	}

	return slot1
}

func allToHitBonuses(t *testing.T, cases []struct {
	name    string
	level   int
	draw    int32
	want    int
	wantErr bool
},
) ([]int, []error) {
	t.Helper()

	got := make([]int, len(cases))
	errs := make([]error, len(cases))
	for i, tc := range cases {
		rnd := &tapeRnd{t: t, values: []int32{tc.draw}}
		got[i], errs[i] = itemgen.CalculateToHitBonus(tc.level, rnd.rnd)
	}

	return got, errs
}

func realPrefixes(t *testing.T) []itemgen.Affix {
	t.Helper()

	rows, err := itemdat.LoadItemPrefixes(testsupport.Corpus(t, "items", "assets/txtdata/items/item_prefixes.tsv"))
	if err != nil {
		t.Fatalf("LoadItemPrefixes: %v", err)
	}
	affixes, err := itemgen.PrefixesToAffixes(rows)
	if err != nil {
		t.Fatalf("PrefixesToAffixes: %v", err)
	}

	return affixes
}

func seededRnd(seed uint32) func(int32) int32 {
	engine.SetRndSeed(seed)

	return engine.GenerateRnd
}

func assertNames(t *testing.T, pool []itemgen.Affix, want []string) {
	t.Helper()

	if len(pool) != len(want) {
		t.Fatalf("pool size = %d, want %d", len(pool), len(want))
	}
	for i, name := range want {
		if pool[i].Name != name {
			t.Fatalf("pool[%d] = %q, want %q", i, pool[i].Name, name)
		}
	}
}

func assertAffixesMatch(t *testing.T, pool []itemgen.Affix, typ itemgen.AffixItemType, lo, hi int) {
	t.Helper()

	for _, a := range pool {
		if a.Types&typ == 0 || a.MinLevel < lo || a.MinLevel > hi {
			t.Fatalf("pool contains %q (types %v, level %d), want only %v in %d..%d",
				a.Name, a.Types, a.MinLevel, typ, lo, hi)
		}
	}
}

func assertDraws(t *testing.T, got, want []int32) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("draws = %v, want %v", got, want)
	}
	for i, n := range want {
		if got[i] != n {
			t.Fatalf("draw %d = %d, want %d (draws %v)", i, got[i], n, got)
		}
	}
}

func assertPlan(t *testing.T, tc bonusCase, plan itemgen.BonusPlan) {
	t.Helper()

	if plan.MinLevel != tc.wantMin || plan.MaxLevel != 40 {
		t.Fatalf("%s: level range = %d..%d, want %d..40", tc.name, plan.MinLevel, plan.MaxLevel, tc.wantMin)
	}
	if plan.Flags != tc.wantFlags {
		t.Fatalf("%s: flags = %v, want %v", tc.name, plan.Flags, tc.wantFlags)
	}
	if plan.StaffSpell != tc.wantStaff || plan.NoBonus != tc.wantNoBonus {
		t.Fatalf("%s: plan = %+v, want staffspell=%v nobonus=%v", tc.name, plan, tc.wantStaff, tc.wantNoBonus)
	}
}
