// Package missiles_test ports the devilGo-mwlm slice-1 acceptance tests:
// the behavior registry must cover every base misdat.tsv MissileID (row
// count asserted at runtime from the on-disk TSV — never hardcoded, per
// the AC fix), and the 16.16 motion arithmetic must be exact for scripted
// velocities including negative deltas (missiles.cpp:2810-2830).
package missiles_test

import (
	"os"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

const baseMisdatPath = "../../../internal/testdata/gamedata/misdat/misdat.tsv"

// baseMisdatRows reads and parses the base misdat.tsv corpus
// (LoadMisdat, Source/tables/misdat.cpp:373-396).
func baseMisdatRows(t *testing.T) []misdat.MissileData {
	t.Helper()
	data, err := os.ReadFile(baseMisdatPath)
	if err != nil {
		t.Fatalf("read committed misdat fixture: %v", err)
	}
	rows, err := misdat.ParseMisdat(data)
	if err != nil {
		t.Fatalf("misdat.ParseMisdat: %v", err)
	}
	return rows
}

// TestRegistryCoversMisdat: the registry must hold a Behavior for every
// MissileID in the base misdat.tsv. The expected count is taken from the
// on-disk file at test runtime (AC fix: never hardcoded). Row order ==
// MissileID ordinal (GetMissileData, Source/tables/misdat.cpp:451).
func TestRegistryCoversMisdat(t *testing.T) {
	rows := baseMisdatRows(t)
	want := len(rows)
	if got := missiles.RegistryCount(); got != want {
		t.Fatalf("RegistryCount() = %d, want %d (misdat.tsv rows counted at runtime)", got, want)
	}
	ids := make([]missiles.MissileID, len(rows))
	for i := range rows {
		ids[i] = missiles.MissileID(i) // row order == enum ordinal (misdat.cpp:451)
	}
	if missing := missiles.MissingFromRegistry(ids); len(missing) != 0 {
		t.Fatalf("MissingFromRegistry = %v, want empty", missing)
	}
	// The two named slice-1 stubs (Arrow, Firebolt) resolve to callable
	// behavior pairs — placeholders, never panics.
	for _, id := range []missiles.MissileID{missiles.Arrow, missiles.Firebolt} {
		b, ok := missiles.BehaviorFor(id)
		if !ok {
			t.Fatalf("BehaviorFor(%d) not registered", id)
		}
		if b.Add == nil || b.Process == nil {
			t.Fatalf("BehaviorFor(%d) has nil Add/Process", id)
		}
	}
}

// TestMotion16Dot16: scripted pixel velocities drive the 16.16 fixed
// point travel arithmetic (MoveMissile traveled += velocity<<16,
// missiles.cpp:2810-2830 semantics; UpdateMissilePos pixelsTravelled =
// traveled >> 16, missiles.cpp:213-215) — exact, including negative
// deltas (traveled is persisted as two int32, loadsave.cpp:852-853).
func TestMotion16Dot16(t *testing.T) {
	tests := []struct {
		name         string
		startTile    [2]int
		steps        []missiles.PositionDisplacement
		wantTraveled [2]int32
		wantTile     [2]int
		wantMid      [2]int32
	}{
		{
			name:         "zero velocity holds position",
			startTile:    [2]int{63, 63},
			steps:        []missiles.PositionDisplacement{{DeltaX: 0, DeltaY: 0}},
			wantTraveled: [2]int32{63 << 16, 63 << 16},
			wantTile:     [2]int{63, 63},
			wantMid:      [2]int32{63, 63},
		},
		{
			name:         "arrow diagonal 32,16 pixels",
			startTile:    [2]int{63, 63},
			steps:        []missiles.PositionDisplacement{{DeltaX: 32, DeltaY: 16}},
			wantTraveled: [2]int32{95 << 16, 79 << 16},
			wantTile:     [2]int{95, 79},
			wantMid:      [2]int32{95, 79},
		},
		{
			name:      "negative velocities accumulate exactly",
			startTile: [2]int{0, 0},
			steps: []missiles.PositionDisplacement{
				{DeltaX: -3, DeltaY: -5},
				{DeltaX: -4, DeltaY: 2},
			},
			wantTraveled: [2]int32{-7 << 16, -3 << 16},
			wantTile:     [2]int{-7, -3},
			wantMid:      [2]int32{-7, -3},
		},
		{
			name:      "multi step scripted velocities",
			startTile: [2]int{5, 5},
			steps: []missiles.PositionDisplacement{
				{DeltaX: 10, DeltaY: 7},
				{DeltaX: 0, DeltaY: 3},
				{DeltaX: -2, DeltaY: -1},
			},
			wantTraveled: [2]int32{13 << 16, 14 << 16},
			wantTile:     [2]int{13, 14},
			wantMid:      [2]int32{13, 14},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m missiles.Missile
			m.SetTraveledFromTile(tt.startTile[0], tt.startTile[1])
			wantX := int32(tt.startTile[0]) << 16
			wantY := int32(tt.startTile[1]) << 16
			if m.Position.Traveled.DeltaX != wantX || m.Position.Traveled.DeltaY != wantY {
				t.Fatalf("SetTraveledFromTile: traveled = (%d, %d), want (%d, %d)",
					m.Position.Traveled.DeltaX, m.Position.Traveled.DeltaY, wantX, wantY)
			}
			for _, v := range tt.steps {
				missiles.Step(&m, v)
			}
			if m.Position.Traveled.DeltaX != tt.wantTraveled[0] || m.Position.Traveled.DeltaY != tt.wantTraveled[1] {
				t.Errorf("traveled = (%d, %d), want (%d, %d)",
					m.Position.Traveled.DeltaX, m.Position.Traveled.DeltaY, tt.wantTraveled[0], tt.wantTraveled[1])
			}
			if m.Position.Tile.X != tt.wantTile[0] || m.Position.Tile.Y != tt.wantTile[1] {
				t.Errorf("tile = (%d, %d), want (%d, %d)",
					m.Position.Tile.X, m.Position.Tile.Y, tt.wantTile[0], tt.wantTile[1])
			}
			mid := m.MidPosition()
			if mid.DeltaX != tt.wantMid[0] || mid.DeltaY != tt.wantMid[1] {
				t.Errorf("mid = (%d, %d), want (%d, %d)",
					mid.DeltaX, mid.DeltaY, tt.wantMid[0], tt.wantMid[1])
			}
		})
	}
}

// TestMidPositionFloorsFractionalTraveled: mid-position = traveled >> 16
// is an arithmetic shift — fractional pixels floor toward negative
// infinity, matching C++ int32 right-shift semantics (missiles.cpp:215).
func TestMidPositionFloorsFractionalTraveled(t *testing.T) {
	m := &missiles.Missile{
		Position: missiles.Position{
			Traveled: missiles.PositionDisplacement{
				DeltaX: -(1 << 16) - 32768, // -1.5 px
				DeltaY: 2<<16 + 49152,      // 2.75 px
			},
		},
	}
	mid := m.MidPosition()
	if mid.DeltaX != -2 || mid.DeltaY != 2 {
		t.Fatalf("mid = (%d, %d), want (-2, 2)", mid.DeltaX, mid.DeltaY)
	}
	missiles.Step(m, missiles.PositionDisplacement{DeltaX: 1, DeltaY: 1})
	if m.Position.Traveled.DeltaX != -32768 || m.Position.Traveled.DeltaY != 3<<16+49152 {
		t.Fatalf("traveled = (%d, %d), want (-32768, %d)",
			m.Position.Traveled.DeltaX, m.Position.Traveled.DeltaY, 3<<16+49152)
	}
	if m.Position.Tile.X != -1 || m.Position.Tile.Y != 3 {
		t.Fatalf("tile = (%d, %d), want (-1, 3)", m.Position.Tile.X, m.Position.Tile.Y)
	}
}

// TestPathDistance: WalkingDistance port (Source/engine/point.hpp:154-159)
// — max(|dx|, |dy|), the tile-hop count MoveMissile uses for axis-aligned
// travel (missiles.cpp:560-563). Direction-independent.
func TestPathDistance(t *testing.T) {
	tests := []struct {
		name string
		from missiles.TilePosition
		to   missiles.TilePosition
		want int
	}{
		{"same tile", missiles.TilePosition{X: 5, Y: 5}, missiles.TilePosition{X: 5, Y: 5}, 0},
		{"diagonal dominated by x", missiles.TilePosition{X: 5, Y: 5}, missiles.TilePosition{X: 8, Y: 6}, 3},
		{"diagonal dominated by y", missiles.TilePosition{X: 5, Y: 5}, missiles.TilePosition{X: 6, Y: 9}, 4},
		{"reverse direction", missiles.TilePosition{X: 8, Y: 6}, missiles.TilePosition{X: 5, Y: 5}, 3},
		{"negative deltas", missiles.TilePosition{X: -2, Y: 3}, missiles.TilePosition{X: 1, Y: -4}, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := missiles.PathDistance(tt.from, tt.to); got != tt.want {
				t.Fatalf("PathDistance(%v, %v) = %d, want %d", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
