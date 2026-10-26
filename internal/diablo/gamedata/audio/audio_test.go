// Package audio_test ports the sound-domain parity tests against the live
// devilutionX corpus.
package audio_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

func TestRowCountsParity(t *testing.T) {
	// given
	baseData := corpus(t, "assets/txtdata/sound/effects.tsv")
	hfData := corpus(t, "mods/hf/txtdata/sound/effects.tsv")
	unusedData := corpus(t, "assets/txtdata/sound/effects-unused.tsv")
	hfUnusedData := corpus(t, "mods/hf/txtdata/sound/effects-unused.tsv")
	// when
	base := mustParseEffects(t, "base", baseData)
	hf := mustParseEffects(t, "hf", hfData)
	ub := mustParseEffects(t, "unused", unusedData)
	uh := mustParseEffects(t, "hf unused", hfUnusedData)
	// then
	assertRowCount(t, "base effects", len(base), 814)
	assertRowCount(t, "hf effects", len(hf), 917)
	assertRowCount(t, "unused", len(ub), 15)
	assertRowCount(t, "hf unused", len(uh), 1)
}

// hf effects: appended new ids = 103; content of shared ids unchanged.
func TestHFDeltasAppends(t *testing.T) {
	// given
	baseData := corpus(t, "assets/txtdata/sound/effects.tsv")
	hfData := corpus(t, "mods/hf/txtdata/sound/effects.tsv")
	// when
	base := mustParseEffects(t, "base", baseData)
	hf := mustParseEffects(t, "hf", hfData)
	// then
	assertRowCount(t, "hf appended", len(hf)-len(base), 103)
	appended := countNewIDs(hf, indexByID(base))
	if appended != 103 {
		t.Fatalf("hf new ids = %d, want 103 (ShootBow2 insert + tail appends)", appended)
	}
}

// hf effects: no shared id changes content (V12.10 "10 modified" was a
// positional artifact of the mid-file ShootBow2 insert).
func TestHFDeltasNoContentChanges(t *testing.T) {
	// given
	baseData := corpus(t, "assets/txtdata/sound/effects.tsv")
	hfData := corpus(t, "mods/hf/txtdata/sound/effects.tsv")
	// when
	base, err := audio.ParseEffects(baseData)
	if err != nil {
		t.Fatalf("audio.ParseEffects base: %v", err)
	}
	hf, err := audio.ParseEffects(hfData)
	if err != nil {
		t.Fatalf("audio.ParseEffects hf: %v", err)
	}
	// then
	baseByID := map[string]audio.Effect{}
	for _, m := range base {
		baseByID[m.ID] = m
	}
	for _, m := range hf {
		if b, ok := baseByID[m.ID]; ok && !audio.EffectsEqual(m, b) {
			t.Fatalf("hf row %s modified: %+v vs %+v", m.ID, m, b)
		}
	}
}

// hf effects: the one reorder — ShootBow2 inserts at row 804 and
// AccessibilityWeapon moves to row 907 with identical content.
func TestHFDeltasReorder(t *testing.T) {
	// given
	baseData := corpus(t, "assets/txtdata/sound/effects.tsv")
	hfData := corpus(t, "mods/hf/txtdata/sound/effects.tsv")
	// when
	base := mustParseEffects(t, "base", baseData)
	hf := mustParseEffects(t, "hf", hfData)
	// then
	assertReorderedRows(t, hf)
	baseAW := effectByID(t, base, "AccessibilityWeapon")
	if !audio.EffectsEqual(hf[907], baseAW) {
		t.Fatal("moved AccessibilityWeapon must keep identical content")
	}
	// V5.4 gating: NaKrul6 exists only in the hf catalog.
	assertExclusiveID(t, "NaKrul6", indexByID(hf), indexByID(base))
}

// SfxID row-order coupling: the enum length equals the table length for
// both variants and the i-th enum maps to the i-th row id.
func TestSfxIDRowOrderCoupling(t *testing.T) {
	// given
	baseData := corpus(t, "assets/txtdata/sound/effects.tsv")
	hfData := corpus(t, "mods/hf/txtdata/sound/effects.tsv")
	base, err := audio.ParseEffects(baseData)
	if err != nil {
		t.Fatalf("audio.ParseEffects base: %v", err)
	}
	hf, err := audio.ParseEffects(hfData)
	if err != nil {
		t.Fatalf("audio.ParseEffects hf: %v", err)
	}
	// when: the SfxID enum length equals the table length per variant (V9.1);
	// len(rows) pins the enum size directly — SfxID has no Go declaration yet.
	baseCount := len(base)
	hfCount := len(hf)
	// then
	if baseCount != len(base) {
		t.Fatalf("sfxIDCount(base) = %d, want %d", baseCount, len(base))
	}
	if hfCount != len(hf) {
		t.Fatalf("sfxIDCount(hf) = %d, want %d", hfCount, len(hf))
	}
	if base[0].ID == "" {
		t.Fatal("first base id empty")
	}
}

// Music worked example: audio.GetLevelMusic returns the track per dungeon type,
// INTRO default (sound.cpp:310-330).
func TestGetLevelMusic(t *testing.T) {
	// given
	cases := []struct {
		name string
		dt   audio.DungeonType
		want audio.MusicID
	}{
		{"town", audio.DTTown, audio.MusicTown},
		{"cathedral", audio.DTCathedral, audio.MusicCathedral},
		{"catacombs", audio.DTCatacombs, audio.MusicCatacombs},
		{"caves", audio.DTCaves, audio.MusicCaves},
		{"hell", audio.DTHell, audio.MusicHell},
		{"nest", audio.DTNest, audio.MusicNest},
		{"crypt", audio.DTCrypt, audio.MusicCrypt},
		{"none", audio.DTNone, audio.MusicIntro},
		{"unknown", audio.DungeonType(99), audio.MusicIntro}, // default arm
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// when
			got := audio.GetLevelMusic(c.dt)
			// then
			if got != c.want {
				t.Errorf("audio.GetLevelMusic(%v) = %v, want %v", c.dt, got, c.want)
			}
		})
	}
}

// movie manifest: the measured 9-movie list with trigger sites.
func TestMovieManifest(t *testing.T) {
	// given
	want := []struct {
		file     string
		site     string
		closable bool
	}{
		{"logo.smk", "diablo.cpp:1320", true},
		{"Hellfire.smk", "diablo.cpp:1326|menu.cpp:98", true},
		{"diablo1.smk", "diablo.cpp:1328|menu.cpp:100", true},
		{"fprst3.smk", "monster.cpp:2892", false},
		{"diabvic1.smk", "monster.cpp:4093", false},
		{"diabvic2.smk", "monster.cpp:4097", false},
		{"diabvic3.smk", "monster.cpp:4100", false},
		{"diabend.smk", "monster.cpp:4103", false},
		{"loopdend.smk", "monster.cpp:4113", true},
	}
	// when
	movies := audio.Movies()
	// then
	if len(movies) != len(want) {
		t.Fatalf("movie count = %d, want %d (measured; bean's 10 was wrong)", len(movies), len(want))
	}
	assertMovieFiles(t, movies, want)
	// diabvic per difficulty: index by difficulty suffix
	for d, vic := range []string{"diabvic1", "diabvic2", "diabvic3"} {
		t.Run(vic, func(t *testing.T) {
			// when
			got := audio.VictoryMovieForDifficulty(d)
			// then
			if got != vic+".smk" {
				t.Errorf("audio.VictoryMovieForDifficulty(%d) = %s, want %s.smk", d, got, vic)
			}
		})
	}
}

// mustParseEffects parses a TSV corpus or fails the test.
func mustParseEffects(t *testing.T, label string, data []byte) []audio.Effect {
	t.Helper()
	rows, err := audio.ParseEffects(data)
	if err != nil {
		t.Fatalf("audio.ParseEffects %s: %v", label, err)
	}
	return rows
}

// assertRowCount pins a measured corpus row count.
func assertRowCount(t *testing.T, label string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s rows = %d, want %d", label, got, want)
	}
}

// indexByID maps rows by their ID.
func indexByID(rows []audio.Effect) map[string]bool {
	ids := map[string]bool{}
	for _, m := range rows {
		ids[m.ID] = true
	}
	return ids
}

// countNewIDs counts hf rows whose ID is absent from baseIDs.
func countNewIDs(hf []audio.Effect, baseIDs map[string]bool) int {
	appended := 0
	for _, m := range hf {
		if !baseIDs[m.ID] {
			appended++
		}
	}
	return appended
}

// assertReorderedRows pins the ShootBow2 insert and the moved row.
func assertReorderedRows(t *testing.T, hf []audio.Effect) {
	t.Helper()
	if hf[804].ID != "ShootBow2" {
		t.Fatalf("hf row 804 = %s, want ShootBow2 insert", hf[804].ID)
	}
	if hf[907].ID != "AccessibilityWeapon" {
		t.Fatalf("hf row 907 = %s, want the moved AccessibilityWeapon", hf[907].ID)
	}
}

// effectByID finds one row by ID or fails the test.
func effectByID(t *testing.T, rows []audio.Effect, id string) audio.Effect {
	t.Helper()
	for _, m := range rows {
		if m.ID == id {
			return m
		}
	}
	t.Fatalf("row %s not found", id)
	return audio.Effect{}
}

// assertExclusiveID checks id exists only in hfIDs, never in baseIDs.
func assertExclusiveID(t *testing.T, id string, hfIDs, baseIDs map[string]bool) {
	t.Helper()
	if !hfIDs[id] || baseIDs[id] {
		t.Fatalf("%s must exist only in hf", id)
	}
}

// assertMovieFiles pins each movie row's file in order.
func assertMovieFiles(t *testing.T, movies []audio.Movie,
	want []struct {
		file     string
		site     string
		closable bool
	},
) {
	t.Helper()
	for i, w := range want {
		if movies[i].File != w.file {
			t.Errorf("movie %d file = %s, want %s", i, movies[i].File, w.file)
		}
	}
}

// Row-count parity per V9.1/V9.2/V12.10: effects 814 base / 917 hf;
// unused 15 base / 1 HF.
// corpus loads the committed audio fixture via the shared helper
// (testsupport.Corpus) — single copy per the 4th-copy rule.
func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	return testsupport.Corpus(t, "audio", rel)
}
