package levels

// placeSlimePool is PlaceSlimePool, the nest arm of PlacePool
// (levels/drlg_l3.cpp:1497-1503): it stamps the four slime pool patterns and
// reports a usable level only when at least three of them fit
// (levels/drlg_l3.cpp:1421-1437). A pool that placed fewer makes GenerateLevel
// reroll the layout, which is why this returns the flag the decoration call
// sites discard.
func (d *Dungeon) placeSlimePool() bool {
	placed := 0

	for _, pool := range hiveSlimePools {
		if d.placeSlimePattern(pool.miniset, pool.chance) {
			placed++
		}
	}

	return placed >= hiveSlimePoolMinimum
}

// placeSlimePattern is the caves PlaceMiniSetRandom as PlaceSlimePool calls it
// (levels/drlg_l3.cpp:1390-1411): every fitting position takes the pattern once
// the dice agree, and the caller reads whether any position did.
func (d *Dungeon) placeSlimePattern(miniset Miniset, chance int32) bool {
	placed := false

	for sy := range DMAXY - miniset.Size.Height {
		for sx := range DMAXX - miniset.Size.Width {
			if d.placeSlimePatternAt(miniset, Point{X: sx, Y: sy}, chance) {
				placed = true
			}
		}
	}

	return placed
}

// placeSlimePatternAt is one candidate position of that scan
// (levels/drlg_l3.cpp:1396-1409).
func (d *Dungeon) placeSlimePatternAt(miniset Miniset, at Point, chance int32) bool {
	if !miniset.Matches(d, at, true) || !CanReplaceTile(d, miniset.Replace[0][0], at) {
		return false
	}

	if d.Rnd.GenerateRnd(100) >= chance {
		return false
	}

	miniset.Place(d, at, false)

	return true
}
