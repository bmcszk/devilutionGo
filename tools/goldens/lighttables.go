package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

// Independent Go transcription of devilutionX MakeLightTable
// (Source/engine/light_tables.cpp @ ca91d778fba92a1f680afb1fb2cdf09337ff1b42)
// for the light-table TSV goldens under internal/testdata/gamedata/lighttables.
// It deliberately does NOT call internal/light: the goldens are the independent
// oracle the light tests compare against. cone.tsv stays probe-only
// (dvg_probe lighttables); this refresher regenerates the seven level-type
// tables and both falloff branches.

// dungeon_type ordinals used by the level-type switch
// (devilutionX Source/levels/gendung_defs.hpp:15-22).
const (
	dtypeTown      = 0
	dtypeCathedral = 1
	dtypeCatacombs = 2
	dtypeCaves     = 3
	dtypeHell      = 4
	dtypeNest      = 5
	dtypeCrypt     = 6
)

const (
	// lightTSVSource is the # source header line: the upstream file every
	// byte here is transcribed from.
	lightTSVSource = "devilutionX/Source/engine/light_tables.cpp"
	// lightRowBytes is the TSV row width, matching the committed goldens'
	// "# dims: 16x<rows>" headers and scripts/convert_lighttables.py.
	lightRowBytes = 16
)

// shadeSteps ports the per-shade range sizes (light_tables.cpp:24-25):
// 18 ranges of 16 or 8 steps.
var shadeSteps = [...]int{16, 16, 16, 16, 16, 16, 16, 16, 8, 8, 8, 8, 16, 16, 16, 16, 16, 16}

// lighttableTargets lists every light-table golden the subcommand regenerates:
// seven level-type tables (16x256) plus the quadratic falloff (the committed
// falloffs.tsv, dumped from NEST) and the linear falloff branch.
var lighttableTargets = []lighttableTarget{
	{"town_tables", func() []byte { return flattenLightTables(dtypeTown) }},
	{"cathedral_tables", func() []byte { return flattenLightTables(dtypeCathedral) }},
	{"catacombs_tables", func() []byte { return flattenLightTables(dtypeCatacombs) }},
	{"caves_tables", func() []byte { return flattenLightTables(dtypeCaves) }},
	{"hell_tables", func() []byte { return flattenLightTables(dtypeHell) }},
	{"nest_tables", func() []byte { return flattenLightTables(dtypeNest) }},
	{"crypt_tables", func() []byte { return flattenLightTables(dtypeCrypt) }},
	{"falloffs", func() []byte { return flattenLightFalloffs(dtypeNest) }},
	{"falloffs_linear", func() []byte { return flattenLightFalloffs(dtypeTown) }},
}

// lighttableTarget is one golden file and the transcription that fills it.
type lighttableTarget struct {
	name string
	data func() []byte
}

// runLighttables writes every light-table golden TSV into the output directory
// (positional path 1).
func runLighttables(paths []string) error {
	outDir := paths[0]

	for _, target := range lighttableTargets {
		path := filepath.Join(outDir, target.name+".tsv")
		if err := os.WriteFile(path, encodeLightTSV(target.name, target.data()), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	slog.Info("goldens lighttables done", "files", len(lighttableTargets), "out", outDir)

	return nil
}

// encodeLightTSV renders one byte table as the committed golden format the
// light tests decode: # table/# dims/# source headers, then rows of 16
// tab-separated decimals (the scripts/convert_lighttables.py layout).
func encodeLightTSV(name string, data []byte) []byte {
	rows := (len(data) + lightRowBytes - 1) / lightRowBytes

	out := []byte("# table: " + name + "\n" +
		"# dims: " + strconv.Itoa(lightRowBytes) + "x" + strconv.Itoa(rows) + "\n" +
		"# source: " + lightTSVSource + "\n")

	for r := range rows {
		row := data[r*lightRowBytes : min((r+1)*lightRowBytes, len(data))]
		for i, b := range row {
			if i > 0 {
				out = append(out, '\t')
			}

			out = append(out, strconv.AppendUint(nil, uint64(b), 10)...)
		}

		out = append(out, '\n')
	}

	return out
}

// makeLightTables ports MakeLightTable (light_tables.cpp:24-69): the 16-shade
// ramp, the pitch-black last table, then the per-level-type overrides.
func makeLightTables(levelType int) [16][256]uint8 {
	tables := buildShadeRamp()
	tables[15] = [256]uint8{} // Make last shade pitch black (light_tables.cpp:46)

	switch levelType {
	case dtypeHell:
		applyBloodWall(&tables) // light_tables.cpp:47-61
	case dtypeNest, dtypeCrypt:
		applyLavaBright(&tables) // light_tables.cpp:62-69
	default:
		// TOWN/CATHEDRAL/CATACOMBS/CAVES keep the base ramp (light_tables.cpp:47).
	}

	return tables
}

// buildShadeRamp ports the shade-step loop (light_tables.cpp:23-45): one ramp
// per shade, index 0 forced black, white (255) collapsing to black.
func buildShadeRamp() [16][256]uint8 {
	var tables [16][256]uint8

	for shade := range 16 {
		fillShadeTable(&tables[shade], shade)
	}

	return tables
}

// fillShadeTable ports one MakeLightTable shade iteration
// (light_tables.cpp:26-44) into one 256-entry table.
func fillShadeTable(table *[256]uint8, shade int) {
	colorIndex := 0

	for _, steps := range shadeSteps {
		shading := uint8(shade * steps / 16)
		shadeStart := uint8(colorIndex)
		shadeEnd := shadeStart + uint8(steps) - 1

		for step := range steps {
			colorIndex = writeShadedColor(table, colorIndex, shadeStart, shadeEnd, shading, step)
		}
	}
}

// writeShadedColor ports the inner MakeLightTable step (light_tables.cpp:34-43):
// index 0 stays black; white (255) and out-of-range colors collapse to black.
// Returns the next color index.
func writeShadedColor(table *[256]uint8, colorIndex int, shadeStart, shadeEnd, shading uint8, step int) int {
	if colorIndex == 0 { // Black (light_tables.cpp:35-38)
		return colorIndex + 1
	}

	color := int(shadeStart) + step + int(shading)
	if color > int(shadeEnd) || color == 255 { // White (light_tables.cpp:40-42)
		color = 0
	}

	table[colorIndex] = uint8(color)

	return colorIndex + 1
}

// applyBloodWall ports the DTYPE_HELL branch (light_tables.cpp:47-61): the
// blood-wall color band mirrored around the 16-color boundary on every
// non-black shade.
func applyBloodWall(tables *[16][256]uint8) {
	const (
		rangeEnd = 16
		shades   = 15 // LightTables.size() - 1 (light_tables.cpp:48)
	)

	for i := range shades {
		for j := range rangeEnd {
			color := uint8((rangeEnd - 1) << 4 / shades * (shades - i) / rangeEnd * (j + 1)) // :52
			color = 1 + (color >> 4)                                                         // :53
			tables[i][j+1] = color                                                           // :55
			tables[i][31-j] = color                                                          // :57
		}
	}
}

// applyLavaBright ports the DTYPE_NEST/CRYPT branch (light_tables.cpp:62-69):
// an identity 0..15 lava band on every shade; the darkest shade's band
// collapses to color 1 with index 0 staying black.
func applyLavaBright(tables *[16][256]uint8) {
	for t := range 16 {
		for k := range 16 {
			tables[t][k] = uint8(k) // std::iota (light_tables.cpp:64-65)
		}
	}

	tables[15][0] = 0 // light_tables.cpp:66

	for k := 1; k < 16; k++ {
		tables[15][k] = 1 // std::fill_n (light_tables.cpp:67)
	}
}

// makeLightFalloffs ports the LightFalloffs loop (light_tables.cpp:76-101):
// quadratic over-exposed falloff on NEST/CRYPT, linear lean falloff otherwise.
func makeLightFalloffs(levelType int) [16][128]uint8 {
	var falloffs [16][128]uint8

	for radius := range 16 {
		maxDistance := (radius + 1) * 8

		for distance := range 128 {
			falloffs[radius][distance] = lightFalloff(levelType, radius, maxDistance, distance)
		}
	}

	return falloffs
}

// lightFalloff ports one LightFalloffs cell (light_tables.cpp:83-100).
func lightFalloff(levelType, radius, maxDistance, distance int) uint8 {
	if distance > maxDistance {
		return 15 // light_tables.cpp:84-87
	}

	factor := float32(distance) / float32(maxDistance) // light_tables.cpp:89

	scaled := falloffScaled(levelType, radius, factor) // light_tables.cpp:91-98
	scaled += 0.5                                      // Round up (light_tables.cpp:99)

	return uint8(scaled) // light_tables.cpp:100
}

// falloffScaled ports the per-distance scaling branch (light_tables.cpp:91-98):
// quadratic falloff with over-exposure on NEST/CRYPT, linear lean falloff
// otherwise. float32 is used exactly as the upstream C++ float math.
func falloffScaled(levelType, radius int, factor float32) float32 {
	const (
		maxDarkness   = float32(15)
		maxBrightness = float32(0)
	)

	if levelType == dtypeNest || levelType == dtypeCrypt {
		// quardratic falloff with over exposure (light_tables.cpp:91-95)
		brightness := float32(radius) * float32(1.25)
		scaled := factor*factor*brightness + (maxDarkness - brightness)

		return max(maxBrightness, scaled) // std::max (light_tables.cpp:96)
	}

	// Leaner falloff (light_tables.cpp:97-98)
	return factor * maxDarkness
}

// flattenLightTables serializes one level-type table set, shade-major.
func flattenLightTables(levelType int) []byte {
	tables := makeLightTables(levelType)

	out := make([]byte, 0, len(tables)*256)
	for i := range tables {
		out = append(out, tables[i][:]...)
	}

	return out
}

// flattenLightFalloffs serializes one level-type falloff set, radius-major.
func flattenLightFalloffs(levelType int) []byte {
	falloffs := makeLightFalloffs(levelType)

	out := make([]byte, 0, len(falloffs)*128)
	for i := range falloffs {
		out = append(out, falloffs[i][:]...)
	}

	return out
}
