//nolint:testpackage // internal-state test: pokes unexported identifiers
package txtdata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// spellRow holds the header and the first two data rows of the committed
// spelldat fixture (internal/testdata/gamedata/spells/spelldat.tsv lines 1-3:
// header + Firebolt + Healing).
var spellRow = mustReadSpellRows()

func TestLoadSpellDataGoldenRow(t *testing.T) {
	// given
	tbl, err := enginetxtdata.Parse([]byte(spellRow))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// when
	spells, err := loadSpellData(tbl)
	// then
	if err != nil {
		t.Fatalf("loadSpellData: %v", err)
	}
	if len(spells) != 2 {
		t.Fatalf("spells = %d, want 2", len(spells))
	}
	verifySpellIdentity(t, spells[0])
	verifySpellCosts(t, spells[0])
	verifySpellFlags(t, spells[0])
	verifySpellManaStaff(t, spells[0])
}

func verifySpellIdentity(t *testing.T, s spellData) {
	t.Helper()
	if s.Name != "Firebolt" || s.SoundID != "CastFire" {
		t.Fatalf("name/sound = %q/%q", s.Name, s.SoundID)
	}
}

func verifySpellCosts(t *testing.T, s spellData) {
	t.Helper()
	if s.BookCost10 != 100 || s.StaffCost10 != 5 || s.ManaCost != 6 {
		t.Fatalf("costs = %d/%d/%d", s.BookCost10, s.StaffCost10, s.ManaCost)
	}
}

func verifySpellFlags(t *testing.T, s spellData) {
	t.Helper()
	if len(s.Flags) != 2 || s.Flags[0] != "Fire" || s.Flags[1] != "Targeted" {
		t.Fatalf("flags = %v", s.Flags)
	}
}

func verifySpellManaStaff(t *testing.T, s spellData) {
	t.Helper()
	if s.ManaMultiplier != 1 || s.MinMana != 3 || s.StaffMin != 40 || s.StaffMax != 80 {
		t.Fatalf("mana/staff = %d/%d/%d/%d", s.ManaMultiplier, s.MinMana, s.StaffMin, s.StaffMax)
	}
}

func TestLoadSpellDataMissilesPaddedToNull(t *testing.T) {
	tbl, _ := enginetxtdata.Parse([]byte(spellRow))
	// when
	spells, err := loadSpellData(tbl)
	// then
	if err != nil {
		t.Fatalf("loadSpellData: %v", err)
	}
	if len(spells[0].Missiles) != 2 || spells[0].Missiles[0] != "Firebolt" || spells[0].Missiles[1] != "Null" {
		t.Fatalf("missiles = %v, want [Firebolt Null]", spells[0].Missiles)
	}
}

func TestLoadSpellDataMissingColumn(t *testing.T) {
	// given
	bad := "id	name\r\nFirebolt	Firebolt\r\n"
	tbl, err := enginetxtdata.Parse([]byte(bad))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// when
	_, err = loadSpellData(tbl)
	// then
	if err == nil {
		t.Fatal("loadSpellData = nil error, want missing-column error")
	}
}

// mustReadSpellRows reads the first three records of the committed spelldat
// TSV fixture and re-joins them with CRLF line endings.
func mustReadSpellRows() string {
	path := filepath.Join("..", "..", "testdata", "gamedata", "spells", "spelldat.tsv")
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("read %q: %v", path, err))
	}
	records := make([]string, 0, 3)
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}
		records = append(records, line)
		if len(records) == 3 {
			break
		}
	}
	if len(records) != 3 {
		panic(fmt.Sprintf("%q: got %d records, want 3", path, len(records)))
	}
	return strings.Join(records, "\r\n") + "\r\n"
}
