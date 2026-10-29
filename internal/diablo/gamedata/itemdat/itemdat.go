// Package itemdat loads the item data tables (itemdat.tsv, item_prefixes.tsv,
// item_suffixes.tsv, unique_itemdat.tsv) and holds the Go registry replacing
// the C++ itemdat.cpp tables.
package itemdat

import (
	"fmt"
	"strings"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// ItemData is one itemdat.tsv row (23 columns, exact order).
type ItemData struct {
	ID              string // id
	DropRate        int    // dropRate
	Class           string // class
	EquipType       string // equipType
	CursorGraphic   string // cursorGraphic
	ItemType        string // itemType
	UniqueBaseItem  string // uniqueBaseItem
	Name            string // name
	ShortName       string // shortName
	MinMonsterLevel int    // minMonsterLevel
	Durability      int    // durability
	MinDamage       int    // minDamage
	MaxDamage       int    // maxDamage
	MinArmor        int    // minArmor
	MaxArmor        int    // maxArmor
	MinStrength     int    // minStrength
	MinMagic        int    // minMagic
	MinDexterity    int    // minDexterity
	SpecialEffects  string // specialEffects
	MiscID          string // miscId
	Spell           string // spell
	Usable          bool   // usable
	Value           int    // value
}

// ItemPrefix is one item_prefixes.tsv row.
type ItemPrefix struct {
	Name        string   // name
	Power       string   // power
	PowerValue1 int      // power.value1
	PowerValue2 int      // power.value2
	MinLevel    int      // minLevel
	ItemTypes   []string // itemTypes (comma-separated in TSV)
	Alignment   string   // alignment
	Chance      int      // chance
	Useful      bool     // useful
	MinVal      int      // minVal
	MaxVal      int      // maxVal
	MultVal     int      // multVal
}

// ItemSuffix is one item_suffixes.tsv row (same structure as ItemPrefix).
type ItemSuffix struct {
	Name        string   // name
	Power       string   // power
	PowerValue1 int      // power.value1
	PowerValue2 int      // power.value2
	MinLevel    int      // minLevel
	ItemTypes   []string // itemTypes (comma-separated in TSV)
	Alignment   string   // alignment
	Chance      int      // chance
	Useful      bool     // useful
	MinVal      int      // minVal
	MaxVal      int      // maxVal
	MultVal     int      // multVal
}

// UniqueItemData is one unique_itemdat.tsv row (up to 6 powers).
type UniqueItemData struct {
	Name           string // name
	CursorGraphic  string // cursorGraphic
	UniqueBaseItem string // uniqueBaseItem
	MinLevel       int    // minLevel
	Value          int    // value
	Power0         string // power0
	Power0Value1   int    // power0.value1
	Power0Value2   int    // power0.value2
	Power1         string // power1
	Power1Value1   int    // power1.value1
	Power1Value2   int    // power1.value2
	Power2         string // power2
	Power2Value1   int    // power2.value1
	Power2Value2   int    // power2.value2
	Power3         string // power3
	Power3Value1   int    // power3.value1
	Power3Value2   int    // power3.value2
	Power4         string // power4
	Power4Value1   int    // power4.value1
	Power4Value2   int    // power4.value2
	Power5         string // power5
	Power5Value1   int    // power5.value1
	Power5Value2   int    // power5.value2
}

// LoadItemDat parses the base itemdat.tsv.
func LoadItemDat(data []byte) ([]ItemData, error) {
	t, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse itemdat: %w", err)
	}
	items := make([]ItemData, 0, len(t.Rows))
	for _, row := range t.Rows {
		if len(row) < 23 {
			continue // skip malformed
		}
		usable := row[21] == "true"
		items = append(items, ItemData{
			ID:              row[0],
			DropRate:        parseInt(row[1]),
			Class:           row[2],
			EquipType:       row[3],
			CursorGraphic:   row[4],
			ItemType:        row[5],
			UniqueBaseItem:  row[6],
			Name:            row[7],
			ShortName:       row[8],
			MinMonsterLevel: parseInt(row[9]),
			Durability:      parseInt(row[10]),
			MinDamage:       parseInt(row[11]),
			MaxDamage:       parseInt(row[12]),
			MinArmor:        parseInt(row[13]),
			MaxArmor:        parseInt(row[14]),
			MinStrength:     parseInt(row[15]),
			MinMagic:        parseInt(row[16]),
			MinDexterity:    parseInt(row[17]),
			SpecialEffects:  row[18],
			MiscID:          row[19],
			Spell:           row[20],
			Usable:          usable,
			Value:           parseInt(row[22]),
		})
	}
	return items, nil
}

// LoadItemPrefixes parses item_prefixes.tsv.
func LoadItemPrefixes(data []byte) ([]ItemPrefix, error) {
	return loadAffixes(data, func(row []string) ItemPrefix {
		return ItemPrefix{
			Name: row[0], Power: row[1], PowerValue1: parseInt(row[2]), PowerValue2: parseInt(row[3]),
			MinLevel: parseInt(row[4]), ItemTypes: splitItemTypes(row[5]), Alignment: row[6],
			Chance: parseInt(row[7]), Useful: row[8] == "true",
			MinVal: parseInt(row[9]), MaxVal: parseInt(row[10]), MultVal: parseInt(row[11]),
		}
	})
}

// LoadItemSuffixes parses item_suffixes.tsv (same format as prefixes).
func LoadItemSuffixes(data []byte) ([]ItemSuffix, error) {
	return loadAffixes(data, func(row []string) ItemSuffix {
		return ItemSuffix{
			Name: row[0], Power: row[1], PowerValue1: parseInt(row[2]), PowerValue2: parseInt(row[3]),
			MinLevel: parseInt(row[4]), ItemTypes: splitItemTypes(row[5]), Alignment: row[6],
			Chance: parseInt(row[7]), Useful: row[8] == "true",
			MinVal: parseInt(row[9]), MaxVal: parseInt(row[10]), MultVal: parseInt(row[11]),
		}
	})
}

// LoadUniqueItemDat parses unique_itemdat.tsv.
func LoadUniqueItemDat(data []byte) ([]UniqueItemData, error) {
	t, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse unique_itemdat: %w", err)
	}
	uniques := make([]UniqueItemData, 0, len(t.Rows))
	for _, row := range t.Rows {
		if len(row) < 23 {
			continue // skip malformed
		}
		uniques = append(uniques, UniqueItemData{
			Name:           row[0],
			CursorGraphic:  row[1],
			UniqueBaseItem: row[2],
			MinLevel:       parseInt(row[3]),
			Value:          parseInt(row[4]),
			Power0:         row[5],
			Power0Value1:   parseInt(row[6]),
			Power0Value2:   parseInt(row[7]),
			Power1:         row[8],
			Power1Value1:   parseInt(row[9]),
			Power1Value2:   parseInt(row[10]),
			Power2:         row[11],
			Power2Value1:   parseInt(row[12]),
			Power2Value2:   parseInt(row[13]),
			Power3:         row[14],
			Power3Value1:   parseInt(row[15]),
			Power3Value2:   parseInt(row[16]),
			Power4:         row[17],
			Power4Value1:   parseInt(row[18]),
			Power4Value2:   parseInt(row[19]),
			Power5:         row[20],
			Power5Value1:   parseInt(row[21]),
			Power5Value2:   parseInt(row[22]),
		})
	}
	return uniques, nil
}

// loadAffixes is the shared loader for the prefix/suffix tables (identical format).
func loadAffixes[T any](data []byte, rowTo func([]string) T) ([]T, error) {
	t, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse affix table: %w", err)
	}
	out := make([]T, 0, len(t.Rows))
	for _, row := range t.Rows {
		if len(row) < 12 {
			continue
		}
		out = append(out, rowTo(row))
	}
	return out, nil
}

// splitItemTypes splits a comma-separated itemTypes string.
func splitItemTypes(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// parseInt safely parses an integer, returns 0 on error.
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	var v int
	// best-effort: malformed numeric cells intentionally parse as 0, per upstream behavior
	_, _ = fmt.Sscanf(s, "%d", &v)
	return v
}
