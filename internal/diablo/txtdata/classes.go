// Package txtdata reads the Diablo-specific TSV tables; the table mechanism
// (parsing, mod overrides) lives in internal/engine/txtdata.
package txtdata

import (
	"fmt"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// ClassData is one classdat.tsv row.
// ClassData is one classes.tsv row (classes.cpp).
type ClassData struct {
	ClassName  string
	FolderName string
}

// LoadClasses parses classdat.tsv (header className,folderName,portrait,inv).
// LoadClasses parses classes.tsv rows (classes.cpp ClassData table).
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
// LoadClassTables parses the per-class attribute tables
// (SourceD1 monster.cpp-style two-column key/value tables).
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
