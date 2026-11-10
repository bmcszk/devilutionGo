package render_test

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render"
)

const (
	l1PieceCount = 453

	// l1MinFile is the retail cathedral MIN record: levels\l1data\l1.min
	// (dun_tile_data.cpp:81), 9060 bytes = 4530 u16 = 453 pieces of 10 blocks.
	l1MinFile   = "min/l1.min.hex"
	townMinFile = "min/town.min.hex"

	// townMinPieces notes the retail town MIN record: nlevels\towndata\
	// town.min (dun_tile_data.cpp:73), 44128 bytes = 22064 u16 = 1379 pieces
	// of 16 blocks.
	townPieceCount = 1379
)

// l1SolFile is the retail cathedral SOL record from the offline producer run
// (internal/testdata/render/producer/go-tiles): levels\l1data\l1.sol
// (dun_tile_data.cpp:107), 453 pieces of 10 micro blocks (4530 u16 in l1.min).
const l1SolFile = "sol/l1.sol.hex"

// lSolFiles pairs the four dungeon levels whose SOL hand-patches the port
// includes (dun_tile_data.cpp:107-159).
var lSolFiles = map[render.LevelType]string{
	render.LevelCathedral: "sol/l1.sol.hex",
	render.LevelCaves:     "sol/l3.sol.hex",
	render.LevelHell:      "sol/l4.sol.hex",
	render.LevelCrypt:     "sol/l5.sol.hex",
}

// cathedralUpstreamIndexes is the cathedral hand-patch list
// (dun_tile_data.cpp:109-133).
var cathedralUpstreamIndexes = []int{
	9, 15, 16, 20, 21, 27, 28, 51, 56, 58, 61, 63, 65, 72,
	208, 247, 253, 257, 323, 403, 24, 450,
}

// TestSOLDataParsesRetailRecords loads the real cathedral SOL record and pins
// its leading values (vendor bytes od-dumped when the fixture was produced).
func TestSOLDataParsesRetailRecords(t *testing.T) {
	// given
	raw := readFlatSize(t, filepath.Join(fixtureDir, "tiles", l1SolFile), 453)

	// when
	sol, err := render.LoadSOLData(raw)
	// then
	if err != nil {
		t.Fatalf("LoadSOLData: %v", err)
	}

	prefix := []render.TileProperties{0x1f, 0x0, 0x9f, 0x0, 0x2f, 0xaf, 0x0, 0xf}
	for i, want := range prefix {
		if sol[i] != want {
			t.Fatalf("sol[%d] = %#x, want %#x", i, sol[i], want)
		}
	}
}

// TestSOLDataRejectsOversizedRecords bounds the record at MAXTILES; upstream
// reads the record straight into SOLData[MAXTILES] (dun_tile_data.hpp:87)
// with no room to overrun.
func TestSOLDataRejectsOversizedRecords(t *testing.T) {
	// given
	raw := make([]byte, render.MaxTiles+1)

	// when
	_, err := render.LoadSOLData(raw)

	// then
	if err == nil {
		t.Fatal("LoadSOLData accepted a record larger than MAXTILES")
	}
}

// TestSOLHandPatchesCathedral pins the cathedral arch/wall/pillar additions,
// index-exact per dun_tile_data.cpp:109-133.
func TestSOLHandPatchesCathedral(t *testing.T) {
	// given
	base := loadSOL(t, render.LevelCathedral)
	arch := render.TilePropertiesBlockLight | render.TilePropertiesBlockMissile
	wall := render.TilePropertiesBlockMissile
	pillar := render.TilePropertiesBlockLight
	cases := []struct {
		index    int
		property render.TileProperties
	}{
		{9, arch},
		{15, arch},
		{16, arch},
		{20, arch},
		{21, arch},
		{27, wall},
		{28, wall},
		{51, arch},
		{56, arch},
		{58, arch},
		{61, arch},
		{63, arch},
		{65, arch},
		{72, arch},
		{208, arch},
		{247, arch},
		{253, arch},
		{257, arch},
		{323, arch},
		{403, pillar},
		{24, pillar},
		{450, arch},
	}

	// when
	sol := render.ApplySOLHandPatches(base, render.LevelCathedral)

	// then
	expectSOLHave(t, sol, cases)
}

// TestSOLHandPatchesCaves pins the fence-substitution and floor-pixel fixes
// (dun_tile_data.cpp:140-147).
func TestSOLHandPatchesCaves(t *testing.T) {
	// given
	base := loadSOL(t, render.LevelCaves)

	// when
	sol := render.ApplySOLHandPatches(base, render.LevelCaves)

	// then: the fence sub-tiles become solid and tile 170 keeps its floor
	// pixels from drawing.
	expectSOLHave(t, sol, []struct {
		index    int
		property render.TileProperties
	}{
		{170, render.TilePropertiesBlockMissile},
		{481, render.TilePropertiesSolid},
		{487, render.TilePropertiesSolid},
	})
}

// TestSOLHandPatchesSolidClears pins the "incorrectly marked as being solid"
// clears of hell and crypt (dun_tile_data.cpp:151, :158).
func TestSOLHandPatchesSolidClears(t *testing.T) {
	// given
	l4 := loadSOL(t, render.LevelHell)
	l5 := loadSOL(t, render.LevelCrypt)

	// when
	got := render.ApplySOLHandPatches(l4, render.LevelHell)
	crypt := render.ApplySOLHandPatches(l5, render.LevelCrypt)

	// then
	if got[210] != 0 {
		t.Errorf("hell sol[210] = %#x, want 0", got[210])
	}
	if crypt[142] != 0 {
		t.Errorf("crypt sol[142] = %#x, want 0", crypt[142])
	}
}

// TestSOLPatchTableShape pins the patch table rows against the exact upstream
// index list (dun_tile_data.cpp:109-158).
func TestSOLPatchTableShape(t *testing.T) {
	// given
	want := map[render.LevelType][]int{
		render.LevelCathedral: cathedralUpstreamIndexes,
		render.LevelCaves:     {170, 481, 487},
		render.LevelHell:      {210},
		render.LevelCrypt:     {142},
	}

	// when
	got := map[render.LevelType][]int{}
	for level := range want {
		patches := render.SOLPatchesFor(level)
		indexes := make([]int, 0, len(patches))

		for _, patch := range patches {
			indexes = append(indexes, patch.Index)
		}

		got[level] = indexes
	}

	// then
	for level, expected := range want {
		if !slices.Equal(got[level], expected) {
			t.Fatalf("%v patch rows %v, want %v", level, got[level], expected)
		}
	}
}

// TestMinScrambleSpotValues pins the first pieces of the retail cathedral and
// town MIN records decoded through the on-disk scramble
// (dun_tile_data.cpp:186). Every pinned value was od-read from the vendor
// dump and hand-decoded with the upstream formula.
func TestMinScrambleSpotValuesCathedralHandDecoded(t *testing.T) {
	// given
	pieces := loadMinPieces(t, l1MinFile, l1PieceCount*10)

	// when
	got, err := render.DecodeDungeonPieces(pieces, render.LevelCathedral, render.SOLData{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	// then
	if got.TileLen != 10 {
		t.Fatalf("microTileLen = %d, want 10", got.TileLen)
	}

	if len(got.Micros) != l1PieceCount {
		t.Fatalf("piece count %d, want %d", len(got.Micros), l1PieceCount)
	}

	// piece 0: mt[0] left trapezoid frame 4, mt[1] right triangle
	// frame 5, mt[2] square frame 3, mt[4] square frame 2,
	// mt[6] transparent square frame 1.
	carries := map[int]render.LevelCelBlock{0: 0x4004, 1: 0x3005, 2: 0x0003, 4: 0x0002, 6: 0x1001}
	empties := []int{3, 5, 7, 8, 9}
	expectMicros(t, got.Micros[0], carries, empties)

	// piece 1: a plain wall with the floor triangle pair 6/7.
	carries = map[int]render.LevelCelBlock{0: 0x2006, 1: 0x3007}
	empties = []int{2, 3, 4, 5, 6, 7, 8, 9}
	expectMicros(t, got.Micros[1], carries, empties)
}

// TestMinScrambleSpotValuesTown pins the first retail town piece decoding
// through the blocks-16 scramble (dun_tile_data.cpp:171-174).
func TestMinScrambleSpotValuesTown(t *testing.T) {
	// given
	pieces := loadMinPieces(t, townMinFile, townPieceCount*16)

	// when
	got, err := render.DecodeDungeonPieces(pieces, render.LevelTown, render.SOLData{})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	// then: mt[0] left triangle frame 1, mt[1] right triangle frame 2.
	if got.TileLen != 16 {
		t.Fatalf("microTileLen = %d, want 16", got.TileLen)
	}

	if len(got.Micros) != townPieceCount {
		t.Fatalf("piece count %d, want %d", len(got.Micros), townPieceCount)
	}

	carries := map[int]render.LevelCelBlock{0: 0x2001, 1: 0x3002}
	empties := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	expectMicros(t, got.Micros[0], carries, empties)
}

// TestRenumberLevelCelBlocks pins the upper-bound subtraction of
// ComputeCelBlockAdjustments' application (dun_tile_data.cpp:210-216):
// std::upper_bound finds the first adjustment whose Frame is strictly
// greater than the block's frame, and that entry's Subtract is deducted.
func TestRenumberLevelCelBlocks(t *testing.T) {
	// ComputeCelBlockAdjustments over the kept frames [1,3]
	// (reencode_dun_cels.cpp:305-320) — the gap after frame 1 leaves row
	// (3, 0), the trailing row (0xFFFF, 1) covers every frame past the last
	// kept frame
	adjustments := []render.CelBlockAdjustment{
		{Frame: 3, Subtract: 0},
		{Frame: 0xFFFF, Subtract: 1},
	}

	cases := []struct {
		name  string
		frame uint16
		want  uint16
	}{
		{name: "gap", frame: 3, want: 2},         // the Frame==3 row is skipped, the trailing row subtracts 1
		{name: "below", frame: 1, want: 1},       // nothing subtracts before the first kept frame
		{name: "above total", frame: 5, want: 4}, // only the 0xFFFF row lies beyond, subtract 1
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			micros := render.DungeonMicros{{render.LevelCelBlock(0x3000 | tc.frame)}}

			// when
			render.RenumberLevelCelBlocks(micros, adjustments)

			// then
			got := micros[0][0].Frame()
			if got != tc.want {
				t.Fatalf("frame = %d, want %d", got, tc.want)
			}
		})
	}

	t.Run("above all without trailing row stays unchanged", func(t *testing.T) {
		// given: upstream builds the 0xFFFF row only after dropped frames
		// (reencode_dun_cels.cpp:316-318); with just the (3,0) row a frame
		// beyond it finds no entry
		micros := render.DungeonMicros{{render.LevelCelBlock(0x3005)}}

		// when
		render.RenumberLevelCelBlocks(micros, adjustments[:1])

		// then
		if got := micros[0][0].Frame(); got != 5 {
			t.Fatalf("frame = %d, want 5", got)
		}
	})
}

// expectSOLHave checks one patched record's expected bits per index.
func expectSOLHave(
	t *testing.T,
	sol render.SOLData,
	cases []struct {
		index    int
		property render.TileProperties
	},
) {
	t.Helper()

	for _, want := range cases {
		if sol[want.index]&want.property != want.property {
			t.Errorf("sol[%d] = %#x, missing %#x", want.index, sol[want.index], want.property)
		}
	}
}

// readFlatSize decodes a hex fixture, optionally checking the byte count.
func readFlatSize(t *testing.T, path string, want ...int) []byte {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	decoded, err := hex.DecodeString(strings.Join(strings.Fields(string(body)), ""))
	if err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}

	if len(want) > 0 && len(decoded) != want[0] {
		t.Fatalf("fixture %s: %d bytes, want %d", path, len(decoded), want[0])
	}

	return decoded
}

// loadSOL loads the vendor SOL record of one level.
func loadSOL(t *testing.T, level render.LevelType) render.SOLData {
	t.Helper()

	raw := readFlatSize(t, filepath.Join(fixtureDir, "tiles", lSolFiles[level]))
	sol, err := render.LoadSOLData(raw)
	if err != nil {
		t.Fatalf("LoadSOLData(%v): %v", level, err)
	}

	return sol
}

// loadMinPieces decodes a retail MIN record hex fixture of one record word.
func loadMinPieces(t *testing.T, name string, words int) []render.LevelCelBlock {
	t.Helper()

	raw := readFlatSize(t, filepath.Join(fixtureDir, "tiles", name), words*2)
	pieces, err := render.LoadMinPieces(raw)
	if err != nil {
		t.Fatalf("LoadMinPieces(%s): %v", name, err)
	}

	return pieces
}

// expectMicros checks one decoded piece against expected values and empties.
func expectMicros(
	t *testing.T,
	micros render.MICROS,
	carries map[int]render.LevelCelBlock,
	empties []int,
) {
	t.Helper()

	for block, want := range carries {
		if micros[block] != want {
			t.Errorf("mt[%d] = %#x, want %#x", block, micros[block], want)
		}
	}

	for _, block := range empties {
		if micros[block] != 0 {
			t.Errorf("mt[%d] = %#x, want 0", block, micros[block])
		}
	}
}
