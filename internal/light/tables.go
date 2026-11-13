// Package light provides the light, falloff and cone tables used for tile lighting.
package light

import (
	"math"
)

// DTYPE_* are dungeon_type's ordinals: the values .sv files carry as
// quest._qlvltype (SaveQuest WriteLE<uint8_t>, loadsave.cpp:1775; read back at
// loadsave.cpp:993) and in the header leveltype field (WriteBE<uint32_t>,
// loadsave.cpp:2797; read at :2502). Verified at devilutionX e1ad0c2b6
// (Source/levels/gendung_defs.hpp:15-26): TOWN=0, CATHEDRAL=1, CATACOMBS=2,
// CAVES=3, HELL=4, NEST=5, CRYPT=6, LAST=CRYPT, NONE=-1. The light-table
// switch dispatches on these ordinals (light_tables.cpp leveltype branches).
const (
	DTYPE_TOWN      LevelType = 0
	DTYPE_CATHEDRAL LevelType = 1
	DTYPE_CATACOMBS LevelType = 2
	DTYPE_CAVES     LevelType = 3
	DTYPE_HELL      LevelType = 4
	DTYPE_NEST      LevelType = 5
	DTYPE_CRYPT     LevelType = 6
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

// LevelType is dungeon_type (levels/gendung_defs.hpp:15-26). This package owns
// the type rather than importing internal/diablo/gamedata/leveldat, so the
// engine-side light tables stay independent of the Diablo layer (ADR-008); the
// Diablo side maps its own enum onto this one.
type LevelType uint8

// LightTables bundles the light, falloff and cone tables built for one level
// type (light_tables.cpp MakeLighTables result). FullyLit and FullyDark are
// nil for the level types that have no such table.
type LightTables struct {
	Tables    [16][256]uint8
	Falloffs  [16][128]uint8
	Cone      [8][8][16][16]uint8
	FullyLit  *[256]uint8
	FullyDark *[256]uint8
}

// MakeLightTables builds the light, falloff and cone tables for a level type.
//
// Ported 1:1 from light_tables.cpp:24-119 in three phases:
// buildLightTables (MakeLightTable + hell/nest branches),
// buildLightFalloffs (LightFalloffs), buildLightCone (LightConeInterpolations).
func MakeLightTables(levelType LevelType) LightTables {
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

	return LightTables{
		Tables:    tables,
		Falloffs:  buildLightFalloffs(levelType),
		Cone:      buildLightCone(),
		FullyLit:  fullyLit,
		FullyDark: fullyDark,
	}
}

// buildLightTables ports the MakeLightTable shade ramp (light_tables.cpp:24-45)
// and the per-level-type overrides: DTYPE_HELL (:47-61) and DTYPE_NEST/CRYPT (:62-69).
// The highest shade is pitch black.
func buildLightTables(levelType LevelType) [16][256]uint8 {
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

	for s := range LightsMax + 1 {
		c := 0

		for _, steps := range shadeSteps {
			shading := uint8(s * steps / 16)
			shadeStart := c

			shadeEnd := shadeStart + steps - 1
			for step := range steps {
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
	for i := range LightsMax {
		for j := range 16 {
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
	for i := range NumLightingLevels {
		for k := range 16 {
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
func buildLightFalloffs(levelType LevelType) [16][128]uint8 {
	var falloffs [16][128]uint8

	const maxDarkness = float32(15)

	for radius := range NumLightRadiuses {
		maxDistance := (radius + 1) * 8
		for distance := range 128 {
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
func falloffScaled(levelType LevelType, radius int, factor, maxDarkness float32) float32 {
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

	for offsetY := range 8 {
		for offsetX := range 8 {
			fillConeOffsets(&cone[offsetX][offsetY], offsetX, offsetY)
		}
	}

	return cone
}

// fillConeOffsets ports one LightConeInterpolations sub-table
// (light_tables.cpp:107-114): Euclidean distance from the offset origin.
func fillConeOffsets(plane *[16][16]uint8, offsetX, offsetY int) {
	for y := range 16 {
		for x := range 16 {
			a := 8*x - offsetX
			b := 8*y - offsetY
			plane[x][y] = uint8(math.Sqrt(float64(a*a + b*b)))
		}
	}
}
