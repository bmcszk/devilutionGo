// Package missiles: the belt and trigger seams the StealPotions and Warp add
// bodies reach.
package missiles

// BeltItem is the belt-slot record AddStealPotions reads (player.h:220).
type BeltItem struct {
	// Type is Item::_itype (tables/itemdat.h:290-305).
	Type int
	// MiscID is Item::_iMiscId (tables/itemdat.h:433-453).
	MiscID int
	// Seed is Item::_iSeed, preserved across the re-init (missiles.cpp:1441-1443).
	Seed int
}

// beltCallbacks is embedded in ProcessCallbacks so AddStealPotions can walk a
// player's belt without those globals (missiles.cpp:1392-1456).
type beltCallbacks struct {
	// PlayerIndexAtPosition ports the PlayerAtPosition lookup when it has to
	// answer "no player" too (missiles.cpp:1398).
	PlayerIndexAtPosition func(tile TilePosition) (int, bool)
	// BeltItem reads one belt slot (player.h:220).
	BeltItem func(player, slot int) BeltItem
	// RemoveBeltItem ports Player::RemoveSpdBarItem (missiles.cpp:1415).
	RemoveBeltItem func(player, slot int)
	// ItemIndexForMisc ports ItemMiscIdIdx (objects.cpp:4387-4395).
	ItemIndexForMisc func(miscID int) int
	// ReinitBeltItem ports InitializeItem with the seed preserved and the stat
	// flag forced (missiles.cpp:1440-1444).
	ReinitBeltItem func(player, slot, itemIndex int)
}

// warpCallbacks is embedded in ProcessCallbacks so AddWarp can read the level
// and trigger tables (missiles.cpp:1507-1580).
type warpCallbacks struct {
	// IsSetLevel reports the setlevel flag (missiles.cpp:1521).
	IsSetLevel func() bool
	// SetLevelNum reports setlvlnum (levels/gendung_defs.hpp:48-60).
	SetLevelNum func() int
	// TriggerCount reports numtrigs (missiles.cpp:1520).
	TriggerCount func() int
	// TriggerMessage reports trigs[i]._tmsg (interfac.h:24-33).
	TriggerMessage func(i int) int
	// TriggerPosition reports trigs[i].position (missiles.cpp:1521).
	TriggerPosition func(i int) TilePosition
	// PosOkPlayer ports PosOkPlayer(player, target) (missiles.cpp:1574).
	PosOkPlayer func(player, x, y int) bool
}
