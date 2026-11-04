package levels

// Point is a mega-tile or world-tile coordinate (engine/point.hpp:20-40).
type Point struct{ X, Y int }

// Size is a width/height pair in tiles (engine/size.hpp).
type Size struct{ Width, Height int }

// Rectangle is a position plus a size (engine/rectangle.hpp:20-60).
type Rectangle struct {
	Position Point
	Size     Size
}

// MegaToWorld converts a mega-tile coordinate to the northmost of its four
// world tiles (engine/point.hpp:164-168).
func (p Point) MegaToWorld() Point {
	return Point{X: 16 + 2*p.X, Y: 16 + 2*p.Y}
}

// WorldToMega converts a world-tile coordinate back to its mega tile
// (engine/point.hpp:170-173).
func (p Point) WorldToMega() Point {
	return Point{X: (p.X - 16) / 2, Y: (p.Y - 16) / 2}
}

// Add returns p displaced by q (engine/point.hpp operator+).
func (p Point) Add(q Point) Point { return Point{X: p.X + q.X, Y: p.Y + q.Y} }

// Contains reports whether p lies inside r (engine/rectangle.hpp:41-52).
func (r Rectangle) Contains(p Point) bool {
	return p.X >= r.Position.X && p.X < r.Position.X+r.Size.Width &&
		p.Y >= r.Position.Y && p.Y < r.Position.Y+r.Size.Height
}
