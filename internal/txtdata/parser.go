// Package txtdata parses devilutionX txtdata TSV tables
// (assets/txtdata/Readme.md EBNF; Source/data/file.hpp DataFile).
package txtdata

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// utf8BOM is stripped before parsing, as DataFile's constructor does
// (data/file.hpp:48-52).
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Table is one parsed txtdata file: the header record names the columns and
// every following record is a data row (Readme.md: first record SHOULD be
// the header; all files contain at least one record).
type Table struct {
	Header []string
	Rows   [][]string
}

// Column returns the index of the named header column.
func (t *Table) Column(name string) (int, error) {
	for i, h := range t.Header {
		if h == name {
			return i, nil
		}
	}
	return 0, fmt.Errorf("txtdata: no column %q", name)
}

// Parse decodes a whole txtdata file. A trailing newline on the last record
// is optional (Readme.md format spec); values containing tabs are rejected.
func Parse(data []byte) (*Table, error) {
	data = bytes.TrimPrefix(data, utf8BOM)
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, errors.New("txtdata: empty file")
	}
	records := splitRecords(data)
	if len(records) == 0 {
		return nil, errors.New("txtdata: no records")
	}
	t := &Table{Header: splitFields(records[0])}
	for _, rec := range records[1:] {
		fields := splitFields(rec)
		for _, f := range fields {
			if strings.ContainsRune(f, '\t') {
				return nil, fmt.Errorf("txtdata: field value contains tab: %q", f)
			}
		}
		t.Rows = append(t.Rows, fields)
	}
	return t, nil
}

// splitRecords splits on newlines, tolerating CRLF (upstream corpus uses
// CRLF) and a missing trailing newline.
func splitRecords(data []byte) []string {
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSuffix(l, "\r")
		if l == "" {
			continue
		}
		out = append(out, l)
	}
	return out
}

func splitFields(record string) []string {
	return strings.Split(record, "\t")
}
