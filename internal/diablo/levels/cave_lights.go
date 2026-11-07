package levels

// This file ports PlaceCaveLights, the caves' static light pass
// (levels/drlg_l3.cpp:2154-2168), called from CreateL3Dungeon right after Pass3
// (levels/drlg_l3.cpp:2192-2200 via PlaceLights, :2185-2188).

const (
	// caveLightRadius is the radius PlaceCaveLights lights every matching tile
	// with (levels/drlg_l3.cpp:2159, 2161, 2163).
	caveLightRadius uint8 = 7

	// The caves light their lava, walls and ceiling pieces with one radius, in
	// three separate tile ranges (levels/drlg_l3.cpp:2158-2165). The predicates
	// stay disjoint branches upstream; they are only ranges here.
	caveLightBandFirst    uint16 = 55
	caveLightBandLast     uint16 = 146
	caveLightSecondFirst  uint16 = 153
	caveLightSecondLast   uint16 = 160
	caveLightSingletonLow uint16 = 149
	caveLightSingletonHi  uint16 = 151
)

// PlaceCaveLights collects the caves' static light sources: every world tile
// whose dPiece is inside the 55-146 or 153-160 band, or is 149 or 151, lit at
// radius 7 (levels/drlg_l3.cpp:2154-2168). CreateL3Dungeon calls it right after
// Pass3 (:2198-2199).
func PlaceCaveLights(pieces PieceGrid) []Light {
	sources := make([]Light, 0)

	for x := range MAXDUNX {
		for y := range MAXDUNY {
			if !isCaveLightTile(pieces.Piece[x][y]) {
				continue
			}

			sources = append(sources, Light{Position: Point{X: x, Y: y}, Radius: caveLightRadius})
		}
	}

	return sources
}

// isCaveLightTile is PlaceCaveLights' three-branch tile test
// (levels/drlg_l3.cpp:2158-2165). dPiece is uint16 upstream too
// (levels/dun_tile_data.hpp:99), so the comparisons are unsigned as written.
func isCaveLightTile(tile uint16) bool {
	if tile >= caveLightBandFirst && tile <= caveLightBandLast {
		return true
	}
	if tile >= caveLightSecondFirst && tile <= caveLightSecondLast {
		return true
	}

	return tile == caveLightSingletonLow || tile == caveLightSingletonHi
}
