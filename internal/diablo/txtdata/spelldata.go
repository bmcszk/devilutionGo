package txtdata

import (
	"fmt"
	"strconv"
	"strings"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// spellData is one row of the spell table.
type spellData struct {
	Name           string
	SoundID        string
	BookCost10     int
	StaffCost10    int
	ManaCost       int
	Flags          []string
	BookLevel      int
	StaffLevel     int
	MinInt         int
	Missiles       []string // always holds at least two entries; missing ones are "Null"
	ManaMultiplier int
	MinMana        int
	StaffMin       int
	StaffMax       int
}

// loadSpellData returns one spellData per table row.
func loadSpellData(t *enginetxtdata.Table) ([]spellData, error) {
	out := make([]spellData, 0, len(t.Rows)+1)
	for _, row := range t.Rows {
		s, err := spellFromRow(t, row)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func spellFromRow(t *enginetxtdata.Table, row []string) (spellData, error) {
	var s spellData
	cell := func(name string) (string, error) {
		i, err := t.Column(name)
		if err != nil {
			return "", err
		}
		if i >= len(row) {
			return "", fmt.Errorf("txtdata: row too short for %q", name)
		}
		return row[i], nil
	}
	number := func(name string) (int, error) {
		v, err := cell(name)
		if err != nil {
			return 0, err
		}
		return strconv.Atoi(v)
	}
	var err error
	if _, err = cell("id"); err != nil {
		return s, err // the id column is validated but not stored
	}
	if s.Name, err = cell("name"); err != nil {
		return s, err
	}
	if s.SoundID, err = cell("soundId"); err != nil {
		return s, err
	}
	nums := []struct {
		name string
		dst  *int
	}{
		{"bookCost10", &s.BookCost10},
		{"staffCost10", &s.StaffCost10},
		{"manaCost", &s.ManaCost},
		{"bookLevel", &s.BookLevel},
		{"staffLevel", &s.StaffLevel},
		{"minIntelligence", &s.MinInt},
		{"manaMultiplier", &s.ManaMultiplier},
		{"minMana", &s.MinMana},
		{"staffMin", &s.StaffMin},
		{"staffMax", &s.StaffMax},
	}
	for _, f := range nums {
		if *f.dst, err = number(f.name); err != nil {
			return s, err
		}
	}
	flags, err := cell("flags")
	if err != nil {
		return s, err
	}
	if flags != "" {
		s.Flags = strings.Split(flags, ",")
	}
	missiles, err := cell("missiles")
	if err != nil {
		return s, err
	}
	s.Missiles = strings.Split(missiles, ",")
	for len(s.Missiles) < 2 {
		s.Missiles = append(s.Missiles, "Null")
	}
	return s, nil
}

// nullSpell is the placeholder spell prepended to every spell table.
func nullSpell() spellData {
	return spellData{
		Name:     "Null",
		SoundID:  "None",
		Flags:    []string{"Fire"},
		Missiles: []string{"Null", "Null"},
	}
}

// loadSpellDataWithNull returns the spell table with the Null spell first.
func loadSpellDataWithNull(t *enginetxtdata.Table) ([]spellData, error) {
	out := []spellData{nullSpell()}
	rest, err := loadSpellData(t)
	if err != nil {
		return nil, err
	}
	return append(out, rest...), nil
}

// spellIDByName returns the row index of the spell with the given id.
func spellIDByName(t *enginetxtdata.Table, id string) (int, error) {
	ii, err := t.Column("id")
	if err != nil {
		return -1, err
	}
	for i, row := range t.Rows {
		if ii < len(row) && row[ii] == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("txtdata: unknown spell id %q", id)
}
