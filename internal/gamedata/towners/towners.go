// Package towners ports the town domain tables: towners.tsv
// (Source/towners.cpp loader region) and quest_dialog.tsv (10x18 base /
// 13x25 hf), plus the store price caps (items.h:36-37).
package towners

import (
	"errors"
	"strings"
)

// Vendor/gold caps (items.h:36-37; items.h:28-30 MaxGold + cursor
// thresholds are re-declared here as the town-domain view).
const (
	MaxVendorValue   = 140000
	MaxVendorValueHF = 200000
	MaxGold          = 5000
	GoldCursorSmall  = 1000
	GoldCursorMedium = 2500
	GoldCursorLarge  = 5000
)

// Towner is one towners.tsv row (11 columns; gossipTexts and animOrder are
// comma lists).
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

// ParseTowners parses towners.tsv in row order. hf appends three TOWN_COW
// rows that carry one trailing empty column; tolerated and dropped.
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
		// Numeric fields tolerate "" as 0: the hf-appended TOWN_COW rows
		// carry empty numeric/path cells (upstream reader defaults).
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

// ParseQuestDialog parses quest_dialog.tsv: row 0 is the header
// (towner_type + Q_* columns); each data row is [towner_type, TEXT_*...].
// Returns the rows and the column count (18 base / 25 hf).
// QuestDialog is the parsed quest_dialog.tsv: header (towner_type + Q_*
// columns), data rows [towner_type, TEXT_*...], total column count
// (19 base / 25 hf).
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
