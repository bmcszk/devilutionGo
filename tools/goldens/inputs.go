package main

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

// townInputFiles are the fixed town `.dun` files CreateTown loads
// (levels/town.cpp:202-207), by artifact base name.
var townInputFiles = []string{"sector1s", "sector2s", "sector3s", "sector4s", "automap"}

// dunLayer is the header and tile layer of one `.dun` file.
type dunLayer struct {
	width, height int
	tiles         []uint16
}

// runTownInputs converts the five fixed town `.dun` files of an offline asset
// extraction into the committed tile-layer fixtures under
// <corpus-dir>/inputs/<base>.csv. The encoder is
// testsupport.EncodeDunTilesCSV, so a refresh cannot drift from the format the
// parity test decodes.
func runTownInputs(paths []string) error {
	rawDir, corpusDir := paths[0], paths[1]
	outDir := filepath.Join(corpusDir, "inputs")

	for _, base := range townInputFiles {
		if err := convertTownInput(rawDir, outDir, base); err != nil {
			return err
		}
	}

	slog.Info("goldens town-inputs done", "files", len(townInputFiles), "corpus", outDir)

	return nil
}

// convertTownInput writes one raw `.dun` file as a tile-layer fixture.
func convertTownInput(rawDir, outDir, base string) error {
	data, err := os.ReadFile(filepath.Join(rawDir, base+".dun"))
	if err != nil {
		return fmt.Errorf("read %s.dun: %w", base, err)
	}

	layer, err := parseDunTiles(data)
	if err != nil {
		return fmt.Errorf("parse %s.dun: %w", base, err)
	}

	csvData := testsupport.EncodeDunTilesCSV(layer.width, layer.height, layer.tiles)

	path := filepath.Join(outDir, base+".csv")
	if err := os.WriteFile(path, csvData, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

// parseDunTiles reads the dimensions and the tile layer of one `.dun` file.
func parseDunTiles(data []byte) (dunLayer, error) {
	if len(data) < 4 {
		return dunLayer{}, fmt.Errorf("%d bytes, want at least 4", len(data))
	}

	layer := dunLayer{
		width:  int(binary.LittleEndian.Uint16(data[0:])),
		height: int(binary.LittleEndian.Uint16(data[2:])),
	}

	count := layer.width * layer.height
	if len(data) < 4+2*count {
		return dunLayer{}, fmt.Errorf("%d bytes, want %d for %dx%d", len(data), 4+2*count, layer.width, layer.height)
	}

	layer.tiles = make([]uint16, count)
	for i := range layer.tiles {
		layer.tiles[i] = binary.LittleEndian.Uint16(data[4+2*i:])
	}

	return layer, nil
}
