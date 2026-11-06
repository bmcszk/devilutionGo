package levels

// This file ports PlaceCryptLights, the Hellfire crypt's static light pass
// (levels/crypt.cpp:817-838), called from CreateL5Dungeon right after Pass3
// (levels/drlg_l1.cpp:1305-1312).

// cryptLavaRadius is the radius PlaceCryptLights lights each lava tile with
// (levels/crypt.cpp:832).
const cryptLavaRadius uint8 = 3

// cryptLavaTiles are the crypt's lava and fire tiles, in upstream order and
// without duplicates (levels/crypt.cpp:819-826).
var cryptLavaTiles = []uint16{
	124, 128, 130, 132, 133, 134, 135, 139, 141, 143, 145, 156, 164, 166,
	167, 168, 169, 170, 182, 190, 192, 195, 196, 199, 200, 254, 266, 273,
	276, 281, 282, 283, 284, 285, 286, 287, 288, 290, 302, 316, 434, 435,
	436, 437, 445, 446, 447, 453, 457, 460, 466, 470, 477, 479, 484, 485,
	486, 490, 507, 537, 557, 559, 561, 563, 564, 568, 569, 572, 578, 580,
	584, 585, 589, 592, 593, 594, 595, 596, 597, 598, 599, 600, 601,
}

// PlaceCryptLights collects the crypt's static light sources: every world tile
// whose dPiece is one of cryptLavaTiles, lit at radius 3
// (levels/crypt.cpp:828-837).
//
// Upstream DoLighting is called per matching tile and stops scanning the tile
// list there, so a tile that matches twice still yields one source.
func PlaceCryptLights(pieces PieceGrid) []Light {
	sources := make([]Light, 0, len(cryptLavaTiles))

	for x := range MAXDUNX {
		for y := range MAXDUNY {
			if isCryptLavaTile(pieces.Piece[x][y]) {
				sources = append(sources, Light{Position: Point{X: x, Y: y}, Radius: cryptLavaRadius})
			}
		}
	}

	return sources
}

// isCryptLavaTile reports whether a dPiece id is in the lava table.
func isCryptLavaTile(tile uint16) bool {
	for _, lavaTile := range cryptLavaTiles {
		if tile == lavaTile {
			return true
		}
	}

	return false
}
