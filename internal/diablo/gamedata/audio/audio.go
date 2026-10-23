// Package audio ports the sound domain: the effects catalog, the
// level→music map, and the movie manifest (V9.4, measured .smk sites).
package audio

import (
	"strings"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
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
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, err
	}
	rows := tbl.Rows
	out := make([]Effect, 0, len(rows))
	for _, r := range rows {
		out = append(out, Effect{ID: r[0], Flags: strings.Split(r[1], ","), Path: r[2]})
	}
	return out, nil
}

// DungeonType mirrors dungeon_type; canonical values live in leveldat
// (TOWN=0..CRYPT=6, NONE=-1).
type DungeonType = leveldat.DungeonType

const (
	DTTown      = leveldat.DTTown
	DTCathedral = leveldat.DTCathedral
	DTCatacombs = leveldat.DTCatacombs
	DTCaves     = leveldat.DTCaves
	DTHell      = leveldat.DTHell
	DTNest      = leveldat.DTNest
	DTCrypt     = leveldat.DTCrypt
	DTLast      = leveldat.DTLast
	DTNone      = leveldat.DTNone
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

// moviesManifest is the measured manifest: every .smk referenced in the tree
// (V9.4) — 9 entries; the bean body's "exactly 10" did not survive
// measurement.
var moviesManifest = []movie{
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

// movies returns the movie manifest (do not mutate).
func movies() []movie { return moviesManifest }

// victoryMovieForDifficulty returns the Diablo-victory movie for a
// 0-based difficulty.
func victoryMovieForDifficulty(difficulty int) string {
	if difficulty < 0 || difficulty > 2 {
		return "diabvic1.smk"
	}
	return []string{"diabvic1.smk", "diabvic2.smk", "diabvic3.smk"}[difficulty]
}
