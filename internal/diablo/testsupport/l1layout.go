package testsupport

// The L1 level image: the parity-true cathedral layout the l1_parity gate pins
// (internal/diablo/levels/l1_parity_test.go, corpus internal/testdata/drlg),
// loaded as the world grid a session walks on. The fixture loader lives here
// because opening committed goldens is test scaffolding, not simulation code.

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// l1CorpusDir, l1PieceDir and l1SOLDir are the committed pieces of the
// parity-true L1 layout, relative to the repo root.
const (
	l1CorpusDir = "internal/testdata/drlg"
	l1PieceDir  = l1CorpusDir + "/dpiece"
	l1SOLDir    = "internal/testdata/render/tiles/sol"
)

// SOL flag bits the cathedral patch table writes: BlockLight is 1<<1 and
// BlockMissile is 1<<2 (levels/dun_tile.hpp:106-114).
const (
	solBlockLight   byte = 1 << 1
	solBlockMissile byte = 1 << 2
)

// cathedralSOLPatches is the cathedral patch list (dun_tile_data.cpp:109-133),
// one entry per upstream line.
var cathedralSOLPatches = []solPatch{
	{9, solBlockLight | solBlockMissile},
	{15, solBlockLight | solBlockMissile},
	{16, solBlockLight | solBlockMissile},
	{20, solBlockLight | solBlockMissile},
	{21, solBlockLight | solBlockMissile},
	{27, solBlockMissile},
	{28, solBlockMissile},
	{51, solBlockLight | solBlockMissile},
	{56, solBlockLight | solBlockMissile},
	{58, solBlockLight | solBlockMissile},
	{61, solBlockLight | solBlockMissile},
	{63, solBlockLight | solBlockMissile},
	{65, solBlockLight | solBlockMissile},
	{72, solBlockLight | solBlockMissile},
	{208, solBlockLight | solBlockMissile},
	{247, solBlockLight | solBlockMissile},
	{253, solBlockLight | solBlockMissile},
	{257, solBlockLight | solBlockMissile},
	{323, solBlockLight | solBlockMissile},
	{403, solBlockLight},
	{24, solBlockLight},
	{450, solBlockLight | solBlockMissile},
}

// solPatch is one hand patch: the record slot and the flags it gains.
type solPatch struct {
	index int
	add   byte
}

// L1Layout is one committed cathedral level image plus the tile the level enters
// on (ViewPosition, the fourth fixture record).
type L1Layout struct {
	// Image is the world grid the session loads: dPiece and the SOL record's
	// classification, both in the generator's own x-major indexing.
	Image *world.LevelLayout
	// Entry is the level's ViewPosition.
	Entry world.Tile
	// EntryWalkable reports whether the entry tile is not solid — the check
	// that the loaded layout really is walkable where the level says the
	// player arrives.
	EntryWalkable bool
}

// LoadL1Layout decodes one committed cathedral layout: the dPiece fixture, the
// retail SOL record with its cathedral hand patches, and the ViewPosition the
// tile-layer fixture carries. seed picks the corpus variant.
func LoadL1Layout(t *testing.T, level, seed uint32) *L1Layout {
	t.Helper()

	name := fmt.Sprintf("l%d-dpiece-main-%d.csv", level, seed)
	records := readCorpusCSV(t, filepath.Join(l1PieceDir, name))

	image := &world.LevelLayout{
		Piece: parsePieces(t, name, records),
		SOL:   applyCathedralSOLPatches(readCathedralSOL(t)),
	}

	entry := readEntryTile(t, level, seed)

	return &L1Layout{
		Image:         image,
		Entry:         entry,
		EntryWalkable: !solidAt(image, entry),
	}
}

// parsePieces reads the dPiece record of one fixture: the header record, then
// MAXDUNX*MAXDUNY x-major piece ids (dpiece_golden_test.go:150-158).
func parsePieces(t *testing.T, name string, records [][]string) []int {
	t.Helper()

	if len(records) < 2 {
		t.Fatalf("%s: %d records, want at least 2", name, len(records))
	}

	fields := records[1]
	if len(fields) != world.MAXDUNX*world.MAXDUNY {
		t.Fatalf("%s: %d pieces, want %d", name, len(fields), world.MAXDUNX*world.MAXDUNY)
	}

	pieces := make([]int, len(fields))

	for i, field := range fields {
		value, err := strconv.Atoi(field)
		if err != nil {
			t.Fatalf("%s: piece %d %q: %v", name, i, field, err)
		}

		pieces[i] = value
	}

	return pieces
}

// readCathedralSOL decodes the retail cathedral SOL record (the hex dump the
// render fixture keeps, dun_tile_data.cpp:107).
func readCathedralSOL(t *testing.T) []byte {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(repoRoot(t), l1SOLDir, "l1.sol.hex"))
	if err != nil {
		t.Fatalf("read cathedral SOL record: %v", err)
	}

	decoded, err := decodeHexDump(string(raw))
	if err != nil {
		t.Fatalf("decode cathedral SOL record: %v", err)
	}

	return decoded
}

// applyCathedralSOLPatches is the cathedral branch of LoadLevelSOLData
// (dun_tile_data.cpp:109-133): each patched entry gains its flags in the SOL
// record's own bit layout (levels/dun_tile.hpp:106-114).
func applyCathedralSOLPatches(record []byte) []byte {
	patched := make([]byte, len(record))
	copy(patched, record)

	for _, patch := range cathedralSOLPatches {
		if patch.index < len(patched) {
			patched[patch.index] |= patch.add
		}
	}

	return patched
}

// readEntryTile reads the ViewPosition record of one tile-layer fixture
// (golden_test.go:99-105: "ViewX,ViewY,leveltype").
func readEntryTile(t *testing.T, level, seed uint32) world.Tile {
	t.Helper()

	name := fmt.Sprintf("l%d-main-%d.csv", level, seed)
	records := readCorpusCSV(t, filepath.Join(l1CorpusDir, name))

	if len(records) != 4 || len(records[3]) != 3 {
		t.Fatalf("%s: %d records of %d last-record fields, want 4 of 3", name, len(records), len(records[3]))
	}

	x, err := strconv.Atoi(records[3][0])
	if err != nil {
		t.Fatalf("%s: parse ViewX: %v", name, err)
	}

	y, err := strconv.Atoi(records[3][1])
	if err != nil {
		t.Fatalf("%s: parse ViewY: %v", name, err)
	}

	return world.Tile{X: x, Y: y}
}

// solidAt reports whether the image classifies a tile as solid.
func solidAt(image *world.LevelLayout, tile world.Tile) bool {
	index := tile.X*world.MAXDUNY + tile.Y
	if index < 0 || index >= len(image.Piece) {
		return true
	}

	return world.SOLTileProperties(image.SOL[image.Piece[index]])&world.TileSolid != 0
}

// readCorpusCSV parses one committed fixture under the repo root.
func readCorpusCSV(t *testing.T, rel string) [][]string {
	t.Helper()

	file, err := os.Open(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("open fixture %s: %v", rel, err)
	}

	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse fixture %s: %v", rel, err)
	}

	return records
}

// decodeHexDump decodes whitespace-separated hex bytes.
func decodeHexDump(dump string) ([]byte, error) {
	joined := strings.Join(strings.Fields(dump), "")

	decoded := make([]byte, len(joined)/2)

	for i := range decoded {
		value, err := strconv.ParseUint(joined[2*i:2*i+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("hex byte %d: %w", i, err)
		}

		decoded[i] = byte(value)
	}

	return decoded, nil
}

// repoRoot resolves the repo root from the running test's working directory: go
// test runs with the package dir as cwd, so a fixed relative prefix would only
// serve one caller. The walk in internal/integration reads the same fixtures and
// must resolve them to the same tree.
func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve working directory: %v", err)
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod above %s: the fixture tree is not reachable", dir)
		}

		dir = parent
	}
}
