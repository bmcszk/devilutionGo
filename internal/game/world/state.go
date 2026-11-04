package world

// Tile is a dungeon tile coordinate (dun_tile_data.hpp positions).
type Tile struct{ X, Y int }

// State is the slice the movement handlers own: the grid plus the one tile the
// player stands on. The only write path is StepTo, which the movement handler owns.
type State struct {
	grid *Grid
	hero Tile
}

// NewState builds the slice on a fresh grid with the player at origin.
func NewState() *State {
	return &State{grid: NewGrid()}
}

// Grid returns the grid the slice owns (read access for the tile predicates
// and for the SOL data loader).
func (s *State) Grid() *Grid { return s.grid }

// Hero returns the tile the player stands on.
func (s *State) Hero() Tile { return s.hero }

// StepTo walks the player to tile when the step is possible: the destination is
// one tile away inside the dungeon, and the grid lets the player step there
// (CanStep reads the diagonal corners, tile_properties.cpp:66-86). Longer walks
// are a path, which this slice does not carry yet; a refused step moves nobody,
// so the handler emits no event (law 2).
func (s *State) StepTo(tile Tile) bool {
	if s.Range(tile) != StepReach || !InDungeonBounds(tile.X, tile.Y) {
		return false
	}

	if !s.grid.CanStep(s.hero.X, s.hero.Y, tile.X, tile.Y) || !s.grid.IsTileWalkable(tile.X, tile.Y) {
		return false
	}

	s.hero = tile
	s.grid.DFlags[tile.Y][tile.X] |= FlagExplored

	return true
}

// Range reports the Chebyshev distance from the player to tile: the number of
// diagonal steps between them, which is how the click gate measures reach.
func (s *State) Range(tile Tile) int {
	dx, dy := absInt(tile.X-s.hero.X), absInt(tile.Y-s.hero.Y)

	return max(dx, dy)
}

// TargetAt reports whether the tile holds a live target. The grid slice holds
// no monsters yet (they land with the entity wave), so no tile has one; the
// movement slice answers honestly instead of inventing an occupant.
func (*State) TargetAt(Tile) (alive, ok bool) { return false, false }

func absInt(v int) int {
	if v < 0 {
		return -v
	}

	return v
}
