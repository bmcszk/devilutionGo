// Package leveldat_test ports the level-content parity tests: trigger piece
// lists (trigs.cpp:31-59), enums, setmap manifest, theme table.
package leveldat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat"
)

// Trigger piece-list cardinalities must equal trigs.cpp:31-59 exactly.
func TestTriggerPieceListCardinalities(t *testing.T) {
	want := map[string]int{
		"TownDownList":  10,
		"TownWarp1List": 12,
		"TownCryptList": 8,
		"TownHiveList":  4,
		"L1UpList":      11,
		"L1DownList":    9,
		"L2UpList":      2,
		"L2DownList":    4,
		"L2TWarpUpList": 2,
		"L3UpList":      14,
		"L3DownList":    8,
		"L3TWarpUpList": 14,
		"L4UpList":      3,
		"L4DownList":    5,
		"L4TWarpUpList": 3,
	}
	for name, n := range want {
		got := leveldat.PieceList(name)
		if len(got) != n {
			t.Errorf("%s has %d pieces, want %d", name, len(got), n)
		}
	}
}

// Spot values: TownDownList starts {715,714,718...}; L1UpList ends 139;
// L3TWarpUpList contains 181 and 557..559.
func TestTriggerPieceListValues(t *testing.T) {
	td := leveldat.PieceList("TownDownList")
	if td[0] != 715 || td[1] != 714 || td[2] != 718 {
		t.Fatalf("TownDownList head = %v", td[:3])
	}
	l1up := leveldat.PieceList("L1UpList")
	if l1up[0] != 126 || l1up[len(l1up)-1] != 139 {
		t.Fatalf("L1UpList = %v", l1up)
	}
	l3tw := leveldat.PieceList("L3TWarpUpList")
	if l3tw[0] != 181 || l3tw[len(l3tw)-1] != 559 {
		t.Fatalf("L3TWarpUpList = %v", l3tw)
	}
}

// Enum parity: DungeonType 7 + NONE=-1; SetLevelID 9 values incl. sentinels;
// ThemeID 17 + THEME_NONE=-1.
func TestEnumParity(t *testing.T) {
	if leveldat.DTLast != leveldat.DTCrypt || leveldat.DTNone != -1 {
		t.Fatalf("DungeonType sentinels wrong: last=%d none=%d", leveldat.DTLast, leveldat.DTNone)
	}
	if leveldat.SLFirstArena != leveldat.SLArenaChurch || leveldat.SLLast != leveldat.SLArenaCircleOfLife {
		t.Fatal("SetLevelID arena sentinels wrong")
	}
	if got := leveldat.SLArenaCircleOfLife - leveldat.SLFirstArena; got != 2 {
		t.Fatalf("arena count = %d, want 3 entries 0..2", got)
	}
	if leveldat.ThemeCount != 17 || leveldat.ThemeNone != -1 {
		t.Fatalf("ThemeID counts wrong: %d/%d", leveldat.ThemeCount, leveldat.ThemeNone)
	}
}

// Town trigger worked examples: (25,29) -> NEXT; warp tiles gated by
// IsWarpOpen with their _tlvl (Catacombs 49,21 -> 5; Caves 17,69 -> 9;
// Hell 41,80 -> 13; Nest 80,62 -> 17; Crypt 36,24 -> 21).
func TestTownTriggers(t *testing.T) {
	tg := leveldat.TownTriggers()
	found := false
	for _, tr := range tg {
		if tr.X == 25 && tr.Y == 29 {
			found = true
			if tr.Effect != leveldat.TrigNextLevel {
				t.Fatalf("(25,29) effect = %v, want NEXT", tr.Effect)
			}
		}
	}
	if !found {
		t.Fatal("town (25,29) trigger missing")
	}
	warps := map[[2]int]struct {
		lvl int
		dt  leveldat.DungeonType
	}{
		{49, 21}: {5, leveldat.DTCatacombs},
		{17, 69}: {9, leveldat.DTCaves},
		{41, 80}: {13, leveldat.DTHell},
		{80, 62}: {17, leveldat.DTNest},
		{36, 24}: {21, leveldat.DTCrypt},
	}
	got := map[[2]int]bool{}
	for _, tr := range tg {
		key := [2]int{tr.X, tr.Y}
		if w, ok := warps[key]; ok {
			got[key] = true
			if tr.TargetLevel != w.lvl {
				t.Errorf("warp %v target level = %d, want %d", key, tr.TargetLevel, w.lvl)
			}
			if tr.RequiresWarpOpen != w.dt {
				t.Errorf("warp %v gate = %v, want %v", key, tr.RequiresWarpOpen, w.dt)
			}
		}
	}
	if len(got) != len(warps) {
		t.Fatalf("found %d of %d warp triggers", len(got), len(warps))
	}
}

// L4 piece 369 -> NEXT only when Q_BETRAYER == QUEST_DONE (both branches).
func TestL4BetrayerGate(t *testing.T) {
	rules := leveldat.L4PieceRules()
	r, ok := rules[369]
	if !ok {
		t.Fatal("piece 369 rule missing")
	}
	if !r.RequiresQuestDone {
		t.Fatal("369 must require Q_BETRAYER done")
	}
	if r.Effect != leveldat.TrigNextLevel {
		t.Fatalf("369 effect = %v, want NEXT", r.Effect)
	}
	// contrast: 82 -> PREV unconditionally
	r82, ok := rules[82]
	if !ok || r82.RequiresQuestDone || r82.Effect != leveldat.TrigPrevLevel {
		t.Fatalf("82 rule = %+v, want unconditional PREV", r82)
	}
}

// Setmap manifest: entry + visual tiles per V11.4.
func TestSetmapManifest(t *testing.T) {
	m := leveldat.SetmapManifest()
	cases := []struct {
		id     leveldat.SetLevelID
		x, y   int
		vx, vy int
	}{
		{leveldat.SLSkelking, 83, 44, 0, 0},
		{leveldat.SLBonechamb, 70, 40, 0, 0},
		{leveldat.SLPoisonwater, 31, 83, 0, 0},
		{leveldat.SLVilebetrayer, 35, 36, 0, 0},
		{leveldat.SLArenaChurch, 29, 22, 28, 20},
		{leveldat.SLArenaHell, 34, 26, 33, 26},
		{leveldat.SLArenaCircleOfLife, 30, 26, 29, 26},
	}
	for _, c := range cases {
		s, ok := m[c.id]
		if !ok {
			t.Errorf("setmap %v missing", c.id)
			continue
		}
		if s.EntryX != c.x || s.EntryY != c.y {
			t.Errorf("%v entry = (%d,%d), want (%d,%d)", c.id, s.EntryX, s.EntryY, c.x, c.y)
		}
		if c.vx != 0 && (s.VisualX != c.vx || s.VisualY != c.vy) {
			t.Errorf("%v visual = (%d,%d), want (%d,%d)", c.id, s.VisualX, s.VisualY, c.vx, c.vy)
		}
	}
}
