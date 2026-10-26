// Package missiles_test: missiles.CheckMissileCol branch coverage
// (devilGo-mwlm slice-2c). Ports the branches implied by upstream
// Source/cpp:479-538 (missiles.CheckMissileCol) and 857-870
// (IsMissileBlockedByTile): the bounds guard, the monster dispatch
// table, the player dispatch table, every flag combination
// (keepOnCollision / isDamageShifted / gbIsHellfire-blocked /
// onlyHitWalking), the wall + breakable-object branch, and the
// hit-sound tail. Entity pipelines are the port's hitCallbacks seams
// driven with recording spies; the world grid is a scripted
// world.Grid; the single RNG consumer (the hellfire blocked-shot
// rotation, cpp:529) replays rotation_test.go's deterministic
// LCG tape.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// Shared fixture constants: the probe tile sits inside bounds, and the
// roll argument pack mirrors one scripted missiles.CheckMissileCol call
// (cpp:479 missiles.DamageType/minDamage/maxDamage arguments).
const (
	probeX = 5
	probeY = 5

	testMinDamage = 7
	testMaxDamage = 13
	testDist      = 4

	// testLifespan is the Duration colMissile seeds; "alive" assertions
	// expect this value back.
	testLifespan = 10
)

// rollWant carries the per-assertion roll expectations that vary
// between cases (cpp:496, 498, 516, 519 and 523 forward the
// argument pack); wantOneRoll asserts the fixed pack
// (testMinDamage/testMaxDamage/testDist/missiles.Arrow) plus these fields.
// keepOnCollision records the probe's keep argument
// (!dontDeleteOnCollision upstream); slice 3a-ii asserts it for the
// explosion re-probes (cpp:2911, 3112, 3623, 4056 pass true).
type rollWant struct {
	damageType missiles.DamageType
	shifted    bool
	keep       missiles.KeepOnCollision
}

// rollCall records one damage-roll dispatch with named fields.
type rollCall struct {
	minDamage, maxDamage, dist int
	missileType                missiles.MissileID
	damageType                 missiles.DamageType
	shifted                    bool
	keep                       missiles.KeepOnCollision
}

// playerMHitArgs records one PlayerMHit dispatch: the recorded roll call
// plus the monster side (cpp:517-519 &monster, cpp:522
// nullptr for trap-or-self kills) and the death reason
// (cpp:519 deathReason::MonsterOrTrap, cpp:522).
type playerMHitArgs struct {
	roll        rollCall
	monsterSide *int
	deathReason missiles.DeathReason
}

// colSpies scripts the hitCallbacks outcomes and records every dispatch.
// Zero value: MonsterTargeted reports not-targeted and every roll
// misses; newColSpies flips MonsterTargeted to the common targeted case.
type colSpies struct {
	// Scripted outcomes (cpp:489 gate, 496/498/516/519/523 rolls).
	targetMonster     bool
	conflict          bool
	trapHit           bool
	monsterMHit       bool
	plr2PlrHit        bool
	plr2PlrBlocked    bool
	playerMHitHit     bool
	playerMHitBlocked bool

	// Recordings.
	targetedSeen       []int16
	onlyHitWalkingSeen *missiles.Direction
	conflictSeen       []monsterPair
	trapHitSeen        []rollCall
	monsterMHitSeen    []rollCall
	plr2PlrSeen        []rollCall
	playerMHitSeen     []playerMHitArgs
	breakSeen          int
	hitSoundSeen       int
}

// monsterPair is one MonsterFactionConflict recording: target and source
// monster ids.
type monsterPair struct {
	target, source int
}

type monsterRollDispatchRow = struct {
	name     string
	source   int
	caster   missiles.Caster
	conflict bool
	wantTrap bool
	wantMHit bool
}

type playerRollDispatchRow = struct {
	name        string
	source      int
	caster      missiles.Caster
	wantPlr2Plr bool
	wantPHit    bool
	wantSide    *int
	wantDeath   missiles.DeathReason
}

// TestCheckMissileColOutOfBounds covers the bounds guard
// (cpp:481-482): out-of-dungeon probes are complete no-ops —
// zero result, untouched missile, no dispatch, no sound.
func TestCheckMissileColOutOfBounds(t *testing.T) {
	// given
	s := newColSpies()
	w := colWorld(s)
	m := colMissile()

	// when
	res := missiles.CheckMissileCol(w, &m, missiles.TilePosition{X: world.MAXDUNX, Y: probeY}, missiles.MissileColProbe{
		DamageType:     missiles.DamagePhysical,
		MinDamage:      testMinDamage,
		MaxDamage:      testMaxDamage,
		Shifted:        false,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	// then
	if res != (missiles.CheckMissileColResult{}) {
		t.Fatalf("result = %+v, want zero value", res)
	}

	if m.Duration != testLifespan || m.HitFlag {
		t.Fatalf("missile mutated: Duration = %d, HitFlag = %v", m.Duration, m.HitFlag)
	}

	if len(s.targetedSeen) != 0 || s.hitSoundSeen != 0 {
		t.Fatalf("callbacks fired: targeted = %v, hitSound = %d", s.targetedSeen, s.hitSoundSeen)
	}
}

// TestCheckMissileColMonsterRollDispatch verifies the monster-roll
// dispatch table (cpp:491-498): a trap or an opposing-faction
// monster caster rolls MonsterTrapHit; TARGET_BOTH / TARGET_MONSTERS
// roll MonsterMHit; a same-faction monster caster rolls neither.
func TestCheckMissileColMonsterRollDispatch(t *testing.T) {
	tests := []monsterRollDispatchRow{
		// cpp:491 IsTrap() short-circuits to MonsterTrapHit (496).
		{name: "trap rolls MonsterTrapHit", source: -1, caster: missiles.TargetBoth, wantTrap: true},
		// cpp:492-495 TARGET_PLAYERS + opposing faction ->
		// MonsterTrapHit (496).
		{
			name:   "monster missiles.Caster, opposing faction rolls MonsterTrapHit",
			source: 3, caster: missiles.TargetPlayers, conflict: true, wantTrap: true,
		},
		// cpp:492 faction guard fails and cpp:497
		// IsAnyOf(TARGET_BOTH, TARGET_MONSTERS) fails -> no roll at all.
		{
			name:   "monster missiles.Caster, same faction rolls nothing",
			source: 3, caster: missiles.TargetPlayers,
		},
		// cpp:497-498.
		{name: "TARGET_BOTH rolls MonsterMHit", source: 2, caster: missiles.TargetBoth, wantMHit: true},
		{name: "TARGET_MONSTERS rolls MonsterMHit", source: 2, caster: missiles.TargetMonsters, wantMHit: true},
	}
	// when
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			monsterRollDispatchCase(t, tc)
		})
	}
}

// TestCheckMissileColRollArgumentPassThrough verifies the shared roll
// argument pack lands verbatim: missiles.DamageType, isDamageShifted, min/max,
// _midist and _mitype (cpp:496).
func TestCheckMissileColRollArgumentPassThrough(t *testing.T) {
	s := newColSpies()
	s.trapHit = true
	w := colWorld(s)
	placeMonster(w, 1)

	m := colMissile()
	m.Source = -1 // trap

	// when
	missiles.CheckMissileCol(w, &m, colProbe(), missiles.MissileColProbe{
		DamageType:     missiles.DamageFire,
		MinDamage:      testMinDamage,
		MaxDamage:      testMaxDamage,
		Shifted:        true,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	// then
	wantOneRoll(t, s.trapHitSeen, rollWant{
		damageType: missiles.DamageFire,
		shifted:    true,
	}, "MonsterTrapHit")
}

// TestCheckMissileColMonsterTargetGate covers the dMonster gate
// (cpp:485-489): an empty cell never dispatches; a cell the
// injected targeted-gate rejects never rolls; onlyHitWalking passes
// through the seam verbatim; and a walking cell (-3) still resolves
// Monsters[abs(mid)-1] = index 2 for the faction check (cpp:487).
func TestCheckMissileColMonsterTargetGate(t *testing.T) {
	t.Run("empty cell never dispatches", func(t *testing.T) {
		monsterGateEmptyCell(t)
	})

	t.Run("not-targeted cell never rolls", func(t *testing.T) {
		monsterGateNotTargeted(t)
	})

	t.Run("onlyHitWalking passes through the seam", func(t *testing.T) {
		monsterGateOnlyHitWalking(t)
	})

	t.Run("walking cell resolves abs(mid)-1 for the faction check", func(t *testing.T) {
		monsterGateWalkingCell(t)
	})
}

// monsterHitFlagsRow is one TestCheckMissileColMonsterHitFlags case.
func TestCheckMissileColPlayerRollDispatch(t *testing.T) {
	for _, tc := range playerRollDispatchTable() {
		t.Run(tc.name, func(t *testing.T) {
			playerRollDispatchCase(t, tc)
		})
	}
}

// TestCheckMissileColPlayerGate covers the player-cell gates: without
// onlyHitWalking the probe skips negative (walking) dPlayer cells
// (cpp:511 ignoreMovingPlayers + player.cpp:2044), and
// TARGET_MONSTERS never rolls against its own source player
// (cpp:515 player id == _misource guard).
func colProbe() missiles.TilePosition {
	return missiles.TilePosition{X: probeX, Y: probeY}
}

func colMissile() missiles.Missile {
	m := missiles.Missile{
		Type:     missiles.Arrow,
		Source:   0,
		Caster:   int(missiles.TargetMonsters),
		Dist:     testDist,
		Duration: testLifespan,
	}
	m.SetTraveledFromTile(probeX, probeY)

	return m
}

func newColSpies() *colSpies {
	return &colSpies{targetMonster: true}
}

func colWorld(s *colSpies) *missiles.HitWorld {
	return &missiles.HitWorld{Grid: world.NewGrid(), HitCallbacks: s.callbacks()}
}

func placeMonster(w *missiles.HitWorld, id int16) {
	w.DMonster[probeY][probeX] = id
}

func placePlayer(w *missiles.HitWorld, id int16) {
	w.DPlayer[probeY][probeX] = id
}

func setTileProps(w *missiles.HitWorld, props world.TileProperties) {
	const scriptedPiece = 1378 // top SOL slot of the LoadLevelSOLData range

	w.Grid.DPiece[probeY][probeX] = scriptedPiece
	w.Grid.SOLData[scriptedPiece] = props
}

func placeObject(w *missiles.HitWorld, obj world.WorldObject) {
	w.Grid.Objects = append(w.Grid.Objects, obj)
	w.Grid.DObject[probeY][probeX] = len(w.Grid.Objects)
}

func fixedRoll() rollCall {
	return rollCall{
		minDamage:   testMinDamage,
		maxDamage:   testMaxDamage,
		dist:        testDist,
		missileType: missiles.Arrow,
	}
}

func wantOneRoll(t *testing.T, got []rollCall, want rollWant, label string) {
	t.Helper()

	if len(got) != 1 {
		t.Fatalf("%s: %d dispatches recorded, want exactly 1 (%+v)", label, len(got), got)
	}

	wantCall := fixedRoll()
	wantCall.damageType = want.damageType
	wantCall.shifted = want.shifted

	wantCall.keep = want.keep
	if got[0] != wantCall {
		t.Fatalf("%s: args = %+v, want %+v", label, got[0], wantCall)
	}
}

func wantNoRolls(t *testing.T, got []rollCall, label string) {
	t.Helper()

	if len(got) != 0 {
		t.Fatalf("%s: fired %d times, want 0 (%+v)", label, len(got), got)
	}
}

func wantOnePlayerMHit(t *testing.T, got []playerMHitArgs, monsterSide *int, reason missiles.DeathReason) {
	t.Helper()

	if len(got) != 1 {
		t.Fatalf("PlayerMHit: %d dispatches, want exactly 1 (%+v)", len(got), got)
	}

	call := got[0]
	if call.roll != fixedRoll() {
		t.Fatalf("PlayerMHit args = %+v, want %+v", call.roll, fixedRoll())
	}

	if (call.monsterSide == nil) != (monsterSide == nil) {
		t.Fatalf("PlayerMHit monsterSide nil-ness = %v, want %v",
			call.monsterSide == nil, monsterSide == nil)
	}

	if monsterSide != nil && *call.monsterSide != *monsterSide {
		t.Fatalf("PlayerMHit monsterSide = %d, want %d", *call.monsterSide, *monsterSide)
	}

	if call.deathReason != reason {
		t.Fatalf("PlayerMHit deathReason = %d, want %d (cpp:522)", call.deathReason, reason)
	}
}

// callbacks wires the spies into the missiles.HitWorld seam (collision.go
// hitCallbacks).
func (s *colSpies) callbacks() missiles.HitCallbacks {
	return missiles.HitCallbacks{
		MonsterTargeted:        s.monsterTargetedCallback(),
		MonsterFactionConflict: s.monsterFactionConflictCallback(),
		MonsterTrapHit:         s.monsterTrapHitCallback(),
		MonsterMHit:            s.monsterMHitCallback(),
		Plr2PlrMHit:            s.plr2PlrMHitCallback(),
		PlayerMHit:             s.playerMHitCallback(),
		BreakObjectMissile:     func() { s.breakSeen++ },
		PlayHitSound:           func() { s.hitSoundSeen++ },
	}
}

// monsterTargetedCallback records and answers the dMonster gate probe.
func (s *colSpies) monsterTargetedCallback() func(mid int16, onlyHitWalking *missiles.Direction) bool {
	return func(mid int16, onlyHitWalking *missiles.Direction) bool {
		s.targetedSeen = append(s.targetedSeen, mid)
		s.onlyHitWalkingSeen = onlyHitWalking

		return s.targetMonster
	}
}

// monsterFactionConflictCallback records and answers the trap-branch faction
// guard.
func (s *colSpies) monsterFactionConflictCallback() func(targetMonster, sourceMonster int) bool {
	return func(targetMonster, sourceMonster int) bool {
		s.conflictSeen = append(s.conflictSeen, monsterPair{targetMonster, sourceMonster})
		return s.conflict
	}
}

// monsterTrapHitCallback records the MonsterTrapHit roll and answers it.
func (s *colSpies) monsterTrapHitCallback() func(minDamage, maxDamage, dist int,
	missileType missiles.MissileID, damageType missiles.DamageType,
	shifted bool, keepOnCollision missiles.KeepOnCollision,
) bool {
	return func(minDamage, maxDamage, dist int,
		missileType missiles.MissileID, damageType missiles.DamageType,
		shifted bool, keepOnCollision missiles.KeepOnCollision,
	) bool {
		s.trapHitSeen = append(s.trapHitSeen,
			rollCall{
				minDamage: minDamage, maxDamage: maxDamage, dist: dist,
				missileType: missileType, damageType: damageType, shifted: shifted,
				keep: keepOnCollision,
			})

		return s.trapHit
	}
}

// monsterMHitCallback records the MonsterMHit roll and answers it.
func (s *colSpies) monsterMHitCallback() func(minDamage, maxDamage, dist int,
	missileType missiles.MissileID, damageType missiles.DamageType,
	shifted bool, keepOnCollision missiles.KeepOnCollision,
) bool {
	return func(minDamage, maxDamage, dist int,
		missileType missiles.MissileID, damageType missiles.DamageType,
		shifted bool, keepOnCollision missiles.KeepOnCollision,
	) bool {
		s.monsterMHitSeen = append(s.monsterMHitSeen,
			rollCall{
				minDamage: minDamage, maxDamage: maxDamage, dist: dist,
				missileType: missileType, damageType: damageType, shifted: shifted,
				keep: keepOnCollision,
			})

		return s.monsterMHit
	}
}

// plr2PlrMHitCallback records the player-vs-player roll and answers it.
func (s *colSpies) plr2PlrMHitCallback() func(minDamage, maxDamage, dist int,
	missileType missiles.MissileID, damageType missiles.DamageType, shifted bool,
) (bool, bool) {
	return func(minDamage, maxDamage, dist int,
		missileType missiles.MissileID, damageType missiles.DamageType, shifted bool,
	) (bool, bool) {
		s.plr2PlrSeen = append(s.plr2PlrSeen,
			rollCall{
				minDamage: minDamage, maxDamage: maxDamage, dist: dist,
				missileType: missileType, damageType: damageType, shifted: shifted,
			})

		return s.plr2PlrHit, s.plr2PlrBlocked
	}
}

// playerMHitCallback records the PlayerMHit dispatch and answers it.
func (s *colSpies) playerMHitCallback() func(monsterSide *int, minDamage, maxDamage, dist int,
	missileType missiles.MissileID, damageType missiles.DamageType, shifted bool,
	reason missiles.DeathReason,
) (bool, bool) {
	return func(monsterSide *int, minDamage, maxDamage, dist int,
		missileType missiles.MissileID, damageType missiles.DamageType, shifted bool,
		reason missiles.DeathReason,
	) (bool, bool) {
		s.playerMHitSeen = append(s.playerMHitSeen, playerMHitArgs{
			roll: rollCall{
				minDamage: minDamage, maxDamage: maxDamage, dist: dist,
				missileType: missileType, damageType: damageType, shifted: shifted,
			},
			monsterSide: monsterSide,
			deathReason: reason,
		})

		return s.playerMHitHit, s.playerMHitBlocked
	}
}

// monsterMHitKeepSeen reports whether the last MonsterMHit dispatch was a
// keep-on-collision probe (slice 3a-ii explosion re-probes,
// cpp:2911/3111).
func (s *colSpies) monsterMHitKeepSeen() bool {
	return len(s.monsterMHitSeen) > 0 &&
		s.monsterMHitSeen[len(s.monsterMHitSeen)-1].keep != missiles.KeepOnCollision(false)
}

func monsterRollDispatchCase(t *testing.T, tc monsterRollDispatchRow) {
	t.Helper()

	s := newColSpies()
	s.conflict = tc.conflict
	s.trapHit = tc.wantTrap
	s.monsterMHit = tc.wantMHit
	w := colWorld(s)
	placeMonster(w, 1)

	m := colMissile()
	m.Source = tc.source
	m.Caster = int(tc.caster)

	res := missiles.CheckMissileCol(w, &m, colProbe(), missiles.MissileColProbe{
		DamageType:     missiles.DamagePhysical,
		MinDamage:      testMinDamage,
		MaxDamage:      testMaxDamage,
		Shifted:        false,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	// then
	if res.MonsterHit != (tc.wantTrap || tc.wantMHit) {
		t.Fatalf("MonsterHit = %v, want %v", res.MonsterHit, tc.wantTrap || tc.wantMHit)
	}

	if tc.wantTrap {
		wantOneRoll(t, s.trapHitSeen, rollWant{}, "MonsterTrapHit")
	} else {
		wantNoRolls(t, s.trapHitSeen, "MonsterTrapHit")
	}

	if tc.wantMHit {
		wantOneRoll(t, s.monsterMHitSeen, rollWant{}, "MonsterMHit")
	} else {
		wantNoRolls(t, s.monsterMHitSeen, "MonsterMHit")
	}

	wantNoRolls(t, s.plr2PlrSeen, "Plr2PlrMHit")

	if len(s.playerMHitSeen) != 0 {
		t.Fatalf("PlayerMHit fired on the monster branch: %+v", s.playerMHitSeen)
	}
}

func monsterGateEmptyCell(t *testing.T) {
	t.Helper()

	s := newColSpies()
	w := colWorld(s)
	m := colMissile()

	// when
	res := missiles.CheckMissileCol(w, &m, colProbe(), missiles.MissileColProbe{
		DamageType:     missiles.DamagePhysical,
		MinDamage:      testMinDamage,
		MaxDamage:      testMaxDamage,
		Shifted:        false,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	// then
	if res.MonsterHit || len(s.targetedSeen) != 0 {
		t.Fatalf("empty cell dispatched: hit = %v, targeted = %v", res.MonsterHit, s.targetedSeen)
	}

	if m.Duration != testLifespan || m.HitFlag {
		t.Fatalf("missile mutated: Duration = %d, HitFlag = %v", m.Duration, m.HitFlag)
	}
}

func monsterGateNotTargeted(t *testing.T) {
	t.Helper()

	s := newColSpies()
	s.targetMonster = false // cpp:489 gate rejects
	w := colWorld(s)
	placeMonster(w, 1)

	m := colMissile()

	// when
	res := missiles.CheckMissileCol(w, &m, colProbe(), missiles.MissileColProbe{
		DamageType:     missiles.DamagePhysical,
		MinDamage:      testMinDamage,
		MaxDamage:      testMaxDamage,
		Shifted:        false,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	// then
	if res.MonsterHit {
		t.Fatal("MonsterHit = true, want false behind a closed gate")
	}

	wantNoRolls(t, s.trapHitSeen, "MonsterTrapHit")
	wantNoRolls(t, s.monsterMHitSeen, "MonsterMHit")

	if m.Duration != testLifespan || m.HitFlag {
		t.Fatalf("missile mutated: Duration = %d, HitFlag = %v", m.Duration, m.HitFlag)
	}
}

func monsterGateOnlyHitWalking(t *testing.T) {
	t.Helper()

	s := newColSpies()
	w := colWorld(s)
	placeMonster(w, 1)

	m := colMissile()
	wallDir := missiles.SouthEast

	// when
	missiles.CheckMissileCol(w, &m, colProbe(), missiles.MissileColProbe{
		DamageType:     missiles.DamagePhysical,
		MinDamage:      testMinDamage,
		MaxDamage:      testMaxDamage,
		Shifted:        false,
		Keep:           false,
		OnlyHitWalking: &wallDir,
	})

	// then
	if s.onlyHitWalkingSeen == nil || *s.onlyHitWalkingSeen != wallDir {
		t.Fatalf("onlyHitWalking seen = %v, want %d", s.onlyHitWalkingSeen, wallDir)
	}
}

func monsterGateWalkingCell(t *testing.T) {
	t.Helper()

	s := newColSpies()
	s.conflict = true
	s.trapHit = true
	w := colWorld(s)
	placeMonster(w, -3) // walking spawn (cpp:487 std::abs(mid) - 1)

	m := colMissile()
	m.Source = 3
	m.Caster = int(missiles.TargetPlayers)

	// when
	res := missiles.CheckMissileCol(w, &m, colProbe(), missiles.MissileColProbe{
		DamageType:     missiles.DamagePhysical,
		MinDamage:      testMinDamage,
		MaxDamage:      testMaxDamage,
		Shifted:        false,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	// then
	if !res.MonsterHit {
		t.Fatal("MonsterHit = false, want trap-path hit")
	}

	if len(s.conflictSeen) != 1 || s.conflictSeen[0] != (monsterPair{target: 2, source: 3}) {
		t.Fatalf("faction conflict args = %v, want [[2 3]] (abs(-3)-1 = 2)", s.conflictSeen)
	}

	wantOneRoll(t, s.trapHitSeen, rollWant{}, "MonsterTrapHit")
}

func playerRollDispatchCase(t *testing.T, tc playerRollDispatchRow) {
	t.Helper()

	s := newColSpies()
	s.plr2PlrHit = tc.wantPlr2Plr
	s.playerMHitHit = tc.wantPHit
	w := colWorld(s)
	placePlayer(w, 1) // dPlayer cell -> player index 0

	m := colMissile()
	m.Source = tc.source
	m.Caster = int(tc.caster)

	// when
	res := missiles.CheckMissileCol(w, &m, colProbe(), missiles.MissileColProbe{
		DamageType:     missiles.DamagePhysical,
		MinDamage:      testMinDamage,
		MaxDamage:      testMaxDamage,
		Shifted:        false,
		Keep:           false,
		OnlyHitWalking: nil,
	})

	// then
	if res.PlayerHit != (tc.wantPlr2Plr || tc.wantPHit) {
		t.Fatalf("PlayerHit = %v, want %v", res.PlayerHit, tc.wantPlr2Plr || tc.wantPHit)
	}

	if tc.wantPlr2Plr {
		wantOneRoll(t, s.plr2PlrSeen, rollWant{}, "Plr2PlrMHit")
	} else {
		wantNoRolls(t, s.plr2PlrSeen, "Plr2PlrMHit")
	}

	if tc.wantPHit {
		wantOnePlayerMHit(t, s.playerMHitSeen, tc.wantSide, tc.wantDeath)
	} else if len(s.playerMHitSeen) != 0 {
		t.Fatalf("PlayerMHit fired: %+v", s.playerMHitSeen)
	}
}

func playerRollDispatchTable() []playerRollDispatchRow {
	sourceOf := func(idx int) *int { return &idx }

	return []playerRollDispatchRow{
		// cpp:514-516: TARGET_MONSTERS + different player id ->
		// Plr2PlrMHit.
		{
			name:   "TARGET_MONSTERS hits another player via Plr2PlrMHit",
			source: 1, caster: missiles.TargetMonsters, wantPlr2Plr: true,
		},
		// cpp:517-519: monster side -> PlayerMHit(&monster,
		// deathReason::MonsterOrTrap).
		{
			name:   "monster side rolls PlayerMHit with the source monster",
			source: 4, caster: missiles.TargetPlayers, wantPHit: true,
			wantSide: sourceOf(4), wantDeath: missiles.DeathMonsterOrTrap,
		},
		// cpp:522-523: TARGET_BOTH, player source ->
		// PlayerMHit(nullptr, deathReason::Player).
		{
			name:   "TARGET_BOTH player source rolls PlayerMHit with missiles.DeathByPlayer",
			source: 0, caster: missiles.TargetBoth, wantPHit: true,
			wantSide: nil, wantDeath: missiles.DeathByPlayer,
		},
		// cpp:522-523: trap source -> PlayerMHit(nullptr,
		// deathReason::MonsterOrTrap).
		{
			name:   "trap source rolls PlayerMHit with missiles.DeathMonsterOrTrap",
			source: -1, caster: missiles.TargetBoth, wantPHit: true,
			wantSide: nil, wantDeath: missiles.DeathMonsterOrTrap,
		},
	}
}
