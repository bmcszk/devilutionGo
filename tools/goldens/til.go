package main

import (
	"bytes"
	"crypto/sha256"
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

// megaTileCols is the tile width of a `.til` member: four uint16 micro tiles
// per mega tile (levels/dun_tile_data.hpp:56-61).
const megaTileCols = 4

// runTil converts every raw `.til` member in rawDir into the committed
// assets/<name>.tal.csv table and reports the sha256 of the bytes it
// re-encodes to, which is what assets/manifest.txt pins.
func runTil(paths []string) error {
	rawDir, corpusDir := paths[0], paths[1]

	bases, err := tilBases(rawDir)
	if err != nil {
		return err
	}

	for _, base := range bases {
		hash, err := convertTil(rawDir, corpusDir, base)
		if err != nil {
			return err
		}
		slog.Info("goldens til converted", "file", base+".til.csv", "sha256", hash)
	}

	return nil
}

// tilBases lists the raw member names ("l5") in rawDir, sorted.
func tilBases(rawDir string) ([]string, error) {
	entries, err := os.ReadDir(rawDir)
	if err != nil {
		return nil, fmt.Errorf("read raw dir: %w", err)
	}

	bases := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".til") {
			continue
		}
		bases = append(bases, strings.TrimSuffix(name, ".til"))
	}

	if len(bases) == 0 {
		return nil, fmt.Errorf("no .til members in %s", rawDir)
	}

	slices.Sort(bases)

	return bases, nil
}

// convertTil decodes one headerless `.til` member, writes its CSV and returns
// the sha256 of the member bytes.
func convertTil(rawDir, corpusDir, base string) (string, error) {
	data, err := os.ReadFile(filepath.Join(rawDir, base+".til"))
	if err != nil {
		return "", fmt.Errorf("read %s.til: %w", base, err)
	}

	values, err := parseTilMember(data)
	if err != nil {
		return "", fmt.Errorf("parse %s.til: %w", base, err)
	}

	path := filepath.Join(corpusDir, base+".til.csv")
	if err := os.WriteFile(path, encodeTilCSV(values), 0o644); err != nil {
		return "", fmt.Errorf("write %s.til.csv: %w", path, err)
	}

	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// parseTilMember decodes a `.til` member: count*cols uint16 LE micro tiles,
// no header (the CSV's "count,cols" record is derived from the length).
func parseTilMember(data []byte) ([]uint16, error) {
	if len(data) == 0 || len(data)%(megaTileCols*2) != 0 {
		return nil, errors.New("length is not a whole number of mega tiles")
	}

	values := make([]uint16, len(data)/2)
	for i := range values {
		values[i] = binary.LittleEndian.Uint16(data[i*2:])
	}

	return values, nil
}

// encodeTilCSV writes the committed table: a "count,cols" record, then
// count*cols micro tiles in file order.
func encodeTilCSV(values []uint16) []byte {
	var out bytes.Buffer

	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{strconv.Itoa(len(values) / megaTileCols), strconv.Itoa(megaTileCols)})
	_ = writer.Write(encodeUint16Row(values))
	writer.Flush()

	return out.Bytes()
}
