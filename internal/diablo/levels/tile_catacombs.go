package levels

// This file holds the catacombs tile vocabulary and the layer addressing the
// catacombs passes rely on. drlg_l2.cpp uses bare tile ids (it has no `enum
// Tile` of its own, unlike drlg_l1.cpp), so only the ids the generator gives a
// name through repeated use are named here; the rest stay numeric with the
// translating line cited.

// catacombFloor is the catacombs floor tile id, the id the transparency flood
// and the theme-room scan key on (levels/drlg_l2.cpp:2694, 2704).
const catacombFloor Tile = 3

// catacombVWall and catacombHWall are the room wall ids CreateThemeRoom paints
// on a room's left/right and top/bottom edges (levels/gendung.cpp:107-112).
const (
	catacombVWall Tile = 1
	catacombHWall Tile = 2
)

// catacombBlindWall is the tile DrawBlind stamps to close the Halls of the
// Blind's outer wall (levels/drlg_quests.cpp:74-83).
const catacombBlindWall Tile = 154

// The layout alphabet. 'A', 'B', 'C' and 'E' mark the four quadrants of a room
// rectangle so a hall can turn away from the corner it reached
// (levels/drlg_l2.cpp:1708-1713, 1887-1894).
const (
	cellEmpty layoutCell = ' '
	cellRoom  layoutCell = '.'
	cellWall  layoutCell = '#'
	cellDoor  layoutCell = 'D'
	cellHall  layoutCell = ','
	cellQuadA layoutCell = 'A'
	cellQuadB layoutCell = 'B'
	cellQuadC layoutCell = 'C'
	cellQuadE layoutCell = 'E'
)

// layoutCell is the catacombs layout alphabet, the character layer the room
// builder fills before any tile exists (levels/drlg_l2.cpp:58-59).
type layoutCell byte

// flatTile reads the mega-tile layer at the flat address the C array uses, so
// the out-of-range neighbours the catacombs passes read out of habit resolve
// the way they do upstream: `dungeon` is a flat DMAXX*DMAXY array followed
// immediately by `pdungeon` (levels/dun_tile_data.cpp:32-33), so an x of DMAXX
// continues into pdungeon and a y of DMAXY walks into the next row. Reads
// before the layer land in DungeonMask's bit storage, which cannot hold a tile
// id, so they read 0 — the same value the passes' `!= 3` style tests see.
func (d *Dungeon) flatTile(x, y int) Tile {
	index := x*DMAXY + y

	switch {
	case index < 0:
		return 0
	case index < DMAXX*DMAXY:
		return d.Tiles[index/DMAXY][index%DMAXY]
	case index < 2*DMAXX*DMAXY:
		prev := index - DMAXX*DMAXY
		return d.PrevTiles[prev/DMAXY][prev%DMAXY]
	default:
		return 0
	}
}
