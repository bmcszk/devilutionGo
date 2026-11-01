package itemgen

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
)

// Difficulty levels (levels/gendung_defs.hpp:39-44).
const (
	DifficultyNormal    = 0
	DifficultyNightmare = 1
	DifficultyHell      = 2
)

// GoldMaxLimit is GOLD_MAX_LIMIT (items.h:31).
const GoldMaxLimit = 5000

// difficultyLevelOffset is the 16/32 level bump of the gold roll
// (items.cpp:2381-2387).
const difficultyLevelOffset = 16

// ItemAttrs is the projected item state GetItemAttrs fills in
// (items.cpp:2331-2392).
type ItemAttrs struct {
	ItemType     string
	Cursor       string
	Name         string
	Identified   string
	Class        string
	MinDamage    int
	MaxDamage    int
	AC           int
	MiscID       string
	Spell        string
	Value        int
	Durability   int
	MaxDur       int
	MinStrength  int
	MinMagic     int
	MinDexterity int
	Quality      int
	// GoldCursor is GetGoldCursor of the rolled amount; it stays 0 for rows
	// that are not gold (:2391).
	GoldCursor int
}

// GetItemAttrs ports GetItemAttrs (items.cpp:2331-2392): copy the base row onto
// the item, roll the AC band, and override the value with the gold amount when
// the row is gold. Book spells and Hellfire oils stay caller steps.
func GetItemAttrs(
	row itemdat.ItemData,
	difficulty, currLevel int,
	hellDungeon bool,
	rnd func(int32) int32,
) (ItemAttrs, error) {
	attrs := ItemAttrs{
		ItemType:     row.ItemType,
		Cursor:       row.CursorGraphic,
		Name:         row.Name,
		Identified:   row.Name,
		Class:        row.Class,
		MinDamage:    row.MinDamage,
		MaxDamage:    row.MaxDamage,
		AC:           rollArmor(row, rnd),
		MiscID:       row.MiscID,
		Spell:        row.Spell,
		Value:        row.Value,
		Durability:   row.Durability,
		MaxDur:       row.Durability,
		MinStrength:  row.MinStrength,
		MinMagic:     row.MinMagic,
		MinDexterity: row.MinDexterity,
		Quality:      QualityNormal,
	}
	if row.ItemType != "Gold" {
		return attrs, nil
	}

	value, err := goldAmount(difficulty, currLevel, hellDungeon, rnd)
	if err != nil {
		return ItemAttrs{}, err
	}
	attrs.Value = value
	attrs.GoldCursor = GetGoldCursor(value)

	return attrs, nil
}

// rollArmor is `iMinAC + GenerateRnd(iMaxAC - iMinAC + 1)` (:2338).
func rollArmor(row itemdat.ItemData, rnd func(int32) int32) int {
	return row.MinArmor + int(rnd(int32(row.MaxArmor-row.MinArmor+1)))
}

// goldAmount is the gold value roll (:2374-2391).
//
//nolint:revive // hellDungeon is upstream's leveltype == DTYPE_HELL gate.
func goldAmount(difficulty, currLevel int, hellDungeon bool, rnd func(int32) int32) (int, error) {
	var level int
	switch difficulty {
	case DifficultyNormal:
		level = currLevel
	case DifficultyNightmare:
		level = currLevel + difficultyLevelOffset
	case DifficultyHell:
		level = currLevel + 2*difficultyLevelOffset
	default:
		return 0, fmt.Errorf("unknown difficulty %d", difficulty)
	}

	rndv := 5*level + int(rnd(int32(10*level)))
	if hellDungeon {
		rndv += rndv / 8
	}

	return min(rndv, GoldMaxLimit), nil
}
