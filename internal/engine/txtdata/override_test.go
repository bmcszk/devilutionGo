package txtdata_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// Fixtures mirror the V0.2 measured deltas (dg-requirements-append.md:3734-3746):
// spelldat.tsv base 36 rows / hf 51 — appended rows, a few edited, same header.
const spellHeader = "id\tname\tsoundId\tbookCost10\tstaffCost10\tmanaCost\tflags\tbookLevel" +
	"\tstaffLevel\tminIntelligence\tmissiles\tmanaMultiplier\tminMana\tstaffMin\tstaffMax"

// baseRows are the two retail spelldat.tsv data rows the override tests
// modify (assets/txtdata/spells/spelldat.tsv lines 2-3).
const baseRows = "Firebolt\tFirebolt\tCastFire\t100\t5\t6\tFire,Targeted\t1\t1\t15\tFirebolt\t1\t3\t40\t80\r\n" +
	"Healing\tHealing\tCastHealing\t100\t5\t5\tMagic,AllowedInTown\t1\t1\t17\tHealing\t3\t1\t20\t40\r\n"

func TestOverrideModifiesExistingRow(t *testing.T) {
	// given
	base := baseTable(t)
	mod, err := txtdata.Parse(spellRowBytes(
		"Firebolt\tFireboltHF\tCastFire\t999\t5\t6\tFire,Targeted\t1\t1\t15\tFirebolt\t1\t3\t40\t80"))
	if err != nil {
		t.Fatalf("Parse mod: %v", err)
	}
	stack, err := txtdata.NewOverrideStack(base, mod)
	if err != nil {
		t.Fatalf("newOverrideStack: %v", err)
	}
	// when
	merged, err := stack.Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// then
	if len(merged.Rows) != 2 {
		t.Fatalf("rows = %d, want 2 (override, not append)", len(merged.Rows))
	}
	if merged.Rows[0][1] != "FireboltHF" || merged.Rows[0][3] != "999" {
		t.Fatalf("override did not apply: %v", merged.Rows[0])
	}
}

func TestOverrideAppendsNewRows(t *testing.T) {
	// given
	base := baseTable(t)
	mod, err := txtdata.Parse(spellRowBytes(
		"AppendedSpell\tNew\tCastNew\t10\t1\t1\tMagic\t1\t1\t5\tNull\t1\t1\t10\t20"))
	if err != nil {
		t.Fatalf("Parse mod: %v", err)
	}
	stack, err := txtdata.NewOverrideStack(base, mod)
	if err != nil {
		t.Fatalf("newOverrideStack: %v", err)
	}
	// when
	merged, err := stack.Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// then
	if len(merged.Rows) != 3 {
		t.Fatalf("rows = %d, want 3 (append)", len(merged.Rows))
	}
	if merged.Rows[2][1] != "New" {
		t.Fatalf("appended row = %v", merged.Rows[2])
	}
}

func TestOverrideLaterModWins(t *testing.T) {
	// given
	base := baseTable(t)
	modA, _ := txtdata.Parse(spellRowBytes("Firebolt\tFromA\tCastFire\t1\t1\t1\tFire\t1\t1\t1\tNull\t1\t1\t1\t1"))
	modB, _ := txtdata.Parse(spellRowBytes("Firebolt\tFromB\tCastFire\t2\t2\t2\tFire\t2\t2\t2\tNull\t2\t2\t2\t2"))
	stack, err := txtdata.NewOverrideStack(base, modA, modB)
	if err != nil {
		t.Fatalf("newOverrideStack: %v", err)
	}
	// when
	merged, err := stack.Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	// then
	if merged.Rows[0][1] != "FromB" {
		t.Fatalf("later mod must win: got %v", merged.Rows[0])
	}
}

func TestOverrideKeyColumnMissing(t *testing.T) {
	// given
	base, _ := txtdata.Parse([]byte("a\tb\r\n1\t2\r\n"))
	mod, _ := txtdata.Parse([]byte("id\tx\r\n1\t2\r\n"))
	// when
	_, err := txtdata.NewOverrideStack(base, mod)
	// then
	if err == nil {
		t.Fatal("newOverrideStack = nil error, want key-column error on base")
	}
}

func spellRowBytes(values string) []byte {
	return []byte(spellHeader + "\r\n" + values + "\r\n")
}

func baseTable(t *testing.T) *txtdata.Table {
	t.Helper()
	tbl, err := txtdata.Parse([]byte(spellHeader + "\r\n" + baseRows))
	if err != nil {
		t.Fatalf("Parse base: %v", err)
	}
	return tbl
}
