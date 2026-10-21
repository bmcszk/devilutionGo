// Package towners loads the town domain tables (towners.tsv,
// quest_dialog.tsv) and holds the town store price caps.
package towners

import (
	"errors"
	"strings"
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
		if len(cols) == len(header)+1 && cols[len(cols)-1] == "" {
			cols = cols[:len(header)]
		}
		if len(cols) != len(header) {
			return nil, errors.New("towners: ragged row")
		}
		rows = append(rows, cols)
	}
	if header == nil {
		return nil, errors.New("towners: empty table")
	}
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
	if s == "" {
		return 0
	}
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
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var rows [][]string
	var ncols int
	var header []string
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		cols := strings.Split(ln, "\t")
		if header == nil {
			header = cols
			ncols = len(cols)
			continue
		}
		if len(cols) != ncols {
			return QuestDialog{}, errors.New("towners: ragged dialog row")
		}
		rows = append(rows, cols)
	}
	if header == nil {
		return QuestDialog{}, errors.New("towners: empty dialog table")
	}
	return QuestDialog{Header: header, Rows: rows, Cols: ncols}, nil
}
