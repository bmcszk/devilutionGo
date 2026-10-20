// Package textdat ports the text catalog: textdat.tsv + textdat-unused.tsv
// (Source/tables/textdat.cpp:106 loader; _speech_id textdat.h:20-448).
package textdat

import (
	"errors"
	"fmt"
	"strings"
)

// TextRow is one textdat.tsv row: txtstrid|txtstr|scrlltxt|sfxnr.
// TxtStrID is the _speech_id enum NAME (e.g. TEXT_KING1); its index is the
// 0-based row order (textdat.h:20-448 declares them in table order).
type TextRow struct {
	TxtStrID string
	TxtStr   string
	Scrlltxt bool   // scroll-text flag (InitQTextMsg speech)
	Sfxnr    string // sfx id name, references effects.tsv ids ("" = none)
}

// ParseTextdat parses a textdat table in row order. TxtStrID maps the
// _speech_id enum order (dense 0-based upstream).
func ParseTextdat(data []byte) ([]TextRow, error) {
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
	out := make([]TextRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, TextRow{
			TxtStrID: r[0],
			TxtStr:   r[1],
			Scrlltxt: r[2] == "1" || strings.EqualFold(r[2], "true"),
			Sfxnr:    r[3],
		})
	}
	return out, nil
}
