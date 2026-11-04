package levels

// This file drives the catacombs generator: CreateL2Dungeon's GenerateLevel
// (levels/drlg_l2.cpp:2679-2829) for DTYPE_CATACOMBS, levels 5-8.

// GenerateL2 runs the catacombs generator for the level d was created with.
// The caller seeds d.Rnd; GenerateLevel only re-seeds when it finds a level
// seed recorded by an earlier run of the same level
// (levels/drlg_l2.cpp:2681-2682, `LevelSeeds[currlevel]`), which d.LevelSeed
// models.
//
// Pass3 (levels/drlg_l2.cpp:2821-2828) and InitDungeonPieces
// (levels/drlg_l2.cpp:1657-1684) are not part of this port: they write
// dPiece/dSpecial through the mega-tile asset table, which is asset-pipeline
// scope, not generator scope (recorded on devilGo-4ybb).
func GenerateL2(d *Dungeon, opts Options) error {
	if d.LevelSeed != 0 {
		d.Rnd.Seed(d.LevelSeed)
	}

	for {
		done, err := d.rollCatacombs(opts)
		if err != nil {
			return err
		}
		if done {
			break
		}
	}

	d.fixCatacombLockout()
	d.fixCatacombDoors()
	d.fixCatacombDirtTiles()

	d.decorateCatacombs()

	d.catacombSubstitution()
	d.applyCatacombShadowPatterns()

	d.copyToPrevTiles()

	return d.drawCatacombQuests(opts)
}

// rollCatacombs is one GenerateLevel attempt: layout, tile patterns, set
// piece, transparency then stairs; GenerateLevel rerolls until the stairs fit
// (levels/drlg_l2.cpp:2685-2726).
func (d *Dungeon) rollCatacombs(opts Options) (bool, error) {
	d.LevelSeed = d.Rnd.State()
	d.InitTrans()

	if !d.createCatacombs(opts.Quests) {
		return false, nil
	}

	d.fixCatacombTilesPatterns()

	if err := d.placeCatacombSetPiece(opts); err != nil {
		return false, err
	}

	d.FloodTransparencyValues(catacombFloor)
	d.fixCatacombTransparency()

	return d.placeCatacombStairs(opts.Entry), nil
}

// decorateCatacombs runs the decoration passes GenerateLevel executes once the
// layout is settled, in upstream order (levels/drlg_l2.cpp:2704-2812).
func (d *Dungeon) decorateCatacombs() {
	d.placeCatacombThemeRooms()

	for _, door := range l2CtrDoor {
		d.placeCatacombMiniSetRandom(door, 100)
	}

	// VARCH33-40 come before VARCH1-32 in upstream call order.
	for _, arch := range l2VArch[32:] {
		d.placeCatacombMiniSetRandom(arch, 100)
	}

	for _, arch := range l2VArch[:32] {
		d.placeCatacombMiniSetRandom(arch, 100)
	}

	for _, arch := range l2HArch {
		d.placeCatacombMiniSetRandom(arch, 100)
	}

	d.placeCatacombMiniSetRandom(l2CrushColumn, 99)
	d.scatterCatacombWallVariants()
}

// scatterCatacombWallVariants places the 1x1 wall variant replacements, the
// two stitched rooms and the ten big-room minisets
// (levels/drlg_l2.cpp:2794-2812).
func (d *Dungeon) scatterCatacombWallVariants() {
	walls := []struct {
		search, replace Tile
		rndPer          int32
	}{
		{1, 80, 10},
		{1, 81, 10},
		{1, 82, 10},
		{2, 84, 10},
		{2, 85, 10},
		{2, 86, 10},
		{8, 87, 50},
	}

	for _, wall := range walls {
		d.placeCatacombMiniSetRandom1x1(wall.search, wall.replace, wall.rndPer)
	}

	d.placeCatacombMiniSetRandom(l2Pancreas1, 1)
	d.placeCatacombMiniSetRandom(l2Pancreas2, 1)

	bigRooms := []struct {
		miniset Miniset
		rndPer  int32
	}{
		{l2Big[0], 3},
		{l2Big[1], 3},
		{l2Big[2], 3},
		{l2Big[3], 3},
		{l2Big[4], 3},
		{l2Big[5], 20},
		{l2Big[6], 20},
		{l2Big[7], 3},
		{l2Big[8], 20},
		{l2Big[9], 20},
	}

	for _, room := range bigRooms {
		d.placeCatacombMiniSetRandom(room.miniset, room.rndPer)
	}
}
