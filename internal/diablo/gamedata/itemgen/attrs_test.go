package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

func TestGetItemAttrsCopiesRow(t *testing.T) {
	// given
	row := attrsRow()
	row.MinArmor, row.MaxArmor = 10, 14
	rnd := &tapeRnd{t: t, values: []int32{2}}

	// when
	attrs, err := itemgen.GetItemAttrs(row, itemgen.DifficultyNormal, 5, false, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetItemAttrs: %v", err)
	}
	if attrs.AC != 12 || rnd.args[0] != 5 {
		t.Fatalf("AC = %d draw = %d, want 12 from GenerateRnd(14-10+1)", attrs.AC, rnd.args[0])
	}
	assertRowCopy(t, attrs)
}

func TestGetItemAttrsGoldRoll(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		currLevel  int
		hell       bool
		draw       int32
		want       int
	}{
		{name: "normal", difficulty: itemgen.DifficultyNormal, currLevel: 4, draw: 0, want: 20},
		{name: "nightmare offset 16", difficulty: itemgen.DifficultyNightmare, currLevel: 4, draw: 0, want: 100},
		{name: "hell offset 32", difficulty: itemgen.DifficultyHell, currLevel: 4, draw: 0, want: 180},
		{name: "hell dungeon bonus", difficulty: itemgen.DifficultyHell, currLevel: 4, draw: 0, hell: true, want: 202},
		{
			name:       "clamped to the limit",
			difficulty: itemgen.DifficultyHell,
			currLevel:  320, // level 352: 5*352+3519 = 5279 clamps to 5000
			draw:       3519,
			want:       itemgen.GoldMaxLimit,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			row := itemdat.ItemData{ID: "GOLD", ItemType: "Gold", CursorGraphic: "GOLD", Value: 1}
			// first draw is the AC band GenerateRnd(1), then the gold roll.
			rnd := &tapeRnd{t: t, values: []int32{0, tc.draw}}

			// when
			attrs, err := itemgen.GetItemAttrs(row, tc.difficulty, tc.currLevel, tc.hell, rnd.rnd)
			// then
			assertGoldValue(t, attrs, err, tc.want)
		})
	}
}

func TestGetItemAttrsUnknownDifficulty(t *testing.T) {
	// given
	row := itemdat.ItemData{ID: "GOLD", ItemType: "Gold"}

	// when
	_, err := itemgen.GetItemAttrs(row, 7, 4, false, seededRnd(1))

	// then
	if err == nil {
		t.Fatal("err = nil, want the unknown-difficulty error")
	}
}

func attrsRow() itemdat.ItemData {
	return itemdat.ItemData{
		ID:             "SWORD",
		Name:           "Sword",
		ShortName:      "Sword",
		CursorGraphic:  "SWORD",
		ItemType:       "Sword",
		Class:          "Weapon",
		MinDamage:      2,
		MaxDamage:      10,
		MinArmor:       0,
		MaxArmor:       0,
		MinStrength:    18,
		MinMagic:       0,
		MinDexterity:   0,
		Durability:     60,
		Value:          1200,
		MiscID:         "NONE",
		UniqueBaseItem: "SWORD",
	}
}

func assertRowCopy(t *testing.T, attrs itemgen.ItemAttrs) {
	t.Helper()

	if attrs.MinDamage != 2 || attrs.MaxDamage != 10 || attrs.Durability != 60 || attrs.MaxDur != 60 {
		t.Fatalf("attrs = %+v, want the row damage and durability copied", attrs)
	}
	if attrs.Value != 1200 || attrs.Quality != itemgen.QualityNormal {
		t.Fatalf("attrs = %+v, want value 1200 and NORMAL quality", attrs)
	}
	if attrs.MinStrength != 18 || attrs.Name != "Sword" || attrs.ItemType != "Sword" {
		t.Fatalf("attrs = %+v, want the requirement and identity fields", attrs)
	}
}

func assertGoldValue(t *testing.T, attrs itemgen.ItemAttrs, err error, want int) {
	t.Helper()

	if err != nil {
		t.Fatalf("GetItemAttrs: %v", err)
	}
	if attrs.Value != want {
		t.Fatalf("value = %d, want %d", attrs.Value, want)
	}
	if attrs.Quality != itemgen.QualityNormal {
		t.Fatalf("quality = %d, want NORMAL", attrs.Quality)
	}
}
