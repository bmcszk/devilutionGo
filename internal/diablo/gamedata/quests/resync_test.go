package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// ResyncMPQuests activates the multiplayer quests as the player approaches
// their level: the skeleton king and the butcher within one level either way,
// the betrayer, crypt, defiler and Na-Krul one level above
// (quests.cpp:349-397).
func TestResyncMPQuestsActivatesByLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     int
		wantSent  []int8
		wantSpawn int
	}{
		{name: "level1", level: 1, wantSent: []int8{6}},
		{name: "level2", level: 2, wantSent: []int8{12, 6}},
		{name: "level3", level: 3, wantSent: []int8{12, 6}},
		{name: "level4", level: 4, wantSent: []int8{12}},
		{name: "level5", level: 5, wantSent: nil},
		{name: "level9", level: 9, wantSent: nil},
		{name: "level13", level: 13, wantSent: nil},
		{name: "level14", level: 14, wantSent: []int8{15}},
		{name: "level16", level: 16, wantSent: []int8{16, 20}},
		{name: "level20", level: 20, wantSent: []int8{21}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			rows := mustRows(t)
			qs := mustInit(t, enums.SimFlags{Multiplayer: true})
			rec := &killRecorder{}
			// when
			quests.ResyncMPQuests(qs, rows, quests.ResyncContext{
				Level: tt.level,
				Flags: enums.SimFlags{Multiplayer: true},
			}, resyncEffects(rec))
			// then
			assertResyncSent(t, rec.sent, tt.wantSent)
			assertResyncStates(t, qs, tt.wantSent)
		})
	}
}

// The AltBoy object spawns when the betrayer quest is available, which needs
// the player on the quest level itself (questdat.cpp:64-74,
// quests.cpp:376-377).
func TestResyncMPQuestsAltBoy(t *testing.T) {
	tests := []struct {
		name      string
		level     int
		wantSpawn int
	}{
		{name: "oneAbove", level: 14, wantSpawn: 0},
		{name: "onQuestLevel", level: 15, wantSpawn: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			rows := mustRows(t)
			qs := mustInit(t, enums.SimFlags{Multiplayer: true})
			rec := &killRecorder{}
			// when
			quests.ResyncMPQuests(qs, rows, quests.ResyncContext{
				Level: tt.level,
				Flags: enums.SimFlags{Multiplayer: true},
			}, resyncEffects(rec))
			// then
			if rec.altBoy != tt.wantSpawn {
				t.Fatalf("AltBoy spawns = %d, want %d", rec.altBoy, tt.wantSpawn)
			}
		})
	}
}

// The shareware build resyncs nothing (quests.cpp:351-352).
func TestResyncMPQuestsSpawnGate(t *testing.T) {
	// given
	rows := mustRows(t)
	qs := mustInit(t, enums.SimFlags{Spawn: true})
	rec := &killRecorder{}
	// when
	quests.ResyncMPQuests(qs, rows, quests.ResyncContext{
		Level: 2,
		Flags: enums.SimFlags{Spawn: true},
	}, resyncEffects(rec))
	// then
	if len(rec.sent) != 0 || rec.altBoy != 0 {
		t.Fatalf("spawn resync fired: sent %v altBoy %d", rec.sent, rec.altBoy)
	}
}

// The veil quest spawns the golden elixir once, one level above its own level,
// in single player only (quests.cpp:464-468).
func TestResyncQuestsGoldenElixir(t *testing.T) {
	tests := []struct {
		name   string
		level  int
		flags  enums.SimFlags
		state  quests.QuestState
		var1   uint8
		expect int
	}{
		{name: "oneLevelAbove", level: 15, state: quests.QuestActive, expect: 1},
		{name: "questLevel", level: 14, state: quests.QuestActive},
		{name: "twoLevelsAbove", level: 16, state: quests.QuestActive},
		{name: "questNotActive", level: 15, state: quests.QuestInit},
		{name: "elixirAlreadyGiven", level: 15, state: quests.QuestActive, var1: 1},
		{name: "multiplayer", level: 15, state: quests.QuestActive, flags: enums.SimFlags{Multiplayer: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInit(t, tt.flags)
			qs[4].Active = tt.state
			qs[4].Var1 = tt.var1
			rec := &killRecorder{}
			// when
			quests.ResyncQuests(qs, quests.ResyncContext{Level: tt.level, Flags: tt.flags},
				quests.ResyncSPEffects{SpawnGoldenElixir: rec.spawnElixir, SendQuest: rec.sendQuest})
			// then
			assertElixir(t, rec.elixir, tt.expect)
			assertVeilVar1(t, qs[4].Var1, tt.var1, tt.expect)
		})
	}
}

// assertElixir pins the spawn count.
func assertElixir(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("elixir spawns = %d, want %d", got, want)
	}
}

// assertVeilVar1 pins the quest variable: the spawn marks QS_ELIXIRGIVEN, and
// an untouched quest keeps its value.
func assertVeilVar1(t *testing.T, got, before uint8, spawns int) {
	t.Helper()

	want := before
	if spawns == 1 {
		want = 1
	}

	if got != want {
		t.Fatalf("var1 = %d, want %d", got, want)
	}
}
