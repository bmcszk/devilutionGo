package itemgen_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestSetupAllItemsUniqueCursor(t *testing.T) {
	tests := []struct {
		name       string
		cursor     string
		wantCursor string
	}{
		{name: "override replaces the base graphic", cursor: "HELM_OF_SPIRITS", wantCursor: "HELM_OF_SPIRITS"},
		{name: "default keeps the base graphic", cursor: "", wantCursor: "CLEAVER"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			cursor := tc.cursor

			// when
			plan, err := applyCursorCase(t, cursor)

			// then
			assertPlanCursor(t, plan, err, tc.wantCursor)
		})
	}
}

func applyCursorCase(t *testing.T, cursor string) (itemgen.SetupPlan, error) {
	t.Helper()

	items, uniques := cursorTables(cursor)
	req := itemgen.SetupRequest{ItemIndex: 1, Seed: 7, Level: 10, Uper: 15}
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0}}

	plan, err := itemgen.SetupAllItems(items, uniques, nil, nil, req, rnd.rnd, func(int) {})
	if err != nil {
		return itemgen.SetupPlan{}, fmt.Errorf("applyCursorCase: %w", err)
	}

	return plan, nil
}

func assertPlanCursor(t *testing.T, plan itemgen.SetupPlan, err error, want string) {
	t.Helper()

	if err != nil {
		t.Fatalf("SetupAllItems: %v", err)
	}
	if plan.Cursor != want {
		t.Fatalf("cursor = %q, want %q (ICURS_DEFAULT keeps the base)", plan.Cursor, want)
	}
}

func cursorTables(cursor string) ([]itemdat.ItemData, []itemdat.UniqueItemData) {
	items := []itemdat.ItemData{
		{ID: "SWORD", Name: "Sword", ItemType: "Sword", CursorGraphic: "SWORD", Durability: 100},
		{ID: "CLEAVER", Name: "Cleaver", ItemType: "Sword", CursorGraphic: "CLEAVER", Durability: 120},
	}
	uniques := []itemdat.UniqueItemData{{
		Name: "The Butcher's Cleaver", CursorGraphic: cursor, UniqueBaseItem: "CLEAVER",
		MinLevel: 1, Value: 2500,
	}}

	return items, uniques
}
