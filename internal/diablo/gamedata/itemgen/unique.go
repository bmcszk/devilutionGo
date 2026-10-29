package itemgen

import (
	"errors"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
)

// UniqueNone is UITEM_INVALID (tables/itemdat.h:186).
const UniqueNone = -1

// DurIndestructible is DUR_INDESTRUCTIBLE (items.h:35).
const DurIndestructible = 255

// errUniqueOffsetOutOfRange is the uidOffset bounds check (items.cpp:1444-1446).
var errUniqueOffsetOutOfRange = errors.New("uid offset beyond the valid uniques")

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

// isJewelryMisc covers the IMISC_STAFF/RING/AMULET term of the chain (:1486).
func isJewelryMisc(miscID string) bool {
	return miscID == "STAFF" || miscID == "RING" || miscID == "AMULET"
}
