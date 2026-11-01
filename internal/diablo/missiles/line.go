// Package missiles: the line-of-sight walk AddGuardian and the Apocalypse
// rows test their candidate tiles with.
package missiles

import "github.com/bmcszk/devilutionGo/internal/game/world"

// LineClear ports LineClear (monster.cpp:1599-1670): a Bresenham walk from
// startPoint to endPoint that stops at the first tile whose isClear predicate
// fails, and reports whether it reached endPoint. The walk never tests
// startPoint itself (monster.cpp:1633, 1666).
func LineClear(isClear func(x, y int) bool, startPoint, endPoint TilePosition) bool {
	position := startPoint

	dx := endPoint.X - position.X
	dy := endPoint.Y - position.Y

	if absInt(dx) > absInt(dy) {
		if dx < 0 {
			position, endPoint = endPoint, position
			dx, dy = -dx, -dy
		}

		position = lineClearAlongX(position, endPoint, startPoint, dx, dy, isClear)
	} else {
		if dy < 0 {
			position, endPoint = endPoint, position
			dy, dx = -dy, -dx
		}

		position = lineClearAlongY(position, endPoint, startPoint, dx, dy, isClear)
	}

	return position == endPoint
}

// LineClearMissile ports LineClearMissile (monster.cpp:4393-4396): the walk
// that only stops on tiles blocking a missile (PosOkMissile,
// monster.cpp:3281-3284).
func LineClearMissile(w *HitWorld, startPoint, endPoint TilePosition) bool {
	return LineClear(func(x, y int) bool {
		return !w.Grid.TileHasAny(x, y, world.TileBlockMissile)
	}, startPoint, endPoint)
}

// lineClearAlongX ports the |dx| > |dy| arm (monster.cpp:1605-1637).
func lineClearAlongX(position, endPoint, startPoint TilePosition, dx, dy int,
	isClear func(x, y int) bool,
) TilePosition {
	var d, yIncD, dIncD, dIncH int

	if dy > 0 {
		d, dIncD, dIncH, yIncD = 2*dy-dx, 2*dy, 2*(dy-dx), 1
	} else {
		d, dIncD, dIncH, yIncD = 2*dy+dx, 2*dy, 2*(dx+dy), -1
	}

	for done := false; !done && position != endPoint; {
		if (d <= 0) != (yIncD < 0) {
			d += dIncD
		} else {
			d += dIncH
			position.Y += yIncD
		}

		position.X++
		done = position != startPoint && !isClear(position.X, position.Y)
	}

	return position
}

// lineClearAlongY ports the |dx| <= |dy| arm (monster.cpp:1639-1670).
func lineClearAlongY(position, endPoint, startPoint TilePosition, dx, dy int,
	isClear func(x, y int) bool,
) TilePosition {
	var d, xIncD, dIncD, dIncH int

	if dx > 0 {
		d, dIncD, dIncH, xIncD = 2*dx-dy, 2*dx, 2*(dx-dy), 1
	} else {
		d, dIncD, dIncH, xIncD = 2*dx+dy, 2*dx, 2*(dy+dx), -1
	}

	for done := false; !done && position != endPoint; {
		if (d <= 0) != (xIncD < 0) {
			d += dIncD
		} else {
			d += dIncH
			position.X += xIncD
		}

		position.Y++
		done = position != startPoint && !isClear(position.X, position.Y)
	}

	return position
}

// absInt is upstream's std::abs for the two axis deltas.
func absInt(v int) int {
	if v < 0 {
		return -v
	}

	return v
}
