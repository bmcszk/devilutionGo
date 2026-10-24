// Package textdat loads the text catalog tables (textdat.tsv,
// textdat-unused.tsv).
package textdat

import (
	"strings"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// textRow is one textdat.tsv row: txtstrid|txtstr|scrlltxt|sfxnr.
// TxtStrID is the speech id enum name; the row index is its enum value.
type textRow struct {
	TxtStrID string
	TxtStr   string
	Scrlltxt bool
	Sfxnr    string // effect id; "" means no sound
}

// TextRow is the exported textdat row type
// (upstream textdat struct, textdat.cpp / textdat.tsv columns).
type TextRow = textRow

// ParseTextdat parses a textdat table into one textRow per row, in file
// order.
func ParseTextdat(data []byte) ([]textRow, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, err
	}
	out := make([]textRow, 0, len(tbl.Rows))
	for _, r := range tbl.Rows {
		out = append(out, textRow{
			TxtStrID: r[0],
			TxtStr:   r[1],
			Scrlltxt: r[2] == "1" || strings.EqualFold(r[2], "true"),
			Sfxnr:    r[3],
		})
	}
	return out, nil
}
