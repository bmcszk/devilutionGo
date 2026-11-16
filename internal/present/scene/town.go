package scene

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// Retail record paths of the town's render inputs, upstream spellings
// (diablo.cpp:1371-1387 LoadLvlGFX DTYPE_TOWN, dun_tile_data.cpp:103-106
// LoadLevelSOLData + :72-77 LoadMinData, town.cpp:202-207).
const (
	townCelPath = `levels\towndata\town.cel`
	townTilPath = `levels\towndata\town.til`
	townMinPath = `levels\towndata\town.min`
	townSolPath = `levels\towndata\town.sol`
	townPalPath = `levels\towndata\town.pal`
)

// sectorPaths lists the fixed sector .dun resources with their placements
// (levels/town.go TownSector* constants, town.cpp:202-205).
var sectorPaths = []struct {
	path string
	at   levels.Point
}{
	{levels.TownSector1, levels.Point{X: 46, Y: 46}},
	{levels.TownSector2, levels.Point{X: 46, Y: 0}},
	{levels.TownSector3, levels.Point{X: 0, Y: 46}},
	{levels.TownSector4, levels.Point{X: 0, Y: 0}},
}

// TownScene is one composed town frame: the index surface plus the palette
// state it presents through.
type TownScene struct {
	Surface render.Surface
	State   render.PaletteState
}

// townRawRecords holds the raw archive bytes one compose reads.
type townRawRecords struct {
	palData []byte
	celData []byte
	minData []byte
	tilData []byte
	solData []byte
}

// townRecords bundles the loaded render inputs of one town compose.
type townRecords struct {
	pal       *assets.Palette
	cel       []byte
	micros    render.DungeonMicros
	sol       render.SOLData
	megaTiles []levels.MegaTile
	files     map[string][]byte
}

// ArchivesPresent reports whether any directory in dirs holds a readable
// retail archive; callers skip instead of failing when it is false
// (mpq_smoke_test.go findAssetArchives pattern).
func ArchivesPresent(dirs []string) bool {
	for _, dir := range dirs {
		for _, name := range []string{"DIABDAT.MPQ", "hellfire.mpq"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return true
			}
		}
	}

	return false
}

// Archive preference mirrors LoadLvlGFX: the hellfire nlevels variants win
// when installed (diablo.cpp:1374-1383, dun_tile_data.cpp:72-77); the stack
// resolves per record, and hellfire.mpq holds no levels\towndata members, so
// the diabdat records serve where nlevels lacks them.
func ComposeTownFrame(dirs []string) (TownScene, error) {
	stack := assets.NewMPQStack()
	defer func() { _ = stack.Close() }()

	stack.Load(dirs, "hellfire", assets.HellfireMpqPriority)
	stack.Load(dirs, "DIABDAT", assets.MainMpqPriority, ".MPQ")
	stack.Load(dirs, "devilutionx", assets.DevilutionXMpqPriority)

	town, err := loadTownRecords(stack)
	if err != nil {
		return TownScene{}, err
	}

	grid := buildTownGrid(town.files, town.megaTiles)

	state := render.NewPaletteState(*town.pal, light.MakeLightTables(light.DTYPE_TOWN), assets.BlendingFor(0))

	world, view := townWorld(grid, town.micros, town.sol)

	viewData := &townWorldView{grid: grid, micros: town.micros, sol: town.sol, light: townLight()}
	surface := renderTownFrame(world, view, town.cel, state, viewData)

	return TownScene{Surface: surface, State: state}, nil
}

// loadTownRecords reads every record ComposeTownFrame needs and decodes the
// MIN + SOL chains; the first missing archive record names itself.
func loadTownRecords(stack *assets.MPQStack) (townRecords, error) {
	raw, files, err := readTownFiles(stack)
	if err != nil {
		return townRecords{}, err
	}

	pal, err := assets.LoadPaletteFromBytes(raw.palData)
	if err != nil {
		return townRecords{}, fmt.Errorf("decode %s: %w", townPalPath, err)
	}

	megaTiles, err := parseMegaTiles(raw.tilData)
	if err != nil {
		return townRecords{}, err
	}

	pieces, err := decodeTownPieces(raw.celData, raw.minData, raw.solData)
	if err != nil {
		return townRecords{}, err
	}

	sol, err := render.LoadSOLData(raw.solData)
	if err != nil {
		return townRecords{}, fmt.Errorf("LoadSOLData: %w", err)
	}

	return townRecords{
		pal:       pal,
		cel:       pieces.cel,
		micros:    pieces.micros,
		sol:       render.ApplySOLHandPatches(sol, render.LevelTown),
		megaTiles: megaTiles,
		files:     files,
	}, nil
}

// readTownFiles reads the raw records plus the sector layers.
func readTownFiles(stack *assets.MPQStack) (townRawRecords, map[string][]byte, error) {
	var raw townRawRecords

	for _, record := range []struct {
		path string
		dst  *[]byte
	}{
		{townPalPath, &raw.palData},
		{townCelPath, &raw.celData},
		{townMinPath, &raw.minData},
		{townTilPath, &raw.tilData},
		{townSolPath, &raw.solData},
	} {
		data, err := readRecord(stack, record.path)
		if err != nil {
			return raw, nil, err
		}

		*record.dst = data
	}

	files, err := loadTownDunFiles(stack)
	if err != nil {
		return raw, nil, err
	}

	return raw, files, nil
}

// readRecord fetches one archive member by upstream path, naming the record
// when no archive serves it.
func readRecord(stack *assets.MPQStack, path string) ([]byte, error) {
	data, ok := stack.ReadFile(path)
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("read %s: not found in archives", path)
	}

	return data, nil
}

// buildTownGrid runs CreateTown on a fresh town dungeon with the real light
// init (town.cpp:359-387, gendung.cpp:301-307 InitLight's town-0 default).
func buildTownGrid(files map[string][]byte, megaTiles []levels.MegaTile) levels.PieceGrid {
	dungeon := levels.NewDungeon(levels.Town, 0)
	dungeon.InitLight()

	return levels.CreateTown(dungeon, levels.TownOptions{
		Entry:     levels.EntryMain,
		Files:     files,
		MegaTiles: megaTiles,
		Quests:    levels.Quests{PWATERFountain: true},
	})
}
