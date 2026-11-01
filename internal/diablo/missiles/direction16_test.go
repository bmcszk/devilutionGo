// Package missiles_test pins GetDirection16 (missiles.cpp:987-1020): the
// flip/mirror decomposition is easy to get subtly wrong, so every branch of
// the upstream tree gets a hand-computed row. Expected values are the
// Direction16 ordinals of tables/misdat.h:147-164.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

var direction16Cases = []direction16Case{
	{
		name: "equal points", from: [2]int{0, 0}, to: [2]int{0, 0},
		want: missiles.Direction16SouthSouthWest, wantName: "South_SouthWest (cpp:1005)",
	},
	{
		name: "pure y span", from: [2]int{0, 0}, to: [2]int{0, 5},
		want: missiles.Direction16SouthWest, wantName: "SouthWest (cpp:1003)",
	},
	{
		name: "shallow y span", from: [2]int{0, 0}, to: [2]int{5, 10},
		want: missiles.Direction16SouthSouthWest, wantName: "South_SouthWest (cpp:1005)",
	},
	{
		name: "square span", from: [2]int{0, 0}, to: [2]int{5, 5},
		want: missiles.Direction16South, wantName: "South (cpp:1002)",
	},
	{
		name: "x dominant", from: [2]int{0, 0}, to: [2]int{10, 5},
		want: missiles.Direction16SouthSouthEast, wantName: "South_SouthEast (cpp:996-998,1018)",
	},
	{
		name: "pure x span", from: [2]int{0, 0}, to: [2]int{5, 0},
		want: missiles.Direction16SouthEast, wantName: "SouthEast (cpp:1018)",
	},
	{
		name: "negative x", from: [2]int{0, 0}, to: [2]int{-5, 0},
		want: missiles.Direction16NorthWest, wantName: "NorthWest (cpp:1010-1019)",
	},
	{
		name: "negative y", from: [2]int{0, 0}, to: [2]int{0, -5},
		want: missiles.Direction16NorthEast, wantName: "NorthEast (cpp:1014-1016)",
	},
	{
		name: "north east diagonal", from: [2]int{0, 0}, to: [2]int{5, -5},
		want: missiles.Direction16East, wantName: "East (cpp:1015)",
	},
	{
		name: "south west diagonal", from: [2]int{0, 0}, to: [2]int{-5, 5},
		want: missiles.Direction16West, wantName: "West (cpp:1011)",
	},
	{
		name: "relative to source", from: [2]int{7, 9}, to: [2]int{17, 19},
		want: missiles.Direction16South, wantName: "South (same offset as square span)",
	},
}

// direction16Case is one hand-computed GetDirection16 row.
type direction16Case struct {
	name     string
	from     [2]int
	to       [2]int
	want     missiles.Direction16
	wantName string
}

func TestGetDirection16(t *testing.T) {
	for _, tt := range direction16Cases {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := missiles.GetDirection16(
				missiles.TilePosition{X: tt.from[0], Y: tt.from[1]},
				missiles.TilePosition{X: tt.to[0], Y: tt.to[1]},
			)

			// then
			if got != tt.want {
				t.Fatalf("GetDirection16 = %d, want %d %s", got, tt.want, tt.wantName)
			}
		})
	}
}
