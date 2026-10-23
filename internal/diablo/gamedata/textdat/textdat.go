// Package textdat loads the text catalog tables (textdat.tsv,
// textdat-unused.tsv).
package textdat

import (
	"errors"
	"fmt"
	"strings"
)

// textRow is one textdat.tsv row: txtstrid|txtstr|scrlltxt|sfxnr.
// TxtStrID is the speech id enum name; the row index is its enum value.
type textRow struct {
	TxtStrID string
	TxtStr   string
	Scrlltxt bool
	Sfxnr    string // effect id; "" means no sound
}

// ParseTextdat parses a textdat table into one textRow per row, in file
// order.
func ParseTextdat(data []byte) ([]textRow, error) {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var rows [][]string
	var header []string
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		cols := strings.Split(ln, "\t")
		if header == nil {
			header = cols
			continue
		}
		if len(cols) != len(header) {
			return nil, fmt.Errorf("textdat: row has %d cols, want %d: %q", len(cols), len(header), ln)
		}
		rows = append(rows, cols)
	}
	if header == nil {
		return nil, errors.New("textdat: empty table")
	}
	out := make([]textRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, textRow{
			TxtStrID: r[0],
			TxtStr:   r[1],
			Scrlltxt: r[2] == "1" || strings.EqualFold(r[2], "true"),
			Sfxnr:    r[3],
		})
	}
	return out, nil
}
