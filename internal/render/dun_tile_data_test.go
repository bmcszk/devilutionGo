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

// loadSOL loads the vendor SOL record of one level with its hand patches.
func loadSOL(t *testing.T, level render.LevelType) render.SOLData {
	t.Helper()

	raw := readFlatSize(t, filepath.Join(fixtureDir, "tiles", lSolFiles[level]))
	sol, err := render.LoadSOLData(raw)
	if err != nil {
		t.Fatalf("LoadSOLData(%v): %v", level, err)
	}

	return sol
}
