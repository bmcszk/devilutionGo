package misdat_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

// parseMisdat parses corpus or fails the test.
func parseMisdat(t *testing.T, rel string) []misdat.MissileData {
	t.Helper()
	rows, err := misdat.ParseMisdat(corpus(t, rel))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat %s: %v", rel, err)
	}
	return rows
}

// parseSprites parses a missile_sprites.tsv copy or fails the test.
func parseSprites(t *testing.T, rel string) []misdat.MissileSprite {
	t.Helper()
	rows, err := misdat.ParseMissileSprites(corpus(t, rel))
	if err != nil {
		t.Fatalf("misdat.ParseMissileSprites %s: %v", rel, err)
	}
	return rows
}

// wantRows fails unless rows has the wanted length.
func wantRows(t *testing.T, what string, rows, want int) {
	t.Helper()
	if rows != want {
		t.Fatalf("%s rows = %d, want %d", what, rows, want)
	}
}

func TestRowCountsParity(t *testing.T) {
	// when
	base := parseMisdat(t, "assets/txtdata/missiles/misdat.tsv")
	hf := parseMisdat(t, "mods/hf/txtdata/missiles/misdat.tsv")
	sb := parseSprites(t, "assets/txtdata/missiles/missile_sprites.tsv")
	sh := parseSprites(t, "mods/hf/txtdata/missiles/missile_sprites.tsv")

	// then
	wantRows(t, "base misdat", len(base), 68)
	wantRows(t, "hf misdat", len(hf), 108)
	wantRows(t, "base sprite", len(sb), 46)
	wantRows(t, "hf sprite", len(sh), 59)
}

func TestHFMisdatStrictSuperset(t *testing.T) {
	// given
	base := parseMisdat(t, "assets/txtdata/missiles/misdat.tsv")

	// when
	hf := parseMisdat(t, "mods/hf/txtdata/missiles/misdat.tsv")

	// then
	for i := range base {
		if hfRowDiffers(base[i], hf[i]) {
			t.Fatalf("hf row %d (%s) modifies base row", i, base[i].ID)
		}
	}
	wantRows(t, "hf appended", len(hf)-len(base), 40)
}

// hfRowDiffers reports whether the hf copy of base row i changed any field.
func hfRowDiffers(base, hf misdat.MissileData) bool {
	return hf.ID != base.ID || hf.AddFn != base.AddFn || hf.ProcessFn != base.ProcessFn ||
		hf.CastSound != base.CastSound || hf.HitSound != base.HitSound ||
		hf.Graphic != base.Graphic || hf.MovementDistribution != base.MovementDistribution ||
		strings.Join(hf.Flags, ",") != strings.Join(base.Flags, ",")
}

func TestRegistryCompleteness(t *testing.T) {
	base := parseMisdat(t, "assets/txtdata/missiles/misdat.tsv")
	adds, procs := misdatNames(base)
	// control rows carry an empty addFn upstream (nullptr)
	delete(adds, "")
	wantRows(t, "base distinct addFn (excl. empty)", len(adds), 50)
	delete(procs, "")
	wantRows(t, "base distinct processFn (excl. empty)", len(procs), 36)
	checkRegistered(t, "addFn", adds, misdat.AddRegistry)
	checkReferenced(t, "processFn", procs, misdat.ProcessRegistry)
	for _, name := range misdat.BaseProcessNames {
		if !procs[name] {
			t.Errorf("processFn %q registered but not referenced by any base row", name)
		}
	}
}

// checkReferenced fails on referenced names missing from the registry.
func checkReferenced[V any](t *testing.T, what string, referenced map[string]bool, registry map[string]V) {
	t.Helper()
	for name := range referenced {
		if _, ok := registry[name]; !ok {
			t.Errorf("%s %q referenced by data but not registered", what, name)
		}
	}
}

// misdatNames collects the distinct addFn/processFn column values of rows.
func misdatNames(rows []misdat.MissileData) (adds, procs map[string]bool) {
	adds = map[string]bool{}
	procs = map[string]bool{}
	for _, m := range rows {
		adds[m.AddFn] = true
		procs[m.ProcessFn] = true
	}
	return adds, procs
}

// checkRegistered fails on any mismatch between the names a data corpus
// references and the names a registry implements, in both directions.
func checkRegistered[V any](t *testing.T, what string, referenced map[string]bool, registry map[string]V) {
	t.Helper()
	for name := range referenced {
		if _, ok := registry[name]; !ok {
			t.Errorf("%s %q referenced by data but not registered", what, name)
		}
	}
	for name := range registry {
		if !referenced[name] {
			t.Errorf("%s %q registered but not referenced by any row", what, name)
		}
	}
}

func TestAddOpenNestSpawnsFourBigExplosions(t *testing.T) {
	// given
	hf := parseMisdat(t, "mods/hf/txtdata/missiles/misdat.tsv")
	if hf[97].ID != "OpenNest" || hf[97].AddFn != "AddOpenNest" {
		t.Fatalf("hf data row 97 = %s/%s, want OpenNest/AddOpenNest", hf[97].ID, hf[97].AddFn)
	}
	w := NewStubWorld()
	self := w.Spawn("OpenNest", 80, 62)

	// when
	misdat.AddOpenNest(w, self)

	// then
	got := w.Missiles()
	wantRows(t, "spawned", len(got), 4)
	checkBigExplosionSpawns(t, got)
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
	rows, err := misdat.ParseMissileSprites(corpus(t, "assets/txtdata/missiles/missile_sprites.tsv"))
	if err != nil {
		t.Fatalf("misdat.ParseMissileSprites: %v", err)
	}
	for _, s := range rows {
		if len(s.FrameDelay) != 16 || len(s.FrameLength) != 16 {
			t.Fatalf("%s: frame arrays not 16 wide", s.ID)
		}
		checkDedup(t, s.ID, s.FrameDelay)
	}
}

// checkBigExplosionSpawns verifies the four BigExplosion positions AddOpenNest
// spawns (Hive niche corners).
func checkBigExplosionSpawns(t *testing.T, got []*misdat.MissileRef) {
	t.Helper()
	want := [][2]int{{80, 62}, {81, 62}, {80, 63}, {81, 63}}
	for i, p := range want {
		if got[i].Kind != "BigExplosion" || got[i].X != p[0] || got[i].Y != p[1] {
			t.Errorf("spawn %d = %s@(%d,%d), want BigExplosion@(%d,%d)", i, got[i].Kind, got[i].X, got[i].Y, p[0], p[1])
		}
	}
}

// checkDedup verifies the run-collapse of one row's delay table.
func checkDedup(t *testing.T, id string, delay []int) {
	t.Helper()
	// when
	dd := misdat.ToIndex(delay)

	// then
	if len(dd.Values()) != countRuns(delay) && !slices.Equal(dd.Values(), distinct(delay)) {
		t.Fatalf("%s: delay dedup %v wrong for %v", id, dd.Values(), delay)
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
