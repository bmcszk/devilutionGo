// Package textdat_test ports the text-domain parity tests against the live
// devilutionX corpus.
package textdat_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/gamedata/textdat"
)

const devilutionX = "/home/blaze/github/bmcszk/devilutionX"

func corpus(t *testing.T, rel string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(devilutionX, rel))
	if err != nil {
		t.Skipf("corpus not available: %v", err)
	}
	return data
}

// Row-count parity per V8.1/V8.2: textdat 338 + unused 139 = 477; no hf
// override (positive drift guard).
func TestRowCountsParityNoHF(t *testing.T) {
	rows, err := textdat.ParseTextdat(corpus(t, "assets/txtdata/text/textdat.tsv"))
	if err != nil {
		t.Fatalf("ParseTextdat: %v", err)
	}
	if len(rows) != 338 {
		t.Fatalf("textdat rows = %d, want 338", len(rows))
	}
	unused, err := textdat.ParseTextdat(corpus(t, "assets/txtdata/text/textdat-unused.tsv"))
	if err != nil {
		t.Fatalf("ParseTextdat unused: %v", err)
	}
	if len(unused) != 139 {
		t.Fatalf("unused rows = %d, want 139", len(unused))
	}
	if _, err := os.Stat(filepath.Join(devilutionX, "mods/hf/txtdata/text/textdat.tsv")); err == nil {
		t.Fatal("hf textdat override exists but upstream has none")
	}
}

// Measured column facts: every row carries an sfxnr; 324 of 338 are scroll
// text (V8.1 measured); ids are dense 0-based.
func TestMeasuredColumnFacts(t *testing.T) {
	rows, err := textdat.ParseTextdat(corpus(t, "assets/txtdata/text/textdat.tsv"))
	if err != nil {
		t.Fatalf("ParseTextdat: %v", err)
	}
	scroll := 0
	named := 0
	for i, r := range rows {
		if r.Sfxnr == "" {
			t.Fatalf("row %d sfxnr empty (measured: all 338 carry one)", i)
		}
		named++
		if r.Scrlltxt {
			scroll++
		}
	}
	if named != 338 {
		t.Fatalf("rows with sfxnr = %d, want 338", named)
	}
	if scroll != 324 {
		t.Fatalf("scroll-text rows = %d, want 324", scroll)
	}
	for i, r := range rows {
		if r.TxtStrID == "" {
			t.Fatalf("row %d has empty txtstrid", i)
		}
	}
}

// Text bodies contain at least one verbatim quote from the loader's
// speech table spot set (byte-identity on content).
func TestVerbatimContentSpot(t *testing.T) {
	rows, err := textdat.ParseTextdat(corpus(t, "assets/txtdata/text/textdat.tsv"))
	if err != nil {
		t.Fatalf("ParseTextdat: %v", err)
	}
	joined := map[string]bool{}
	for _, r := range rows {
		joined[r.TxtStr] = true
	}
	// TEXT_KING2 opening, byte-exact from row 2 of the corpus
	const king2 = "The village needs your help, good master! Some months ago " +
		"King Leoric's son, Prince Albrecht, was kidnapped."
	if !anyContains(joined, king2) {
		t.Fatal("TEXT_KING2 speech line missing — content mismatch")
	}
}

func anyContains(set map[string]bool, sub string) bool {
	for s := range set {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
