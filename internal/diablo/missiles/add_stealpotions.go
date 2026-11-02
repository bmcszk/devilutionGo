// Package missiles: the Hellfire trap that spoils a player's belt.
package missiles

import "github.com/bmcszk/devilutionGo/internal/engine"

// Belt and misc-id constants of AddStealPotions (player.h:40,
// tables/itemdat.h:290-453).
const (
	maxBeltItems   = 8
	itemTypeMisc   = 0
	idiNone        = -1
	imiscFullHeal  = 2
	imiscHeal      = 3
	imiscMana      = 6
	imiscFullMana  = 7
	imiscRejuv     = 18
	imiscFullRejuv = 19
	stealSearchMax = 2
)

// AddStealPotions ports AddStealPotions (missiles.cpp:1392-1456): every player
// within 2 tiles of the trap has its belt spoiled — each misc item takes a coin
// flip, potions are removed, full potions are demoted to their lesser form, a
// rejuv becomes a random lesser potion and a full rejuv's replacement rolls a
// d3 — with one "PodPop" sound per player and a redraw; the row then deletes
// itself.
func AddStealPotions(w *HitWorld, m *Missile, _ *AddParameter) {
	start := m.Position.Start

	engine.Crawl(0, stealSearchMax, func(d engine.Displacement) (engine.Displacement, bool) {
		stealBeltAt(w, TilePosition{X: start.X + d.DeltaX, Y: start.Y + d.DeltaY})

		return engine.Displacement{}, false // cpp:1453 — never stop the crawl
	})

	m.DelFlag = true // cpp:1455
}

// stealBeltAt spoils the belt of the player standing on tile, if any
// (missiles.cpp:1396-1454).
func stealBeltAt(w *HitWorld, tile TilePosition) {
	player, ok := w.ProcessCallbacks.PlayerIndexAtPosition(tile)
	if !ok {
		return
	}

	played := false

	for slot := range maxBeltItems {
		if stealBeltSlot(w, player, slot) && !played { // cpp:1446-1449
			w.ProcessCallbacks.PlaySfxLoc("PodPop", tile.X, tile.Y)
			played = true
		}
	}

	w.ProcessCallbacks.RedrawEverything() // cpp:1451
}

// stealBeltSlot spoils one belt slot and reports whether the row counts as a
// changed potion, which is what triggers the sound (missiles.cpp:1403-1450).
func stealBeltSlot(w *HitWorld, player, slot int) bool {
	item := w.ProcessCallbacks.BeltItem(player, slot)

	itemIndex := idiNone
	if item.Type == itemTypeMisc {
		if flipCoin(w) { // cpp:1407-1408
			return false
		}

		var ok bool

		itemIndex, ok = stolenItemIndex(w, player, slot, item.MiscID)
		if !ok {
			return false // cpp:1436-1437 — not a potion, sound not played
		}
	}

	if itemIndex != idiNone { // cpp:1440-1445
		w.ProcessCallbacks.ReinitBeltItem(player, slot, itemIndex)
	}

	return true
}

// stolenItemIndex maps a misc id to its spoiled replacement
// (missiles.cpp:1409-1438); ok is false for the ids that fall through the
// switch without touching the belt.
func stolenItemIndex(w *HitWorld, player, slot, miscID int) (int, bool) {
	switch miscID {
	case imiscFullHeal:
		return w.ProcessCallbacks.ItemIndexForMisc(imiscHeal), true
	case imiscHeal, imiscMana:
		w.ProcessCallbacks.RemoveBeltItem(player, slot)

		return idiNone, true
	case imiscFullMana:
		return w.ProcessCallbacks.ItemIndexForMisc(imiscMana), true
	case imiscRejuv:
		return w.ProcessCallbacks.ItemIndexForMisc(pickRandomlyAmong(w, imiscHeal, imiscMana)), true
	case imiscFullRejuv:
		return w.ProcessCallbacks.ItemIndexForMisc(fullRejuvReplacement(w)), true
	default:
		return idiNone, false
	}
}

// fullRejuvReplacement is the d3 table of AddStealPotions
// (missiles.cpp:1424-1434).
func fullRejuvReplacement(w *HitWorld) int {
	switch w.ProcessCallbacks.GenerateRnd(3) {
	case 0:
		return imiscFullMana
	case 1:
		return imiscFullHeal
	default:
		return imiscRejuv
	}
}

// pickRandomlyAmong ports PickRandomlyAmong over two entries (missiles.cpp:
// 1421): a two-way draw picking the first on zero.
func pickRandomlyAmong(w *HitWorld, first, second int) int {
	if w.ProcessCallbacks.GenerateRnd(2) == 0 {
		return first
	}

	return second
}
