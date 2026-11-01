// Package missiles: the eight-way best-fit direction classifier.
package missiles

// GetDirection ports GetDirection (engine/point.hpp:246-284): it picks the
// direction whose 22.5° cone contains the displacement, comparing 5*dx with
// 2*dy to approximate tan(22.5°).
func GetDirection(start, destination TilePosition) Direction {
	mx := destination.X - start.X
	my := destination.Y - start.Y

	if mx >= 0 {
		return getDirectionEast(my, mx)
	}

	return getDirectionWest(-mx, my)
}

// getDirectionEast ports the mx >= 0 half (point.hpp:252-268).
func getDirectionEast(my, mx int) Direction {
	if my >= 0 {
		if 5*mx <= my*2 {
			return SouthWest
		}

		_dir := South
		if 5*my <= mx*2 {
			_dir = SouthEast
		}

		return _dir
	}

	my = -my
	if 5*mx <= my*2 {
		return NorthEast
	}

	_dir := East
	if 5*my <= mx*2 {
		_dir = SouthEast
	}

	return _dir
}

// getDirectionWest ports the mx < 0 half (point.hpp:269-282).
func getDirectionWest(mx, my int) Direction {
	_dir := West
	if my >= 0 {
		if 5*mx <= my*2 {
			return SouthWest
		}

		if 5*my <= mx*2 {
			_dir = NorthWest
		}

		return _dir
	}

	my = -my
	if 5*mx <= my*2 {
		return NorthEast
	}

	_dir = North
	if 5*my <= mx*2 {
		_dir = NorthWest
	}

	return _dir
}
