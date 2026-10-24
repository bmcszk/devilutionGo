// Package leveldat_test covers the level-content tables from outside the
// package (black-box).
package leveldat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat"
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
			got := leveldat.PieceList(name)
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
			got := leveldat.PieceList(c.name)
			// then
			assertPieceListShape(t, c.name, got, c.head, c.tail)
		})
	}
}

// assertPieceListShape checks a piece list's head prefix and tail entry.
func assertPieceListShape(t *testing.T, name string, got, head []uint16, tail uint16) {
	t.Helper()
	if len(got) < len(head) {
		t.Fatalf("%s = %v, too short for head %v", name, got, head)
	}
	for i, w := range head {
		if got[i] != w {
			t.Errorf("%s head[%d] = %d, want %d", name, i, got[i], w)
		}
	}
	if got[len(got)-1] != tail {
		t.Errorf("%s tail = %d, want %d", name, got[len(got)-1], tail)
	}
}

func TestEnumParity(t *testing.T) {
	if leveldat.DTLast != leveldat.DTCrypt || leveldat.DTNone != -1 {
		t.Fatalf("DungeonType sentinels wrong: last=%d none=%d", leveldat.DTLast, leveldat.DTNone)
	}
	if leveldat.SLFirstArena != leveldat.SLArenaChurch || leveldat.SLLast != leveldat.SLArenaCircleOfLife {
		t.Fatal("setLevelID arena sentinels wrong")
	}
	if got := leveldat.SLArenaCircleOfLife - leveldat.SLFirstArena; got != 2 {
		t.Fatalf("arena count = %d, want 3 entries 0..2", got)
	}
	if leveldat.ThemeCount != 17 || leveldat.ThemeNone != -1 {
		t.Fatalf("ThemeID counts wrong: %d/%d", leveldat.ThemeCount, leveldat.ThemeNone)
	}
}

func TestTownTriggers(t *testing.T) {
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
	// when
	tg := leveldat.TownTriggers()
	// then
	assertCathedralStairs(t, tg)
	got := assertTownWarps(t, tg, warps)
	if len(got) != len(warps) {
		t.Fatalf("found %d of %d warp triggers", len(got), len(warps))
	}
}

// assertCathedralStairs checks the (25,29) NEXT trigger exists.
func assertCathedralStairs(t *testing.T, tg []leveldat.Trigger) {
	t.Helper()
	for _, tr := range tg {
		if tr.X == 25 && tr.Y == 29 {
			if tr.Effect != leveldat.TrigNextLevel {
				t.Fatalf("(25,29) effect = %v, want NEXT", tr.Effect)
			}
			return
		}
	}
	t.Fatal("town (25,29) trigger missing")
}

// assertTownWarps checks each gated warp trigger's level and gate; returns
// the set of matched warp keys.
func assertTownWarps(t *testing.T, tg []leveldat.Trigger,
	warps map[[2]int]struct {
		lvl int
		dt  leveldat.DungeonType
	},
) map[[2]int]bool {
	t.Helper()
	got := map[[2]int]bool{}
	for _, tr := range tg {
		key := [2]int{tr.X, tr.Y}
		w, ok := warps[key]
		if !ok {
			continue
		}
		got[key] = true
		if tr.TargetLevel != w.lvl {
			t.Errorf("warp %v target level = %d, want %d", key, tr.TargetLevel, w.lvl)
		}
		if tr.RequiresWarpOpen != w.dt {
			t.Errorf("warp %v gate = %v, want %v", key, tr.RequiresWarpOpen, w.dt)
		}
	}
	return got
}

func TestL4BetrayerGate(t *testing.T) {
	// when
	rules := leveldat.L4PieceRules()
	// then
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
	r82, ok := rules[82]
	if !ok || r82.RequiresQuestDone || r82.Effect != leveldat.TrigPrevLevel {
		t.Fatalf("82 rule = %+v, want unconditional PREV", r82)
	}
}

func TestSetmapManifest(t *testing.T) {
	cases := []struct {
		name   string
		bid    leveldat.SetLevelID
		x, y   int
		vx, vy int
	}{
		{"skelking", leveldat.SLSkelking, 83, 44, 0, 0},
		{"bonechamb", leveldat.SLBonechamb, 70, 40, 0, 0},
		{"poisonwater", leveldat.SLPoisonwater, 31, 83, 0, 0},
		{"vilebetrayer", leveldat.SLVilebetrayer, 35, 36, 0, 0},
		{"arena church", leveldat.SLArenaChurch, 29, 22, 28, 20},
		{"arena hell", leveldat.SLArenaHell, 34, 26, 33, 26},
		{"arena circle of life", leveldat.SLArenaCircleOfLife, 30, 26, 29, 26},
	}
	// when
	m := leveldat.SetmapManifestCopy()
	// then
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, ok := m[c.bid]
			if !ok {
				t.Errorf("setmap %v missing", c.bid)
				return
			}
			assertSetmapEntry(t, c.bid, s, c.x, c.y, c.vx, c.vy)
		})
	}
}

// assertSetmapEntry checks a manifest entry's entry tile and, when wantVX is
// non-zero, its visual tile (vx == 0 leaves the visual tile unchecked).
func assertSetmapEntry(t *testing.T, id leveldat.SetLevelID, s leveldat.Setmap,
	x, y, wantVX, wantVY int,
) {
	t.Helper()
	if s.EntryX != x || s.EntryY != y {
		t.Errorf("%v entry = (%d,%d), want (%d,%d)", id, s.EntryX, s.EntryY, x, y)
	}
	if wantVX != 0 && (s.VisualX != wantVX || s.VisualY != wantVY) {
		t.Errorf("%v visual = (%d,%d), want (%d,%d)", id, s.VisualX, s.VisualY, wantVX, wantVY)
	}
}
