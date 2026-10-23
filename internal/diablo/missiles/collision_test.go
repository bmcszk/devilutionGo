// Package missiles_test: CheckMissileCol branch coverage
// (devilGo-mwlm slice-2c). Ports the branches implied by upstream
// Source/cpp:479-538 (CheckMissileCol) and 857-870
// (IsMissileBlockedByTile): the bounds guard, the monster dispatch
// table, the player dispatch table, every flag combination
// (keepOnCollision / isDamageShifted / gbIsHellfire-blocked /
// onlyHitWalking), the wall + breakable-object branch, and the
// hit-sound tail. Entity pipelines are the port's hitCallbacks seams
// driven with recording spies; the world grid is a scripted
// world.Grid; the single RNG consumer (the hellfire blocked-shot
// rotation, cpp:529) replays rotation_test.go's deterministic
// LCG tape.
//
//nolint:testpackage // internal-state test: pokes unexported identifiers
package missiles

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// Shared fixture constants: the probe tile sits inside bounds, and the
// roll argument pack mirrors one scripted CheckMissileCol call
// (cpp:479 damageType/minDamage/maxDamage arguments).
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
// (testMinDamage/testMaxDamage/testDist/Arrow) plus these fields.
// keepOnCollision records the probe's keepOnCollision argument
// (!dontDeleteOnCollision upstream); slice 3a-ii asserts it for the
// explosion re-probes (cpp:2911, 3112, 3623, 4056 pass true).
type rollWant struct {
	damageType      damageType
	shifted         bool
	keepOnCollision bool
}

// rollCall records one damage-roll dispatch with named fields.
type rollCall struct {
	minDamage, maxDamage, dist int
	missileType                missileID
	damageType                 damageType
	shifted                    bool
	keepOnCollision            bool
}

// playerMHitArgs records one PlayerMHit dispatch: the recorded roll call
// plus the monster side (cpp:517-519 &monster, cpp:522
// nullptr for trap-or-self kills) and the death reason
// (cpp:519 deathReason::MonsterOrTrap, cpp:522).
type playerMHitArgs struct {
	roll        rollCall
	monsterSide *int
	deathReason deathReason
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
	onlyHitWalkingSeen *Direction
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

// TestCheckMissileColOutOfBounds covers the bounds guard
// (cpp:481-482): out-of-dungeon probes are complete no-ops —
// zero result, untouched missile, no dispatch, no sound.
func TestCheckMissileColOutOfBounds(t *testing.T) {
	// given
	s := newColSpies()
	w := colWorld(s)
	m := colMissile()

	// when
	res := CheckMissileCol(w, &m, tilePosition{X: world.MAXDUNX, Y: probeY},
		DamagePhysical, testMinDamage, testMaxDamage, false, false, nil)

	// then
	if res != (checkMissileColResult{}) {
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
	tests := []struct {
		name     string
		source   int
		caster   caster
		conflict bool
		wantTrap bool
		wantMHit bool
	}{
		// cpp:491 IsTrap() short-circuits to MonsterTrapHit (496).
		{name: "trap rolls MonsterTrapHit", source: -1, caster: targetBoth, wantTrap: true},
		// cpp:492-495 TARGET_PLAYERS + opposing faction ->
		// MonsterTrapHit (496).
		{name: "monster caster, opposing faction rolls MonsterTrapHit",
			source: 3, caster: targetPlayers, conflict: true, wantTrap: true},
		// cpp:492 faction guard fails and cpp:497
		// IsAnyOf(TARGET_BOTH, TARGET_MONSTERS) fails -> no roll at all.
		{name: "monster caster, same faction rolls nothing",
			source: 3, caster: targetPlayers},
		// cpp:497-498.
		{name: "TARGET_BOTH rolls MonsterMHit", source: 2, caster: targetBoth, wantMHit: true},
		{name: "TARGET_MONSTERS rolls MonsterMHit", source: 2, caster: targetMonsters, wantMHit: true},
	}
	// when
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newColSpies()
			s.conflict = tc.conflict
			s.trapHit = tc.wantTrap
			s.monsterMHit = tc.wantMHit
			w := colWorld(s)
			placeMonster(w, 1)
			m := colMissile()
			m.Source = tc.source
			m.caster = int(tc.caster)

			res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
				testMinDamage, testMaxDamage, false, false, nil)

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
		})
	}
}

// TestCheckMissileColRollArgumentPassThrough verifies the shared roll
// argument pack lands verbatim: damageType, isDamageShifted, min/max,
// _midist and _mitype (cpp:496).
func TestCheckMissileColRollArgumentPassThrough(t *testing.T) {
	s := newColSpies()
	s.trapHit = true
	w := colWorld(s)
	placeMonster(w, 1)
	m := colMissile()
	m.Source = -1 // trap

	// when
	CheckMissileCol(w, &m, colProbe(), DamageFire,
		testMinDamage, testMaxDamage, true /* isDamageShifted */, false, nil)

	// then
	wantOneRoll(t, s.trapHitSeen, rollWant{
		damageType: DamageFire,
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
		s := newColSpies()
		w := colWorld(s)
		m := colMissile()

		// when
		res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if res.MonsterHit || len(s.targetedSeen) != 0 {
			t.Fatalf("empty cell dispatched: hit = %v, targeted = %v", res.MonsterHit, s.targetedSeen)
		}
		if m.Duration != testLifespan || m.HitFlag {
			t.Fatalf("missile mutated: Duration = %d, HitFlag = %v", m.Duration, m.HitFlag)
		}
	})

	t.Run("not-targeted cell never rolls", func(t *testing.T) {
		s := newColSpies()
		s.targetMonster = false // cpp:489 gate rejects
		w := colWorld(s)
		placeMonster(w, 1)
		m := colMissile()

		// when
		res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if res.MonsterHit {
			t.Fatal("MonsterHit = true, want false behind a closed gate")
		}
		wantNoRolls(t, s.trapHitSeen, "MonsterTrapHit")
		wantNoRolls(t, s.monsterMHitSeen, "MonsterMHit")
		if m.Duration != testLifespan || m.HitFlag {
			t.Fatalf("missile mutated: Duration = %d, HitFlag = %v", m.Duration, m.HitFlag)
		}
	})

	t.Run("onlyHitWalking passes through the seam", func(t *testing.T) {
		s := newColSpies()
		w := colWorld(s)
		placeMonster(w, 1)
		m := colMissile()
		wallDir := SouthEast

		// when
		CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, &wallDir)

		// then
		if s.onlyHitWalkingSeen == nil || *s.onlyHitWalkingSeen != wallDir {
			t.Fatalf("onlyHitWalking seen = %v, want %d", s.onlyHitWalkingSeen, wallDir)
		}
	})

	t.Run("walking cell resolves abs(mid)-1 for the faction check", func(t *testing.T) {
		s := newColSpies()
		s.conflict = true
		s.trapHit = true
		w := colWorld(s)
		placeMonster(w, -3) // walking spawn (cpp:487 std::abs(mid) - 1)
		m := colMissile()
		m.Source = 3
		m.caster = int(targetPlayers)

		// when
		res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if !res.MonsterHit {
			t.Fatal("MonsterHit = false, want trap-path hit")
		}
		if len(s.conflictSeen) != 1 || s.conflictSeen[0] != (monsterPair{target: 2, source: 3}) {
			t.Fatalf("faction conflict args = %v, want [[2 3]] (abs(-3)-1 = 2)", s.conflictSeen)
		}
		wantOneRoll(t, s.trapHitSeen, rollWant{}, "MonsterTrapHit")
	})
}

// TestCheckMissileColMonsterHitFlags covers the monster-hit flag block
// (cpp:503-507): a hit zeroes duration unless
// dontDeleteOnCollision and always sets _miHitFlag; a miss leaves the
// missile untouched.
func TestCheckMissileColMonsterHitFlags(t *testing.T) {
	tests := []struct {
		name         string
		rollHits     bool
		keep         bool
		wantDuration int
		wantHitFlag  bool
		wantSound    int
	}{
		// cpp:503-506: hit + !dontDeleteOnCollision -> duration = 0
		// (and the cpp:548 sound tail fires).
		{name: "hit deletes on collision", rollHits: true, keep: false,
			wantDuration: 0, wantHitFlag: true, wantSound: 1},
		// cpp:503-505: hit + dontDeleteOnCollision keeps flying.
		{name: "hit with keepOnCollision stays alive", rollHits: true, keep: true,
			wantDuration: testLifespan, wantHitFlag: true},
		// cpp:503 branch skipped entirely on a miss.
		{name: "miss leaves the missile alive", rollHits: false, keep: false,
			wantDuration: testLifespan, wantHitFlag: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newColSpies()
			s.trapHit = tc.rollHits // trap path: the simplest hit script
			w := colWorld(s)
			placeMonster(w, 1)
			m := colMissile()
			m.Source = -1

			// when
			res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
				testMinDamage, testMaxDamage, false, tc.keep, nil)

			// then
			if res.MonsterHit != tc.rollHits {
				t.Fatalf("MonsterHit = %v, want %v", res.MonsterHit, tc.rollHits)
			}
			if m.Duration != tc.wantDuration {
				t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
			}
			if m.HitFlag != tc.wantHitFlag {
				t.Fatalf("HitFlag = %v, want %v (cpp:506)", m.HitFlag, tc.wantHitFlag)
			}
			if s.hitSoundSeen != tc.wantSound {
				t.Fatalf("hit-sound fired %d times, want %d (cpp:548)", s.hitSoundSeen, tc.wantSound)
			}
		})
	}
}

// TestCheckMissileColPlayerRollDispatch verifies the player-roll
// dispatch (cpp:513-523): TARGET_MONSTERS vs another player
// rolls Plr2PlrMHit; a monster side rolls PlayerMHit carrying the source
// monster; TARGET_BOTH / trap roll PlayerMHit with a nil monster side
// and the death reason following sourceType (cpp:522).
func TestCheckMissileColPlayerRollDispatch(t *testing.T) {
	sourceOf := func(idx int) *int { return &idx }
	tests := []struct {
		name        string
		source      int
		caster      caster
		wantPlr2Plr bool
		wantPHit    bool
		wantSide    *int
		wantDeath   deathReason
	}{
		// cpp:514-516: TARGET_MONSTERS + different player id ->
		// Plr2PlrMHit.
		{name: "TARGET_MONSTERS hits another player via Plr2PlrMHit",
			source: 1, caster: targetMonsters, wantPlr2Plr: true},
		// cpp:517-519: monster side -> PlayerMHit(&monster,
		// deathReason::MonsterOrTrap).
		{name: "monster side rolls PlayerMHit with the source monster",
			source: 4, caster: targetPlayers, wantPHit: true,
			wantSide: sourceOf(4), wantDeath: DeathMonsterOrTrap},
		// cpp:522-523: TARGET_BOTH, player source ->
		// PlayerMHit(nullptr, deathReason::Player).
		{name: "TARGET_BOTH player source rolls PlayerMHit with DeathByPlayer",
			source: 0, caster: targetBoth, wantPHit: true,
			wantSide: nil, wantDeath: DeathByPlayer},
		// cpp:522-523: trap source -> PlayerMHit(nullptr,
		// deathReason::MonsterOrTrap).
		{name: "trap source rolls PlayerMHit with DeathMonsterOrTrap",
			source: -1, caster: targetBoth, wantPHit: true,
			wantSide: nil, wantDeath: DeathMonsterOrTrap},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newColSpies()
			s.plr2PlrHit = tc.wantPlr2Plr
			s.playerMHitHit = tc.wantPHit
			w := colWorld(s)
			placePlayer(w, 1) // dPlayer cell -> player index 0
			m := colMissile()
			m.Source = tc.source
			m.caster = int(tc.caster)

			// when
			res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
				testMinDamage, testMaxDamage, false, false, nil)

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
		})
	}
}

// TestCheckMissileColPlayerGate covers the player-cell gates: without
// onlyHitWalking the probe skips negative (walking) dPlayer cells
// (cpp:511 ignoreMovingPlayers + player.cpp:2044), and
// TARGET_MONSTERS never rolls against its own source player
// (cpp:515 player id == _misource guard).
func TestCheckMissileColPlayerGate(t *testing.T) {
	t.Run("walking cell skipped without onlyHitWalking", func(t *testing.T) {
		s := newColSpies()
		s.plr2PlrHit = true
		w := colWorld(s)
		placePlayer(w, -1) // negative = walking auxiliary spawn
		m := colMissile()
		m.Source = 1
		m.caster = int(targetMonsters)

		// when
		res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if res.PlayerHit || len(s.plr2PlrSeen) != 0 {
			t.Fatalf("walking cell rolled: hit = %v, plr2plr = %v", res.PlayerHit, s.plr2PlrSeen)
		}
		if m.Duration != testLifespan || m.HitFlag {
			t.Fatalf("missile mutated: Duration = %d, HitFlag = %v", m.Duration, m.HitFlag)
		}
	})

	t.Run("TARGET_MONSTERS skips its own source player", func(t *testing.T) {
		s := newColSpies()
		s.plr2PlrHit = true
		w := colWorld(s)
		placePlayer(w, 1) // player index 0 == m.Source below
		m := colMissile()
		m.Source = 0
		m.caster = int(targetMonsters)

		// when
		res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if res.PlayerHit || len(s.plr2PlrSeen) != 0 {
			t.Fatalf("self-source rolled: hit = %v, plr2plr = %v (cpp:515)",
				res.PlayerHit, s.plr2PlrSeen)
		}
		if m.Duration != testLifespan {
			t.Fatalf("Duration = %d, want untouched %d", m.Duration, testLifespan)
		}
	})
}

// TestCheckMissileColPlayerHitFlags covers the player-hit flag block
// (cpp:527-532) off the hellfire path: keepOnCollision and the
// plain delete, plus the blocked-report-without-hit case, where blocked
// travels out on the result but the missile still survives.
func TestCheckMissileColPlayerHitFlags(t *testing.T) {
	tests := []struct {
		name           string
		hit            bool
		blocked        bool
		keep           bool
		wantDuration   int
		wantHitFlag    bool
		wantResBlocked bool
	}{
		// cpp:527-532: hit + !gbIsHellfire-blocked + no keep ->
		// duration = 0, _miHitFlag = true.
		{name: "hit deletes on collision", hit: true,
			wantDuration: 0, wantHitFlag: true},
		// cpp:531-532: dontDeleteOnCollision keeps the missile.
		{name: "hit with keepOnCollision stays alive", hit: true, keep: true,
			wantDuration: testLifespan, wantHitFlag: true},
		// cpp:527: no isPlayerHit -> neither duration nor rotate
		// run, but the blocked out-param still reaches the caller.
		{name: "miss that reports blocked still leaves the missile alive",
			hit: false, blocked: true,
			wantDuration: testLifespan, wantResBlocked: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newColSpies()
			s.plr2PlrHit = tc.hit
			s.plr2PlrBlocked = tc.blocked
			w := colWorld(s)
			placePlayer(w, 1)
			m := colMissile()
			m.Source = 1
			m.caster = int(targetMonsters)

			// when
			res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
				testMinDamage, testMaxDamage, false, tc.keep, nil)

			// then
			if res.PlayerHit != tc.hit {
				t.Fatalf("PlayerHit = %v, want %v", res.PlayerHit, tc.hit)
			}
			if res.Blocked != tc.wantResBlocked {
				t.Fatalf("Blocked = %v, want %v", res.Blocked, tc.wantResBlocked)
			}
			if m.Duration != tc.wantDuration {
				t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
			}
			if m.HitFlag != tc.wantHitFlag {
				t.Fatalf("HitFlag = %v, want %v (cpp:532)", m.HitFlag, tc.wantHitFlag)
			}
		})
	}
}

// TestCheckMissileColHellfireBlockedShot covers cpp:528-530:
// gbIsHellfire && blocked rotates the missile 45° through the injected
// Rnd seam (rotateBlockedMissile, cpp:449-465) instead of
// deleting it; every other combination takes the plain delete path.
func TestCheckMissileColHellfireBlockedShot(t *testing.T) {
	// rotationTape offsets: tape[0] = 1 -> PickRandomlyAmong index 1 ->
	// +1 (right); tape[1] = 0 -> index 0 -> -1 (left).
	const (
		drawRight = 0
		drawLeft  = 1
	)
	tests := []struct {
		name         string
		firebolt     bool
		hellfire     bool
		blocked      bool
		keep         bool
		frame        int
		tapeFrom     int
		wantFrame    int
		wantDuration int
	}{
		// cpp:528-530: hellfire + blocked -> rotate, missile
		// survives (duration untouched).
		{name: "hellfire blocked arrow rotates right and survives",
			hellfire: true, blocked: true, frame: 5, tapeFrom: drawRight,
			wantFrame: 6, wantDuration: testLifespan},
		{name: "hellfire blocked arrow rotates left and survives",
			hellfire: true, blocked: true, frame: 5, tapeFrom: drawLeft,
			wantFrame: 4, wantDuration: testLifespan},
		// cpp:458-464: non-arrow graphics rotate the 0-indexed
		// frame group (_mimfnum) instead of _miAnimFrame.
		{name: "hellfire blocked firebolt rotates the frame group",
			firebolt: true, hellfire: true, blocked: true, frame: 5, tapeFrom: drawRight,
			wantFrame: 6, wantDuration: testLifespan},
		// cpp:529: rotation wins over dontDeleteOnCollision.
		{name: "keepOnCollision still rotates but survives",
			hellfire: true, blocked: true, keep: true, frame: 5, tapeFrom: drawRight,
			wantFrame: 6, wantDuration: testLifespan},
		// cpp:531: without gbIsHellfire the plain delete runs.
		{name: "diablo blocked shot deletes",
			blocked: true, frame: 5, wantFrame: 5, wantDuration: 0},
		// cpp:531: hellfire without blocked is a plain delete too.
		{name: "hellfire unblocked deletes",
			hellfire: true, frame: 5, wantFrame: 5, wantDuration: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newColSpies()
			s.plr2PlrHit = true
			s.plr2PlrBlocked = tc.blocked
			w := colWorld(s)
			w.IsHellfire = tc.hellfire
			w.Rnd = rotationTapeAt(t, tc.tapeFrom)
			placePlayer(w, 1)
			m := colMissile()
			if tc.firebolt {
				m.AnimType = 1 // Fireball graphic (misdat.h:37; non-arrow branch)
				m.MFNum = tc.frame
			} else {
				m.AnimFrame = tc.frame
			}
			m.Source = 1
			m.caster = int(targetMonsters)

			// when
			res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
				testMinDamage, testMaxDamage, false, tc.keep, nil)

			// then
			if !res.PlayerHit || !m.HitFlag {
				t.Fatalf("hit flags = (PlayerHit %v, HitFlag %v), want true/true", res.PlayerHit, m.HitFlag)
			}
			if res.Blocked != tc.blocked {
				t.Fatalf("Blocked = %v, want %v", res.Blocked, tc.blocked)
			}
			gotFrame := m.AnimFrame
			if tc.firebolt {
				gotFrame = m.MFNum
			}
			if gotFrame != tc.wantFrame {
				t.Fatalf("rotated frame = %d, want %d (cpp:529)", gotFrame, tc.wantFrame)
			}
			if m.Duration != tc.wantDuration {
				t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
			}
		})
	}
}

// TestCheckMissileColWallBranch covers the wall / breakable-object
// branch (cpp:536-544): a missile-blocking tile sets
// WallBlocked, breaks a breakable object on the tile, zeroes duration
// unless keepOnCollision, and CLEARS _miHitFlag (an assignment, not an
// accumulate — cpp:544), overriding earlier hit branches.
func TestCheckMissileColWallBranch(t *testing.T) {
	tests := []struct {
		name         string
		tile         world.TileProperties
		obj          *world.WorldObject
		keep         bool
		wantWall     bool
		wantDuration int
		wantBreaks   int
	}{
		// cpp:863: the SOL BlockMissile bit blocks the missile.
		{name: "missile-blocking wall stops the missile", tile: world.TileBlockMissile,
			wantWall: true, wantDuration: 0},
		// cpp:863 checks BlockMissile only — a plain solid tile
		// (TileProperties::Solid, no BlockMissile) does NOT stop
		{name: "solid-but-open-to-missiles tile passes", tile: world.TileSolid,
			wantDuration: testLifespan},
		{name: "open floor passes", tile: world.TileNone,
			wantDuration: testLifespan},
		// cpp:543-544: keepOnCollision keeps the duration while
		// HitFlag still clears.
		{name: "keepOnCollision survives the wall", tile: world.TileBlockMissile, keep: true,
			wantWall: true, wantDuration: testLifespan},
		// cpp:537-539: breakable object on the wall tile breaks.
		{name: "breakable object on a wall tile breaks",
			tile: world.TileBlockMissile, obj: &world.WorldObject{Breakable: true},
			wantWall: true, wantDuration: 0, wantBreaks: 1},
		// cpp:538 IsBreakable() false: the tile still blocks but
		// nothing breaks.
		{name: "plain object blocks without breaking",
			obj: &world.WorldObject{}, wantWall: true, wantDuration: 0},
		// cpp:866-869: _oMissFlag lets the missile pass through.
		{name: "object allowing missiles passes",
			obj: &world.WorldObject{MissFlag: true}, wantDuration: testLifespan},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newColSpies()
			w := colWorld(s)
			setTileProps(w, tc.tile)
			if tc.obj != nil {
				placeObject(w, *tc.obj)
			}
			m := colMissile()
			m.HitFlag = true // preset: cpp:544 must CLEAR, not keep

			// when
			res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
				testMinDamage, testMaxDamage, false, tc.keep, nil)

			// then
			if res.WallBlocked != tc.wantWall {
				t.Fatalf("WallBlocked = %v, want %v", res.WallBlocked, tc.wantWall)
			}
			if s.breakSeen != tc.wantBreaks {
				t.Fatalf("BreakObjectMissile fired %d times, want %d (cpp:539)",
					s.breakSeen, tc.wantBreaks)
			}
			if m.Duration != tc.wantDuration {
				t.Fatalf("Duration = %d, want %d", m.Duration, tc.wantDuration)
			}
			if tc.wantWall && m.HitFlag {
				t.Fatal("HitFlag = true, want cleared by the wall branch (cpp:544)")
			}
			if !tc.wantWall && !m.HitFlag {
				t.Fatal("HitFlag = false on a pass-through tile, want the preset untouched (wall branch must not run)")
			}
		})
	}
}

// TestIsMissileBlockedByTile ports the wall helper directly
// (Source/cpp:857-870): out of bounds blocks, the SOL
// BlockMissile bit blocks, and a present object blocks unless
// _oMissFlag allows missiles through.
func TestIsMissileBlockedByTile(t *testing.T) {
	t.Run("out of bounds blocks", func(t *testing.T) {
		g := world.NewGrid()

		// when / then
		if !IsMissileBlockedByTile(g, world.MAXDUNY, probeX) {
			t.Fatal("out-of-bounds probe = false, want true (cpp:859-861)")
		}
	})

	t.Run("BlockMissile SOL bit blocks", func(t *testing.T) {
		w := &hitWorld{Grid: world.NewGrid()}
		setTileProps(w, world.TileBlockMissile)

		// when / then
		if !IsMissileBlockedByTile(w.Grid, probeX, probeY) {
			t.Fatal("blocked tile = false, want true (cpp:863)")
		}
	})

	t.Run("plain solid floor does not block missiles", func(t *testing.T) {
		w := &hitWorld{Grid: world.NewGrid()}
		setTileProps(w, world.TileSolid)

		// when / then
		if IsMissileBlockedByTile(w.Grid, probeX, probeY) {
			t.Fatal("solid floor = true, want false (cpp:863 BlockMissile only)")
		}
	})

	t.Run("object without _oMissFlag blocks", func(t *testing.T) {
		w := &hitWorld{Grid: world.NewGrid()}
		placeObject(w, world.WorldObject{SolidFlag: true})

		// when / then
		if !IsMissileBlockedByTile(w.Grid, probeX, probeY) {
			t.Fatal("object = false, want true (cpp:869)")
		}
	})

	t.Run("object with _oMissFlag passes", func(t *testing.T) {
		w := &hitWorld{Grid: world.NewGrid()}
		placeObject(w, world.WorldObject{MissFlag: true})

		// when / then
		if IsMissileBlockedByTile(w.Grid, probeX, probeY) {
			t.Fatal("missile-permeable object = true, want false (cpp:866-869)")
		}
	})

	t.Run("empty tile passes", func(t *testing.T) {
		g := world.NewGrid()

		// when / then
		if IsMissileBlockedByTile(g, probeX, probeY) {
			t.Fatal("open floor = true, want false")
		}
	})
}

// TestCheckMissileColHitSoundTail covers the sound tail
// (cpp:545-548): the hook fires iff duration reached 0 —
// including entry at 0 with no branch touching the missile — and stays
// silent while the missile lives. The row's hitSound != SfxID::None
// decision rides inside the hook (audio is engine-owned).
func TestCheckMissileColHitSoundTail(t *testing.T) {
	t.Run("zero duration at entry fires the hook", func(t *testing.T) {
		s := newColSpies()
		w := colWorld(s)
		m := colMissile()
		m.Duration = 0

		// when
		CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if s.hitSoundSeen != 1 {
			t.Fatalf("hit-sound fired %d times, want 1 (cpp:548)", s.hitSoundSeen)
		}
	})

	t.Run("alive missile stays silent", func(t *testing.T) {
		s := newColSpies()
		w := colWorld(s)
		m := colMissile()

		// when
		CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if s.hitSoundSeen != 0 {
			t.Fatalf("hit-sound fired %d times, want 0", s.hitSoundSeen)
		}
	})

	t.Run("wall-zeroed missile fires the hook", func(t *testing.T) {
		s := newColSpies()
		w := colWorld(s)
		setTileProps(w, world.TileBlockMissile)
		m := colMissile()

		// when
		CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if m.Duration != 0 || s.hitSoundSeen != 1 {
			t.Fatalf("Duration = %d, sound = %d; want 0/1 (tail runs after the wall branch)",
				m.Duration, s.hitSoundSeen)
		}
	})

	t.Run("kill-zeroed missile fires the hook", func(t *testing.T) {
		s := newColSpies()
		s.trapHit = true
		w := colWorld(s)
		placeMonster(w, 1)
		m := colMissile()
		m.Source = -1

		// when
		CheckMissileCol(w, &m, colProbe(), DamagePhysical,
			testMinDamage, testMaxDamage, false, false, nil)

		// then
		if m.Duration != 0 || s.hitSoundSeen != 1 {
			t.Fatalf("Duration = %d, sound = %d; want 0/1 (tail runs after the monster branch)",
				m.Duration, s.hitSoundSeen)
		}
	})
}

// TestCheckMissileColBranchOrdering proves the statement order of the
// port (cpp:484-548): monster, player and wall branches on one
// tile all fire, the wall branch's _miHitFlag = false (cpp:544)
// lands after the hit branches' true, and the sound tail runs last.
func TestCheckMissileColBranchOrdering(t *testing.T) {
	// given
	s := newColSpies()
	s.trapHit = true
	s.playerMHitHit = true
	w := colWorld(s)
	placeMonster(w, 1)
	placePlayer(w, 1)
	setTileProps(w, world.TileBlockMissile)
	placeObject(w, world.WorldObject{Breakable: true})
	m := colMissile()
	m.Source = -1 // trap: monster branch rolls MonsterTrapHit...
	m.caster = int(targetBoth)
	// ...and the player branch rolls PlayerMHit(nullptr, MonsterOrTrap)
	// (cpp:522-523).

	// when
	res := CheckMissileCol(w, &m, colProbe(), DamagePhysical,
		testMinDamage, testMaxDamage, false, false, nil)

	// then
	if !res.MonsterHit || !res.PlayerHit || !res.WallBlocked {
		t.Fatalf("result = %+v, want all three branches taken", res)
	}
	if len(s.trapHitSeen) != 1 {
		t.Fatalf("MonsterTrapHit dispatches = %d, want 1", len(s.trapHitSeen))
	}
	wantOnePlayerMHit(t, s.playerMHitSeen, nil, DeathMonsterOrTrap)
	if s.breakSeen != 1 {
		t.Fatalf("BreakObjectMissile fired %d times, want 1", s.breakSeen)
	}
	if m.Duration != 0 {
		t.Fatalf("Duration = %d, want 0", m.Duration)
	}
	if m.HitFlag {
		t.Fatal("HitFlag = true, want the wall branch's false to win (cpp:544)")
	}
	if s.hitSoundSeen != 1 {
		t.Fatalf("hit-sound fired %d times, want 1 (tail runs last)", s.hitSoundSeen)
	}
}

// colProbe is the probe position (cpp:479 Point position).
func colProbe() tilePosition {
	return tilePosition{X: probeX, Y: probeY}
}

// colMissile seeds the missile under test: a TARGET_MONSTERS Arrow at
// tile (5, 5) with 10 ticks of life left. Tests override Source/caster
// per branch under test.
func colMissile() Missile {
	m := Missile{
		Type:     Arrow,
		Source:   0,
		caster:   int(targetMonsters),
		Dist:     testDist,
		Duration: testLifespan,
	}
	m.SetTraveledFromTile(probeX, probeY)
	return m
}

// newColSpies defaults the monster-target gate to open (the common case
// behind cpp:489).
func newColSpies() *colSpies {
	return &colSpies{targetMonster: true}
}

// colWorld builds an open-floor hitWorld with the spies wired: SOL entry
// 0 is plain walkable (world.NewGrid), no entities, non-hellfire.
func colWorld(s *colSpies) *hitWorld {
	return &hitWorld{Grid: world.NewGrid(), hitCallbacks: s.callbacks()}
}

// placeMonster writes a dMonster occupancy cell (cpp:485;
// 1-based, negative = walking spawn).
func placeMonster(w *hitWorld, id int16) {
	w.DMonster[probeY][probeX] = id
}

// placePlayer writes a dPlayer occupancy cell (PlayerAtPosition,
// Source/player.cpp:2040-2049).
func placePlayer(w *hitWorld, id int16) {
	w.DPlayer[probeY][probeX] = id
}

// setTileProps points the probe tile's dPiece at a scripted SOL entry
// (TileHasAny, tile_properties.cpp:11-41 — the grid seam
// IsMissileBlockedByTile reads, cpp:863).
func setTileProps(w *hitWorld, props world.TileProperties) {
	const scriptedPiece = 1378 // top SOL slot of the LoadLevelSOLData range
	w.Grid.DPiece[probeY][probeX] = scriptedPiece
	w.Grid.SOLData[scriptedPiece] = props
}

// placeObject writes an object onto the probe tile (1-based dObject,
// FindObjectAtPosition — cpp:537 and 866).
func placeObject(w *hitWorld, obj world.WorldObject) {
	w.Grid.Objects = append(w.Grid.Objects, obj)
	w.Grid.DObject[probeY][probeX] = len(w.Grid.Objects)
}

// fixedRoll is the argument pack every fixture call forwards
// (cpp:496/498/516/519/523 pass-through); tests prove
// damageType/shifted/keepOnCollision pass through via rollWant.
func fixedRoll() rollCall {
	return rollCall{
		minDamage:   testMinDamage,
		maxDamage:   testMaxDamage,
		dist:        testDist,
		missileType: Arrow,
	}
}

// wantOneRoll asserts exactly one dispatch: the fixed fixture pack plus
// the case's named want fields (cpp:496/498/516/519/523).
func wantOneRoll(t *testing.T, got []rollCall, want rollWant, label string) {
	t.Helper()
	if len(got) != 1 {
		t.Fatalf("%s: %d dispatches recorded, want exactly 1 (%+v)", label, len(got), got)
	}
	wantCall := fixedRoll()
	wantCall.damageType = want.damageType
	wantCall.shifted = want.shifted
	wantCall.keepOnCollision = want.keepOnCollision
	if got[0] != wantCall {
		t.Fatalf("%s: args = %+v, want %+v", label, got[0], wantCall)
	}
}

// wantNoRolls asserts a dispatch never fired.
func wantNoRolls(t *testing.T, got []rollCall, label string) {
	t.Helper()
	if len(got) != 0 {
		t.Fatalf("%s: fired %d times, want 0 (%+v)", label, len(got), got)
	}
}

// wantOnePlayerMHit asserts exactly one PlayerMHit dispatch with the
// scripted argument pack, monster side, and death reason
// (cpp:519/523: &monster or nullptr + deathReason).
func wantOnePlayerMHit(t *testing.T, got []playerMHitArgs, monsterSide *int, deathReason deathReason) {
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
	if call.deathReason != deathReason {
		t.Fatalf("PlayerMHit deathReason = %d, want %d (cpp:522)", call.deathReason, deathReason)
	}
}

// callbacks wires the spies into the hitWorld seam (collision.go
// hitCallbacks).
func (s *colSpies) callbacks() hitCallbacks {
	return hitCallbacks{
		MonsterTargeted: func(mid int16, onlyHitWalking *Direction) bool {
			s.targetedSeen = append(s.targetedSeen, mid)
			s.onlyHitWalkingSeen = onlyHitWalking
			return s.targetMonster
		},
		MonsterFactionConflict: func(targetMonster, sourceMonster int) bool {
			s.conflictSeen = append(s.conflictSeen, monsterPair{targetMonster, sourceMonster})
			return s.conflict
		},
		MonsterTrapHit: func(minDamage, maxDamage, dist int,
			missileType missileID, damageType damageType,
			shifted, keepOnCollision bool) bool {
			s.trapHitSeen = append(s.trapHitSeen,
				rollCall{minDamage: minDamage, maxDamage: maxDamage, dist: dist,
					missileType: missileType, damageType: damageType, shifted: shifted,
					keepOnCollision: keepOnCollision})
			return s.trapHit
		},
		MonsterMHit: func(minDamage, maxDamage, dist int,
			missileType missileID, damageType damageType,
			shifted, keepOnCollision bool) bool {
			s.monsterMHitSeen = append(s.monsterMHitSeen,
				rollCall{minDamage: minDamage, maxDamage: maxDamage, dist: dist,
					missileType: missileType, damageType: damageType, shifted: shifted,
					keepOnCollision: keepOnCollision})
			return s.monsterMHit
		},
		Plr2PlrMHit: func(minDamage, maxDamage, dist int,
			missileType missileID, damageType damageType, shifted bool) (bool, bool) {
			s.plr2PlrSeen = append(s.plr2PlrSeen,
				rollCall{minDamage: minDamage, maxDamage: maxDamage, dist: dist,
					missileType: missileType, damageType: damageType, shifted: shifted})
			return s.plr2PlrHit, s.plr2PlrBlocked
		},
		PlayerMHit: func(monsterSide *int, minDamage, maxDamage, dist int,
			missileType missileID, damageType damageType, shifted bool,
			deathReason deathReason) (bool, bool) {
			s.playerMHitSeen = append(s.playerMHitSeen, playerMHitArgs{
				roll: rollCall{minDamage: minDamage, maxDamage: maxDamage, dist: dist,
					missileType: missileType, damageType: damageType, shifted: shifted},
				monsterSide: monsterSide,
				deathReason: deathReason,
			})
			return s.playerMHitHit, s.playerMHitBlocked
		},
		BreakObjectMissile: func() { s.breakSeen++ },
		PlayHitSound:       func() { s.hitSoundSeen++ },
	}
}

// monsterMHitKeepSeen reports whether the last MonsterMHit dispatch was a
// keepOnCollision probe (slice 3a-ii explosion re-probes,
// cpp:2911/3111).
func (s *colSpies) monsterMHitKeepSeen() bool {
	return len(s.monsterMHitSeen) > 0 &&
		s.monsterMHitSeen[len(s.monsterMHitSeen)-1].keepOnCollision
}
