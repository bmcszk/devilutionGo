package monstdat_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/monstdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

func TestRowCountsParity(t *testing.T) {
	// given
	baseCorpus := corpus(t, "assets/txtdata/monsters/monstdat.tsv")
	hfCorpus := corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv")
	uniqCorpus := corpus(t, "assets/txtdata/monsters/unique_monstdat.tsv")
	// when
	base, err := monstdat.ParseMonstdat(baseCorpus)
	if err != nil {
		t.Fatalf("ParseMonstdat base: %v", err)
	}

	hf, err := monstdat.ParseMonstdat(hfCorpus)
	if err != nil {
		t.Fatalf("ParseMonstdat hf: %v", err)
	}

	uniq, err := monstdat.ParseUniqueMonstdat(uniqCorpus)
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

func TestHFMonstdatStrictSuperset(t *testing.T) {
	// given
	baseCorpus := corpus(t, "assets/txtdata/monsters/monstdat.tsv")
	hfCorpus := corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv")
	// when
	base := mustParseMonstdat(t, "base", baseCorpus)
	hf := mustParseMonstdat(t, "hf", hfCorpus)
	// then
	if len(hf)-len(base) != 26 {
		t.Fatalf("hf appended = %d, want 26", len(hf)-len(base))
	}

	assertSharedRowsUnchanged(t, base, hf)
	assertAppendsAreNew(t, base, hf)
}

func TestAIRegistryParity(t *testing.T) {
	// given
	baseCorpus := corpus(t, "assets/txtdata/monsters/monstdat.tsv")
	hfCorpus := corpus(t, "mods/hf/txtdata/monsters/monstdat.tsv")
	// when
	base := mustParseMonstdat(t, "base", baseCorpus)
	hf := mustParseMonstdat(t, "hf", hfCorpus)
	used := distinctAIValues(base, hf)
	// then
	if len(used) != 31 {
		t.Fatalf("distinct ai values = %d, want 31", len(used))
	}

	assertAIInRegistry(t, used)

	if len(monstdat.AiProc) != 40 {
		t.Fatalf("monstdat.AiProc entries = %d, want 40", len(monstdat.AiProc))
	}

	assertSingleNilProc(t)
}

func TestTreasureDecoding(t *testing.T) {
	// given
	baseCorpus := corpus(t, "assets/txtdata/monsters/monstdat.tsv")
	// when
	base := mustParseMonstdat(t, "base", baseCorpus)
	// then
	uniqCount, uniqIDs, noDropCount := treasureCounts(base)
	if uniqCount != 2 {
		t.Fatalf("monstdat.T_UNIQ rows = %d (%v), want 2 (MT_SKING, MT_CLEAVER)", uniqCount, uniqIDs)
	}

	if noDropCount != 4 {
		t.Fatalf("monstdat.T_NODROP rows = %d, want 4 (Fiend/Blink/Gloom/Familiar)", noDropCount)
	}

	assertCleaverTreasure(t, base)
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
			if got := monstdat.MissileTypeFor(p.monster); got != p.want {
				t.Errorf("monstdat.MissileTypeFor(%s) = %q, want %q", p.monster, got, p.want)
			}
		})
	}

	if got := monstdat.MissileTypeFor("Zombie"); got != "Arrow" {
		t.Errorf("default monstdat.MissileTypeFor(Zombie) = %q, want Arrow", got)
	}
}

// mustParseMonstdat parses a monstdat corpus or fails the test.
func mustParseMonstdat(t *testing.T, label string, data []byte) []monstdat.MonsterData {
	t.Helper()

	rows, err := monstdat.ParseMonstdat(data)
	if err != nil {
		t.Fatalf("ParseMonstdat %s: %v", label, err)
	}

	return rows
}

// assertSharedRowsUnchanged checks hf keeps base rows byte-identical in order.
func assertSharedRowsUnchanged(t *testing.T, base, hf []monstdat.MonsterData) {
	t.Helper()

	for i := range base {
		if hf[i].ID != base[i].ID || strings.Join(hf[i].Raw(), "\t") != strings.Join(base[i].Raw(), "\t") {
			t.Fatalf("hf row %d (%s) modifies base row", i, base[i].ID)
		}
	}
}

// assertAppendsAreNew checks the appended hf rows carry no base id.
func assertAppendsAreNew(t *testing.T, base, hf []monstdat.MonsterData) {
	t.Helper()

	baseIDs := map[string]bool{}
	for _, m := range base {
		baseIDs[m.ID] = true
	}

	for i := len(base); i < len(hf); i++ {
		if baseIDs[hf[i].ID] {
			t.Fatalf("hf appended row %d (%s) duplicates a base id", i, hf[i].ID)
		}
	}
}

// distinctAIValues collects the non-empty ai values used by both corpora.
func distinctAIValues(base, hf []monstdat.MonsterData) map[string]bool {
	used := map[string]bool{}
	for _, m := range base {
		used[m.AI] = true
	}

	for _, m := range hf {
		used[m.AI] = true
	}

	delete(used, "")

	return used
}

// assertAIInRegistry checks every used ai value is in monstdat.AiProcByName.
func assertAIInRegistry(t *testing.T, used map[string]bool) {
	t.Helper()

	for name := range used {
		if _, ok := monstdat.AiProcByName[name]; !ok {
			t.Errorf("ai %q used by data but not in monstdat.AiProc registry", name)
		}
	}
}

// assertSingleNilProc checks the registry's only nil proc is FireMan.
func assertSingleNilProc(t *testing.T) {
	t.Helper()

	nilCount := 0

	for id, fn := range monstdat.AiProc {
		if fn == nil {
			nilCount++

			if id != monstdat.AIFireMan {
				t.Errorf("nil monstdat.AiProc at %v, want only FireMan", id)
			}
		}
	}

	if nilCount != 1 {
		t.Fatalf("nil monstdat.AiProc entries = %d, want exactly 1 (FireMan)", nilCount)
	}
}

// treasureCounts counts monstdat.T_UNIQ and monstdat.T_NODROP rows.
func treasureCounts(rows []monstdat.MonsterData) (uniq int, uniqIDs []string, noDrop int) {
	for _, m := range rows {
		if m.Treasure&monstdat.T_UNIQ != 0 {
			uniq++

			uniqIDs = append(uniqIDs, m.ID)
		}

		if m.Treasure&monstdat.T_NODROP != 0 {
			noDrop++
		}
	}

	return uniq, uniqIDs, noDrop
}

// assertCleaverTreasure pins MT_CLEAVER's treasure. Upstream monstdat.cpp:37-39:
// Uniq(item) = monstdat.T_UNIQ + item, and UITEM_CLEAVER is the first
// _unique_items ordinal (items.h:47-48), i.e. 0 — so CLEAVER's treasure is
// exactly monstdat.T_UNIQ with no extra bits.
func assertCleaverTreasure(t *testing.T, rows []monstdat.MonsterData) {
	t.Helper()

	for _, m := range rows {
		if m.ID == "MT_CLEAVER" && m.Treasure != monstdat.T_UNIQ {
			t.Fatalf("MT_CLEAVER treasure = %#x, want monstdat.T_UNIQ|UITEM_CLEAVER", m.Treasure)
		}
	}
}

func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	return testsupport.Corpus(t, "monstdat", rel)
}
