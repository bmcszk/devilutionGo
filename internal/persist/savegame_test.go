package persist_test

import (
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/persist"
)

type saveGameState struct {
	name   string
	header persist.GameHeader
	want   string
}

func TestGameMagic(t *testing.T) {
	// when each game mode picks its magic (loadsave.cpp:2765-2776)
	cases := []struct {
		hellfire, spawn bool
		want            string
	}{
		{false, false, persist.MagicRetail},
		{true, false, persist.MagicHellfire},
		{false, true, persist.MagicShareware},
		{true, true, persist.MagicSharewareHellfire},
	}
	// then the tokens match the upstream ASCII values
	for _, tc := range cases {
		got, err := persist.GameMagic(persist.GameMode{Hellfire: tc.hellfire, Spawn: tc.spawn})
		if err != nil {
			t.Fatalf("GameMagic: %v", err)
		}

		if got != tc.want {
			t.Fatalf("GameMagic(%v,%v) = %q, want %q", tc.hellfire, tc.spawn, got, tc.want)
		}
	}
}

func TestLevelCountsFor(t *testing.T) {
	// when the level triple is selected per game (loadsave.cpp:2778-2787)
	hellfire := persist.LevelCountsFor(persist.GameMode{Hellfire: true})
	diablo := persist.LevelCountsFor(persist.GameMode{})
	// then Hellfire has 25 levels / 24 quests / 15 premium, Diablo 17/16/6
	if hellfire != (persist.LevelCounts{Levels: 25, Quests: 24, SmithPremium: 15}) {
		t.Fatalf("hellfire = %+v", hellfire)
	}

	if diablo != (persist.LevelCounts{Levels: 17, Quests: 16, SmithPremium: 6}) {
		t.Fatalf("diablo = %+v", diablo)
	}
}

func TestClampMissiles(t *testing.T) {
	// when the missile count is capped for vanilla compatibility
	// (loadsave.cpp:55, :2793-2795)
	cases := []struct{ in, want int }{
		{0, 0}, {7, 7}, {125, 125}, {126, 125}, {1000, 125},
	}
	// then anything above 125 is clamped
	for _, tc := range cases {
		if got := persist.ClampMissiles(tc.in); got != tc.want {
			t.Fatalf("ClampMissiles(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestGameHeaderGoldenImage(t *testing.T) {
	// given a Diablo game header on dungeon level 3
	header := persist.GameHeader{
		CurrLevel:      3,
		LevelType:      5,
		ViewX:          160,
		ViewY:          176,
		InventoryFlag:  true,
		ActiveMonsters: 7,
		ActiveItems:    2,
		ActiveMissiles: 200,
		ActiveObjects:  9,
	}
	// when
	encoded, err := header.Encode()
	// then the magic is little-endian ASCII, the counts big-endian, and the
	// missile count is clamped to 125
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	const want = "5245544c00" + // "RETL" + setlevel=0
		"00000000" + // setlvlnum BE
		"00000003" + // currlevel BE
		"00000005" + // leveltype BE
		"000000a0" + // viewx BE
		"000000b0" + // viewy BE
		"01" + "00" + // invflag, charflag
		"00000007" + // monsters BE
		"00000002" + // items BE
		"0000007d" + // missiles BE (125)
		"00000009" // objects BE
	if got := hex.EncodeToString(encoded); got != want {
		t.Fatalf("Encode = %s, want %s", got, want)
	}

	if len(encoded) != persist.GameHeaderSize {
		t.Fatalf("header size = %d, want %d", len(encoded), persist.GameHeaderSize)
	}
}

func TestGameHeaderRoundTrip(t *testing.T) {
	// given a Hellfire set-level header
	header := persist.GameHeader{
		Hellfire:       true,
		SetLevel:       true,
		SetLevelNum:    21,
		CurrLevel:      9,
		LevelType:      2,
		ViewX:          -4,
		ViewY:          8,
		CharFlag:       true,
		ActiveMissiles: 125,
	}
	// when
	encoded, err := header.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, magic, err := persist.DecodeGameHeader(encoded)
	// then the header round-trips and reports the Hellfire magic
	if err != nil {
		t.Fatalf("DecodeGameHeader: %v", err)
	}

	if magic != persist.MagicHellfire || decoded != header {
		t.Fatalf("decoded = %+v magic %q, want %+v", decoded, magic, header)
	}
}

func TestDecodeGameHeaderBadMagic(t *testing.T) {
	// given an image with an unknown magic token
	buf := make([]byte, persist.GameHeaderSize)
	copy(buf, "XXXX")
	// when
	_, magic, err := persist.DecodeGameHeader(buf)
	// then the decode reports the invalid game state
	if !errors.Is(err, persist.ErrInvalidGameState) {
		t.Fatalf("err = %v, want ErrInvalidGameState (magic %q)", err, magic)
	}
}

func TestSaveGameStatesGolden(t *testing.T) {
	// given the five scripted game states in testdata/savegame_states.tsv
	states := loadSaveGameStates(t)
	// when each header is encoded with the values the game state carries
	for _, state := range states {
		// then the stream matches the independently generated hex image
		encoded, err := state.header.Encode()
		if err != nil {
			t.Fatalf("%s: Encode: %v", state.name, err)
		}

		if got := hex.EncodeToString(encoded); got != state.want {
			t.Fatalf("%s: stream = %s, want %s", state.name, got, state.want)
		}
	}
}

func loadSaveGameStates(t *testing.T) []saveGameState {
	t.Helper()

	raw, err := os.ReadFile("testdata/savegame_states.tsv")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var states []saveGameState
	for line := range strings.SplitSeq(strings.TrimSpace(string(raw)), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		states = append(states, parseSaveGameState(t, line))
	}

	if len(states) != 5 {
		t.Fatalf("fixture rows = %d, want 5", len(states))
	}

	return states
}

func parseSaveGameState(t *testing.T, line string) saveGameState {
	t.Helper()

	fields := strings.Split(line, "\t")
	if len(fields) != 15 {
		t.Fatalf("fixture row %q has %d fields, want 15", fields[0], len(fields))
	}

	numbers := make([]int, 0, 13)
	for _, field := range fields[2:14] {
		value, err := strconv.Atoi(field)
		if err != nil {
			t.Fatalf("fixture row %q: %v", fields[0], err)
		}

		numbers = append(numbers, value)
	}

	header := persist.GameHeader{
		Hellfire:       fields[1] == persist.MagicHellfire || fields[1] == persist.MagicSharewareHellfire,
		Spawn:          fields[1] == persist.MagicShareware || fields[1] == persist.MagicSharewareHellfire,
		SetLevel:       numbers[0] != 0,
		SetLevelNum:    uint32(numbers[1]),
		CurrLevel:      uint32(numbers[2]),
		LevelType:      uint32(numbers[3]),
		ViewX:          int32(numbers[4]),
		ViewY:          int32(numbers[5]),
		InventoryFlag:  numbers[6] != 0,
		CharFlag:       numbers[7] != 0,
		ActiveMonsters: int32(numbers[8]),
		ActiveItems:    int32(numbers[9]),
		ActiveMissiles: int32(numbers[10]),
		ActiveObjects:  int32(numbers[11]),
	}

	return saveGameState{name: fields[0], header: header, want: fields[14]}
}
