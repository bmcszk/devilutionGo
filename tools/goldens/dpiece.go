package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// dpieceArtifact is one dvg_probe `dpiece` dump: the 112x112 dPiece
// (mega-tile expansion) and dSpecial layers, both x-major, so cell (x, y)
// sits at x*height+y — the order the probe copies out of the C arrays
// (dun_tile_data.hpp:99,136).
type dpieceArtifact struct {
	Width    int
	Height   int
	DPiece   []uint16
	DSpecial []int8
}

// runDpiece converts every `dpiece` artifact in artifactDir into the committed
// dPiece/dSpecial golden CSV under corpusDir.
func runDpiece(paths []string) error {
	artifactDir, corpusDir := paths[0], paths[1]

	bases, err := dpieceArtifactBases(artifactDir)
	if err != nil {
		return err
	}

	for _, base := range bases {
		if err := convertDpieceArtifact(artifactDir, corpusDir, base); err != nil {
			return err
		}
	}

	slog.Info("goldens dpiece done", "files", len(bases), "corpus", corpusDir)

	return nil
}

// convertDpieceArtifact parses one probe dump and writes its golden CSV.
func convertDpieceArtifact(artifactDir, corpusDir, base string) error {
	data, err := os.ReadFile(filepath.Join(artifactDir, base+".bin"))
	if err != nil {
		return fmt.Errorf("read %s.bin: %w", base, err)
	}

	grids, err := parseDpieceArtifact(data)
	if err != nil {
		return fmt.Errorf("parse %s: %w", base, err)
	}

	if err := os.WriteFile(filepath.Join(corpusDir, base+".csv"), encodeDpieceCSV(grids), 0o644); err != nil {
		return fmt.Errorf("write %s.csv: %w", base, err)
	}

	return nil
}

// dpieceArtifactBases lists the artifact base names ("l5-main-2588") that hold
// a dPiece dump, sorted for deterministic output.
func dpieceArtifactBases(artifactDir string) ([]string, error) {
	entries, err := os.ReadDir(artifactDir)
	if err != nil {
		return nil, fmt.Errorf("read artifact dir: %w", err)
	}

	bases := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".bin") {
			continue
		}
		bases = append(bases, strings.TrimSuffix(name, ".bin"))
	}

	if len(bases) == 0 {
		return nil, fmt.Errorf("no .bin artifacts in %s", artifactDir)
	}

	slices.Sort(bases)

	return bases, nil
}

// parseDpieceArtifact decodes one probe dump: two little-endian uint16
// dimensions, then width*height uint16 dPiece cells, then width*height int8
// dSpecial cells.
func parseDpieceArtifact(data []byte) (*dpieceArtifact, error) {
	const headerSize = 4
	if len(data) < headerSize {
		return nil, errors.New("short header")
	}

	width := int(binary.LittleEndian.Uint16(data[0:2]))
	height := int(binary.LittleEndian.Uint16(data[2:4]))
	cellCount := width * height
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("bad dimensions %dx%d", width, height)
	}
	if want := headerSize + cellCount*3; len(data) != want {
		return nil, fmt.Errorf("length %d, want %d for %dx%d", len(data), want, width, height)
	}

	grids := &dpieceArtifact{
		Width:    width,
		Height:   height,
		DPiece:   make([]uint16, cellCount),
		DSpecial: make([]int8, cellCount),
	}

	pieceBytes := data[headerSize : headerSize+cellCount*2]
	for i := range grids.DPiece {
		grids.DPiece[i] = binary.LittleEndian.Uint16(pieceBytes[i*2:])
	}

	specialBytes := data[headerSize+cellCount*2:]
	for i := range grids.DSpecial {
		grids.DSpecial[i] = int8(specialBytes[i])
	}

	return grids, nil
}

// encodeDpieceCSV writes the golden CSV: a "width,height" record, the dPiece
// row (x-major), and the dSpecial row only when dSpecial carries a non-zero
// cell (the generator-side door/arch markers; cathedral, caves and hell leave
// it empty at generation time).
func encodeDpieceCSV(grids *dpieceArtifact) []byte {
	var out bytes.Buffer

	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{strconv.Itoa(grids.Width), strconv.Itoa(grids.Height)})
	_ = writer.Write(encodeUint16Row(grids.DPiece))
	if slices.ContainsFunc(grids.DSpecial, func(value int8) bool { return value != 0 }) {
		_ = writer.Write(encodeInt8Row(grids.DSpecial))
	}
	writer.Flush()

	return out.Bytes()
}

func encodeUint16Row(cells []uint16) []string {
	row := make([]string, len(cells))
	for i, value := range cells {
		row[i] = strconv.Itoa(int(value))
	}
	return row
}

func encodeInt8Row(cells []int8) []string {
	row := make([]string, len(cells))
	for i, value := range cells {
		row[i] = strconv.Itoa(int(value))
	}
	return row
}
