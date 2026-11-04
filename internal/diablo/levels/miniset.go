package levels

// Miniset is a search/replace pattern placed into the dungeon layer
// (levels/gendung.h:37-72). Search and replace are indexed [y][x], like
// upstream.
type Miniset struct {
	Size    Size
	Search  [6][6]Tile
	Replace [6][6]Tile
}

// Matches reports whether the miniset fits at pos. A zero search cell is a
// wildcard and respectProtected makes any protected cell a mismatch
// (levels/gendung.h:48-63).
//
//nolint:revive // respectProtected is upstream Miniset::matches' parameter, not control coupling
func (m Miniset) Matches(d *Dungeon, pos Point, respectProtected bool) bool {
	for yy := range m.Size.Height {
		for xx := range m.Size.Width {
			if m.Search[yy][xx] != 0 && d.Tiles[xx+pos.X][yy+pos.Y] != m.Search[yy][xx] {
				return false
			}

			if respectProtected && d.Protected.Test(xx+pos.X, yy+pos.Y) {
				return false
			}
		}
	}

	return true
}

// Place writes the non-zero replace cells at pos and optionally protects them
// (levels/gendung.h:65-82).
//
//nolint:revive // protect is upstream Miniset::place's parameter, not control coupling
func (m Miniset) Place(d *Dungeon, pos Point, protect bool) {
	for y := range m.Size.Height {
		for x := range m.Size.Width {
			if m.Replace[y][x] == 0 {
				continue
			}

			d.Tiles[x+pos.X][y+pos.Y] = m.Replace[y][x]
			if protect {
				d.Protected.Set(x+pos.X, y+pos.Y, true)
			}
		}
	}
}

// CanReplaceTile is the compatibility workaround PlaceMiniSetRandom applies
// before a random 1x1 replacement (levels/drlg_l1.cpp:350-366). Note the
// constants: the function lives in drlg_l1.cpp, so a crypt replacement is
// still screened with the CATHEDRAL range [VWallEnd2, VWall8] = [84, 100];
// that is upstream behaviour, not a port bug.
//
//nolint:cyclop,revive // 1:1 port of levels/drlg_l1.cpp:350-366; complexity is upstream's
func CanReplaceTile(d *Dungeon, replace Tile, tile Point) bool {
	if replace < catVWallEnd2 || replace > catVWall8 {
		return true
	}

	inBounds := func(p Point) bool {
		return p.X >= 0 && p.X < DMAXX && p.Y >= 0 && p.Y < DMAXY
	}

	// BUGFIX: p2 is a workaround for a bug, only p1 should have been used
	// (fixing this breaks compatibility). Direction::NorthWest is the
	// isometric displacement {-1,0}; see engine/displacement.hpp:257-263.
	p2 := tile.Add(Point{X: -1})
	compare := func(p1 Point) bool {
		return inBounds(p1) && inBounds(p2) &&
			d.Tiles[p1.X][p1.Y] >= catVWallEnd2 && d.Tiles[p2.X][p2.Y] <= catVWall8
	}

	return !compare(p2) &&
		!compare(tile.Add(Point{X: 1})) &&
		!compare(tile.Add(Point{Y: 1})) &&
		!compare(tile.Add(Point{Y: -1}))
}
