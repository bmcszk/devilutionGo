package levels

import "fmt"

// This file ports the caves set pieces and the decoration passes around them:
// the stairs, the anvil of fury room, the town warp and the miscellaneous
// 1x1 substitutions (levels/drlg_l3.cpp:1371-1419, 1801-1888, 1929-1957).

// l3AnvilDun is the anvil of fury room resource PlaceAnvil loads
// (levels/drlg_l3.cpp:1803).
const l3AnvilDun = "levels/l3data/anvil.dun"

// l3StairsTries is PlaceMiniSet's default try count
// (levels/gendung.h:73-75).
const l3StairsTries = 199

// placeAnvilTries bounds PlaceAnvil's scan (levels/drlg_l3.cpp:1814).
const placeAnvilTries = 198

// placeCaveStairs places the up stairs, the down stairs and, on the first caves
// level, the town warp, then records the view position of the entry
// (levels/drlg_l3.cpp:1929-1957).
func (d *Dungeon) placeCaveStairs(opts Options) bool {
	up, ok := d.PlaceMiniSet(l3StairsUp, l3StairsTries, false)
	if !ok {
		return false
	}

	if opts.Entry == EntryMain {
		d.ViewPosition = up.MegaToWorld().Add(Point{X: 1, Y: 3})
	}

	down, ok := d.PlaceMiniSet(l3StairsDown, l3StairsTries, false)
	if !ok {
		return false
	}

	if opts.Entry == EntryPrev {
		d.ViewPosition = down.MegaToWorld().Add(Point{X: 3, Y: 1})
	}

	return d.placeTownWarp(opts.Entry)
}

// placeTownWarp places the stairs back to town on level 9 only
// (levels/drlg_l3.cpp:1948-1956).
func (d *Dungeon) placeTownWarp(entry Entry) bool {
	if d.CurrLevel != 9 {
		return true
	}

	warp, ok := d.PlaceMiniSet(l3TownWarp, l3StairsTries, false)
	if !ok {
		return false
	}

	if entry == EntryTwarPDn {
		d.ViewPosition = warp.MegaToWorld().Add(Point{X: 1, Y: 3})
	}

	return true
}

// warp opens the town warp tiles and drops the stray dirt of the anvil island
// (levels/drlg_l3.cpp:1847-1863).
func (d *Dungeon) warp() {
	for j := range DMAXY {
		for i := range DMAXX {
			if d.isTownWarpCorner(i, j) {
				d.openTownWarp(i, j)

				return
			}

			d.dropDirtBelowFloor(i, j)
		}
	}
}

// dropDirtBelowFloor floors a dirt tile whose south-east neighbour is already
// floor (levels/drlg_l3.cpp:1851-1853).
func (d *Dungeon) dropDirtBelowFloor(i, j int) {
	if d.Tiles[i][j] == 5 && d.flatTile(i+1, j+1) == caveFloor {
		d.Tiles[i][j] = caveFloor
	}
}

// isTownWarpCorner reports whether the four town warp tiles sit at i,j.
func (d *Dungeon) isTownWarpCorner(i, j int) bool {
	return d.Tiles[i][j] == 125 && d.Tiles[i+1][j] == 125 &&
		d.Tiles[i][j+1] == 125 && d.Tiles[i+1][j+1] == 125
}

// openTownWarp writes the four town warp tiles
// (levels/drlg_l3.cpp:1856-1859).
func (d *Dungeon) openTownWarp(i, j int) {
	d.Tiles[i][j] = 156
	d.Tiles[i+1][j] = 155
	d.Tiles[i][j+1] = 153
	d.Tiles[i+1][j+1] = 154
}

// hallOfHeroes opens the hero tiles the theme rooms leave walled in
// (levels/drlg_l3.cpp:1865-1888).
func (d *Dungeon) hallOfHeroes() {
	d.dropOpenHeroFloors()
	d.openHeroCorners()
}

// dropOpenHeroFloors is HallOfHeroes' first scan: every dirt tile with open
// floor to its south-east becomes floor (levels/drlg_l3.cpp:1867-1871).
func (d *Dungeon) dropOpenHeroFloors() {
	for j := range DMAXY {
		for i := range DMAXX {
			d.dropDirtBelowFloor(i, j)
		}
	}
}

// openHeroCorners is HallOfHeroes' second scan over the hero corners
// (levels/drlg_l3.cpp:1873-1886).
func (d *Dungeon) openHeroCorners() {
	for j := range DMAXY {
		for i := range DMAXX {
			d.openHeroCorner(i, j)
		}
	}
}

// openHeroCorner is HallOfHeroes' second scan: a walled hero corner next to the
// double ceiling tile opens into floor (levels/drlg_l3.cpp:1872-1886).
func (d *Dungeon) openHeroCorner(i, j int) {
	if d.Tiles[i][j] != 5 || d.flatTile(i+1, j+1) != 12 {
		return
	}

	switch {
	case d.flatTile(i+1, j) == caveFloor:
		d.Tiles[i][j] = caveFloor
		d.Tiles[i][j+1] = caveFloor
		d.Tiles[i+1][j+1] = caveFloor
	case d.flatTile(i, j+1) == caveFloor:
		d.Tiles[i][j] = caveFloor
		d.Tiles[i+1][j] = caveFloor
		d.Tiles[i+1][j+1] = caveFloor
	default:
	}
}

// placeAnvil drops the anvil of fury room on a patch of open floor and protects
// it (levels/drlg_l3.cpp:1801-1845).
func (d *Dungeon) placeAnvil(opts Options) (bool, error) {
	data, err := dungeonSetPiece(opts, l3AnvilDun)
	if err != nil {
		return false, fmt.Errorf("place anvil: %w", err)
	}

	size := GetDunSize(data)
	area := Size{Width: size.Width + 2, Height: size.Height + 2}
	origin, ok := d.findAnvilSpot(area)
	if !ok {
		return false, nil
	}

	d.PlaceDunTiles(data, origin.Add(Point{X: 1, Y: 1}), caveFloor)
	d.SetPiece = Rectangle{Position: origin, Size: area}
	d.protectArea(d.SetPiece)
	d.closeAnvilIsland()

	return true, nil
}

// findAnvilSpot is PlaceAnvil's placement scan: it starts at a random corner and
// walks the map once looking for open floor (levels/drlg_l3.cpp:1807-1836).
func (d *Dungeon) findAnvilSpot(area Size) (Point, bool) {
	spot := Point{
		X: int(d.Rnd.GenerateRnd(int32(DMAXX - area.Width))),
		Y: int(d.Rnd.GenerateRnd(int32(DMAXY - area.Height))),
	}

	for range placeAnvilTries + 1 {
		spot = d.advanceAnvilSpot(spot, area)

		if d.anvilAreaFree(spot, area) {
			return spot, true
		}

		spot.X++
	}

	return Point{}, false
}

// advanceAnvilSpot wraps the scan to the next row once it runs off the right
// edge, and back to the first row at the bottom edge
// (levels/drlg_l3.cpp:1810-1820).
func (*Dungeon) advanceAnvilSpot(spot Point, area Size) Point {
	if spot.X != DMAXX-area.Width {
		return spot
	}

	spot.X = 0
	spot.Y++

	if spot.Y == DMAXY-area.Height {
		spot.Y = 0
	}

	return spot
}

// anvilAreaFree reports whether every tile of the area is unprotected floor.
func (d *Dungeon) anvilAreaFree(origin Point, area Size) bool {
	for j := range area.Height {
		for i := range area.Width {
			if d.Protected.Test(origin.X+i, origin.Y+j) || d.Tiles[origin.X+i][origin.Y+j] != caveFloor {
				return false
			}
		}
	}

	return true
}

// protectArea marks every tile of area as protected.
func (d *Dungeon) protectArea(area Rectangle) {
	for j := range area.Size.Height {
		for i := range area.Size.Width {
			d.Protected.Set(area.Position.X+i, area.Position.Y+j, true)
		}
	}
}

// closeAnvilIsland walls the three tiles River would otherwise enter the anvil
// island through; Warp opens them again (levels/drlg_l3.cpp:1841-1844).
func (d *Dungeon) closeAnvilIsland() {
	for i := range 3 {
		d.Tiles[d.SetPiece.Position.X+7+i][d.SetPiece.Position.Y+5] = 2
	}
}
