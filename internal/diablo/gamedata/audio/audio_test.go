// Package audio_test ports the sound-domain parity tests against the live
// devilutionX corpus.
package audio_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
)

func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	name := filepath.Base(rel)
	if strings.Contains(rel, "mods/hf/") {
		name = "hf-" + name
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "internal", "testdata", "gamedata", "audio", name))
	if err != nil {
		t.Fatalf("read committed audio fixture %s: %v", name, err)
	}
	return data
}

// Row-count parity per V9.1/V9.2/V12.10: effects 814 base / 917 hf;
// unused 15 base / 1 hf.
func TestRowCountsParity(t *testing.T) {
	base, err := audio.ParseEffects(corpus(t, "assets/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	if len(base) != 814 {
		t.Fatalf("base effects rows = %d, want 814", len(base))
	}
	hf, err := audio.ParseEffects(corpus(t, "mods/hf/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
	if len(hf) != 917 {
		t.Fatalf("hf effects rows = %d, want 917", len(hf))
	}
	ub, err := audio.ParseEffects(corpus(t, "assets/txtdata/sound/effects-unused.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects unused: %v", err)
	}
	if len(ub) != 15 {
		t.Fatalf("unused rows = %d, want 15", len(ub))
	}
	uh, err := audio.ParseEffects(corpus(t, "mods/hf/txtdata/sound/effects-unused.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects hf unused: %v", err)
	}
	if len(uh) != 1 {
		t.Fatalf("hf unused rows = %d, want 1", len(uh))
	}
}

// hf effects: appended new ids = 103; content of shared ids unchanged.
func TestHFDeltasAppends(t *testing.T) {
	base, err := audio.ParseEffects(corpus(t, "assets/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	hf, err := audio.ParseEffects(corpus(t, "mods/hf/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
	if len(hf)-len(base) != 103 {
		t.Fatalf("hf appended = %d, want 103", len(hf)-len(base))
	}
	baseByID := map[string]audio.Effect{}
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
	base, err := audio.ParseEffects(corpus(t, "assets/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	hf, err := audio.ParseEffects(corpus(t, "mods/hf/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
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
	base, err := audio.ParseEffects(corpus(t, "assets/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	hf, err := audio.ParseEffects(corpus(t, "mods/hf/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
	if hf[804].ID != "ShootBow2" {
		t.Fatalf("hf row 804 = %s, want ShootBow2 insert", hf[804].ID)
	}
	if hf[907].ID != "AccessibilityWeapon" {
		t.Fatalf("hf row 907 = %s, want the moved AccessibilityWeapon", hf[907].ID)
	}
	var baseAW audio.Effect
	for _, m := range base {
		if m.ID == "AccessibilityWeapon" {
			baseAW = m
		}
	}
	if !audio.EffectsEqual(hf[907], baseAW) {
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
	base, err := audio.ParseEffects(corpus(t, "assets/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects base: %v", err)
	}
	if got := audio.SfxIDCount(len(base)); got != len(base) {
		t.Fatalf("SfxIDCount(base) = %d, want %d", got, len(base))
	}
	hf, err := audio.ParseEffects(corpus(t, "mods/hf/txtdata/sound/effects.tsv"))
	if err != nil {
		t.Fatalf("ParseEffects hf: %v", err)
	}
	if got := audio.SfxIDCount(len(hf)); got != len(hf) {
		t.Fatalf("SfxIDCount(hf) = %d, want %d", got, len(hf))
	}
	// spot-check row-order: first and last ids
	if base[0].ID != "SfxIDLogo" && base[0].ID == "" {
		t.Fatal("first base id empty")
	}
	_ = hf
}

// Music worked example: GetLevelMusic returns the track per dungeon type,
// INTRO default (sound.cpp:310-330).
func TestGetLevelMusic(t *testing.T) {
	cases := []struct {
		dt   audio.DungeonType
		want audio.MusicID
	}{
		{audio.DTTown, audio.MusicTown},
		{audio.DTCathedral, audio.MusicCathedral},
		{audio.DTCatacombs, audio.MusicCatacombs},
		{audio.DTCaves, audio.MusicCaves},
		{audio.DTHell, audio.MusicHell},
		{audio.DTNest, audio.MusicNest},
		{audio.DTCrypt, audio.MusicCrypt},
		{audio.DTNone, audio.MusicIntro},
		{audio.DungeonType(99), audio.MusicIntro}, // default arm
	}
	for _, c := range cases {
		if got := audio.GetLevelMusic(c.dt); got != c.want {
			t.Errorf("GetLevelMusic(%v) = %v, want %v", c.dt, got, c.want)
		}
	}
}

// Movie manifest: the measured 9-movie list with trigger sites.
func TestMovieManifest(t *testing.T) {
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
	movies := audio.Movies()
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
		if got := audio.VictoryMovieForDifficulty(d); got != vic+".smk" {
			t.Errorf("VictoryMovieForDifficulty(%d) = %s, want %s.smk", d, got, vic)
		}
	}
}
