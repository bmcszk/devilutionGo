package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// randomUniqueFixture scripts the RandomUnique inputs and records the calls.
type randomUniqueFixture struct {
	rnd       *tapeRnd
	uniques   []itemdat.UniqueItemData
	dropped   map[int]bool
	multi     bool
	discarded int
	seeds     int
	setups    []itemgen.SetupRequest
	plans     []itemgen.SetupPlan
}

// TestRandomUniqueSkippedWhenNotUnique pins the CreateInfo guard of
// TryRandomUniqueItem: without CF_UNIQUE there is nothing to do
// (items.cpp:3316-3318).
// TestRandomUniqueSkippedWhenUniqueMisc pins the IMISC_UNIQUE term
// (items.cpp:3316-3318).
// TestRandomUniqueFallbackWhenAllDropped pins the uids-empty branch: uper drops
// to 1, a uper-15 level gains +4, and SetupAllItems reruns until the roll
// misses the unique quality (items.cpp:3337-3346).
// TestRandomUniqueReverseCompatibleLoop pins the reverse-compatible regen path:
// onlygood is forced and the level target comes from the unique's own min level
// (items.cpp:3387-3406).
// TestRandomUniqueOffsetPath pins the non-reverse-compatible branch: one rerun
// with the uid offset, then the CF_UIDOFFSET tag (items.cpp:3407-3416).

func TestRandomUniqueSkippedWhenNotUnique(t *testing.T) {
	// given
	f := randomUniqueEnv(t)

	// when
	out, err := itemgen.RandomUnique(f.env(), itemgen.RandomUniqueRequest{ItemIndex: 0})
	// then
	if err != nil {
		t.Fatalf("RandomUnique: %v", err)
	}
	if !out.Skipped || !f.untouched() {
		t.Fatalf("out = %+v draws = %v, want an untouched skip", out, f.draws())
	}
}

func TestRandomUniqueSkippedWhenUniqueMisc(t *testing.T) {
	// given
	f := randomUniqueEnv(t)
	req := itemgen.RandomUniqueRequest{CreateInfo: itemgen.CreateInfoUnique, MiscID: "UNIQUE"}

	// when
	out, err := itemgen.RandomUnique(f.env(), req)
	// then
	if err != nil {
		t.Fatalf("RandomUnique: %v", err)
	}
	if !out.Skipped || f.discarded != 0 {
		t.Fatalf("out = %+v discarded = %d, want a skip before the discard", out, f.discarded)
	}
}

func TestRandomUniqueFallbackWhenAllDropped(t *testing.T) {
	// given
	f := randomUniqueEnv(t)
	f.dropped = map[int]bool{0: true, 1: true, 2: true}
	req := itemgen.RandomUniqueRequest{
		CreateInfo: itemgen.CreateInfoUnique, BaseItemID: "SWORD", Level: 10, Uper: 15,
	}
	f.plans = []itemgen.SetupPlan{{Quality: itemgen.QualityUnique}, {Quality: itemgen.QualityNormal}}

	// when
	out, err := itemgen.RandomUnique(f.env(), req)
	// then
	if err != nil {
		t.Fatalf("RandomUnique: %v", err)
	}
	if !out.Fallback || out.UID != itemgen.UniqueNone || out.Regenerated != 2 {
		t.Fatalf("out = %+v, want the fallback to stop on the non-unique rerun", out)
	}
	assertFallbackSetup(t, f.setups[0])
}

func TestRandomUniqueReverseCompatibleLoop(t *testing.T) {
	// given
	f := randomUniqueEnv(t)
	req := itemgen.RandomUniqueRequest{
		CreateInfo: itemgen.CreateInfoUnique, BaseItemID: "SWORD", Level: 10, Uper: 1,
	}
	f.rnd.values = []int32{0, 1, 0, 0}
	f.plans = []itemgen.SetupPlan{{UID: 0}, {UID: 1}}

	// when
	out, err := itemgen.RandomUnique(f.env(), req)
	// then
	if err != nil {
		t.Fatalf("RandomUnique: %v", err)
	}
	if !out.ReverseCompat || out.UID != 1 || out.Regenerated != 2 {
		t.Fatalf("out = %+v, want uid 1 after two reruns", out)
	}
	assertReverseSetup(t, f.setups[0], 1, 3)
}

func TestRandomUniqueOffsetPath(t *testing.T) {
	// given
	f := randomUniqueEnv(t)
	req := itemgen.RandomUniqueRequest{
		CreateInfo: itemgen.CreateInfoUnique, BaseItemID: "SWORD", Level: 10, Uper: 1, Seed: 77,
	}

	// when
	out, err := itemgen.RandomUnique(f.env(), req)
	// then
	if err != nil {
		t.Fatalf("RandomUnique: %v", err)
	}
	if out.ReverseCompat || out.UID != 0 || out.Offset != 2 || out.Regenerated != 0 {
		t.Fatalf("out = %+v, want uid 0 at offset 2 without a regen loop", out)
	}
	if len(f.setups) != 1 || f.setups[0].UidOffset != 2 || f.setups[0].Seed != 77 {
		t.Fatalf("setups = %+v, want one rerun with the original seed and offset 2", f.setups)
	}
	if want := (2 << 1) & itemgen.BuffUIDOffsetMask; out.Plan.Buff != want {
		t.Fatalf("Buff = %d, want %d (uidOffset << 1)", out.Plan.Buff, want)
	}
}

func randomUniqueEnv(t *testing.T) *randomUniqueFixture {
	t.Helper()

	return &randomUniqueFixture{
		rnd:     &tapeRnd{t: t, values: []int32{0, 0, 0, 0, 0, 0}},
		uniques: offsetUniques(),
		dropped: map[int]bool{},
	}
}

func (f *randomUniqueFixture) env() itemgen.RandomUniqueEnv {
	return itemgen.RandomUniqueEnv{
		Uniques:     f.uniques,
		Dropped:     f.dropped,
		Multiplayer: f.multi,
		Rnd:         f.rnd.rnd,
		Discard:     func(n int) { f.discarded += n },
		NextSeed:    f.nextSeed,
		Setup:       f.setup,
	}
}

func (f *randomUniqueFixture) setup(req itemgen.SetupRequest) (itemgen.SetupPlan, error) {
	f.setups = append(f.setups, req)
	if len(f.plans) == 0 {
		return itemgen.SetupPlan{}, nil
	}
	plan := f.plans[0]
	f.plans = f.plans[1:]

	return plan, nil
}

func (f *randomUniqueFixture) nextSeed() uint32 {
	f.seeds++

	return uint32(f.seeds)
}

func (f *randomUniqueFixture) untouched() bool {
	return len(f.setups) == 0 && f.discarded == 0
}

func (f *randomUniqueFixture) draws() []int32 {
	return f.rnd.args
}

func assertFallbackSetup(t *testing.T, req itemgen.SetupRequest) {
	t.Helper()

	if req.Level != 14 || req.Uper != 1 || req.OnlyGood || req.Seed != 1 {
		t.Fatalf("setup = %+v, want level 14 (10+4), uper 1, onlygood false, seed 1", req)
	}
}

func assertReverseSetup(t *testing.T, req itemgen.SetupRequest, wantUper, wantLevel int) {
	t.Helper()

	if !req.OnlyGood {
		t.Fatalf("setup = %+v, want onlygood forced for the reverse-compatible path", req)
	}
	if req.Uper != wantUper || req.Level != wantLevel || req.UidOffset != 0 {
		t.Fatalf("setup = %+v, want uper %d level %d offset 0", req, wantUper, wantLevel)
	}
}
