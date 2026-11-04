package levels

// This file drives the hell generator: CreateL4Dungeon's GenerateLevel
// (levels/drlg_l4.cpp:1228-1236) for DTYPE_HELL levels 13-16. Only the stage
// skeleton is here; the layout, walls, set pieces, stairs and theme rooms land
// one stage file at a time.

// hellRock is the tile InitDungeonFlags fills the hell layout with
// (levels/drlg_l4.cpp:191, memset(dungeon, 30, sizeof(dungeon))). Hell has no
// Tile enum upstream, so the stages spell their ids inline.
const hellRock Tile = 30

// GenerateL4 runs the hell generator for the level d was created with
// (levels/drlg_l4.cpp:1228-1236, CreateL4Dungeon).
//
// No stage is ported yet, so the dungeon is the hell rock tile the generator
// starts from (levels/drlg_l4.cpp:191, memset(dungeon, 30, sizeof(dungeon))).
// That keeps the run deterministic while the parity gate stays red.
func GenerateL4(d *Dungeon, opts Options) error {
	if d.LevelSeed != 0 {
		d.Rnd.Seed(d.LevelSeed)
	}

	_ = opts

	for x := range DMAXX {
		for y := range DMAXY {
			d.Tiles[x][y] = hellRock
			d.PrevTiles[x][y] = hellRock
		}
	}

	return nil
}
