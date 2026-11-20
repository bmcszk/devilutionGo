package world

// Tile is a dungeon tile coordinate (dun_tile_data.hpp positions).
type Tile struct{ X, Y int }

// State is the slice the movement handlers own: the grid plus the one tile the
// player stands on. The write paths are StepTo, LoadLevel, CreateFromPack and
// TakeFloorItem — one per verb handler that owns a field.
type State struct {
	grid  *Grid
	hero  Tile
	level int
	kind  int
	// loaded records that a dungeon level has been loaded, so the gate can
	// tell the first descend from a reload of the same level.
	loaded bool
	// pack is the hero the session raised, zero until CreateFromPack runs.
	pack       StartingPack
	packRaised bool
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

// Level implements kernel.LevelView: the level number and dungeon_type the
// session loaded, and ok=false in town and before the first descend.
func (s *State) Level() (level, levelType int, ok bool) { return s.level, s.kind, s.loaded }

// CreateFromPack raises the hero the pack describes onto tile — the state half
// of CreatePlayer (player.cpp:2283-2330): the class, its starting stats and its
// starting vitals. The tile must be somewhere the player may stand, so a hero is
// never raised into a wall; ok=false changes nothing.
func (s *State) CreateFromPack(pack StartingPack, tile Tile) bool {
	if !s.posOkPlayer(tile) {
		return false
	}

	s.hero = tile
	s.pack = pack
	s.packRaised = true

	return true
}

// RaisedPack reports the hero the session raised: the pack and its tile, and
// ok=false while no CreatePack has raised one.
func (s *State) RaisedPack() (StartingPack, Tile, bool) {
	return s.pack, s.hero, s.packRaised
}

// HeroRaised reports whether a hero has been raised, the state a save needs to
// have anything to freeze.
func (s *State) HeroRaised() bool { return s.packRaised }

// SpawnAllowed reports whether a hero may be raised on tile: the tile rules the
// player's own position already answers, so a create cannot land in a wall.
func (s *State) SpawnAllowed(tile Tile) bool { return s.posOkPlayer(tile) }

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

// Sub returns the component-wise difference a - b.
func (a Tile) Sub(b Tile) Tile { return Tile{X: a.X - b.X, Y: a.Y - b.Y} }

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

// sessionRecord freezes the session into the value a save slot carries: the
// hero's tile, the loaded level's identity and the pack the hero was raised
// from. A session that raised no hero has nothing worth freezing, so it reports
// false and the save stays silent (ADR-018 law 2).
func (s *State) sessionRecord() (SessionRecord, bool) {
	if !s.packRaised {
		return SessionRecord{}, false
	}

	return SessionRecord{
		Level: s.level, LevelType: s.kind, Loaded: s.loaded, Hero: s.hero, Pack: s.pack,
	}, true
}

// restoreSession puts a saved session back: the level image the source serves is
// loaded and the hero is dropped onto the saved tile through the same
// SyncInitPlrPos search a descend runs (player.cpp:3235-3271), so a level whose
// saved tile is a wall still restores onto a legal one. A town record has no
// level to load, so only the hero and the pack move; ok=false changes nothing.
func (s *State) restoreSession(record SessionRecord, source LevelSource) (Place, bool) {
	if !record.Loaded {
		s.hero, s.pack, s.packRaised = record.Hero, record.Pack, true

		return Place{}, true
	}

	layout, _, ok := source.LevelImage(record.Level, record.LevelType)
	if !ok {
		return Place{}, false
	}

	place, ok := s.LoadLevel(layout, record.Hero, record.Level, record.LevelType)
	if !ok {
		return Place{}, false
	}

	s.pack, s.packRaised = record.Pack, true

	return place, true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}

	return v
}
