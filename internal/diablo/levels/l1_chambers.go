package levels

import (
	"errors"
	"fmt"
)

// This file ports the chamber decoration stage of the cathedral generator: the
// three chambers and the halls joining them, the hard-coded pillar pattern, and
// the quest set piece dropped into one of the chambers
// (levels/drlg_l1.cpp:804-874, 1036-1075, 1330-1356).

// ErrMissingSetPiece reports a set piece the caller did not supply.
var ErrMissingSetPiece = errors.New("missing set piece data")

// generateChamber decorates one 10x10 chamber: floor, the four pillars, and the
// arch/door connection in each direction the layout joins this chamber
// (levels/drlg_l1.cpp:804-855).
//
//nolint:revive // connectPrevious/connectNext/verticalLayout are upstream GenerateChamber's parameters
func (d *Dungeon) generateChamber(position Point, connectPrevious, connectNext, verticalLayout bool) {
	if connectPrevious {
		if verticalLayout {
			d.Tiles[position.X+2][position.Y] = catHArch
			d.Tiles[position.X+3][position.Y] = catHArch
			d.Tiles[position.X+4][position.Y] = catCorner
			d.Tiles[position.X+7][position.Y] = catVArchEnd
			d.Tiles[position.X+8][position.Y] = catHArch
			d.Tiles[position.X+9][position.Y] = catHWall
		} else {
			d.Tiles[position.X][position.Y+2] = catVArch
			d.Tiles[position.X][position.Y+3] = catVArch
			d.Tiles[position.X][position.Y+4] = catCorner
			d.Tiles[position.X][position.Y+7] = catHArchEnd
			d.Tiles[position.X][position.Y+8] = catVArch
			d.Tiles[position.X][position.Y+9] = catVWall
		}
	}

	if connectNext {
		d.connectNextChamber(position, verticalLayout)
	}

	for y := 1; y < 11; y++ {
		for x := 1; x < 11; x++ {
			d.Tiles[position.X+x][position.Y+y] = catFloor
			d.Chamber.Set(position.X+x, position.Y+y, true)
		}
	}

	d.Tiles[position.X+4][position.Y+4] = catPillar
	d.Tiles[position.X+7][position.Y+4] = catPillar
	d.Tiles[position.X+4][position.Y+7] = catPillar
	d.Tiles[position.X+7][position.Y+7] = catPillar
}

// connectNextChamber draws the arch connection on the far side of a chamber:
// the run leaves the chamber towards the next one, or towards the dirt when the
// chamber is the last one (levels/drlg_l1.cpp:819-855).
//
//nolint:revive // verticalLayout is upstream GenerateChamber's parameter
func (d *Dungeon) connectNextChamber(position Point, verticalLayout bool) {
	if verticalLayout {
		position.Y += 11
		d.Tiles[position.X+2][position.Y] = catHArchVWall
		d.Tiles[position.X+3][position.Y] = catHArch
		d.Tiles[position.X+4][position.Y] = catHArchEnd
		d.Tiles[position.X+7][position.Y] = catDArch
		d.Tiles[position.X+8][position.Y] = catHArch

		if d.Tiles[position.X+9][position.Y] != catDWall {
			d.Tiles[position.X+9][position.Y] = catHDirtCorner
		}

		return
	}

	position.X += 11
	d.Tiles[position.X][position.Y+2] = catHWallVArch
	d.Tiles[position.X][position.Y+3] = catVArch
	d.Tiles[position.X][position.Y+4] = catVArchEnd
	d.Tiles[position.X][position.Y+7] = catDArch
	d.Tiles[position.X][position.Y+8] = catVArch

	if d.Tiles[position.X][position.Y+9] != catDWall {
		d.Tiles[position.X][position.Y+9] = catHDirtCorner
	}
}

// generateHall draws the two arch rows of a hall run starting at start
// (levels/drlg_l1.cpp:857-874).
//
//nolint:revive // verticalLayout is upstream GenerateHall's parameter
func (d *Dungeon) generateHall(start Point, length int, verticalLayout bool) {
	if verticalLayout {
		for i := start.Y; i < start.Y+length; i++ {
			d.Tiles[start.X][i] = catVArch
			d.Tiles[start.X+3][i] = catVArch
		}

		return
	}

	for i := start.X; i < start.X+length; i++ {
		d.Tiles[i][start.Y] = catHArch
		d.Tiles[i][start.Y+3] = catHArch
	}
}

// fillChambers draws the three chambers, the halls between them and then the
// quest set piece (levels/drlg_l1.cpp:1036-1075).
func (d *Dungeon) fillChambers(opts Options) error {
	d.drawChambers()
	d.drawHalls()

	return d.initSetPiece(opts)
}

// drawChambers draws the chambers the layout stage kept
// (levels/drlg_l1.cpp:1038-1051).
func (d *Dungeon) drawChambers() {
	chamber1 := Point{X: 0, Y: 14}
	chamber3 := Point{X: 28, Y: 14}

	if d.VerticalLayout {
		swapXY(&chamber1)
		swapXY(&chamber3)
	}

	if d.HasChamber1 {
		d.generateChamber(chamber1, false, true, d.VerticalLayout)
	}

	if d.HasChamber2 {
		d.generateChamber(Point{X: 14, Y: 14}, d.HasChamber1, d.HasChamber3, d.VerticalLayout)
	}

	if d.HasChamber3 {
		d.generateChamber(chamber3, true, false, d.VerticalLayout)
	}
}

// drawHalls draws the halls joining the chambers
// (levels/drlg_l1.cpp:1053-1067).
func (d *Dungeon) drawHalls() {
	hall1 := Point{X: 12, Y: 18}
	hall2 := Point{X: 26, Y: 18}

	if d.VerticalLayout {
		swapXY(&hall1)
		swapXY(&hall2)
	}

	if !d.HasChamber2 {
		d.generateHall(hall1, 16, d.VerticalLayout)

		return
	}

	if d.HasChamber1 {
		d.generateHall(hall1, 2, d.VerticalLayout)
	}

	if d.HasChamber3 {
		d.generateHall(hall2, 2, d.VerticalLayout)
	}
}

// initSetPiece drops the quest set piece into one chamber
// (levels/drlg_l1.cpp:430-447). The crypt branch is unit C's.
func (d *Dungeon) initSetPiece(opts Options) error {
	path := setPiecePath(opts.Quests)
	if path == "" {
		return nil
	}

	data, ok := opts.SetPieces[path]
	if !ok {
		return fmt.Errorf("set piece %q: %w", path, ErrMissingSetPiece)
	}

	position := d.selectChamber()

	d.PlaceDunTiles(data, position, catFloor)
	d.SetPiece = Rectangle{Position: position, Size: GetDunSize(data)}

	return nil
}

// setPiecePath is the .dun InitSetPiece loads for the quests available on this
// level, or "" when the level needs none (levels/drlg_l1.cpp:432-442).
func setPiecePath(quests Quests) string {
	switch {
	case quests.Butcher:
		return "levels/l1data/rnd6.dun"
	case quests.Skelking:
		return "levels/l1data/skngdo.dun"
	case quests.LTBanner:
		return "levels/l1data/banner2.dun"
	default:
		return ""
	}
}

// selectChamber picks the chamber the set piece goes into and returns its
// world-tile origin (levels/drlg_l1.cpp:1330-1356).
func (d *Dungeon) selectChamber() Point {
	chamber := d.pickChamber()

	switch chamber {
	case 1:
		if d.VerticalLayout {
			return Point{X: 16, Y: 2}
		}

		return Point{X: 2, Y: 16}
	case 3:
		if d.VerticalLayout {
			return Point{X: 16, Y: 30}
		}

		return Point{X: 30, Y: 16}
	default:
		return Point{X: 16, Y: 16}
	}
}

// pickChamber draws the chamber number SelectChamber chooses from the chamber
// availability flags (levels/drlg_l1.cpp:1332-1350).
//
//nolint:revive,cyclop // the availability branch chain is upstream SelectChamber's logic, verbatim
func (d *Dungeon) pickChamber() int {
	switch {
	case d.HasChamber1 && d.HasChamber2 && d.HasChamber3:
		return int(d.Rnd.GenerateRnd(3)) + 1
	case d.HasChamber1 && d.HasChamber2:
		return d.pickRandomlyAmong([2]int{2, 1})
	case d.HasChamber1 && d.HasChamber3:
		return d.pickRandomlyAmong([2]int{3, 1})
	case d.HasChamber2 && d.HasChamber3:
		return d.pickRandomlyAmong([2]int{2, 3})
	default:
		// FirstRoom forces chamber 2 when either side chamber is missing, so a
		// single-chamber layout is always chamber 2 (levels/drlg_l1.cpp:1345-1349).
		return 2
	}
}

// pickRandomlyAmong returns one of two values, drawing GenerateRnd(2)
// (engine/random.hpp PickRandomlyAmong).
func (d *Dungeon) pickRandomlyAmong(values [2]int) int {
	return values[d.Rnd.GenerateRnd(2)]
}
