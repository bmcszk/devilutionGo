// Package itemdat_test tests the itemdat TSV loading.
package itemdat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
)

func TestLoadItemDat(t *testing.T) {
	// given
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/itemdat.tsv")

	// when
	items, err := itemdat.LoadItemDat(data)
	if err != nil {
		t.Fatalf("LoadItemDat: %v", err)
	}

	// then
	assertRow(t, items, 0, "IDI_GOLD")
	assertRow(t, items, 32, "IDI_LGTFORGE")
	assertRow(t, items, 39, "IDI_RUNEBOMB")
	t.Logf("LoadItemDat: %d items loaded", len(items))
}

func TestLoadItemPrefixes(t *testing.T) {
	// given
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/item_prefixes.tsv")

	// when
	prefixes, err := itemdat.LoadItemPrefixes(data)
	if err != nil {
		t.Fatalf("LoadItemPrefixes: %v", err)
	}

	// then
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
	// given
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/item_suffixes.tsv")

	// when
	suffixes, err := itemdat.LoadItemSuffixes(data)
	if err != nil {
		t.Fatalf("LoadItemSuffixes: %v", err)
	}

	// then
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
	// given
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/unique_itemdat.tsv")

	// when
	uniques, err := itemdat.LoadUniqueItemDat(data)
	if err != nil {
		t.Fatalf("LoadUniqueItemDat: %v", err)
	}

	// then
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
	// given
	data := testsupport.Corpus(t, "items", "mods/hf/txtdata/items/item_prefixes.tsv")

	// when
	prefixes, err := itemdat.LoadItemPrefixes(data)
	if err != nil {
		t.Fatalf("LoadItemPrefixes hf: %v", err)
	}

	// then
	// AC: 86 hf rows (+3 appended)
	if len(prefixes) != 86 {
		t.Fatalf("hf item_prefixes rows = %d, want 86", len(prefixes))
	}
	t.Logf("LoadItemPrefixes hf: %d prefixes loaded", len(prefixes))
}

func TestLoadHFItemSuffixes(t *testing.T) {
	// given
	data := testsupport.Corpus(t, "items", "mods/hf/txtdata/items/item_suffixes.tsv")

	// when
	suffixes, err := itemdat.LoadItemSuffixes(data)
	if err != nil {
		t.Fatalf("LoadItemSuffixes hf: %v", err)
	}

	// then
	// AC: 98 hf rows (+3 appended, 20 modified)
	if len(suffixes) != 98 {
		t.Fatalf("hf item_suffixes rows = %d, want 98", len(suffixes))
	}
	t.Logf("LoadItemSuffixes hf: %d suffixes loaded", len(suffixes))
}

func TestLoadHFUniqueItemDat(t *testing.T) {
	// given
	data := testsupport.Corpus(t, "items", "mods/hf/txtdata/items/unique_itemdat.tsv")

	// when
	uniques, err := itemdat.LoadUniqueItemDat(data)
	if err != nil {
		t.Fatalf("LoadUniqueItemDat hf: %v", err)
	}

	// then
	// AC: 110 hf rows (+20 appended, 2 modified)
	if len(uniques) != 110 {
		t.Fatalf("hf unique_itemdat rows = %d, want 110", len(uniques))
	}
	t.Logf("LoadUniqueItemDat hf: %d uniques loaded", len(uniques))
}

func TestTotalRowCount(t *testing.T) {
	// given
	// Total base rows: 168 + 83 + 95 + 90 = 436
	data := testsupport.Corpus(t, "items", "assets/txtdata/items/itemdat.tsv")
	items, _ := itemdat.LoadItemDat(data)
	data = testsupport.Corpus(t, "items", "assets/txtdata/items/item_prefixes.tsv")
	prefixes, _ := itemdat.LoadItemPrefixes(data)
	data = testsupport.Corpus(t, "items", "assets/txtdata/items/item_suffixes.tsv")
	suffixes, _ := itemdat.LoadItemSuffixes(data)
	data = testsupport.Corpus(t, "items", "assets/txtdata/items/unique_itemdat.tsv")
	uniques, _ := itemdat.LoadUniqueItemDat(data)

	// when
	total := len(items) + len(prefixes) + len(suffixes) + len(uniques)

	// then
	if total != 436 {
		t.Fatalf("total base rows = %d, want 436", total)
	}
	t.Logf("Total base rows: %d", total)
}

// assertRow checks the expected length and the ID at the given index.
func assertRow(t *testing.T, items []itemdat.ItemData, idx int, wantID string) {
	t.Helper()
	if len(items) != 168 {
		t.Fatalf("itemdat rows = %d, want 168", len(items))
	}
	if items[idx].ID != wantID {
		t.Fatalf("row %d ID = %q, want %s", idx+1, items[idx].ID, wantID)
	}
}
