// Package quests ports the quest table loader: questdat.tsv rows
// quest_id / quest_state enums, and initialiseQuestPools.
package quests

import (
	"errors"
	"strconv"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// questID mirrors quest_id in declaration order.
type questID int8

const (
	QRock questID = iota
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

	QInvalid questID = -1
)

// QuestIDCount is the named-entry count (Q_INVALID excluded).
const QuestIDCount = 24

// questState mirrors quest_state.
type questState uint8

const (
	QuestNotAvail questState = iota // quest did not spawn this game
	QuestInit                       // spawned, waiting to trigger
	QuestActive                     // in progress
	QuestDone                       // log closed and finished
	// Hive states occupy 7-10 upstream (explicit values kept).
	QuestHiveTease1 questState = 7
	QuestHiveTease2 questState = 8
	QuestHiveActive questState = 9
	QuestHiveDone   questState = 10

	QuestInvalid questState = 0xFF
)

// questRow is one questdat.tsv row
// (columns qdlvl|qdmultlvl|qlvlt|bookOrder|qdrnd|qslvl|isSinglePlayerOnly|qdmsg|qlstr).
// Column mapping: r[0] is qdlvl — the dungeon LEVEL number (int), not a name;
// qlstr (last column) is the display string. Rows are keyed by position:
// upstream emplaces each record into the QuestsData vector in file order
// (Source/tables/questdat.cpp:35-53, LoadQuestDatFromFile; qdlvl read via
// reader.readInt at questdat.cpp:45).
type questRow struct {
	ID                 string // row key: quest position in questdat.tsv (questdat.cpp:35-53 emplace_back per record)
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

// ParseQuestdat maps row i to questID(i) — the row key is position.
func ParseQuestdat(data []byte) ([]questRow, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, err
	}
	rows := tbl.Rows
	out := make([]questRow, 0, len(rows))
	for _, r := range rows {
		row := questRow{
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

// initialiseQuestPools ports: five pickRandomlyAmong
// pools (four when the seed hits the vanilla 988045466 bug path, which
// discards one draw instead). Each picked quest becomes QUEST_NOTAVAIL.
func initialiseQuestPools(seed uint32, setRndSeed func(uint32), rnd func(int32) int32) []questState {
	setRndSeed(seed)
	qs := make([]questState, QuestIDCount)
	for i := range qs {
		qs[i] = QuestInit
	}
	pick := func(ids ...questID) {
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
