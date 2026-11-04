package levels

// This file ports the stair placement of the cathedral generator
// (levels/drlg_l1.cpp:1155-1210).

// placeStairs places this level type's stairs, reporting whether the attempt
// produced a usable level (levels/drlg_l1.cpp:1203-1210).
func (d *Dungeon) placeStairs(opts Options) bool {
	if d.LevelType == Crypt {
		return d.placeCryptStairs(opts)
	}

	return d.placeCathedralStairs(opts)
}

// placeCathedralStairs places the poison-water entrance and the up and down
// stairs, retrying the whole level when the caller's original cathedral has no
// stairs (levels/drlg_l1.cpp:1155-1201).
func (d *Dungeon) placeCathedralStairs(opts Options) bool {
	// The entrance is only placed when the quest owns the level; a failed
	// placement only marks the attempt as unusable (drlg_l1.cpp:1162-1166).
	success := !opts.Quests.PWATER || d.placePWATEREntrance(opts)

	if !d.placeStairsUp(opts) {
		if opts.OriginalCathedral {
			return false
		}

		success = false
	}

	if !d.placeStairsDown(opts) {
		success = false
	}

	return success
}

// placePWATEREntrance places the Poisoned Water Supply entrance, opens a
// transparency region over its lower half and records the position the quest
// teleports to (levels/drlg_l1.cpp:1161-1175).
func (d *Dungeon) placePWATEREntrance(opts Options) bool {
	position, ok := d.PlaceMiniSet(pwaterIn, DMAXX*DMAXY, true)
	if !ok {
		return false
	}

	savedVal := d.TransVal
	d.TransVal = 0
	d.MRectTrans(Rectangle{
		Position: position.Add(Point{X: 0, Y: 2}),
		Size:     Size{Width: 5, Height: 2},
	})
	d.TransVal = savedVal

	d.PWATERPosition = position.MegaToWorld().Add(Point{X: 5, Y: 6})
	if opts.Entry == EntryRtnLvl {
		d.ViewPosition = d.PWATERPosition
	}

	return true
}

// stairsUpMiniset is the stairs-up pattern PlaceCathedralStairs asks for: the
// original cathedral's corner stairs when the banner quest does not own the
// level, the plain corner stairs otherwise (levels/drlg_l1.cpp:1178).
func stairsUpMiniset(opts Options) Miniset {
	if opts.OriginalCathedral && !opts.Quests.LTBanner {
		return l5StairSup
	}

	return stairSup
}

// placeStairsUp places the stairs up and points the camera at them, reporting
// whether the level has usable stairs (levels/drlg_l1.cpp:1177-1184).
func (d *Dungeon) placeStairsUp(opts Options) bool {
	position, ok := d.PlaceMiniSet(stairsUpMiniset(opts), DMAXX*DMAXY, true)
	if !ok {
		return false
	}

	if opts.Entry == EntryMain {
		d.ViewPosition = position.MegaToWorld().Add(Point{X: 3, Y: 4})
	}

	return true
}

// placeStairsDown places the stairs down, or, when the banner quest owns the
// level, points the camera at the banner set piece
// (levels/drlg_l1.cpp:1186-1198).
func (d *Dungeon) placeStairsDown(opts Options) bool {
	if opts.Quests.LTBanner {
		if opts.Entry == EntryPrev {
			d.ViewPosition = d.SetPiece.Position.MegaToWorld().Add(Point{X: 3, Y: 11})
		}

		return true
	}

	position, ok := d.PlaceMiniSet(stairDown, DMAXX*DMAXY, true)
	if !ok {
		return false
	}

	if opts.Entry == EntryPrev {
		d.ViewPosition = position.MegaToWorld().Add(Point{X: 3, Y: 3})
	}

	return true
}
