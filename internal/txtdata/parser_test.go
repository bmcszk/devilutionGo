package txtdata_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/txtdata"
)

const sampleTSV = "id\tname\tbookCost10\n" +
	"1\tFirebolt\t100\n" +
	"2\tHealing\t200"

func TestParseHeaderAndRows(t *testing.T) {
	tbl, err := txtdata.Parse([]byte(sampleTSV))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(tbl.Header) != 3 || tbl.Header[1] != "name" {
		t.Fatalf("header = %v", tbl.Header)
	}
	if len(tbl.Rows) != 2 || tbl.Rows[0][1] != "Firebolt" {
		t.Fatalf("rows = %v", tbl.Rows)
	}
}

func TestParseStripsBOM(t *testing.T) {
	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, []byte(sampleTSV)...)
	tbl, err := txtdata.Parse(withBOM)
	if err != nil {
		t.Fatalf("Parse with BOM: %v", err)
	}
	if tbl.Header[0] != "id" {
		t.Fatalf("BOM not stripped: header[0] = %q", tbl.Header[0])
	}
}

func TestParseCRLEparated(t *testing.T) {
	crlf := strings.ReplaceAll(sampleTSV, "\n", "\r\n")
	tbl, err := txtdata.Parse([]byte(crlf))
	if err != nil {
		t.Fatalf("Parse CRLF: %v", err)
	}
	if len(tbl.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(tbl.Rows))
	}
}

func TestParseEmptyIsError(t *testing.T) {
	for _, in := range []string{"", "\xef\xbb\xbf"} {
		if _, err := txtdata.Parse([]byte(in)); err == nil {
			t.Fatalf("Parse(%q) = nil error, want empty-file error", in)
		}
	}
}

func TestColumnLookup(t *testing.T) {
	tbl, _ := txtdata.Parse([]byte(sampleTSV))
	if i, err := tbl.Column("bookCost10"); err != nil || i != 2 {
		t.Fatalf("Column(bookCost10) = %d, %v", i, err)
	}
	if _, err := tbl.Column("nope"); err == nil {
		t.Fatal("Column(nope) = nil error, want error")
	}
}
