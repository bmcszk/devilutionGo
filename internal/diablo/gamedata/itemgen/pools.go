package itemgen

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
)

// Quality tiers (items.h:42-46).
const (
	QualityNormal = 0
	QualityMagic  = 1
	QualityUnique = 2
)

// rndAllGoldThreshold is the RndAllItems gold gate: >25 means gold.
const rndAllGoldThreshold = 25

// allItemsCursor is RndAllItems' level multiplier (:1394).
const allItemsCursor = 2

// UniqueRoll is the TryRandomUniqueItem outcome (items.cpp:3315-3360).
type UniqueRoll struct {
	UID        int  // UniqueItems index, or UniqueNone when no roll happened
	Rolled     bool // a candidate was picked
	Downgraded bool // every valid unique was already dropped: uper 1, level +4
	Uper       int
	Level      int
}

// RndAllItems ports items.cpp:1392-1401: 26% of the time a level-filtered
// unweighted item, otherwise gold.
func RndAllItems(items []itemdat.ItemData, currLevel int, env DropEnv, rnd func(int32) int32) (int, error) {
	if rnd(100) > rndAllGoldThreshold {
		return itemGoldRow, nil
	}

	idx, err := GetItemIndexForDroppableItem(items, env, false, func(item itemdat.ItemData) bool {
		return currLevel*allItemsCursor >= item.MinMonsterLevel
	}, rnd)
	if err != nil {
		return ItemNone, fmt.Errorf("all-items pool at level %d: %w", currLevel, err)
	}

	return idx, nil
}

// RndTypeItems ports items.cpp:1403-1413: a level-, type- and misc-id filtered
// unweighted pick; miscID "" means imid == -1 (any).
func RndTypeItems(
	items []itemdat.ItemData,
	itemType, miscID string,
	lvl int,
	env DropEnv,
	rnd func(int32) int32,
) (int, error) {
	idx, err := GetItemIndexForDroppableItem(items, env, false, func(item itemdat.ItemData) bool {
		if lvl*allItemsCursor < item.MinMonsterLevel {
			return false
		}
		if item.ItemType != itemType {
			return false
		}

		return miscID == "" || item.MiscID == miscID
	}, rnd)
	if err != nil {
		return ItemNone, fmt.Errorf("type %q misc %q pool at level %d: %w", itemType, miscID, lvl, err)
	}

	return idx, nil
}

// TryRandomUniqueItem ports TryRandomUniqueItem's pure part
// (items.cpp:3315-3360): keep only undropped candidates (multiplayer keeps all,
// :3334), pick with max(0, GenerateRnd(len)) (:3351), and when none remain
// drop to uper 1 with the +4 level adjustment for uper 15 (:3340-3345).
func TryRandomUniqueItem(
	validUniques []int,
	dropped []bool,
	multiplayer bool,
	uper, mLevel int,
	rnd func(int32) int32,
) UniqueRoll {
	uids := undroppedUniques(validUniques, dropped, multiplayer)
	if len(uids) == 0 {
		return UniqueRoll{UID: UniqueNone, Downgraded: true, Uper: 1, Level: downgradeLevel(uper, mLevel)}
	}

	return UniqueRoll{
		UID:    uids[max(rnd(int32(len(uids))), 0)],
		Rolled: true,
		Uper:   uper,
		Level:  mLevel,
	}
}

// undroppedUniques is the UniqueItemFlags filter (:3328-3336).
//
//nolint:revive // multiplayer is upstream's gbIsMultiplayer gate, not control coupling.
func undroppedUniques(validUniques []int, dropped []bool, multiplayer bool) []int {
	uids := make([]int, 0, len(validUniques))
	for _, uid := range validUniques {
		if uid < 0 || uid >= len(dropped) || !dropped[uid] || multiplayer {
			uids = append(uids, uid)
		}
	}

	return uids
}

// downgradeLevel is the `if (uper == 15) mLevel += 4` adjustment (:3341-3343).
func downgradeLevel(uper, mLevel int) int {
	if uper == 15 {
		return mLevel + 4
	}

	return mLevel
}
