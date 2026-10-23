// Package leveldat_test covers the level-content tables.
//
//nolint:testpackage // internal-state test: pokes unexported identifiers
package leveldat

import (
	"testing"
)

func TestTriggerPieceListCardinalities(t *testing.T) {
	// given
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
		t.Run(name, func(t *testing.T) {
			// when
			got := pieceList(name)
			// then
			if len(got) != n {
				t.Errorf("%s has %d pieces, want %d", name, len(got), n)
			}
		})
	}
}

func TestTriggerPieceListValues(t *testing.T) {
	cases := []struct {
		name string
		head []uint16
		tail uint16
	}{
		{"TownDownList", []uint16{715, 714, 718}, 726},
		{"L1UpList", []uint16{126}, 139},
		{"L3TWarpUpList", []uint16{181}, 559},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// when
			got := pieceList(c.name)
			// then
			if len(got) < len(c.head) {
				t.Fatalf("%s = %v, too short for head %v", c.name, got, c.head)
			}
			for i, w := range c.head {
				if got[i] != w {
					t.Errorf("%s head[%d] = %d, want %d", c.name, i, got[i], w)
				}
			}
			if got[len(got)-1] != c.tail {
				t.Errorf("%s tail = %d, want %d", c.name, got[len(got)-1], c.tail)
			}
		})
	}
}

func TestEnumParity(t *testing.T) {
	if DTLast != DTCrypt || DTNone != -1 {
		t.Fatalf("DungeonType sentinels wrong: last=%d none=%d", DTLast, DTNone)
	}
	if SLFirstArena != SLArenaChurch || SLLast != SLArenaCircleOfLife {
		t.Fatal("setLevelID arena sentinels wrong")
	}
	if got := SLArenaCircleOfLife - SLFirstArena; got != 2 {
		t.Fatalf("arena count = %d, want 3 entries 0..2", got)
	}
	if ThemeCount != 17 || ThemeNone != -1 {
		t.Fatalf("ThemeID counts wrong: %d/%d", ThemeCount, ThemeNone)
	}
}

func TestTownTriggers(t *testing.T) {
	warps := map[[2]int]struct {
		lvl int
		dt  DungeonType
	}{
		{49, 21}: {5, DTCatacombs},
		{17, 69}: {9, DTCaves},
		{41, 80}: {13, DTHell},
		{80, 62}: {17, DTNest},
		{36, 24}: {21, DTCrypt},
	}
	// when
	tg := townTriggers()
	// then
	found := false
	for _, tr := range tg {
		if tr.X == 25 && tr.Y == 29 {
			found = true
			if tr.Effect != TrigNextLevel {
				t.Fatalf("(25,29) effect = %v, want NEXT", tr.Effect)
			}
		}
	}
	if !found {
		t.Fatal("town (25,29) trigger missing")
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

func TestL4BetrayerGate(t *testing.T) {
	// when
	rules := l4PieceRules()
	// then
	r, ok := rules[369]
	if !ok {
		t.Fatal("piece 369 rule missing")
	}
	if !r.RequiresQuestDone {
		t.Fatal("369 must require Q_BETRAYER done")
	}
	if r.Effect != TrigNextLevel {
		t.Fatalf("369 effect = %v, want NEXT", r.Effect)
	}
	r82, ok := rules[82]
	if !ok || r82.RequiresQuestDone || r82.Effect != TrigPrevLevel {
		t.Fatalf("82 rule = %+v, want unconditional PREV", r82)
	}
}

func TestSetmapManifest(t *testing.T) {
	cases := []struct {
		name   string
		id     setLevelID
		x, y   int
		vx, vy int
	}{
		{"skelking", SLSkelking, 83, 44, 0, 0},
		{"bonechamb", SLBonechamb, 70, 40, 0, 0},
		{"poisonwater", SLPoisonwater, 31, 83, 0, 0},
		{"vilebetrayer", SLVilebetrayer, 35, 36, 0, 0},
		{"arena church", SLArenaChurch, 29, 22, 28, 20},
		{"arena hell", SLArenaHell, 34, 26, 33, 26},
		{"arena circle of life", SLArenaCircleOfLife, 30, 26, 29, 26},
	}
	// when
	m := setmapManifestCopy()
	// then
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, ok := m[c.id]
			if !ok {
				t.Errorf("setmap %v missing", c.id)
				return
			}
			if s.EntryX != c.x || s.EntryY != c.y {
				t.Errorf("%v entry = (%d,%d), want (%d,%d)", c.id, s.EntryX, s.EntryY, c.x, c.y)
			}
			// vx == 0 leaves the visual tile unchecked.
			if c.vx != 0 && (s.VisualX != c.vx || s.VisualY != c.vy) {
				t.Errorf("%v visual = (%d,%d), want (%d,%d)", c.id, s.VisualX, s.VisualY, c.vx, c.vy)
			}
		})
	}
}
