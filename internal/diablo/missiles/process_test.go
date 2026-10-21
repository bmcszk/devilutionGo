// Package missiles_test: devilGo-mwlm slice 3a-beta acceptance tests —
// the motion primitives and the ProcessArrow/ProcessGenericProjectile
// ports (Source/missiles.cpp:173-249, 552-663, 2932-3016). Every expected
// number below is upstream-exact: the 16.16 tapes were derived by running
// the ported arithmetic line by line (float32 truncation included), and
// the conversion probes reuse the displacement.go golden values.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// Scripted 16.16 velocities for the tapes:
//   - axis 11 px/tick (screen-x direction) — crosses tiles every ~6 ticks,
//     first tile change on tick 3 (tape below);
//   - diagonal 64×32 px/tick — Manhattan distance 2 per tick, so
//     MoveMissile's interpolated-tile loop runs; the increment is exactly
//     32×16 px (the iso tile diagonal, (32<<16)/denominator with
//     denominator = 2*speed.dy = speed.dx).
const (
	axisVX = 11 << 16
	axisVY = 0
	diagVX = 64 << 16
	diagVY = 32 << 16
)

// Scripted Arrow damage seam values (SourceDamage stand-in for the
// player _pIMinDam/_pIMaxDam read, missiles.cpp:2938-2942).
const (
	arrowMinDam = 3
	arrowMaxDam = 9
)

// freshMissile seeds the upstream AddMissile state (missiles.cpp:2823-2827):
// tile = start = src, traveled/velocity/offset zero, no light, hash 0.
// (Slice-1's SetTraveledFromTile is the loadsave reload shape, not the
// spawn shape — AddMissile leaves traveled zeroed.)
func freshMissile(x, y int) missiles.Missile {
	var m missiles.Missile
	m.Position.Tile = missiles.TilePosition{X: x, Y: y}
	m.Position.Start = m.Position.Tile
	return m
}

// procWorld builds an open-floor HitWorld (world.NewGrid: SOL entry 0 is
// walkable, no objects) with empty occupancy tables.
func procWorld() *missiles.HitWorld {
	return &missiles.HitWorld{Grid: world.NewGrid(),
		ProcessCallbacks: missiles.NoopProcessCallbacks()}
}

// lightSpies records every light-hook dispatch (ChangeLightOffset,
// lighting.h:64; ChangeLight, lighting.h:65; AddUnLight, lighting.h:61).
type lightSpies struct {
	offsets    [][3]int // lid, offsetX, offsetY
	changes    [][4]int // lid, x, y, radius
	unlights   []int    // lid
	spawnCalls []spawnCall
	sourceDam  []bool
}

// spawnCall records one SpawnMissile dispatch (the AddMissile tail of
// the explosion switch, missiles.cpp:2982-3001).
type spawnCall struct {
	src, dst missiles.TilePosition
	dir      missiles.Direction
	spawnID  missiles.MissileID
	caster   int
	source   int
	parent   *missiles.Missile
}

// wire installs the recording hooks on the world's process sink; the
// returned func restores the previous callbacks (defer in each test).
func (s *lightSpies) wire(w *missiles.HitWorld) func() {
	prev := w.ProcessCallbacks
	cb := missiles.NoopProcessCallbacks()
	cb.ChangeLightOffset = func(lid int, offsetX, offsetY int8) {
		s.offsets = append(s.offsets, [3]int{lid, int(offsetX), int(offsetY)})
	}
	cb.ChangeLight = func(lid int, x, y int, radius uint8) {
		s.changes = append(s.changes, [4]int{lid, x, y, int(radius)})
	}
	cb.AddUnLight = func(lid int) {
		s.unlights = append(s.unlights, lid)
	}
	cb.SpawnMissile = func(src, dst missiles.TilePosition, dir missiles.Direction,
		spawnID missiles.MissileID, caster, source int, parent *missiles.Missile) {
		s.spawnCalls = append(s.spawnCalls, spawnCall{src, dst, dir, spawnID, caster, source, parent})
	}
	cb.SourceDamage = func(m *missiles.Missile) (int, int) {
		s.sourceDam = append(s.sourceDam, true)
		return arrowMinDam, arrowMaxDam
	}
	w.ProcessCallbacks = cb
	return func() { w.ProcessCallbacks = prev }
}

// TestUpdateMissileVelocity: velocity = worldToNormalScreen(tile -
// destination) * velocityInPixels (missiles.cpp:173-186); same-tile
// short-circuits to {0,0}. Expected vectors: unit normals from
// displacement.hpp:363-370 float32 math ((±46340, ∓23170) for the
// cardinal directions), scaled by 16.
func TestUpdateMissileVelocity(t *testing.T) {
	tests := []struct {
		name        string
		tile        missiles.TilePosition
		destination missiles.TilePosition
		pixels      int
		want        missiles.PositionDisplacement
	}{
		{"same tile zeroes velocity", missiles.TilePosition{X: 64, Y: 64},
			missiles.TilePosition{X: 64, Y: 64}, 16, missiles.PositionDisplacement{}},
		{"tile east of destination", missiles.TilePosition{X: 67, Y: 64},
			missiles.TilePosition{X: 64, Y: 64}, 16,
			missiles.PositionDisplacement{DeltaX: -741440, DeltaY: -370720}},
		{"tile south of destination", missiles.TilePosition{X: 64, Y: 67},
			missiles.TilePosition{X: 64, Y: 64}, 16,
			missiles.PositionDisplacement{DeltaX: 741440, DeltaY: -370720}},
		{"tile north of destination", missiles.TilePosition{X: 64, Y: 61},
			missiles.TilePosition{X: 64, Y: 64}, 16,
			missiles.PositionDisplacement{DeltaX: -741440, DeltaY: 370720}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := freshMissile(tt.tile.X, tt.tile.Y)
			missiles.UpdateMissileVelocity(&m, tt.destination, tt.pixels)
			if m.Position.Velocity != tt.want {
				t.Fatalf("velocity = %v, want %v", m.Position.Velocity, tt.want)
			}
		})
	}
}

// TestStopMissile: velocity zeroed; offset snaps to the renderer offset
// only when the renderer tile matches (missiles.h:49-54).
func TestStopMissile(t *testing.T) {
	t.Run("renderer on tile snaps offset", func(t *testing.T) {
		m := freshMissile(64, 64)
		m.Position.Velocity = missiles.PositionDisplacement{DeltaX: 5, DeltaY: 5}
		m.Position.Offset = missiles.PositionDisplacement{DeltaX: 100, DeltaY: 200}
		m.Position.TileForRendering = m.Position.Tile
		m.Position.OffsetForRendering = missiles.PositionDisplacement{DeltaX: 7, DeltaY: 9}
		missiles.StopMissile(&m)
		if m.Position.Velocity != (missiles.PositionDisplacement{}) {
			t.Fatalf("velocity = %v, want zero", m.Position.Velocity)
		}
		if m.Position.Offset != (missiles.PositionDisplacement{DeltaX: 7, DeltaY: 9}) {
			t.Fatalf("offset = %v, want renderer offset {7 9}", m.Position.Offset)
		}
	})
	t.Run("renderer elsewhere keeps offset", func(t *testing.T) {
		m := freshMissile(64, 64)
		m.Position.Velocity = missiles.PositionDisplacement{DeltaX: 5, DeltaY: 5}
		m.Position.Offset = missiles.PositionDisplacement{DeltaX: 100, DeltaY: 200}
		m.Position.TileForRendering = missiles.TilePosition{X: 1, Y: 1}
		m.Position.OffsetForRendering = missiles.PositionDisplacement{DeltaX: 7, DeltaY: 9}
		missiles.StopMissile(&m)
		if m.Position.Offset != (missiles.PositionDisplacement{DeltaX: 100, DeltaY: 200}) {
			t.Fatalf("offset = %v, want {100 200}", m.Position.Offset)
		}
	})
}

// TestPutMissile: the dFlags registration (missiles.cpp:191-211) —
// Missile bit always, MissileFireWall/MissileLightningWall per type,
// DelFlag short-circuit (including out-of-bounds raising it), and the
// MissilePreFlag global.
func TestPutMissile(t *testing.T) {
	type want struct {
		flags   world.DungeonFlag
		delFlag bool
		preFlag bool
	}
	tests := []struct {
		name    string
		typ     missiles.MissileID
		tile    missiles.TilePosition
		delFlag bool
		preFlag bool
		want    want
	}{
		{"arrow sets missile bit", missiles.Arrow,
			missiles.TilePosition{X: 64, Y: 64}, false, false,
			want{flags: world.FlagMissile}},
		{"firewall adds fire bit", missiles.FireWall,
			missiles.TilePosition{X: 64, Y: 64}, false, false,
			want{flags: world.FlagMissile | world.FlagMissileFire}},
		{"lightningwall adds lightning bit", missiles.LightningWall,
			missiles.TilePosition{X: 64, Y: 64}, false, false,
			want{flags: world.FlagMissile | world.FlagMissileLight}},
		{"pre-flagged missile raises the global", missiles.Arrow,
			missiles.TilePosition{X: 64, Y: 64}, false, true,
			want{flags: world.FlagMissile, preFlag: true}},
		{"del flag skips registration", missiles.Arrow,
			missiles.TilePosition{X: 64, Y: 64}, true, false,
			want{delFlag: true}},
		{"out of bounds raises del flag", missiles.Arrow,
			missiles.TilePosition{X: -1, Y: 64}, false, false,
			want{delFlag: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &missiles.HitWorld{Grid: world.NewGrid()}

			m := freshMissile(tt.tile.X, tt.tile.Y)
			m.Type = tt.typ
			m.DelFlag = tt.delFlag
			m.PreFlag = tt.preFlag
			missiles.PutMissile(w, &m)

			if m.DelFlag != tt.want.delFlag {
				t.Fatalf("DelFlag = %v, want %v", m.DelFlag, tt.want.delFlag)
			}
			if w.ProcessCallbacks.MissilePreFlag != tt.want.preFlag {
				t.Fatalf("MissilePreFlag = %v, want %v",
					w.ProcessCallbacks.MissilePreFlag, tt.want.preFlag)
			}
			if !world.InDungeonBounds(tt.tile.X, tt.tile.Y) || tt.want.delFlag {
				return // nothing may be written
			}
			got := w.Grid.DFlags[tt.tile.Y][tt.tile.X]
			if got != tt.want.flags {
				t.Fatalf("dFlags = %v, want %v", got, tt.want.flags)
			}
		})
	}
}

// TestUpdateMissilePos: traveled>>16 → screenToMissile → tile = start +
// tileOffset; offset = pixels + worldToScreen(tileOffset); light offset =
// screenToLight(pixels) - tileOffset*8 (missiles.cpp:213-224). Golden
// vectors hand-run through the exact integer/float32 arithmetic.
func TestUpdateMissilePos(t *testing.T) {
	tests := []struct {
		name        string
		traveled    missiles.PositionDisplacement
		wantTile    missiles.TilePosition
		wantOffset  missiles.PositionDisplacement
		wantLightXY [2]int
	}{
		{"fractional diagonal floors the pixel vector",
			missiles.PositionDisplacement{DeltaX: -11 << 16, DeltaY: 13<<16 + 32768},
			missiles.TilePosition{X: 0, Y: 1},
			missiles.PositionDisplacement{DeltaX: 21, DeltaY: -3},
			[2]int{1, -4}},
		{"negative pixel vector wraps through iso space",
			missiles.PositionDisplacement{DeltaX: -32 << 16, DeltaY: -8 << 16},
			missiles.TilePosition{X: -1, Y: 0},
			missiles.PositionDisplacement{DeltaX: 0, DeltaY: 8},
			[2]int{2, 2}},
		{"whole tiles inside the same tile keep offset small",
			missiles.PositionDisplacement{DeltaX: 3 << 16, DeltaY: 0},
			missiles.TilePosition{X: 0, Y: 0},
			missiles.PositionDisplacement{DeltaX: 3, DeltaY: 0},
			[2]int{0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spies := &lightSpies{}
			w := &missiles.HitWorld{}
			restore := spies.wire(w)
			defer restore()

			m := freshMissile(0, 0)
			m.Lid = 3
			m.Position.Traveled = tt.traveled
			missiles.UpdateMissilePos(w, &m)

			if m.Position.Tile != tt.wantTile {
				t.Errorf("tile = %v, want %v", m.Position.Tile, tt.wantTile)
			}
			if m.Position.Offset != tt.wantOffset {
				t.Errorf("offset = %v, want %v", m.Position.Offset, tt.wantOffset)
			}
			if len(spies.offsets) != 1 {
				t.Fatalf("ChangeLightOffset calls = %d, want 1", len(spies.offsets))
			}
			got := spies.offsets[0]
			if got[0] != 3 || got[1] != tt.wantLightXY[0] || got[2] != tt.wantLightXY[1] {
				t.Errorf("ChangeLightOffset = %v, want lid 3 offset %v", got, tt.wantLightXY)
			}
		})
	}
}

// TestMoveMissilePos: the charging-monster correction (missiles.cpp:233-251)
// — East→SE, West→SW, South/SW/SE→South; the tile moves by the direction
// delta and the offset by worldToScreen(delta). The IsTileAvailable gate
// is the injected seam.
func TestMoveMissilePos(t *testing.T) {
	tests := []struct {
		name       string
		dir        missiles.Direction
		available  bool
		wantCalled bool
		wantDelta  missiles.TilePosition
		wantOff    missiles.PositionDisplacement
	}{
		{"east shifts south-east", missiles.East, true, true,
			missiles.TilePosition{X: 1, Y: 0}, missiles.PositionDisplacement{DeltaX: -32, DeltaY: -16}},
		{"west shifts south-west", missiles.West, true, true,
			missiles.TilePosition{X: 0, Y: 1}, missiles.PositionDisplacement{DeltaX: 32, DeltaY: -16}},
		{"south shifts south", missiles.South, true, true,
			missiles.TilePosition{X: 1, Y: 1}, missiles.PositionDisplacement{DeltaX: 0, DeltaY: -32}},
		{"north is untouched", missiles.North, true, false,
			missiles.TilePosition{X: 0, Y: 0}, missiles.PositionDisplacement{}},
		{"unavailable tile is untouched", missiles.East, false, true,
			missiles.TilePosition{X: 0, Y: 0}, missiles.PositionDisplacement{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := 0
			w := &missiles.HitWorld{}
			cb := missiles.NoopProcessCallbacks()
			cb.IsTileAvailable = func(sourceMonster, x, y int) bool {
				called++
				return tt.available
			}
			w.ProcessCallbacks = cb

			m := freshMissile(64, 64)
			m.MFNum = int(tt.dir)
			missiles.MoveMissilePos(w, &m)

			if called != boolCount(tt.wantCalled) {
				t.Fatalf("IsTileAvailable calls = %d, want %d", called, boolCount(tt.wantCalled))
			}
			if m.Position.Tile != (missiles.TilePosition{X: 64 + tt.wantDelta.X, Y: 64 + tt.wantDelta.Y}) {
				t.Fatalf("tile = %v, want %+v", m.Position.Tile, tt.wantDelta)
			}
			if m.Position.Offset != tt.wantOff {
				t.Fatalf("offset = %v, want %v", m.Position.Offset, tt.wantOff)
			}
		})
	}
}

// boolCount is the test-side 0/1 view of a control boolean.
func boolCount(b bool) int {
	if b {
		return 1
	}
	return 0
}

// moveTape drives MoveMissile tick by tick and records what the checkTile
// spy saw, so the table rows below assert exact visit order.
type moveTape struct {
	world   *missiles.HitWorld
	probeIn func(tile missiles.TilePosition) bool // scripted checkTile outcome
	probes  []missiles.TilePosition
}

func newMoveTape(w *missiles.HitWorld) *moveTape { return &moveTape{world: w} }

func (mt *moveTape) checkTile(tile missiles.TilePosition) bool {
	mt.probes = append(mt.probes, tile)
	if mt.probeIn != nil {
		return mt.probeIn(tile)
	}
	return true
}

// tickMove runs one MoveMissile tick and asserts the full post-state
// (wantChanged is upstream's return: false only for the no-motion tick,
// missiles.cpp:565).
func tickMove(t *testing.T, mt *moveTape, m *missiles.Missile, dontMove, wantChanged bool, //nolint:revive // tape fixture: the upstream MoveMissile state pack
	wantProbes []missiles.TilePosition, wantTile missiles.TilePosition,
	wantTraveled, wantOffset missiles.PositionDisplacement, wantLight [2]int) {
	t.Helper()

	spies := &lightSpies{}
	restore := spies.wire(mt.world)
	defer restore()

	changed := missiles.MoveMissile(mt.world, m, mt.checkTile, dontMove)
	if changed != wantChanged {
		t.Fatalf("MoveMissile changed = %v, want %v", changed, wantChanged)
	}
	if len(mt.probes) != len(wantProbes) {
		t.Fatalf("probes = %v, want %v", mt.probes, wantProbes)
	}
	for i, want := range wantProbes {
		if mt.probes[i] != want {
			t.Fatalf("probe[%d] = %v, want %v", i, mt.probes[i], want)
		}
	}
	if m.Position.Tile != wantTile {
		t.Fatalf("tile = %v, want %v", m.Position.Tile, wantTile)
	}
	if m.Position.Traveled != wantTraveled {
		t.Fatalf("traveled = %v, want %v", m.Position.Traveled, wantTraveled)
	}
	if m.Position.Offset != wantOffset {
		t.Fatalf("offset = %v, want %v", m.Position.Offset, wantOffset)
	}
	if len(spies.offsets) != 1 {
		t.Fatalf("ChangeLightOffset calls = %d, want 1", len(spies.offsets))
	}
	if wantLight != [2]int{0, 0} {
		got := spies.offsets[0]
		if got[1] != wantLight[0] || got[2] != wantLight[1] {
			t.Fatalf("light offset = %v, want %v", got[1:], wantLight)
		}
	}
}

// TestMoveMissileAxisTape: 11 px/tick screen-x travel from (64,64),
// open floor — ticks 0-1 stay inside the start tile (possibleVisitTiles
// = 0 → false, no probes), tick 2 lands on (65,63) and probes exactly
// the destination tile (missiles.cpp:552-628).
func TestMoveMissileAxisTape(t *testing.T) {
	w := procWorld()
	m := freshMissile(64, 64)
	m.Lid = 7
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: axisVX, DeltaY: axisVY}
	mt := newMoveTape(w)

	tickMove(t, mt, &m, false, false, nil,
		missiles.TilePosition{X: 64, Y: 64},
		missiles.PositionDisplacement{DeltaX: 720896, DeltaY: 0},
		missiles.PositionDisplacement{DeltaX: 11, DeltaY: 0}, [2]int{1, -1})
	t.Run("tick2", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, false, false,
			nil,
			missiles.TilePosition{X: 64, Y: 64},
			missiles.PositionDisplacement{DeltaX: 1441792, DeltaY: 0},
			missiles.PositionDisplacement{DeltaX: 22, DeltaY: 0}, [2]int{2, -2})
	})
	t.Run("tick3", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, false, true,
			[]missiles.TilePosition{{X: 65, Y: 63}},
			missiles.TilePosition{X: 65, Y: 63},
			missiles.PositionDisplacement{DeltaX: 2162688, DeltaY: 0},
			missiles.PositionDisplacement{DeltaX: -31, DeltaY: 0}, [2]int{-4, 4})
	})
}

// TestMoveMissileDiagonalTape: 64×32 px/tick from (64,64) — Manhattan
// distance 2 per tick, so the interpolation loop walks the in-between
// tiles in exact order ((65,64) then the destination (66,64); the corner
// probe never fires because the 32×16 px increment keeps dy%16 and dx%32
// aligned) (missiles.cpp:566-622).
func TestMoveMissileDiagonalTape(t *testing.T) {
	w := procWorld()
	m := freshMissile(64, 64)
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: diagVX, DeltaY: diagVY}
	mt := newMoveTape(w)

	tickMove(t, mt, &m, false, true,
		[]missiles.TilePosition{{X: 65, Y: 64}, {X: 66, Y: 64}},
		missiles.TilePosition{X: 66, Y: 64},
		missiles.PositionDisplacement{DeltaX: 4194304, DeltaY: 2097152},
		missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0}, [2]int{0, 0})
	t.Run("tick2", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, false, true,
			[]missiles.TilePosition{{X: 67, Y: 64}, {X: 68, Y: 64}},
			missiles.TilePosition{X: 68, Y: 64},
			missiles.PositionDisplacement{DeltaX: 8388608, DeltaY: 4194304},
			missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0}, [2]int{0, 0})
	})
}

// TestMoveMissileStopSemantics: the three collision endings — the tail
// probe revert (velocity step undone), the interpolated-tile revert (one
// increment undone), and the pass-through that keeps travelling
// (missiles.cpp:611-628).
func TestMoveMissileStopSemantics(t *testing.T) {
	tests := []struct {
		name string
		vel  missiles.PositionDisplacement
		dm   bool // ifCheckTileFailsDontMoveToTile

		wantProbes []missiles.TilePosition
		wantTile   missiles.TilePosition
		wantTrav   missiles.PositionDisplacement
		wantOff    missiles.PositionDisplacement
		wantVel    missiles.PositionDisplacement
	}{
		{
			name: "single-tile step blocked reverts the whole velocity",
			vel:  missiles.PositionDisplacement{DeltaX: 32 << 16, DeltaY: 0}, dm: true,
			wantProbes: []missiles.TilePosition{{X: 65, Y: 63}},
			wantTile:   missiles.TilePosition{X: 64, Y: 64},
			wantTrav:   missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0},
			wantOff:    missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0},
			wantVel:    missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0},
		},
		{
			name: "single-tile step blocked without dontMove lands on the probed tile",
			vel:  missiles.PositionDisplacement{DeltaX: 32 << 16, DeltaY: 0}, dm: false,
			wantProbes: []missiles.TilePosition{{X: 65, Y: 63}},
			wantTile:   missiles.TilePosition{X: 65, Y: 63},
			wantTrav:   missiles.PositionDisplacement{DeltaX: 32 << 16, DeltaY: 0},
			wantOff:    missiles.PositionDisplacement{DeltaX: -32, DeltaY: 0},
			wantVel:    missiles.PositionDisplacement{DeltaX: 32 << 16, DeltaY: 0},
		},
		{
			name: "interpolated tile blocked reverts one increment",
			vel:  missiles.PositionDisplacement{DeltaX: diagVX, DeltaY: diagVY}, dm: true,
			wantProbes: []missiles.TilePosition{{X: 65, Y: 64}},
			wantTile:   missiles.TilePosition{X: 64, Y: 64},
			wantTrav:   missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0},
			wantOff:    missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0},
			wantVel:    missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0},
		},
		{
			name: "interpolated tile blocked without dontMove lands there",
			vel:  missiles.PositionDisplacement{DeltaX: diagVX, DeltaY: diagVY}, dm: false,
			wantProbes: []missiles.TilePosition{{X: 65, Y: 64}},
			wantTile:   missiles.TilePosition{X: 65, Y: 64},
			wantTrav:   missiles.PositionDisplacement{DeltaX: 32 << 16, DeltaY: 16 << 16},
			wantOff:    missiles.PositionDisplacement{DeltaX: 0, DeltaY: 0},
			wantVel:    missiles.PositionDisplacement{DeltaX: diagVX, DeltaY: diagVY},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := procWorld()
			m := freshMissile(64, 64)
			m.Position.Velocity = tt.vel
			mt := newMoveTape(w)
			mt.probeIn = func(missiles.TilePosition) bool { return false }

			changed := missiles.MoveMissile(w, &m, mt.checkTile, tt.dm)
			if !changed {
				t.Fatal("MoveMissile changed = false, want true")
			}
			if len(mt.probes) != len(tt.wantProbes) || mt.probes[0] != tt.wantProbes[0] {
				t.Fatalf("probes = %v, want %v", mt.probes, tt.wantProbes)
			}
			if m.Position.Tile != tt.wantTile {
				t.Errorf("tile = %v, want %v", m.Position.Tile, tt.wantTile)
			}
			if m.Position.Traveled != tt.wantTrav {
				t.Errorf("traveled = %v, want %v", m.Position.Traveled, tt.wantTrav)
			}
			if m.Position.Offset != tt.wantOff {
				t.Errorf("offset = %v, want %v", m.Position.Offset, tt.wantOff)
			}
			if m.Position.Velocity != tt.wantVel {
				t.Errorf("velocity = %v, want %v", m.Position.Velocity, tt.wantVel)
			}
		})
	}
}

// TestMoveMissileAndCheckMissileColHash: the slow-missile double-hit
// guard — with the tile unchanged, CheckMissileCol reruns only when the
// dMonster^dPlayer hash differs from lastCollisionTargetHash
// (missiles.cpp:636-663).
func TestMoveMissileAndCheckMissileColHash(t *testing.T) {
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &missiles.HitWorld{Grid: world.NewGrid(), HitCallbacks: spies.callbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks()}
	w.DMonster[64][64] = 3

	m := freshMissile(64, 64)
	m.Type = missiles.Arrow
	m.Source = 0
	m.Caster = int(missiles.TargetMonsters)
	m.Dist = testDist
	m.Duration = testLifespan

	missiles.MoveMissileAndCheckMissileCol(w, &m, missiles.DamagePhysical, testMinDamage, testMaxDamage, true, false)
	if got := len(spies.monsterMHitSeen); got != 1 {
		t.Fatalf("after call 1: MonsterMHit rolls = %d, want 1", got)
	}
	if m.LastCollisionTargetHash != 3 {
		t.Fatalf("hash = %d, want 3", m.LastCollisionTargetHash)
	}

	// Same occupant: the guard suppresses the second probe.
	missiles.MoveMissileAndCheckMissileCol(w, &m, missiles.DamagePhysical, testMinDamage, testMaxDamage, true, false)
	if got := len(spies.monsterMHitSeen); got != 1 {
		t.Fatalf("after call 2: MonsterMHit rolls = %d, want 1 (guard)", got)
	}

	// A different occupant re-arms the probe.
	w.DMonster[64][64] = 5
	missiles.MoveMissileAndCheckMissileCol(w, &m, missiles.DamagePhysical, testMinDamage, testMaxDamage, true, false)
	if got := len(spies.monsterMHitSeen); got != 2 {
		t.Fatalf("after call 3: MonsterMHit rolls = %d, want 2", got)
	}
}

// TestProcessArrow: duration countdown, _midist increment, source damage
// seam pass-through (player branch), the move, deletion at 0 and the
// PutMissile tail (missiles.cpp:2932-2958). Velocity = axis tape.
func TestProcessArrow(t *testing.T) {
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &missiles.HitWorld{Grid: world.NewGrid(), HitCallbacks: spies.callbacks()}
	spies2 := &lightSpies{}
	restore2 := spies2.wire(w)
	defer restore2()

	m := freshMissile(64, 64)
	m.Type = missiles.Arrow
	m.Source = 0 // player 0
	m.Caster = int(missiles.TargetMonsters)
	m.Dist = 0
	m.Duration = 3
	// 32 px/tick axis velocity. The 32 px iso diagonal maps to the tile
	// delta (+1,-1), but screenToMissile only lands there every other
	// tick: tick 1 reaches (65,63), tick 2 stays (65,63), tick 3 reaches
	// (66,62) — the upstream tile cadence. A monster on each tile makes
	// each moving tick roll MonsterMHit once (spy misses → the arrow
	// survives); missiles.cpp:2955 roll pack.
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: 32 << 16, DeltaY: 0}
	w.DMonster[63][65] = 3
	w.DMonster[62][66] = 5

	// Tick 1: tile change to (65,63) — roll#1 carries dist 1; the alive
	// missile is registered in dFlags.
	missiles.ProcessArrow(w, &m)
	if m.Duration != 2 || m.Dist != 1 || m.DelFlag {
		t.Fatalf("tick 1: duration = %d dist = %d delFlag = %v, want 2 1 false",
			m.Duration, m.Dist, m.DelFlag)
	}
	roll := spies.monsterMHitSeen[0]
	wantRoll := rollArgs{arrowMinDam, arrowMaxDam, 1, missiles.Arrow, missiles.DamagePhysical, false, false}
	if roll != wantRoll {
		t.Fatalf("tick 1: roll = %+v, want %+v", roll, wantRoll)
	}
	if w.Grid.DFlags[63][65]&world.FlagMissile == 0 {
		t.Fatal("tick 1: FlagMissile missing at (65, 63)")
	}

	// Tick 2: no tile change — the lastCollisionTargetHash guard
	// suppresses the on-tile re-probe of monster 3; only the countdowns
	// tick.
	missiles.ProcessArrow(w, &m)
	if m.Duration != 1 || m.Dist != 2 || m.DelFlag {
		t.Fatalf("tick 2: duration = %d dist = %d delFlag = %v, want 1 2 false",
			m.Duration, m.Dist, m.DelFlag)
	}
	if rolls := len(spies.monsterMHitSeen); rolls != 1 {
		t.Fatalf("tick 2: MonsterMHit rolls = %d, want 1 (hash guard)", rolls)
	}

	// Tick 3: tile change to (66,62) — roll#2 carries dist 3; duration
	// 0 deletes the missile and the PutMissile tail early-outs.
	missiles.ProcessArrow(w, &m)
	if m.Duration != 0 || m.Dist != 3 || !m.DelFlag {
		t.Fatalf("tick 3: duration = %d dist = %d delFlag = %v, want 0 3 true",
			m.Duration, m.Dist, m.DelFlag)
	}
	roll = spies.monsterMHitSeen[1]
	wantRoll = rollArgs{arrowMinDam, arrowMaxDam, 3, missiles.Arrow, missiles.DamagePhysical, false, false}
	if roll != wantRoll {
		t.Fatalf("tick 3: roll = %+v, want %+v", roll, wantRoll)
	}
	if w.Grid.DFlags[62][66]&world.FlagMissile != 0 {
		t.Fatal("tick 3: deleted missile registered in dFlags")
	}
}

// TestProcessGenericProjectileLightTape: the alive branch tracks the tile
// into var1/var2 and repositions the light; the hit branch deletes,
// spawns the explosion via the seam, removes the light and skips the
// dFlags write (missiles.cpp:2963-3016). Movement = diagonal tape.
func TestProcessGenericProjectileLightTape(t *testing.T) { //nolint:revive // tick-tape assertions are one upstream sequence (missiles.cpp:2963-3016)
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &missiles.HitWorld{Grid: world.NewGrid(), HitCallbacks: spies.callbacks()}
	light := &lightSpies{}
	restoreLight := light.wire(w)
	defer restoreLight()

	m := freshMissile(64, 64)
	m.Type = missiles.Firebolt
	m.Source = 0
	m.Caster = int(missiles.TargetMonsters)
	m.Dam = 12
	m.Lid = 7
	m.Duration = 2
	m.MFNum = int(missiles.South)
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: diagVX, DeltaY: diagVY}

	// Tick 1: alive — var1/var2 track the tile, ChangeLight(tile, 8).
	missiles.ProcessGenericProjectile(w, &m)
	if m.Duration != 1 {
		t.Fatalf("tick 1: duration = %d, want 1", m.Duration)
	}
	if m.Var1 != 66 || m.Var2 != 64 {
		t.Fatalf("tick 1: var = (%d, %d), want (66, 64)", m.Var1, m.Var2)
	}
	if len(light.changes) != 1 || light.changes[0] != [4]int{7, 66, 64, 8} {
		t.Fatalf("tick 1: ChangeLight = %v, want [7 66 64 8]", light.changes)
	}
	if got := w.Grid.DFlags[64][66] & world.FlagMissile; got == 0 {
		t.Fatal("tick 1: FlagMissile missing at (66, 64)")
	}
	if len(light.unlights) != 0 || len(light.spawnCalls) != 0 {
		t.Fatalf("tick 1: unlights = %v spawns = %v, want none", light.unlights, light.spawnCalls)
	}

	// Tick 2: hit — delete, spawn MagmaBallExplosion at the tile, unlight.
	missiles.ProcessGenericProjectile(w, &m)
	if m.Duration != 0 || !m.DelFlag {
		t.Fatalf("tick 2: duration = %d delFlag = %v, want 0 true", m.Duration, m.DelFlag)
	}
	if len(light.spawnCalls) != 1 {
		t.Fatalf("tick 2: spawns = %d, want 1", len(light.spawnCalls))
	}
	call := light.spawnCalls[0]
	if call.src != (missiles.TilePosition{X: 68, Y: 64}) ||
		call.dst != (missiles.TilePosition{}) ||
		call.dir != missiles.South ||
		call.spawnID != missiles.MagmaBallExplosion ||
		call.caster != int(missiles.TargetMonsters) || call.source != 0 ||
		call.parent != &m {
		t.Fatalf("tick 2: spawn = %+v, want MagmaBallExplosion at (68, 64) dir South parent self", call)
	}
	if len(light.unlights) != 1 || light.unlights[0] != 7 {
		t.Fatalf("tick 2: unlights = %v, want [7]", light.unlights)
	}
	if w.Grid.DFlags[64][68]&world.FlagMissile != 0 {
		t.Fatal("tick 2: deleted missile registered in dFlags")
	}
	if len(light.changes) != 1 {
		t.Fatalf("tick 2: ChangeLight calls = %d, want 1 (alive branch only)", len(light.changes))
	}
}

// TestProcessGenericProjectileSpawnIDSwitch: the explosion dispatch
// (missiles.cpp:2982-3001) — base-table rows only; hellfire flare rows
// land with the hellfire slice. NoLight missiles touch no light hook.
func TestProcessGenericProjectileSpawnIDSwitch(t *testing.T) {
	tests := []struct {
		name          string
		typ           missiles.MissileID
		wantExplosion missiles.MissileID
	}{
		{"firebolt", missiles.Firebolt, missiles.MagmaBallExplosion},
		{"magmaball", missiles.MagmaBall, missiles.MagmaBallExplosion},
		{"bloodstar", missiles.BloodStar, missiles.BloodStarExplosion},
		{"acid", missiles.Acid, missiles.AcidSplat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spies := newColSpies() // monster-target gate open (collision_test.go)
			w := &missiles.HitWorld{Grid: world.NewGrid(), HitCallbacks: spies.callbacks()}
			light := &lightSpies{}
			restore := light.wire(w)
			defer restore()

			m := freshMissile(64, 64)
			m.Type = tt.typ
			m.Caster = int(missiles.TargetMonsters)
			m.Source = 2
			m.Duration = 1
			m.Lid = missiles.NoLight // no light plumbing on this path
			m.Position.Velocity = missiles.PositionDisplacement{DeltaX: axisVX, DeltaY: axisVY}

			missiles.ProcessGenericProjectile(w, &m)

			if !m.DelFlag {
				t.Fatal("DelFlag not set")
			}
			if len(light.spawnCalls) != 1 {
				t.Fatalf("spawns = %d, want 1", len(light.spawnCalls))
			}
			if got := light.spawnCalls[0].spawnID; got != tt.wantExplosion {
				t.Fatalf("spawnID = %v, want %v", got, tt.wantExplosion)
			}
			if len(light.unlights) != 0 || len(light.changes) != 0 {
				t.Fatalf("unlights = %v changes = %v, want none with NoLight",
					light.unlights, light.changes)
			}
			// ChangeLightOffset's own NO_LIGHT guard (lighting.cpp:332)
			// makes the UpdateMissilePos feed harmless — the hook fires,
			// but never with a real light id.
			for _, off := range light.offsets {
				if off[0] != missiles.NoLight {
					t.Fatalf("ChangeLightOffset lid = %d, want NoLight", off[0])
				}
			}
		})
	}
}
