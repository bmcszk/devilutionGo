package levels

// placeNestStairs is PlaceNestStairs: the stairs up (the town warp on the first
// nest level) and, on every nest level but the last, the stairs down, each with
// the view position of the entry (levels/drlg_l3.cpp:1959-1980).
func (d *Dungeon) placeNestStairs(opts Options) bool {
	up, ok := d.PlaceMiniSet(hiveUpStairs(d.CurrLevel), placeMiniSetTries, false)
	if !ok {
		return false
	}

	if opts.Entry == EntryMain || opts.Entry == EntryTwarPDn {
		d.ViewPosition = up.MegaToWorld().Add(Point{X: 1, Y: 3})
	}

	if d.CurrLevel == hiveLowestLevel {
		return true
	}

	down, ok := d.PlaceMiniSet(hiveStairsDown, placeMiniSetTries, false)
	if !ok {
		return false
	}

	if opts.Entry == EntryPrev {
		d.ViewPosition = down.MegaToWorld().Add(Point{X: 3, Y: 1})
	}

	return true
}

// hiveUpStairs is the stairs-up pattern PlaceNestStairs asks for: the nest
// entrance on the first nest level, the plain stairs above it
// (levels/drlg_l3.cpp:1964).
func hiveUpStairs(level int) Miniset {
	if level == hiveFirstLevel {
		return hiveTownWarp
	}

	return hiveStairsUp
}
