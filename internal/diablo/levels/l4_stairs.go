package levels

// This file ports the hell stairs pass: PlaceStairs places the stairs up first,
// then the level-dependent exit, and points the camera at the entrance the
// player arrived by (levels/drlg_l4.cpp:1101-1142).

// placeHellStairs is PlaceStairs. It reports false when a miniset does not fit,
// which makes GenerateLevel roll the whole layout again.
func (d *Dungeon) placeHellStairs(opts Options) bool {
	position, ok := d.PlaceMiniSet(l4UStairs, hellPlacementTries, false)
	if !ok {
		return false
	}

	if opts.Entry == EntryMain {
		d.ViewPosition = position.MegaToWorld().Add(Point{X: 6, Y: 6})
	}

	if d.CurrLevel == hellPentagramLevel {
		return d.placeHellGate(opts)
	}

	return d.placeHellExits(opts)
}

// placeHellGate places the pentagram portal every level 15 attempt needs
// (levels/drlg_l4.cpp:1124-1131).
func (d *Dungeon) placeHellGate(opts Options) bool {
	penta, ok := d.PlaceMiniSet(l4Penta2, hellPlacementTries, false)
	if !ok {
		return false
	}

	d.DiabloPosition = penta

	if opts.Entry == EntryPrev {
		d.ViewPosition = penta.MegaToWorld().Add(Point{X: 6, Y: 5})
	}

	return true
}

// placeHellExits places the level's exits: the stairs down unless level 16 or the
// warlord room stands in for them, and the town warp on level 13
// (levels/drlg_l4.cpp:1108-1123).
func (d *Dungeon) placeHellExits(opts Options) bool {
	if d.CurrLevel != hellQuadLevel && !d.placeHellStairsDown(opts) {
		return false
	}

	if d.CurrLevel != warlordLevel {
		return true
	}

	warp, ok := d.PlaceMiniSet(l4TWarp, hellPlacementTries, false)
	if !ok {
		return false
	}

	if opts.Entry == EntryTwarPDn {
		d.ViewPosition = warp.MegaToWorld().Add(Point{X: 6, Y: 6})
	}

	return true
}

// placeHellStairsDown places the stairs down, or points the camera at the
// warlord room that replaced them (levels/drlg_l4.cpp:1110-1122).
func (d *Dungeon) placeHellStairsDown(opts Options) bool {
	if opts.Quests.Warlord {
		if opts.Entry == EntryPrev {
			d.ViewPosition = d.SetPiece.Position.MegaToWorld().Add(Point{X: 7, Y: 7})
		}

		return true
	}

	down, ok := d.PlaceMiniSet(l4DStairs, hellPlacementTries, false)
	if !ok {
		return false
	}

	if opts.Entry == EntryPrev {
		d.ViewPosition = down.MegaToWorld().Add(Point{X: 7, Y: 5})
	}

	return true
}
