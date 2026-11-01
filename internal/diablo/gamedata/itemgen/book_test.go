package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/testsupport"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

func TestGetBookSpellPicksNthEligible(t *testing.T) {
	tests := []struct {
		name  string
		env   itemgen.SpellEnv
		level int
		draw  int32
		want  int
	}{
		// level 1 walk: Firebolt(1), Healing(2), ChargedBolt(30), HolyBolt(31)
		{name: "first eligible", env: itemgen.SpellEnv{}, level: 1, draw: 0, want: 1},
		{name: "third eligible skips the level gates", env: itemgen.SpellEnv{}, level: 1, draw: 2, want: 30},
		{name: "spawn caps the level", env: itemgen.SpellEnv{Spawn: true}, level: 9, draw: 1, want: 2},
		{name: "multiplayer keeps HealOther", env: itemgen.SpellEnv{Multiplayer: true}, level: 1, draw: 4, want: 34},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			spells := realSpells(t)
			rnd := &tapeRnd{t: t, values: []int32{tc.draw}}

			// when
			book, err := itemgen.GetBookSpell(spells, tc.env, tc.level, rnd.rnd)

			// then
			assertBookSpell(t, book, err, rnd, len(spells), tc.want)
		})
	}
}

func TestGetBookSpellBookDetails(t *testing.T) {
	// given
	spells := realSpells(t)
	rnd := &tapeRnd{t: t, values: []int32{0}}

	// when
	book, err := itemgen.GetBookSpell(spells, itemgen.SpellEnv{}, 1, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetBookSpell: %v", err)
	}
	if book.Name != "Firebolt" || book.MinMagic != 15 || book.Value != 1000 || book.Cursor != itemgen.CursorBookRed {
		t.Fatalf("book = %+v, want Firebolt minInt 15 value 1000 cursor red", book)
	}
}

func TestGetBookSpellNoEligibleSpell(t *testing.T) {
	// given
	spells := []txtdata.SpellData{
		{Name: "Null"},
		{Name: "Firebolt", BookLevel: 99, BookCost10: 100},
	}
	rnd := &tapeRnd{t: t, values: []int32{1}}

	// when
	_, err := itemgen.GetBookSpell(spells, itemgen.SpellEnv{}, 1, rnd.rnd)

	// then
	if err == nil {
		t.Fatal("err = nil, want the no-eligible-spell error instead of upstream's spin")
	}
}

func TestGetBookSpellCursorPerMagicType(t *testing.T) {
	tests := []struct {
		name string
		draw int32
		want int
	}{
		{name: "fire", draw: 0, want: itemgen.CursorBookRed},
		{name: "lightning", draw: 2, want: itemgen.CursorBookBlue},
		{name: "magic", draw: 1, want: itemgen.CursorBookGrey},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			spells := realSpells(t)
			rnd := &tapeRnd{t: t, values: []int32{tc.draw}}

			// when
			book, err := itemgen.GetBookSpell(spells, itemgen.SpellEnv{}, 1, rnd.rnd)

			// then
			assertBookCursor(t, book, err, tc.want)
		})
	}
}

func assertBookSpell(t *testing.T, book itemgen.BookSpell, err error, rnd *tapeRnd, tableSize, want int) {
	t.Helper()

	if err != nil {
		t.Fatalf("GetBookSpell: %v", err)
	}
	if book.Spell != want {
		t.Fatalf("spell = %d, want %d", book.Spell, want)
	}
	if rnd.args[0] != int32(tableSize) {
		t.Fatalf("draw arg = %d, want SpellsData.size %d", rnd.args[0], tableSize)
	}
}

func assertBookCursor(t *testing.T, book itemgen.BookSpell, err error, want int) {
	t.Helper()

	if err != nil {
		t.Fatalf("GetBookSpell: %v", err)
	}
	if book.Cursor != want {
		t.Fatalf("cursor = %d, want %d", book.Cursor, want)
	}
}

func realSpells(t *testing.T) []txtdata.SpellData {
	t.Helper()

	table, err := enginetxtdata.Parse(testsupport.Corpus(t, "spells", "assets/txtdata/spells/spelldat.tsv"))
	if err != nil {
		t.Fatalf("enginetxtdata.Parse: %v", err)
	}
	spells, err := txtdata.LoadSpellDataWithNull(table)
	if err != nil {
		t.Fatalf("LoadSpellDataWithNull: %v", err)
	}

	return spells
}
