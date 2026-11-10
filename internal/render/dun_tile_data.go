package render

// This file ports the dungeon tile data layer: LevelCelBlock and TileProperties
// (Source/levels/dun_tile.hpp:83-119), the SOL record loader with its per-level
// hand patches (Source/levels/dun_tile_data.cpp:99-164) and the MIN block
// decode that fills DPieceMicros (dun_tile_data.cpp:166-217).
//
// SOLData and DPieceMicros are functions' data parameters, not globals: the
// level machinery injects them into the renderer.

import (
	"fmt"
)

const (
	// MaxTiles is MAXTILES, the SOL record bound (dun_tile_data.hpp:29).
	MaxTiles = 1379
)

// The dungeon types a level loads tile data for (dun_tile_data.cpp:71-91).
const (
	LevelTown LevelType = iota
	LevelCathedral
	LevelCatacombs
	LevelCaves
	LevelHell
	LevelNest
	LevelCrypt
)

// The six micro-tile shapes (dun_tile.hpp:20-44).
const (
	TileSquare TileType = iota
	TileTransparentSquare
	TileLeftTriangle
	TileRightTriangle
	TileLeftTrapezoid
	TileRightTrapezoid
)

// The `.sol` flag bits (dun_tile.hpp:107-113).
const (
	TilePropertiesSolid TileProperties = 1 << iota
	TilePropertiesBlockLight
	TilePropertiesBlockMissile
	TilePropertiesTransparent
	TilePropertiesTransparentLeft
	TilePropertiesTransparentRight
	TilePropertiesTrap
)

// solPatches is the patch table written as code, one row per upstream branch
// (MANDATORY-CONTROVERSY minute: data tables do not ride a file of CSV).
var solPatches = map[LevelType][]SOLPatch{
	LevelCathedral: {
		// Fix incorrectly marked arched tiles (dun_tile_data.cpp:109-133).
		{Index: 9, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 15, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 16, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 20, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 21, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 27, Add: TilePropertiesBlockMissile},
		{Index: 28, Add: TilePropertiesBlockMissile},
		{Index: 51, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 56, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 58, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 61, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 63, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 65, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 72, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 208, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 247, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 253, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 257, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 323, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
		{Index: 403, Add: TilePropertiesBlockLight},
		// Fix incorrectly marked pillar tile (dun_tile_data.cpp:130-131).
		{Index: 24, Add: TilePropertiesBlockLight},
		// Fix incorrectly marked wall tile (dun_tile_data.cpp:132-133).
		{Index: 450, Add: TilePropertiesBlockLight | TilePropertiesBlockMissile},
	},
	LevelCaves: {
		// The graphics for tile 48 sub-tile 171 frame 461 are partly
		// incorrect, as they have a few pixels that should belong to the
		// solid tile 49 instead (dun_tile_data.cpp:140-143).
		{Index: 170, Add: TilePropertiesBlockMissile},
		// Fence sub-tiles 481 and 487 are substitutes for solid sub-tiles
		// 473 and 479 but are not marked as solid (:144-147).
		{Index: 481, Add: TilePropertiesSolid},
		{Index: 487, Add: TilePropertiesSolid},
	},
	LevelHell: {
		// Tile is incorrectly marked as being solid (dun_tile_data.cpp:151).
		{Index: 210, Replace: 0, SetClear: true},
	},
	LevelCrypt: {
		// Tile is incorrectly marked as being solid (dun_tile_data.cpp:158).
		{Index: 142, Replace: 0, SetClear: true},
	},
}

// LevelType is dungeon_type; ordinal order stays the upstream one
// (levels/gendung_defs.hpp:15-26), so the wiring can pass leveldat values
// straight through.
type LevelType int8

// TileType is the micro-tile shape and encoding, upstream ordinals kept
// (levels/dun_tile.hpp:19-81). Every encoding lays rows bottom row first.
type TileType uint8

// TileProperties is one `.sol` entry's flags (dun_tile.hpp:106-114).
type TileProperties uint8

// LevelCelBlock is one MIN block of the level CEL file: 12 frame bits plus a
// 3-bit tile type (dun_tile.hpp:83-101).
type LevelCelBlock uint16

// MICROS is one tile's 16 MIN blocks (dun_tile.hpp:119-121).
type MICROS [16]LevelCelBlock

// SOLData is the per-tile properties record of the active level type
// (dun_tile_data.hpp:87, `TileProperties SOLData[MAXTILES]`).
type SOLData [MaxTiles]TileProperties

// SOLPatch is one hand-patched SOL entry, the vendor correction upstream
// applies at load (dun_tile_data.cpp:109-158).
type SOLPatch struct {
	// Index is the 0-based SOLData slot the patch touches.
	Index int
	// Add is the set of properties OR'ed into the entry.
	Add TileProperties
	// Replace and SetClear, together, clear the entry to exactly Replace;
	// upstream uses that for the hell and crypt "incorrectly marked as
	// being solid" fixes (dun_tile_data.cpp:151, :158).
	Replace  TileProperties
	SetClear bool
}

// HasValue reports a used block, the `data != 0` test (dun_tile.hpp:87).
func (b LevelCelBlock) HasValue() bool { return b != 0 }

// Type is the 3-bit shape of the block's frame (dun_tile.hpp:91).
func (b LevelCelBlock) Type() TileType {
	return TileType(uint16(b) & 0x7000 >> 12)
}

// Frame is the 1-based frame index in pDungeonCels (dun_tile.hpp:100).
func (b LevelCelBlock) Frame() uint16 {
	return uint16(b) & 0xFFF
}

// LoadSOLData parses one retail `.sol` record, the byte-for-byte
// LoadIntegralFileInMem (dun_tile_data.cpp:103-158). Records are capped at
// MAXTILES because the upstream destination array has no room over that.
func LoadSOLData(data []byte) (SOLData, error) {
	if len(data) > MaxTiles {
		return SOLData{}, fmt.Errorf("sol record %d bytes exceeds MAXTILES %d", len(data), MaxTiles)
	}

	var values SOLData
	for i, value := range data {
		values[i] = TileProperties(value)
	}

	return values, nil
}

// SOLPatchesFor returns the hand patches of one level type; the returned slice
// shares the table's storage and is valid read-only.
func SOLPatchesFor(level LevelType) []SOLPatch {
	return solPatches[level]
}

// ApplySOLHandPatches overlays one level type's hand patches on the loaded
// record, the per-leveltype branch after each Load call of
// LoadLevelSOLData (dun_tile_data.cpp:107-159).
func ApplySOLHandPatches(values SOLData, level LevelType) SOLData {
	for _, patch := range SOLPatchesFor(level) {
		if patch.SetClear {
			values[patch.Index] = patch.Replace

			continue
		}

		values[patch.Index] |= patch.Add
	}

	return values
}
