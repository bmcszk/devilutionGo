package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestGetOilTypeLevelPool(t *testing.T) {
	tests := []struct {
		name  string
		env   itemgen.SpellEnv
		lvl   int
		draw  int32
		want  string
		value int
	}{
		// level-1 pool in table order: 0 Accuracy, 2 Sharpness, 5 Blacksmith, 8 Hardening
		{
			name: "level 1 pool is the four level-1 oils", env: itemgen.SpellEnv{Hellfire: true},
			lvl: 1, draw: 1, want: "Oil of Sharpness", value: 500,
		},
		// level-4 pool: 0, 2, 4 Skill, 5, 8
		{
			name: "level 4 adds Oil of Skill", env: itemgen.SpellEnv{Hellfire: true},
			lvl: 4, draw: 2, want: "Oil of Skill", value: 1500,
		},
		// level-5 pool is 0, 2, 4, 5, 6 Fortitude, 8
		{
			name: "level 5 adds Oil of Fortitude", env: itemgen.SpellEnv{Hellfire: true},
			lvl: 5, draw: 4, want: "Oil of Fortitude", value: 2500,
		},
		{
			name: "level 10 drops the level-17 oil", env: itemgen.SpellEnv{Hellfire: true},
			lvl: 10, draw: 1, want: "Oil of Mastery", value: 2500,
		},
		{
			name: "multiplayer pins Blacksmith and Imperviousness",
			env:  itemgen.SpellEnv{Hellfire: true, Multiplayer: true},
			lvl:  1, draw: 0, want: "Blacksmith Oil", value: 100,
		},
		{
			name:  "level 0 promotes to 1",
			env:   itemgen.SpellEnv{Hellfire: true},
			lvl:   0,
			draw:  0,
			want:  "Oil of Accuracy",
			value: 500,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			rnd := &tapeRnd{t: t, values: []int32{tc.draw}}

			// when
			oil, err := itemgen.GetOilType(tc.env, tc.lvl, rnd.rnd)
			// then
			assertOil(t, oil, err, tc.want, tc.value)
		})
	}
}

func TestGetOilTypeDrawBounds(t *testing.T) {
	// given
	rnd := &tapeRnd{t: t, values: []int32{3, 9}}

	// when
	if _, err := itemgen.GetOilType(itemgen.SpellEnv{Hellfire: true}, 1, rnd.rnd); err != nil {
		t.Fatalf("GetOilType level 1: %v", err)
	}
	if _, err := itemgen.GetOilType(itemgen.SpellEnv{Hellfire: true}, 99, rnd.rnd); err != nil {
		t.Fatalf("GetOilType level 99: %v", err)
	}

	// then
	if rnd.args[0] != 4 || rnd.args[1] != 10 {
		t.Fatalf("draw args = %v, want 4 at level 1 and 10 at level 99", rnd.args)
	}
}

func TestGetGoldCursor(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{name: "small at the limit", value: itemgen.GoldSmallLimit, want: itemgen.CursorGoldSmall},
		{name: "small below the limit", value: 1, want: itemgen.CursorGoldSmall},
		{name: "medium just above small", value: itemgen.GoldSmallLimit + 1, want: itemgen.CursorGoldMedium},
		{name: "medium below the medium limit", value: itemgen.GoldMediumLimit - 1, want: itemgen.CursorGoldMedium},
		{name: "large at the medium limit", value: itemgen.GoldMediumLimit, want: itemgen.CursorGoldLarge},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			cursor := itemgen.GetGoldCursor(tc.value)

			// then
			if cursor != tc.want {
				t.Fatalf("cursor = %d, want %d", cursor, tc.want)
			}
		})
	}
}

func TestGetItemAttrsGoldCursor(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		currLevel  int
		draw       int32
		want       int
	}{
		{name: "small", difficulty: itemgen.DifficultyNormal, currLevel: 4, draw: 0, want: itemgen.CursorGoldSmall},
		{
			name:       "medium",
			difficulty: itemgen.DifficultyNormal,
			currLevel:  100,
			draw:       800,
			want:       itemgen.CursorGoldMedium,
		}, // 500+800
		{
			name:       "large",
			difficulty: itemgen.DifficultyHell,
			currLevel:  200,
			draw:       1500,
			want:       itemgen.CursorGoldLarge,
		}, // 1160+1500
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			row := itemdat.ItemData{ID: "GOLD", ItemType: "Gold", CursorGraphic: "GOLD"}
			rnd := &tapeRnd{t: t, values: []int32{0, tc.draw}}

			// when
			attrs, err := itemgen.GetItemAttrs(row, tc.difficulty, tc.currLevel, false, rnd.rnd)

			// then
			assertGoldCursor(t, attrs, err, tc.want)
		})
	}
}

func assertGoldCursor(t *testing.T, attrs itemgen.ItemAttrs, err error, want int) {
	t.Helper()

	if err != nil {
		t.Fatalf("GetItemAttrs: %v", err)
	}
	if attrs.GoldCursor != want {
		t.Fatalf("gold cursor = %d, want %d (value %d)", attrs.GoldCursor, want, attrs.Value)
	}
}

func assertOil(t *testing.T, oil itemgen.Oil, err error, want string, value int) {
	t.Helper()

	if err != nil {
		t.Fatalf("GetOilType: %v", err)
	}
	if oil.Name != want || oil.Value != value {
		t.Fatalf("oil = %+v, want %s value %d", oil, want, value)
	}
}
