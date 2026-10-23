// Package audio ports the sound domain: the effects catalog, the
// level→music map, and the movie manifest (V9.4, measured .smk sites).
package audio

import (
	"errors"
	"fmt"
	"strings"
)

// Effect is one effects.tsv row: id|flags|path.
type Effect struct {
	ID    string
	Flags []string
	Path  string
}

// effectsEqual reports byte-level equality of two rows (flags order included).
func effectsEqual(a, b Effect) bool {
	if a.ID != b.ID || a.Path != b.Path || len(a.Flags) != len(b.Flags) {
		return false
	}
	for i := range a.Flags {
		if a.Flags[i] != b.Flags[i] {
			return false
		}
	}
	return true
}

// ParseEffects parses an effects.tsv (id|flags|path) in row order.
func ParseEffects(data []byte) ([]Effect, error) {
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
			return nil, fmt.Errorf("audio: row has %d cols, want %d: %q", len(cols), len(header), ln)
		}
		rows = append(rows, cols)
	}
	if header == nil {
		return nil, errors.New("audio: empty table")
	}
	out := make([]Effect, 0, len(rows))
	for _, r := range rows {
		out = append(out, Effect{ID: r[0], Flags: strings.Split(r[1], ","), Path: r[2]})
	}
	return out, nil
}

// sfxIDCount pins SfxID to row order: the enum length equals the table
// length per variant (V9.1).
func sfxIDCount(rows int) int { return rows }

// DungeonType mirrors dungeon_type
// TOWN=0..CRYPT=6, NONE=-1.
type DungeonType int8

const (
	DTTown      DungeonType = 0
	DTCathedral DungeonType = 1
	DTCatacombs DungeonType = 2
	DTCaves     DungeonType = 3
	DTHell      DungeonType = 4
	DTNest      DungeonType = 5
	DTCrypt     DungeonType = 6
	DTLast      DungeonType = 6
	DTNone      DungeonType = -1
)

// musicID mirrors _music_id: the tracks getLevelMusic returns.
type musicID int8

const (
	MusicTown musicID = iota
	MusicCathedral
	MusicCatacombs
	MusicCaves
	MusicHell
	MusicNest
	MusicCrypt
	MusicIntro
)

// getLevelMusic ports engine/sound.cpp:310-330 verbatim; the default case
// (DT_NONE or any unmapped type) falls through to the intro track.
func getLevelMusic(dungeonType DungeonType) musicID {
	switch dungeonType {
	case DTTown:
		return MusicTown
	case DTCathedral:
		return MusicCathedral
	case DTCatacombs:
		return MusicCatacombs
	case DTCaves:
		return MusicCaves
	case DTHell:
		return MusicHell
	case DTNest:
		return MusicNest
	case DTCrypt:
		return MusicCrypt
	default:
		return MusicIntro
	}
}

// movie is one in-game movie with its upstream trigger site(s).
type movie struct {
	File     string
	Site     string // upstream play_movie / PlayInGameMovie call site(s)
	Closable bool   // user_can_close argument
}

// Movies is the measured manifest: every .smk referenced in the tree
// (V9.4) — 9 entries; the bean body's "exactly 10" did not survive
// measurement.
var movies = []movie{
	{"logo.smk", "diablo.cpp:1320", true},
	{"Hellfire.smk", "diablo.cpp:1326|menu.cpp:98", true},
	{"diablo1.smk", "diablo.cpp:1328|menu.cpp:100", true},
	{"fprst3.smk", "monster.cpp:2892", false},
	{"diabvic1.smk", "monster.cpp:4093", false},
	{"diabvic2.smk", "monster.cpp:4097", false},
	{"diabvic3.smk", "monster.cpp:4100", false},
	{"diabend.smk", "monster.cpp:4103", false},
	{"loopdend.smk", "monster.cpp:4113", true},
}

// Movies returns the movie manifest (do not mutate).
func Movies() []movie { return movies }

// victoryMovieForDifficulty returns the Diablo-victory movie for a
// 0-based difficulty.
func victoryMovieForDifficulty(difficulty int) string {
	if difficulty < 0 || difficulty > 2 {
		return "diabvic1.smk"
	}
	return []string{"diabvic1.smk", "diabvic2.smk", "diabvic3.smk"}[difficulty]
}
