package levels

// This file ports ApplyShadowsPatterns (levels/drlg_l1.cpp:332-395): the pass
// that stamps the 2x2 shadow patterns onto the corners the decoration passes
// left, and then picks the fence-aware variant of the shadow tiles.

// applyShadowsPatterns runs the two ApplyShadowsPatterns scans
// (levels/drlg_l1.cpp:332-395).
func (d *Dungeon) applyShadowsPatterns() {
	d.stampShadowPatterns()
	d.fixShadowFences()
}

// stampShadowPatterns matches the 2x2 base-type window around every tile
// against the shadow table and writes the shadow tiles it finds
// (levels/drlg_l1.cpp:332-367).
func (d *Dungeon) stampShadowPatterns() {
	for y := 1; y < DMAXY; y++ {
		for x := 1; x < DMAXX; x++ {
			d.stampShadowPatternAt(x, y)
		}
	}
}

// stampShadowPatternAt applies every matching shadow pattern at one window
// (levels/drlg_l1.cpp:336-365). The window is read before any pattern runs, so
// patterns do not see each other's writes.
func (d *Dungeon) stampShadowPatternAt(x, y int) {
	var window [2][2]Tile
	window[0][0] = baseTypes[d.Tiles[x][y]]
	window[1][0] = baseTypes[d.Tiles[x-1][y]]
	window[0][1] = baseTypes[d.Tiles[x][y-1]]
	window[1][1] = baseTypes[d.Tiles[x-1][y-1]]

	for _, shadow := range shadowPatterns {
		if shadow.matches(window) {
			d.applyShadowPattern(x, y, shadow)
		}
	}
}

// applyShadowPattern writes the three shadow tiles one matched pattern asks
// for, skipping protected tiles (levels/drlg_l1.cpp:348-364).
func (d *Dungeon) applyShadowPattern(x, y int, shadow shadowPattern) {
	if shadow.nv1 != 0 && !d.Protected.Test(x-1, y-1) {
		d.Tiles[x-1][y-1] = shadow.nv1
	}

	if shadow.nv2 != 0 && !d.Protected.Test(x, y-1) {
		d.Tiles[x][y-1] = shadow.nv2
	}

	if shadow.nv3 != 0 && !d.Protected.Test(x-1, y) {
		d.Tiles[x-1][y] = shadow.nv3
	}
}

// matches reports whether the 2x2 base-type window fits the pattern's
// requirements; a zero requirement matches anything
// (levels/drlg_l1.cpp:337-346).
func (s shadowPattern) matches(window [2][2]Tile) bool {
	if s.strig != window[0][0] {
		return false
	}

	if s.s1 != 0 && s.s1 != window[1][1] {
		return false
	}

	if s.s2 != 0 && s.s2 != window[0][1] {
		return false
	}

	return s.s3 == 0 || s.s3 == window[1][0]
}

// fixShadowFences picks the fence-aware shadow variants of the tiles the
// pattern scan stamped (levels/drlg_l1.cpp:369-394).
func (d *Dungeon) fixShadowFences() {
	for y := 1; y < DMAXY; y++ {
		for x := 1; x < DMAXX; x++ {
			if d.Protected.Test(x-1, y) {
				continue
			}

			d.Tiles[x-1][y] = fenceShadowVariant(d.Tiles[x-1][y], d.Tiles[x][y])
		}
	}
}

// fenceShadowVariant is the per-tile half of FixShadowFences: a shadow next to
// a fence keeps its fence variant (levels/drlg_l1.cpp:374-391).
func fenceShadowVariant(tile, east Tile) Tile {
	if !isFence(east) {
		return tile
	}

	switch tile {
	case catFloor12:
		return catFloor14
	case catHArchShadow:
		return catHArchShadow2
	case catHWallShadow:
		return catHWallShadow2
	default:
		return tile
	}
}

// isFence reports whether a tile is one of the fence tiles the shadow variant
// keys on (levels/drlg_l1.cpp:377).
func isFence(tile Tile) bool {
	switch tile {
	case catDFence, catVFenceEnd, catVFence, catHWallVFence, catHArchVFence, catHArchVDoor:
		return true
	default:
		return false
	}
}
