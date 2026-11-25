package persist_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

// TestHeroStoreSetRehomesMultiRowLegacyFile pins the legacy sweep: a demo-era
// multi-row single_0.db (several heroes in one file) is re-homed ONCE at set
// construction — each row into its own single_<slot>.db, the original renamed
// single_0.db.pre-per-hero (never deleted — the rename IS the rollback), and
// a second construction is a no-op. A legitimate per-hero single_0.db (one
// row, slot 0) must NOT be touched.
func TestHeroStoreSetRehomesMultiRowLegacyFile(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()

	// given a legacy single_0.db holding heroes on slots 0 AND 2
	writeLegacyFile(t, dir)

	// when the set is constructed
	set := persist.NewHeroStoreSet(dir, heroSetNS)

	// then the rows were re-homed: slot 0 kept, slot 2 has its own file, and
	// the original survived as the .pre-per-hero marker
	assertRehomed(t, set, dir)

	// and a second construction is a no-op (the marker holds)
	reopened := persist.NewHeroStoreSet(dir, heroSetNS)

	rows, err := reopened.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes second construction: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("ListHeroes after second construction = %d rows, want unchanged 2", len(rows))
	}
}

// TestHeroStoreSetLeavesLegitSingleSlotZeroFile pins the discriminator: a
// legitimate per-hero single_0.db (exactly one row, slot 0) is NOT swept —
// no .pre-per-hero marker appears and the file keeps its row.
func TestHeroStoreSetLeavesLegitSingleSlotZeroFile(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()

	// given a legitimate per-hero single_0.db with only the slot-0 hero
	set := persist.NewHeroStoreSet(dir, heroSetNS)

	if err := set.SaveHero(ctx, heroRow(0, "Only Hero")); err != nil {
		t.Fatalf("SaveHero slot 0: %v", err)
	}

	// when the set is constructed again over the same dir
	persist.NewHeroStoreSet(dir, heroSetNS)

	// then nothing was swept
	if _, err := os.Stat(filepath.Join(dir, "single_0.db.pre-per-hero")); !os.IsNotExist(err) {
		t.Fatal("single_0.db.pre-per-hero exists, want none (legit file must not be swept)")
	}

	row, err := set.LoadHero(ctx, 0)
	if err != nil {
		t.Fatalf("LoadHero slot 0: %v", err)
	}

	if row.Name != "Only Hero" {
		t.Fatalf("slot 0 hero = %q, want Only Hero", row.Name)
	}
}

// writeLegacyFile writes the demo-era shape: a multi-row single_0.db.
func writeLegacyFile(t *testing.T, dir string) {
	t.Helper()

	legacy, err := persist.OpenSQLiteStore(context.Background(), filepath.Join(dir, "single_0.db"))
	if err != nil {
		t.Fatalf("OpenSQLiteStore legacy: %v", err)
	}

	for _, row := range []persist.HeroRecord{heroRow(0, "Stay Put"), heroRow(2, "Moved")} {
		if err := legacy.SaveHero(context.Background(), row); err != nil {
			t.Fatalf("SaveHero slot %d: %v", row.Slot, err)
		}
	}

	if err := legacy.Close(); err != nil {
		t.Fatalf("legacy Close: %v", err)
	}
}

// assertRehomed checks the sweep's outcome: two heroes on their own slots and
// the .pre-per-hero marker present.
func assertRehomed(t *testing.T, set *persist.HeroStoreSet, dir string) {
	t.Helper()

	ctx := context.Background()

	rows, err := set.ListHeroes(ctx)
	if err != nil {
		t.Fatalf("ListHeroes after re-home: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("ListHeroes after re-home = %d rows (%v), want 2", len(rows), rows)
	}

	slot2, err := set.LoadHero(ctx, 2)
	if err != nil {
		t.Fatalf("LoadHero slot 2 after re-home: %v", err)
	}

	if slot2.Name != "Moved" {
		t.Fatalf("slot 2 hero = %q, want Moved", slot2.Name)
	}

	if _, err := os.Stat(filepath.Join(dir, "single_0.db.pre-per-hero")); err != nil {
		t.Fatalf("single_0.db.pre-per-hero missing after sweep: %v", err)
	}
}
