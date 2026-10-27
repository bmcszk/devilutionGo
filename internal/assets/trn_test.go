package assets_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/assets"
)

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

// filledTRN builds a valid 256-byte table whose entries are all v.
func filledTRN(v byte) []byte {
	table := make([]byte, 256)
	for i := range table {
		table[i] = v
	}

	return table
}
