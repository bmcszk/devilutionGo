package assets_test

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// gfxFixtureDir holds the committed gfx goldens (internal/testdata/gfx).
const gfxFixtureDir = "../testdata/gfx"

// rndLvlPalCase is one TestLoadRndLvlPalDeterministicPerSeed row.
type rndLvlPalCase struct {
	name      string
	levelType int
	alternate bool
	rv        int32
	wantPath  string
	wantDraws int
}

// fakeSource is an in-memory AssetSource.
type fakeSource map[string][]byte

func (f fakeSource) ReadFile(name string) ([]byte, bool) {
	data, ok := f[name]

	return data, ok
}

func TestLevelPaletteNameTable(t *testing.T) {
	// given
	// palette.cpp:202-230: town and crypt are fixed, nest adds the
	// alternateNestArt flag to rv, every other level embeds its ordinal.
	tests := []struct {
		name      string
		levelType int
		rv        int
		alternate bool
		want      string
	}{
		{name: "town", levelType: 0, want: `levels\towndata\town.pal`},
		{name: "cathedral", levelType: 1, rv: 1, want: `levels\l1data\l1_1.pal`},
		{name: "catacombs", levelType: 2, rv: 2, want: `levels\l2data\l2_2.pal`},
		{name: "caves", levelType: 3, rv: 3, want: `levels\l3data\l3_3.pal`},
		{name: "hell", levelType: 4, rv: 4, want: `levels\l4data\l4_4.pal`},
		{name: "crypt ignores rv", levelType: 6, rv: 3, want: `nlevels\l5data\l5base.pal`},
		{name: "nest alternate art rv1", levelType: 5, rv: 1, alternate: true, want: `nlevels\l6data\l6base1.pal`},
		{name: "nest default art rv1", levelType: 5, rv: 1, want: `nlevels\l6data\l6base2.pal`},
		{name: "nest alternate art rv4", levelType: 5, rv: 4, alternate: true, want: `nlevels\l6data\l6base4.pal`},
		{name: "nest default art rv4", levelType: 5, rv: 4, want: `nlevels\l6data\l6base5.pal`},
	}
	// when every branch derives its path
	got := make([]string, len(tests))
	for i, tt := range tests {
		got[i] = assets.LevelPaletteName(tt.levelType, tt.rv, tt.alternate)
	}

	// then
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when the derived path is compared with the upstream literal
			got := got[i]
			// then
			if got != tt.want {
				t.Fatalf("LevelPaletteName(%d, %d, %v) = %q, want %q",
					tt.levelType, tt.rv, tt.alternate, got, tt.want)
			}
		})
	}
}

func TestBlendingRangesPerLevelType(t *testing.T) {
	// given
	// palette.cpp:188-200: caves and crypt skip 1..31, nest skips 1..15,
	// the remaining level types blend the full palette.
	tests := []struct {
		name      string
		levelType int
		want      assets.Blending
	}{
		{name: "town", levelType: 0, want: assets.Blending{Full: true}},
		{name: "cathedral", levelType: 1, want: assets.Blending{Full: true}},
		{name: "catacombs", levelType: 2, want: assets.Blending{Full: true}},
		{name: "caves", levelType: 3, want: assets.Blending{SkipFrom: 1, SkipTo: 31}},
		{name: "hell", levelType: 4, want: assets.Blending{Full: true}},
		{name: "nest", levelType: 5, want: assets.Blending{SkipFrom: 1, SkipTo: 15}},
		{name: "crypt", levelType: 6, want: assets.Blending{SkipFrom: 1, SkipTo: 31}},
	}
	// when every level type is classified
	got := make([]assets.Blending, len(tests))
	for i, tt := range tests {
		got[i] = assets.BlendingFor(tt.levelType)
	}

	// then
	for i, tt := range tests {
		if got[i] != tt.want {
			t.Errorf("%s: BlendingFor(%d) = %+v, want %+v", tt.name, tt.levelType, got[i], tt.want)
		}
	}
}

func TestLoadRndLvlPalDeterministicPerSeed(t *testing.T) {
	// given
	// palette.cpp:202-230: town and crypt return before the RNG draw; every
	// other type draws RandomIntBetween(1, 4) exactly once.
	tests := []rndLvlPalCase{
		{name: "town draws nothing", levelType: 0, wantPath: `levels\towndata\town.pal`, wantDraws: 0},
		{name: "crypt draws nothing", levelType: 6, wantPath: `nlevels\l5data\l5base.pal`, wantDraws: 0},
		{name: "cathedral rv3", levelType: 1, rv: 3, wantPath: `levels\l1data\l1_3.pal`, wantDraws: 1},
		{name: "hell rv4", levelType: 4, rv: 4, wantPath: `levels\l4data\l4_4.pal`, wantDraws: 1},
		{name: "nest default art", levelType: 5, rv: 4, wantPath: `nlevels\l6data\l6base5.pal`, wantDraws: 1},
		{
			name: "nest alternate art", levelType: 5, alternate: true, rv: 1,
			wantPath: `nlevels\l6data\l6base1.pal`, wantDraws: 1,
		},
	}
	// when every case draws with its stubbed RNG
	gotPaths := make([]string, len(tests))
	gotDraws := make([]int, len(tests))
	for i, tt := range tests {
		gotPaths[i], gotDraws[i] = runRndLvlPalCase(t, tt)
	}

	// then
	for i, tt := range tests {
		if gotPaths[i] != tt.wantPath {
			t.Errorf("%s: path = %q, want %q", tt.name, gotPaths[i], tt.wantPath)
		}

		if gotDraws[i] != tt.wantDraws {
			t.Errorf("%s: RNG draws = %d, want %d", tt.name, gotDraws[i], tt.wantDraws)
		}
	}
}

func TestLoadRndLvlPalBlendingPerLevelType(t *testing.T) {
	// given
	tests := []struct {
		levelType int
		want      assets.Blending
	}{
		{levelType: 3, want: assets.Blending{SkipFrom: 1, SkipTo: 31}},
		{levelType: 5, want: assets.Blending{SkipFrom: 1, SkipTo: 15}},
		{levelType: 1, want: assets.Blending{Full: true}},
	}
	// when
	got := make([]assets.Blending, len(tests))
	for i, tt := range tests {
		_, got[i], _ = assets.LoadRndLvlPal(tt.levelType, false, func(int32, int32) int32 { return 2 })
	}

	// then
	for i, tt := range tests {
		if got[i] != tt.want {
			t.Errorf("level type %d: blending = %+v, want %+v", tt.levelType, got[i], tt.want)
		}
	}
}

func TestLoadPaletteFromStackMissingIsError(t *testing.T) {
	// given
	src := fakeSource{}
	// when
	p, err := assets.LoadPaletteFromStack(src, `levels\towndata\town.pal`)
	// then
	if err == nil {
		t.Fatal("expected an error for a missing palette asset")
	}

	if p != nil {
		t.Fatalf("palette = %v, want nil", p)
	}

	if !strings.Contains(err.Error(), `levels\towndata\town.pal`) {
		t.Fatalf("error %q does not name the missing path", err)
	}
}

func TestLoadPaletteFromStack(t *testing.T) {
	// given
	// The committed ui_art/diablo.pal golden (768 bytes), served under the
	// upstream path.
	raw := readHexFixture(t, filepath.Join(gfxFixtureDir, "pal/diablo.pal.hex"))
	src := fakeSource{`ui_art\diablo.pal`: raw}
	// when
	p, err := assets.LoadPaletteFromStack(src, `ui_art\diablo.pal`)
	// then
	if err != nil {
		t.Fatalf("LoadPaletteFromStack: %v", err)
	}

	triples := [][3]uint8{{0, 0, 0}, {191, 0, 0}, {0, 191, 0}, {0, 255, 255}, {255, 255, 255}}
	indices := []int{0, 1, 2, 254, 255}
	for i, idx := range indices {
		if p[idx] != triples[i] {
			t.Errorf("palette[%d] = %v, want %v", idx, p[idx], triples[i])
		}
	}
}

func TestCycleColorsRotatesRangeForwardAndReverse(t *testing.T) {
	// given
	// palette.cpp:51-91: std::rotate by one inside [from, to]; outside the
	// range is untouched.
	identity := identityPalette()
	forward, reverse := identity, identity
	// when
	assets.CycleColors(&forward, 1, 3)
	assets.CycleColorsReverse(&reverse, 1, 3)
	// then
	if forward[0] != identity[0] || forward[4] != identity[4] {
		t.Error("CycleColors touched colors outside [1, 3]")
	}

	if forward[1] != identity[2] || forward[2] != identity[3] || forward[3] != identity[1] {
		t.Errorf("CycleColors = %v, want [2 3 1] shift", forward[1:4])
	}

	if reverse[1] != identity[3] || reverse[2] != identity[1] || reverse[3] != identity[2] {
		t.Errorf("CycleColorsReverse = %v, want [3 1 2] shift", reverse[1:4])
	}
}

func TestPaletteUpdateCavesCyclesOneToThirtyOne(t *testing.T) {
	// given
	// palette.cpp:363-367: every call cycles [1, 31].
	p := identityPalette()
	// when
	assets.PaletteUpdateForCaves(&p, 0)
	// then
	if p[1] != identityPalette()[2] || p[31] != identityPalette()[1] {
		t.Fatalf("caves cycle = %v..%v, want 2..1", p[1], p[31])
	}

	if p[0] != identityPalette()[0] || p[32] != identityPalette()[32] {
		t.Fatal("caves cycle touched colors outside [1, 31]")
	}
}

func TestPaletteUpdateCryptAlternatesLavaEveryOtherFrame(t *testing.T) {
	// given
	// palette.cpp:372-386: the lava range [1, 15] cycles on every other call,
	// the glow range [16, 31] on every call.
	p := identityPalette()
	base := identityPalette()
	// when
	assets.PaletteUpdateForCrypt(&p, 0)
	lavaAfterFirst := p[1]
	glowAfterFirst := p[16]
	assets.PaletteUpdateForCrypt(&p, 1)
	lavaAfterSecond := p[1]
	glowAfterSecond := p[16]
	assets.PaletteUpdateForCrypt(&p, 2)
	lavaAfterThird := p[1]
	// then
	if lavaAfterFirst != base[15] || glowAfterFirst != base[31] {
		t.Fatalf("frame 0 = lava %v glow %v, want %v %v", lavaAfterFirst, glowAfterFirst, base[15], base[31])
	}

	if lavaAfterSecond != lavaAfterFirst {
		t.Fatalf("frame 1 lava = %v, want unchanged %v", lavaAfterSecond, lavaAfterFirst)
	}

	if glowAfterSecond != base[30] {
		t.Fatalf("frame 1 glow = %v, want %v", glowAfterSecond, base[30])
	}

	if lavaAfterThird != base[14] {
		t.Fatalf("frame 2 lava = %v, want %v", lavaAfterThird, base[14])
	}
}

func TestPaletteUpdateHiveEveryThirdFrame(t *testing.T) {
	// given
	// palette.cpp:389-405: bubbles [1, 8] and waves [9, 15] cycle only on
	// every third call.
	p := identityPalette()
	base := identityPalette()
	// when
	assets.PaletteUpdateForHive(&p, 0)
	first := p[1]
	assets.PaletteUpdateForHive(&p, 1)
	second := p[1]
	assets.PaletteUpdateForHive(&p, 2)
	third := p[1]
	bubble := p[9]
	// then
	if first != base[1] || second != base[1] {
		t.Fatalf("hive cycled before the third frame: %v %v", first, second)
	}

	if third != base[8] || bubble != base[15] {
		t.Fatalf("hive cycle = %v..%v, want %v..%v", third, bubble, base[8], base[15])
	}

	if p[0] != base[0] || p[16] != base[16] {
		t.Fatal("hive cycle touched colors outside [1, 15]")
	}
}

func identityPalette() assets.Palette {
	var p assets.Palette
	for i := range 256 {
		p[i] = [3]uint8{byte(i), 0, 0}
	}

	return p
}

// readHexFixture reads a committed hex dump; a missing golden fails the test.
func readHexFixture(tb testing.TB, path string) []byte {
	tb.Helper()

	dump, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read fixture %s: %v", path, err)
	}

	raw, err := hex.DecodeString(strings.Join(strings.Fields(string(dump)), ""))
	if err != nil {
		tb.Fatalf("decode fixture %s: %v", path, err)
	}

	return raw
}

// runRndLvlPalCase resolves one row with a counting RNG stub.
func runRndLvlPalCase(t *testing.T, tt rndLvlPalCase) (path string, draws int) {
	t.Helper()

	nextInt := func(lo, hi int32) int32 {
		draws++
		if lo != 1 || hi != 4 {
			t.Errorf("%s: nextInt(%d, %d), want (1, 4)", tt.name, lo, hi)
		}

		return tt.rv
	}

	path, _, err := assets.LoadRndLvlPal(tt.levelType, tt.alternate, nextInt)
	if err != nil {
		t.Fatalf("%s: LoadRndLvlPal: %v", tt.name, err)
	}

	return path, draws
}
