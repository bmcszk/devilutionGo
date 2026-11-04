package levels

// This file drives the caves generator: CreateL3Dungeon's GenerateLevel
// (levels/drlg_l3.cpp:1991-2148) for DTYPE_CAVES levels 9-12. The nest branch of
// the same function and Pass3 are not part of this port.

// l3Isles are the isle minisets GenerateLevel scatters after the stairs, with
// the placement percentage of each call (levels/drlg_l3.cpp:2075-2081).
var l3Isles = [...]struct {
	miniset Miniset
	chance  int32
}{
	{miniset: l3Isle1, chance: 70},
	{miniset: l3Isle2, chance: 70},
	{miniset: l3Isle3, chance: 30},
	{miniset: l3Isle4, chance: 30},
	{miniset: l3Isle1, chance: 100},
	{miniset: l3Isle2, chance: 100},
	{miniset: l3Isle5, chance: 90},
}

// l3Decorations are the stalagmite and crevice minisets, in upstream order
// (levels/drlg_l3.cpp:2090-2111).
var l3Decorations = [...]struct {
	miniset Miniset
	chance  int32
}{
	{miniset: l3Tite1, chance: 10},
	{miniset: l3Tite2, chance: 10},
	{miniset: l3Tite3, chance: 10},
	{miniset: l3Tite6, chance: 20},
	{miniset: l3Tite7, chance: 20},
	{miniset: l3Tite8, chance: 20},
	{miniset: l3Tite9, chance: 20},
	{miniset: l3Tite10, chance: 20},
	{miniset: l3Tite11, chance: 30},
	{miniset: l3Tite12, chance: 20},
	{miniset: l3Tite13, chance: 20},
	{miniset: l3Crev1, chance: 30},
	{miniset: l3Crev2, chance: 30},
	{miniset: l3Crev3, chance: 30},
	{miniset: l3Crev4, chance: 30},
	{miniset: l3Crev5, chance: 30},
	{miniset: l3Crev6, chance: 30},
	{miniset: l3Crev7, chance: 30},
	{miniset: l3Crev8, chance: 30},
	{miniset: l3Crev9, chance: 30},
	{miniset: l3Crev10, chance: 30},
	{miniset: l3Crev11, chance: 30},
}

// l3CreviceSubs are the 1x1 substitutions that finish the caves decoration: the
// floor tiles turned into lava rims and wall variants
// (levels/drlg_l3.cpp:2112-2116).
var l3CreviceSubs = [...]struct {
	search  Tile
	replace Tile
	chance  int32
}{
	{search: 7, replace: 106, chance: 25},
	{search: 7, replace: 107, chance: 25},
	{search: 7, replace: 108, chance: 25},
	{search: 9, replace: 109, chance: 25},
	{search: 10, replace: 110, chance: 25},
}

// GenerateL3 runs the caves generator for the level d was created with
// (levels/drlg_l3.cpp:2185-2193, CreateL3Dungeon). The caller seeds d.Rnd;
// GenerateL3 only re-seeds when it finds a level seed recorded by an earlier
// run of the same level (levels/drlg_l3.cpp:1993-1994), which d.LevelSeed
// models.
//
// Pass3 (levels/drlg_l3.cpp:2149-2152) is not part of this port: it writes
// dPiece through the mega-tile asset table, which is asset-pipeline scope.
func GenerateL3(d *Dungeon, opts Options) error {
	if d.LevelSeed != 0 {
		d.Rnd.Seed(d.LevelSeed)
	}

	d.InitTrans()

	for {
		done, err := d.rollCaves(opts)
		if err != nil {
			return err
		}
		if done {
			break
		}
	}

	d.dumpStage("stairsok")
	d.decorateCaves(opts)
	d.dumpStage("theme")
	d.copyToPrevTiles()

	return nil
}

// rollCaves is one GenerateLevel attempt: layout, mega tiles, stairs, the anvil
// room, then the lava pool. A placed pool ends the reroll loop
// (levels/drlg_l3.cpp:1996-2037).
func (d *Dungeon) rollCaves(opts Options) (bool, error) {
	d.LevelSeed = d.Rnd.State()
	d.initCaveDungeonFlags()

	d.carveCaveLayout(opts)

	if d.getFloorArea() < caveMinArea || !d.lockout() {
		return false, nil
	}

	d.dumpStage("layout")
	d.makeMegas()
	d.dumpStage("megas")

	return d.placeCaveFeatures(opts)
}

// placeCaveFeatures places the stairs, the anvil quest room and the lava pool
// once the layout is sound; a failed placement rejects the attempt
// (levels/drlg_l3.cpp:2026-2037).
func (d *Dungeon) placeCaveFeatures(opts Options) (bool, error) {
	if !d.placeCaveStairs(opts) {
		return false, nil
	}

	if opts.Quests.Anvil {
		placed, err := d.placeAnvil(opts)
		if err != nil {
			return false, err
		}
		if !placed {
			return false, nil
		}
	}

	return d.placeLavaPool(), nil
}

// carveAnvilPads floors the square the anvil room needs
// (levels/drlg_l3.cpp:2012-2018).
func (d *Dungeon) carveAnvilPads() {
	x1 := int(d.Rnd.GenerateRnd(10)) + 10
	y1 := int(d.Rnd.GenerateRnd(10)) + 10

	d.floorArea(x1, y1, x1+12, y1+12)
}

// carveCaveLayout is the layout half of one attempt: the first room, its four
// neighbouring blocks, the anvil floor pad and the four fill passes
// (levels/drlg_l3.cpp:1998-2019).
func (d *Dungeon) carveCaveLayout(opts Options) {
	x1 := int(d.Rnd.GenerateRnd(20)) + 10
	y1 := int(d.Rnd.GenerateRnd(20)) + 10
	x2 := x1 + 2
	y2 := y1 + 2

	d.fillRoom(x1, y1, x2, y2)

	d.createBlock(Point{X: x1, Y: y1}, 2, 0)
	d.createBlock(Point{X: x2, Y: y1}, 2, 1)
	d.createBlock(Point{X: x1, Y: y2}, 2, 2)
	d.createBlock(Point{X: x1, Y: y1}, 2, 3)

	if opts.Quests.Anvil {
		d.carveAnvilPads()
	}

	d.fillDiagonals()
	d.fillSingles()
	d.fillStraights()
	d.fillDiagonals()
	d.clearEdges()
}

// decorateCaves runs the caves decoration passes GenerateLevel executes once the
// layout is settled, in upstream order (levels/drlg_l3.cpp:2090-2117).
func (d *Dungeon) decorateCaves(opts Options) {
	d.poolFix()
	d.warp()

	for _, isle := range l3Isles {
		d.placeCaveMiniSetRandom(isle.miniset, isle.chance)
	}

	d.hallOfHeroes()
	d.river()

	if opts.Quests.Anvil {
		d.openAnvilIsland()
	}

	d.placeCaveThemeRooms()
	d.fence()

	for _, decoration := range l3Decorations {
		d.placeCaveMiniSetRandom(decoration.miniset, decoration.chance)
	}

	for _, sub := range l3CreviceSubs {
		d.placeCaveMiniSetRandom1x1(sub.search, sub.replace, sub.chance)
	}
}

// openAnvilIsland undoes closeAnvilIsland and turns the bridge tile River cut
// into the island back into a crossing (levels/drlg_l3.cpp:2103-2109).
func (d *Dungeon) openAnvilIsland() {
	for i := range 3 {
		d.Tiles[d.SetPiece.Position.X+7+i][d.SetPiece.Position.Y+5] = caveFloor
	}

	x, y := d.SetPiece.Position.X+10, d.SetPiece.Position.Y+5
	if d.Tiles[x][y] == 17 || d.Tiles[x][y] == 18 {
		d.Tiles[x][y] = 45
	}
}

// placeCaveMiniSetRandom is the caves PlaceMiniSetRandom: every fitting position
// takes the miniset once the dice agree (levels/drlg_l3.cpp:1390-1411). The
// caves call sites discard its placed flag the way upstream does at 1415 and
// 2029-2031, so the port returns nothing.
func (d *Dungeon) placeCaveMiniSetRandom(miniset Miniset, chance int32) {
	for sy := range DMAXY - miniset.Size.Height {
		for sx := range DMAXX - miniset.Size.Width {
			d.placeCaveMinisetAt(miniset, Point{X: sx, Y: sy}, chance)
		}
	}
}

// placeCaveMinisetAt is one candidate position of the random miniset pass: when
// the miniset fits and the dice agree it is stamped in place
// (levels/drlg_l3.cpp:1396-1409).
//
// The respectProtected=true guard is upstream's default argument
// (levels/gendung.h:45-53; the caves call at levels/drlg_l3.cpp:1401 passes no
// second argument), so protected set-piece tiles reject a candidate here. The
// shared Cathedral pass deliberately passes false instead — that is the
// upstream "Match bug from Crypt levels if false" (levels/drlg_l1.cpp:1258)
// and is not the same call: do not converge the two flags.
func (d *Dungeon) placeCaveMinisetAt(miniset Miniset, at Point, chance int32) {
	if !miniset.Matches(d, at, true) || !CanReplaceTile(d, miniset.Replace[0][0], at) {
		return
	}

	if d.Rnd.GenerateRnd(100) >= chance {
		return
	}

	miniset.Place(d, at, false)
}

// placeCaveMiniSetRandom1x1 is the 1x1 form of the same pass
// (levels/drlg_l3.cpp:1416-1419).
func (d *Dungeon) placeCaveMiniSetRandom1x1(search, replace Tile, chance int32) {
	d.placeCaveMiniSetRandom(Miniset{
		Size:    Size{Width: 1, Height: 1},
		Search:  [6][6]Tile{{search}},
		Replace: [6][6]Tile{{replace}},
	}, chance)
}
