package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// UniqueMonsterType indices are the array positions UniqueMonstersData uses
// (monster.h:60-75), so the order is the 1:1 contract.
func TestUniqueMonsterTypeOrder(t *testing.T) {
	// given
	want := []struct {
		index uint8
		name  string
	}{
		{index: 0, name: "Garbud"},
		{index: 1, name: "SkeletonKing"},
		{index: 2, name: "Zhar"},
		{index: 3, name: "SnotSpill"},
		{index: 4, name: "Lazarus"},
		{index: 5, name: "RedVex"},
		{index: 6, name: "BlackJade"},
		{index: 7, name: "Lachdan"},
		{index: 8, name: "WarlordOfBlood"},
		{index: 9, name: "Butcher"},
		{index: 10, name: "HorkDemon"},
		{index: 11, name: "Defiler"},
		{index: 12, name: "NaKrul"},
		{index: 255, name: ""},
	}
	// when
	gotNames := make([]string, 0, len(want))
	for _, w := range want {
		gotNames = append(gotNames, quests.UniqueMonsterName(w.index))
	}
	// then
	for i, w := range want {
		if got := gotNames[i]; got != w.name {
			t.Fatalf("unique type %d name = %q, want %q", w.index, got, w.name)
		}
	}
}

// The kill hook reads a monster through its table id and its UniqueType index
// (monster.h:261), which is what entities.Monster carries.
func TestVictimFromMonster(t *testing.T) {
	// given
	monster := struct {
		id     string
		unique uint8
	}{id: "MT_ADVOCATE", unique: 4}
	// when
	got := quests.VictimFromMonster(monster.id, monster.unique)
	// then
	if got.MonsterID != "MT_ADVOCATE" || got.UniqueType != quests.UniqueLazarus {
		t.Fatalf("victim = %q/%q, want MT_ADVOCATE/Lazarus", got.MonsterID, got.UniqueType)
	}
}

// An ordinary monster has UniqueMonsterType::None (monster.h:74) and gets the
// empty unique name, so no unique branch fires.
func TestVictimFromOrdinaryMonster(t *testing.T) {
	// when
	got := quests.VictimFromMonster("MT_NZOMBIE", quests.UniqueTypeNone)
	// then
	if got.UniqueType != "" {
		t.Fatalf("unique type = %q, want empty", got.UniqueType)
	}
}
