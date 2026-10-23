// Package quests ports the quest table loader: questdat.tsv rows
// quest_id / quest_state enums, and InitialiseQuestPools.
package quests

import (
	"errors"
	"strconv"
	"strings"
)

// QuestID mirrors quest_id in declaration order.
type QuestID int8

const (
	QRock QuestID = iota
	QMushroom
	QGarbud
	QZhar
	QVeil
	QDiablo
	QButcher
	QLtbanner
	QBlind
	QBlood
	QAnvil
	QWarlord
	QSkelking
	QPwater
	QSchamb
	QBetrayer
	QGrave
	QFarmer
	QGirl
	QTrader
	QDefiler
	QNakrul
	QCornstn
	QJersey

	QInvalid QuestID = -1
)

// QuestIDCount is the named-entry count (Q_INVALID excluded).
const QuestIDCount = 24

// QuestState mirrors quest_state.
type QuestState uint8

const (
	QuestNotAvail QuestState = iota // quest did not spawn this game
	QuestInit                       // spawned, waiting to trigger
	QuestActive                     // in progress
	QuestDone                       // log closed and finished
	// Hive states occupy 7-10 upstream (explicit values kept).
	QuestHiveTease1 QuestState = 7
	QuestHiveTease2 QuestState = 8
	QuestHiveActive QuestState = 9
	QuestHiveDone   QuestState = 10

	QuestInvalid QuestState = 0xFF
)

// QuestRow is one questdat.tsv row
// qdlvl|qdmultlvl|qlvlt|bookOrder|qdrnd|qslvl|isSinglePlayerOnly|qdmsg|qlstr.
type QuestRow struct {
	ID                 string // qlstr column is the display name; the row key is position
	QDLvl              int
	QDMultLvl          int
	QLvlT              string
	BookOrder          int
	QDRnd              int
	QSLvl              string
	IsSinglePlayerOnly bool
	QDMsg              string
	QLStr              string
}

// ParseQuestdat maps row i to QuestID(i) — the row key is position.
func ParseQuestdat(data []byte) ([]QuestRow, error) {
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
			return nil, errors.New("quests: ragged row")
		}
		rows = append(rows, cols)
	}
	if header == nil {
		return nil, errors.New("quests: empty table")
	}
	out := make([]QuestRow, 0, len(rows))
	for _, r := range rows {
		row := QuestRow{
			ID:                 r[0],
			QLvlT:              r[2],
			QSLvl:              r[5],
			IsSinglePlayerOnly: r[6] == "true",
			QDMsg:              r[7],
			QLStr:              r[8],
		}
		var err error
		if row.QDLvl, err = strconv.Atoi(r[0]); err != nil {
			return nil, errors.New("quests: qdlvl not numeric")
		}
		if row.QDMultLvl, err = strconv.Atoi(r[1]); err != nil {
			return nil, errors.New("quests: qdmultlvl not numeric")
		}
		if row.BookOrder, err = strconv.Atoi(r[3]); err != nil {
			return nil, errors.New("quests: bookOrder not numeric")
		}
		if row.QDRnd, err = strconv.Atoi(r[4]); err != nil {
			return nil, errors.New("quests: qdrnd not numeric")
		}
		out = append(out, row)
	}
	return out, nil
}

// NotAvailableInSpawn reports V12.9: all quests are QUEST_NOTAVAIL when the
// spawn (shareware) gate is on.
func NotAvailableInSpawn(QuestRow) bool { return true }

// pickAmong mirrors DiabloGenerator::pickRandomlyAmong
// on the global LCG: GenerateRnd with the
// negative-clamp the C++ applies.
func pickAmong(n int, rng func(int32) int32) int {
	idx := rng(int32(n))
	if idx < 0 {
		idx = 0
	}
	return int(idx)
}

// InitialiseQuestPools ports: five pickRandomlyAmong
// pools (four when the seed hits the vanilla 988045466 bug path, which
// discards one draw instead). Each picked quest becomes QUEST_NOTAVAIL.
func InitialiseQuestPools(seed uint32, setRndSeed func(uint32), rnd func(int32) int32) []QuestState {
	setRndSeed(seed)
	qs := make([]QuestState, QuestIDCount)
	for i := range qs {
		qs[i] = QuestInit
	}
	pick := func(ids ...QuestID) {
		qs[ids[pickAmong(len(ids), rnd)]] = QuestNotAvail
	}
	pick(QSkelking, QPwater)
	if seed == 988045466 {
		// Vanilla reads QuestGroup1[-2] here; net effect: one discarded draw
		// and The Butcher stays available.
		rnd(1)
	} else {
		pick(QButcher, QLtbanner, QGarbud)
	}
	pick(QBlind, QRock, QBlood)
	pick(QMushroom, QZhar, QAnvil)
	pick(QVeil, QWarlord)
	return qs
}
