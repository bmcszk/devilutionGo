package txtdata_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

func TestLoadClasses(t *testing.T) {
	// given
	const classdatTSV = "className\tfolderName\tportrait\tinv\r\n" +
		"Warrior\twarr\t0\tinv\r\n" +
		"Rogue\trogue\t1\tinv_rog\r\n"

	tbl := mustParse(t, []byte(classdatTSV))

	// when
	classes, err := txtdata.LoadClasses(tbl)
	if err != nil {
		t.Fatalf("txtdata.LoadClasses: %v", err)
	}

	// then
	if len(classes) != 2 {
		t.Fatalf("classes = %d, want 2", len(classes))
	}

	if classes[0].ClassName != "Warrior" || classes[0].FolderName != "warr" {
		t.Fatalf("class0 = %+v", classes[0])
	}
}

func TestLoadClassTables(t *testing.T) {
	// given
	const classTableTSV = "Attribute\tValue\r\nclassFlags\tCriticalStrike\r\n"

	tbl := mustParse(t, []byte(classTableTSV))

	// when
	tables, err := txtdata.LoadClassTables(tbl)
	if err != nil {
		t.Fatalf("txtdata.LoadClassTables: %v", err)
	}

	// then
	if len(tables) != 1 || tables[0][0] != "classFlags" || tables[0][1] != "CriticalStrike" {
		t.Fatalf("tables = %v", tables)
	}
}
