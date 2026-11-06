package testsupport

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"strconv"
)

// EncodeDunTilesCSV writes the committed tile-layer fixture format the fixed
// `.dun` set pieces use: "width,height" then width*height row-major tile ids.
// It is the encoder twin of DecodeDunTilesCSV, so a refresh cannot drift from
// the bytes the loaders read.
func EncodeDunTilesCSV(width, height int, tiles []uint16) []byte {
	var out bytes.Buffer

	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{strconv.Itoa(width), strconv.Itoa(height)})

	fields := make([]string, len(tiles))
	for i, tile := range tiles {
		fields[i] = strconv.Itoa(int(tile))
	}
	_ = writer.Write(fields)
	writer.Flush()

	return out.Bytes()
}

// DecodeDunTilesCSV rebuilds the `.dun` prefix a loader reads from one
// tile-layer fixture: two little-endian dimensions followed by the tile layer as
// little-endian uint16 values (levels/gendung.cpp:576-580, town.cpp:26-31).
func DecodeDunTilesCSV(csvData []byte) ([]byte, error) {
	reader := csv.NewReader(bytes.NewReader(csvData))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse tile-layer CSV: %w", err)
	}
	if len(records) != 2 {
		return nil, fmt.Errorf("tile-layer CSV holds %d records, want 2", len(records))
	}

	width, height, err := dunTilesDims(records[0])
	if err != nil {
		return nil, err
	}
	if got, want := len(records[1]), width*height; got != want {
		return nil, fmt.Errorf("tile layer holds %d tiles, want %d", got, want)
	}

	data := make([]byte, 4+2*width*height)
	binary.LittleEndian.PutUint16(data[0:], uint16(width))
	binary.LittleEndian.PutUint16(data[2:], uint16(height))

	for i, field := range records[1] {
		tile, err := strconv.ParseUint(field, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("tile %d %q: %w", i, field, err)
		}
		binary.LittleEndian.PutUint16(data[4+2*i:], uint16(tile))
	}

	return data, nil
}

// dunTilesDims reads the "width,height" header record.
func dunTilesDims(record []string) (width, height int, err error) {
	if len(record) != 2 {
		return 0, 0, fmt.Errorf("tile-layer header %v, want width,height", record)
	}

	values := make([]int, 2)
	for i, field := range record {
		parsed, err := strconv.Atoi(field)
		if err != nil {
			return 0, 0, fmt.Errorf("header dimension %q: %w", field, err)
		}
		values[i] = parsed
	}

	if values[0] <= 0 || values[1] <= 0 {
		return 0, 0, fmt.Errorf("header dimensions %v must be positive", values)
	}

	return values[0], values[1], nil
}
