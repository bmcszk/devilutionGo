package monsterspawn

// PlaceGroup's own limits, each the literal of the upstream expression at
// devilutionX monster.cpp (commit e1ad0c2b6).
const (
	// PackRetries is the `try1 < 10` rewind budget of one pack
	// (monster.cpp:312): a failed attempt unplaces its monsters and re-anchors.
	PackRetries = 10
	// ScatterIterations is the `try2 < 100` budget of one attempt's scatter walk
	// (monster.cpp:341).
	ScatterIterations = 100
	// BandOrigin and BandSpan give the anchor band 16..95:
	// GenerateRnd(80) + 16 on both axes (monster.cpp:329-330).
	BandOrigin = 16
	BandSpan   = 80
	// LeashRadius is the `std::abs(xp - x1) >= 4` leash limit
	// (monster.cpp:344): a leashed minion stays within three tiles of the
	// attempt's anchor tile, on each axis.
	LeashRadius = 4
	// MinionHitPointFactor is the `maxHitPoints *= 2` leader-minion doubling
	// (monster.cpp:352).
	MinionHitPointFactor = 2
	// DirectionCount is the GenerateRnd(8) argument of every direction roll
	// (monster.cpp:304, :323, :341).
	DirectionCount = 8
	// SkeletonKingPackSize is the pack size PlaceUniqueMonst requests for the
	// Skeleton King (monster.cpp:541).
	SkeletonKingPackSize = 30
)

// Placement flags AddMonsterType stores per level monster type
// (monster.h:133-138, enum PlaceFlags).
const (
	// PlaceScatter marks a type PlaceGroup scatters (PLACE_SCATTER, 1 << 0).
	PlaceScatter = 1 << iota
	// PlaceSpecial marks a type placed by its own special rule
	// (PLACE_SPECIAL, 1 << 1).
	PlaceSpecial
	// PlaceUnique marks a unique monster's type (PLACE_UNIQUE, 1 << 2).
	PlaceUnique
)
