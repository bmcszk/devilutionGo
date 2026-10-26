// Package missiles_test: devilGo-mwlm slice 3a-beta acceptance tests —
// the motion primitives and the missiles.ProcessArrow/missiles.ProcessGenericProjectile
// ports (Source/missiles.cpp:173-249, 552-663, 2932-3016). Every expected
// number below is upstream-exact: each literal is hand-transcribed from the
// cited upstream arithmetic (displacement.hpp / missiles.cpp formulas, cited
// at each derivation), not from the port's output.
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
//     missiles.MoveMissile's interpolated-tile loop runs; the increment is exactly
//     32×16 px (the iso tile diagonal, (32<<16)/denominator with
//     denominator = 2*speed.dy = speed.dx).
const (
	axisVX = 11 << 16
	axisVY = 0
	diagVX = 64 << 16
	diagVY = 32 << 16
)

// Scripted missiles.Arrow damage seam values (SourceDamage stand-in for the
// player _pIMinDam/_pIMaxDam read, cpp:2938-2942).
const (
	arrowMinDam = 3
	arrowMaxDam = 9
)

// lightSpies records every light-hook dispatch (ChangeLightOffset,
// lighting.h:64; ChangeLight, lighting.h:65; AddUnLight, lighting.h:61).
type lightSpies struct {
	offsets    []lightOffsetCall // ChangeLightOffset (lighting.h:64)
	changes    []changeLightCall // ChangeLight (lighting.h:65)
	unlights   []int             // AddUnLight (lid)
	spawnCalls []spawnCall
	sourceDam  []bool
}

// changeLightCall and lightOffsetCall record one dispatch each
// (lighting.h:64/65); changeLightCall is shared with process_ii_test.go.
type lightOffsetCall struct {
	lid, offsetX, offsetY int
}

// spawnCall records one SpawnMissile dispatch (the AddMissile tail of
// the explosion switch, cpp:2982-3001).
type spawnCall struct {
	src, dst missiles.TilePosition
	dir      missiles.Direction
	spawnID  missiles.MissileID
	caster   int
	source   int
	parent   *missiles.Missile
}

type wantPutMissile struct {
	flags   world.DungeonFlag
	delFlag bool
	preFlag bool
}

type putMissileRow = struct {
	name    string
	typ     missiles.MissileID
	tile    missiles.TilePosition
	delFlag bool
	preFlag bool
	want    wantPutMissile
}

type updateMissilePosRow = struct {
	name        string
	traveled    missiles.PositionDisplacement
	wantTile    missiles.TilePosition
	wantOffset  missiles.PositionDisplacement
	wantLightXY missiles.TilePosition
}

type moveMissilePosRow = struct {
	name       string
	dir        missiles.Direction
	available  bool
	wantCalled bool
	wantDelta  missiles.TilePosition
	wantOff    missiles.PositionDisplacement
}

// moveTape drives missiles.MoveMissile tick by tick and records what the checkTile
// spy saw, so the table rows below assert exact visit order.
type moveTape struct {
	world   *missiles.HitWorld
	probeIn func(tile missiles.TilePosition) bool // scripted checkTile outcome
	probes  []missiles.TilePosition
}

// moveExpectation bundles the post-tick state assertions of tickMove.
type moveExpectation struct {
	changed          bool
	probes           []missiles.TilePosition
	tile             missiles.TilePosition
	traveled, offset missiles.PositionDisplacement
	light            missiles.TilePosition
}

type moveMissileStopRow = struct {
	name string
	vel  missiles.PositionDisplacement
	dm   bool // ifCheckTileFailsDontMoveToTile

	wantProbes []missiles.TilePosition
	wantTile   missiles.TilePosition
	wantTrav   missiles.PositionDisplacement
	wantOff    missiles.PositionDisplacement
	wantVel    missiles.PositionDisplacement
}

type projSpawnSwitchRow = struct {
	name          string
	typ           missiles.MissileID
	wantExplosion missiles.MissileID
}

// TestUpdateMissileVelocity: velocity = missiles.WorldToNormalScreen(tile -
// destination) * velocityInPixels (missiles.cpp:173-186); same-tile
// short-circuits to {0,0}. Literals derived from the upstream formulas
// (displacement.hpp:151-199, 363-371):
//
//	rotated = {dy-dx, -(dy+dx)}; magnitude = sqrtf(dx*dx+dy*dy);
//	normalized = (rotated<<16) / magnitude (C++ float div, trunc toward 0);
//	missiles.WorldToNormalScreen = {normalized.x, normalized.y/2}; then * 16 px.
//
// e.g. (3,0): rotated = {-3,-3}, mag = sqrt(18), -3<<16/4.2426 = -46340 (trunc),
// y/2 = -23170; velocity = {-46340*16, -23170*16} = {-741440, -370720}.
func TestUpdateMissileVelocity(t *testing.T) {
	// given
	tests := []struct {
		name        string
		tile        missiles.TilePosition
		destination missiles.TilePosition
		pixels      int
		want        missiles.PositionDisplacement
	}{
		{
			"same tile zeroes velocity",
			missiles.TilePosition{X: 64, Y: 64},
			missiles.TilePosition{X: 64, Y: 64},
			16,
			missiles.PositionDisplacement{},
		},
		{
			"tile east of destination",
			missiles.TilePosition{X: 67, Y: 64},
			missiles.TilePosition{X: 64, Y: 64},
			16,
			missiles.PositionDisplacement{DeltaX: -741440, DeltaY: -370720},
		},
		{
			"tile south of destination",
			missiles.TilePosition{X: 64, Y: 67},
			missiles.TilePosition{X: 64, Y: 64},
			16,
			missiles.PositionDisplacement{DeltaX: 741440, DeltaY: -370720},
		},
		{
			"tile north of destination",
			missiles.TilePosition{X: 64, Y: 61},
			missiles.TilePosition{X: 64, Y: 64},
			16,
			missiles.PositionDisplacement{DeltaX: -741440, DeltaY: 370720},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			m := freshMissile(tt.tile.X, tt.tile.Y)

			// when
			missiles.UpdateMissileVelocity(&m, tt.destination, tt.pixels)

			// then
			if m.Position.Velocity != tt.want {
				t.Fatalf("velocity = %v, want %v", m.Position.Velocity, tt.want)
			}
		})
	}
}

func TestStopMissile(t *testing.T) {
	t.Run("renderer on tile snaps offset", func(t *testing.T) {
		// given
		m := freshMissile(64, 64)
		m.Position.Velocity = missiles.PositionDisplacement{DeltaX: 5, DeltaY: 5}
		m.Position.Offset = missiles.PositionDisplacement{DeltaX: 100, DeltaY: 200}
		m.Position.TileForRendering = m.Position.Tile
		m.Position.OffsetForRendering = missiles.PositionDisplacement{DeltaX: 7, DeltaY: 9}

		// when
		missiles.StopMissile(&m)

		// then
		if m.Position.Velocity != (missiles.PositionDisplacement{}) {
			t.Fatalf("velocity = %v, want zero", m.Position.Velocity)
		}

		if m.Position.Offset != (missiles.PositionDisplacement{DeltaX: 7, DeltaY: 9}) {
			t.Fatalf("offset = %v, want renderer offset {7 9}", m.Position.Offset)
		}
	})
	t.Run("renderer elsewhere keeps offset", func(t *testing.T) {
		// given
		m := freshMissile(64, 64)
		m.Position.Velocity = missiles.PositionDisplacement{DeltaX: 5, DeltaY: 5}
		m.Position.Offset = missiles.PositionDisplacement{DeltaX: 100, DeltaY: 200}
		m.Position.TileForRendering = missiles.TilePosition{X: 1, Y: 1}
		m.Position.OffsetForRendering = missiles.PositionDisplacement{DeltaX: 7, DeltaY: 9}

		// when
		missiles.StopMissile(&m)

		// then
		if m.Position.Offset != (missiles.PositionDisplacement{DeltaX: 100, DeltaY: 200}) {
			t.Fatalf("offset = %v, want {100 200}", m.Position.Offset)
		}
	})
}

// TestPutMissile: the dFlags registration (cpp:191-211) —
// missiles.Missile bit always, MissileFireWall/MissileLightningWall per type,
// DelFlag short-circuit (including out-of-bounds raising it), and the
// MissilePreFlag global.
func TestPutMissile(t *testing.T) {
	// given
	tests := []putMissileRow{
		{
			"arrow sets missile bit", missiles.Arrow,
			missiles.TilePosition{X: 64, Y: 64},
			false, false,
			wantPutMissile{flags: world.FlagMissile},
		},
		{
			"firewall adds fire bit", missiles.FireWall,
			missiles.TilePosition{X: 64, Y: 64},
			false, false,
			wantPutMissile{flags: world.FlagMissile | world.FlagMissileFire},
		},
		{
			"lightningwall adds lightning bit", missiles.LightningWall,
			missiles.TilePosition{X: 64, Y: 64},
			false, false,
			wantPutMissile{flags: world.FlagMissile | world.FlagMissileLight},
		},
		{
			"pre-flagged missile raises the global", missiles.Arrow,
			missiles.TilePosition{X: 64, Y: 64},
			false, true,
			wantPutMissile{flags: world.FlagMissile, preFlag: true},
		},
		{
			"del flag skips registration", missiles.Arrow,
			missiles.TilePosition{X: 64, Y: 64},
			true, false,
			wantPutMissile{delFlag: true},
		},
		{
			"out of bounds raises del flag", missiles.Arrow,
			missiles.TilePosition{X: -1, Y: 64},
			false, false,
			wantPutMissile{delFlag: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			putMissileCase(t, tt)
		})
	}
}

// TestUpdateMissilePos: traveled>>16 → missiles.ScreenToMissile → tile = start +
// tileOffset; offset = pixels + missiles.WorldToScreen(tileOffset); light offset =
// screenToLight(pixels) - tileOffset*8 (missiles.cpp:213-224). Literals
// derived from the upstream matrix formulas (displacement.hpp:151-190):
//
//	missiles.WorldToScreen(d)   = {32*(dy-dx), -16*(dy+dx)}
//	missiles.ScreenToMissile(d) = {(2*dy+dx±32)/64, (2*dy-dx±32)/64} (sign picks
//	  rounding offset; C++ int division truncates toward zero)
//	screenToLight(d)   = {(2*dy+dx)/8, (2*dy-dx)/8}
//
// e.g. pixels (-11,13): missiles.ScreenToMissile = {(15+32)/64, (37+32)/64} = (0,1);
// offset = (-11,13) + missiles.WorldToScreen(0,1)=(32,-16) = (21,-3);
// light = (15/8, 37/8) = (1,4) - (0,1)*8 = (1,-4).
func TestUpdateMissilePos(t *testing.T) {
	// given
	tests := []updateMissilePosRow{
		{
			"fractional diagonal floors the pixel vector",
			missiles.PositionDisplacement{DeltaX: -11 << 16, DeltaY: 13<<16 + 32768},
			missiles.TilePosition{X: 0, Y: 1},
			missiles.PositionDisplacement{DeltaX: 21, DeltaY: -3},
			missiles.TilePosition{X: 1, Y: -4},
		},
		{
			"negative pixel vector wraps through iso space",
			missiles.PositionDisplacement{DeltaX: -32 << 16, DeltaY: -8 << 16},
			missiles.TilePosition{X: -1, Y: 0},
			missiles.PositionDisplacement{DeltaX: 0, DeltaY: 8},
			missiles.TilePosition{X: 2, Y: 2},
		},
		{
			"whole tiles inside the same tile keep offset small",
			missiles.PositionDisplacement{DeltaX: 3 << 16, DeltaY: 0},
			missiles.TilePosition{X: 0, Y: 0},
			missiles.PositionDisplacement{DeltaX: 3, DeltaY: 0},
			missiles.TilePosition{X: 0, Y: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateMissilePosCase(t, tt)
		})
	}
}

// TestMoveMissilePos: the charging-monster correction (cpp:233-251)
// — missiles.East→SE, missiles.West→SW, missiles.South/SW/SE→missiles.South; the tile moves by the direction
// delta and the offset by missiles.WorldToScreen(delta). The IsTileAvailable gate
// is the injected seam.
func TestMoveMissilePos(t *testing.T) {
	// given
	tests := []moveMissilePosRow{
		{
			"east shifts south-east", missiles.East, true, true,
			missiles.TilePosition{X: 1, Y: 0},
			missiles.PositionDisplacement{DeltaX: -32, DeltaY: -16},
		},
		{
			"west shifts south-west", missiles.West, true, true,
			missiles.TilePosition{X: 0, Y: 1},
			missiles.PositionDisplacement{DeltaX: 32, DeltaY: -16},
		},
		{
			"south shifts south", missiles.South, true, true,
			missiles.TilePosition{X: 1, Y: 1},
			missiles.PositionDisplacement{DeltaX: 0, DeltaY: -32},
		},
		{
			"north is untouched", missiles.North, true, false,
			missiles.TilePosition{X: 0, Y: 0},
			missiles.PositionDisplacement{},
		},
		{
			"unavailable tile is untouched", missiles.East, false, true,
			missiles.TilePosition{X: 0, Y: 0},
			missiles.PositionDisplacement{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moveMissilePosCase(t, tt)
		})
	}
}

// TestMoveMissileAxisTape: 11 px/tick screen-x travel from (64,64),
// open floor — ticks 0-1 stay inside the start tile (possibleVisitTiles
// = 0 → false, no probes), tick 2 lands on (65,63) and probes exactly
// the destination tile (cpp:552-628).
func TestMoveMissileAxisTape(t *testing.T) {
	// given
	w := procWorld()
	m := freshMissile(64, 64)
	m.Lid = 7
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: axisVX, DeltaY: axisVY}
	mt := newMoveTape(w)

	// when/then — one missiles.MoveMissile tick with its full post-state
	tickMove(t, mt, &m, moveExpectation{
		changed:  false,
		probes:   nil,
		tile:     missiles.TilePosition{X: 64, Y: 64},
		traveled: missiles.PositionDisplacement{DeltaX: 720896, DeltaY: 0},
		offset:   missiles.PositionDisplacement{DeltaX: 11, DeltaY: 0},
		light:    missiles.TilePosition{X: 1, Y: -1},
	})
	t.Run("tick2", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, moveExpectation{
			changed:  false,
			probes:   nil,
			tile:     missiles.TilePosition{X: 64, Y: 64},
			traveled: missiles.PositionDisplacement{DeltaX: 1441792, DeltaY: 0},
			offset:   missiles.PositionDisplacement{DeltaX: 22, DeltaY: 0},
			light:    missiles.TilePosition{X: 2, Y: -2},
		})
	})
	t.Run("tick3", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, moveExpectation{
			changed:  true,
			probes:   []missiles.TilePosition{{X: 65, Y: 63}},
			tile:     missiles.TilePosition{X: 65, Y: 63},
			traveled: missiles.PositionDisplacement{DeltaX: 2162688, DeltaY: 0},
			offset:   missiles.PositionDisplacement{DeltaX: -31, DeltaY: 0},
			light:    missiles.TilePosition{X: -4, Y: 4},
		})
	})
}

// TestMoveMissileDiagonalTape: 64×32 px/tick from (64,64) — Manhattan
// distance 2 per tick, so the interpolation loop walks the in-between
// tiles in exact order ((65,64) then the destination (66,64); the corner
// probe never fires because the 32×16 px increment keeps dy%16 and dx%32
// aligned) (cpp:566-622).
func TestMoveMissileDiagonalTape(t *testing.T) {
	// given
	w := procWorld()
	m := freshMissile(64, 64)
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: diagVX, DeltaY: diagVY}
	mt := newMoveTape(w)

	tickMove(t, mt, &m, moveExpectation{
		changed:  true,
		probes:   []missiles.TilePosition{{X: 65, Y: 64}, {X: 66, Y: 64}},
		tile:     missiles.TilePosition{X: 66, Y: 64},
		traveled: missiles.PositionDisplacement{DeltaX: 4194304, DeltaY: 2097152},
		light:    missiles.TilePosition{},
	})
	t.Run("tick2", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, moveExpectation{
			changed:  true,
			probes:   []missiles.TilePosition{{X: 67, Y: 64}, {X: 68, Y: 64}},
			tile:     missiles.TilePosition{X: 68, Y: 64},
			traveled: missiles.PositionDisplacement{DeltaX: 8388608, DeltaY: 4194304},
			light:    missiles.TilePosition{},
		})
	})
}

// TestMoveMissileStopSemantics pins the upstream stop-on-collision
// semantics (missiles.cpp:611-628).
func TestMoveMissileStopSemantics(t *testing.T) {
	// given
	tests := []moveMissileStopRow{
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
			moveMissileStopCase(t, tt)
		})
	}
}

// TestMoveMissileAndCheckMissileColHash: the slow-missile double-hit
// guard — with the tile unchanged, missiles.CheckMissileCol reruns only when the
// dMonster^dPlayer hash differs from lastCollisionTargetHash
// (cpp:636-663).
func TestMoveMissileAndCheckMissileColHash(t *testing.T) {
	// given
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &missiles.HitWorld{
		Grid: world.NewGrid(), HitCallbacks: spies.callbacks(),
		ProcessCallbacks: missiles.NoopProcessCallbacks(),
	}
	w.DMonster[64][64] = 3

	m := freshMissile(64, 64)
	m.Type = missiles.Arrow
	m.Source = 0
	m.Caster = int(missiles.TargetMonsters)
	m.Dist = testDist
	m.Duration = testLifespan

	// when/then — three missiles.MoveMissileAndCheckMissileCol rounds over the
	// changing dMonster occupant
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
// missiles.PutMissile tail (cpp:2932-2958). Velocity = axis tape.
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
	// delta (+1,-1), but missiles.ScreenToMissile only lands there every other
	// tick: tick 1 reaches (65,63), tick 2 stays (65,63), tick 3 reaches
	// (66,62) — the upstream tile cadence. A monster on each tile makes
	// each moving tick roll MonsterMHit once (spy misses → the arrow
	// survives); cpp:2955 roll pack.
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: 32 << 16, DeltaY: 0}
	w.DMonster[63][65] = 3
	w.DMonster[62][66] = 5

	// Tick 1: tile change to (65,63) — roll#1 carries dist 1; the alive
	// missile is registered in dFlags.
	// when
	missiles.ProcessArrow(w, &m)

	// then
	arrowTick1(t, spies, w, m)
	// Tick 2: no tile change — the lastCollisionTargetHash guard
	// suppresses the on-tile re-probe of monster 3; only the countdowns
	// tick.
	// when
	missiles.ProcessArrow(w, &m)

	// then
	arrowTick2(t, spies, m)

	// Tick 3: tile change to (66,62) — roll#2 carries dist 3; duration
	// 0 deletes the missile and the missiles.PutMissile tail early-outs.
	// when
	missiles.ProcessArrow(w, &m)

	// then
	arrowTick3(t, spies, w, m)
}

// TestProcessGenericProjectileLightTape: the alive branch tracks the tile
// into var1/var2 and repositions the light; the hit branch deletes,
// spawns the explosion via the seam, removes the light and skips the
// dFlags write (cpp:2963-3016). Movement = diagonal tape; the assertions
// stay inline so the tape reads in upstream order.
func TestProcessGenericProjectileLightTape(t *testing.T) {
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
	// when
	missiles.ProcessGenericProjectile(w, &m)

	// then
	projTapeTick1(t, light, w, m)
	// Tick 2: hit — delete, spawn missiles.MagmaBallExplosion at the tile, unlight.
	// when
	missiles.ProcessGenericProjectile(w, &m)

	// then
	projTapeTick2(t, light, w, &m)
}

// TestProcessGenericProjectileSpawnIDSwitch: the explosion dispatch
// (cpp:2982-3001) — base-table rows only; hellfire flare rows
// land with the hellfire slice. missiles.NoLight missiles.missiles touch no light hook.
func TestProcessGenericProjectileSpawnIDSwitch(t *testing.T) {
	// given
	tests := []projSpawnSwitchRow{
		{"firebolt", missiles.Firebolt, missiles.MagmaBallExplosion},
		{"magmaball", missiles.MagmaBall, missiles.MagmaBallExplosion},
		{"bloodstar", missiles.BloodStar, missiles.BloodStarExplosion},
		{"acid", missiles.Acid, missiles.AcidSplat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projSpawnSwitchCase(t, tt)
		})
	}
}

func newMoveTape(w *missiles.HitWorld) *moveTape { return &moveTape{world: w} }

func (mt *moveTape) checkTile(tile missiles.TilePosition) bool {
	mt.probes = append(mt.probes, tile)
	if mt.probeIn != nil {
		return mt.probeIn(tile)
	}

	return true
}

func putMissileCase(t *testing.T, tt putMissileRow) {
	t.Helper()
	// given
	w := &missiles.HitWorld{Grid: world.NewGrid()}

	m := freshMissile(tt.tile.X, tt.tile.Y)
	m.Type = tt.typ
	m.DelFlag = tt.delFlag
	m.PreFlag = tt.preFlag

	// when
	missiles.PutMissile(w, &m)

	// then
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
}

func updateMissilePosCase(t *testing.T, tt updateMissilePosRow) {
	t.Helper()
	// given
	spies := &lightSpies{}
	w := &missiles.HitWorld{}

	restore := spies.wire(w)
	defer restore()

	m := freshMissile(0, 0)
	m.Lid = 3
	m.Position.Traveled = tt.traveled

	// when
	missiles.UpdateMissilePos(w, &m)

	// then
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
	if got.lid != 3 || got.offsetX != tt.wantLightXY.X || got.offsetY != tt.wantLightXY.Y {
		t.Errorf("ChangeLightOffset = %+v, want lid 3 offset %v", got, tt.wantLightXY)
	}
}

func moveMissilePosCase(t *testing.T, tt moveMissilePosRow) {
	t.Helper()
	// given
	called := 0
	w := &missiles.HitWorld{}
	cb := missiles.NoopProcessCallbacks()
	cb.IsTileAvailable = func(_, _, _ int) bool {
		called++
		return tt.available
	}
	w.ProcessCallbacks = cb

	m := freshMissile(64, 64)
	m.MFNum = int(tt.dir)

	// when
	missiles.MoveMissilePos(w, &m)

	// then
	if called != expectedProbeCount(tt) {
		t.Fatalf("IsTileAvailable calls = %d, want %d", called, expectedProbeCount(tt))
	}

	if m.Position.Tile != (missiles.TilePosition{X: 64 + tt.wantDelta.X, Y: 64 + tt.wantDelta.Y}) {
		t.Fatalf("tile = %v, want %+v", m.Position.Tile, tt.wantDelta)
	}

	if m.Position.Offset != tt.wantOff {
		t.Fatalf("offset = %v, want %v", m.Position.Offset, tt.wantOff)
	}
}

func expectedProbeCount(tt moveMissilePosRow) int {
	if tt.wantCalled {
		return 1
	}

	return 0
}

// tickMove runs one missiles.MoveMissile tick and asserts the full post-state
// (wantChanged is upstream's return: false only for the no-motion tick,
// cpp:565).
func tickMove(t *testing.T, mt *moveTape, m *missiles.Missile, want moveExpectation) {
	t.Helper()

	wantChanged := want.changed

	spies := &lightSpies{}

	restore := spies.wire(mt.world)
	defer restore()

	changed := missiles.MoveMissile(mt.world, m, mt.checkTile, false)
	if changed != wantChanged {
		t.Fatalf("moveMissile changed = %v, want %v", changed, wantChanged)
	}

	expectProbes(t, mt.probes, want.probes)
	expectMoveState(t, m, want.tile, want.traveled, want.offset)

	if len(spies.offsets) != 1 {
		t.Fatalf("ChangeLightOffset calls = %d, want 1", len(spies.offsets))
	}

	expectLightOffset(t, spies.offsets, want.light)
}

func expectMoveState(t *testing.T, m *missiles.Missile, wantTile missiles.TilePosition,
	wantTraveled, wantOffset missiles.PositionDisplacement,
) {
	t.Helper()

	if m.Position.Tile != wantTile {
		t.Fatalf("tile = %v, want %v", m.Position.Tile, wantTile)
	}

	if m.Position.Traveled != wantTraveled {
		t.Fatalf("traveled = %v, want %v", m.Position.Traveled, wantTraveled)
	}

	if m.Position.Offset != wantOffset {
		t.Fatalf("offset = %v, want %v", m.Position.Offset, wantOffset)
	}
}

func expectProbes(t *testing.T, got, want []missiles.TilePosition) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("probes = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("probe[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func expectLightOffset(t *testing.T, got []lightOffsetCall, want missiles.TilePosition) {
	t.Helper()

	if len(got) != 1 {
		t.Fatalf("ChangeLightOffset calls = %d, want 1", len(got))
	}

	if want != (missiles.TilePosition{}) {
		if got[0].offsetX != want.X || got[0].offsetY != want.Y {
			t.Fatalf("light offset = (%d, %d), want %v", got[0].offsetX, got[0].offsetY, want)
		}
	}
}
