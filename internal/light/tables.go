// Package light ports upstream engine/light_tables.cpp (MakeLightTable).
package light

import "math"

// DungeonType mirrors upstream dungeon_type (gendung_defs.hpp:15-24);
// ordinals matter for parity.
type DungeonType int

const (
	DTYPE_TOWN DungeonType = iota
	DTYPE_CATHEDRAL
	DTYPE_CATACOMBS
	DTYPE_CAVES
	DTYPE_HELL
	DTYPE_NEST
	DTYPE_CRYPT
)

const (
	// LightsMax is the highest light shade index (light_tables.cpp).
	LightsMax = 15
	// NumLightingLevels is the number of light tables (light_tables.cpp).
	NumLightingLevels = 16
	// LightTableSize is the number of entries per light table (light_tables.cpp).
	LightTableSize = 256
	// NumLightRadiuses is the number of light falloff radius tables (light_tables.cpp).
	NumLightRadiuses = 16
)

// MakeLightTables ports MakeLightTable (light_tables.cpp:24-119).
// fullyLit/fullyDark are nil where upstream sets the pointer to nullptr.
//
//nolint:revive // The 5-result signature is mandated by the port contract (upstream light_tables.cpp).
func MakeLightTables(levelType DungeonType) (tables [16][256]uint8, falloffs [16][128]uint8, cone [8][8][16][16]uint8, fullyLit, fullyDark *[256]uint8) {
	shadeSteps := []int{16, 16, 16, 16, 16, 16, 16, 16, 8, 8, 8, 8, 16, 16, 16, 16, 16, 16}
	for s := 0; s < LightsMax+1; s++ {
		c := 0
		for _, steps := range shadeSteps {
			shading := uint8(s * steps / 16)
			shadeStart := c
			shadeEnd := shadeStart + steps - 1
			for step := 0; step < steps; step++ {
				if c == 0 {
					tables[s][c] = 0
					c++
					continue
				}
				color := shadeStart + step + int(shading)
				if color > shadeEnd || color == 255 {
					color = 0
				}
				tables[s][c] = uint8(color)
				c++
			}
		}
	}
	// The last table is pitch black (light_tables.cpp).
	tables[LightsMax] = [LightTableSize]uint8{}
	fullyLit = &tables[0]
	fullyDark = &tables[LightsMax]

	switch levelType {
	case DTYPE_HELL:
		for i := 0; i < LightsMax; i++ {
			for j := 0; j < 16; j++ {
				color := uint8(((16 - 1) << 4) / 15 * (15 - i) / 16 * (j + 1))
				color = uint8(1 + (color >> 4))
				tables[i][j+1] = color
				tables[i][31-j] = color
			}
		}
		fullyLit = nil
	case DTYPE_NEST, DTYPE_CRYPT:
		for i := 0; i < NumLightingLevels; i++ {
			for k := 0; k < 16; k++ {
				tables[i][k] = uint8(k)
			}
		}
		tables[LightsMax][0] = 0
		for k := 1; k < 16; k++ {
			tables[LightsMax][k] = 1
		}
		fullyDark = nil
	}
	// Falloffs (light_tables.cpp).
	const maxDarkness = float32(15)
	for radius := 0; radius < NumLightRadiuses; radius++ {
		maxDistance := (radius + 1) * 8
		for distance := 0; distance < 128; distance++ {
			if distance > maxDistance {
				falloffs[radius][distance] = 15
				continue
			}
			factor := float32(distance) / float32(maxDistance)
			var scaled float32
			if levelType == DTYPE_NEST || levelType == DTYPE_CRYPT {
				brightness := float32(radius) * 1.25
				scaled = factor*factor*brightness + (15 - brightness)
				if scaled < 0 {
					scaled = 0
				}
			} else {
				scaled = factor * maxDarkness
			}
			scaled += 0.5
			falloffs[radius][distance] = uint8(scaled)
		}
	}
	// Light cone tables (light_tables.cpp).
	for offsetY := 0; offsetY < 8; offsetY++ {
		for offsetX := 0; offsetX < 8; offsetX++ {
			for y := 0; y < 16; y++ {
				for x := 0; x < 16; x++ {
					a := 8*x - offsetX
					b := 8*y - offsetY
					cone[offsetX][offsetY][x][y] = uint8(math.Sqrt(float64(a*a + b*b)))
				}
			}
		}
	}

	return tables, falloffs, cone, fullyLit, fullyDark
}
