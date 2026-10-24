// Package missiles_test: devilGo-mwlm slice 3a-beta acceptance tests —
// the motion primitives and the ProcessArrow/ProcessGenericProjectile
// ports (Source/missiles.cpp:173-249, 552-663, 2932-3016). Every expected
// number below is upstream-exact: each literal is hand-transcribed from the
// cited upstream arithmetic (displacement.hpp / missiles.cpp formulas, cited
// at each derivation), not from the port's output.
//
// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package missiles

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// Scripted 16.16 velocities for the tapes:
//   - axis 11 px/tick (screen-x direction) — crosses tiles every ~6 ticks,
//     first tile change on tick 3 (tape below);
//   - diagonal 64×32 px/tick — Manhattan distance 2 per tick, so
//     moveMissile's interpolated-tile loop runs; the increment is exactly
//     32×16 px (the iso tile diagonal, (32<<16)/denominator with
//     denominator = 2*speed.dy = speed.dx).
const (
	axisVX = 11 << 16
	axisVY = 0
	diagVX = 64 << 16
	diagVY = 32 << 16
)

// Scripted Arrow damage seam values (SourceDamage stand-in for the
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
	src, dst tilePosition
	dir      Direction
	spawnID  missileID
	caster   int
	source   int
	parent   *Missile
}

// TestUpdateMissileVelocity: velocity = worldToNormalScreen(tile -
// destination) * velocityInPixels (missiles.cpp:173-186); same-tile
// short-circuits to {0,0}. Literals derived from the upstream formulas
// (displacement.hpp:151-199, 363-371):
//
//	rotated = {dy-dx, -(dy+dx)}; magnitude = sqrtf(dx*dx+dy*dy);
//	normalized = (rotated<<16) / magnitude (C++ float div, trunc toward 0);
//	worldToNormalScreen = {normalized.x, normalized.y/2}; then * 16 px.
//
// e.g. (3,0): rotated = {-3,-3}, mag = sqrt(18), -3<<16/4.2426 = -46340 (trunc),
// y/2 = -23170; velocity = {-46340*16, -23170*16} = {-741440, -370720}.
func TestUpdateMissileVelocity(t *testing.T) {
	// given
	tests := []struct {
		name        string
		tile        tilePosition
		destination tilePosition
		pixels      int
		want        positionDisplacement
	}{
		{"same tile zeroes velocity", tilePosition{X: 64, Y: 64},
			tilePosition{X: 64, Y: 64}, 16, positionDisplacement{}},
		{"tile east of destination", tilePosition{X: 67, Y: 64},
			tilePosition{X: 64, Y: 64}, 16,
			positionDisplacement{DeltaX: -741440, DeltaY: -370720}},
		{"tile south of destination", tilePosition{X: 64, Y: 67},
			tilePosition{X: 64, Y: 64}, 16,
			positionDisplacement{DeltaX: 741440, DeltaY: -370720}},
		{"tile north of destination", tilePosition{X: 64, Y: 61},
			tilePosition{X: 64, Y: 64}, 16,
			positionDisplacement{DeltaX: -741440, DeltaY: 370720}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			m := freshMissile(tt.tile.X, tt.tile.Y)

			// when
			updateMissileVelocity(&m, tt.destination, tt.pixels)

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
		m.Position.Velocity = positionDisplacement{DeltaX: 5, DeltaY: 5}
		m.Position.Offset = positionDisplacement{DeltaX: 100, DeltaY: 200}
		m.Position.TileForRendering = m.Position.Tile
		m.Position.OffsetForRendering = positionDisplacement{DeltaX: 7, DeltaY: 9}

		// when
		stopMissile(&m)

		// then
		if m.Position.Velocity != (positionDisplacement{}) {
			t.Fatalf("velocity = %v, want zero", m.Position.Velocity)
		}
		if m.Position.Offset != (positionDisplacement{DeltaX: 7, DeltaY: 9}) {
			t.Fatalf("offset = %v, want renderer offset {7 9}", m.Position.Offset)
		}
	})
	t.Run("renderer elsewhere keeps offset", func(t *testing.T) {
		// given
		m := freshMissile(64, 64)
		m.Position.Velocity = positionDisplacement{DeltaX: 5, DeltaY: 5}
		m.Position.Offset = positionDisplacement{DeltaX: 100, DeltaY: 200}
		m.Position.TileForRendering = tilePosition{X: 1, Y: 1}
		m.Position.OffsetForRendering = positionDisplacement{DeltaX: 7, DeltaY: 9}

		// when
		stopMissile(&m)

		// then
		if m.Position.Offset != (positionDisplacement{DeltaX: 100, DeltaY: 200}) {
			t.Fatalf("offset = %v, want {100 200}", m.Position.Offset)
		}
	})
}

// TestPutMissile: the dFlags registration (cpp:191-211) —
// Missile bit always, MissileFireWall/MissileLightningWall per type,
// DelFlag short-circuit (including out-of-bounds raising it), and the
// MissilePreFlag global.
type wantPutMissile struct {
	flags   world.DungeonFlag
	delFlag bool
	preFlag bool
}

type putMissileRow = struct {
	name    string
	typ     missileID
	tile    tilePosition
	delFlag bool
	preFlag bool
	want    wantPutMissile
}

func TestPutMissile(t *testing.T) {
	// given
	tests := []putMissileRow{
		{"arrow sets missile bit", Arrow,
			tilePosition{X: 64, Y: 64}, false, false,
			wantPutMissile{flags: world.FlagMissile}},
		{"firewall adds fire bit", FireWall,
			tilePosition{X: 64, Y: 64}, false, false,
			wantPutMissile{flags: world.FlagMissile | world.FlagMissileFire}},
		{"lightningwall adds lightning bit", LightningWall,
			tilePosition{X: 64, Y: 64}, false, false,
			wantPutMissile{flags: world.FlagMissile | world.FlagMissileLight}},
		{"pre-flagged missile raises the global", Arrow,
			tilePosition{X: 64, Y: 64}, false, true,
			wantPutMissile{flags: world.FlagMissile, preFlag: true}},
		{"del flag skips registration", Arrow,
			tilePosition{X: 64, Y: 64}, true, false,
			wantPutMissile{delFlag: true}},
		{"out of bounds raises del flag", Arrow,
			tilePosition{X: -1, Y: 64}, false, false,
			wantPutMissile{delFlag: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			putMissileCase(t, tt)
		})
	}
}

func putMissileCase(t *testing.T, tt putMissileRow) {
	t.Helper()
	// given
	w := &hitWorld{Grid: world.NewGrid()}

	m := freshMissile(tt.tile.X, tt.tile.Y)
	m.Type = tt.typ
	m.DelFlag = tt.delFlag
	m.PreFlag = tt.preFlag

	// when
	PutMissile(w, &m)

	// then
	if m.DelFlag != tt.want.delFlag {
		t.Fatalf("DelFlag = %v, want %v", m.DelFlag, tt.want.delFlag)
	}
	if w.processCallbacks.MissilePreFlag != tt.want.preFlag {
		t.Fatalf("MissilePreFlag = %v, want %v",
			w.processCallbacks.MissilePreFlag, tt.want.preFlag)
	}
	if !world.InDungeonBounds(tt.tile.X, tt.tile.Y) || tt.want.delFlag {
		return // nothing may be written
	}
	got := w.Grid.DFlags[tt.tile.Y][tt.tile.X]
	if got != tt.want.flags {
		t.Fatalf("dFlags = %v, want %v", got, tt.want.flags)
	}
}

// TestUpdateMissilePos: traveled>>16 → screenToMissile → tile = start +
// tileOffset; offset = pixels + worldToScreen(tileOffset); light offset =
// screenToLight(pixels) - tileOffset*8 (missiles.cpp:213-224). Literals
// derived from the upstream matrix formulas (displacement.hpp:151-190):
//
//	worldToScreen(d)   = {32*(dy-dx), -16*(dy+dx)}
//	screenToMissile(d) = {(2*dy+dx±32)/64, (2*dy-dx±32)/64} (sign picks
//	  rounding offset; C++ int division truncates toward zero)
//	screenToLight(d)   = {(2*dy+dx)/8, (2*dy-dx)/8}
//
// e.g. pixels (-11,13): screenToMissile = {(15+32)/64, (37+32)/64} = (0,1);
// offset = (-11,13) + worldToScreen(0,1)=(32,-16) = (21,-3);
// light = (15/8, 37/8) = (1,4) - (0,1)*8 = (1,-4).
type updateMissilePosRow = struct {
	name        string
	traveled    positionDisplacement
	wantTile    tilePosition
	wantOffset  positionDisplacement
	wantLightXY tilePosition
}

func TestUpdateMissilePos(t *testing.T) {
	// given
	tests := []updateMissilePosRow{
		{"fractional diagonal floors the pixel vector",
			positionDisplacement{DeltaX: -11 << 16, DeltaY: 13<<16 + 32768},
			tilePosition{X: 0, Y: 1},
			positionDisplacement{DeltaX: 21, DeltaY: -3},
			tilePosition{X: 1, Y: -4}},
		{"negative pixel vector wraps through iso space",
			positionDisplacement{DeltaX: -32 << 16, DeltaY: -8 << 16},
			tilePosition{X: -1, Y: 0},
			positionDisplacement{DeltaX: 0, DeltaY: 8},
			tilePosition{X: 2, Y: 2}},
		{"whole tiles inside the same tile keep offset small",
			positionDisplacement{DeltaX: 3 << 16, DeltaY: 0},
			tilePosition{X: 0, Y: 0},
			positionDisplacement{DeltaX: 3, DeltaY: 0},
			tilePosition{X: 0, Y: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateMissilePosCase(t, tt)
		})
	}
}

func updateMissilePosCase(t *testing.T, tt updateMissilePosRow) {
	t.Helper()
	// given
	spies := &lightSpies{}
	w := &hitWorld{}
	restore := spies.wire(w)
	defer restore()

	m := freshMissile(0, 0)
	m.Lid = 3
	m.Position.Traveled = tt.traveled

	// when
	updateMissilePos(w, &m)

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

// TestMoveMissilePos: the charging-monster correction (cpp:233-251)
// — East→SE, West→SW, South/SW/SE→South; the tile moves by the direction
// delta and the offset by worldToScreen(delta). The IsTileAvailable gate
// is the injected seam.
type moveMissilePosRow = struct {
	name       string
	dir        Direction
	available  bool
	wantCalled bool
	wantDelta  tilePosition
	wantOff    positionDisplacement
}

func TestMoveMissilePos(t *testing.T) {
	// given
	tests := []moveMissilePosRow{
		{"east shifts south-east", East, true, true,
			tilePosition{X: 1, Y: 0}, positionDisplacement{DeltaX: -32, DeltaY: -16}},
		{"west shifts south-west", West, true, true,
			tilePosition{X: 0, Y: 1}, positionDisplacement{DeltaX: 32, DeltaY: -16}},
		{"south shifts south", South, true, true,
			tilePosition{X: 1, Y: 1}, positionDisplacement{DeltaX: 0, DeltaY: -32}},
		{"north is untouched", North, true, false,
			tilePosition{X: 0, Y: 0}, positionDisplacement{}},
		{"unavailable tile is untouched", East, false, true,
			tilePosition{X: 0, Y: 0}, positionDisplacement{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moveMissilePosCase(t, tt)
		})
	}
}

func moveMissilePosCase(t *testing.T, tt moveMissilePosRow) {
	t.Helper()
	// given
	called := 0
	w := &hitWorld{}
	cb := noopProcessCallbacks()
	cb.IsTileAvailable = func(_, _, _ int) bool {
		called++
		return tt.available
	}
	w.processCallbacks = cb

	m := freshMissile(64, 64)
	m.MFNum = int(tt.dir)

	// when
	moveMissilePos(w, &m)

	// then
	if called != expectedProbeCount(tt) {
		t.Fatalf("IsTileAvailable calls = %d, want %d", called, expectedProbeCount(tt))
	}
	if m.Position.Tile != (tilePosition{X: 64 + tt.wantDelta.X, Y: 64 + tt.wantDelta.Y}) {
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

// moveTape drives moveMissile tick by tick and records what the checkTile
// spy saw, so the table rows below assert exact visit order.
type moveTape struct {
	world   *hitWorld
	probeIn func(tile tilePosition) bool // scripted checkTile outcome
	probes  []tilePosition
}

func newMoveTape(w *hitWorld) *moveTape { return &moveTape{world: w} }

func (mt *moveTape) checkTile(tile tilePosition) bool {
	mt.probes = append(mt.probes, tile)
	if mt.probeIn != nil {
		return mt.probeIn(tile)
	}
	return true
}

// tickMove runs one moveMissile tick and asserts the full post-state
// (wantChanged is upstream's return: false only for the no-motion tick,
// cpp:565).
// moveExpectation bundles the post-tick state assertions of tickMove.
type moveExpectation struct {
	changed          bool
	probes           []tilePosition
	tile             tilePosition
	traveled, offset positionDisplacement
	light            tilePosition
}

func tickMove(t *testing.T, mt *moveTape, m *Missile, dontMove dontMoveOnFailedProbe, want moveExpectation) {
	t.Helper()
	wantChanged := want.changed

	spies := &lightSpies{}
	restore := spies.wire(mt.world)
	defer restore()

	changed := moveMissile(mt.world, m, mt.checkTile, dontMove)
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

func expectMoveState(t *testing.T, m *Missile, wantTile tilePosition, wantTraveled, wantOffset positionDisplacement) {
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

func expectProbes(t *testing.T, got, want []tilePosition) {
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

func expectLightOffset(t *testing.T, got []lightOffsetCall, want tilePosition) {
	t.Helper()
	if len(got) != 1 {
		t.Fatalf("ChangeLightOffset calls = %d, want 1", len(got))
	}
	if want != (tilePosition{}) {
		if got[0].offsetX != want.X || got[0].offsetY != want.Y {
			t.Fatalf("light offset = (%d, %d), want %v", got[0].offsetX, got[0].offsetY, want)
		}
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
	m.Position.Velocity = positionDisplacement{DeltaX: axisVX, DeltaY: axisVY}
	mt := newMoveTape(w)

	// when/then — one moveMissile tick with its full post-state
	tickMove(t, mt, &m, false, moveExpectation{
		changed:  false,
		probes:   nil,
		tile:     tilePosition{X: 64, Y: 64},
		traveled: positionDisplacement{DeltaX: 720896, DeltaY: 0},
		offset:   positionDisplacement{DeltaX: 11, DeltaY: 0},
		light:    tilePosition{X: 1, Y: -1},
	})
	t.Run("tick2", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, false, moveExpectation{
			changed:  false,
			probes:   nil,
			tile:     tilePosition{X: 64, Y: 64},
			traveled: positionDisplacement{DeltaX: 1441792, DeltaY: 0},
			offset:   positionDisplacement{DeltaX: 22, DeltaY: 0},
			light:    tilePosition{X: 2, Y: -2},
		})
	})
	t.Run("tick3", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, false, moveExpectation{
			changed:  true,
			probes:   []tilePosition{{X: 65, Y: 63}},
			tile:     tilePosition{X: 65, Y: 63},
			traveled: positionDisplacement{DeltaX: 2162688, DeltaY: 0},
			offset:   positionDisplacement{DeltaX: -31, DeltaY: 0},
			light:    tilePosition{X: -4, Y: 4},
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
	m.Position.Velocity = positionDisplacement{DeltaX: diagVX, DeltaY: diagVY}
	mt := newMoveTape(w)

	tickMove(t, mt, &m, false, moveExpectation{
		changed:  true,
		probes:   []tilePosition{{X: 65, Y: 64}, {X: 66, Y: 64}},
		tile:     tilePosition{X: 66, Y: 64},
		traveled: positionDisplacement{DeltaX: 4194304, DeltaY: 2097152},
		light:    tilePosition{},
	})
	t.Run("tick2", func(t *testing.T) {
		mt.probes = nil
		tickMove(t, mt, &m, false, moveExpectation{
			changed:  true,
			probes:   []tilePosition{{X: 67, Y: 64}, {X: 68, Y: 64}},
			tile:     tilePosition{X: 68, Y: 64},
			traveled: positionDisplacement{DeltaX: 8388608, DeltaY: 4194304},
			light:    tilePosition{},
		})
	})
}

// TestMoveMissileStopSemantics pins the upstream stop-on-collision
// semantics (missiles.cpp:611-628).
type moveMissileStopRow = struct {
	name string
	vel  positionDisplacement
	dm   bool // ifCheckTileFailsDontMoveToTile

	wantProbes []tilePosition
	wantTile   tilePosition
	wantTrav   positionDisplacement
	wantOff    positionDisplacement
	wantVel    positionDisplacement
}

func TestMoveMissileStopSemantics(t *testing.T) {
	// given
	tests := []moveMissileStopRow{
		{
			name: "single-tile step blocked reverts the whole velocity",
			vel:  positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}, dm: true,
			wantProbes: []tilePosition{{X: 65, Y: 63}},
			wantTile:   tilePosition{X: 64, Y: 64},
			wantTrav:   positionDisplacement{DeltaX: 0, DeltaY: 0},
			wantOff:    positionDisplacement{DeltaX: 0, DeltaY: 0},
			wantVel:    positionDisplacement{DeltaX: 0, DeltaY: 0},
		},
		{
			name: "single-tile step blocked without dontMove lands on the probed tile",
			vel:  positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}, dm: false,
			wantProbes: []tilePosition{{X: 65, Y: 63}},
			wantTile:   tilePosition{X: 65, Y: 63},
			wantTrav:   positionDisplacement{DeltaX: 32 << 16, DeltaY: 0},
			wantOff:    positionDisplacement{DeltaX: -32, DeltaY: 0},
			wantVel:    positionDisplacement{DeltaX: 32 << 16, DeltaY: 0},
		},
		{
			name: "interpolated tile blocked reverts one increment",
			vel:  positionDisplacement{DeltaX: diagVX, DeltaY: diagVY}, dm: true,
			wantProbes: []tilePosition{{X: 65, Y: 64}},
			wantTile:   tilePosition{X: 64, Y: 64},
			wantTrav:   positionDisplacement{DeltaX: 0, DeltaY: 0},
			wantOff:    positionDisplacement{DeltaX: 0, DeltaY: 0},
			wantVel:    positionDisplacement{DeltaX: 0, DeltaY: 0},
		},
		{
			name: "interpolated tile blocked without dontMove lands there",
			vel:  positionDisplacement{DeltaX: diagVX, DeltaY: diagVY}, dm: false,
			wantProbes: []tilePosition{{X: 65, Y: 64}},
			wantTile:   tilePosition{X: 65, Y: 64},
			wantTrav:   positionDisplacement{DeltaX: 32 << 16, DeltaY: 16 << 16},
			wantOff:    positionDisplacement{DeltaX: 0, DeltaY: 0},
			wantVel:    positionDisplacement{DeltaX: diagVX, DeltaY: diagVY},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moveMissileStopCase(t, tt)
		})
	}
}

func moveMissileStopCase(t *testing.T, tt moveMissileStopRow) {
	t.Helper()
	// given
	w := procWorld()
	m := freshMissile(64, 64)
	m.Position.Velocity = tt.vel
	mt := newMoveTape(w)
	mt.probeIn = func(tilePosition) bool { return false }

	// when
	changed := moveMissile(w, &m, mt.checkTile, dontMoveOnFailedProbe(tt.dm))

	// then
	if !changed {
		t.Fatal("moveMissile changed = false, want true")
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
}

// TestMoveMissileAndCheckMissileColHash: the slow-missile double-hit
// guard — with the tile unchanged, CheckMissileCol reruns only when the
// dMonster^dPlayer hash differs from lastCollisionTargetHash
// (cpp:636-663).
func TestMoveMissileAndCheckMissileColHash(t *testing.T) {
	// given
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: spies.callbacks(),
		processCallbacks: noopProcessCallbacks()}
	w.DMonster[64][64] = 3

	m := freshMissile(64, 64)
	m.Type = Arrow
	m.Source = 0
	m.caster = int(targetMonsters)
	m.Dist = testDist
	m.Duration = testLifespan

	// when/then — three moveMissileAndCheckMissileCol rounds over the
	// changing dMonster occupant
	moveMissileAndCheckMissileCol(w, &m, DamagePhysical, testMinDamage, testMaxDamage, true, false)
	if got := len(spies.monsterMHitSeen); got != 1 {
		t.Fatalf("after call 1: MonsterMHit rolls = %d, want 1", got)
	}
	if m.LastCollisionTargetHash != 3 {
		t.Fatalf("hash = %d, want 3", m.LastCollisionTargetHash)
	}

	// Same occupant: the guard suppresses the second probe.
	moveMissileAndCheckMissileCol(w, &m, DamagePhysical, testMinDamage, testMaxDamage, true, false)
	if got := len(spies.monsterMHitSeen); got != 1 {
		t.Fatalf("after call 2: MonsterMHit rolls = %d, want 1 (guard)", got)
	}

	// A different occupant re-arms the probe.
	w.DMonster[64][64] = 5
	moveMissileAndCheckMissileCol(w, &m, DamagePhysical, testMinDamage, testMaxDamage, true, false)
	if got := len(spies.monsterMHitSeen); got != 2 {
		t.Fatalf("after call 3: MonsterMHit rolls = %d, want 2", got)
	}
}

// TestProcessArrow: duration countdown, _midist increment, source damage
// seam pass-through (player branch), the move, deletion at 0 and the
// PutMissile tail (cpp:2932-2958). Velocity = axis tape.
func TestProcessArrow(t *testing.T) {
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: spies.callbacks()}
	spies2 := &lightSpies{}
	restore2 := spies2.wire(w)
	defer restore2()

	m := freshMissile(64, 64)
	m.Type = Arrow
	m.Source = 0 // player 0
	m.caster = int(targetMonsters)
	m.Dist = 0
	m.Duration = 3
	// 32 px/tick axis velocity. The 32 px iso diagonal maps to the tile
	// delta (+1,-1), but screenToMissile only lands there every other
	// tick: tick 1 reaches (65,63), tick 2 stays (65,63), tick 3 reaches
	// (66,62) — the upstream tile cadence. A monster on each tile makes
	// each moving tick roll MonsterMHit once (spy misses → the arrow
	// survives); cpp:2955 roll pack.
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}
	w.DMonster[63][65] = 3
	w.DMonster[62][66] = 5

	// Tick 1: tile change to (65,63) — roll#1 carries dist 1; the alive
	// missile is registered in dFlags.
	// when
	ProcessArrow(w, &m)

	// then
	arrowTick1(t, spies, w, m)
	// Tick 2: no tile change — the lastCollisionTargetHash guard
	// suppresses the on-tile re-probe of monster 3; only the countdowns
	// tick.
	// when
	ProcessArrow(w, &m)

	// then
	arrowTick2(t, spies, m)

	// Tick 3: tile change to (66,62) — roll#2 carries dist 3; duration
	// 0 deletes the missile and the PutMissile tail early-outs.
	// when
	ProcessArrow(w, &m)

	// then
	arrowTick3(t, spies, w, m)
}

func arrowTick1(t *testing.T, spies *colSpies, w *hitWorld, m Missile) {
	t.Helper()
	// then
	if m.Duration != 2 || m.Dist != 1 || m.DelFlag {
		t.Fatalf("tick 1: duration = %d dist = %d delFlag = %v, want 2 1 false",
			m.Duration, m.Dist, m.DelFlag)
	}
	roll := spies.monsterMHitSeen[0]
	wantRoll := rollCall{minDamage: arrowMinDam, maxDamage: arrowMaxDam, dist: 1,
		missileType: Arrow, damageType: DamagePhysical}
	if roll != wantRoll {
		t.Fatalf("tick 1: roll = %+v, want %+v", roll, wantRoll)
	}
	if w.Grid.DFlags[63][65]&world.FlagMissile == 0 {
		t.Fatal("tick 1: FlagMissile missing at (65, 63)")
	}
}

func arrowTick2(t *testing.T, spies *colSpies, m Missile) {
	t.Helper()
	if m.Duration != 1 || m.Dist != 2 || m.DelFlag {
		t.Fatalf("tick 2: duration = %d dist = %d delFlag = %v, want 1 2 false",
			m.Duration, m.Dist, m.DelFlag)
	}
	if rolls := len(spies.monsterMHitSeen); rolls != 1 {
		t.Fatalf("tick 2: MonsterMHit rolls = %d, want 1 (hash guard)", rolls)
	}
}

func arrowTick3(t *testing.T, spies *colSpies, w *hitWorld, m Missile) {
	t.Helper()
	if m.Duration != 0 || m.Dist != 3 || !m.DelFlag {
		t.Fatalf("tick 3: duration = %d dist = %d delFlag = %v, want 0 3 true",
			m.Duration, m.Dist, m.DelFlag)
	}
	roll := spies.monsterMHitSeen[1]
	wantRoll := rollCall{minDamage: arrowMinDam, maxDamage: arrowMaxDam, dist: 3,
		missileType: Arrow, damageType: DamagePhysical}
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
// dFlags write (cpp:2963-3016). Movement = diagonal tape; the assertions
// stay inline so the tape reads in upstream order.
func TestProcessGenericProjectileLightTape(t *testing.T) {
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: spies.callbacks()}
	light := &lightSpies{}
	restoreLight := light.wire(w)
	defer restoreLight()

	m := freshMissile(64, 64)
	m.Type = Firebolt
	m.Source = 0
	m.caster = int(targetMonsters)
	m.Dam = 12
	m.Lid = 7
	m.Duration = 2
	m.MFNum = int(South)
	m.Position.Velocity = positionDisplacement{DeltaX: diagVX, DeltaY: diagVY}

	// Tick 1: alive — var1/var2 track the tile, ChangeLight(tile, 8).
	// when
	ProcessGenericProjectile(w, &m)

	// then
	projTapeTick1(t, light, w, m)
	// Tick 2: hit — delete, spawn MagmaBallExplosion at the tile, unlight.
	// when
	ProcessGenericProjectile(w, &m)

	// then
	projTapeTick2(t, light, w, &m)
}

func projTapeTick1(t *testing.T, light *lightSpies, w *hitWorld, m Missile) {
	t.Helper()
	if m.Duration != 1 {
		t.Fatalf("tick 1: duration = %d, want 1", m.Duration)
	}
	if m.Var1 != 66 || m.Var2 != 64 {
		t.Fatalf("tick 1: var = (%d, %d), want (66, 64)", m.Var1, m.Var2)
	}
	expectChangeLight(t, light.changes, changeLightCall{lid: 7, x: 66, y: 64, radius: 8}, "lid 7 at (66, 64) radius 8")
	expectFlagMissileAt(t, w.Grid, 66, 64)
	if len(light.unlights) != 0 || len(light.spawnCalls) != 0 {
		t.Fatalf("tick 1: unlights = %v spawns = %v, want none", light.unlights, light.spawnCalls)
	}
}

func projTapeTick2(t *testing.T, light *lightSpies, w *hitWorld, m *Missile) {
	t.Helper()
	if m.Duration != 0 || !m.DelFlag {
		t.Fatalf("tick 2: duration = %d delFlag = %v, want 0 true", m.Duration, m.DelFlag)
	}
	if len(light.spawnCalls) != 1 {
		t.Fatalf("tick 2: spawns = %d, want 1", len(light.spawnCalls))
	}
	expectExplosionSpawnCall(t, light, m)
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

func expectExplosionSpawnCall(t *testing.T, light *lightSpies, parent *Missile) {
	t.Helper()
	call := light.spawnCalls[0]
	if call.src != (tilePosition{X: 68, Y: 64}) ||
		call.dst != (tilePosition{}) ||
		call.dir != South ||
		call.spawnID != MagmaBallExplosion ||
		call.caster != int(targetMonsters) || call.source != 0 ||
		call.parent != parent {
		t.Fatalf("tick 2: spawn = %+v, want MagmaBallExplosion at (68, 64) dir South parent self", call)
	}
}

// TestProcessGenericProjectileSpawnIDSwitch: the explosion dispatch
// (cpp:2982-3001) — base-table rows only; hellfire flare rows
// land with the hellfire slice. NoLight missiles touch no light hook.
type projSpawnSwitchRow = struct {
	name          string
	typ           missileID
	wantExplosion missileID
}

func TestProcessGenericProjectileSpawnIDSwitch(t *testing.T) {
	// given
	tests := []projSpawnSwitchRow{
		{"firebolt", Firebolt, MagmaBallExplosion},
		{"magmaball", MagmaBall, MagmaBallExplosion},
		{"bloodstar", BloodStar, BloodStarExplosion},
		{"acid", Acid, AcidSplat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projSpawnSwitchCase(t, tt)
		})
	}
}

func projSpawnSwitchCase(t *testing.T, tt projSpawnSwitchRow) {
	t.Helper()
	// given
	spies := newColSpies() // monster-target gate open (collision_test.go)
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: spies.callbacks()}
	light := &lightSpies{}
	restore := light.wire(w)
	defer restore()

	m := freshMissile(64, 64)
	m.Type = tt.typ
	m.caster = int(targetMonsters)
	m.Source = 2
	m.Duration = 1
	m.Lid = NoLight // no light plumbing on this path
	m.Position.Velocity = positionDisplacement{DeltaX: axisVX, DeltaY: axisVY}

	// when
	ProcessGenericProjectile(w, &m)

	// then
	if !m.DelFlag {
		t.Fatal("DelFlag not set")
	}
	expectExplosionSpawn(t, light, tt.wantExplosion)
}

func expectExplosionSpawn(t *testing.T, light *lightSpies, wantExplosion missileID) {
	t.Helper()
	if len(light.spawnCalls) != 1 {
		t.Fatalf("spawns = %d, want 1", len(light.spawnCalls))
	}
	if got := light.spawnCalls[0].spawnID; got != wantExplosion {
		t.Fatalf("spawnID = %v, want %v", got, wantExplosion)
	}
	if len(light.unlights) != 0 || len(light.changes) != 0 {
		t.Fatalf("unlights = %v changes = %v, want none with NoLight",
			light.unlights, light.changes)
	}
	// ChangeLightOffset's own NO_LIGHT guard (lighting.cpp:332)
	// makes the updateMissilePos feed harmless — the hook fires,
	// but never with a real light id.
	for _, off := range light.offsets {
		if off.lid != NoLight {
			t.Fatalf("ChangeLightOffset lid = %d, want NoLight", off.lid)
		}
	}
}

// freshMissile seeds the upstream AddMissile state (cpp:2823-2827):
// tile = start = src, traveled/velocity/offset zero, no light, hash 0.
// (Slice-1's SetTraveledFromTile is the loadsave reload shape, not the
// spawn shape — AddMissile leaves traveled zeroed.)
func freshMissile(x, y int) Missile {
	var m Missile
	m.Position.Tile = tilePosition{X: x, Y: y}
	m.Position.Start = m.Position.Tile
	return m
}

func procWorld() *hitWorld {
	return &hitWorld{Grid: world.NewGrid(),
		processCallbacks: noopProcessCallbacks()}
}

// wire installs the recording hooks on the world's process sink; the
// returned func restores the previous callbacks (defer in each test).
func (s *lightSpies) wire(w *hitWorld) func() {
	prev := w.processCallbacks
	cb := noopProcessCallbacks()
	cb.ChangeLightOffset = func(lid int, offsetX, offsetY int8) {
		s.offsets = append(s.offsets, lightOffsetCall{lid: lid, offsetX: int(offsetX), offsetY: int(offsetY)})
	}
	cb.ChangeLight = func(lid int, x, y int, radius uint8) {
		s.changes = append(s.changes, changeLightCall{lid: lid, x: x, y: y, radius: uint8(radius)})
	}
	cb.AddUnLight = func(lid int) {
		s.unlights = append(s.unlights, lid)
	}
	cb.SpawnMissile = func(src, dst tilePosition, dir Direction,
		spawnID missileID, caster, source int, parent *Missile) {
		s.spawnCalls = append(s.spawnCalls, spawnCall{src, dst, dir, spawnID, caster, source, parent})
	}
	cb.SourceDamage = func(_ *Missile) (int, int) {
		s.sourceDam = append(s.sourceDam, true)
		return arrowMinDam, arrowMaxDam
	}
	w.processCallbacks = cb
	return func() { w.processCallbacks = prev }
}

func expectFlagMissileAt(t *testing.T, g *world.Grid, x, y int) {
	t.Helper()
	if g.DFlags[y][x]&world.FlagMissile == 0 {
		t.Fatalf("FlagMissile missing at (%d, %d)", x, y)
	}
}
