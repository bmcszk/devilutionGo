package txtdata_test

import (
	"strings"
	"testing"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

const sampleTSV = "id\tname\tbookCost10\n" +
	"1\tFirebolt\t100\n" +
	"2\tHealing\t200"

func TestParseHeaderAndRows(t *testing.T) {
	// sampleTSV holds three columns and two data rows.
	// when the payload is parsed
	tbl, err := enginetxtdata.Parse([]byte(sampleTSV))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// then
	if len(tbl.Header) != 3 || tbl.Header[1] != "name" {
		t.Fatalf("header = %v", tbl.Header)
	}

	if len(tbl.Rows) != 2 || tbl.Rows[0][1] != "Firebolt" {
		t.Fatalf("rows = %v", tbl.Rows)
	}
}

func TestParseStripsBOM(t *testing.T) {
	// given
	withBOM := append([]byte{0xEF, 0xBB, 0xBF}, []byte(sampleTSV)...)
	// when
	tbl, err := enginetxtdata.Parse(withBOM)
	if err != nil {
		t.Fatalf("Parse with BOM: %v", err)
	}
	// then
	if tbl.Header[0] != "id" {
		t.Fatalf("BOM not stripped: header[0] = %q", tbl.Header[0])
	}
}

func TestParseCRLFSeparated(t *testing.T) {
	// given
	crlf := strings.ReplaceAll(sampleTSV, "\n", "\r\n")
	// when
	tbl, err := enginetxtdata.Parse([]byte(crlf))
	if err != nil {
		t.Fatalf("Parse CRLF: %v", err)
	}
	// then
	if len(tbl.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(tbl.Rows))
	}
}

func TestParseEmptyIsError(t *testing.T) {
	// given one empty or BOM-only payload per case
	inputs := []string{"", "\xef\xbb\xbf"}

	// when every payload is parsed
	errs := make([]error, len(inputs))
	for i, in := range inputs {
		_, errs[i] = enginetxtdata.Parse([]byte(in))
	}

	// then every payload reports an empty-file error
	for i, in := range inputs {
		t.Run(in, func(t *testing.T) {
			if errs[i] == nil {
				t.Fatal("Parse = nil error, want empty-file error")
			}
		})
	}
}

func TestColumnLookup(t *testing.T) {
	// given
	tbl, _ := enginetxtdata.Parse([]byte(sampleTSV))
	// when
	i, err := tbl.Column("bookCost10")
	_, nopeErr := tbl.Column("nope")
	// then
	if err != nil || i != 2 {
		t.Fatalf("Column(bookCost10) = %d, %v", i, err)
	}

	if nopeErr == nil {
		t.Fatal("Column(nope) = nil error, want error")
	}
}
