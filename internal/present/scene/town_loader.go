package scene

import (
	"encoding/binary"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// renumbered micro blocks the renderer reads.
type townPieces struct {
	cel    []byte
	frames []render.DunFrame
	micros render.DungeonMicros
}

// loadTownDunFiles reads the five fixed sector layers CreateTown loads
// (town.cpp:202-207); the map keys are the upstream path strings.
func loadTownDunFiles(stack *assets.MPQStack) (map[string][]byte, error) {
	files := make(map[string][]byte, len(sectorPaths)+1)
	for _, sector := range sectorPaths {
		data, ok := stack.ReadFile(sector.path)
		if !ok {
			return nil, fmt.Errorf("read %s: not found in archives", sector.path)
		}

		files[sector.path] = data
	}

	automap, ok := stack.ReadFile(levels.TownAutomap)
	if !ok {
		return nil, fmt.Errorf("read %s: not found in archives", levels.TownAutomap)
	}

	files[levels.TownAutomap] = automap

	return files, nil
}

// parseMegaTiles parses the town.til pMegaTiles table: count*4 little-endian
// micro tiles (diablo.cpp:1380-1385). The committed corpus confirms 376 mega
// tiles = 3008 bytes for town (internal/testdata/drlg/assets/town.til.csv
// header "376,4").
func parseMegaTiles(data []byte) ([]levels.MegaTile, error) {
	const megaWords = 4
	if len(data) == 0 || len(data)%(megaWords*2) != 0 {
		return nil, fmt.Errorf("%s: %d bytes is not a mega tile table", townTilPath, len(data))
	}

	count := len(data) / (megaWords * 2)
	tiles := make([]levels.MegaTile, count)
	for i := range tiles {
		off := i * megaWords * 2
		tiles[i] = levels.MegaTile{
			Micro1: binary.LittleEndian.Uint16(data[off:]),
			Micro2: binary.LittleEndian.Uint16(data[off+2:]),
			Micro3: binary.LittleEndian.Uint16(data[off+4:]),
			Micro4: binary.LittleEndian.Uint16(data[off+6:]),
		}
	}

	return tiles, nil
}

// townPieces is decodeTownPieces' result: the re-encoded CEL and the

// decodeTownPieces runs the MIN decode + re-encode chain:
// LoadSOLData -> LoadMinPieces -> DecodeDungeonPieces (town, 16 blocks) ->
// ReencodeDungeonCels + ComputeCelBlockAdjustments +
// RenumberLevelCelBlocks (diablo.cpp:3421 SetDungeonMicros).
func decodeTownPieces(celData, minData, solData []byte) (townPieces, error) {
	sol, err := render.LoadSOLData(solData)
	if err != nil {
		return townPieces{}, fmt.Errorf("LoadSOLData: %w", err)
	}

	sol = render.ApplySOLHandPatches(sol, render.LevelTown)

	pieces, err := render.LoadMinPieces(minData)
	if err != nil {
		return townPieces{}, fmt.Errorf("LoadMinPieces: %w", err)
	}

	decoded, err := render.DecodeDungeonPieces(pieces, render.LevelTown, sol)
	if err != nil {
		return townPieces{}, fmt.Errorf("DecodeDungeonPieces: %w", err)
	}

	reencoded := render.ReencodeDungeonCels(celData, decoded.Frames)
	render.RenumberLevelCelBlocks(decoded.Micros, render.ComputeCelBlockAdjustments(decoded.Frames))

	return townPieces{cel: reencoded, frames: decoded.Frames, micros: decoded.Micros}, nil
}
