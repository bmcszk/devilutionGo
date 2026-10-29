// Package itemdat_test tests the itemdat TSV loading.
package itemdat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

func TestLoadItemDat(t *testing.T) {
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/itemdat.tsv")
	items, err := itemdat.LoadItemDat(data)
	if err != nil {
		t.Fatalf("LoadItemDat: %v", err)
	}
	// AC: 168 base rows (header excluded) = 168 data rows
	// file has 169 lines (1 header + 168 data)
	if len(items) != 168 {
		t.Fatalf("itemdat rows = %d, want 168", len(items))
	}
	// Worked example: row 0 is IDI_GOLD (index 0)
	if len(items) > 0 && items[0].ID != "IDI_GOLD" {
		t.Fatalf("row 0 ID = %q, want IDI_GOLD", items[0].ID)
	}
	// Row 33 is IDI_LGTFORGE (index 32)
	if len(items) > 32 && items[32].ID != "IDI_LGTFORGE" {
		t.Fatalf("row 33 ID = %q, want IDI_LGTFORGE", items[32].ID)
	}
	// Row 40 is IDI_RUNEBOMB (index 39)
	if len(items) > 39 && items[39].ID != "IDI_RUNEBOMB" {
		t.Fatalf("row 40 ID = %q, want IDI_RUNEBOMB", items[39].ID)
	}
	t.Logf("LoadItemDat: %d items loaded", len(items))
}

func TestLoadItemPrefixes(t *testing.T) {
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/item_prefixes.tsv")
	prefixes, err := itemdat.LoadItemPrefixes(data)
	if err != nil {
		t.Fatalf("LoadItemPrefixes: %v", err)
	}
	// AC: 83 base rows
	if len(prefixes) != 83 {
		t.Fatalf("item_prefixes rows = %d, want 83", len(prefixes))
	}
	// First should be "Tin"
	if len(prefixes) > 0 && prefixes[0].Name != "Tin" {
		t.Fatalf("prefix[0] = %q, want Tin", prefixes[0].Name)
	}
	t.Logf("LoadItemPrefixes: %d prefixes loaded", len(prefixes))
}

func TestLoadItemSuffixes(t *testing.T) {
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/item_suffixes.tsv")
	suffixes, err := itemdat.LoadItemSuffixes(data)
	if err != nil {
		t.Fatalf("LoadItemSuffixes: %v", err)
	}
	// AC: 95 base rows
	if len(suffixes) != 95 {
		t.Fatalf("item_suffixes rows = %d, want 95", len(suffixes))
	}
	// First should be "quality"
	if len(suffixes) > 0 && suffixes[0].Name != "quality" {
		t.Fatalf("suffix[0] = %q, want quality", suffixes[0].Name)
	}
	t.Logf("LoadItemSuffixes: %d suffixes loaded", len(suffixes))
}

func TestLoadUniqueItemDat(t *testing.T) {
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/unique_itemdat.tsv")
	uniques, err := itemdat.LoadUniqueItemDat(data)
	if err != nil {
		t.Fatalf("LoadUniqueItemDat: %v", err)
	}
	// AC: 90 base rows
	if len(uniques) != 90 {
		t.Fatalf("unique_itemdat rows = %d, want 90", len(uniques))
	}
	// First should be "The Butcher's Cleaver"
	if len(uniques) > 0 && uniques[0].Name != "The Butcher's Cleaver" {
		t.Fatalf("unique[0] = %q, want The Butcher's Cleaver", uniques[0].Name)
	}
	t.Logf("LoadUniqueItemDat: %d uniques loaded", len(uniques))
}

func TestLoadHFItemPrefixes(t *testing.T) {
	data := testsupport.Corpus(t, "items", "mods/hf/txtdata/items/item_prefixes.tsv")
	prefixes, err := itemdat.LoadItemPrefixes(data)
	if err != nil {
		t.Fatalf("LoadItemPrefixes hf: %v", err)
	}
	// AC: 86 hf rows (+3 appended)
	if len(prefixes) != 86 {
		t.Fatalf("hf item_prefixes rows = %d, want 86", len(prefixes))
	}
	t.Logf("LoadItemPrefixes hf: %d prefixes loaded", len(prefixes))
}

func TestLoadHFItemSuffixes(t *testing.T) {
	data := testsupport.Corpus(t, "items", "mods/hf/txtdata/items/item_suffixes.tsv")
	suffixes, err := itemdat.LoadItemSuffixes(data)
	if err != nil {
		t.Fatalf("LoadItemSuffixes hf: %v", err)
	}
	// AC: 98 hf rows (+3 appended, 20 modified)
	if len(suffixes) != 98 {
		t.Fatalf("hf item_suffixes rows = %d, want 98", len(suffixes))
	}
	t.Logf("LoadItemSuffixes hf: %d suffixes loaded", len(suffixes))
}

func TestLoadHFUniqueItemDat(t *testing.T) {
	data := testsupport.Corpus(t, "items", "mods/hf/txtdata/items/unique_itemdat.tsv")
	uniques, err := itemdat.LoadUniqueItemDat(data)
	if err != nil {
		t.Fatalf("LoadUniqueItemDat hf: %v", err)
	}
	// AC: 110 hf rows (+20 appended, 2 modified)
	if len(uniques) != 110 {
		t.Fatalf("hf unique_itemdat rows = %d, want 110", len(uniques))
	}
	t.Logf("LoadUniqueItemDat hf: %d uniques loaded", len(uniques))
}

func TestTotalRowCount(t *testing.T) {
	// Total base rows: 168 + 83 + 95 + 90 = 436
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/itemdat.tsv")
	items, _ := itemdat.LoadItemDat(data)
	data = testsupport.Corpus(t, "items", "assets/txtdata/items/item_prefixes.tsv")
	prefixes, _ := itemdat.LoadItemPrefixes(data)
	data = testsupport.Corpus(t, "items", "assets/txtdata/items/item_suffixes.tsv")
	suffixes, _ := itemdat.LoadItemSuffixes(data)
	data = testsupport.Corpus(t, "items", "assets/txtdata/items/unique_itemdat.tsv")
	uniques, _ := itemdat.LoadUniqueItemDat(data)
	total := len(items) + len(prefixes) + len(suffixes) + len(uniques)
	if total != 436 {
		t.Fatalf("total base rows = %d, want 436", total)
	}
	t.Logf("Total base rows: %d", total)
}