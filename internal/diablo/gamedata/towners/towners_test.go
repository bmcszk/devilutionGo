// Package towners_test ports the town-domain parity tests against the live
// devilutionX corpus.
package towners_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/towners"
)

func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	name := filepath.Base(rel)
	if strings.Contains(rel, "mods/hf/") {
		name = "hf-" + name
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "internal", "testdata", "gamedata", "towners", name))
	if err != nil {
		t.Fatalf("read committed towners fixture %s: %v", name, err)
	}
	return data
}

// Row-count parity: towners 12 base / 15 hf (+3 appended); quest_dialog
// 10x18 base / 13x25 hf.
func TestRowCountsParity(t *testing.T) {
	// when
	base, err := towners.ParseTowners(corpus(t, "assets/txtdata/towners/towners.tsv"))
	hf, err2 := towners.ParseTowners(corpus(t, "mods/hf/txtdata/towners/towners.tsv"))
	// then
	if err != nil {
		t.Fatalf("ParseTowners base: %v", err)
	}
	if len(base) != 12 {
		t.Fatalf("base towners = %d, want 12", len(base))
	}
	if err2 != nil {
		t.Fatalf("ParseTowners hf: %v", err2)
	}
	if len(hf) != 15 {
		t.Fatalf("hf towners = %d, want 15", len(hf))
	}
	qbase, err3 := towners.ParseQuestDialog(corpus(t, "assets/txtdata/towners/quest_dialog.tsv"))
	qhf, err4 := towners.ParseQuestDialog(corpus(t, "mods/hf/txtdata/towners/quest_dialog.tsv"))
	// then
	if err3 != nil {
		t.Fatalf("ParseQuestDialog base: %v", err3)
	}
	if len(qbase.Rows) != 10 || qbase.Cols-1 != 18 {
		t.Fatalf("base quest_dialog = %d rows x %d quest cols, want 10x18", len(qbase.Rows), qbase.Cols-1)
	}
	if err4 != nil {
		t.Fatalf("ParseQuestDialog hf: %v", err4)
	}
	if len(qhf.Rows) != 13 || qhf.Cols-1 != 24 {
		t.Fatalf("hf quest_dialog = %d rows x %d quest cols, want 13x24", len(qhf.Rows), qhf.Cols-1)
	}
}

// hf towners: shared prefix content-identical, +3 appended.
func TestHFTownersAppendOnly(t *testing.T) {
	// when
	base, err := towners.ParseTowners(corpus(t, "assets/txtdata/towners/towners.tsv"))
	hf, err2 := towners.ParseTowners(corpus(t, "mods/hf/txtdata/towners/towners.tsv"))
	// then
	if err != nil {
		t.Fatalf("ParseTowners base: %v", err)
	}
	if err2 != nil {
		t.Fatalf("ParseTowners hf: %v", err2)
	}
	for i := range base {
		if hf[i].Type != base[i].Type || hf[i].Name != base[i].Name ||
			hf[i].X != base[i].X || hf[i].Y != base[i].Y {
			t.Fatalf("hf row %d (%s) modifies base row", i, base[i].Name)
		}
	}
}

// Dialog matrix: cells are TEXT_* ids or empty (TEXT_NONE); the mushroom
// forced cells hold (quests.cpp:61-62).
func TestDialogMatrixCells(t *testing.T) {
	// when
	qd, err := towners.ParseQuestDialog(corpus(t, "assets/txtdata/towners/quest_dialog.tsv"))
	// then
	if err != nil {
		t.Fatalf("ParseQuestDialog: %v", err)
	}
	// find the healer + witch rows by towner_type
	talkerRow := map[string]int{}
	for i, row := range qd.Rows {
		talkerRow[row[0]] = i
	}
	hr, wr := talkerRow["TOWN_HEALER"], talkerRow["TOWN_WITCH"]
	mush := slices.Index(qd.Header, "Q_MUSHROOM")
	if mush < 0 {
		t.Fatal("Q_MUSHROOM column missing")
	}
	if qd.Rows[hr][mush] != "" && qd.Rows[hr][mush] != "TEXT_NONE" {
		t.Fatalf("healer Q_MUSHROOM = %q, want TEXT_NONE", qd.Rows[hr][mush])
	}
	if qd.Rows[wr][mush] != "TEXT_MUSH9" {
		t.Fatalf("witch Q_MUSHROOM = %q, want TEXT_MUSH9", qd.Rows[wr][mush])
	}
}

// Price caps table: the vendor constants asserted from the data table
// (store caps; MaxVendorValue 140000 base / 200000 hf; MaxGold 5000).
func TestPriceCapsTable(t *testing.T) {
	if towners.MaxVendorValue != 140000 || towners.MaxVendorValueHF != 200000 {
		t.Fatalf("MaxVendorValue = %d/%d, want 140000/200000", towners.MaxVendorValue, towners.MaxVendorValueHF)
	}
	if towners.MaxGold != 5000 {
		t.Fatalf("MaxGold = %d, want 5000", towners.MaxGold)
	}
	if towners.GoldCursorSmall != 1000 || towners.GoldCursorMedium != 2500 || towners.GoldCursorLarge != 5000 {
		t.Fatal("gold cursor thresholds wrong")
	}
}
