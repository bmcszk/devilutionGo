package levels

// Quests is the quest availability the cathedral generator consults
// (drlg_l1.cpp:386-394, 1116-1145). Each field is a Quest::IsAvailable()
// result for the level being generated; the caller decides it, the generator
// only reads the snapshot.
type Quests struct {
	// PWATER is Quests[Q_PWATER].IsAvailable().
	PWATER bool
	// LTBanner is Quests[Q_LTBANNER].IsAvailable().
	LTBanner bool
	// Butcher is Quests[Q_BUTCHER].IsAvailable().
	Butcher bool
	// Skelking is Quests[Q_SKELKING].IsAvailable() (and not multiplayer).
	Skelking bool
	// CornerStone is CornerStone.isAvailable() (items.cpp:4865-4868).
	CornerStone bool
}

// Options carries the inputs CreateL5Dungeon reads besides the dungeon state.
type Options struct {
	// Entry is the level entry (lvl_entry).
	Entry Entry
	// OriginalCathedral is MyPlayer->pOriginalCathedral.
	OriginalCathedral bool
	// Quests is the quest availability snapshot.
	Quests Quests
	// SetPieces maps a .dun resource path to its bytes, replacing the
	// upstream LoadFileInMem calls.
	SetPieces map[string][]byte
}

// GenerateL1 runs the cathedral generator (CreateL5Dungeon's GenerateLevel).
//
//nolint:revive // d is the generator state; GenerateL1's body lands in the next commit of this lane
func GenerateL1(d *Dungeon, opts Options) error {
	return nil
}
