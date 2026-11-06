package levels

// This file drives the hive generator: CreateL3Dungeon's GenerateLevel for
// DTYPE_NEST levels 17-20 (levels/drlg_l3.cpp:1991-2147) with the nest variants
// of PlaceStairs (:1982-1989) and PlacePool (:1497-1503), then the nest
// decoration block (:2034-2089). The nest runs no PoolFix, Warp, HallOfHeroes,
// River, theme room or fence pass: upstream keeps those in the caves branch
// (:2090-2143).
//
// Pass3 (:2149-2152) is ExpandPieces in piece.go. PlaceHiveLights (:2169-2178)
// is the CreateL3Dungeon PlaceLights step (:2185-2187) and takes the expanded
// dPiece grid, because GenerateLevel itself never touches dPiece.

// GenerateNest runs the hive generator for the nest level d was created with
// (levels/drlg_l3.cpp:2192-2200, CreateL3Dungeon). The caller seeds d.Rnd;
// like GenerateL3 it only re-seeds when a level seed is already recorded
// (levels/drlg_l3.cpp:1993-1994).
func GenerateNest(d *Dungeon, opts Options) error {
	if d.LevelSeed != 0 {
		d.Rnd.Seed(d.LevelSeed)
	}

	d.InitTrans()

	for {
		done, err := d.rollNest(opts)
		if err != nil {
			return err
		}
		if done {
			break
		}
	}

	d.dumpStage("stairsok")
	d.decorateNest()
	d.dumpStage("theme")
	d.copyToPrevTiles()

	return nil
}

// rollNest is one nest GenerateLevel attempt. It shares the caves layout half
// (levels/drlg_l3.cpp:1996-2019) and swaps only the two family pieces: the
// nest stairs and the slime pool that ends the reroll loop
// (levels/drlg_l3.cpp:2026-2037).
func (d *Dungeon) rollNest(opts Options) (bool, error) {
	d.LevelSeed = d.Rnd.State()
	d.initCaveDungeonFlags()

	d.carveCaveLayout(opts)

	if d.getFloorArea() < caveMinArea || !d.lockout() {
		return false, nil
	}

	d.dumpStage("layout")
	d.makeMegas()
	d.dumpStage("megas")

	return d.placeNestFeatures(opts)
}

// placeNestFeatures places the nest stairs, the anvil quest room and the slime
// pool once the layout is sound; a pool that placed fewer than three patterns
// rejects the attempt (levels/drlg_l3.cpp:2026-2037).
func (d *Dungeon) placeNestFeatures(opts Options) (bool, error) {
	if !d.placeNestStairs(opts) {
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

	return d.placeSlimePool(), nil
}

// decorateNest is the nest half of GenerateLevel's tail, in upstream order:
// the isles, the floor substitutions, the two tunnel pattern sweeps, the two
// hive centre pieces and the surface substitutions
// (levels/drlg_l3.cpp:2034-2089).
func (d *Dungeon) decorateNest() {
	for _, isle := range hiveIsles {
		d.placeCaveMiniSetRandom(isle.miniset, isle.chance)
	}

	for _, sub := range hiveFloorSubs {
		d.placeCaveMiniSetRandom1x1(sub.search, sub.replace, sub.chance)
	}

	for _, pattern := range hiveTunnelRise {
		d.placeCaveMiniSetRandom(pattern.miniset, pattern.chance)
	}

	for _, pattern := range hiveTunnelFall {
		d.placeCaveMiniSetRandom(pattern.miniset, pattern.chance)
	}

	for _, pattern := range hiveCentres {
		d.placeCaveMiniSetRandom(pattern.miniset, pattern.chance)
	}

	for _, sub := range hiveSurfaceSubs {
		d.placeCaveMiniSetRandom1x1(sub.search, sub.replace, sub.chance)
	}
}

// PlaceHiveLights is CreateL3Dungeon's PlaceLights for nest levels
// (levels/drlg_l3.cpp:2169-2178): every world tile of dPiece in the 381-456
// band gets a radius-9 light. addLight is the caller's DoLighting
// (levels/lighting.cpp), which this package does not own.
func PlaceHiveLights(pieces PieceGrid, addLight func(x, y int, radius uint8)) {
	for j := range MAXDUNY {
		for i := range MAXDUNX {
			if pieces.Piece[i][j] < hiveLightFirstTile || pieces.Piece[i][j] > hiveLightLastTile {
				continue
			}

			addLight(i, j, hiveLightRadius)
		}
	}
}
