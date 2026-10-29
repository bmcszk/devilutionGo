package itemgen

// Create-info flags (items.h:158-171 icreateinfo_flag).
const (
	CreateInfoLevel        = (1 << 6) - 1
	CreateInfoOnlyGood     = 1 << 6
	CreateInfoUper15       = 1 << 7
	CreateInfoUper1        = 1 << 8
	CreateInfoUnique       = 1 << 9
	CreateInfoSmith        = 1 << 10
	CreateInfoSmithPremium = 1 << 11
	CreateInfoBoy          = 1 << 12
	CreateInfoWitch        = 1 << 13
	CreateInfoHealer       = 1 << 14
	CreateInfoPregen       = 1 << 15
)

// maxAffixMinLevel is the min(minlvl, 25) cap SetupAllItems applies.
const maxAffixMinLevel = 25

// witchSpellChance is Adria's `GenerateRnd(100) <= 5` gate (:2271).
const witchSpellChance = 5

// LevelWindow is the minlvl/maxlvl pair GetItemPowerPrefixAndSuffix receives.
type LevelWindow struct {
	MinLevel int
	MaxLevel int
	Iblvl    int
}

// ItemLevelWindow ports the minlvl/maxlvl decision tree of SetupAllItems
// (items.cpp:2249-2279): the vendor branches, the witch 5% staff upgrade, the
// default GetItemBLevel path, and the min(minlvl, 25) cap at :2279.
func ItemLevelWindow(createInfo int, miscID string, rnd func(int32) int32) LevelWindow {
	lvl := createInfo & CreateInfoLevel
	onlygood := createInfo&(CreateInfoOnlyGood|CreateInfoSmithPremium|CreateInfoBoy|CreateInfoWitch) != 0

	var window LevelWindow
	switch {
	case createInfo&CreateInfoSmithPremium != 0:
		window = LevelWindow{MinLevel: lvl / 2, MaxLevel: lvl, Iblvl: lvl}
	case createInfo&CreateInfoBoy != 0:
		window = LevelWindow{MinLevel: lvl, MaxLevel: lvl * 2, Iblvl: lvl * 2}
	case createInfo&CreateInfoWitch != 0:
		window = witchWindow(lvl, miscID, rnd)
	default:
		iblvl := GetItemBLevel(lvl, miscID, onlygood, createInfo&CreateInfoUper15 != 0, rnd)
		window = LevelWindow{MinLevel: iblvl / 2, MaxLevel: iblvl, Iblvl: iblvl}
	}

	window.MinLevel = min(window.MinLevel, maxAffixMinLevel)

	return window
}

// AffixTypeForItemType ports the affix item-type switch of SetupAllItems
// (items.cpp:2283-2332); ok is false for the no-affix item types (None, Misc,
// Gold and anything the switch does not name, which upstream leaves as None).
func AffixTypeForItemType(itemType string) (AffixItemType, bool) {
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
	default:
		return 0, false // None, Misc, Gold and unnamed types
	}
}

// witchWindow is Adria's branch: a 5% 2*lvl upgrade, also granted to staffs
// (:2268-2276).
func witchWindow(lvl int, miscID string, rnd func(int32) int32) LevelWindow {
	iblvl := -1
	if rnd(100) <= witchSpellChance {
		iblvl = 2 * lvl
	}
	if iblvl == -1 && miscID == "STAFF" {
		iblvl = 2 * lvl
	}

	return LevelWindow{MinLevel: iblvl / 2, MaxLevel: iblvl, Iblvl: iblvl}
}
