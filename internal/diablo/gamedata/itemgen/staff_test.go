package itemgen_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// Spell ids from the pinned SpellID enum order (tables/spelldat.h:30-69).
const (
	spellFireboltID    = 1
	spellResurrectID   = 32
	spellTelekinesisID = 33
	spellHealOtherID   = 34
	spellBloodStarID   = 35

	staffSpellTableSize = 40
)

func TestGetStaffSpellVanillaFallback(t *testing.T) {
	// given
	spells := staffSpells(map[int]int{spellFireboltID: 1}, nil)
	rnd := &tapeRnd{t: t, values: []int32{0}} // FlipCoin(4) hits

	// when
	roll, err := itemgen.GetStaffSpell(spells, 10, itemgen.SpellEnv{}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetStaffSpell: %v", err)
	}
	if !roll.AffixFallback {
		t.Fatal("AffixFallback = false, want the vanilla FlipCoin(4) branch")
	}
	if roll.MinLevel != 5 || roll.MaxLevel != 10 {
		t.Fatalf("levels = %d..%d, want 5..10 (GetItemPower lvl/2, lvl)", roll.MinLevel, roll.MaxLevel)
	}
	assertDraws(t, rnd.args, []int32{4})
}

func TestGetStaffSpellHellfireUsesSpellPath(t *testing.T) {
	// given
	spells := staffSpells(map[int]int{spellFireboltID: 1}, nil)
	rnd := &tapeRnd{t: t, values: []int32{1, 0}}

	// when
	roll, err := itemgen.GetStaffSpell(spells, 10, itemgen.SpellEnv{Hellfire: true}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetStaffSpell: %v", err)
	}
	if roll.AffixFallback {
		t.Fatal("AffixFallback = true, want the spell path under Hellfire")
	}
	if roll.Spell != spellFireboltID || roll.Charges != 40 || roll.MinMag != 15 || roll.Value != 400 {
		t.Fatalf("roll = %+v, want firebolt/40 charges/minmag 15/value 400", roll)
	}
	assertDraws(t, rnd.args, []int32{40, 41})
}

func TestGetStaffSpellSinglePlayerSkips(t *testing.T) {
	// given
	levels := map[int]int{spellResurrectID: 1, spellTelekinesisID: 1, spellHealOtherID: 1, spellBloodStarID: 1}
	single := staffSpells(levels, nil)
	multi := staffSpells(levels, nil)
	oneDraw := func() *tapeRnd { return &tapeRnd{t: t, values: []int32{0, 0}} }

	// when
	singleRoll, err := itemgen.GetStaffSpell(single, 2, itemgen.SpellEnv{Hellfire: true}, oneDraw().rnd)
	if err != nil {
		t.Fatalf("GetStaffSpell single: %v", err)
	}
	multiRoll, err := itemgen.GetStaffSpell(multi, 2, itemgen.SpellEnv{Hellfire: true, Multiplayer: true}, oneDraw().rnd)
	if err != nil {
		t.Fatalf("GetStaffSpell multi: %v", err)
	}

	// then
	if singleRoll.Spell != spellTelekinesisID {
		t.Fatalf("single-player spell = %d, want Telekinesis (Resurrect skipped)", singleRoll.Spell)
	}
	if multiRoll.Spell != spellResurrectID {
		t.Fatalf("multiplayer spell = %d, want Resurrect", multiRoll.Spell)
	}
}

func TestGetStaffSpellSkipsHighLevelSpells(t *testing.T) {
	// given
	spells := staffSpells(map[int]int{spellFireboltID: 5, spellFireboltID + 1: 1}, nil)
	rnd := &tapeRnd{t: t, values: []int32{0, 0}}

	// when
	roll, err := itemgen.GetStaffSpell(spells, 2, itemgen.SpellEnv{Hellfire: true}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetStaffSpell: %v", err)
	}
	if roll.Spell != spellFireboltID+1 {
		t.Fatalf("spell = %d, want %d (staff level 5 is above l=1)", roll.Spell, spellFireboltID+1)
	}
}

func TestGetSpellStaffLevelSpawnExclusions(t *testing.T) {
	// given
	spells := staffSpells(map[int]int{spellBloodStarID: 3}, map[int]string{spellBloodStarID: "BloodStar"})

	// when
	spawnLevel := itemgen.GetSpellStaffLevel(spells, spellBloodStarID, itemgen.SpellEnv{Spawn: true})
	normalLevel := itemgen.GetSpellStaffLevel(spells, spellBloodStarID, itemgen.SpellEnv{})
	offTable := itemgen.GetSpellStaffLevel(spells, 99, itemgen.SpellEnv{})

	// then
	if spawnLevel != -1 {
		t.Fatalf("spawn staff level = %d, want -1 for BloodStar (spells.cpp:334)", spawnLevel)
	}
	if normalLevel != 3 {
		t.Fatalf("staff level = %d, want 3", normalLevel)
	}
	if offTable != -1 {
		t.Fatalf("off-table staff level = %d, want -1", offTable)
	}
}

func TestGetStaffSpellSpawnCapsPowerLevel(t *testing.T) {
	// given
	spells := staffSpells(map[int]int{spellFireboltID: 1}, nil)
	rnd := &tapeRnd{t: t, values: []int32{0, 0}}

	// when
	roll, err := itemgen.GetStaffSpell(spells, 15, itemgen.SpellEnv{Hellfire: true, Spawn: true}, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetStaffSpell: %v", err)
	}
	if roll.MaxLevel != 10 {
		t.Fatalf("MaxLevel = %d, want the spawn cap 10 (items.cpp:1248)", roll.MaxLevel)
	}
}

func TestGetStaffPrefixCoin(t *testing.T) {
	// given
	prefixes := staffPrefixAffixes()
	missed := &tapeRnd{t: t, values: []int32{1}}    // FlipCoin(10) fails
	forced := &tapeRnd{t: t, values: []int32{1, 0}} // onlygood skips the coin
	hit := &tapeRnd{t: t, values: []int32{0, 0}}    // FlipCoin(10) hits

	// when
	_, missedOK := itemgen.GetStaffPrefix(prefixes, 10, false, missed.rnd)
	prefix, forcedOK := itemgen.GetStaffPrefix(prefixes, 10, true, forced.rnd)
	_, hitOK := itemgen.GetStaffPrefix(prefixes, 10, false, hit.rnd)

	// then
	if missedOK {
		t.Fatal("GetStaffPrefix ok = true, want false (FlipCoin(10) failed)")
	}
	if !forcedOK || prefix.Name != "of Fire" {
		t.Fatalf("onlygood prefix = %+v ok=%v, want of Fire", prefix, forcedOK)
	}
	if !hitOK {
		t.Fatal("GetStaffPrefix ok = false, want a prefix on a coin hit")
	}
}

func TestGetStaffPowerSavesPrefix(t *testing.T) {
	// given
	prefixes := staffPrefixAffixes()
	rnd := &tapeRnd{t: t, values: []int32{0, 0, 0}}

	// when
	power, err := itemgen.GetStaffPower(prefixes, 10, false, rnd.rnd)
	// then
	if err != nil {
		t.Fatalf("GetStaffPower: %v", err)
	}
	if !power.HasPrefix || power.Prefix.Name != "of Fire" {
		t.Fatalf("power = %+v, want the of Fire prefix", power)
	}
	if power.Bonus.Value != 1 || power.Values.Add1 != 100 || power.Values.Mult1 != 2 {
		t.Fatalf("values = %+v bonus = %+v, want Add1 100 / Mult1 2", power.Values, power.Bonus)
	}
	assertDraws(t, rnd.args, []int32{10, 1, 5})
}

func staffSpells(levels map[int]int, names map[int]string) []txtdata.SpellData {
	spells := make([]txtdata.SpellData, staffSpellTableSize)
	for i := range spells {
		spells[i] = txtdata.SpellData{StaffLevel: -1, StaffMin: 40, StaffMax: 80, StaffCost10: 5, MinInt: 15}
	}
	for id, level := range levels {
		spells[id].StaffLevel = level
	}
	for id, name := range names {
		spells[id].Name = name
	}

	return spells
}

func staffPrefixAffixes() []itemgen.Affix {
	return []itemgen.Affix{{
		Name: "of Fire", Power: "FIREDAM", Value1: 1, Value2: 5, MinLevel: 1,
		Types: itemgen.AffixStaff, Alignment: itemgen.AlignAny, Chance: 1, Useful: true,
		MinVal: 100, MaxVal: 500, MultVal: 2,
	}}
}
