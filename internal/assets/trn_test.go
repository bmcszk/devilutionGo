package assets_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

// trnCensusSize is the fixed TRN table length (lighting.cpp:37).
const trnCensusSize = 256

// trnCensusCount is how many TRN fixtures are committed, so one cannot
// silently disappear from internal/testdata/gfx/trn.
const trnCensusCount = 6

// trnCensusDigests pins the decoded bytes of every committed TRN fixture; a
// fixture without an entry fails the census instead of passing unnoticed.
var trnCensusDigests = map[string]string{
	"fonts_red.trn.hex":                "a22400b6230646cd079e8249323ea01005d6ef0aea0875ad0e8d7b501b54e076",
	"fonts_whitegold.trn.hex":          "49562f773f976c3a63be50f1826e41feb24048968b26580681151b6fb542dba4",
	"gendata_pause.trn.hex":            "6efd80223719c22995185e565b6e1a2eb4f22d7d41dac594d19e549b107fd421",
	"plrgfx_warrior_whu_whufm.trn.hex": "a2f0c9b53396aa7192a770551bc2b60562fd29b5b71196c0a0c7a80bb013c240",
	"plrgfx_warrior_whu_whulm.trn.hex": "a2f0c9b53396aa7192a770551bc2b60562fd29b5b71196c0a0c7a80bb013c240",
	"plrgfx_warrior_whu_whuqm.trn.hex": "a2f0c9b53396aa7192a770551bc2b60562fd29b5b71196c0a0c7a80bb013c240",
}

// recordingSource records every requested path before serving it.
type recordingSource struct {
	files fakeSource
	order *[]string
}

func (r recordingSource) ReadFile(name string) ([]byte, bool) {
	*r.order = append(*r.order, name)

	return r.files.ReadFile(name)
}

func TestDecodeTRNRejectsWrongSize(t *testing.T) {
	// given
	// A TRN file is exactly 256 bytes (lighting.cpp:214-219 arrays).
	sizes := []int{0, 255, 257}
	// when every size is decoded
	errs := make([]error, len(sizes))
	for i, size := range sizes {
		_, errs[i] = assets.DecodeTRN(make([]byte, size))
	}

	// then
	for i, size := range sizes {
		if errs[i] == nil {
			t.Errorf("DecodeTRN(%d bytes) = nil error, want size error", size)
		}
	}
}

func TestDecodeTRNGolden(t *testing.T) {
	// given
	// Committed gendata/pause.trn (devilutionX assets, 256 bytes).
	raw := readHexFixture(t, filepath.Join(gfxFixtureDir, "trn/gendata_pause.trn.hex"))
	// when
	trn, err := assets.DecodeTRN(raw)
	// then
	if err != nil {
		t.Fatalf("DecodeTRN: %v", err)
	}

	// Spot values pin the file bytes independently of the loader.
	want := map[int]byte{0: 0, 1: 227, 2: 228, 15: 0, 128: 224, 255: 0}
	for idx, w := range want {
		if trn[idx] != w {
			t.Errorf("trn[%d] = %d, want %d", idx, trn[idx], w)
		}
	}
}

func TestApplyTRNRemapsIndicesInPlace(t *testing.T) {
	// given
	// clx_render.cpp:423: dst = trn[src], in place.
	var trn assets.TRN
	for i := range trn {
		trn[i] = byte(255 - i)
	}

	pixels := []byte{0, 1, 2, 255}
	want := []byte{255, 254, 253, 0}
	// when
	assets.ApplyTRN(pixels, trn)
	// then
	for i := range want {
		if pixels[i] != want[i] {
			t.Fatalf("pixels[%d] = %d, want %d", i, pixels[i], want[i])
		}
	}
}

func TestLoadTrnsRequestsInfraStonePauseInOrder(t *testing.T) {
	// given
	// lighting.cpp:214-219 loads the three engine tables in this exact order.
	var order []string
	src := recordingSource{files: fakeSource{
		`plrgfx\infra.trn`:  filledTRN(1),
		`plrgfx\stone.trn`:  filledTRN(2),
		`gendata\pause.trn`: filledTRN(3),
	}, order: &order}
	// when
	reg, err := assets.LoadTrns(src)
	// then
	if err != nil {
		t.Fatalf("LoadTrns: %v", err)
	}

	want := []string{`plrgfx\infra.trn`, `plrgfx\stone.trn`, `gendata\pause.trn`}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Fatalf("requested = %v, want %v", order, want)
	}

	if reg == nil || reg.Infravision != assets.TRN(filledTRN(1)) {
		t.Fatalf("registry = %+v, want the infra table", reg)
	}
}

func TestLoadTrnsStopsAtFirstFailure(t *testing.T) {
	// given
	// stone.trn is absent, so pause.trn must never be requested.
	var order []string
	src := recordingSource{files: fakeSource{`plrgfx\infra.trn`: filledTRN(1)}, order: &order}
	// when
	_, err := assets.LoadTrns(src)
	// then
	if err == nil {
		t.Fatal("expected an error for the missing stone.trn")
	}

	if !strings.Contains(err.Error(), `plrgfx\stone.trn`) {
		t.Fatalf("error %q does not name the missing path", err)
	}

	if len(order) != 2 {
		t.Fatalf("requested = %v, want a stop after stone.trn", order)
	}
}

func TestClassTRNMissingIsNilNotError(t *testing.T) {
	// given
	// trn.cpp:29-45 uses LoadOptionalFileInMem: a missing class TRN is not an
	// error.
	reg := &assets.TRNRegistry{}
	// when
	trn, ok, err := reg.ClassTRN(fakeSource{}, "warrior")
	// then
	if err != nil {
		t.Fatalf("ClassTRN: %v", err)
	}

	if ok || trn != (assets.TRN{}) {
		t.Fatalf("ClassTRN = %v, %v, want zero TRN and false", trn, ok)
	}
}

func TestClassTRNPathAndContent(t *testing.T) {
	// given
	raw := readHexFixture(t, filepath.Join(gfxFixtureDir, "trn/fonts_red.trn.hex"))
	src := fakeSource{`plrgfx\warrior.trn`: raw}
	reg := &assets.TRNRegistry{}
	// when
	trn, ok, err := reg.ClassTRN(src, "warrior")
	// then
	if err != nil || !ok {
		t.Fatalf("ClassTRN = ok %v, err %v", ok, err)
	}

	if trn[0] != 0 || trn[255] != 255 {
		t.Fatalf("trn endpoints = %d, %d, want 0, 255", trn[0], trn[255])
	}
}

func TestPlayerGraphicTRNPath(t *testing.T) {
	// given
	// trn.cpp:48-60: <graphic path>.trn sits next to the CL2.
	raw := readHexFixture(t, filepath.Join(gfxFixtureDir, "trn/plrgfx_warrior_whu_whuqm.trn.hex"))
	src := fakeSource{`plrgfx\warrior\whu\whuqm.trn`: raw}
	reg := &assets.TRNRegistry{}
	// when
	trn, ok, err := reg.PlayerGraphicTRN(src, `plrgfx\warrior\whu\whuqm`)
	// then
	if err != nil || !ok {
		t.Fatalf("PlayerGraphicTRN = ok %v, err %v", ok, err)
	}

	if trn[128] != 255 || trn[254] != 253 {
		t.Fatalf("trn[128], trn[254] = %d, %d, want 255, 253", trn[128], trn[254])
	}
}

func TestMonsterTRNMissingIsError(t *testing.T) {
	// given
	// monster.cpp:172-178 loads the monster TRN with LoadFileInMem (required).
	reg := &assets.TRNRegistry{}
	// when
	_, err := reg.MonsterTRN(fakeSource{}, "zombie/bluered")
	// then
	if err == nil {
		t.Fatal("expected an error for the missing monster TRN")
	}

	if !strings.Contains(err.Error(), `monsters\zombie/bluered.trn`) {
		t.Fatalf("error %q does not name the monster TRN path", err)
	}
}

func TestMonsterTRNPathAndContent(t *testing.T) {
	// given
	raw := readHexFixture(t, filepath.Join(gfxFixtureDir, "trn/plrgfx_warrior_whu_whufm.trn.hex"))
	src := fakeSource{`monsters\zombie/bluered.trn`: raw}
	reg := &assets.TRNRegistry{}
	// when
	trn, err := reg.MonsterTRN(src, "zombie/bluered")
	// then
	if err != nil {
		t.Fatalf("MonsterTRN: %v", err)
	}

	if trn[254] != 253 {
		t.Fatalf("trn[254] = %d, want 253", trn[254])
	}
}

func TestMonsterTRNSanitizedReplaces255With0(t *testing.T) {
	// given
	// monster.cpp:178: std::replace(255 -> 0) before the table is applied.
	var trn assets.TRN
	trn[0] = 255
	trn[10] = 7
	trn[255] = 255
	// when
	got := assets.MonsterTRNSanitized(trn)
	// then
	if got[0] != 0 || got[255] != 0 {
		t.Fatalf("sanitized entries = %d, %d, want 0, 0", got[0], got[255])
	}

	if got[10] != 7 {
		t.Fatalf("sanitized entry 10 = %d, want 7", got[10])
	}

	if trn[0] != 255 {
		t.Fatal("MonsterTRNSanitized mutated its input")
	}
}

func TestMonsterTRNSkipAnimOneForCouncilorFamily(t *testing.T) {
	// given
	// monster.cpp:181-184: animation 1 is skipped for these four types.
	tests := []struct {
		name        string
		monsterType assets.MonsterType
		animIndex   int
		want        bool
	}{
		{name: "councilor anim 1", monsterType: assets.MTCounslr, animIndex: 1, want: true},
		{name: "magistr anim 1", monsterType: assets.MTMagistr, animIndex: 1, want: true},
		{name: "cabalist anim 1", monsterType: assets.MTCabalist, animIndex: 1, want: true},
		{name: "advocate anim 1", monsterType: assets.MTAdvocate, animIndex: 1, want: true},
		{name: "councilor anim 0", monsterType: assets.MTCounslr, animIndex: 0, want: false},
		{name: "councilor anim 2", monsterType: assets.MTCounslr, animIndex: 2, want: false},
	}
	// when
	got := make([]bool, len(tests))
	for i, tt := range tests {
		got[i] = assets.MonsterTRNSkipAnim(tt.monsterType, tt.animIndex)
	}

	// then
	for i, tt := range tests {
		if got[i] != tt.want {
			t.Errorf("%s: MonsterTRNSkipAnim = %v, want %v", tt.name, got[i], tt.want)
		}
	}
}

func TestTRNFixtureCensus(t *testing.T) {
	// Every .trn.hex committed under internal/testdata/gfx/trn is a census
	// member; a missing directory or file fails, never skips.
	// when the committed fixture directory is enumerated
	names := trnFixtureNames(t)
	// then
	if len(names) != trnCensusCount {
		t.Fatalf("committed TRN fixtures = %d, want %d", len(names), trnCensusCount)
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			// when
			trn, raw := decodeTRNFixture(t, name)
			// then
			expectTRNCensusEntry(t, name, trn, raw)
		})
	}
}

// trnFixtureNames lists the committed TRN fixtures by file name.
func TestMonsterTypeFromMonsterIDMapsCouncilorFamily(t *testing.T) {
	// given
	// monstdat.h:266-269 places MT_COUNSLR..MT_ADVOCATE at _monster_id
	// ordinals 105..108 (0 is MT_NZOMBIE), not at the local 0..3 values.
	tests := []struct {
		name      string
		monsterID int
		want      assets.MonsterType
	}{
		{name: "councilor", monsterID: 105, want: assets.MTCounslr},
		{name: "magistr", monsterID: 106, want: assets.MTMagistr},
		{name: "cabalist", monsterID: 107, want: assets.MTCabalist},
		{name: "advocate", monsterID: 108, want: assets.MTAdvocate},
	}
	// when every ordinal is converted
	got := make([]assets.MonsterType, len(tests))
	for i, tt := range tests {
		got[i], _ = assets.MonsterTypeFromMonsterID(tt.monsterID)
	}

	// then the conversion maps each C++ ordinal to its type
	for i, tt := range tests {
		if got[i] != tt.want {
			t.Errorf("%s: MonsterTypeFromMonsterID(%d) = %v, want %v", tt.name, tt.monsterID, got[i], tt.want)
		}
	}
}

func TestMonsterTypeFromMonsterIDRejectsOthers(t *testing.T) {
	// given ids outside the councilor family, plus the neighbours of the range
	ids := []int{-1, 0, 8, 104, 109, 110, 200}
	// when each is converted
	errs := make([]error, len(ids))
	for i, id := range ids {
		_, errs[i] = assets.MonsterTypeFromMonsterID(id)
	}

	// then every one is an error
	for i, id := range ids {
		if errs[i] == nil {
			t.Errorf("MonsterTypeFromMonsterID(%d) = nil error, want out-of-range error", id)
		}
	}
}

func TestMonsterTRNSkipAnimThroughConversion(t *testing.T) {
	// given the four C++ ordinals
	ids := []int{105, 106, 107, 108}
	// when each is converted and asked about animation 1 and 0
	skipOne := make([]bool, len(ids))
	skipZero := make([]bool, len(ids))
	for i, id := range ids {
		monsterType, err := assets.MonsterTypeFromMonsterID(id)
		if err != nil {
			t.Fatalf("MonsterTypeFromMonsterID(%d): %v", id, err)
		}

		skipOne[i] = assets.MonsterTRNSkipAnim(monsterType, 1)
		skipZero[i] = assets.MonsterTRNSkipAnim(monsterType, 0)
	}

	// then animation 1 is skipped for all four and animation 0 never is
	for i, id := range ids {
		if !skipOne[i] || skipZero[i] {
			t.Errorf("id %d: skip anim1 = %v, skip anim0 = %v, want true, false", id, skipOne[i], skipZero[i])
		}
	}
}

func trnFixtureNames(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join(gfxFixtureDir, "trn"))
	if err != nil {
		t.Fatalf("read committed TRN corpus: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if _, ok := strings.CutSuffix(entry.Name(), ".trn.hex"); ok {
			names = append(names, entry.Name())
		}
	}

	return names
}

// decodeTRNFixture loads one committed hex dump and decodes it.
func decodeTRNFixture(t *testing.T, name string) (assets.TRN, []byte) {
	t.Helper()

	raw := readHexFixture(t, filepath.Join(gfxFixtureDir, "trn", name))

	trn, err := assets.DecodeTRN(raw)
	if err != nil {
		t.Fatalf("%s: DecodeTRN: %v", name, err)
	}

	return trn, raw
}

// expectTRNCensusEntry pins one fixture's shape, bytes, and round-trip identity.
func expectTRNCensusEntry(t *testing.T, name string, trn assets.TRN, raw []byte) {
	t.Helper()

	if len(raw) != trnCensusSize {
		t.Fatalf("%s: decoded bytes = %d, want %d", name, len(raw), trnCensusSize)
	}

	wantDigest, ok := trnCensusDigests[name]
	if !ok {
		t.Fatalf("%s has no pinned digest; add it to trnCensusDigests", name)
	}

	if got := hex.EncodeToString(digestOf(raw)); got != wantDigest {
		t.Fatalf("%s digest = %s, want %s", name, got, wantDigest)
	}

	probe := trnCensusProbeTable()
	assets.ApplyTRN(probe, trn)

	for i, want := range trn {
		if probe[i] != want {
			t.Fatalf("%s: ApplyTRN(trn, %d) = %d, want trn[%d] = %d", name, i, probe[i], i, want)
		}
	}
}

func digestOf(raw []byte) []byte {
	sum := sha256.Sum256(raw)

	return sum[:]
}

// trnCensusProbeTable is the identity index sequence: applying a table to it
// must reproduce the table itself.
func trnCensusProbeTable() []byte {
	probe := make([]byte, trnCensusSize)
	for i := range probe {
		probe[i] = byte(i)
	}

	return probe
}

// filledTRN builds a valid 256-byte table whose entries are all v.
func filledTRN(v byte) []byte {
	table := make([]byte, 256)
	for i := range table {
		table[i] = v
	}

	return table
}
