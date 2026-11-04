package levels

import "github.com/bmcszk/devilutionGo/internal/engine"

// Grid dimensions (levels/gendung_defs.hpp:6-9).
const (
	// DMAXX and DMAXY are the mega-tile dimensions of the dungeon layer.
	DMAXX = 40
	DMAXY = 40
	// MAXDUNX and MAXDUNY are the world-tile dimensions of the doubled layers.
	MAXDUNX = 16 + DMAXX*2 + 16
	MAXDUNY = 16 + DMAXY*2 + 16
)

// Dungeon types (levels/gendung_defs.hpp:15-26).
const (
	Town LevelType = iota
	Cathedral
	Catacombs
	Caves
	Hell
	Nest
	Crypt
)

// Level entries (levels/gendung_defs.hpp:28-38).
const (
	EntryMain Entry = iota
	EntryPrev
	EntrySetLvl
	EntryRtnLvl
)

// Tile is a dungeon mega-tile id (upstream `enum Tile : uint8_t`). Each level
// type has its own vocabulary: see the cat* and crypt* constants.
type Tile uint8

// LevelType is dungeon_type (levels/gendung_defs.hpp:15-26).
type LevelType int8

// Entry is lvl_entry (levels/gendung_defs.hpp:28-38). Only the two entries the
// fixture producer drives are modelled; the generator distinguishes them and
// the rest are game-flow entries.
type Entry uint8

// Rnd is the random source a generator draws from (engine/random.hpp). Tests
// script it; EngineRnd wraps the ported LCG.
type Rnd struct {
	GenerateRnd func(v int32) int32
	// RandomIntLessThan is the clamped GenerateRnd (engine/random.hpp:391-394).
	RandomIntLessThan func(v int32) int32
	FlipCoin          func(frequency uint) bool
	Discard           func(count uint)
	Seed              func(seed uint32)
	State             func() uint32
}

// Dungeon is the mutable generator state. Upstream these arrays are globals
// (levels/dun_tile_data.cpp:31-62); one Dungeon value is one set of them.
// Every array keeps the upstream [x][y] indexing.
type Dungeon struct {
	// Tiles is `dungeon`: the mega-tile layer the generator writes.
	Tiles [DMAXX][DMAXY]Tile
	// PrevTiles is `pdungeon`: the copy the automap and set pieces read.
	PrevTiles [DMAXX][DMAXY]Tile
	// Trans is `dTransVal`, the doubled transparency window.
	Trans [MAXDUNX][MAXDUNY]int8
	// TransVal is the current transparency region id (`TransVal`, int8_t).
	TransVal int8
	// TransList is the "region touched this game" flags (`TransList`).
	TransList [256]bool
	// ViewPosition is `ViewPosition`: where the camera starts.
	ViewPosition Point
	// SetPiece and SetPieceRoom are the placed set piece and its exclusion
	// rectangle (`SetPiece`, `SetPieceRoom`).
	SetPiece     Rectangle
	SetPieceRoom Rectangle
	// Protected marks tiles no later pass may overwrite (`Protected`).
	Protected Bitset2d
	// Chamber marks chamber interiors (`Chamber`).
	Chamber Bitset2d
	// Mask marks the carved mega tiles (`DungeonMask`).
	Mask Bitset2d
	// VerticalLayout records whether FirstRoom laid the chambers out
	// vertically (drlg_l1.cpp:26).
	VerticalLayout bool
	// HasChamber1 reports whether chamber 1 is part of the layout
	// (drlg_l1.cpp:28).
	HasChamber1 bool
	// HasChamber2 reports whether chamber 2 is part of the layout
	// (drlg_l1.cpp:30).
	HasChamber2 bool
	// HasChamber3 reports whether chamber 3 is part of the layout
	// (drlg_l1.cpp:32).
	HasChamber3 bool
	// CurrLevel is currlevel and LevelType is leveltype.
	CurrLevel int
	LevelType LevelType
	// Rnd is the random source.
	Rnd Rnd
	// LevelSeed is LevelSeeds[currlevel]: the RNG state the generator re-seeds
	// from when the level is generated again in the same game
	// (levels/drlg_l1.cpp:1214-1215, 1236). Zero means "no recorded seed".
	LevelSeed uint32
	// PWATERPosition is Quests[Q_PWATER].position, the world-tile position of
	// the Poisoned Water Supply entrance (levels/drlg_l1.cpp:1169-1170).
	PWATERPosition Point
	// CheckQuests, when set, is DRLG_CheckQuests: the quest hook the generator
	// calls once the layer is final (levels/drlg_l1.cpp:1294). It writes only
	// pdungeon, never the mega-tile layer.
	CheckQuests func(setPiece Point)
	// DumpStage, when set, is called after each named generator pass with that
	// pass's name. It is the offline stage-oracle hook: the C++ side calls
	// DumpStage() in a probe build only (levels/drlg_l1.cpp:1240, 1247, 1256).
	DumpStage func(name string)
}

// EngineRnd is the LCG-backed source the game itself uses
// (internal/engine/random.go, engine/random.cpp).
func EngineRnd() Rnd {
	return Rnd{
		GenerateRnd:       engine.GenerateRnd,
		RandomIntLessThan: engine.RandomIntLessThan,
		FlipCoin:          engine.FlipCoin,
		Discard:           engine.DiscardRandomValues,
		Seed:              engine.SetRndSeed,
		State:             engine.GetLCGEngineState,
	}
}

// NewDungeon returns a dungeon with empty 40x40 masks and the real LCG source.
func NewDungeon(levelType LevelType, currLevel int) *Dungeon {
	return &Dungeon{
		Protected: NewBitset2d(DMAXX, DMAXY),
		Chamber:   NewBitset2d(DMAXX, DMAXY),
		Mask:      NewBitset2d(DMAXX, DMAXY),
		CurrLevel: currLevel,
		LevelType: levelType,
		Rnd:       EngineRnd(),
	}
}
