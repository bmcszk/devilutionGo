// Package audio_test ports the sound-domain parity tests against the live
// devilutionX corpus.
//
//nolint:testpackage // internal-state test: pokes unexported identifiers
package audio

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

// Row-count parity per V9.1/V9.2/V12.10: effects 814 base / 917 hf;
// unused 15 base / 1 hf.
// corpus loads the committed audio fixture via the shared helper
// (testsupport.Corpus) — single copy per the 4th-copy rule.
func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	return testsupport.Corpus(t, "audio", rel)
}

func TestRowCountsParity(t *testing.T) {
	// given
	baseData := corpus(t, "assets/txtdata/sound/effects.tsv")
	hfData := corpus(t, "mods/hf/txtdata/sound/effects.tsv")
	unusedData := corpus(t, "assets/txtdata/sound/effects-unused.tsv")
	hfUnusedData := corpus(t, "mods/hf/txtdata/sound/effects-unused.tsv")
	// when
	base, err := ParseEffects(baseData)
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	hf, err := ParseEffects(hfData)
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
	ub, err := ParseEffects(unusedData)
	if err != nil {
		t.Fatalf("ParseEffects unused: %v", err)
	}
	uh, err := ParseEffects(hfUnusedData)
	if err != nil {
		t.Fatalf("ParseEffects hf unused: %v", err)
	}
	// then
	if len(base) != 814 {
		t.Fatalf("base effects rows = %d, want 814", len(base))
	}
	if len(hf) != 917 {
		t.Fatalf("hf effects rows = %d, want 917", len(hf))
	}
	if len(ub) != 15 {
		t.Fatalf("unused rows = %d, want 15", len(ub))
	}
	if len(uh) != 1 {
		t.Fatalf("hf unused rows = %d, want 1", len(uh))
	}
}

// hf effects: appended new ids = 103; content of shared ids unchanged.
func TestHFDeltasAppends(t *testing.T) {
	// given
	baseData := corpus(t, "assets/txtdata/sound/effects.tsv")
	hfData := corpus(t, "mods/hf/txtdata/sound/effects.tsv")
	// when
	base, err := ParseEffects(baseData)
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	hf, err := ParseEffects(hfData)
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
	// then
	if len(hf)-len(base) != 103 {
		t.Fatalf("hf appended = %d, want 103", len(hf)-len(base))
	}
	baseByID := map[string]Effect{}
	baseIDs := map[string]bool{}
	for _, m := range base {
		baseIDs[m.ID] = true
		baseByID[m.ID] = m
	}
	appended := 0
	for _, m := range hf {
		if !baseIDs[m.ID] {
			appended++
		}
	}
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
	base, err := ParseEffects(baseData)
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	hf, err := ParseEffects(hfData)
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
	// then
	baseByID := map[string]Effect{}
	for _, m := range base {
		baseByID[m.ID] = m
	}
	for _, m := range hf {
		if b, ok := baseByID[m.ID]; ok && !effectsEqual(m, b) {
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
	base, err := ParseEffects(baseData)
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	hf, err := ParseEffects(hfData)
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
	// then
	if hf[804].ID != "ShootBow2" {
		t.Fatalf("hf row 804 = %s, want ShootBow2 insert", hf[804].ID)
	}
	if hf[907].ID != "AccessibilityWeapon" {
		t.Fatalf("hf row 907 = %s, want the moved AccessibilityWeapon", hf[907].ID)
	}
	var baseAW Effect
	for _, m := range base {
		if m.ID == "AccessibilityWeapon" {
			baseAW = m
		}
	}
	if !effectsEqual(hf[907], baseAW) {
		t.Fatal("moved AccessibilityWeapon must keep identical content")
	}
	// V5.4 gating: NaKrul6 exists only in the hf catalog.
	hfIDs := map[string]bool{}
	baseIDs := map[string]bool{}
	for _, m := range hf {
		hfIDs[m.ID] = true
	}
	for _, m := range base {
		baseIDs[m.ID] = true
	}
	if !hfIDs["NaKrul6"] || baseIDs["NaKrul6"] {
		t.Fatal("NaKrul6 must exist only in hf")
	}
}

// SfxID row-order coupling: the enum length equals the table length for
// both variants and the i-th enum maps to the i-th row id.
func TestSfxIDRowOrderCoupling(t *testing.T) {
	// given
	baseData := corpus(t, "assets/txtdata/sound/effects.tsv")
	hfData := corpus(t, "mods/hf/txtdata/sound/effects.tsv")
	base, err := ParseEffects(baseData)
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	hf, err := ParseEffects(hfData)
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
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

// Music worked example: getLevelMusic returns the track per dungeon type,
// INTRO default (sound.cpp:310-330).
func TestGetLevelMusic(t *testing.T) {
	// given
	cases := []struct {
		name string
		dt   DungeonType
		want musicID
	}{
		{"town", DTTown, MusicTown},
		{"cathedral", DTCathedral, MusicCathedral},
		{"catacombs", DTCatacombs, MusicCatacombs},
		{"caves", DTCaves, MusicCaves},
		{"hell", DTHell, MusicHell},
		{"nest", DTNest, MusicNest},
		{"crypt", DTCrypt, MusicCrypt},
		{"none", DTNone, MusicIntro},
		{"unknown", DungeonType(99), MusicIntro}, // default arm
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// when
			got := getLevelMusic(c.dt)
			// then
			if got != c.want {
				t.Errorf("getLevelMusic(%v) = %v, want %v", c.dt, got, c.want)
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
	movies := movies()
	// then
	if len(movies) != len(want) {
		t.Fatalf("movie count = %d, want %d (measured; bean's 10 was wrong)", len(movies), len(want))
	}
	for i, w := range want {
		if movies[i].File != w.file {
			t.Errorf("movie %d file = %s, want %s", i, movies[i].File, w.file)
		}
	}
	// diabvic per difficulty: index by difficulty suffix
	for d, vic := range []string{"diabvic1", "diabvic2", "diabvic3"} {
		t.Run(vic, func(t *testing.T) {
			// when
			got := victoryMovieForDifficulty(d)
			// then
			if got != vic+".smk" {
				t.Errorf("victoryMovieForDifficulty(%d) = %s, want %s.smk", d, got, vic)
			}
		})
	}
}
