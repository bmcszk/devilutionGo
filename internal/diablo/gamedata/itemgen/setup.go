package itemgen

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// uniqueMiscMaxSeed is the `iseed > 109` guard of the IMISC_UNIQUE branch
// (items.cpp:3305); 109 is the last unique_itemdat row index the vanilla table
// marks as a misc unique holder.
const uniqueMiscMaxSeed = 109

// BuffHellfire is CF_HELLFIRE in dwBuff (items.h:179).
const BuffHellfire = 1 << 0

// BuffUIDOffsetMask is CF_UIDOFFSET = ((1<<4)-1)<<1 (items.h:180).
const BuffUIDOffsetMask = ((1 << 4) - 1) << 1

// SetupRequest is the SetupAllItems input (items.cpp:3269-3275). It carries no
// unique-dropped flags: upstream filters those in TryRandomUniqueItem, while
// SetupAllItems only rolls CheckUnique against the level-filtered list.
type SetupRequest struct {
	ItemIndex      int
	Seed           uint32
	Level          int
	Uper           int
	OnlyGood       bool
	Pregen         bool
	UidOffset      int
	ForceNotUnique bool
	// Spells is the spell table the staff branch needs. A nil table stands in
	// for GetItemBonus' allowspells=false: staffs then take plain staff
	// affixes instead of spells (items.cpp:1337-1341).
	Spells []txtdata.SpellData
	// Env is the game-mode state the staff spell walk reads.
	Env SpellEnv
	// Hero is the player stat slice the mana/life conversion powers scale
	// (items.cpp:1019-1028); the zero value leaves those powers inert.
	Hero HeroStats
}

// SetupPlan is the SetupAllItems outcome for one item.
type SetupPlan struct {
	CreateInfo int
	Iblvl      int
	Quality    int
	Name       string
	// Cursor is the unique-row graphic override (GetUniqueItem's UICurs); it
	// stays empty when the row leaves ICURS_DEFAULT, meaning keep the base
	// item's graphic (items.cpp:1464-1465).
	Cursor string
	UID    int
	Powers []AffixBonus
	Values AffixValues
	// Spell, Charges and MinMagic are the GetStaffSpell outcome; they stay
	// zero for non-staff items and for staffs on the affix fallback.
	Spell    int
	Charges  int
	MinMagic int
	// Value is _ivalue: the base item value plus the staff spell share
	// (charges*staffCost()/5, items.cpp:1271-1275).
	Value int
	// Ivalue is _iIvalue: the base value, the staff share, the unique table
	// value, or the CalcItemValue result once an affix landed
	// (items.cpp:606-609, 1153, 1226-1227, 1465).
	Ivalue int
	// Buff is the item's dwBuff word: CF_HELLFIRE for Hellfire games
	// (items.cpp:2966-2967) plus the uid-offset bits the caller ORs in
	// (:3412).
	Buff       int
	Durability int
	Cleared    bool // the IMISC_UNIQUE seed/base mismatch that clears the item
}

// UidOffsetBits is the `(uidOffset << 1) & CF_UIDOFFSET` stamp that
// TryRandomUniqueItem adds after regenerating a forced unique
// (items.cpp:3412).
func UidOffsetBits(uidOffset int) int {
	return (uidOffset << 1) & BuffUIDOffsetMask
}

// SetupAllItems ports SetupAllItems' pure decision flow (items.cpp:3269-3312):
// create-info flags, the GetItemBLevel gate, the CheckUnique roll with its
// GetItemBonus / GetUniqueItem split, the unique durability skip, and the
// IMISC_UNIQUE seed-carried uid path.
func SetupAllItems(
	items []itemdat.ItemData,
	uniques []itemdat.UniqueItemData,
	prefixes, suffixes []Affix,
	req SetupRequest,
	rnd func(int32) int32,
	discard func(int),
) (SetupPlan, error) {
	if req.ItemIndex < 0 || req.ItemIndex >= len(items) {
		return SetupPlan{}, fmt.Errorf("item index %d out of range", req.ItemIndex)
	}
	item := items[req.ItemIndex]
	plan := SetupPlan{
		CreateInfo: createInfoFor(req), UID: UniqueNone, Quality: QualityNormal,
		Cursor: item.CursorGraphic, // GetItemAttrs' base _iCurs (:2334)
		Value:  item.Value,         // GetItemAttrs sets both _ivalue and _iIvalue
		Ivalue: item.Value,         // to the base value (:2612-2613)
	}
	if req.Env.Hellfire {
		plan.Buff |= BuffHellfire // GetItemAttrs' gbIsHellfire dwBuff bit
	}

	if item.MiscID == "UNIQUE" {
		return setupUniqueMisc(uniques, item, req, plan, rnd)
	}

	plan.Iblvl = GetItemBLevel(req.Level, item.MiscID, req.OnlyGood, req.Uper == 15, rnd)
	if plan.Iblvl == -1 {
		return plan, nil
	}

	return setupRolledItem(uniques, prefixes, suffixes, item, req, plan, rnd, discard)
}

// createInfoFor packs the caller-provided flags (:3273-3281).
func createInfoFor(req SetupRequest) int {
	info := req.Level
	if req.Pregen {
		info |= CreateInfoPregen
	}
	if req.OnlyGood {
		info |= CreateInfoOnlyGood
	}
	switch req.Uper {
	case 15:
		info |= CreateInfoUper15
	case 1:
		info |= CreateInfoUper1
	default:
		// uper other than 1/15 sets no flag (:3280-3283)
	}

	return info
}

// setupRolledItem is the non-unique-misc body (:3283-3300).
func setupRolledItem(
	uniques []itemdat.UniqueItemData,
	prefixes, suffixes []Affix,
	item itemdat.ItemData,
	req SetupRequest,
	plan SetupPlan,
	rnd func(int32) int32,
	discard func(int),
) (SetupPlan, error) {
	uid, err := rollUnique(uniques, item, req, plan.Iblvl, rnd, discard)
	if err != nil {
		return SetupPlan{}, err
	}
	if uid == UniqueNone {
		return applyBonus(prefixes, suffixes, item, req, plan, rnd)
	}

	return applyUnique(uniques, uid, item, plan, req.Hero, rnd)
}

// uniqueKey is the base-item key upstream compares against
// UniqueItems[uid].UIItemId (items.cpp:3325, 3207): the itemdat
// `uniqueBaseItem` column. Rows without that column (synthetic and legacy
// tables) fall back to the row ID.
func uniqueKey(item itemdat.ItemData) string {
	if item.UniqueBaseItem != "" {
		return item.UniqueBaseItem
	}

	return item.ID
}

// rollUnique is the CheckUnique / forced-discard block (:3285-3293). It filters
// the unique list by iblvl, not the raw level: uper15 already raised the item
// level by 4, so CheckUnique sees the same window GetItemBLevel returned.
func rollUnique(
	uniques []itemdat.UniqueItemData,
	item itemdat.ItemData,
	req SetupRequest,
	iblvl int,
	rnd func(int32) int32,
	discard func(int),
) (int, error) {
	if req.ForceNotUnique {
		discard(1) // CheckUnique

		return UniqueNone, nil
	}

	valid := ValidUniques(uniques, uniqueKey(item), iblvl)
	uid, err := CheckUnique(valid, req.Uper, req.UidOffset, rnd, discard)
	if err != nil {
		return UniqueNone, fmt.Errorf("unique check for %q: %w", item.ID, err)
	}

	return uid, nil
}

// applyBonus is the GetItemBonus branch (:3295).
func applyBonus(
	prefixes, suffixes []Affix,
	item itemdat.ItemData,
	req SetupRequest,
	plan SetupPlan,
	rnd func(int32) int32,
) (SetupPlan, error) {
	types, ok := AffixTypeForItemType(item.ItemType)
	if !ok {
		plan.Durability = ItemRndDur(item.Durability, item.Durability, rnd)

		return plan, nil
	}

	if types == AffixStaff && len(req.Spells) > 0 {
		return applyStaffSpell(prefixes, item, req, plan, rnd)
	}

	return applyAffixPower(prefixes, suffixes, item, req, plan, types, rnd)
}

// applyAffixPower is the GetItemPower call of GetItemBonus (:1320-1342 and the
// staff fallback at :1241).
func applyAffixPower(
	prefixes, suffixes []Affix,
	item itemdat.ItemData,
	req SetupRequest,
	plan SetupPlan,
	types AffixItemType,
	rnd func(int32) int32,
) (SetupPlan, error) {
	// GetItemBonus clamps the lower bound before calling GetItemPower
	// (items.cpp:1311), so a deep item still reaches mid-level affixes.
	power, err := GetItemPower(prefixes, suffixes, ItemPowerRequest{
		BaseName:  item.Name,
		BaseValue: item.Value,
		MinLevel:  min(plan.Iblvl/2, maxAffixMinLevel),
		MaxLevel:  plan.Iblvl,
		Types:     types,
		OnlyGood:  req.OnlyGood,
	}, rnd)
	if err != nil {
		return SetupPlan{}, fmt.Errorf("item power for %q: %w", item.ID, err)
	}
	plan.Values = power.Values
	plan.Name = power.Name
	if power.Magic { // CalcItemValue only runs with an affix (:1226-1227)
		plan.Quality = QualityMagic
		plan.Ivalue = power.Value
	}
	plan.Durability = ItemRndDur(item.Durability, item.Durability, rnd)

	return plan, nil
}

// applyStaffSpell is GetStaffSpell followed by GetStaffPower
// (items.cpp:1237-1280, 1286-1308): assign the spell, its charges, the magic
// requirement and the value share, then roll the staff prefix. The
// GenerateStaffName step stays with the caller because it needs the panel width
// check.
func applyStaffSpell(
	prefixes []Affix,
	item itemdat.ItemData,
	req SetupRequest,
	plan SetupPlan,
	rnd func(int32) int32,
) (SetupPlan, error) {
	roll, err := GetStaffSpell(req.Spells, plan.Iblvl, req.Env, rnd)
	if err != nil {
		return SetupPlan{}, fmt.Errorf("staff spell for %q: %w", item.ID, err)
	}
	if roll.AffixFallback {
		return applyAffixPower(prefixes, nil, item, req, plan, AffixStaff, rnd)
	}

	plan.Spell = roll.Spell
	plan.Charges = roll.Charges
	plan.MinMagic = roll.MinMag
	// The spell share lands on both value slots (:1271-1275).
	plan.Value += roll.Value
	plan.Ivalue += roll.Value

	affix, err := GetStaffPower(prefixes, plan.Iblvl, req.OnlyGood, req.Hero, rnd)
	if err != nil {
		return SetupPlan{}, fmt.Errorf("staff power for %q: %w", item.ID, err)
	}
	if affix.HasPrefix {
		plan.Values = affix.Values
		plan.Quality = QualityMagic
	}
	// GetStaffPower re-prices on every path (:1153).
	plan.Ivalue = CalcItemValue(plan.Values, plan.Ivalue)
	plan.Durability = ItemRndDur(item.Durability, item.Durability, rnd)

	return plan, nil
}

// applyUnique is the GetUniqueItem branch (:3297).
func applyUnique(
	uniques []itemdat.UniqueItemData,
	uid int,
	item itemdat.ItemData,
	plan SetupPlan,
	hero HeroStats,
	rnd func(int32) int32,
) (SetupPlan, error) {
	applied, err := ApplyUniqueItem(uniques[uid], uid, item.MiscID, plan.CreateInfo, hero, rnd)
	if err != nil {
		return SetupPlan{}, fmt.Errorf("unique %d: %w", uid, err)
	}
	plan.Quality = QualityUnique
	plan.Name = applied.Name
	plan.Ivalue = applied.Value // _iIvalue = uniqueItemData.UIValue (:1465)
	if applied.Cursor != "" {   // ICURS_DEFAULT keeps the base graphic (:1464)
		plan.Cursor = applied.Cursor
	}
	plan.UID = applied.UID
	plan.Powers = applied.Powers
	plan.CreateInfo = applied.CreateInfo
	plan.Durability = item.Durability // uniques skip ItemRndDur (:3299-3300)

	return plan, nil
}

// setupUniqueMisc is the IMISC_UNIQUE branch (:3302-3310): the uid rides in
// the seed, and a seed/base mismatch clears the item.
func setupUniqueMisc(
	uniques []itemdat.UniqueItemData,
	item itemdat.ItemData,
	req SetupRequest,
	plan SetupPlan,
	rnd func(int32) int32,
) (SetupPlan, error) {
	if int(req.Seed) > uniqueMiscMaxSeed || int(req.Seed) >= len(uniques) {
		plan.Cleared = true

		return plan, nil
	}
	if uniques[req.Seed].UniqueBaseItem != item.ID {
		plan.Cleared = true

		return plan, nil
	}

	return applyUnique(uniques, int(req.Seed), item, plan, req.Hero, rnd)
}
