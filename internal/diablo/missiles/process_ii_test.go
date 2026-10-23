// Package missiles_test: devilGo-mwlm slice 3a-ii acceptance tests —
// ProcessElementalArrow, ProcessFireball, ProcessMissileExplosion,
// ProcessChargedBolt and ProcessHolyBolt (Source/missiles.cpp:2848-2930,
// 3086-3163, 3613-3630, 3979-4018, 4020-4057). All expected numbers are
// transcribed from the cited upstream arithmetic (the branch constants and
// light radii are verbatim missiles.cpp values cited at each assertion); the
// animation-phase ids use the MissileGraphicID ordinals cited in process.go.
//
//nolint:testpackage // internal-state test: pokes unexported identifiers
package missiles

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// iiSpies wires every seam the slice-3a-ii fns read and records the
// dispatches; scripted fields drive the outcomes.
type iiSpies struct {
	// Scripted outcomes.
	sourceDamMin, sourceDamMax int // SourceDamage (player/monster pack)
	elemMin, elemMax           int // ElementDamage (elemental player pack)
	rnd                        int // GenerateRnd constant tape
	addLightLid                int // AddLight return (the new light id)
	transAt                    func(x, y int) bool
	checkBlock                 func(fromX, fromY, toX, toY int) bool

	// Recordings.
	changes    []changeLightCall // ChangeLight (lighting.h:65)
	addLights  []addLightCall    // AddLight (lighting.h:66)
	unlights   []int             // AddUnLight (lid)
	sourceSeen []bool            // SourceDamage dispatches
	elemSeen   []damageType
}

// changeLightCall records one ChangeLight dispatch
// (lighting.h:65) — shared with process_test.go's lightSpies.
type changeLightCall struct {
	lid, x, y int
	radius    uint8
}

// addLightCall records one AddLight dispatch (lighting.h:66).
type addLightCall struct {
	x, y   int
	radius uint8
}

func newIISpies() *iiSpies {
	return &iiSpies{
		sourceDamMin: 3, sourceDamMax: 9,
		elemMin: 11, elemMax: 13,
		rnd:         4,
		addLightLid: 42,
		transAt:     func(int, int) bool { return false },
		checkBlock:  func(int, int, int, int) bool { return false },
	}
}

// wire installs the recording hooks on the world's process sink; the
// returned func restores the previous callbacks.
func (s *iiSpies) wire(w *hitWorld) func() {
	prev := w.processCallbacks
	cb := noopProcessCallbacks()
	cb.ChangeLight = func(lid, x, y int, radius uint8) {
		s.changes = append(s.changes, changeLightCall{lid: lid, x: x, y: y, radius: radius})
	}
	cb.AddLight = func(x, y int, radius uint8) int {
		s.addLights = append(s.addLights, addLightCall{x: x, y: y, radius: uint8(radius)})
		return s.addLightLid
	}
	cb.AddUnLight = func(lid int) {
		s.unlights = append(s.unlights, lid)
	}
	cb.SourceDamage = func(m *Missile) (int, int) {
		s.sourceSeen = append(s.sourceSeen, true)
		return s.sourceDamMin, s.sourceDamMax
	}
	cb.ElementDamage = func(m *Missile, elem damageType) (int, int) {
		s.elemSeen = append(s.elemSeen, elem)
		return s.elemMin, s.elemMax
	}
	cb.GenerateRnd = func(n int) int { return s.rnd }
	cb.TransAt = func(x, y int) bool { return s.transAt(x, y) }
	cb.CheckBlock = func(fromX, fromY, toX, toY int) bool {
		return s.checkBlock(fromX, fromY, toX, toY)
	}
	w.processCallbacks = cb
	return func() { w.processCallbacks = prev }
}

// makeSolid marks (x, y) solid through the scripted SOL entry
// (tile_properties.cpp:21-29 — the IsTileSolid grid path the fireball
// wall-adjust reads).
func makeSolid(w *hitWorld, x, y int) {
	const scriptedPiece = 1378
	w.Grid.DPiece[y][x] = scriptedPiece
	w.Grid.SOLData[scriptedPiece] = world.TileSolid
}

// TestProcessElementalArrowFlight: the flight phase — dist tally,
// source-scaled physical roll via the seam, light tracking radius 5 and
// the dFlags registration (cpp:2854-2865, 2913-2918).
func TestProcessElementalArrowFlight(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)
	w.DMonster[63][65] = 3 // on the (65,63) landing tile

	m := freshMissile(64, 64)
	m.Type = FireArrow
	m.Source = 0
	m.caster = int(targetMonsters)
	m.Duration = 5
	m.AnimType = int(GraphicArrow)
	m.AnimLen = 15
	m.Lid = 4
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// when
	ProcessElementalArrow(w, &m)

	// then
	if m.Duration != 4 || m.Dist != 1 {
		t.Fatalf("duration = %d dist = %d, want 4 1", m.Duration, m.Dist)
	}
	if m.Position.Tile != (tilePosition{X: 65, Y: 63}) {
		t.Fatalf("tile = %v, want (65, 63)", m.Position.Tile)
	}
	roll := col.monsterMHitSeen[0]
	wantRoll := rollCall{minDamage: 3, maxDamage: 9, dist: 1, missileType: FireArrow,
		damageType: DamagePhysical}
	if roll != wantRoll {
		t.Fatalf("roll = %+v, want %+v", roll, wantRoll)
	}
	if len(spies.changes) != 1 || spies.changes[0] != (changeLightCall{lid: 4, x: 65, y: 63, radius: 5}) {
		t.Fatalf("ChangeLight = %+v, want lid 4 at (65, 63) radius 5", spies.changes)
	}
	if w.Grid.DFlags[63][65]&world.FlagMissile == 0 {
		t.Fatal("FlagMissile missing at (65, 63)")
	}
	if m.DelFlag {
		t.Fatal("DelFlag set early")
	}
}

// TestProcessElementalArrowTrap: the trap flight branch rolls
// GenerateRnd(10)+1+currlevel twice (cpp:2858-2862) — scripted
// rnd 4, currlevel 2 → (7, 9) through the MonsterTrapHit dispatch.
func TestProcessElementalArrowTrap(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)
	w.processCallbacks.Currlevel = 2 // trap dice base (cpp:2858-2862)
	// The 32 px/tick cadence visits (65,63),(65,63),(66,62),(66,62),
	// (67,61): the monster sits on the final tile, so both rolls land
	// on the duration-0 tick (flight probe + explosion re-probe).
	w.DMonster[61][67] = 3

	m := freshMissile(64, 64)
	m.Type = LightningArrow
	m.Source = -1 // trap (IsTrap, h:150-152)
	m.caster = int(targetMonsters)
	m.Duration = 5
	m.AnimLen = 15
	m.AnimType = int(GraphicArrow)
	m.Lid = NoLight
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// Ticks 1-4: setup — no monster under the flight path, so no rolls.
	for range 4 {
		ProcessElementalArrow(w, &m)
	}
	if len(col.trapHitSeen) != 0 {
		t.Fatalf("flight rolls before the landing tile = %d, want 0",
			len(col.trapHitSeen))
	}

	// Tick 5, the tick under test: the duration-0 explosion tick — the
	// flight probe (7, 9) physical, then the same trap dice in the
	// LIGHTNING re-probe.
	// when
	ProcessElementalArrow(w, &m)

	// then
	if len(col.trapHitSeen) != 2 {
		t.Fatalf("MonsterTrapHit rolls = %d, want 2", len(col.trapHitSeen))
	}
	flightRoll := col.trapHitSeen[0]
	wantFlight := rollCall{minDamage: 7, maxDamage: 9, dist: 5,
		missileType: LightningArrow, damageType: DamagePhysical}
	if flightRoll != wantFlight {
		t.Fatalf("flight roll = %+v, want %+v", flightRoll, wantFlight)
	}
	roll := col.trapHitSeen[1]
	wantRoll := rollCall{minDamage: 7, maxDamage: 9, dist: 5,
		missileType: LightningArrow, damageType: DamageLightning,
		keepOnCollision: true}
	if roll != wantRoll {
		t.Fatalf("explosion roll = %+v, want %+v", roll, wantRoll)
	}
	// NoLight: ChangeLight still feeds (the upstream lighting guard
	// no-ops it, lighting.cpp:332) but never with a real lid.
	for _, c := range spies.changes {
		if c.lid != NoLight {
			t.Fatalf("ChangeLight lid = %d, want NoLight", c.lid)
		}
	}
	if m.Duration != 14 {
		t.Fatalf("duration = %d, want 14 (AnimLen-1)", m.Duration)
	}
}

// TestProcessElementalArrowExplosion: the flight tick that hits duration 0
// transitions to the explosion animation (cpp:2866-2911):
// setDefaultFrameGroup, duration = _miAnimLen-1, stopMissile, the
// LightningArrow explosion roll (player lightning damage seam, lightning
// type, ChargedBolt anim) through a keepOnCollision CheckMissileCol.
func TestProcessElementalArrowExplosion(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)
	w.DMonster[63][65] = 3 // on the landing tile the explosion probes

	m := freshMissile(64, 64)
	m.Type = LightningArrow
	m.Source = 0
	m.caster = int(targetMonsters)
	m.Duration = 2
	m.AnimType = int(GraphicArrow)
	m.AnimLen = 15
	m.Lid = 4
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// Tick 1 is setup: an alive flight tick over the monster tile (rolls
	// once, spy misses → the arrow survives).
	ProcessElementalArrow(w, &m)
	if m.Duration != 1 {
		t.Fatalf("tick 1: duration = %d, want 1", m.Duration)
	}

	// when: Tick 2, the duration-0 explosion transition.
	ProcessElementalArrow(w, &m)

	// then
	if m.Duration != 14 { // _miAnimLen - 1
		t.Fatalf("duration = %d, want 14", m.Duration)
	}
	if m.AnimType != int(GraphicChargedBolt) {
		t.Fatalf("animType = %d, want ChargedBolt (%d)",
			m.AnimType, GraphicChargedBolt)
	}
	if m.MFNum != 0 {
		t.Fatalf("frame group = %d, want 0 (default)", m.MFNum)
	}
	if m.Position.Velocity != (positionDisplacement{}) {
		t.Fatalf("velocity = %v, want zero (stopMissile)", m.Position.Velocity)
	}
	if len(spies.elemSeen) != 1 || spies.elemSeen[0] != DamageLightning {
		t.Fatalf("element damage dispatches = %v, want [Lightning]", spies.elemSeen)
	}
	roll := col.monsterMHitSeen[len(col.monsterMHitSeen)-1]
	wantRoll := rollCall{minDamage: 11, maxDamage: 13, dist: 2,
		missileType: LightningArrow, damageType: DamageLightning,
		keepOnCollision: true}
	if roll != wantRoll {
		t.Fatalf("roll = %+v, want %+v", roll, wantRoll)
	}
	if !col.monsterMHitKeepSeen() {
		t.Fatal("explosion re-probe not keepOnCollision")
	}
	if m.DelFlag {
		t.Fatal("DelFlag set early (duration restarted at AnimLen-1)")
	}
	if len(spies.changes) != 1 || spies.changes[0] != (changeLightCall{lid: 4, x: 65, y: 63, radius: 5}) {
		t.Fatalf("ChangeLight = %+v, want lid 4 at (65, 63) radius 5", spies.changes)
	}
}

// TestProcessElementalArrowAnimPhase: the ChargedBolt/MagmaBallExplosion
// animation phase only advances the light by the animation frame
// (cpp:2851-2853) — ChangeLight(frame+5) then the radius-5 tile
// tracker; no move, no dist, no roll.
func TestProcessElementalArrowAnimPhase(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)

	m := freshMissile(64, 64)
	m.Type = FireArrow
	m.Duration = 2
	m.AnimType = int(GraphicChargedBolt)
	m.AnimFrame = 3
	m.Lid = 4

	// when
	ProcessElementalArrow(w, &m)

	// then
	if m.Duration != 1 || m.Dist != 0 {
		t.Fatalf("duration = %d dist = %d, want 1 0", m.Duration, m.Dist)
	}
	if len(col.monsterMHitSeen) != 0 || len(col.trapHitSeen) != 0 {
		t.Fatal("anim phase rolled damage")
	}
	// The anim phase only feeds ChangeLight(frame+5)
	// (cpp:2852); the radius-5 tracker is flight-only
	// (cpp:2913-2918 else-branch).
	want := changeLightCall{lid: 4, x: 64, y: 64, radius: 8}
	if len(spies.changes) != 1 || spies.changes[0] != want {
		t.Fatalf("ChangeLight = %+v, want %+v", spies.changes, want)
	}
}

// TestProcessFireballFlight: the flight tick tracks the light radius 8
// (cpp:3152-3157).
func TestProcessFireballFlight(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)

	m := freshMissile(64, 64)
	m.Type = Fireball
	m.caster = int(targetMonsters)
	m.Dam = 12
	m.Duration = 3
	m.Lid = 6
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// when
	ProcessFireball(w, &m)

	// then
	if m.Duration != 2 {
		t.Fatalf("duration = %d, want 2", m.Duration)
	}
	if m.Position.Tile != (tilePosition{X: 65, Y: 63}) {
		t.Fatalf("tile = %v, want (65, 63)", m.Position.Tile)
	}
	if len(spies.changes) != 1 || spies.changes[0] != (changeLightCall{lid: 6, x: 65, y: 63, radius: 8}) {
		t.Fatalf("ChangeLight = %+v, want lid 6 at (65, 63) radius 8", spies.changes)
	}
	if graphicID(m.AnimType) == GraphicBigExplosion {
		t.Fatal("animType entered the explosion phase in flight")
	}
}

// TestProcessFireballImpactRing: the impact tick (duration hits 0) probes
// the 9-entry offset ring through CheckBlock
// (cpp:3103-3112) — blocking every tile but (66,62) pins the
// East offset math ((1,-1) from the impact tile mp=(65,63): the 64 px
// step truncates to one tile, (65,63)); the wall-adjust !TransAt branch
// shifts the tile south and the offset up (cpp:3115-3118).
func TestProcessFireballImpactRing(t *testing.T) {
	// given
	spies := newIISpies()
	spies.checkBlock = func(_, _, toX, toY int) bool {
		return toX != 66 || toY != 62 // only the East ring tile is visible (De Morgan)
	}
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)
	w.DMonster[62][66] = 3 // on the visible East tile

	m := freshMissile(64, 64)
	m.Type = Fireball
	m.caster = int(targetMonsters)
	m.Dam = 12
	m.Duration = 1 // the 64 px step lands on (65,63) in one tick
	m.AnimFrame = 1
	m.AnimLen = 15
	m.Lid = 6
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// when
	ProcessFireball(w, &m)

	// then: the roll comes from the one visible ring tile.
	if len(col.monsterMHitSeen) != 1 {
		t.Fatalf("MonsterMHit rolls = %d, want 1", len(col.monsterMHitSeen))
	}
	roll := col.monsterMHitSeen[0]
	wantRoll := rollCall{minDamage: 12, maxDamage: 12, dist: 0,
		missileType: Fireball, damageType: DamageFire,
		keepOnCollision: true}
	if roll != wantRoll {
		t.Fatalf("roll = %+v, want %+v", roll, wantRoll)
	}
	// The impact feeds the AnimFrame radius (cpp:3107); the
	// flight tracker never fires (the single tick is the impact).
	want := changeLightCall{lid: 6, x: 65, y: 63, radius: 1}
	if len(spies.changes) != 1 || spies.changes[0] != want {
		t.Fatalf("ChangeLight = %+v, want %+v", spies.changes, want)
	}
	// Wall-adjust: !TransAt(mp) → tile += South, offset.dy -= 32.
	if m.Position.Tile != (tilePosition{X: 66, Y: 64}) {
		t.Fatalf("tile = %v, want (66, 64)", m.Position.Tile)
	}
	// flight offset (-32, 0); branch 1 subtracts 32 from dy → (-32, -32).
	if m.Position.Offset != (positionDisplacement{DeltaX: -32, DeltaY: -32}) {
		t.Fatalf("offset = %v, want (-32, -32)", m.Position.Offset)
	}
	if m.AnimType != int(GraphicBigExplosion) || m.Duration != 14 || m.MFNum != 0 {
		t.Fatalf("anim = %d duration = %d group = %d, want BigExplosion 14 0",
			m.AnimType, m.Duration, m.MFNum)
	}
	if m.Position.Velocity != (positionDisplacement{}) {
		t.Fatalf("velocity = %v, want zero", m.Position.Velocity)
	}
	if w.Grid.DFlags[64][66]&world.FlagMissile == 0 {
		t.Fatal("FlagMissile missing at the shifted tile (66, 64)")
	}
}

// TestProcessFireballWallAdjustDX: the velocity.dx>0 compound branch —
// transparent (0,+1) with a solid SW neighbor shifts only the offset x
// (cpp:3128-3131).
func TestProcessFireballWallAdjustDX(t *testing.T) {
	// given
	spies := newIISpies()
	// mp (65,63) and mp+(0,+1) (65,64) transparent; mp+SW (65,64)
	// solid — only branch 3 (dx>0) fires (cpp:3124-3129).
	spies.transAt = func(x, y int) bool { return y == 63 || y == 64 }
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)
	makeSolid(w, 65, 64)
	spies.checkBlock = func(int, int, int, int) bool { return true }

	m := freshMissile(64, 64)
	m.Type = Fireball
	m.caster = int(targetMonsters)
	m.Dam = 12
	m.Duration = 1
	m.AnimLen = 15
	m.Lid = NoLight
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// when
	ProcessFireball(w, &m)

	// then: flight offset (-32, 0); branch 3 subtracts 32 from dx → (-64, 0).
	if m.Position.Offset != (positionDisplacement{DeltaX: -64, DeltaY: 0}) {
		t.Fatalf("offset = %v, want (-64, 0)", m.Position.Offset)
	}
	if m.Position.Tile != (tilePosition{X: 65, Y: 63}) {
		t.Fatalf("tile = %v, want (65, 63) (no south shift)", m.Position.Tile)
	}
}

// TestProcessFireballBurnout: the BigExplosion animation phase only waits
// for duration 0 (cpp:3090-3094) — the missile stays registered
// until the final tick deletes it and removes the light.
func TestProcessFireballBurnout(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)

	m := freshMissile(64, 64)
	m.Type = Fireball
	m.AnimType = int(GraphicBigExplosion)
	m.Duration = 2
	m.Lid = 6

	// The first tick is setup: the burn-out mid-flight, still registered.
	ProcessFireball(w, &m)
	if m.Duration != 1 || m.DelFlag {
		t.Fatalf("duration = %d delFlag = %v, want 1 false", m.Duration, m.DelFlag)
	}
	if w.Grid.DFlags[64][64]&world.FlagMissile == 0 {
		t.Fatal("FlagMissile missing at (64, 64)")
	}

	// when
	ProcessFireball(w, &m)

	// then
	if !m.DelFlag {
		t.Fatal("DelFlag not set at burn-out end")
	}
	if len(spies.unlights) != 1 || spies.unlights[0] != 6 {
		t.Fatalf("unlights = %v, want [6]", spies.unlights)
	}
}

// TestProcessMissileExplosionRamp: the ExpLight ramp (cpp:3613-3630)
// — var1 0 lights the tile at radius 9 through AddLight, later ticks
// ChangeLight down the table, duration 0 deletes and unlights.
func TestProcessMissileExplosionRamp(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)

	m := freshMissile(64, 64)
	m.Type = MagmaBallExplosion
	m.Duration = 3
	m.Lid = NoLight

	// when: the seed tick.
	ProcessMissileExplosion(w, &m)
	// then
	if m.Lid != 42 {
		t.Fatalf("lid = %d, want 42 (AddLight return)", m.Lid)
	}
	if len(spies.addLights) != 1 || spies.addLights[0] != (addLightCall{x: 64, y: 64, radius: 9}) {
		t.Fatalf("AddLight = %+v, want (64, 64) radius 9", spies.addLights)
	}
	if m.Var1 != 1 {
		t.Fatalf("var1 = %d, want 1", m.Var1)
	}

	// when: the first ramp tick.
	ProcessMissileExplosion(w, &m)
	// then
	if len(spies.changes) != 1 || spies.changes[0] != (changeLightCall{lid: 42, x: 64, y: 64, radius: 10}) {
		t.Fatalf("ChangeLight = %v, want [42 64 64 10] (ExpLight[1])", spies.changes)
	}

	// when: the final tick.
	ProcessMissileExplosion(w, &m)
	// then
	if !m.DelFlag {
		t.Fatal("DelFlag not set at duration 0")
	}
	if len(spies.unlights) != 1 || spies.unlights[0] != 42 {
		t.Fatalf("unlights = %v, want [42]", spies.unlights)
	}
	// Upstream leaves the previous frame's registration in dFlags
	// (RedoMissileFlags clears the missile layer per frame); the port
	// asserts only that the final tick wrote nothing new.
	if len(spies.changes) != 1 {
		t.Fatalf("ChangeLight calls = %d, want 1 (ExpLight[1] only)", len(spies.changes))
	}
}

// TestProcessChargedBoltSteering: var3 0 re-aims through the BPath table —
// mirnd 0 reads BPath[0] = -1 = a left turn from East → NorthEast, the
// velocity re-aims at 8 px/tick, var3 recharges to 16 and the light rides
// at radius var1 (cpp:3983-3996, 4010-4012).
func TestProcessChargedBoltSteering(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)

	m := freshMissile(64, 64)
	m.Type = ChargedBolt
	m.caster = int(targetMonsters)
	m.Dam = 5
	m.Duration = 5
	m.AnimType = int(GraphicChargedBolt)
	m.AnimLen = 15
	m.Lid = 3
	m.Var2 = int(East) // base direction
	m.Var3 = 0
	m.Mirnd = 0

	// when
	ProcessChargedBolt(w, &m)

	// then
	if m.Mirnd != 1 {
		t.Fatalf("mirnd = %d, want 1", m.Mirnd)
	}
	if m.Var3 != 16 {
		t.Fatalf("var3 = %d, want 16", m.Var3)
	}
	// velocity = worldToNormalScreen((0,1)) * 8 = (370720, -185360).
	wantV := positionDisplacement{DeltaX: 370720, DeltaY: -185360}
	if m.Position.Velocity != wantV {
		t.Fatalf("velocity = %v, want %v", m.Position.Velocity, wantV)
	}
	if m.Position.Tile != (tilePosition{X: 64, Y: 64}) {
		t.Fatalf("tile = %v, want (64, 64) (8 px does not cross a tile)", m.Position.Tile)
	}
	if len(spies.changes) != 1 || spies.changes[0] != (changeLightCall{lid: 3, x: 64, y: 64, radius: 0}) {
		t.Fatalf("ChangeLight = %v, want [3 64 64 0] (var1)", spies.changes)
	}
}

// TestProcessChargedBoltHit: the hit branch freezes the bolt into its
// Lightning burn-out (cpp:4003-4009) — var1 8, offset/velocity
// zero, Lightning anim, duration restarted at _miAnimLen.
func TestProcessChargedBoltHit(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	col.monsterMHit = true // the roll lands
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)
	w.DMonster[63][65] = 3 // on the (65,63) landing tile

	m := freshMissile(64, 64)
	m.Type = ChargedBolt
	m.caster = int(targetMonsters)
	m.Dam = 5
	m.Duration = 5
	m.AnimType = int(GraphicChargedBolt)
	m.AnimLen = 7
	m.Lid = 3
	m.Var2 = int(South) // (1,1) delta
	m.Var3 = 16         // no steering this tick
	// 32 px/tick → the bolt lands on (65,63), where the monster waits.
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// when
	ProcessChargedBolt(w, &m)

	// then
	if !m.HitFlag {
		t.Fatal("HitFlag not set")
	}
	if m.Var1 != 8 || m.MFNum != 0 {
		t.Fatalf("var1 = %d frameGroup = %d, want 8 0", m.Var1, m.MFNum)
	}
	if m.Position.Velocity != (positionDisplacement{}) {
		t.Fatalf("velocity = %v, want zero", m.Position.Velocity)
	}
	if m.AnimType != int(GraphicLightning) {
		t.Fatalf("animType = %d, want Lightning (%d)", m.AnimType, GraphicLightning)
	}
	if m.Duration != 7 { // restarted at _miAnimLen
		t.Fatalf("duration = %d, want 7", m.Duration)
	}
	if len(spies.changes) != 1 || spies.changes[0] != (changeLightCall{lid: 3, x: 65, y: 63, radius: 8}) {
		t.Fatalf("ChangeLight = %v, want [3 65 63 8]", spies.changes)
	}
}

// TestProcessChargedBoltBurnout: the Lightning phase skips steering, the
// move and the light — only the duration counts down and the missile is
// deleted at 0 (cpp:3982-4016: the move, the collision probe and
// ChangeLight all sit inside the non-Lightning guard).
func TestProcessChargedBoltBurnout(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)

	m := freshMissile(64, 64)
	m.Type = ChargedBolt
	m.Duration = 1
	m.AnimType = int(GraphicLightning)
	m.Lid = 3
	m.Var1 = 8

	// when
	ProcessChargedBolt(w, &m)

	// then
	if !m.DelFlag {
		t.Fatal("DelFlag not set at burn-out")
	}
	if len(spies.unlights) != 1 || spies.unlights[0] != 3 {
		t.Fatalf("unlights = %v, want [3]", spies.unlights)
	}
	if len(spies.changes) != 0 {
		t.Fatalf("ChangeLight = %v, want none (burn-out skips the light)", spies.changes)
	}
}

// TestProcessChargedBoltBurnoutNoReProbe: the burn-out tick of a frozen
// bolt only counts the duration down (cpp:3999-4011 is inside
// the non-Lightning guard) — the monster standing on the frozen tile is
// not probed again and the duration is not restarted at _miAnimLen.
func TestProcessChargedBoltBurnoutNoReProbe(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)
	w.DMonster[64][64] = 3 // the monster on the frozen tile

	m := freshMissile(64, 64)
	m.Type = ChargedBolt
	m.caster = int(targetMonsters)
	m.Dam = 5
	m.Duration = 2
	m.AnimType = int(GraphicLightning)
	m.AnimLen = 7
	m.Lid = 3
	m.Var1 = 8

	// The first tick is setup: the burn-out mid-flight, no re-probe yet.
	ProcessChargedBolt(w, &m)
	if len(col.monsterMHitSeen) != 0 {
		t.Fatalf("MonsterMHit rolls = %d, want 0 (burn-out skips the probe)", len(col.monsterMHitSeen))
	}
	if m.Duration != 1 {
		t.Fatalf("duration = %d, want 1 (no AnimLen restart)", m.Duration)
	}
	if m.AnimType != int(GraphicLightning) {
		t.Fatalf("animType = %d, want Lightning (%d)", m.AnimType, GraphicLightning)
	}
	if m.DelFlag {
		t.Fatal("DelFlag set before the duration ran out")
	}

	// when: the final burn-out tick.
	ProcessChargedBolt(w, &m)

	// then
	if !m.DelFlag {
		t.Fatal("DelFlag not set at burn-out")
	}
	if len(col.monsterMHitSeen) != 0 {
		t.Fatalf("MonsterMHit rolls = %d, want 0", len(col.monsterMHitSeen))
	}
	if len(spies.unlights) != 1 || spies.unlights[0] != 3 {
		t.Fatalf("unlights = %v, want [3]", spies.unlights)
	}
}

// TestProcessHolyBoltTransition: the flight tick that hits duration 0
// snaps into the HolyBoltExplosion animation and stops
// (cpp:4029-4034) — no light tracking on the transition tick.
func TestProcessHolyBoltTransition(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)

	m := freshMissile(64, 64)
	m.Type = HolyBolt
	m.caster = int(targetMonsters)
	m.Dam = 7
	m.Duration = 1
	m.AnimType = int(GraphicHolyBolt)
	m.AnimLen = 9
	m.Lid = 5
	m.Position.Velocity = positionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// when
	ProcessHolyBolt(w, &m)

	// then
	if m.AnimType != int(GraphicHolyBoltExplosion) {
		t.Fatalf("animType = %d, want HolyBoltExplosion (%d)",
			m.AnimType, GraphicHolyBoltExplosion)
	}
	if m.Duration != 8 || m.MFNum != 0 {
		t.Fatalf("duration = %d group = %d, want 8 0", m.Duration, m.MFNum)
	}
	if m.Position.Velocity != (positionDisplacement{}) {
		t.Fatalf("velocity = %v, want zero (stopMissile)", m.Position.Velocity)
	}
	if len(spies.changes) != 0 {
		t.Fatalf("ChangeLight fired on the transition tick: %v", spies.changes)
	}
	if w.Grid.DFlags[63][65]&world.FlagMissile == 0 {
		t.Fatal("FlagMissile missing at (65, 63)")
	}
}

// TestProcessHolyBoltExplosionPhase: the explosion animation rides
// ChangeLight(frame+7) and deletes at duration 0
// (cpp:4036-4041).
func TestProcessHolyBoltExplosionPhase(t *testing.T) {
	// given
	spies := newIISpies()
	col := newColSpies()
	w := &hitWorld{Grid: world.NewGrid(), hitCallbacks: col.callbacks()}
	spies.wire(w)

	m := freshMissile(64, 64)
	m.Type = HolyBolt
	m.AnimType = int(GraphicHolyBoltExplosion)
	m.AnimFrame = 2
	m.Duration = 1
	m.Lid = 5

	// when
	ProcessHolyBolt(w, &m)

	// then
	if len(spies.changes) != 1 || spies.changes[0] != (changeLightCall{lid: 5, x: 64, y: 64, radius: 9}) {
		t.Fatalf("ChangeLight = %v, want [5 64 64 9] (frame+7)", spies.changes)
	}
	if !m.DelFlag {
		t.Fatal("DelFlag not set at explosion burn-out")
	}
	if len(spies.unlights) != 1 || spies.unlights[0] != 5 {
		t.Fatalf("unlights = %v, want [5]", spies.unlights)
	}
}
