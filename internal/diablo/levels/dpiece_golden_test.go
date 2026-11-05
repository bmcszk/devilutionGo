package levels_test

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// pieceDir and assetDir are the committed dPiece corpus and the mega-tile
// tables the Pass3 expansion reads (internal/testdata/drlg/README.md).
const (
	pieceDir = drlgDir + "/dpiece"
	assetDir = drlgDir + "/assets"
)

// pieceGolden is one parsed dPiece fixture: the world-tile grid, the optional
// dSpecial grid, and whether the fixture carries that second record at all
// (levels with an all-zero dSpecial omit it, README "dPiece goldens").
type pieceGolden struct {
	piece      [levels.MAXDUNX][levels.MAXDUNY]uint16
	special    [levels.MAXDUNX][levels.MAXDUNY]int8
	hasSpecial bool
}

// pieceGoldenPaths lists the committed dPiece fixtures.
func pieceGoldenPaths(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(pieceDir, "*.csv"))
	if err != nil {
		t.Fatalf("glob dPiece goldens: %v", err)
	}

	return paths
}

// loadPieceGoldens parses one seed's dPiece fixtures, keyed "l<level>-<entry>".
func loadPieceGoldens(t *testing.T, seed uint32) map[string]*pieceGolden {
	t.Helper()

	suffix := fmt.Sprintf("-%d.csv", seed)
	goldens := make(map[string]*pieceGolden, len(dpieceVariants))
	for _, path := range pieceGoldenPaths(t) {
		base := filepath.Base(path)
		if !strings.HasSuffix(base, suffix) {
			continue
		}

		key := strings.Replace(strings.TrimSuffix(base, suffix), "-dpiece-", "-", 1)
		goldens[key] = loadPieceGolden(t, path)
	}

	return goldens
}

// loadPieceGolden parses one fixture: "112,112", then 112*112 dPiece values and,
// where the level writes any, 112*112 dSpecial values, both x-major.
func loadPieceGolden(t *testing.T, path string) *pieceGolden {
	t.Helper()

	records := readCSVRecords(t, path)
	if want := []int{2, 3}; len(records) != want[0] && len(records) != want[1] {
		t.Fatalf("%s: %d records, want 2 or 3", path, len(records))
	}

	checkPieceHeader(t, path, records[0])

	golden := &pieceGolden{}
	fillPieceRecord(t, path, records[1], &golden.piece)
	if len(records) == 3 {
		golden.hasSpecial = true
		fillSpecialRecord(t, path, records[2], &golden.special)
	}

	return golden
}

// pieceMegaTiles parses assets/l<family>.til.csv into pMegaTiles entries, one
// family at a time (levels/dun_tile_data.hpp:56-61).
func pieceMegaTiles(t *testing.T) map[string][]levels.MegaTile {
	t.Helper()

	families := []string{"l1", "l2", "l3", "l4"}
	tiles := make(map[string][]levels.MegaTile, len(families))
	for _, family := range families {
		tiles[family] = loadMegaTiles(t, family)
	}

	return tiles
}

// loadMegaTiles parses one family's mega-tile table: "count,4", then
// count*4 micro tiles in file order.
func loadMegaTiles(t *testing.T, family string) []levels.MegaTile {
	t.Helper()

	path := filepath.Join(assetDir, family+".til.csv")
	records := readCSVRecords(t, path)
	if len(records) != 2 {
		t.Fatalf("%s: %d records, want 2", path, len(records))
	}

	count, cols := checkTableHeader(t, path, records[0], 4)
	values := parseUint16Record(t, path, records[1])
	if len(values) != count*cols {
		t.Fatalf("%s: %d values, want %d", path, len(values), count*cols)
	}

	tiles := make([]levels.MegaTile, count)
	for tile := range tiles {
		micros := values[tile*cols:]
		tiles[tile] = levels.MegaTile{Micro1: micros[0], Micro2: micros[1], Micro3: micros[2], Micro4: micros[3]}
	}

	return tiles
}

// checkPieceHeader pins the "MAXDUNX,MAXDUNY" record the corpus starts with.
func checkPieceHeader(t *testing.T, path string, record []string) {
	t.Helper()

	if got := parseUint16Record(t, path, record); len(got) != 2 ||
		got[0] != levels.MAXDUNX || got[1] != levels.MAXDUNY {
		t.Fatalf("%s: header %v, want %d,%d", path, record, levels.MAXDUNX, levels.MAXDUNY)
	}
}

// checkTableHeader pins a "count,cols" asset record and returns it.
func checkTableHeader(t *testing.T, path string, record []string, wantCols int) (count, cols int) {
	t.Helper()

	got := parseUint16Record(t, path, record)
	if len(got) != 2 || int(got[1]) != wantCols {
		t.Fatalf("%s: header %v, want <count>,%d", path, record, wantCols)
	}
	if got[0] == 0 {
		t.Fatalf("%s: empty table", path)
	}

	return int(got[0]), int(got[1])
}

// fillPieceRecord reads one dPiece record into grid, index x*MAXDUNY+y.
func fillPieceRecord(t *testing.T, path string, record []string, grid *[levels.MAXDUNX][levels.MAXDUNY]uint16) {
	t.Helper()

	values := parseUint16Record(t, path, record)
	checkPieceRecordLen(t, path, len(values))
	for index, value := range values {
		grid[index/levels.MAXDUNY][index%levels.MAXDUNY] = value
	}
}

// fillSpecialRecord reads one dSpecial record into grid, index x*MAXDUNY+y.
func fillSpecialRecord(t *testing.T, path string, record []string, grid *[levels.MAXDUNX][levels.MAXDUNY]int8) {
	t.Helper()

	checkPieceRecordLen(t, path, len(record))
	for index, field := range record {
		value, err := strconv.ParseInt(field, 10, 8)
		if err != nil {
			t.Fatalf("%s: dSpecial value %d %q: %v", path, index, field, err)
		}
		grid[index/levels.MAXDUNY][index%levels.MAXDUNY] = int8(value)
	}
}

// checkPieceRecordLen pins one record to the full world-tile grid.
func checkPieceRecordLen(t *testing.T, path string, got int) {
	t.Helper()

	if want := levels.MAXDUNX * levels.MAXDUNY; got != want {
		t.Fatalf("%s: %d values, want %d", path, got, want)
	}
}

// parseUint16Record parses one record of uint16 CSV fields.
func parseUint16Record(t *testing.T, path string, record []string) []uint16 {
	t.Helper()

	values := make([]uint16, len(record))
	for i, field := range record {
		value, err := strconv.ParseUint(field, 10, 16)
		if err != nil {
			t.Fatalf("%s: value %d %q: %v", path, i, field, err)
		}
		values[i] = uint16(value)
	}

	return values
}

// readCSVRecords parses a whole fixture, keeping ragged records so the record
// checks above can report the real shape.
func readCSVRecords(t *testing.T, path string) [][]string {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	return records
}

// comparePieces reports the first divergence of the generated grid from the
// fixture, dPiece first and then dSpecial where the fixture carries it.
func comparePieces(t *testing.T, got levels.PieceGrid, golden *pieceGolden) {
	t.Helper()

	checkPieceLayer(t, got.Piece, golden.piece)
	if golden.hasSpecial {
		checkSpecialLayer(t, got.Special, golden.special)
	}
}

// checkPieceLayer compares two dPiece grids index by index.
func checkPieceLayer(t *testing.T, got, want [levels.MAXDUNX][levels.MAXDUNY]uint16) {
	t.Helper()

	for x := range got {
		for y := range got[x] {
			if got[x][y] != want[x][y] {
				t.Fatalf("dPiece[%d][%d] = %d, want %d", x, y, got[x][y], want[x][y])
			}
		}
	}
}

// checkSpecialLayer compares two dSpecial grids index by index.
func checkSpecialLayer(t *testing.T, got, want [levels.MAXDUNX][levels.MAXDUNY]int8) {
	t.Helper()

	for x := range got {
		for y := range got[x] {
			if got[x][y] != want[x][y] {
				t.Fatalf("dSpecial[%d][%d] = %d, want %d", x, y, got[x][y], want[x][y])
			}
		}
	}
}
