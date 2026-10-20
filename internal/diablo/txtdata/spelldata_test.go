package txtdata_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// Golden rows copied from assets/txtdata/spells/spelldat.tsv lines 1-3
// (header + Firebolt + Healing).
var spellRow = strings.Join([]string{
	"id\tname\tsoundId\tbookCost10\tstaffCost10\tmanaCost\tflags\tbookLevel\tstaffLevel\tminIntelligence" +
		"\tmissiles\tmanaMultiplier\tminMana\tstaffMin\tstaffMax",
	"Firebolt\tFirebolt\tCastFire\t100\t5\t6\tFire,Targeted\t1\t1\t15\tFirebolt\t1\t3\t40\t80",
	"Healing\tHealing\tCastHealing\t100\t5\t5\tMagic,AllowedInTown\t1\t1\t17\tHealing\t3\t1\t20\t40",
}, "\r\n") + "\r\n"

func TestLoadSpellDataGoldenRow(t *testing.T) {
	tbl, err := enginetxtdata.Parse([]byte(spellRow))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	spells, err := txtdata.LoadSpellData(tbl)
	if err != nil {
		t.Fatalf("LoadSpellData: %v", err)
	}
	if len(spells) != 2 {
		t.Fatalf("spells = %d, want 2", len(spells))
	}
	verifySpellIdentity(t, spells[0])
	verifySpellCosts(t, spells[0])
	verifySpellFlags(t, spells[0])
	verifySpellManaStaff(t, spells[0])
}

func verifySpellIdentity(t *testing.T, s txtdata.SpellData) {
	t.Helper()
	if s.Name != "Firebolt" || s.SoundID != "CastFire" {
		t.Fatalf("name/sound = %q/%q", s.Name, s.SoundID)
	}
}

func verifySpellCosts(t *testing.T, s txtdata.SpellData) {
	t.Helper()
	if s.BookCost10 != 100 || s.StaffCost10 != 5 || s.ManaCost != 6 {
		t.Fatalf("costs = %d/%d/%d", s.BookCost10, s.StaffCost10, s.ManaCost)
	}
}

func verifySpellFlags(t *testing.T, s txtdata.SpellData) {
	t.Helper()
	if len(s.Flags) != 2 || s.Flags[0] != "Fire" || s.Flags[1] != "Targeted" {
		t.Fatalf("flags = %v", s.Flags)
	}
}

func verifySpellManaStaff(t *testing.T, s txtdata.SpellData) {
	t.Helper()
	if s.ManaMultiplier != 1 || s.MinMana != 3 || s.StaffMin != 40 || s.StaffMax != 80 {
		t.Fatalf("mana/staff = %d/%d/%d/%d", s.ManaMultiplier, s.MinMana, s.StaffMin, s.StaffMax)
	}
}

func TestLoadSpellDataMissilesPaddedToNull(t *testing.T) {
	tbl, _ := enginetxtdata.Parse([]byte(spellRow))
	spells, err := txtdata.LoadSpellData(tbl)
	if err != nil {
		t.Fatalf("LoadSpellData: %v", err)
	}
	if len(spells[0].Missiles) != 2 || spells[0].Missiles[0] != "Firebolt" || spells[0].Missiles[1] != "Null" {
		t.Fatalf("missiles = %v, want [Firebolt Null]", spells[0].Missiles)
	}
}

func TestLoadSpellDataMissingColumn(t *testing.T) {
	bad := "id\tname\r\nFirebolt\tFirebolt\r\n"
	tbl, err := enginetxtdata.Parse([]byte(bad))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := txtdata.LoadSpellData(tbl); err == nil {
		t.Fatal("LoadSpellData = nil error, want missing-column error")
	}
}
