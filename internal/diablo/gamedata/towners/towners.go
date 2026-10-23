// Package towners loads the town domain tables (towners.tsv,
// quest_dialog.tsv) and holds the town store price caps.
package towners

import (
	"strings"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// Caps the town domain enforces on vendor prices and gold piles.
const (
	MaxVendorValue   = 140000
	MaxVendorValueHF = 200000
	MaxGold          = 5000
	GoldCursorSmall  = 1000
	GoldCursorMedium = 2500
	GoldCursorLarge  = 5000
)

// Towner is one towners.tsv row; GossipTexts and AnimOrder come from
// comma-separated columns.
type Towner struct {
	Type        string
	Name        string
	X, Y        int
	Direction   string
	AnimWidth   int
	AnimPath    string
	AnimFrames  int
	AnimDelay   int
	GossipTexts []string
	AnimOrder   []string
}

// ParseTowners parses towners.tsv in row order; a trailing empty column is
// dropped.
func ParseTowners(data []byte) ([]Towner, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, err
	}
	rows := tbl.Rows
	out := make([]Towner, 0, len(rows))
	for _, r := range rows {
		o := Towner{
			Type:        r[0],
			Name:        r[1],
			Direction:   r[4],
			AnimPath:    r[6],
			GossipTexts: splitList(r[9]),
			AnimOrder:   splitList(r[10]),
		}
		// Numeric fields tolerate "" as 0: hf rows carry empty numeric cells.
		o.X = atoiOrZero(r[2])
		o.Y = atoiOrZero(r[3])
		o.AnimWidth = atoiOrZero(r[5])
		o.AnimFrames = atoiOrZero(r[7])
		o.AnimDelay = atoiOrZero(r[8])
		out = append(out, o)
	}
	return out, nil
}

func atoiOrZero(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0
		}
		n = n*10 + int(s[i]-'0')
	}
	return n
}

func splitList(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// QuestDialog is the parsed quest_dialog.tsv matrix: header column names
// (towner_type plus Q_* quest columns), rows starting with the towner type,
// and the column count shared by all rows.
type QuestDialog struct {
	Header []string
	Rows   [][]string
	Cols   int
}

func ParseQuestDialog(data []byte) (QuestDialog, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return QuestDialog{}, err
	}
	return QuestDialog{Header: tbl.Header, Rows: tbl.Rows, Cols: len(tbl.Header)}, nil
}
