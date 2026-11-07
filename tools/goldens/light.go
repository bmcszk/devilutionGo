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

// lightArtifact is one dvg_probe `light` dump: the 112x112 dLight grid as
// dumped right after CreateDungeon, x-major, so cell (x, y) sits at
// x*height+y — the order the probe copies out of the C array
// (levels/dun_tile_data.hpp:105).
type lightArtifact struct {
	Width  int
	Height int
	DLight []uint8
}

// runLight converts every `light` artifact in artifactDir into the committed
// dLight golden CSV under corpusDir.
func runLight(paths []string) error {
	artifactDir, corpusDir := paths[0], paths[1]

	bases, err := lightArtifactBases(artifactDir)
	if err != nil {
		return err
	}

	for _, base := range bases {
		if err := convertLightArtifact(artifactDir, corpusDir, base); err != nil {
			return err
		}
	}

	slog.Info("goldens light done", "files", len(bases), "corpus", corpusDir)

	return nil
}

// convertLightArtifact parses one probe dump and writes its golden CSV.
func convertLightArtifact(artifactDir, corpusDir, base string) error {
	data, err := os.ReadFile(filepath.Join(artifactDir, base+".bin"))
	if err != nil {
		return fmt.Errorf("read %s.bin: %w", base, err)
	}

	grid, err := parseLightArtifact(data)
	if err != nil {
		return fmt.Errorf("parse %s: %w", base, err)
	}

	if err := os.WriteFile(filepath.Join(corpusDir, base+".csv"), encodeLightCSV(grid), 0o644); err != nil {
		return fmt.Errorf("write %s.csv: %w", base, err)
	}

	return nil
}

// lightArtifactBases lists the artifact base names ("l17-light-main-2588")
// that hold a dLight dump, sorted for deterministic output.
func lightArtifactBases(artifactDir string) ([]string, error) {
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

// parseLightArtifact decodes one probe dump: two little-endian uint16
// dimensions, then width*height uint8 dLight shades.
func parseLightArtifact(data []byte) (*lightArtifact, error) {
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
	if want := headerSize + cellCount; len(data) != want {
		return nil, fmt.Errorf("length %d, want %d for %dx%d", len(data), want, width, height)
	}

	return &lightArtifact{
		Width:  width,
		Height: height,
		DLight: slices.Clone(data[headerSize:]),
	}, nil
}

// encodeLightCSV writes the golden CSV: a "width,height" record, then the
// dLight shade row, x-major.
func encodeLightCSV(grid *lightArtifact) []byte {
	var out bytes.Buffer

	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{strconv.Itoa(grid.Width), strconv.Itoa(grid.Height)})
	_ = writer.Write(encodeUint8Row(grid.DLight))
	writer.Flush()

	return out.Bytes()
}
