package gameplay_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/gameplay"
)

// TestItemgenFacadeSpawnsUnique pins that the gameplay entry point is the
// itemgen port itself: a DIFF_NORMAL SpawnUnique resolves the base row through
// the facade and applies the unique (items.cpp:3207-3214).
func TestItemgenFacadeSpawnsUnique(t *testing.T) {
	// given
	deps := gameplay.SpawnDeps{
		Items: []itemdat.ItemData{{
			ID: "IDI_CLEAVER", ItemType: "Sword", Class: "Weapon",
			UniqueBaseItem: "CLEAVER", CursorGraphic: "CLEAVER", Value: 100,
		}},
		Uniques: []itemdat.UniqueItemData{{
			Name: "The Butcher's Cleaver", UniqueBaseItem: "CLEAVER", Value: 3650,
		}},
		Currlevel: 4,
		Rnd:       func(int32) int32 { return 0 },
		Discard:   func(int) {},
		NextSeed:  func() uint32 { return 0 },
	}

	// when
	out, err := gameplay.SpawnUnique(deps, 0, gameplay.LevelNotGiven, gameplay.DifficultyNormal, 1)
	// then
	if err != nil {
		t.Fatalf("SpawnUnique: %v", err)
	}
	if !out.Direct || out.BaseIndex != 0 {
		t.Fatalf("out = %+v, want the DIFF_NORMAL path on base row 0", out)
	}
	if out.Setup.Quality != 0x02 {
		t.Fatalf("quality = %d, want the unique bucket", out.Setup.Quality)
	}
}
