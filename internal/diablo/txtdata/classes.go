// Package txtdata ports the Diablo-specific table readers: the class domain
// tables (classdat loader, playerdat.cpp:345, plus the per-class loaders —
// attributes :171, starting_loadout :209, sprites :227, animations :251,
// sounds), the experience thresholds, the mana-cost formula and the spell
// data. The table MECHANISM (TSV parse, mod override) lives in
// internal/engine/txtdata.
package txtdata

import (
	"fmt"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// ClassData is one classdat.tsv row (playerdat.cpp:345 loader).
type ClassData struct {
	ClassName  string
	FolderName string
}

// ClassTables holds every per-class table for one class folder.
type ClassTables struct {
	Class           string
	Attributes      [][2]string // Attribute,Value pairs
	StartingLoadout [][2]string // Variable,Value pairs
}

// LoadClasses parses classdat.tsv (header className,folderName,portrait,inv).
func LoadClasses(t *enginetxtdata.Table) ([]ClassData, error) {
	ci, err := t.Column("className")
	if err != nil {
		return nil, err
	}
	fi, err := t.Column("folderName")
	if err != nil {
		return nil, err
	}
	var out []ClassData
	for _, row := range t.Rows {
		if ci >= len(row) || fi >= len(row) {
			return nil, fmt.Errorf("txtdata: classdat row too short: %v", row)
		}
		out = append(out, ClassData{ClassName: row[ci], FolderName: row[fi]})
	}
	return out, nil
}

// LoadClassTables parses one per-class table (Attribute,Value or
// Variable,Value two-column layout).
func LoadClassTables(t *enginetxtdata.Table) ([][2]string, error) {
	var out [][2]string
	for _, row := range t.Rows {
		if len(row) < 2 {
			return nil, fmt.Errorf("txtdata: class table row too short: %v", row)
		}
		out = append(out, [2]string{row[0], row[1]})
	}
	return out, nil
}
