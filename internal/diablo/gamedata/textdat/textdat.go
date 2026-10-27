// Package textdat loads the text catalog tables (textdat.tsv,
// textdat-unused.tsv).
package textdat

import (
	"fmt"
	"strings"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// TextRow is one textdat.tsv row: txtstrid|txtstr|scrlltxt|sfxnr.
// TxtStrID is the speech id enum name; the row index is its enum value.
type TextRow struct {
	TxtStrID string
	TxtStr   string
	Scrlltxt bool
	Sfxnr    string // effect id; "" means no sound
}

// ParseTextdat parses a textdat table into one TextRow per row, in file
// order.
func ParseTextdat(data []byte) ([]TextRow, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse textdat table: %w", err)
	}

	out := make([]TextRow, 0, len(tbl.Rows))
	for _, r := range tbl.Rows {
		out = append(out, TextRow{
			TxtStrID: r[0],
			TxtStr:   r[1],
			Scrlltxt: r[2] == "1" || strings.EqualFold(r[2], "true"),
			Sfxnr:    r[3],
		})
	}

	return out, nil
}
