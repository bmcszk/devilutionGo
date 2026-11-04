package levels

// This file ports the crypt shadow pass, which stamps the shadow tiles an
// arch, wall end or pillar casts onto the floor next to it
// (levels/crypt.cpp:486-609).

// cryptShadowRules is ApplyCryptShadowsPatterns' switch in upstream order
// (levels/crypt.cpp:491-605).
var cryptShadowRules = []cryptShadowRule{
	{
		tiles: []Tile{cryptDArch, cryptDArch2, cryptDArch3},
		west:  cryptShadow1, northWest: cryptShadow2, north: cryptShadow3,
	},
	{
		tiles: []Tile{
			cryptHWallEnd, cryptHWallEnd2, cryptHWallEnd3, cryptHWallEnd4,
			cryptPillar, cryptPillar2, cryptPillar3, cryptPillar5, cryptPillar9,
		},
		west: cryptShadow4, northWest: cryptShadow5,
	},
	{
		tiles: []Tile{
			cryptHArchEnd, cryptHArchEnd2, cryptHArchEnd3, cryptHArchEnd4,
			cryptHWallVArch, cryptHWallVArch2, cryptHWallVArch3, cryptHWallVArch4,
			cryptVArch, cryptVArch4, cryptVArch5, cryptVArch6, cryptVArch7,
		},
		west: cryptShadow1, northWest: cryptShadow2,
	},
	{
		tiles: []Tile{cryptVArchEnd, cryptVArchEnd2, cryptVArchEnd4},
		west:  cryptShadow4, northWest: cryptShadow5, north: cryptShadow3,
	},
	{
		tiles: []Tile{
			cryptHArch, cryptHArch2,
			cryptHArchVWall, cryptHArchVWall2, cryptHArchVWall3, cryptHArchVWall4,
		},
		north: cryptShadow3,
	},
	{
		tiles: []Tile{cryptHArch5, cryptHArch6},
		north: cryptShadow6,
	},
	{
		tiles: []Tile{cryptVArch2},
		west:  cryptShadow9, northWest: cryptShadow10,
	},
	{
		tiles: []Tile{cryptVArchEnd3},
		west:  cryptShadow11, northWest: cryptShadow12, north: cryptShadow3,
	},
	{
		tiles: []Tile{cryptVArch3},
		west:  cryptShadow13, northWest: cryptShadow14,
	},
	{
		tiles: []Tile{cryptHArch3, cryptHArch4},
		north: cryptShadow15,
	},
	{
		tiles: []Tile{cryptPillar6, cryptPillar8},
		west:  cryptShadow11, northWest: cryptShadow12,
	},
	{
		tiles: []Tile{cryptDArch4},
		west:  cryptShadow1, northWest: cryptShadow2, north: cryptShadow15,
	},
	{
		tiles: []Tile{cryptPillar11, cryptPillar12, cryptPillarHalf},
		west:  cryptShadow7, northWest: cryptShadow8,
	},
}

// cryptShadowRule is one tile group of ApplyCryptShadowsPatterns and the shadow
// each floor neighbour receives.
type cryptShadowRule struct {
	// tiles are the tile ids that select this rule.
	tiles []Tile
	// west shadows the floor at (i-1,j); northWest shadows (i-1,j-1); north
	// shadows (i,j-1). A zero value leaves that neighbour alone.
	west      Tile
	northWest Tile
	north     Tile
}

// applyCryptShadowsPatterns walks the layer from row 1 and column 1, so the
// neighbours a rule reads are always in bounds
// (levels/crypt.cpp:486-609).
func (d *Dungeon) applyCryptShadowsPatterns() {
	for j := 1; j < DMAXY; j++ {
		for i := 1; i < DMAXX; i++ {
			d.applyCryptShadowAt(i, j)
		}
	}
}

// applyCryptShadowAt stamps the shadow group of one tile onto its floor
// neighbours (levels/crypt.cpp:491-605).
func (d *Dungeon) applyCryptShadowAt(i, j int) {
	for _, rule := range cryptShadowRules {
		if rule.spreads(d.Tiles[i][j]) {
			rule.apply(d, i, j)

			return
		}
	}
}

// spreads reports whether this rule is the one for the given tile.
func (r cryptShadowRule) spreads(tile Tile) bool {
	for _, candidate := range r.tiles {
		if candidate == tile {
			return true
		}
	}

	return false
}

// apply writes the rule's shadow tiles onto the floor neighbours of (i,j).
func (r cryptShadowRule) apply(d *Dungeon, i, j int) {
	stampShadow(d, i-1, j, r.west)
	stampShadow(d, i-1, j-1, r.northWest)
	stampShadow(d, i, j-1, r.north)
}

// stampShadow replaces one floor tile with a shadow tile.
func stampShadow(d *Dungeon, i, j int, shadow Tile) {
	if shadow != 0 && d.Tiles[i][j] == cryptFloor {
		d.Tiles[i][j] = shadow
	}
}
