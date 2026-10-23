//nolint:testpackage // internal-state test: pokes unexported identifiers
package monstdat

import (
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
	data, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "..", "internal", "testdata", "gamedata", "monstdat", name,
	))
	if err != nil {
		t.Fatalf("read committed monstdat fixture %s: %v", name, err)
	}
	return data
}

func TestRowCountsParity(t *testing.T) {
	// given
	baseCorpus := corpus(t, "assets/txtdata/monsters/monstdat.tsv")
	hfCorpus := corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv")
	uniqCorpus := corpus(t, "assets/txtdata/monsters/unique_monstdat.tsv")
	// when
	base, err := ParseMonstdat(baseCorpus)
	if err != nil {
		t.Fatalf("ParseMonstdat base: %v", err)
	}
	hf, err := ParseMonstdat(hfCorpus)
	if err != nil {
		t.Fatalf("ParseMonstdat hf: %v", err)
	}
	uniq, err := ParseUniqueMonstdat(uniqCorpus)
	if err != nil {
		t.Fatalf("ParseUniqueMonstdat: %v", err)
	}
	// then
	if len(base) != 112 {
		t.Fatalf("base monstdat rows = %d, want 112", len(base))
	}
	if len(hf) != 138 {
		t.Fatalf("hf monstdat rows = %d, want 138", len(hf))
	}
	if len(uniq) != 100 {
		t.Fatalf("unique rows = %d, want 100", len(uniq))
	}
}

func TestHFMisdatStrictSuperset(t *testing.T) {
	// given
	baseCorpus := corpus(t, "assets/txtdata/monsters/monstdat.tsv")
	hfCorpus := corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv")
	// when
	base, err := ParseMonstdat(baseCorpus)
	if err != nil {
		t.Fatalf("ParseMonstdat base: %v", err)
	}
	hf, err := ParseMonstdat(hfCorpus)
	if err != nil {
		t.Fatalf("ParseMonstdat hf: %v", err)
	}
	// then
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

func TestAIRegistryParity(t *testing.T) {
	// given
	baseCorpus := corpus(t, "assets/txtdata/monsters/monstdat.tsv")
	hfCorpus := corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv")
	// when
	base, err := ParseMonstdat(baseCorpus)
	if err != nil {
		t.Fatalf("ParseMonstdat base: %v", err)
	}
	hf, err := ParseMonstdat(hfCorpus)
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
	// then
	if len(used) != 31 {
		t.Fatalf("distinct ai values = %d, want 31", len(used))
	}
	for name := range used {
		if _, ok := AiProcByName[name]; !ok {
			t.Errorf("ai %q used by data but not in AiProc registry", name)
		}
	}
	if len(AiProc) != 40 {
		t.Fatalf("AiProc entries = %d, want 40", len(AiProc))
	}
	nilCount := 0
	for id, fn := range AiProc {
		if fn == nil {
			nilCount++
			if id != AIFireMan {
				t.Errorf("nil AiProc at %v, want only FireMan", id)
			}
		}
	}
	if nilCount != 1 {
		t.Fatalf("nil AiProc entries = %d, want exactly 1 (FireMan)", nilCount)
	}
}

func TestTreasureDecoding(t *testing.T) {
	// given
	baseCorpus := corpus(t, "assets/txtdata/monsters/monstdat.tsv")
	// when
	base, err := ParseMonstdat(baseCorpus)
	if err != nil {
		t.Fatalf("ParseMonstdat: %v", err)
	}
	// then
	uniqCount, noDropCount := 0, 0
	var uniqIDs []string
	for _, m := range base {
		if m.Treasure&T_UNIQ != 0 {
			uniqCount++
			uniqIDs = append(uniqIDs, m.ID)
		}
		if m.Treasure&T_NODROP != 0 {
			noDropCount++
		}
	}
	if uniqCount != 2 {
		t.Fatalf("T_UNIQ rows = %d (%v), want 2 (MT_SKING, MT_CLEAVER)", uniqCount, uniqIDs)
	}
	if noDropCount != 4 {
		t.Fatalf("T_NODROP rows = %d, want 4 (Fiend/Blink/Gloom/Familiar)", noDropCount)
	}
	// Upstream monstdat.cpp:37-39: Uniq(item) = T_UNIQ + item, and UITEM_CLEAVER is the
	// first _unique_items ordinal (items.h:47-48), i.e. 0 — so CLEAVER's treasure is
	// exactly T_UNIQ with no extra bits (T_UNIQ|UITEMCLEAVER == T_UNIQ).
	for _, m := range base {
		// UITEMCLEAVER is 0 upstream (items.h:47-48), so T_UNIQ|UITEMCLEAVER == T_UNIQ.
		if m.ID == "MT_CLEAVER" && m.Treasure != T_UNIQ {
			t.Fatalf("MT_CLEAVER treasure = %#x, want T_UNIQ|UITEM_CLEAVER", m.Treasure)
		}
	}
}

func TestGetMissileTypeMapping(t *testing.T) {
	// given
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
		t.Run(p.monster, func(t *testing.T) {
			if got := getMissileType(p.monster); got != p.want {
				t.Errorf("getMissileType(%s) = %q, want %q", p.monster, got, p.want)
			}
		})
	}
	if got := getMissileType("Zombie"); got != "Arrow" {
		t.Errorf("default getMissileType(Zombie) = %q, want Arrow", got)
	}
}
