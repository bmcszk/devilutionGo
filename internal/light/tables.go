// Package light provides the light, falloff and cone tables used for tile lighting.
package light

import (
	"math"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat"
)

// DungeonType aliases the canonical leveldat port; its ordinal values are
// part of the parity contract.
type DungeonType = leveldat.DungeonType

const (
	DTYPE_TOWN      = leveldat.DTTown
	DTYPE_CATHEDRAL = leveldat.DTCathedral
	DTYPE_CATACOMBS = leveldat.DTCatacombs
	DTYPE_CAVES     = leveldat.DTCaves
	DTYPE_HELL      = leveldat.DTHell
	DTYPE_NEST      = leveldat.DTNest
	DTYPE_CRYPT     = leveldat.DTCrypt
)

const (
	// LightsMax is the highest light shade index.
	LightsMax = 15
	// NumLightingLevels is the number of light tables.
	NumLightingLevels = 16
	// LightTableSize is the number of entries per light table.
	LightTableSize = 256
	// NumLightRadiuses is the number of light falloff radius tables.
	NumLightRadiuses = 16
)

// lightTables bundles the tables makeLightTables builds for one level type.
// fullyLit and fullyDark are nil for the level types that have no such table.
type lightTables struct {
	tables    [16][256]uint8
	falloffs  [16][128]uint8
	cone      [8][8][16][16]uint8
	fullyLit  *[256]uint8
	fullyDark *[256]uint8
}

// makeLightTables builds the light, falloff and cone tables for a level type.
//
// Ported 1:1 from light_tables.cpp:24-119 in three phases:
// buildLightTables (MakeLightTable + hell/nest branches),
// buildLightFalloffs (LightFalloffs), buildLightCone (LightConeInterpolations).
func makeLightTables(levelType DungeonType) lightTables {
	tables := buildLightTables(levelType)
	fullyLit := &tables[0]
	fullyDark := &tables[LightsMax]
	// Hell clears the identity table (ceiling color map always applied,
	// light_tables.cpp:60); NEST/CRYPT clear the black table (:68).
	if levelType == DTYPE_HELL {
		fullyLit = nil
	}
	if levelType == DTYPE_NEST || levelType == DTYPE_CRYPT {
		fullyDark = nil
	}
	return lightTables{
		tables:    tables,
		falloffs:  buildLightFalloffs(levelType),
		cone:      buildLightCone(),
		fullyLit:  fullyLit,
		fullyDark: fullyDark,
	}
}

// buildLightTables ports the MakeLightTable shade ramp (light_tables.cpp:24-45)
// and the per-level-type overrides: DTYPE_HELL (:47-61) and DTYPE_NEST/CRYPT (:62-69).
// The highest shade is pitch black.
func buildLightTables(levelType DungeonType) [16][256]uint8 {
	var tables [16][256]uint8
	buildShadeRamp(&tables)
	// The highest shade is pitch black.
	tables[LightsMax] = [LightTableSize]uint8{}

	switch levelType {
	case DTYPE_HELL:
		applyHellBloodWallOverride(&tables) // light_tables.cpp:47-61
	case DTYPE_NEST, DTYPE_CRYPT:
		applyLavaBand(&tables) // light_tables.cpp:62-69
	default:
		// DTYPE_TOWN and DTYPE_CATHEDRAL use the base tables unchanged.
	}

	return tables
}

// buildShadeRamp ports the MakeLightTable shade-step loop (light_tables.cpp:24-45):
// 18 ranges of 16 or 8 steps, index 0 forced black, white (255) collapsing to 0.
func buildShadeRamp(tables *[16][256]uint8) {
	shadeSteps := []int{16, 16, 16, 16, 16, 16, 16, 16, 8, 8, 8, 8, 16, 16, 16, 16, 16, 16}
	for s := 0; s < LightsMax+1; s++ {
		c := 0
		for _, steps := range shadeSteps {
			shading := uint8(s * steps / 16)
			shadeStart := c
			shadeEnd := shadeStart + steps - 1
			for step := 0; step < steps; step++ {
				c = writeShadedColor(tables, s, c, shadeStart, shadeEnd, step, int(shading))
			}
		}
	}
}

// writeShadedColor ports the inner MakeLightTable step (light_tables.cpp:34-43):
// index 0 stays black; white (255) and out-of-range colors collapse to black.
// Returns the next color index.
func writeShadedColor(tables *[16][256]uint8, shade, c, shadeStart, shadeEnd, step, shading int) int {
	if c == 0 {
		tables[shade][c] = 0
		return c + 1
	}
	color := shadeStart + step + shading
	if color > shadeEnd || color == 255 {
		color = 0
	}
	tables[shade][c] = uint8(color)
	return c + 1
}

// applyHellBloodWallOverride ports the DTYPE_HELL branch (light_tables.cpp:47-61):
// the blood-wall animation color band mirrored around the 16-color boundary.
func applyHellBloodWallOverride(tables *[16][256]uint8) {
	for i := 0; i < LightsMax; i++ {
		for j := 0; j < 16; j++ {
			color := uint8(((16 - 1) << 4) / 15 * (15 - i) / 16 * (j + 1))
			color = uint8(1 + (color >> 4))
			tables[i][j+1] = color
			tables[i][31-j] = color
		}
	}
}

// applyLavaBand ports the DTYPE_NEST/CRYPT branch (light_tables.cpp:62-69):
// an identity 0..15 lava band on every shade, with the darkest table pinned
// to color 1 for indices 1..15.
func applyLavaBand(tables *[16][256]uint8) {
	for i := 0; i < NumLightingLevels; i++ {
		for k := 0; k < 16; k++ {
			tables[i][k] = uint8(k)
		}
	}
	tables[LightsMax][0] = 0
	for k := 1; k < 16; k++ {
		tables[LightsMax][k] = 1
	}
}

// buildLightFalloffs ports LightFalloffs (light_tables.cpp:76-101).
// float32 is used exactly as the upstream C++ float math.
func buildLightFalloffs(levelType DungeonType) [16][128]uint8 {
	var falloffs [16][128]uint8
	const maxDarkness = float32(15)
	for radius := 0; radius < NumLightRadiuses; radius++ {
		maxDistance := (radius + 1) * 8
		for distance := 0; distance < 128; distance++ {
			if distance > maxDistance {
				falloffs[radius][distance] = 15
				continue
			}
			factor := float32(distance) / float32(maxDistance)
			scaled := falloffScaled(levelType, radius, factor, maxDarkness)
			scaled += 0.5
			falloffs[radius][distance] = uint8(scaled)
		}
	}
	return falloffs
}

// falloffScaled ports the per-distance scaling branch (light_tables.cpp:91-98):
// quadratic falloff with over-exposure on NEST/CRYPT, linear lean falloff otherwise.
func falloffScaled(levelType DungeonType, radius int, factor, maxDarkness float32) float32 {
	if levelType == DTYPE_NEST || levelType == DTYPE_CRYPT {
		brightness := float32(radius) * 1.25
		scaled := factor*factor*brightness + (15 - brightness)
		if scaled < 0 {
			return 0
		}
		return scaled
	}
	return factor * maxDarkness
}

// buildLightCone ports LightConeInterpolations (light_tables.cpp:104-117):
// per-offset Euclidean distance tables used by the light cone rasterizer.
func buildLightCone() [8][8][16][16]uint8 {
	var cone [8][8][16][16]uint8
	for offsetY := 0; offsetY < 8; offsetY++ {
		for offsetX := 0; offsetX < 8; offsetX++ {
			fillConeOffsets(&cone[offsetX][offsetY], offsetX, offsetY)
		}
	}
	return cone
}

// fillConeOffsets ports one LightConeInterpolations sub-table
// (light_tables.cpp:107-114): Euclidean distance from the offset origin.
func fillConeOffsets(plane *[16][16]uint8, offsetX, offsetY int) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			a := 8*x - offsetX
			b := 8*y - offsetY
			plane[x][y] = uint8(math.Sqrt(float64(a*a + b*b)))
		}
	}
}
