package levels_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/levels"
)

// TestSelectChamber pins the chamber pick table for every chamber-availability
// combination FirstRoom can leave behind, plus the layout mapping the set piece
// drops into (levels/drlg_l1.cpp:1330-1356).
//
//nolint:revive,funlen // the availability x draw matrix is the AC; splitting it would hide which cell failed
func TestSelectChamber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                         string
		chamber1, chamber2, chamber3 bool
		vertical                     bool
		draw                         int32
		wantChamber                  int
		wantOrigin                   levels.Point
	}{
		{
			name:     "all three chambers draw 1 of 3",
			chamber1: true, chamber2: true, chamber3: true, draw: 0,
			wantChamber: 1, wantOrigin: levels.Point{X: 2, Y: 16},
		},
		{
			name:     "all three chambers draw 2 of 3",
			chamber1: true, chamber2: true, chamber3: true, draw: 1,
			wantChamber: 2, wantOrigin: levels.Point{X: 16, Y: 16},
		},
		{
			name:     "all three chambers draw 3 of 3",
			chamber1: true, chamber2: true, chamber3: true, draw: 2,
			wantChamber: 3, wantOrigin: levels.Point{X: 30, Y: 16},
		},
		{
			name:     "chamber 1 and 2 draw the vanilla order",
			chamber1: true, chamber2: true, draw: 0,
			wantChamber: 2, wantOrigin: levels.Point{X: 16, Y: 16},
		},
		{
			name:     "chamber 1 and 2 draw the other side",
			chamber1: true, chamber2: true, draw: 1,
			wantChamber: 1, wantOrigin: levels.Point{X: 2, Y: 16},
		},
		{
			name:     "chamber 1 and 3 draw the vanilla order",
			chamber1: true, chamber3: true, draw: 0,
			wantChamber: 3, wantOrigin: levels.Point{X: 30, Y: 16},
		},
		{
			name:     "chamber 1 and 3 draw the other side",
			chamber1: true, chamber3: true, draw: 1,
			wantChamber: 1, wantOrigin: levels.Point{X: 2, Y: 16},
		},
		{
			name:     "chamber 2 and 3 draw the vanilla order",
			chamber2: true, chamber3: true, draw: 0,
			wantChamber: 2, wantOrigin: levels.Point{X: 16, Y: 16},
		},
		{
			name:     "chamber 2 and 3 draw the other side",
			chamber2: true, chamber3: true, draw: 1,
			wantChamber: 3, wantOrigin: levels.Point{X: 30, Y: 16},
		},
		{
			name:        "a lone middle chamber draws nothing",
			chamber2:    true,
			wantChamber: 2, wantOrigin: levels.Point{X: 16, Y: 16},
		},
		{
			name:     "a vertical layout swaps the corner origins",
			chamber1: true, chamber2: true, chamber3: true, vertical: true, draw: 0,
			wantChamber: 1, wantOrigin: levels.Point{X: 16, Y: 2},
		},
		{
			name:     "a vertical layout keeps the middle chamber centered",
			chamber1: true, chamber2: true, chamber3: true, vertical: true, draw: 1,
			wantChamber: 2, wantOrigin: levels.Point{X: 16, Y: 16},
		},
		{
			name:     "a vertical layout flips the third chamber too",
			chamber1: true, chamber2: true, chamber3: true, vertical: true, draw: 2,
			wantChamber: 3, wantOrigin: levels.Point{X: 16, Y: 30},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// given: a layout with the named chambers and a scripted draw.
			dungeon := levels.NewDungeon(levels.Cathedral, 1)
			dungeon.HasChamber1 = test.chamber1
			dungeon.HasChamber2 = test.chamber2
			dungeon.HasChamber3 = test.chamber3
			dungeon.VerticalLayout = test.vertical
			dungeon.Rnd.GenerateRnd = func(int32) int32 { return test.draw }

			// when: SelectChamber picks the set piece room.
			chamber, origin := dungeon.SelectChamberForTest()

			// then
			if chamber != test.wantChamber {
				t.Fatalf("chamber = %d, want %d", chamber, test.wantChamber)
			}

			if origin != test.wantOrigin {
				t.Fatalf("origin = %+v, want %+v", origin, test.wantOrigin)
			}
		})
	}
}
