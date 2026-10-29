package entities_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
)

// decodeRowOffset spaces the fixture rows far apart so they never collide.
const decodeRowOffset = 7

type decodeRow struct {
	x, y      int
	raw       int8
	corpseID  int8
	direction int
}

// TestCorpses_DecodeFixtureParity replays the committed decode fixture: raw
// dCorpse bytes map to corpse id bits 0-4 and direction bits 5-7
// (scrollrt.cpp:802-804).
func TestCorpses_DecodeFixtureParity(t *testing.T) {
	// given the committed decode rows
	rows := loadDecodeRows(t)

	// when each byte sits on its tile
	var corpses entities.Corpses
	for _, row := range rows {
		corpses.Tiles[row.y][row.x] = row.raw
	}

	// then every tile decodes to the fixture's id and direction
	for _, row := range rows {
		assertDecoded(t, &corpses, row)
	}
}

// TestCorpses_DecodeFixtureZeroIDDraws pins the deviation guard on the same
// fixture rows: zero id bits means no draw call (scrollrt.cpp:802 would index -1).
func TestCorpses_DecodeFixtureZeroIDDraws(t *testing.T) {
	// given the committed decode rows whose id bits are zero
	rows := zeroIDRows(loadDecodeRows(t))

	// when each of them is rendered through the corpse draw seam
	draws := drawsFor(rows)

	// then the fixture covered that case and nothing was drawn
	if len(rows) == 0 {
		t.Fatal("decode fixture has no zero-id row")
	}

	if draws != 0 {
		t.Fatalf("zero-id rows drew %d call(s), want 0", draws)
	}
}

// loadDecodeRows reads the committed decode fixture, one row per line after the
// comment and header lines.
func loadDecodeRows(t *testing.T) []decodeRow {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "diablo", "corpses", "decode-corpse.tsv")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed decode fixture %s: %v", path, err)
	}

	lines := decodeLines(raw)
	rows := make([]decodeRow, 0, len(lines))

	for i, line := range lines {
		row, err := parseDecodeRow(line, i)
		if err != nil {
			t.Fatalf("decode fixture %s line %d: %v", path, i+1, err)
		}

		rows = append(rows, row)
	}

	if len(rows) == 0 {
		t.Fatalf("decode fixture %s has no rows", path)
	}

	return rows
}

// assertDecoded checks one fixture row against the decoded tile reference.
func assertDecoded(t *testing.T, corpses *entities.Corpses, row decodeRow) {
	t.Helper()

	ref, ok := corpses.CorpseAt(entities.Point{X: row.x, Y: row.y})
	if !ok {
		t.Fatalf("byte %d at [%d][%d] decoded as empty", row.raw, row.x, row.y)
	}

	if ref.Index != row.corpseID || ref.Direction != row.direction {
		t.Fatalf("byte %d decoded as {Index:%d Direction:%d}, want {%d %d}",
			row.raw, ref.Index, ref.Direction, row.corpseID, row.direction)
	}
}

// drawsFor renders each row on its own map and returns the draw call count.
func drawsFor(rows []decodeRow) int {
	total := 0

	for _, row := range rows {
		var corpses entities.Corpses
		corpses.Tiles[row.y][row.x] = row.raw
		drawer := &recordingDrawer{}

		corpses.DrawTileCorpse(drawer, entities.Point{X: row.x, Y: row.y}, 12)
		total += len(drawer.calls)
	}

	return total
}

// zeroIDRows keeps the rows whose corpse id bits are zero.
func zeroIDRows(rows []decodeRow) []decodeRow {
	var zero []decodeRow

	for _, row := range rows {
		if row.raw&entities.CorpseIndexMask == 0 {
			zero = append(zero, row)
		}
	}

	return zero
}

// parseDecodeRow turns one fixture line into a decode row placed on its tile.
func parseDecodeRow(fields [3]string, index int) (decodeRow, error) {
	raw, err := strconv.Atoi(fields[0])
	if err != nil {
		return decodeRow{}, fmt.Errorf("byte: %w", err)
	}

	corpseID, err := strconv.Atoi(fields[1])
	if err != nil {
		return decodeRow{}, fmt.Errorf("corpse id: %w", err)
	}

	direction, err := strconv.Atoi(fields[2])
	if err != nil {
		return decodeRow{}, fmt.Errorf("direction: %w", err)
	}

	return decodeRow{
		x:         1 + index*decodeRowOffset,
		y:         2,
		raw:       int8(raw),
		corpseID:  int8(corpseID),
		direction: direction,
	}, nil
}

// decodeLines splits the fixture and drops its comment and header lines.
func decodeLines(raw []byte) [][3]string {
	var lines [][3]string

	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "byte\t") {
			continue
		}

		lines = append(lines, [3]string(strings.Split(line, "\t")))
	}

	return lines
}
