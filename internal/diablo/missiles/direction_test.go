// Package missiles_test pins GetDirection (engine/point.hpp:246-284).
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

// directionCases pin the classifier as upstream computes it (point.hpp:252-283):
// the 5*mx <= 2*my test only splits off the diagonals, and the mx-dominant
// check that runs after it promotes the cardinal to the south-east
// (point.hpp:264-265) — so east-dominant spans all land on SouthEast and the
// zero displacement ties into SouthWest (0 <= 0).
var directionCases = []struct {
	name string
	dst  missiles.TilePosition
	want missiles.Direction
}{
	{"zero displacement ties to south west", missiles.TilePosition{X: 4, Y: 4}, missiles.SouthWest},
	{"steep south with one column is south west", missiles.TilePosition{X: 5, Y: 9}, missiles.SouthWest},
	{"south east span", missiles.TilePosition{X: 9, Y: 4}, missiles.SouthEast},
	{"south east span with a row offset", missiles.TilePosition{X: 9, Y: 5}, missiles.SouthEast},
	{"north east span becomes south east", missiles.TilePosition{X: 9, Y: 3}, missiles.SouthEast},
	{"east dominant stays east", missiles.TilePosition{X: 6, Y: 0}, missiles.East},
	{"steep north is north east", missiles.TilePosition{X: 4, Y: 0}, missiles.NorthEast},
	{"north west diagonal is north", missiles.TilePosition{X: 0, Y: 0}, missiles.North},
	{"west dominant with one row is north west", missiles.TilePosition{X: 0, Y: 4}, missiles.NorthWest},
	{"steep north west with one column is north east", missiles.TilePosition{X: 3, Y: 0}, missiles.NorthEast},
	{"west dominant stays west", missiles.TilePosition{X: 3, Y: 6}, missiles.West},
	{"steep west with one column is south west", missiles.TilePosition{X: 3, Y: 9}, missiles.SouthWest},
}

// TestGetDirection pins the eight-way classifier.
func TestGetDirection(t *testing.T) {
	// given
	start := missiles.TilePosition{X: 4, Y: 4}
	cases := directionCases

	// when
	subtests := make([]subtest, 0, len(cases))
	for _, tc := range cases {
		subtests = append(subtests, subtest{tc.name, func(t *testing.T) {
			t.Helper()

			// when
			got := missiles.GetDirection(start, tc.dst)

			// then
			if got != tc.want {
				t.Fatalf("GetDirection(%v, %v) = %v, want %v", start, tc.dst, got, tc.want)
			}
		}})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}
