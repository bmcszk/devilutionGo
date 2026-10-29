package itemgen

import (
	"errors"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
)

// UniqueNone is UITEM_INVALID (tables/itemdat.h:186).
const UniqueNone = -1

// DurIndestructible is DUR_INDESTRUCTIBLE (items.h:35).
const DurIndestructible = 255

// errUniqueOffsetOutOfRange is the uidOffset bounds check (items.cpp:1444-1446).
var errUniqueOffsetOutOfRange = errors.New("uid offset beyond the valid uniques")

// UniqueApplication is the GetUniqueItem outcome (items.cpp:1454-1475). The
// empty Cursor means UICurs == ICURS_DEFAULT, so the base graphic is kept.
type UniqueApplication struct {
	Name       string
	Cursor     string
	Value      int
	UID        int
	Powers     []AffixBonus
	Reseed     bool // IMISC_UNIQUE: _iSeed becomes the uid (:1468)
	CreateInfo int  // createInfo | CF_UNIQUE (:1473)
}

// ValidUniques ports GetValidUniques (items.cpp:1417-1428): the rows for a base
// item whose UIMinLvl the roll level reaches, in table order.
func ValidUniques(uniques []itemdat.UniqueItemData, baseItem string, lvl int) []int {
	valid := make([]int, 0, len(uniques))
	for i, unique := range uniques {
		if unique.UniqueBaseItem == baseItem && lvl >= unique.MinLevel {
			valid = append(valid, i)
		}
	}

	return valid
}

// CheckUnique ports CheckUnique (items.cpp:1430-1452): one uper-percent roll,
// the DiscardRandomValues(1) burn at :1441, then the uidOffset selection from
// the highest eligible unique downwards (:1448-1450).
func CheckUnique(
	validUniques []int,
	uper int,
	uidOffset int,
	rnd func(int32) int32,
	discard func(int),
) (int, error) {
	if rnd(100) > int32(uper) {
		return UniqueNone, nil
	}
	if len(validUniques) == 0 {
		return UniqueNone, nil
	}

	discard(1)

	if uidOffset < 0 || uidOffset >= len(validUniques) {
		return UniqueNone, errUniqueOffsetOutOfRange
	}

	return validUniques[len(validUniques)-1-uidOffset], nil
}

// GetItemBLevel ports items.cpp:1481-1493.
//
// successful first roll draws nothing further; uper15 overrides any iblvl.
//
//nolint:revive // onlygood/uper15 are upstream's parameters, not control coupling. The || chain short-circuits, so a
func GetItemBLevel(lvl int, miscID string, onlygood, uper15 bool, rnd func(int32) int32) int {
	iblvl := -1
	if rnd(100) <= 10 || rnd(100) <= int32(lvl) || onlygood || isJewelryMisc(miscID) {
		iblvl = lvl
	}
	if uper15 {
		iblvl = lvl + 4
	}

	return iblvl
}

// ItemRndDur ports ItemRndDur (items.cpp:1475-1478): durability jitter over
// [maxDur/4+1, maxDur*3/4] for anything not indestructible.
func ItemRndDur(durability, maxDur int, rnd func(int32) int32) int {
	if durability <= 0 || durability == DurIndestructible {
		return durability
	}

	return int(rnd(int32(maxDur/2))) + maxDur/4 + 1
}

// ApplyUniqueItem ports GetUniqueItem (items.cpp:1454-1475): apply the power
// list in order until the first invalid slot, then take the unique name, value
// and graphic, and stamp the unique identity onto the item.
func ApplyUniqueItem(
	unique itemdat.UniqueItemData,
	uid int,
	miscID string,
	createInfo int,
	rnd func(int32) int32,
) (UniqueApplication, error) {
	out := UniqueApplication{
		Name:       unique.Name,
		Cursor:     unique.CursorGraphic,
		Value:      unique.Value,
		UID:        uid,
		Reseed:     miscID == "UNIQUE",
		CreateInfo: createInfo | CreateInfoUnique,
	}
	for i, power := range uniquePowers(unique) {
		bonus, err := SaveItemPower(power, rnd)
		if err != nil {
			return UniqueApplication{}, fmt.Errorf("unique %q power %d: %w", unique.Name, i, err)
		}
		out.Powers = append(out.Powers, bonus)
	}

	return out, nil
}

// uniquePowers expands the six power slots, stopping at the first empty one
// (IPL_INVALID, items.cpp:1459-1461).
func uniquePowers(unique itemdat.UniqueItemData) []Affix {
	slots := [...]Affix{
		{Power: unique.Power0, Value1: unique.Power0Value1, Value2: unique.Power0Value2},
		{Power: unique.Power1, Value1: unique.Power1Value1, Value2: unique.Power1Value2},
		{Power: unique.Power2, Value1: unique.Power2Value1, Value2: unique.Power2Value2},
		{Power: unique.Power3, Value1: unique.Power3Value1, Value2: unique.Power3Value2},
		{Power: unique.Power4, Value1: unique.Power4Value1, Value2: unique.Power4Value2},
		{Power: unique.Power5, Value1: unique.Power5Value1, Value2: unique.Power5Value2},
	}
	powers := make([]Affix, 0, len(slots))
	for _, slot := range slots {
		if slot.Power == "" {
			break
		}
		powers = append(powers, slot)
	}

	return powers
}

// isJewelryMisc covers the IMISC_STAFF/RING/AMULET term of the chain (:1486).
func isJewelryMisc(miscID string) bool {
	return miscID == "STAFF" || miscID == "RING" || miscID == "AMULET"
}
