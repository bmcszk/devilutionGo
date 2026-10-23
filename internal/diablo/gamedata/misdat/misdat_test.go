//nolint:testpackage // internal-state test: pokes unexported identifiers
package misdat

import (
	"slices"
	"testing"

	"strings"

	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

func TestRowCountsParity(t *testing.T) {
	// when
	base, baseErr := ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	hf, hfErr := ParseMisdat(corpus(t, "mods/hf/txtdata/missiles/misdat.tsv"))
	sb, sbErr := parseMissileSprites(corpus(t, "assets/txtdata/missiles/missile_sprites.tsv"))
	sh, shErr := parseMissileSprites(corpus(t, "mods/hf/txtdata/missiles/missile_sprites.tsv"))

	// then
	if baseErr != nil {
		t.Fatalf("ParseMisdat base: %v", baseErr)
	}
	if len(base) != 68 {
		t.Fatalf("base misdat rows = %d, want 68", len(base))
	}
	if hfErr != nil {
		t.Fatalf("ParseMisdat hf: %v", hfErr)
	}
	if len(hf) != 108 {
		t.Fatalf("hf misdat rows = %d, want 108", len(hf))
	}
	if sbErr != nil {
		t.Fatalf("parseMissileSprites base: %v", sbErr)
	}
	if len(sb) != 46 {
		t.Fatalf("base sprite rows = %d, want 46", len(sb))
	}
	if shErr != nil {
		t.Fatalf("parseMissileSprites hf: %v", shErr)
	}
	if len(sh) != 59 {
		t.Fatalf("hf sprite rows = %d, want 59", len(sh))
	}
}

func TestHFMisdatStrictSuperset(t *testing.T) {
	// given
	base, err := ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMisdat base: %v", err)
	}

	// when
	hf, err := ParseMisdat(corpus(t, "mods/hf/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMisdat hf: %v", err)
	}

	// then
	baseIDs := map[string]bool{}
	for _, m := range base {
		baseIDs[m.ID] = true
	}
	for i := range base {
		if hf[i].ID != base[i].ID || hf[i].AddFn != base[i].AddFn || hf[i].ProcessFn != base[i].ProcessFn ||
			hf[i].CastSound != base[i].CastSound || hf[i].HitSound != base[i].HitSound ||
			hf[i].Graphic != base[i].Graphic || hf[i].MovementDistribution != base[i].MovementDistribution ||
			strings.Join(hf[i].Flags, ",") != strings.Join(base[i].Flags, ",") {
			t.Fatalf("hf row %d (%s) modifies base row", i, base[i].ID)
		}
	}
	if len(hf)-len(base) != 40 {
		t.Fatalf("hf appended = %d, want 40", len(hf)-len(base))
	}
}

func TestRegistryCompleteness(t *testing.T) {
	base, err := ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMisdat base: %v", err)
	}
	adds := map[string]bool{}
	procs := map[string]bool{}
	for _, m := range base {
		adds[m.AddFn] = true
		procs[m.ProcessFn] = true
	}
	delete(adds, "") // control rows carry an empty addFn upstream (nullptr)
	if len(adds) != 50 {
		t.Fatalf("base distinct addFn = %d, want 50 (excl. empty)", len(adds))
	}
	delete(procs, "")
	if len(procs) != 36 {
		t.Fatalf("base distinct processFn = %d, want 36 (excl. empty)", len(procs))
	}
	for name := range adds {
		if _, ok := AddRegistry[name]; !ok {
			t.Errorf("addFn %q referenced by data but not registered", name)
		}
	}
	for name := range AddRegistry {
		if !adds[name] {
			t.Errorf("addFn %q registered but not referenced by any base row", name)
		}
	}
	for name := range procs {
		if _, ok := ProcessRegistry[name]; !ok {
			t.Errorf("processFn %q referenced by data but not registered", name)
		}
	}
	for _, name := range BaseProcessNames {
		if !procs[name] {
			t.Errorf("processFn %q registered but not referenced by any base row", name)
		}
	}
}

func TestAddOpenNestSpawnsFourBigExplosions(t *testing.T) {
	// given
	hf, err := ParseMisdat(corpus(t, "mods/hf/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMisdat hf: %v", err)
	}
	if hf[97].ID != "OpenNest" || hf[97].AddFn != "AddOpenNest" {
		t.Fatalf("hf data row 97 = %s/%s, want OpenNest/AddOpenNest", hf[97].ID, hf[97].AddFn)
	}
	w := NewStubWorld()
	self := w.Spawn("OpenNest", 80, 62)

	// when
	addOpenNest(w, self)

	// then
	got := w.Missiles()
	if len(got) != 4 {
		t.Fatalf("spawned %d, want 4", len(got))
	}
	want := [][2]int{{80, 62}, {81, 62}, {80, 63}, {81, 63}}
	for i, p := range want {
		if got[i].Kind != "BigExplosion" || got[i].X != p[0] || got[i].Y != p[1] {
			t.Errorf("spawn %d = %s@(%d,%d), want BigExplosion@(%d,%d)", i, got[i].Kind, got[i].X, got[i].Y, p[0], p[1])
		}
	}
	if !self.Deleted {
		t.Fatal("OpenNest missile must self-delete")
	}
	dl := w.Deleted()
	if len(dl) != 1 || dl[0] != self {
		t.Fatalf("deleted = %v, want exactly the self missile", dl)
	}
}

func TestSpriteToIndexDedup(t *testing.T) {
	// given
	rows, err := parseMissileSprites(corpus(t, "assets/txtdata/missiles/missile_sprites.tsv"))
	if err != nil {
		t.Fatalf("parseMissileSprites: %v", err)
	}
	for _, s := range rows {
		if len(s.FrameDelay) != 16 || len(s.FrameLength) != 16 {
			t.Fatalf("%s: frame arrays not 16 wide", s.ID)
		}

		// when
		dd := toIndex(s.FrameDelay)

		// then
		if len(dd.Values()) != countRuns(s.FrameDelay) && !slices.Equal(dd.Values(), distinct(s.FrameDelay)) {
			t.Fatalf("%s: delay dedup %v wrong for %v", s.ID, dd.Values(), s.FrameDelay)
		}
	}
}

// corpus loads the committed misdat fixture via the shared helper
// (testsupport.Corpus) — single copy per the 4th-copy rule.
func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	return testsupport.Corpus(t, "misdat", rel)
}

func countRuns(v []int) int {
	if len(v) == 0 {
		return 0
	}
	n := 1
	for i := 1; i < len(v); i++ {
		if v[i] != v[i-1] {
			n++
		}
	}
	return n
}

func distinct(v []int) []int {
	seen := map[int]bool{}
	var out []int
	for _, x := range v {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
