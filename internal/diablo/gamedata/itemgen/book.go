package itemgen

import (
	"errors"
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// Book cursor ids (tables/itemdat.h:176-178).
const (
	CursorBookGrey = 86
	CursorBookRed  = 87
	CursorBookBlue = 88
)

// spawnCappedBookLevel is the gbIsSpawn GetBookSpell level cap
// (items.cpp:2497-2499).
const spawnCappedBookLevel = 5

// spawnExcludedBookSpells are the GetSpellBookLevel -1 spells under gbIsSpawn
// (spells.cpp:352-363). Books exclude Apocalypse, staves do not.
var spawnExcludedBookSpells = [...]string{
	"StoneCurse", "Guardian", "Golem", "Elemental", "BloodStar", "BoneSpirit",
}

// BookSpell is the GetBookSpell outcome (items.cpp:2489-2545).
type BookSpell struct {
	Spell    int
	Name     string
	MinMagic int
	Value    int // bookCost() added to _ivalue (:2537-2538)
	Cursor   int
}

// GetBookSpell ports GetBookSpell (items.cpp:2489-2545): pick the rv-th
// spell whose book level the item level meets, then append the spell name
// and adjust the magic requirement, value and cursor.
func GetBookSpell(
	spells []txtdata.SpellData,
	env SpellEnv,
	lvl int,
	rnd func(int32) int32,
) (BookSpell, error) {
	if len(spells) == 0 {
		return BookSpell{}, errors.New("empty spell table")
	}
	if lvl == 0 {
		lvl = 1
	}
	if env.Spawn && lvl > spawnCappedBookLevel {
		lvl = spawnCappedBookLevel
	}

	rv := int(rnd(int32(len(spells)))) + 1
	spell, found := nthBookSpell(spells, env, lvl, rv)
	if !found {
		return BookSpell{}, fmt.Errorf("no book spell for level %d roll %d", lvl, rv)
	}

	data := spells[spell]

	return BookSpell{
		Spell:    spell,
		Name:     data.Name,
		MinMagic: data.MinInt,
		Value:    data.BookCost10 * 10,
		Cursor:   bookCursor(data.Flags),
	}, nil
}

// nthBookSpell walks spells from Firebolt, applying the non-multiplayer
// Resurrect/HealOther shortcuts and the table wrap (:2505-2530).
func nthBookSpell(spells []txtdata.SpellData, env SpellEnv, lvl, rv int) (int, bool) {
	s := spellFirebolt
	book := spellFirebolt
	for rv > 0 {
		if level := bookLevel(spells, env, s); level != -1 && lvl >= level {
			rv--
			book = s
		}
		s = nextBookSpell(spells, env, s)
		if s == spellFirebolt && rv > 0 {
			// A full lap without satisfying the roll would loop forever;
			// upstream keeps spinning until the level gate matches.
			return 0, false
		}
	}

	return book, true
}

// nextBookSpell advances s with the two non-multiplayer skips and the wrap
// (:2515-2529).
func nextBookSpell(spells []txtdata.SpellData, env SpellEnv, s int) int {
	s++
	if !env.Multiplayer {
		if s == spellResurrect {
			s = spellTelekinesis
		}
		if s == spellHealOther {
			s = spellBloodStar
		}
	}
	if s >= len(spells) {
		s = 1
	}

	return s
}

// bookLevel is GetSpellBookLevel (spells.cpp:346-364).
func bookLevel(spells []txtdata.SpellData, env SpellEnv, spell int) int {
	if spell < 0 || spell >= len(spells) {
		return -1
	}
	if !env.Spawn {
		return spells[spell].BookLevel
	}
	for _, excluded := range spawnExcludedBookSpells {
		if spells[spell].Name == excluded {
			return -1
		}
	}

	return spells[spell].BookLevel
}

// bookCursor maps MagicType to the book cursor (:2539-2544).
func bookCursor(flags []string) int {
	for _, flag := range flags {
		switch flag {
		case "Lightning":
			return CursorBookBlue
		case "Magic":
			return CursorBookGrey
		}
	}

	return CursorBookRed
}
