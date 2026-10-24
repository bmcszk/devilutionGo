// Package missiles_test: CheckMissileCol gate/branch coverage
// (split from collision_test.go by behavior: gate flags, hit flags,
// hellfire blocked shot, wall branch, IsMissileBlockedByTile,
// hit-sound tail, branch ordering).
//
// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package missiles

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/world"
)

type monsterHitFlagsRow = struct {
	name         string
	rollHits     bool
	keep         bool
	wantDuration int
	wantHitFlag  bool
	wantSound    int
}

// TestCheckMissileColMonsterHitFlags covers the monster-hit flag block
// (cpp:503-507): a hit zeroes duration unless
// dontDeleteOnCollision and always sets _miHitFlag; a miss leaves the
// missile untouched.
func TestCheckMissileColMonsterHitFlags(t *testing.T) {
	tests := []monsterHitFlagsRow{
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
			monsterHitFlagsCase(t, tc)
		})
	}
}

func monsterHitFlagsCase(t *testing.T, tc monsterHitFlagsRow) {
	t.Helper()
	s := newColSpies()
	s.trapHit = tc.rollHits // trap path: the simplest hit script
	w := colWorld(s)
	placeMonster(w, 1)
	m := colMissile()
	m.Source = -1

	// when
	res := CheckMissileCol(w, &m, colProbe(), missileColProbe{
	damageType: DamagePhysical,
	minDamage: testMinDamage,
	maxDamage: testMaxDamage,
	shifted: false,
	keep: keepOnCollision(tc.keep),
	onlyHitWalking: nil,
})

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
}

// TestCheckMissileColPlayerRollDispatch verifies the player-roll
// dispatch (cpp:513-523): TARGET_MONSTERS vs another player
// rolls Plr2PlrMHit; a monster side rolls PlayerMHit carrying the source
// monster; TARGET_BOTH / trap roll PlayerMHit with a nil monster side
// and the death reason following sourceType (cpp:522).

func TestCheckMissileColPlayerGate(t *testing.T) {
	t.Run("walking cell skipped without onlyHitWalking", func(t *testing.T) {
		playerGateWalkingSkipped(t)
	})

	t.Run("TARGET_MONSTERS skips its own source player", func(t *testing.T) {
		playerGateSelfSource(t)
	})
}

func playerGateWalkingSkipped(t *testing.T) {
	t.Helper()
	s := newColSpies()
	s.plr2PlrHit = true
	w := colWorld(s)
	placePlayer(w, -1) // negative = walking auxiliary spawn
	m := colMissile()
	m.Source = 1
	m.caster = int(targetMonsters)

	// when
	res := CheckMissileCol(w, &m, colProbe(), missileColProbe{
	damageType: DamagePhysical,
	minDamage: testMinDamage,
	maxDamage: testMaxDamage,
	shifted: false,
	keep: false,
	onlyHitWalking: nil,
})

	// then
	if res.PlayerHit || len(s.plr2PlrSeen) != 0 {
		t.Fatalf("walking cell rolled: hit = %v, plr2plr = %v", res.PlayerHit, s.plr2PlrSeen)
	}
	if m.Duration != testLifespan || m.HitFlag {
		t.Fatalf("missile mutated: Duration = %d, HitFlag = %v", m.Duration, m.HitFlag)
	}
}

func playerGateSelfSource(t *testing.T) {
	t.Helper()
	s := newColSpies()
	s.plr2PlrHit = true
	w := colWorld(s)
	placePlayer(w, 1) // player index 0 == m.Source below
	m := colMissile()
	m.Source = 0
	m.caster = int(targetMonsters)

	// when
	res := CheckMissileCol(w, &m, colProbe(), missileColProbe{
	damageType: DamagePhysical,
	minDamage: testMinDamage,
	maxDamage: testMaxDamage,
	shifted: false,
	keep: false,
	onlyHitWalking: nil,
})

	// then
	if res.PlayerHit || len(s.plr2PlrSeen) != 0 {
		t.Fatalf("self-source rolled: hit = %v, plr2plr = %v (cpp:515)",
			res.PlayerHit, s.plr2PlrSeen)
	}
	if m.Duration != testLifespan {
		t.Fatalf("Duration = %d, want untouched %d", m.Duration, testLifespan)
	}
}

// TestCheckMissileColPlayerHitFlags covers the player-hit flag block
// (cpp:527-532) off the hellfire path: keepOnCollision and the
// plain delete, plus the blocked-report-without-hit case, where blocked
// travels out on the result but the missile still survives.
func TestCheckMissileColPlayerHitFlags(t *testing.T) {
	tests := []playerHitFlagsRow{
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
			playerHitFlagsCase(t, tc)
		})
	}
}

type playerHitFlagsRow = struct {
	name           string
	hit            bool
	blocked        bool
	keep           bool
	wantDuration   int
	wantHitFlag    bool
	wantResBlocked bool
}

func playerHitFlagsCase(t *testing.T, tc playerHitFlagsRow) {
	t.Helper()
	s := newColSpies()
	s.plr2PlrHit = tc.hit
	s.plr2PlrBlocked = tc.blocked
	w := colWorld(s)
	placePlayer(w, 1)
	m := colMissile()
	m.Source = 1
	m.caster = int(targetMonsters)

	// when
	res := CheckMissileCol(w, &m, colProbe(), missileColProbe{
	damageType: DamagePhysical,
	minDamage: testMinDamage,
	maxDamage: testMaxDamage,
	shifted: false,
	keep: keepOnCollision(tc.keep),
	onlyHitWalking: nil,
})

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
}

// TestCheckMissileColHellfireBlockedShot covers cpp:528-530:
// gbIsHellfire && blocked rotates the missile 45° through the injected
// Rnd seam (rotateBlockedMissile, cpp:449-465) instead of
// deleting it; every other combination takes the plain delete path.
type hellfireBlockedRow = struct {
	name         string
	firebolt     bool
	hellfire     bool
	blocked      bool
	keep         bool
	frame        int
	tapeFrom     int
	wantFrame    int
	wantDuration int
}

func hellfireBlockedCase(t *testing.T, tc hellfireBlockedRow) {
	t.Helper()
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
	res := CheckMissileCol(w, &m, colProbe(), missileColProbe{
	damageType: DamagePhysical,
	minDamage: testMinDamage,
	maxDamage: testMaxDamage,
	shifted: false,
	keep: keepOnCollision(tc.keep),
	onlyHitWalking: nil,
})

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
}

func hellfireBlockedTable() []hellfireBlockedRow {
	// rotationTape offsets: tape[0] = 1 -> PickRandomlyAmong index 1 ->
	// +1 (right); tape[1] = 0 -> index 0 -> -1 (left).
	const (
		drawRight = 0
		drawLeft  = 1
	)
	return []hellfireBlockedRow{
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
}

// TestCheckMissileColHellfireBlockedShot covers cpp:528-530:
// gbIsHellfire && blocked rotates the missile 45° through the injected
// Rnd seam (rotateBlockedMissile, cpp:449-465) instead of
// deleting it; every other combination takes the plain delete path.
func TestCheckMissileColHellfireBlockedShot(t *testing.T) {
	for _, tc := range hellfireBlockedTable() {
		t.Run(tc.name, func(t *testing.T) {
			hellfireBlockedCase(t, tc)
		})
	}
}

// TestCheckMissileColWallBranch covers the wall / breakable-object
// branch (cpp:536-544): a missile-blocking tile sets
// WallBlocked, breaks a breakable object on the tile, zeroes duration
// unless keepOnCollision, and CLEARS _miHitFlag (an assignment, not an
// accumulate — cpp:544), overriding earlier hit branches.
type wallBranchRow = struct {
	name         string
	tile         world.TileProperties
	obj          *world.WorldObject
	keep         bool
	wantWall     bool
	wantDuration int
	wantBreaks   int
}

func wallBranchCase(t *testing.T, tc wallBranchRow) {
	t.Helper()
	s := newColSpies()
	w := colWorld(s)
	setTileProps(w, tc.tile)
	if tc.obj != nil {
		placeObject(w, *tc.obj)
	}
	m := colMissile()
	m.HitFlag = true // preset: cpp:544 must CLEAR, not keep

	// when
	res := CheckMissileCol(w, &m, colProbe(), missileColProbe{
	damageType: DamagePhysical,
	minDamage: testMinDamage,
	maxDamage: testMaxDamage,
	shifted: false,
	keep: keepOnCollision(tc.keep),
	onlyHitWalking: nil,
})

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
	outcome := wallBranchPass
	if tc.wantWall {
		outcome = wallBranchHit
	}
	expectWallHitFlag(t, outcome, wallBranchPreset(m.HitFlag))
}

func wallBranchTable() []wallBranchRow {
	return []wallBranchRow{
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
}

// TestCheckMissileColWallBranch covers the wall / breakable-object
// branch (cpp:536-544): a missile-blocking tile sets
// WallBlocked, breaks a breakable object on the tile, zeroes duration
// unless keepOnCollision, and CLEARS _miHitFlag (an assignment, not an
// accumulate — cpp:544), overriding earlier hit branches.
func TestCheckMissileColWallBranch(t *testing.T) {
	for _, tc := range wallBranchTable() {
		t.Run(tc.name, func(t *testing.T) {
			wallBranchCase(t, tc)
		})
	}
}

// TestIsMissileBlockedByTile ports the wall helper directly
// (Source/cpp:857-870): out of bounds blocks, the SOL
// BlockMissile bit blocks, and a present object blocks unless
// _oMissFlag allows missiles through.
func TestIsMissileBlockedByTile(t *testing.T) {
	t.Run("out of bounds blocks", func(t *testing.T) {
		isBlockedByTileCase(t, world.MAXDUNY, probeX, nil, tileBlocksMissile, "cpp: 859-861")
})

	t.Run("BlockMissile SOL bit blocks", func(t *testing.T) {
		isBlockedByTileCase(t, probeX, probeY,
			func(w *hitWorld) { setTileProps(w, world.TileBlockMissile) },
			tileBlocksMissile, "cpp: 863")
})

	t.Run("plain solid floor does not block missiles", func(t *testing.T) {
		isBlockedByTileCase(t, probeX, probeY,
			func(w *hitWorld) { setTileProps(w, world.TileSolid) },
			tileAllowsMissile, "cpp: 863 BlockMissile only")
})

	t.Run("object without _oMissFlag blocks", func(t *testing.T) {
		isBlockedByTileCase(t, probeX, probeY, nil, tileBlocksMissile, "cpp:869",
			world.WorldObject{SolidFlag: true})
})

	t.Run("object with _oMissFlag passes", func(t *testing.T) {
		isBlockedByTileCase(t, probeX, probeY, nil, tileAllowsMissile, "cpp:866-869",
			world.WorldObject{MissFlag: true})
})

	t.Run("empty tile passes", func(t *testing.T) {
		isBlockedByTileCase(t, probeX, probeY, nil, tileAllowsMissile, "")
	})
}

// blockedTileOutcome is the expected IsMissileBlockedByTile result,
// named instead of a bare control flag.
type blockedTileOutcome int

const (
	tileBlocksMissile blockedTileOutcome = iota
	tileAllowsMissile
)

func isBlockedByTileCase(
	t *testing.T, x, y int, setup func(w *hitWorld),
	outcome blockedTileOutcome, why string, objs ...world.WorldObject,
) {
	t.Helper()
	wantBlocked := outcome == tileBlocksMissile
	w := &hitWorld{Grid: world.NewGrid()}
	if setup != nil {
		setup(w)
	}
	for _, o := range objs {
		placeObject(w, o)
	}

	// when / then
	if got := IsMissileBlockedByTile(w.Grid, x, y); got != wantBlocked {
		t.Fatalf("blocked = %v, want %v (%s)", got, wantBlocked, why)
	}
}

// TestCheckMissileColHitSoundTail covers the sound tail
// (cpp:545-548): the hook fires iff duration reached 0 —
// including entry at 0 with no branch touching the missile — and stays
// silent while the missile lives. The row's hitSound != SfxID::None
// decision rides inside the hook (audio is engine-owned).
func TestCheckMissileColHitSoundTail(t *testing.T) {
	t.Run("zero duration at entry fires the hook", func(t *testing.T) {
		hitSoundTailCase(t, 0, 1, "cpp:548", nil)
	})

	t.Run("alive missile stays silent", func(t *testing.T) {
		hitSoundTailCase(t, testLifespan, 0, "", nil)
	})

	t.Run("wall-zeroed missile fires the hook", func(t *testing.T) {
		hitSoundTailCase(t, testLifespan, 1, "tail runs after the wall branch",
			func(_ *colSpies, w *hitWorld, _ *Missile) {
				setTileProps(w, world.TileBlockMissile)
			})
	})

	t.Run("kill-zeroed missile fires the hook", func(t *testing.T) {
		hitSoundTailCase(t, testLifespan, 1, "tail runs after the monster branch",
			func(s *colSpies, w *hitWorld, m *Missile) {
				s.trapHit = true
				w.hitCallbacks = s.callbacks()
				placeMonster(w, 1)
				m.Source = -1
			})
	})
}

func hitSoundTailCase(t *testing.T, duration, wantSounds int, why string,
	setup func(_ *colSpies, w *hitWorld, m *Missile)) {
	t.Helper()
	s := newColSpies()
	w := colWorld(s)
	m := colMissile()
	m.Duration = duration
	if setup != nil {
		setup(s, w, &m)
	}

	// when
	CheckMissileCol(w, &m, colProbe(), missileColProbe{
	damageType: DamagePhysical,
	minDamage: testMinDamage,
	maxDamage: testMaxDamage,
	shifted: false,
	keep: false,
	onlyHitWalking: nil,
})

	// then
	if s.hitSoundSeen != wantSounds {
		t.Fatalf("hit-sound fired %d times, want %d (%s)", s.hitSoundSeen, wantSounds, why)
	}
	if duration == testLifespan && why == "" {
		// alive-missile case: also assert the missile was not zeroed
		if m.Duration != testLifespan {
			t.Fatalf("Duration = %d, want %d", m.Duration, testLifespan)
		}
	}
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
	res := CheckMissileCol(w, &m, colProbe(), missileColProbe{
	damageType: DamagePhysical,
	minDamage: testMinDamage,
	maxDamage: testMaxDamage,
	shifted: false,
	keep: false,
	onlyHitWalking: nil,
})

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
		t.Fatal("HitFlag = true, want the wall branch's false to win (cpp: 544)")
}
	if s.hitSoundSeen != 1 {
		t.Fatalf("hit-sound fired %d times, want 1 (tail runs last)", s.hitSoundSeen)
	}
}

// expectWallHitFlag asserts HitFlag semantics per branch outcome: the wall
// branch clears it (cpp:544), a pass-through tile leaves the preset untouched.
func expectWallHitFlag(t *testing.T, outcome wallBranchOutcome, preset wallBranchPreset) {
	t.Helper()
	if outcome == wallBranchHit && bool(preset) {
		t.Fatal("HitFlag = true, want cleared by the wall branch (cpp: 544)")
}
	if outcome == wallBranchPass && !bool(preset) {
		t.Fatal("HitFlag = false on a pass-through tile, want the preset untouched (wall branch must not run)")
	}
}

// wallBranchPreset carries the HitFlag value before the wall branch ran,
// instead of a bare control flag.
type wallBranchPreset bool

// wallBranchOutcome names the two HitFlag expectations instead of a bare
// control flag.
type wallBranchOutcome int

const (
	wallBranchHit  wallBranchOutcome = iota // wall branch ran, HitFlag cleared
	wallBranchPass                          // pass-through tile, preset untouched
)
