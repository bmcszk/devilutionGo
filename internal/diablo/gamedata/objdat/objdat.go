// Package objdat loads the object data table (objdat.tsv) and holds the
// object-behaviour enumerations: shrine types and shrine availability.
package objdat

import (
	"errors"
	"fmt"
	"strings"
)

// ThemeCount is the theme count of game version V11.3, excluding THEME_NONE.
const ThemeCount = 17

// ShrineCount is the full Hellfire shrine pool; the sentinel is excluded.
const ShrineCount = 34

// DiabloShrinePool is the Diablo-era shrine pool size.
const DiabloShrinePool = 26

// SelectionRegion spellings as they appear in the objdat TSV column.
const (
	RegionNone   = ""
	RegionBottom = "Bottom"
	RegionMiddle = "Middle"
	RegionTop    = "Top"
)

// objectData is one objdat.tsv row (12 columns).
type objectData struct {
	ID              string
	File            string
	MinLevel        string
	MaxLevel        string
	LevelType       string
	Theme           string
	Quest           string
	Flags           []string
	AnimDelay       int
	AnimLen         int
	AnimWidth       int
	SelectionRegion string
	Animated        bool
	Solid           bool
}

// ParseObjdat parses objdat.tsv in row order; upstream ships no Hellfire
// override.
func ParseObjdat(data []byte) ([]objectData, error) {
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
			return nil, fmt.Errorf("objdat: row has %d cols, want %d: %q", len(cols), len(header), ln)
		}
		rows = append(rows, cols)
	}
	if header == nil {
		return nil, errors.New("objdat: empty table")
	}
	out := make([]objectData, 0, len(rows))
	for _, r := range rows {
		o := objectData{
			ID:              r[0],
			File:            r[1],
			MinLevel:        r[2],
			MaxLevel:        r[3],
			LevelType:       r[4],
			Theme:           r[5],
			Quest:           r[6],
			Flags:           strings.Split(r[7], ","),
			SelectionRegion: r[11],
		}
		for _, f := range o.Flags {
			switch f {
			case "Animated":
				o.Animated = true
			case "Solid":
				o.Solid = true
			default:
				// MissilesPassThrough / Light / Trap / Breakable decode at
				// their consumer sites; Flags stays raw.
			}
		}
		// animDelay/animLen/animWidth are 0 for static objects.
		o.AnimDelay = atoiOrZero(r[8])
		o.AnimLen = atoiOrZero(r[9])
		o.AnimWidth = atoiOrZero(r[10])
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

// shrineGametype is the per-shrine gametype gate: Single shrines spawn in
// single player only, Multi shrines are excluded from single player.
type shrineGametype string

const (
	ShrineTypeAny    shrineGametype = "Any"
	ShrineTypeSingle shrineGametype = "Single"
	ShrineTypeMulti  shrineGametype = "Multi"
)

// ShrineAvailability lists the per-shrine gate in shrine type order; it holds
// ShrineCount entries.
var ShrineAvailability = []shrineGametype{
	ShrineTypeAny,    // Mysterious
	ShrineTypeAny,    // Hidden
	ShrineTypeSingle, // Gloomy
	ShrineTypeSingle, // Weird
	ShrineTypeAny,    // Magical
	ShrineTypeAny,    // Stone
	ShrineTypeAny,    // Religious
	ShrineTypeAny,    // Enchanted
	ShrineTypeSingle, // Thaumaturgic
	ShrineTypeAny,    // Fascinating
	ShrineTypeAny,    // Cryptic
	ShrineTypeAny,    // Magical (L2)
	ShrineTypeAny,    // Eldritch
	ShrineTypeAny,    // Eerie
	ShrineTypeAny,    // Divine
	ShrineTypeAny,    // Holy
	ShrineTypeAny,    // Sacred
	ShrineTypeAny,    // Spiritual
	ShrineTypeMulti,  // Spooky
	ShrineTypeAny,    // Abandoned
	ShrineTypeAny,    // Creepy
	ShrineTypeAny,    // Quiet
	ShrineTypeAny,    // Secluded
	ShrineTypeAny,    // Ornate
	ShrineTypeAny,    // Glimmering
	ShrineTypeMulti,  // Tainted
	ShrineTypeAny,    // Oily
	ShrineTypeAny,    // Glowing
	ShrineTypeAny,    // Mendicant's
	ShrineTypeAny,    // Sparkling
	ShrineTypeAny,    // Town
	ShrineTypeAny,    // Shimmering
	ShrineTypeSingle, // Solar
	ShrineTypeAny,    // Murphy's
}

// SinglePlayerOnly lists the ShrineTypeSingle shrine names.
var SinglePlayerOnly = []string{"Gloomy", "Weird", "Thaumaturgic", "Solar"}

// MultiPlayerExcluded lists the ShrineTypeMulti shrine names.
var MultiPlayerExcluded = []string{"Spooky", "Tainted"}
