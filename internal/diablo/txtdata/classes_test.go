package txtdata_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// Header + two rows copied from assets/txtdata/classes/classdat.tsv.
const classdatTSV = "className\tfolderName\tportrait\tinv\r\n" +
	"Warrior\twarr\t0\tinv\r\n" +
	"Rogue\trogue\t1\tinv_rog\r\n"

func TestLoadClasses(t *testing.T) {
	tbl, err := enginetxtdata.Parse([]byte(classdatTSV))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	classes, err := txtdata.LoadClasses(tbl)
	if err != nil {
		t.Fatalf("LoadClasses: %v", err)
	}
	if len(classes) != 2 {
		t.Fatalf("classes = %d, want 2", len(classes))
	}
	if classes[0].ClassName != "Warrior" || classes[0].FolderName != "warr" {
		t.Fatalf("class0 = %+v", classes[0])
	}
}

func TestLoadClassTables(t *testing.T) {
	tbl, err := enginetxtdata.Parse([]byte("Attribute\tValue\r\nclassFlags\tCriticalStrike\r\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	tables, err := txtdata.LoadClassTables(tbl)
	if err != nil {
		t.Fatalf("LoadClassTables: %v", err)
	}
	if len(tables) != 1 || tables[0][0] != "classFlags" || tables[0][1] != "CriticalStrike" {
		t.Fatalf("tables = %v", tables)
	}
}
