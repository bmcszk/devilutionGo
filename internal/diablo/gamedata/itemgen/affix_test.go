package itemgen_test

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
	"github.com/bmcszk/devilutionGo/internal/engine"
)

const (
	probeFNVOffset = 1469598103934665603
	probeFNVPrime  = 1099511628211
	// probeCaseMinimum guards against a truncated dump passing vacuously:
	// 1000 sampled base tuples + 300 Hellfire + 2 worked examples per table set.
	probeCaseMinimum = 1300
)

// probeCase is one committed row of the SelectAffix probe dump
// (internal/testdata/gamedata/items/affix_pool_probe.tsv), produced by
// DevilutionX's own SelectAffix (items.cpp:1063-1093).
type probeCase struct {
	table              string
	typ                itemgen.AffixItemType
	minlvl, maxlvl     int
	onlygood           bool
	goe                itemgen.GoodOrEvil
	excludeCharges     bool
	eligible, poolSize int
	rowsHash           uint64
	rows               []string
}

// probeResult is what the Go pool reports for one probeCase.
type probeResult struct {
	eligible, poolSize int
	rowsHash           uint64
	rows               []string
}

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
	bonus, err := itemgen.SaveItemPower(affix, itemgen.HeroStats{}, rnd.rnd)
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

// TestSaveItemPowerManaLifeConversion pins the two hero-reading powers: 50% of
// _pMaxManaBase moves mana->life and 40% of _pMaxHPBase moves life->mana
// (items.cpp:1019-1028). Everything else in the power table ignores the hero.
func TestSaveItemPowerManaLifeConversion(t *testing.T) {
	// given
	hero := itemgen.HeroStats{MaxManaBase: 120, MaxHPBase: 200}
	cases := []struct {
		name     string
		power    string
		wantMana int
		wantHP   int
	}{
		{name: "mana to life", power: "MANATOLIFE", wantMana: -60, wantHP: 60},
		{name: "life to mana", power: "LIFETOMANA", wantMana: 80, wantHP: -80},
	}
	// when
	bonuses := make([]itemgen.AffixBonus, 0, len(cases))
	for _, tc := range cases {
		bonuses = append(bonuses, manaConversionBonus(t, hero, tc.power))
	}
	// then
	for i, tc := range cases {
		if bonuses[i].PlayerMana != tc.wantMana || bonuses[i].PlayerHP != tc.wantHP {
			t.Fatalf("%s delta = (%d, %d), want (%d, %d)",
				tc.name, bonuses[i].PlayerMana, bonuses[i].PlayerHP, tc.wantMana, tc.wantHP)
		}
	}
}

// TestEligibleAffixesMatchesCppProbe cross-checks EligibleAffixes against the
// C++ probe's pool: size, distinct-row count, ordered-row hash, and the full
// row list for the worked examples.
func TestEligibleAffixesMatchesCppProbe(t *testing.T) {
	t.Parallel()

	// given
	tables := probeTables(t)
	cases := loadAffixProbe(t)

	// when
	got := make([]probeResult, len(cases))
	for i, c := range cases {
		got[i] = measurePool(tables[c.table], c)
	}

	// then
	requireProbeCaseCount(t, len(cases))
	for i, c := range cases {
		assertProbeCase(t, c, got[i])
	}
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

	if _, err := itemgen.SaveItemAffix(values, first, itemgen.HeroStats{}, rnd); err != nil {
		t.Fatalf("SaveItemAffix first: %v", err)
	}
	slot1 := *values
	if _, err := itemgen.SaveItemAffix(values, second, itemgen.HeroStats{}, rnd); err != nil {
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

func manaConversionBonus(t *testing.T, hero itemgen.HeroStats, power string) itemgen.AffixBonus {
	t.Helper()

	bonus, err := itemgen.SaveItemPower(
		itemgen.Affix{Name: power, Power: power, Value1: 3, Value2: 3}, hero, func(int32) int32 { return 0 },
	)
	if err != nil {
		t.Fatalf("SaveItemPower: %v", err)
	}

	return bonus
}

// probeTables loads the affix tables the probe dumped, keyed by its table column.
func probeTables(t *testing.T) map[string][]itemgen.Affix {
	t.Helper()

	return map[string][]itemgen.Affix{
		"prefixes":    realPrefixes(t),
		"suffixes":    realSuffixes(t),
		"hf-prefixes": hfPrefixes(t),
		"hf-suffixes": hfSuffixes(t),
	}
}

// hfPrefixes loads the Hellfire prefix table through the base converter.
func hfPrefixes(t *testing.T) []itemgen.Affix {
	t.Helper()

	return convertHF(t, "item_prefixes.tsv", itemdat.LoadItemPrefixes, itemgen.PrefixesToAffixes)
}

// hfSuffixes loads the Hellfire suffix table through the base converter.
func hfSuffixes(t *testing.T) []itemgen.Affix {
	t.Helper()

	return convertHF(t, "item_suffixes.tsv", itemdat.LoadItemSuffixes, itemgen.SuffixesToAffixes)
}

// convertHF loads one Hellfire affix table and converts it like the base table.
func convertHF[Row any](
	t *testing.T,
	name string,
	load func([]byte) ([]Row, error),
	convert func([]Row) ([]itemgen.Affix, error),
) []itemgen.Affix {
	t.Helper()

	data := testsupport.Corpus(t, "items", "mods/hf/txtdata/items/"+name)
	rows, err := load(data)
	if err != nil {
		t.Fatalf("load hf %s: %v", name, err)
	}
	affixes, err := convert(rows)
	if err != nil {
		t.Fatalf("convert hf %s: %v", name, err)
	}

	return affixes
}

// measurePool reports the Go pool shape for one probe case: the pool size, the
// distinct-row count and the FNV-1a 64 hash of the ordered row list (the same
// hash the probe writes).
func measurePool(affixes []itemgen.Affix, c probeCase) probeResult {
	pool := itemgen.EligibleAffixes(affixes, c.typ, c.minlvl, c.maxlvl, c.onlygood, c.goe, c.excludeCharges)
	out := probeResult{poolSize: len(pool), rows: make([]string, 0, len(pool))}
	hash := uint64(probeFNVOffset)
	prev := ""
	for i, affix := range pool {
		if i == 0 || affix.Name != prev {
			out.eligible++
			prev = affix.Name
		}
		out.rows = append(out.rows, affix.Name)
		hash = hashRowName(hash, affix.Name)
	}
	out.rowsHash = hash

	return out
}

// hashRowName folds one row name and its separator into the running hash.
func hashRowName(hash uint64, name string) uint64 {
	for _, ch := range []byte(name) {
		hash ^= uint64(ch)
		hash *= probeFNVPrime
	}
	hash ^= uint64('\n')
	hash *= probeFNVPrime

	return hash
}

// loadAffixProbe parses the committed probe dump: the DETAIL lines of a case
// precede its summary line.
func loadAffixProbe(t *testing.T) []probeCase {
	t.Helper()

	var (
		out     []probeCase
		pending []string
	)
	for i, fields := range probeRows(t) {
		if isProbeDetail(fields) {
			pending = append(pending, fields[9])
			continue
		}
		out = append(out, parseProbeRow(t, i+1, fields, pending))
		pending = nil
	}
	requireNoPending(t, pending)

	return out
}

// probeRows splits the dump into field slices, validating the column count.
func probeRows(t *testing.T) [][]string {
	t.Helper()

	lines := strings.Split(string(testsupport.Corpus(t, "items", "affix_pool_probe.tsv")), "\n")
	rows := make([][]string, 0, len(lines))
	for i, line := range lines[1:] {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 10 {
			t.Fatalf("probe line %d: %d fields, want 10", i+2, len(fields))
		}
		rows = append(rows, fields)
	}

	return rows
}

// isProbeDetail reports whether the row is one of a case's expanded pool rows.
func isProbeDetail(fields []string) bool {
	return fields[0] == "DETAIL"
}

// requireNoPending fails when the dump ends with rows that have no case.
func requireNoPending(t *testing.T, pending []string) {
	t.Helper()

	if len(pending) != 0 {
		t.Fatalf("probe dump ends with %d unattached DETAIL rows", len(pending))
	}
}

// requireProbeCaseCount fails when the dump is too short to prove anything.
func requireProbeCaseCount(t *testing.T, got int) {
	t.Helper()

	if got < probeCaseMinimum {
		t.Fatalf("probe dump has %d cases, want at least %d (truncated fixture?)", got, probeCaseMinimum)
	}
}

// parseProbeRow turns one summary row plus its DETAIL rows into a probeCase.
func parseProbeRow(t *testing.T, lineNo int, fields, rows []string) probeCase {
	t.Helper()

	return probeCase{
		table:          fields[0],
		typ:            probeType(t, lineNo, fields[1]),
		minlvl:         probeInt(t, lineNo, fields[2]),
		maxlvl:         probeInt(t, lineNo, fields[3]),
		onlygood:       fields[4] == "1",
		goe:            itemgen.GoodOrEvil(probeInt(t, lineNo, fields[5])),
		excludeCharges: fields[6] == "1",
		eligible:       probeInt(t, lineNo, fields[7]),
		poolSize:       probeInt(t, lineNo, fields[8]),
		rowsHash:       probeUint64(t, lineNo, fields[9]),
		rows:           rows,
	}
}

// probeType maps the probe's type column onto the AffixItemType flags.
func probeType(t *testing.T, lineNo int, name string) itemgen.AffixItemType {
	t.Helper()

	byName := map[string]itemgen.AffixItemType{
		"None":   0, // AffixItemType::None
		"Misc":   itemgen.AffixMisc,
		"Bow":    itemgen.AffixBow,
		"Staff":  itemgen.AffixStaff,
		"Weapon": itemgen.AffixWeapon,
		"Shield": itemgen.AffixShield,
		"Armor":  itemgen.AffixArmor,
	}
	typ, ok := byName[name]
	if !ok {
		t.Fatalf("probe line %d: unknown affix type %q", lineNo, name)
	}

	return typ
}

// probeInt parses an int column, failing with the probe line number.
func probeInt(t *testing.T, lineNo int, value string) int {
	t.Helper()

	n, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("probe line %d: int %q: %v", lineNo, value, err)
	}

	return n
}

// probeUint64 parses a hex hash column, failing with the probe line number.
func probeUint64(t *testing.T, lineNo int, value string) uint64 {
	t.Helper()

	n, err := strconv.ParseUint(value, 16, 64)
	if err != nil {
		t.Fatalf("probe line %d: hash %q: %v", lineNo, value, err)
	}

	return n
}

// assertProbeCase compares one case against the C++ probe's pool.
func assertProbeCase(t *testing.T, c probeCase, got probeResult) {
	t.Helper()

	label := fmt.Sprintf("%s type=%d lvl=%d..%d onlygood=%v goe=%d charges=%v",
		c.table, c.typ, c.minlvl, c.maxlvl, c.onlygood, c.goe, c.excludeCharges)
	if got.poolSize != c.poolSize || got.eligible != c.eligible {
		t.Fatalf("%s: Go pool (size %d, distinct %d), probe (size %d, distinct %d)",
			label, got.poolSize, got.eligible, c.poolSize, c.eligible)
	}
	if got.rowsHash != c.rowsHash {
		t.Fatalf("%s: Go rows hash %016x, probe %016x\nGo rows: %v",
			label, got.rowsHash, c.rowsHash, got.rows)
	}
	if c.rows != nil && !slices.Equal(got.rows, c.rows) {
		t.Fatalf("%s: Go rows %v, probe rows %v", label, got.rows, c.rows)
	}
}
