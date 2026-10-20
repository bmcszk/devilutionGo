package misdat_test

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"

	"os"
	"path/filepath"
	"strings"
	"testing"
)

func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	name := filepath.Base(rel)
	if strings.Contains(rel, "mods/hf/") {
		name = "hf-" + name
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "internal", "testdata", "gamedata", "misdat", name))
	if err != nil {
		t.Fatalf("read committed misdat fixture %s: %v", name, err)
	}
	return data
}

// Row-count parity per V12.12: misdat 68 base / 108 hf; sprites 46 base / 59 hf.
func TestRowCountsParity(t *testing.T) {
	base, err := misdat.ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat base: %v", err)
	}
	if len(base) != 68 {
		t.Fatalf("base misdat rows = %d, want 68", len(base))
	}
	hf, err := misdat.ParseMisdat(corpus(t, "mods/hf/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat hf: %v", err)
	}
	if len(hf) != 108 {
		t.Fatalf("hf misdat rows = %d, want 108", len(hf))
	}
	sb, err := misdat.ParseMissileSprites(corpus(t, "assets/txtdata/missiles/missile_sprites.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMissileSprites base: %v", err)
	}
	if len(sb) != 46 {
		t.Fatalf("base sprite rows = %d, want 46", len(sb))
	}
	sh, err := misdat.ParseMissileSprites(corpus(t, "mods/hf/txtdata/missiles/missile_sprites.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMissileSprites hf: %v", err)
	}
	if len(sh) != 59 {
		t.Fatalf("hf sprite rows = %d, want 59", len(sh))
	}
}

// hf misdat is a strict superset: +40 appended, 0 modified (V12.10).
func TestHFMisdatStrictSuperset(t *testing.T) {
	base, err := misdat.ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat base: %v", err)
	}
	hf, err := misdat.ParseMisdat(corpus(t, "mods/hf/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat hf: %v", err)
	}
	baseIDs := map[string]bool{}
	for _, m := range base {
		baseIDs[m.ID] = true
	}
	// Shared prefix must be row-identical; the rest appended-only.
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

// Registry completeness (V3.2, measured): base 51 add / 37 process; hf 74/46.
// Every name referenced by a row resolves to a registered implementation and
// vice versa.
func TestRegistryCompleteness(t *testing.T) {
	base, err := misdat.ParseMisdat(corpus(t, "assets/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat base: %v", err)
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
		if _, ok := misdat.AddRegistry[name]; !ok {
			t.Errorf("addFn %q referenced by data but not registered", name)
		}
	}
	for name := range misdat.AddRegistry {
		if !adds[name] {
			t.Errorf("addFn %q registered but not referenced by any base row", name)
		}
	}
	for name := range procs {
		if _, ok := misdat.ProcessRegistry[name]; !ok {
			t.Errorf("processFn %q referenced by data but not registered", name)
		}
	}
	for _, name := range misdat.BaseProcessNames {
		if !procs[name] {
			t.Errorf("processFn %q registered but not referenced by any base row", name)
		}
	}
}

// Worked example: AddOpenNest (hf misdat row 99; missiles.cpp:1238-1246)
// spawns 4 BigExplosion at (80,62)(81,62)(80,63)(81,63), self-deletes.
func TestAddOpenNestSpawnsFourBigExplosions(t *testing.T) {
	hf, err := misdat.ParseMisdat(corpus(t, "mods/hf/txtdata/missiles/misdat.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat hf: %v", err)
	}
	if hf[97].ID != "OpenNest" || hf[97].AddFn != "AddOpenNest" {
		t.Fatalf("hf data row 97 = %s/%s, want OpenNest/AddOpenNest", hf[97].ID, hf[97].AddFn)
	}
	w := NewStubWorld()
	self := w.Spawn("OpenNest", 80, 62)
	misdat.AddOpenNest(w, self)
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
	_ = strings.TrimSpace
}

// Sprite array de-dup: misdat.ToIndex equivalence — identical consecutive entries share one index (cpp:47-55).
func TestSpriteToIndexDedup(t *testing.T) {
	rows, err := misdat.ParseMissileSprites(corpus(t, "assets/txtdata/missiles/missile_sprites.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMissileSprites: %v", err)
	}
	for _, s := range rows {
		if len(s.FrameDelay) != 16 || len(s.FrameLength) != 16 {
			t.Fatalf("%s: frame arrays not 16 wide", s.ID)
		}
		// de-dup: distinct values must map to a compacted table preserving order
		dd := misdat.ToIndex(s.FrameDelay)
		if len(dd.Values()) != countRuns(s.FrameDelay) && !equalDistinct(dd.Values(), distinct(s.FrameDelay)) {
			t.Fatalf("%s: delay dedup %v wrong for %v", s.ID, dd.Values(), s.FrameDelay)
		}
	}
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

func equalDistinct(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
