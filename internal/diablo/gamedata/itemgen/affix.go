// Package itemgen ports the item generation rules — affix pools, the
// prefix/suffix lock, value rolls and the monster drop tree — as pure
// functions over the itemdat tables with an injected RNG.
package itemgen

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
)

// AffixItemType flag values, in upstream declaration order (tables/itemdat.h:609-620).
const (
	AffixMisc AffixItemType = 1 << iota
	AffixBow
	AffixStaff
	AffixWeapon
	AffixShield
	AffixArmor
)

// Affix alignment values (GOE_ANY, GOE_EVIL, GOE_GOOD; tables/itemdat.h:603-607).
const (
	AlignAny GoodOrEvil = iota
	AlignEvil
	AlignGood
)

// FlipCoin denominators of GetItemPowerPrefixAndSuffix (items.cpp:1181-1191).
const (
	prefixCoinDenom   = 4
	suffixCoinDenom   = 3
	forceOneCoinDenom = 2
	onlyGoodCoinDenom = 3
)

// maxEligibleAffixes is the StaticVector<const PLStruct*,256> capacity
// (items.cpp:1073). Upstream writes past it; the committed tables stay below.
const maxEligibleAffixes = 256

// chargesPower is IPL_CHARGES (tables/itemdat.h:533).
const chargesPower = "CHARGES"

// AffixItemType mirrors AffixItemType bit flags (tables/itemdat.h:609-620).
type AffixItemType uint8

// GoodOrEvil mirrors goodorevil (tables/itemdat.h:603-607).
type GoodOrEvil uint8

// Affix ports PLStruct (tables/itemdat.h:628-641).
type Affix struct {
	Name      string
	Power     string // item_effect_type name (IPL_*)
	Value1    int    // power.param1
	Value2    int    // power.param2
	MinLevel  int    // PLMinLvl
	Types     AffixItemType
	Alignment GoodOrEvil
	Chance    int // PLChance, the duplicate-count weight
	Useful    bool
	MinVal    int
	MaxVal    int
	MultVal   int
}

// AffixLoadout is the prefix/suffix pair GetItemPowerPrefixAndSuffix reports.
type AffixLoadout struct {
	Prefix *Affix
	Suffix *Affix
}

// affixAllocation is the coin-flip block of GetItemPowerPrefixAndSuffix
// (items.cpp:1181-1193), including the onlygood promotion.
type affixAllocation struct {
	prefix, suffix, onlygood bool
}

// flatAffix is the column set shared by item_prefixes.tsv and item_suffixes.tsv.
type flatAffix struct {
	name, power              string
	value1, value2, minLevel int
	types                    []string
	alignment                string
	chance                   int
	useful                   bool
	minVal, maxVal, multVal  int
}

// PrefixesToAffixes converts item_prefixes.tsv rows into PLStruct ports.
func PrefixesToAffixes(rows []itemdat.ItemPrefix) ([]Affix, error) {
	return convertAffixes(len(rows), func(i int) flatAffix {
		r := rows[i]
		return flatAffix{
			name: r.Name, power: r.Power, value1: r.PowerValue1, value2: r.PowerValue2,
			minLevel: r.MinLevel, types: r.ItemTypes, alignment: r.Alignment, chance: r.Chance,
			useful: r.Useful, minVal: r.MinVal, maxVal: r.MaxVal, multVal: r.MultVal,
		}
	})
}

// SuffixesToAffixes converts item_suffixes.tsv rows into PLStruct ports.
func SuffixesToAffixes(rows []itemdat.ItemSuffix) ([]Affix, error) {
	return convertAffixes(len(rows), func(i int) flatAffix {
		r := rows[i]
		return flatAffix{
			name: r.Name, power: r.Power, value1: r.PowerValue1, value2: r.PowerValue2,
			minLevel: r.MinLevel, types: r.ItemTypes, alignment: r.Alignment, chance: r.Chance,
			useful: r.Useful, minVal: r.MinVal, maxVal: r.MaxVal, multVal: r.MultVal,
		}
	})
}

// SelectAffix ports SelectAffix (items.cpp:1063-1093): filter the table,
// expand each eligible affix by its chance weight, then pick uniformly.
func SelectAffix(
	affixes []Affix,
	typ AffixItemType,
	minlvl, maxlvl int,
	onlygood bool,
	goe GoodOrEvil,
	excludeChargesForStaffs bool,
	rnd func(int32) int32,
) (Affix, bool) {
	pool := EligibleAffixes(affixes, typ, minlvl, maxlvl, onlygood, goe, excludeChargesForStaffs)
	if len(pool) == 0 {
		return Affix{}, false
	}

	// Upstream indexes with GenerateRnd's raw value, a negative one being UB
	// there, so the rare negative is clamped here (DECISIONS in the bean).
	return pool[max(rnd(int32(len(pool))), 0)], true
}

// EligibleAffixes materialises the duplicate-expanded pick pool. Table order
// and per-row chance weights are kept exactly as items.cpp:1085-1087 appends them.
//
//nolint:revive // onlygood is upstream's SelectAffix parameter, not control coupling.
func EligibleAffixes(
	affixes []Affix,
	typ AffixItemType,
	minlvl, maxlvl int,
	onlygood bool,
	goe GoodOrEvil,
	excludeChargesForStaffs bool,
) []Affix {
	pool := make([]Affix, 0, maxEligibleAffixes)
	for _, affix := range affixes {
		if affix.eligible(typ, minlvl, maxlvl, onlygood, goe, excludeChargesForStaffs) {
			pool = appendWeighted(pool, affix)
		}
	}

	return pool
}

// GetItemPowerPrefixAndSuffix ports items.cpp:1179-1207. The drawn prefix's
// alignment becomes the suffix filter (:1195) — the affix lock.
func GetItemPowerPrefixAndSuffix(
	prefixes, suffixes []Affix,
	minlvl, maxlvl int,
	flgs AffixItemType,
	onlygood bool,
	rnd func(int32) int32,
) AffixLoadout {
	alloc := allocateAffixSlots(onlygood, rnd)

	goe := AlignAny
	var out AffixLoadout
	if alloc.prefix {
		if prefix, ok := SelectAffix(prefixes, flgs, minlvl, maxlvl, alloc.onlygood, goe, true, rnd); ok {
			goe = prefix.Alignment
			out.Prefix = &prefix
		}
	}
	if alloc.suffix {
		if suffix, ok := SelectAffix(suffixes, flgs, minlvl, maxlvl, alloc.onlygood, goe, true, rnd); ok {
			out.Suffix = &suffix
		}
	}

	return out
}

// eligible applies the five SelectAffix filters (items.cpp:1074-1082).
func (a Affix) eligible(
	typ AffixItemType,
	minlvl, maxlvl int,
	onlygood bool,
	goe GoodOrEvil,
	excludeChargesForStaffs bool,
) bool {
	return a.Types&typ != 0 &&
		a.MinLevel >= minlvl && a.MinLevel <= maxlvl &&
		(!onlygood || a.Useful) &&
		!a.opposedAlignment(goe) &&
		(!excludeChargesForStaffs || typ != AffixStaff || a.Power != chargesPower)
}

// opposedAlignment reports a good/evil clash with the filter (items.cpp:1079-1080).
func (a Affix) opposedAlignment(goe GoodOrEvil) bool {
	return (goe == AlignGood && a.Alignment == AlignEvil) || (goe == AlignEvil && a.Alignment == AlignGood)
}

// flipCoin ports FlipCoin (random.cpp:105-109): true one time in denom.
func flipCoin(rnd func(int32) int32, denom int) bool {
	return rnd(int32(denom)) == 0
}

// appendWeighted adds one slot per chance point, stopping at the cap where
// upstream overflows its StaticVector (UB, items.cpp:1073).
func appendWeighted(pool []Affix, affix Affix) []Affix {
	for range affix.Chance {
		if len(pool) == maxEligibleAffixes {
			return pool
		}
		pool = append(pool, affix)
	}

	return pool
}

// convertAffixes parses the enum columns of n affix rows into their flags
// (itemTypes/alignment are readEnumList/read at tables/itemdat.cpp:697-698).
func convertAffixes(n int, at func(int) flatAffix) ([]Affix, error) {
	out := make([]Affix, 0, n)
	for i := range n {
		f := at(i)
		types, err := parseAffixItemTypes(f.types)
		if err != nil {
			return nil, fmt.Errorf("affix %q: %w", f.name, err)
		}
		alignment, err := parseAlignment(f.alignment)
		if err != nil {
			return nil, fmt.Errorf("affix %q: %w", f.name, err)
		}
		out = append(out, Affix{
			Name: f.name, Power: f.power, Value1: f.value1, Value2: f.value2,
			MinLevel: f.minLevel, Types: types, Alignment: alignment, Chance: f.chance,
			Useful: f.useful, MinVal: f.minVal, MaxVal: f.maxVal, MultVal: f.multVal,
		})
	}

	return out, nil
}

// parseAffixItemTypes builds the flag set from the comma-split itemTypes cell.
func parseAffixItemTypes(names []string) (AffixItemType, error) {
	var types AffixItemType
	for _, name := range names {
		flag, ok := affixTypeFlag(name)
		if !ok {
			return 0, fmt.Errorf("unknown itemTypes value %q", name)
		}
		types |= flag
	}

	return types, nil
}

// affixTypeFlag maps a ParseAffixItemType name to its flag (tables/itemdat.cpp:547-556).
func affixTypeFlag(name string) (AffixItemType, bool) {
	switch name {
	case "Misc":
		return AffixMisc, true
	case "Bow":
		return AffixBow, true
	case "Staff":
		return AffixStaff, true
	case "Weapon":
		return AffixWeapon, true
	case "Shield":
		return AffixShield, true
	case "Armor":
		return AffixArmor, true
	}

	return 0, false
}

// parseAlignment maps a ParseAffixAlignment name (tables/itemdat.cpp:558-567).
func parseAlignment(name string) (GoodOrEvil, error) {
	switch name {
	case "Any":
		return AlignAny, nil
	case "Evil":
		return AlignEvil, nil
	case "Good":
		return AlignGood, nil
	}

	return 0, fmt.Errorf("unknown alignment %q", name)
}

// allocateAffixSlots decides which affix tables get drawn from.
//
//nolint:revive // onlygood is upstream's GetItemPowerPrefixAndSuffix parameter.
func allocateAffixSlots(onlygood bool, rnd func(int32) int32) affixAllocation {
	prefix := flipCoin(rnd, prefixCoinDenom)
	suffix := !flipCoin(rnd, suffixCoinDenom)
	if !prefix && !suffix {
		// At least try and give each item a prefix or suffix (items.cpp:1183).
		prefix = flipCoin(rnd, forceOneCoinDenom)
		suffix = !prefix
	}
	if !onlygood && !flipCoin(rnd, onlyGoodCoinDenom) {
		onlygood = true
	}

	return affixAllocation{prefix: prefix, suffix: suffix, onlygood: onlygood}
}
