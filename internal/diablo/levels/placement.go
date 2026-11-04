package levels

import "encoding/binary"

// GetDunSize reads the mega-tile dimensions a .dun file starts with
// (levels/gendung.cpp:576-580).
func GetDunSize(dunData []byte) Size {
	return Size{
		Width:  int(binary.LittleEndian.Uint16(dunData[0:])),
		Height: int(binary.LittleEndian.Uint16(dunData[2:])),
	}
}

// PlaceDunTiles copies the .dun tile layer into the dungeon at position,
// protecting every non-zero tile; zero cells become floorID unless it is zero
// (levels/gendung.cpp:503-521).
func (d *Dungeon) PlaceDunTiles(dunData []byte, position Point, floorID Tile) {
	size := GetDunSize(dunData)

	for j := range size.Height {
		for i := range size.Width {
			offset := 4 + 2*(j*size.Width+i)
			tileID := Tile(binary.LittleEndian.Uint16(dunData[offset:]))

			if tileID != 0 {
				d.Tiles[position.X+i][position.Y+j] = tileID
				d.Protected.Set(position.X+i, position.Y+j, true)
			} else if floorID != 0 {
				d.Tiles[position.X+i][position.Y+j] = floorID
			}
		}
	}
}

// placeMiniSetAt decides one candidate position of the PlaceMiniSet scan: it
// wraps at the right/bottom edge, applies the drlg1 x/y<=12 bump, skips
// SetPieceRoom and misfits, and places on success (levels/gendung.cpp:459-501).
//
//nolint:revive // drlg1Quirk is upstream PlaceMiniSet's parameter, not control coupling
func (d *Dungeon) placeMiniSetAt(miniset Miniset, position Point, drlg1Quirk bool) bool {
	sw := miniset.Size.Width
	sh := miniset.Size.Height

	if position.X == DMAXX-sw {
		position.X = 0
		position.Y++
		if position.Y == DMAXY-sh {
			position.Y = 0
		}
	}

	// Limit the position of SetPieces for compatibility with Diablo bug.
	if drlg1Quirk {
		valid := true
		if position.X <= 12 {
			position.X++
			valid = false
		}

		if position.Y <= 12 {
			position.Y++
			valid = false
		}

		if !valid {
			return false
		}
	}

	if d.SetPieceRoom.Contains(position) {
		return false
	}

	if !miniset.Matches(d, position, true) {
		return false
	}

	miniset.Place(d, position, false)

	return true
}

// PlaceMiniSet scans up to tries positions from a random start and places the
// miniset at the first fit (levels/gendung.cpp:459-501). It reports the
// placement position, like the upstream std::optional.
//
//nolint:revive // drlg1Quirk is upstream PlaceMiniSet's parameter, not control coupling
func (d *Dungeon) PlaceMiniSet(miniset Miniset, tries int, drlg1Quirk bool) (Point, bool) {
	sw := miniset.Size.Width
	sh := miniset.Size.Height
	position := Point{
		X: int(d.Rnd.GenerateRnd(int32(DMAXX - sw))),
		Y: int(d.Rnd.GenerateRnd(int32(DMAXY - sh))),
	}

	for range tries {
		if d.placeMiniSetAt(miniset, position, drlg1Quirk) {
			return position, true
		}

		// for (…; i++, position.x++): the bump runs on `continue` too.
		position.X++
	}

	return Point{}, false
}

// placeMiniSetRandomAt applies the random 1x1 replacement at one position
// (levels/drlg_l1.cpp:1263-1267).
func (d *Dungeon) placeMiniSetRandomAt(miniset Miniset, position Point, rndper int32) {
	if !miniset.Matches(d, position, false) {
		return
	}

	// BUGFIX: This code is copied from Cave and should not be applied for crypt
	if !CanReplaceTile(d, miniset.Replace[0][0], position) {
		return
	}

	if d.Rnd.GenerateRnd(100) >= rndper {
		return
	}

	miniset.Place(d, position, false)
}

// PlaceMiniSetRandom scatters random replacements of one tile id over the
// level (levels/drlg_l1.cpp:1251-1268).
func (d *Dungeon) PlaceMiniSetRandom(miniset Miniset, rndper int32) {
	sw := miniset.Size.Width
	sh := miniset.Size.Height

	for sy := range DMAXY - sh {
		for sx := range DMAXX - sw {
			d.placeMiniSetRandomAt(miniset, Point{X: sx, Y: sy}, rndper)
		}
	}
}

// PlaceMiniSetRandom1x1 is the 1x1 helper the crypt substitutions use
// (levels/crypt.cpp:612-615).
func (d *Dungeon) PlaceMiniSetRandom1x1(search, replace Tile, rndper int32) {
	d.PlaceMiniSetRandom(Miniset{
		Size:    Size{Width: 1, Height: 1},
		Search:  [6][6]Tile{{search}},
		Replace: [6][6]Tile{{replace}},
	}, rndper)
}
