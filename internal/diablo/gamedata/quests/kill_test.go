package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// The killing blow on a quest monster closes its quest and the hero comments
// (quests.cpp:227-289).
func TestCheckQuestKillMonsterTypes(t *testing.T) {
	tests := []struct {
		name     string
		victim   quests.KillVictim
		slot     int
		speech   string
		sendMsgs bool
		because  string
	}{
		{
			name:     "skeletonKing",
			victim:   quests.KillVictim{MonsterID: "MT_SKING"},
			slot:     12,
			speech:   quests.SpeechRestWellLeoric,
			sendMsgs: true,
			because:  "MT_SKING closes Q_SKELKING",
		},
		{
			name:     "butcher",
			victim:   quests.KillVictim{MonsterID: "MT_CLEAVER"},
			slot:     6,
			speech:   quests.SpeechSpiritsAvenged,
			sendMsgs: true,
			because:  "MT_CLEAVER closes Q_BUTCHER",
		},
		{
			name:    "gharbad",
			victim:  quests.KillVictim{MonsterID: "MT_NGOATMC", UniqueType: "Garbud"},
			slot:    2,
			speech:  quests.SpeechNotImpressed,
			because: "the Garbud unique closes Q_GARBUD and always broadcasts",
		},
		{
			name:    "zhar",
			victim:  quests.KillVictim{MonsterID: "MT_COUNSLR", UniqueType: "Zhar"},
			slot:    3,
			speech:  quests.SpeechSorryConcentration,
			because: "the Zhar unique closes Q_ZHAR and always broadcasts",
		},
		{
			name:    "warlord",
			victim:  quests.KillVictim{MonsterID: "MT_BTBLACK", UniqueType: "WarlordOfBlood"},
			slot:    11,
			speech:  quests.SpeechReignOfPainEnded,
			because: "the Warlord unique closes Q_WARLORD and always broadcasts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInit(t, enums.SimFlags{Multiplayer: tt.sendMsgs})
			rec := &killRecorder{}
			// when
			quests.CheckQuestKill(qs, quests.KillContext{Victim: tt.victim, SendMsgs: tt.sendMsgs}, rec.hooks())
			// then
			if qs[tt.slot].Active != quests.QuestDone {
				t.Fatalf("quest %d = %d, want QUEST_DONE: %s", tt.slot, qs[tt.slot].Active, tt.because)
			}

			assertKillEffects(t, rec, tt.speech, tt.slot)
		})
	}
}

// The shareware build ignores kills entirely (quests.cpp:227-229).
func TestCheckQuestKillSpawnGate(t *testing.T) {
	// given
	qs := mustInit(t, enums.SimFlags{Spawn: true})
	rec := &killRecorder{}
	// when
	quests.CheckQuestKill(qs, quests.KillContext{
		Victim:   quests.KillVictim{MonsterID: "MT_SKING"},
		Flags:    enums.SimFlags{Spawn: true},
		SendMsgs: true,
	}, rec.hooks())
	// then
	if qs[12].Active != quests.QuestNotAvail {
		t.Fatalf("Q_SKELKING = %d, want QUEST_NOTAVAIL", qs[12].Active)
	}

	if len(rec.speech) != 0 || len(rec.sent) != 0 {
		t.Fatalf("spawn kill fired effects: speech %v sent %v", rec.speech, rec.sent)
	}
}

// sendmsg = false keeps the quest state change but drops the packet
// (quests.cpp:236).
func TestCheckQuestKillWithoutSendMsg(t *testing.T) {
	// given
	qs := mustInit(t, enums.SimFlags{Multiplayer: true})
	rec := &killRecorder{}
	// when
	quests.CheckQuestKill(qs, quests.KillContext{
		Victim: quests.KillVictim{MonsterID: "MT_CLEAVER"},
		Flags:  enums.SimFlags{Multiplayer: true},
	}, rec.hooks())
	// then
	if qs[6].Active != quests.QuestDone {
		t.Fatalf("Q_BUTCHER = %d, want QUEST_DONE", qs[6].Active)
	}

	if len(rec.sent) != 0 {
		t.Fatalf("sent = %v, want no packet without sendmsg", rec.sent)
	}
}

// Lazarus closes Q_BETRAYER, sets var1 = 7, opens Q_DIABLO and opens the way
// down: a red portal in single player, the tile-369 trigger scan in
// multiplayer (quests.cpp:251-289).
func TestCheckQuestKillLazarus(t *testing.T) {
	tests := []struct {
		name       string
		flags      enums.SimFlags
		wantPortal int
		wantTrig   int
		wantVar2   uint8
	}{
		{
			name:       "singlePlayer",
			flags:      enums.SimFlags{},
			wantPortal: 1,
			wantVar2:   4,
		},
		{
			name:     "multiplayer",
			flags:    enums.SimFlags{Multiplayer: true},
			wantTrig: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInit(t, tt.flags)
			rec := &killRecorder{}
			// when
			quests.CheckQuestKill(qs, quests.KillContext{
				Victim:   quests.KillVictim{MonsterID: "MT_ADVOCATE", UniqueType: "Lazarus"},
				Flags:    tt.flags,
				SendMsgs: true,
			}, rec.hooks())
			// then
			assertLazarus(t, qs, tt.wantVar2)
			assertLazarusEffects(t, rec, tt.wantPortal, tt.wantTrig)
		})
	}
}

// An ordinary monster closes nothing (quests.cpp:227-289 has no default arm).
func TestCheckQuestKillIgnoresOrdinaryMonster(t *testing.T) {
	// given
	qs := mustInit(t, enums.SimFlags{})
	rec := &killRecorder{}
	// when
	quests.CheckQuestKill(qs, quests.KillContext{
		Victim:   quests.KillVictim{MonsterID: "MT_NZOMBIE"},
		SendMsgs: true,
	}, rec.hooks())
	// then
	for i, q := range qs {
		if q.Active != quests.QuestInit {
			t.Fatalf("quest %d = %d, want untouched QUEST_INIT", i, q.Active)
		}
	}

	if len(rec.speech) != 0 || len(rec.sent) != 0 {
		t.Fatalf("ordinary kill fired effects: speech %v sent %v", rec.speech, rec.sent)
	}
}
