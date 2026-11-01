package quests_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/enums"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests"
)

// questResult is the expected ForceQuests answer and target.
type questResult struct {
	ok     bool
	name   string
	cursor quests.Position
}

// questCase is one ForceQuests row.
type questCase struct {
	name   string
	flags  enums.SimFlags
	level  int
	cursor quests.Position
	portal quests.Position
	want   questResult
}

// ForceQuests snaps the cursor into a quest portal on the quest's own level and
// names the map, but never in multiplayer or the shareware build
// (quests.cpp:203-225).
func TestForceQuests(t *testing.T) {
	tests := []questCase{
		{
			name:   "skeletonKingPortal",
			level:  3,
			cursor: quests.Position{X: 10, Y: 10},
			portal: quests.Position{X: 12, Y: 12},
			want: questResult{
				ok:     true,
				name:   "King Leoric's Tomb",
				cursor: quests.Position{X: 12, Y: 12},
			},
		},
		{
			name:   "wrongLevel",
			level:  4,
			cursor: quests.Position{X: 10, Y: 10},
			portal: quests.Position{X: 12, Y: 12},
		},
		{
			name:   "multiplayer",
			flags:  enums.SimFlags{Multiplayer: true},
			level:  3,
			cursor: quests.Position{X: 10, Y: 10},
			portal: quests.Position{X: 12, Y: 12},
		},
		{
			name:   "spawn",
			flags:  enums.SimFlags{Spawn: true},
			level:  3,
			cursor: quests.Position{X: 10, Y: 10},
			portal: quests.Position{X: 12, Y: 12},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			qs := mustInit(t, tt.flags)
			positions := make([]quests.Position, len(qs))
			positions[quests.QSkelking] = tt.portal
			contains := exactContains(tt.portal, tt.cursor)
			// when
			got, ok := quests.ForceQuests(qs, positions, tt.cursor,
				quests.ResyncContext{Level: tt.level, Flags: tt.flags}, contains)
			// then
			assertResult(t, questResult{ok: ok, name: got.MapName, cursor: got.Cursor}, tt.want)
		})
	}
}

// exactContains is EntranceBoundaryContains for the single portal a test sets:
// only the exact quest position around the exact cursor counts.
func exactContains(portal, cursor quests.Position) func(a, b quests.Position) bool {
	return func(a, b quests.Position) bool {
		return a == portal && b == cursor
	}
}

// assertResult pins the ForceQuests answer and target.
func assertResult(t *testing.T, got, want questResult) {
	t.Helper()

	if got != want {
		t.Fatalf("ForceQuests = %+v, want %+v", got, want)
	}
}
