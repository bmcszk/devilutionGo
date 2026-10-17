// SpellData mirrors the fields LoadSpellData reads
// (Source/tables/spelldat.cpp:238-262); column names are the txtdata header.
package txtdata

import (
	"fmt"
	"strconv"
	"strings"
)

// SpellData is the typed mirror of upstream tables/spelldat.h:225-252 fields
// that the txtdata loader fills.
type SpellData struct {
	Name           string
	SoundID        string
	BookCost10     int
	StaffCost10    int
	ManaCost       int
	Flags          []string
	BookLevel      int
	StaffLevel     int
	MinInt         int
	Missiles       []string // filled with "Null" when the TSV lists fewer than 2
	ManaMultiplier int
	MinMana        int
	StaffMin       int
	StaffMax       int
}

// LoadSpellData mirrors LoadSpellData (spelldat.cpp:236-266): skip the id
// column, read the named fields in upstream order, pad missiles to two with
// "Null".
func LoadSpellData(t *Table) ([]SpellData, error) {
	out := make([]SpellData, 0, len(t.Rows)+1)
	for _, row := range t.Rows {
		s, err := spellFromRow(t, row)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func spellFromRow(t *Table, row []string) (SpellData, error) {
	var s SpellData
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
		return s, err // reader.advance() skips id (spelldat.cpp:245)
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
		// readEnumArray fillMissing=Null (spelldat.cpp:257)
		s.Missiles = append(s.Missiles, "Null")
	}
	return s, nil
}

// NullSpell mirrors AddNullSpell (spelldat.cpp:20-30): prepended before the
// TSV rows, SFX None, zero costs, Fire flag, both missiles Null.
func NullSpell() SpellData {
	return SpellData{
		Name:     "Null",
		SoundID:  "None",
		Flags:    []string{"Fire"},
		Missiles: []string{"Null", "Null"},
	}
}

// LoadSpellDataWithNull prepends the Null spell before the TSV rows
// (spelldat.cpp:235 AddNullSpell precedes the record loop).
func LoadSpellDataWithNull(t *Table) ([]SpellData, error) {
	out := []SpellData{NullSpell()}
	rest, err := LoadSpellData(t)
	if err != nil {
		return nil, err
	}
	return append(out, rest...), nil
}

// SpellIDByName finds a spell row by its id column value (the loader keys
// ParseSpellID on it, spelldat.cpp:225-229).
func SpellIDByName(t *Table, id string) (int, error) {
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
