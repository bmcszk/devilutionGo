package levels

import (
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// TownLayoutSource is the world.LevelSource adapter for the retail town
// layout: it wraps CreateTown and the committed .dun corpus so a world session
// can load the real Tristram grid through the seam.
type TownLayoutSource struct {
	// megaTiles is the town pMegaTiles table (levels/towndata/town.til).
	megaTiles []MegaTile
	// files maps the TownSector*/TownAutomap paths to their .dun bytes.
	files map[string][]byte
	// sol is the town SOL record the caller loaded from the committed
	// town.sol fixture; LevelImage classifies each served piece through it.
	sol []byte
	// seed is the corpus seed the committed town fixtures were generated
	// from, so the served layout matches the fixture byte for byte.
	seed uint32
}

// NewTownLayoutSource constructs the town layout source from the committed
// corpus: the town mega-tile table, the five fixed .dun layers, the SOL record
// (loaded from town.sol and hand-patched by the caller, dun_tile_data.cpp:107)
// and the corpus seed.
func NewTownLayoutSource(megaTiles []MegaTile, files map[string][]byte, sol []byte, seed uint32) *TownLayoutSource {
	return &TownLayoutSource{
		megaTiles: megaTiles,
		files:     files,
		sol:       sol,
		seed:      seed,
	}
}

// LevelImage implements world.LevelSource: it runs CreateTown on a fresh town
// dungeon pinned to the corpus seed and returns the dPiece/dSpecial grids as a
// world LevelLayout, plus the generator's ViewPosition as the spawn tile.
// ok=false for any (level, levelType) other than (0, DTYPE_TOWN).
func (s *TownLayoutSource) LevelImage(level, levelType int) (*world.LevelLayout, world.Tile, bool) {
	if level != 0 || levelType != int(Town) {
		return nil, world.Tile{}, false
	}

	// The adapter owns a fresh dungeon so the generator is pure; the seed is
	// the corpus seed so the output matches the committed fixture.
	d := NewDungeon(Town, 0)
	d.Rnd.Seed(s.seed)
	d.InitLight()

	grid := CreateTown(d, TownOptions{
		Entry:     EntryMain,
		Files:     s.files,
		MegaTiles: s.megaTiles,
		Quests:    Quests{PWATERFountain: true},
	})

	// Convert the internal PieceGrid [MAXDUNX][MAXDUNY]uint16 into the world
	// LevelLayout's x-major flat slice (x*MAXDUNY+y).
	pieces := make([]int, world.MAXDUNX*world.MAXDUNY)
	for x := range MAXDUNX {
		for y := range MAXDUNY {
			pieces[x*MAXDUNY+y] = int(grid.Piece[x][y])
		}
	}

	// The SOL record the caller loaded classifies every piece; LoadLevelSOLData
	// fills the whole MAXTILES array upstream, so pieces beyond the committed
	// record classify as unwritten (zero) — the same default upstream's array
	// zero-init gives. Copy the record up to the number of pieces; remaining
	// entries are zero-initialized (same semantics as upstream).
	sol := make([]byte, len(pieces))
	copy(sol, s.sol)

	return &world.LevelLayout{
		Piece: pieces,
		SOL:   sol,
		Items: nil,
	}, world.Tile{X: int(d.ViewPosition.X), Y: int(d.ViewPosition.Y)}, true
}
