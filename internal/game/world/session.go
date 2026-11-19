package world

import (
	"encoding/binary"
	"hash/fnv"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// spawnSearchRadius is the FindClosestValidPosition bound of the spawn search
// (player.cpp:3259).
const spawnSearchRadius = 50

// syncInitOffsets is the SyncInitPlrPos offset table (player.cpp:3240).
var syncInitOffsets = [8][2]int{
	{0, 0}, {1, 0}, {0, 1}, {1, 1}, {2, 0}, {0, 2}, {1, 2}, {2, 1},
}

// Place is where a level load put the player, plus the level identity the load
// announces.
type Place struct {
	// Tile is the tile the player stands on after the SyncInitPlrPos search.
	Tile Tile
	// Level and LevelType are the loaded level's identity.
	Level, LevelType int
}

// LevelSource serves the level image a Descend command asks for: the seam the
// diablo layer fills with its committed corpus, so world owns no layout.
type LevelSource interface {
	// LevelImage returns the level's grid image and the tile the player enters
	// on; ok=false when the level cannot be loaded.
	LevelImage(level, levelType int) (*LevelLayout, Tile, bool)
}

// StartingPack is the hero one class's data raises: the base stats and the
// starting life and mana CreatePlayer copies out of the class attributes
// (player.cpp:2287-2330). Life and mana arrive in the 1/64 raw scale upstream
// stores them in (player.h:236-243).
type StartingPack struct {
	// Class is the class ordinal the pack belongs to (classdat.tsv row).
	Class int
	// BaseStr, BaseMag, BaseDex and BaseVit are the class's starting stats.
	BaseStr, BaseMag, BaseDex, BaseVit int
	// HPBase, MaxHPBase, ManaBase and MaxManaBase are the starting vitals.
	HPBase, MaxHPBase, ManaBase, MaxManaBase int
}

// PackSource serves the starting pack a CreatePack command asks for: the seam
// the diablo layer fills from its committed class data, so world owns no table.
type PackSource interface {
	// StartingPack returns the pack for a class ordinal; ok=false when the
	// class has no pack to raise.
	StartingPack(class int) (StartingPack, bool)
}

// LoadLevel replaces the grid with layout and moves the player to the level's
// entry tile — the half of LoadGameLevel (diablo.cpp:3142-3200) this slice owns.
// SyncInitPlrPos then searches the entry tile's window for a tile the player may
// stand on (player.cpp:3235-3271), so a level whose entry is a wall still boots
// onto a legal tile; ok=false leaves the grid loaded and the player in place.
func (s *State) LoadLevel(layout *LevelLayout, spawn Tile, level, levelType int) (Place, bool) {
	if !s.grid.loadImage(layout) {
		return Place{}, false
	}

	tile, ok := s.spawnNear(spawn)
	if !ok {
		return Place{}, false
	}

	s.hero = tile
	s.level = level
	s.kind = levelType
	s.loaded = true
	s.grid.DFlags[tile.Y][tile.X] |= FlagExplored

	return Place{Tile: tile, Level: level, LevelType: levelType}, true
}

// spawnNear is SyncInitPlrPos (player.cpp:3235-3271): the two-by-two window the
// player is dropped into, then FindClosestValidPosition out to 50 tiles
// (player.cpp:3240-3261).
func (s *State) spawnNear(spawn Tile) (Tile, bool) {
	if tile, ok := s.spawnInWindow(spawn); ok {
		return tile, true
	}

	near, ok := engine.FindClosestValidPosition(
		func(p engine.Point) bool { return s.posOkPlayer(Tile{X: p.X, Y: p.Y}) },
		engine.Point{X: spawn.X, Y: spawn.Y}, 1, spawnSearchRadius,
	)

	return Tile{X: near.X, Y: near.Y}, ok
}

// spawnInWindow walks the SyncInitPlrPos offset table in upstream order
// (player.cpp:3240-3247).
func (s *State) spawnInWindow(spawn Tile) (Tile, bool) {
	for _, offset := range syncInitOffsets {
		tile := Tile{X: spawn.X + offset[0], Y: spawn.Y + offset[1]}
		if s.posOkPlayer(tile) {
			return tile, true
		}
	}

	return Tile{}, false
}

// LayoutDigest folds the grid's layout — the piece layer and its SOL
// classification — into one word, so a state hash that folds the digest changes
// whenever the level does. Both layers are read in the layout's own x-major
// order (dun_tile_data.hpp:99), never by map iteration.
func (g *Grid) LayoutDigest() uint64 {
	h := fnv.New64a()

	var buf [4]byte

	for x := range MAXDUNX {
		for y := range MAXDUNY {
			piece := g.DPiece[y][x]

			binary.LittleEndian.PutUint32(buf[:], uint32(piece))
			_, _ = h.Write(buf[:])

			binary.LittleEndian.PutUint32(buf[:], uint32(g.tileProperties(x, y)))
			_, _ = h.Write(buf[:])
		}
	}

	return h.Sum64()
}

// tileProperties is one tile's SOL entry, zero for a piece index the record does
// not cover (the same bound TileHasAny applies).
func (g *Grid) tileProperties(x, y int) TileProperties {
	piece := g.DPiece[y][x]
	if piece < 0 || piece >= len(g.SOLData) {
		return TileNone
	}

	return g.SOLData[piece]
}

// posOkPlayer is this slice's share of PosOkPlayer (player.cpp:3086-3109): the
// tile must be inside the dungeon and walkable. The other-player and monster
// branches of that predicate read registries this slice does not carry yet, so
// it stops at the tile rules instead of inventing occupancy.
func (s *State) posOkPlayer(tile Tile) bool {
	return InDungeonBounds(tile.X, tile.Y) && s.grid.IsTileWalkable(tile.X, tile.Y)
}

// TakeFloorItem removes the item lying on the player's own tile and reports the
// Items index it held — the floor half of InvGetItem, which reads dItem at the
// item's position and clears it through CleanupItems (inv.cpp:1678, :1053-1057).
// A tile with no item changes nothing, so the handler announces nothing.
func (s *State) TakeFloorItem() (int, Tile, bool) {
	tile := s.hero
	if !InDungeonBounds(tile.X, tile.Y) {
		return 0, Tile{}, false
	}

	cell := s.grid.DItem[tile.Y][tile.X]
	if cell == 0 {
		return 0, Tile{}, false
	}

	s.grid.DItem[tile.Y][tile.X] = 0

	return cell - 1, tile, true
}

// RegisterSessionVerbs wires the transitions the walk owns as kernel verbs: a
// Descend loads the level image source serves, a Loot takes the item under the
// player. Each announces its change and stays silent when nothing changed
// (ADR-018 laws 2 and 3).
func RegisterSessionVerbs(k *kernel.Kernel, s *State, source LevelSource) {
	k.Register(kernel.KindDescend, descendHandler(s, source))
	k.Register(kernel.KindLoot, lootHandler(s))
}

// RegisterCreatePack wires the session's hero creation as a kernel verb: the
// pack the class source serves raises the hero on the command's spawn tile and
// the handler announces it (ADR-018 laws 2 and 3).
func RegisterCreatePack(k *kernel.Kernel, s *State, packs PackSource) {
	k.Register(kernel.KindCreatePack, createPackHandler(s, packs))
}

// createPackHandler applies one CreatePack command: A is the class ordinal, B
// and C the spawn tile.
func createPackHandler(s *State, packs PackSource) kernel.Handler {
	return func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		pack, ok := packs.StartingPack(int(c.A))
		if !ok {
			return nil
		}

		tile := Tile{X: int(c.B), Y: int(c.C)}
		if !s.CreateFromPack(pack, tile) {
			return nil
		}

		return []kernel.Event{{Kind: kernel.KindCreatedPack, A: int32(pack.Class), B: int32(tile.X), C: int32(tile.Y)}}
	}
}

// descendHandler applies one Descend command: A is the level number, B the
// dungeon_type the caller expects.
func descendHandler(s *State, source LevelSource) kernel.Handler {
	return func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		layout, spawn, ok := source.LevelImage(int(c.A), int(c.B))
		if !ok {
			return nil
		}

		place, ok := s.LoadLevel(layout, spawn, int(c.A), int(c.B))
		if !ok {
			return nil
		}

		return []kernel.Event{{
			Kind: kernel.KindDescended,
			A:    int32(place.Level),
			B:    int32(place.LevelType),
			C:    int32(place.Tile.Y*MAXDUNX + place.Tile.X),
		}}
	}
}

// lootHandler applies one Loot command to the slice.
func lootHandler(s *State) kernel.Handler {
	return func(kernel.Command, *kernel.Streams) []kernel.Event {
		index, tile, ok := s.TakeFloorItem()
		if !ok {
			return nil
		}

		return []kernel.Event{{Kind: kernel.KindLooted, A: int32(index), B: int32(tile.X), C: int32(tile.Y)}}
	}
}
