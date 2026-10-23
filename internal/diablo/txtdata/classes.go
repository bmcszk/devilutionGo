// Package txtdata reads the Diablo-specific TSV tables; the table mechanism
// (parsing, mod overrides) lives in internal/engine/txtdata.
package txtdata

import (
	"fmt"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// classData is one classdat.tsv row.
type classData struct {
	ClassName  string
	FolderName string
}

// classTables holds every per-class table for one class folder.
//
// TODO remove nolint:unused // Volume II per-class tables land with the errata bean (devilGo-hy8e)
type classTables struct {
	Class           string
	Attributes      [][2]string // Attribute,Value pairs
	StartingLoadout [][2]string // Variable,Value pairs
}

// loadClasses parses classdat.tsv (header className,folderName,portrait,inv).
func loadClasses(t *enginetxtdata.Table) ([]classData, error) {
	ci, err := t.Column("className")
	if err != nil {
		return nil, err
	}
	fi, err := t.Column("folderName")
	if err != nil {
		return nil, err
	}
	var out []classData
	for _, row := range t.Rows {
		if ci >= len(row) || fi >= len(row) {
			return nil, fmt.Errorf("txtdata: classdat row too short: %v", row)
		}
		out = append(out, classData{ClassName: row[ci], FolderName: row[fi]})
	}
	return out, nil
}

// loadClassTables parses one per-class table (Attribute,Value or
// Variable,Value two-column layout).
func loadClassTables(t *enginetxtdata.Table) ([][2]string, error) {
	var out [][2]string
	for _, row := range t.Rows {
		if len(row) < 2 {
			return nil, fmt.Errorf("txtdata: class table row too short: %v", row)
		}
		out = append(out, [2]string{row[0], row[1]})
	}
	return out, nil
}
