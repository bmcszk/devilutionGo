package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
)

// townArtifact is one dvg_probe `town` dump: the world-tile layers CreateTown
// writes (dPiece, dSpecial), the one it leaves zeroed (dTransVal), and the 40x40
// tile layer the sector and automap loads fill. Every layer is x-major, so cell
// (x, y) sits at x*height+y — the order the probe copies out of the C arrays.
type townArtifact struct {
	Width     int
	Height    int
	TileW     int
	TileH     int
	DPiece    []uint16
	DSpecial  []int8
	DTransVal []int8
	Tiles     []uint8
}

// runTown converts every `town` artifact in artifactDir into the committed town
// golden CSV under corpusDir.
func runTown(artifactDir, corpusDir string) error {
	bases, err := townArtifactBases(artifactDir)
	if err != nil {
		return err
	}

	for _, base := range bases {
		if err := convertTownArtifact(artifactDir, corpusDir, base); err != nil {
			return err
		}
	}

	return nil
}

// convertTownArtifact parses one probe dump, checks the invariants the inventory
// documents, and writes its golden CSV.
func convertTownArtifact(artifactDir, corpusDir, base string) error {
	data, err := os.ReadFile(filepath.Join(artifactDir, base+".bin"))
	if err != nil {
		return fmt.Errorf("read %s.bin: %w", base, err)
	}

	grids, err := parseTownArtifact(data)
	if err != nil {
		return fmt.Errorf("parse %s: %w", base, err)
	}

	if err := checkTownTranslucency(grids); err != nil {
		return fmt.Errorf("parse %s: %w", base, err)
	}

	if err := os.WriteFile(filepath.Join(corpusDir, base+".csv"), encodeTownCSV(grids), 0o644); err != nil {
		return fmt.Errorf("write %s.csv: %w", base, err)
	}

	return nil
}

// checkTownTranslucency pins the inventory's claim that no town path writes
// dTransVal: InitGlobals zeroes it (gendung.cpp:384-389) and CreateTown never
// touches it again.
func checkTownTranslucency(grids *townArtifact) error {
	for index, value := range grids.DTransVal {
		if value != 0 {
			return fmt.Errorf("dTransVal[%d][%d] = %d, want 0", index/grids.Height, index%grids.Height, value)
		}
	}

	return nil
}

// parseTownArtifact decodes one probe dump: four little-endian uint16 dimensions
// (world-tile, then tile-layer), width*height uint16 dPiece, width*height int8
// dSpecial, width*height int8 dTransVal, then tileW*tileH uint8 tiles.
func parseTownArtifact(data []byte) (*townArtifact, error) {
	const headerSize = 8
	if len(data) < headerSize {
		return nil, errors.New("short header")
	}

	grids := &townArtifact{
		Width:  int(binary.LittleEndian.Uint16(data[0:2])),
		Height: int(binary.LittleEndian.Uint16(data[2:4])),
		TileW:  int(binary.LittleEndian.Uint16(data[4:6])),
		TileH:  int(binary.LittleEndian.Uint16(data[6:8])),
	}
	if grids.Width <= 0 || grids.Height <= 0 || grids.TileW <= 0 || grids.TileH <= 0 {
		return nil, fmt.Errorf("bad dimensions %dx%d tiles %dx%d", grids.Width, grids.Height, grids.TileW, grids.TileH)
	}

	cells := grids.Width * grids.Height
	want := headerSize + cells*4 + grids.TileW*grids.TileH
	if len(data) != want {
		return nil, fmt.Errorf("length %d, want %d for %dx%d", len(data), want, grids.Width, grids.Height)
	}

	grids.DPiece = decodeUint16Cells(data[headerSize:], cells)
	specialStart := headerSize + cells*2
	grids.DSpecial = decodeInt8Cells(data[specialStart:], cells)
	transValStart := specialStart + cells
	grids.DTransVal = decodeInt8Cells(data[transValStart:], cells)
	grids.Tiles = slices.Clone(data[transValStart+cells:])

	return grids, nil
}

func decodeUint16Cells(data []byte, cells int) []uint16 {
	values := make([]uint16, cells)
	for i := range values {
		values[i] = binary.LittleEndian.Uint16(data[i*2:])
	}
	return values
}

func decodeInt8Cells(data []byte, cells int) []int8 {
	values := make([]int8, cells)
	for i := range values {
		values[i] = int8(data[i])
	}
	return values
}

// encodeTownCSV writes the golden CSV in the dPiece corpus shape — a
// "width,height" record followed by one x-major row per layer — and appends the
// 40x40 tile layer as its own header plus row, because town pins both the
// mega-tile expansion and the automap tile layer CreateTown patches.
func encodeTownCSV(grids *townArtifact) []byte {
	var out bytes.Buffer

	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{strconv.Itoa(grids.Width), strconv.Itoa(grids.Height)})
	_ = writer.Write(encodeUint16Row(grids.DPiece))
	_ = writer.Write(encodeInt8Row(grids.DSpecial))
	_ = writer.Write([]string{strconv.Itoa(grids.TileW), strconv.Itoa(grids.TileH)})
	_ = writer.Write(encodeUint8Row(grids.Tiles))
	writer.Flush()

	return out.Bytes()
}

func encodeUint8Row(cells []uint8) []string {
	row := make([]string, len(cells))
	for i, value := range cells {
		row[i] = strconv.Itoa(int(value))
	}
	return row
}

// townArtifactBases lists the artifact base names ("l0-town-main-2588"), sorted
// for deterministic output.
func townArtifactBases(artifactDir string) ([]string, error) {
	entries, err := os.ReadDir(artifactDir)
	if err != nil {
		return nil, fmt.Errorf("read artifact dir: %w", err)
	}

	bases := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || filepath.Ext(name) != ".bin" {
			continue
		}
		bases = append(bases, name[:len(name)-len(".bin")])
	}

	if len(bases) == 0 {
		return nil, fmt.Errorf("no .bin artifacts in %s", artifactDir)
	}

	slices.Sort(bases)

	return bases, nil
}
