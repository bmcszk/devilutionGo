package levels

// The fixed .dun resources CreateTown loads, by upstream path
// (levels/town.cpp:202-207). The caller supplies their bytes, replacing the
// upstream LoadFileInMem calls.
const (
	// TownSector1 is the 25x25 north-east quadrant, placed at (46,46).
	TownSector1 = `levels\towndata\sector1s.dun`
	// TownSector2 is the 25x23 north-west quadrant, placed at (46,0).
	TownSector2 = `levels\towndata\sector2s.dun`
	// TownSector3 is the 23x25 south-east quadrant, placed at (0,46).
	TownSector3 = `levels\towndata\sector3s.dun`
	// TownSector4 is the 23x23 south-west quadrant, placed at (0,0).
	TownSector4 = `levels\towndata\sector4s.dun`
	// TownAutomap is the 40x40 automap tile layer (`dungeon`).
	TownAutomap = `levels\towndata\automap.dun`
)

// TownOptions is the input of CreateTown: the resources it loads by path, the
// level's mega-tile table and the warp/quest snapshot its closed-warp patches
// read. The RNG is d.Rnd, which the caller seeds (levels/town.cpp:224-226 is the
// only draw in the level, so the seed is an explicit input).
type TownOptions struct {
	// Entry is lvl_entry; it only moves ViewPosition (levels/town.cpp:364-384).
	Entry Entry
	// Files maps the TownSector*/TownAutomap paths to their `.dun` bytes.
	Files map[string][]byte
	// MegaTiles is the town pMegaTiles table (levels/towndata/town.til), which
	// FillSector indexes by tile id minus one (levels/town.cpp:39-44).
	MegaTiles []MegaTile
	// Warps is IsWarpOpen(dungeon_type); a missing key is a closed warp, so
	// every patch runs (levels/town.cpp:210-238).
	Warps map[LevelType]bool
	// Hellfire is gbIsHellfire: it adds the hive and grave patches
	// (levels/town.cpp:228-238).
	Hellfire bool
	// Quests is the quest snapshot; Quests.PWATER selects the live poisoned
	// water fountain (levels/town.cpp:240-244).
	Quests Quests
}

// CreateTown is the town generator (levels/town.cpp:359-387, DrlgTPass3): it
// writes the `dungeon` tile layer into d and returns the dPiece/dSpecial grids
// the same pass builds.
//
// WIP: not ported yet. This shim returns the zeroed grids so TestTownParity has
// a target to diff against the committed goldens; the parity failure it
// produces is the RED state the port is written against.
func CreateTown(d *Dungeon, opts TownOptions) PieceGrid {
	_ = d
	_ = opts

	return PieceGrid{}
}
