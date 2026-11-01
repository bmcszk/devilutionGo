package persist

import "fmt"

// Level-grid constants (Source/levels/gendung_defs.hpp:5-9,
// Source/levels/dun_tile_data.hpp:34-45).
const (
	// DungeonWidth is MAXDUNX.
	DungeonWidth = 112
	// DungeonHeight is MAXDUNY.
	DungeonHeight = 112
	// AutomapWidth is DMAXX.
	AutomapWidth = 40
	// AutomapHeight is DMAXY.
	AutomapHeight = 40
	// DungeonFlagSaved is DungeonFlag::SavedFlags: Populated|Lit|Explored.
	DungeonFlagSaved uint8 = 0xC8
	// DungeonFlagLoaded is DungeonFlag::LoadedFlags.
	DungeonFlagLoaded uint8 = 0xCF
	// dungeonGridCells is MAXDUNX*MAXDUNY.
	dungeonGridCells = DungeonWidth * DungeonHeight
	// automapGridCells is DMAXX*DMAXY.
	automapGridCells = AutomapWidth * AutomapHeight
	// TileGridsSize is the dMonster+dObject+dLight+dPreLight+Automap block
	// SaveLevel writes (loadsave.cpp:1986-2008).
	TileGridsSize = dungeonGridCells*(4+1+1+1) + automapGridCells
	// CorpseGridSize is the dCorpse block SaveLevel writes first
	// (loadsave.cpp:1948-1953).
	CorpseGridSize = dungeonGridCells
)

// LevelHeaderSize is the three BE counts SaveLevel writes after the corpse
// grid (loadsave.cpp:1955-1957): ActiveMonsterCount, ActiveItemCount and
// ActiveObjectCount.
const LevelHeaderSize = 12

// Level entry kinds (Source/loadsave.cpp:1051-1057).
const (
	// LevelRegular is 'l': the current level number.
	LevelRegular LevelKind = 'l'
	// LevelSet is 's': the set-level map number.
	LevelSet LevelKind = 's'
)

// LevelKind selects the level entry suffix (Source/loadsave.cpp:1051-1057).
type LevelKind byte

// LevelHeader is the counter block of a level stream
// (Source/loadsave.cpp:1955-1957).
type LevelHeader struct {
	// ActiveMonsters is ActiveMonsterCount.
	ActiveMonsters int32
	// ActiveItems is ActiveItemCount.
	ActiveItems int32
	// ActiveObjects is ActiveObjectCount.
	ActiveObjects int32
}

// TileGrids is the contiguous grid block of SaveLevel: the big-endian dMonster
// indices followed by the byte grids dObject, dLight, dPreLight and
// AutomapView (Source/loadsave.cpp:1986-2008).
type TileGrids struct {
	// Monsters is dMonster (BE int32 on disk).
	Monsters [dungeonGridCells]int32
	// Objects is dObject.
	Objects [dungeonGridCells]int8
	// Lights is dLight.
	Lights [dungeonGridCells]uint8
	// PreLights is dPreLight.
	PreLights [dungeonGridCells]uint8
	// Automap is AutomapView.
	Automap [automapGridCells]uint8
}

// Encode returns the grid block image.
func (g TileGrids) Encode() []byte {
	var out recordWriter

	for _, monster := range g.Monsters {
		out.be32(uint32(monster))
	}

	for _, object := range g.Objects {
		out.u8(uint8(object))
	}

	out.bytes(g.Lights[:])
	out.bytes(g.PreLights[:])
	out.bytes(g.Automap[:])

	return out.buf
}

// DecodeTileGrids parses the grid block image, which must be TileGridsSize.
func DecodeTileGrids(buf []byte) (TileGrids, error) {
	in := recordReader{buf: buf}

	var grids TileGrids

	for i := range grids.Monsters {
		grids.Monsters[i] = int32(in.be32())
	}

	for i := range grids.Objects {
		grids.Objects[i] = int8(in.u8())
	}

	copy(grids.Lights[:], in.bytes(dungeonGridCells))
	copy(grids.PreLights[:], in.bytes(dungeonGridCells))
	copy(grids.Automap[:], in.bytes(automapGridCells))

	if in.err != nil {
		return TileGrids{}, in.err
	}

	return grids, nil
}

// SavedFlags masks a dFlags cell to the bits SaveLevel stores
// (loadsave.cpp:1978-1984, DungeonFlag::SavedFlags).
func SavedFlags(cell uint8) uint8 {
	return cell & DungeonFlagSaved
}

// EncodeCorpseGrid returns the dCorpse block of a non-town level
// (loadsave.cpp:1948-1953).
func EncodeCorpseGrid(cells [dungeonGridCells]int8) []byte {
	out := make([]byte, 0, CorpseGridSize)

	for _, cell := range cells {
		out = append(out, uint8(cell))
	}

	return out
}

// EncodeFlagGrid returns the masked dFlags block (loadsave.cpp:1978-1984).
func EncodeFlagGrid(cells [dungeonGridCells]uint8) []byte {
	out := make([]byte, 0, dungeonGridCells)

	for _, cell := range cells {
		out = append(out, SavedFlags(cell))
	}

	return out
}

// Encode returns the counter block; every field is big-endian.
func (h LevelHeader) Encode() []byte {
	var out recordWriter

	out.be32(uint32(h.ActiveMonsters))
	out.be32(uint32(h.ActiveItems))
	out.be32(uint32(h.ActiveObjects))

	return out.buf
}

// DecodeLevelHeader parses the counter block.
func DecodeLevelHeader(buf []byte) (LevelHeader, error) {
	in := recordReader{buf: buf}

	header := LevelHeader{
		ActiveMonsters: int32(in.be32()),
		ActiveItems:    int32(in.be32()),
		ActiveObjects:  int32(in.be32()),
	}

	if in.err != nil {
		return LevelHeader{}, in.err
	}

	return header, nil
}

// LevelEntryName is GetLevelNames (Source/loadsave.cpp:1049-1061): the prefix
// plus the kind suffix, then the level number zero-padded to two digits.
func LevelEntryName(prefix string, kind LevelKind, levelNum int) (string, error) {
	if levelNum < 0 || levelNum > 99 {
		return "", fmt.Errorf("persist: level %d out of range 0..99: %w", levelNum, ErrSlotOutOfRange)
	}

	return fmt.Sprintf("%s%c%02d", prefix, byte(kind), levelNum), nil
}

// TempLevelEntryName is GetTempLevelNames (loadsave.cpp:1063-1066).
func TempLevelEntryName(kind LevelKind, levelNum int) (string, error) {
	return LevelEntryName("temp", kind, levelNum)
}

// PermLevelEntryName is GetPermLevelNames (loadsave.cpp:1068-1071).
func PermLevelEntryName(kind LevelKind, levelNum int) (string, error) {
	return LevelEntryName("perm", kind, levelNum)
}
