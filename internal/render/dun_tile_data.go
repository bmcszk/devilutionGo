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
	"slices"
	"sort"
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

// CelBlockAdjustment shifts one frame index after the re-encode, the
// (frame, subtract) pair of ComputeCelBlockAdjustments
// (reencode_dun_cels.cpp:305-320).
type CelBlockAdjustment struct {
	Frame    uint16
	Subtract uint16
}

// DecodedPieces is DecodeDungeonPieces' one structure with the per-tile
// micro list, the ascending frame list and microTileLen.
type DecodedPieces struct {
	Micros  DungeonMicros
	Frames  []DunFrame
	TileLen uint8
}

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

// SOLData is the per-tile properties record of the active level type
// (dun_tile_data.hpp:87, `TileProperties SOLData[MAXTILES]`).
type SOLData [MaxTiles]TileProperties

// DunFrame is one entry of the frame list handed to ReencodeDungeonCels
// (dun_tile_data.cpp:197-200).
type DunFrame struct {
	Frame uint16 // 1-based frame index in the source CEL
	Info  DunFrameInfo
}

// DungeonMicros is DPieceMicros: the decoded MIN blocks of one level, one
// entry per tile (dun_tile_data.hpp:107).
type DungeonMicros []MICROS

// DunFrameInfo is what the re-encoder needs to know about one kept frame
// (reencode_dun_cels.hpp:17-27).
type DunFrameInfo struct {
	// MicroTileIndex is the MIN block that referenced the frame first; only
	// blocks 0 and 1 (the floor pair) turn into foliage triangles.
	MicroTileIndex uint8

	Type       TileType       // the MIN block's shape
	Properties TileProperties // the level piece's SOL entry
}

// IsFloor reports walkable-floor frames, the `!Solid && !BlockMissile &&
// (block 0 or 1)` test of stairs (reencode_dun_cels.hpp:23-27).
func (info DunFrameInfo) IsFloor() bool {
	blocked := TilePropertiesSolid | TilePropertiesBlockMissile

	return info.Properties&blocked == 0 && (info.MicroTileIndex == 0 || info.MicroTileIndex == 1)
}

// IsFloorLeft reports the left half of a floor pair (reencode_dun_cels.hpp:28).
func (info DunFrameInfo) IsFloorLeft() bool { return info.MicroTileIndex == 0 }

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

// MinBlockCounts is the (microTileLen, blocks) pair SetDungeonMicros derives
// from the level type (dun_tile_data.cpp:168-177).
func MinBlockCounts(level LevelType) (uint8, int) {
	switch level {
	case LevelTown:
		return 16, 16
	case LevelHell:
		return 12, 16
	default:
		return 10, 10
	}
}

// MinScramble maps a MIN block index onto its on-disk word index: upstream's
// `pieces[blocks - 2 + (block & 1) - (block & 0xE)]` (dun_tile_data.cpp:186).
func MinScramble(blocks, block int) int {
	return blocks - 2 + (block & 1) - (block & 0xE)
}

// LoadMinPieces decodes a retail `.min` record of uint16 LE words, the
// Swap16LE read (dun_tile_data.cpp:186); big-endian hosts go through the BE
// parity job instead of a second read path.
func LoadMinPieces(data []byte) ([]LevelCelBlock, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("min record of %d bytes is not a u16 array", len(data))
	}

	pieces := make([]LevelCelBlock, len(data)/2)
	for i := range pieces {
		pieces[i] = LevelCelBlock(uint16(data[2*i]) | uint16(data[2*i+1])<<8)
	}

	return pieces, nil
}

// DecodeDungeonPieces loads the MIN record of one level: the per-tile micro
// list ordered by the on-disk scramble, microTileLen, and the frame list of
// every referenced frame sorted ascending with its first occurrence's type
// (dun_tile_data.cpp:166-201).
func DecodeDungeonPieces(pieces []LevelCelBlock, level LevelType, sol SOLData) (DecodedPieces, error) {
	tileLen, blocks := MinBlockCounts(level)

	if remainder := len(pieces) % blocks; remainder != 0 {
		return DecodedPieces{}, fmt.Errorf("min record of %d words is not blocks of %d", len(pieces), blocks)
	}

	micros := make(DungeonMicros, len(pieces)/blocks)
	frames := []DunFrame{}

	for piece := range micros {
		decodePieceMicros(pieces, blocks, piece, micros, sol, &frames)
	}

	slices.SortFunc(frames, func(a, b DunFrame) int { return int(a.Frame) - int(b.Frame) })

	return DecodedPieces{Micros: micros, Frames: frames, TileLen: tileLen}, nil
}

// ComputeCelBlockAdjustments shifts the frame indexes of the micro blocks to
// the re-encoded frame table's compaction (reencode_dun_cels.cpp:305-320).
func ComputeCelBlockAdjustments(frames []DunFrame) []CelBlockAdjustment {
	adjustments := []CelBlockAdjustment{}
	lastFrameIndex := uint16(0)
	adjustment := uint16(0)

	for _, frame := range frames {
		diff := frame.Frame - lastFrameIndex - 1 // uint16 wraps like upstream
		if diff > 0 {
			adjustments = append(adjustments, CelBlockAdjustment{Frame: frame.Frame, Subtract: adjustment})
		}

		adjustment += diff
		lastFrameIndex = frame.Frame
	}

	if adjustment > 0 {
		adjustments = append(adjustments, CelBlockAdjustment{Frame: 0xFFFF, Subtract: adjustment})
	}

	return adjustments
}

// RenumberLevelCelBlocks applies the shifts to the decoded micro blocks
// (dun_tile_data.cpp:203-216).
func RenumberLevelCelBlocks(micros DungeonMicros, adjustments []CelBlockAdjustment) {
	for levelPieceId := range micros {
		for block := range micros[levelPieceId] {
			micros[levelPieceId][block] = renumberCelBlock(micros[levelPieceId][block], adjustments)
		}
	}
}

// renumberCelBlock shifts one micro block whose frame lands after an
// adjustment (dun_tile_data.cpp:207-214).
func renumberCelBlock(levelCelBlock LevelCelBlock, adjustments []CelBlockAdjustment) LevelCelBlock {
	frame := levelCelBlock.Frame()

	index, found := sort.Find(len(adjustments), func(i int) int {
		if adjustments[i].Frame > frame {
			return -1
		}

		return 1
	})
	if !found {
		return levelCelBlock
	}

	subtract := adjustments[index].Subtract

	return LevelCelBlock(uint16(levelCelBlock) - subtract)
}

// registerFrame appends the frame when this is its first reference
// (dun_tile_data.cpp:189-194).
func registerFrame(frames []DunFrame, levelCelBlock LevelCelBlock, block int, properties TileProperties) []DunFrame {
	for _, frame := range frames {
		if frame.Frame == levelCelBlock.Frame() {
			return frames
		}
	}

	return append(frames, DunFrame{
		Frame: levelCelBlock.Frame(),
		Info: DunFrameInfo{
			MicroTileIndex: uint8(block),
			Type:           levelCelBlock.Type(),
			Properties:     properties,
		},
	})
}

// decodePieceMicros fills one tile's micro blocks from the on-disk words and
// registers the first use of each referenced frame (dun_tile_data.cpp:184-195).
func decodePieceMicros(
	pieces []LevelCelBlock,
	blocks int,
	levelPieceId int,
	micros DungeonMicros,
	sol SOLData,
	frames *[]DunFrame,
) {
	for block := range blocks {
		levelCelBlock := pieces[blocks*levelPieceId+MinScramble(blocks, block)]
		micros[levelPieceId][block] = levelCelBlock

		if !levelCelBlock.HasValue() {
			continue
		}

		*frames = registerFrame(*frames, levelCelBlock, block, sol[levelPieceId])
	}
}
