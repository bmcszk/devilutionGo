package itemgen

import "errors"

// CreateInfoUseful is CF_USEFUL = CF_UPER15|CF_UPER1 (items.h:172).
const CreateInfoUseful = CreateInfoUper15 | CreateInfoUper1

// useful item ids in the committed itemdat table (tables/itemdat.h:51-62).
const (
	itemHealPotion   = "IDI_HEAL"
	itemManaPotion   = "IDI_MANA"
	itemTownPortal   = "IDI_PORTAL"
	itemOil          = "IDI_OIL"
	hellfireUsefulD7 = 7
	vanillaUsefulD2  = 2
	portalCoinDenom  = 3
)

// UsefulRequest is the SetupAllUseful input (items.cpp:1517-1519). The RNG
// handed to SetupAllUseful must already be seeded with Seed, mirroring the
// SetRndSeed(iseed) call.
type UsefulRequest struct {
	Seed     uint32
	Level    int
	Hellfire bool
}

// UsefulPlan is the SetupAllUseful outcome before GetItemAttrs runs on it.
type UsefulPlan struct {
	Item       string
	CreateInfo int
}

// SetupAllUseful ports the index selection of SetupAllUseful
// (items.cpp:1517-1574): the Hellfire 1d7 table, or the vanilla
// Mana/Heal pick with its level>1 FlipCoin(3) town-portal promotion.
func SetupAllUseful(req UsefulRequest, rnd func(int32) int32) (UsefulPlan, error) {
	item, err := usefulItem(req, rnd)
	if err != nil {
		return UsefulPlan{}, err
	}

	return UsefulPlan{Item: item, CreateInfo: req.Level | CreateInfoUseful}, nil
}

// usefulItem is the `idx` selection switch (:1523-1545).
func usefulItem(req UsefulRequest, rnd func(int32) int32) (string, error) {
	if req.Hellfire {
		return hellfireUsefulItem(req.Level, rnd)
	}

	item := pickAmong(vanillaUsefulD2, []string{itemManaPotion, itemHealPotion}, rnd)
	if req.Level > 1 && rnd(portalCoinDenom) == 0 {
		item = itemTownPortal
	}

	return item, nil
}

// hellfireUsefulItem is the Hellfire 1d7 switch (:1523-1542).
func hellfireUsefulItem(level int, rnd func(int32) int32) (string, error) {
	switch rnd(hellfireUsefulD7) {
	case 0:
		if level <= 1 {
			return itemHealPotion, nil
		}

		return itemTownPortal, nil
	case 1, 2:
		return itemHealPotion, nil
	case 3:
		if level <= 1 {
			return itemManaPotion, nil
		}

		return itemTownPortal, nil
	case 4, 5:
		return itemManaPotion, nil
	case 6:
		return itemOil, nil
	default:
		return "", errors.New("useful roll out of range")
	}
}

// pickAmong is PickRandomlyAmong (engine/random.hpp:377-383), including the
// max(GenerateRnd(n), 0) clamp.
func pickAmong(n int32, values []string, rnd func(int32) int32) string {
	return values[max(rnd(n), 0)]
}
