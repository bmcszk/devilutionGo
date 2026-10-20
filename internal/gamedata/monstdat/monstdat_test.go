// Package monstdat tests port the monster data tables against the live
// devilutionX corpus (row counts, hf superset shape, AI registry parity).
package monstdat_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/gamedata/monstdat"
)

// corpusRoot resolves the devilutionX checkout that holds the parity corpus.
func corpusRoot(t *testing.T) string {
	t.Helper()
	devilutionX := os.Getenv("DG_DEVILUTIONX")
	if devilutionX == "" {
		t.Skip("set DG_DEVILUTIONX to the devilutionX checkout to run this parity test")
	}
	return devilutionX
}

func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(corpusRoot(t), rel))
	if err != nil {
		t.Skipf("corpus not available: %v", err)
	}
	return data
}

// Row-count parity per V12.12: monstdat 112 base / 138 hf; unique 100 = hf.
func TestRowCountsParity(t *testing.T) {
	base, err := monstdat.ParseMonstdat(corpus(t, "assets/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat base: %v", err)
	}
	if len(base) != 112 {
		t.Fatalf("base monstdat rows = %d, want 112", len(base))
	}
	hf, err := monstdat.ParseMonstdat(corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat hf: %v", err)
	}
	if len(hf) != 138 {
		t.Fatalf("hf monstdat rows = %d, want 138", len(hf))
	}
	uniq, err := monstdat.ParseUniqueMonstdat(corpus(t, "assets/txtdata/monsters/unique_monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseUniqueMonstdat: %v", err)
	}
	if len(uniq) != 100 {
		t.Fatalf("unique rows = %d, want 100", len(uniq))
	}
}

// hf monstdat: first 112 rows identical, +26 appended, 0 modified (V12.10).
func TestHFMisdatStrictSuperset(t *testing.T) {
	base, err := monstdat.ParseMonstdat(corpus(t, "assets/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat base: %v", err)
	}
	hf, err := monstdat.ParseMonstdat(corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat hf: %v", err)
	}
	if len(hf)-len(base) != 26 {
		t.Fatalf("hf appended = %d, want 26", len(hf)-len(base))
	}
	baseIDs := map[string]bool{}
	for _, m := range base {
		baseIDs[m.ID] = true
	}
	for i := range base {
		if hf[i].ID != base[i].ID || strings.Join(hf[i].Raw(), "\t") != strings.Join(base[i].Raw(), "\t") {
			t.Fatalf("hf row %d (%s) modifies base row", i, base[i].ID)
		}
	}
	for i := len(base); i < len(hf); i++ {
		if baseIDs[hf[i].ID] {
			t.Fatalf("hf appended row %d (%s) duplicates a base id", i, hf[i].ID)
		}
	}
}

// AI registry parity: 40 AiProc entries, exactly one nil (FireMan), every
// distinct data ai value resolves (31 measured), no extra keys.
func TestAIRegistryParity(t *testing.T) {
	if len(monstdat.AiProc) != 40 {
		t.Fatalf("AiProc entries = %d, want 40", len(monstdat.AiProc))
	}
	nilCount := 0
	for id, fn := range monstdat.AiProc {
		if fn == nil {
			nilCount++
			if id != monstdat.AIFireMan {
				t.Errorf("nil AiProc at %v, want only FireMan", id)
			}
		}
	}
	if nilCount != 1 {
		t.Fatalf("nil AiProc entries = %d, want exactly 1 (FireMan)", nilCount)
	}
	// every distinct data ai resolves
	base, err := monstdat.ParseMonstdat(corpus(t, "assets/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat base: %v", err)
	}
	hf, err := monstdat.ParseMonstdat(corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat hf: %v", err)
	}
	used := map[string]bool{}
	for _, m := range base {
		used[m.AI] = true
	}
	for _, m := range hf {
		used[m.AI] = true
	}
	delete(used, "")
	if len(used) != 31 {
		t.Fatalf("distinct ai values = %d, want 31", len(used))
	}
	for name := range used {
		if _, ok := monstdat.AiProcByName[name]; !ok {
			t.Errorf("ai %q used by data but not in AiProc registry", name)
		}
	}
}

// Worked example: Uniq(...) treasure decodes to T_UNIQ | id (monstdat.cpp:39);
// T_NODROP rows drop nothing.
func TestTreasureDecoding(t *testing.T) {
	base, err := monstdat.ParseMonstdat(corpus(t, "assets/txtdata/monsters/monstdat.tsv"))
	if err != nil {
		t.Fatalf("ParseMonstdat: %v", err)
	}
	uniqCount, noDropCount := 0, 0
	var uniqIDs []string
	for _, m := range base {
		if m.Treasure&monstdat.T_UNIQ != 0 {
			uniqCount++
			uniqIDs = append(uniqIDs, m.ID)
		}
		if m.Treasure&monstdat.T_NODROP != 0 {
			noDropCount++
		}
	}
	if uniqCount != 2 {
		t.Fatalf("T_UNIQ rows = %d (%v), want 2 (MT_SKING, MT_CLEAVER)", uniqCount, uniqIDs)
	}
	if noDropCount != 4 {
		t.Fatalf("T_NODROP rows = %d, want 4 (Fiend/Blink/Gloom/Familiar)", noDropCount)
	}
	// Butcher's Cleaver decodes to T_UNIQ | UITEM_CLEAVER (monstdat.cpp:306)
	for _, m := range base {
		if m.ID == "MT_CLEAVER" && m.Treasure != monstdat.T_UNIQ|monstdat.UITEMCLEAVER {
			t.Fatalf("MT_CLEAVER treasure = %#x, want T_UNIQ|UITEM_CLEAVER", m.Treasure)
		}
	}
}

// GetMissileType mapping for all documented pairs.
func TestGetMissileTypeMapping(t *testing.T) {
	pairs := []struct {
		monster string
		want    string
	}{
		{"GoatRanged", "Arrow"},
		{"Succubus", "BloodStar"},
		{"LazarusSuccubus", "BloodStar"},
		{"Acid", "Acid"},
		{"FireBat", "Firebolt"},
		{"Torchant", "Fireball"},
		{"Lich", "OrangeFlare"},
		{"ArchLich", "YellowFlare"},
		{"Psychorb", "BlueFlare"},
		{"Necromorb", "RedFlare"},
		{"Magma", "MagmaBall"},
		{"Storm", "ThinLightningControl"},
		{"Diablo", "DiabloApocalypse"},
		{"BoneDemon", "BlueFlare2"},
	}
	for _, p := range pairs {
		if got := monstdat.GetMissileType(p.monster); got != p.want {
			t.Errorf("GetMissileType(%s) = %q, want %q", p.monster, got, p.want)
		}
	}
	if got := monstdat.GetMissileType("Zombie"); got != "Arrow" {
		t.Errorf("default GetMissileType(Zombie) = %q, want Arrow", got)
	}
}
