// Package missiles: the Guardian sentry's bodies.
package missiles

import "github.com/bmcszk/devilutionGo/internal/game/world"

// Guardian frame groups (missiles.h:63-67).
const (
	// GuardianFrameStart is GuardianFrame::Start.
	GuardianFrameStart = 0
	// GuardianFrameIdle is GuardianFrame::Idle.
	GuardianFrameIdle = 1
	// GuardianFrameAttack is GuardianFrame::Attack.
	GuardianFrameAttack = 2
)

// SpellGuardian is SpellID::Guardian (tables/spelldat.h:45): the spell level
// GuardianTryFireAt fires its Firebolts at (missiles.cpp:729).
const SpellGuardian = 13

// guardianArc are the 48 offsets the guardian sweeps every 16 ticks, walked
// backwards along the lines of VisionCrawlTable (missiles.cpp:3830-3873).
var guardianArc = [48][2]int{
	{6, 0},
	{5, 0},
	{4, 0},
	{3, 0},
	{2, 0},
	{1, 0},
	{6, 1},
	{5, 1},
	{4, 1},
	{3, 1},
	{6, 2},
	{2, 1},
	{5, 2},
	{6, 3},
	{4, 2},
	{5, 3},
	{3, 2},
	{1, 1},
	{6, 4},
	{6, 5},
	{5, 4},
	{4, 3},
	{2, 2},
	{5, 5},
	{4, 4},
	{3, 3},
	{6, 6},
	{5, 6},
	{4, 5},
	{3, 4},
	{2, 3},
	{4, 6},
	{3, 5},
	{2, 4},
	{1, 2},
	{3, 6},
	{2, 5},
	{1, 3},
	{0, 1},
	{2, 6},
	{1, 4},
	{1, 5},
	{1, 6},
	{0, 2},
	{0, 3},
	{0, 6},
	{0, 5},
	{0, 4},
}

// ProcessGuardian ports ProcessGuardian (missiles.cpp:3810-3889): the idle
// frame returns on the fire cooldown or when the attack anim runs out
// (cpp:3814-3818), every 16th tick sweeps the 48-offset arc in its four
// mirrored forms and fires at the first monster in reach (cpp:3822-3876), the
// light ramps down over the last frames (cpp:3878-3885) and the expiry drops
// the missile and its light (cpp:3886-3889).
func ProcessGuardian(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3812

	if m.Var2 > 0 { // cpp:3814-3816
		m.Var2--
	}

	if m.Duration == m.Var1 || (m.MFNum == GuardianFrameAttack && m.Var2 == 0) { // cpp:3817-3819
		m.setFrameGroup(GuardianFrameIdle)
	}

	if m.Duration%16 == 0 { // cpp:3823
		guardianSweep(w, m)
	}

	guardianFadeTail(w, m)

	if m.Duration == 0 {
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}

	PutMissile(w, m) // cpp:3891
}

// guardianFadeTail ports the frame-14 animation flip and the light ramp
// (missiles.cpp:3878-3885).
func guardianFadeTail(w *HitWorld, m *Missile) {
	if m.Duration == 14 { // cpp:3878-3881
		m.setFrameGroup(GuardianFrameStart)
		m.AnimFrame = 15
		m.AnimAdd = -1
	}

	m.Var3 += m.AnimAdd // cpp:3883

	if m.Var3 > 15 {
		m.Var3 = 15
	} else if m.Var3 > 0 {
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, uint8(m.Var3))
	}
}

// guardianSweep ports the arc loop (missiles.cpp:3874-3879): the first monster
// any mirrored offset finds ends the sweep.
func guardianSweep(w *HitWorld, m *Missile) {
	for _, offset := range guardianArc {
		x, y := offset[0], offset[1]
		if tryGuardianFire(w, m, m.Position.Tile.X+x, m.Position.Tile.Y+y) ||
			tryGuardianFire(w, m, m.Position.Tile.X+y, m.Position.Tile.Y+x) || // flipXY
			tryGuardianFire(w, m, m.Position.Tile.X+x, m.Position.Tile.Y-y) || // flipY
			tryGuardianFire(w, m, m.Position.Tile.X-x, m.Position.Tile.Y+y) { // flipX
			return
		}
	}
}

// tryGuardianFire ports GuardianTryFireAt (missiles.cpp:709-734): a monster
// the guardian can see — line clear, not a player minion, still alive — takes a
// Firebolt whose damage is GenerateRnd(10) + level/2 + 1 scaled by the spell
// level, and the guardian drops into its 3-tick attack frame.
func tryGuardianFire(w *HitWorld, m *Missile, x, y int) bool {
	target := TilePosition{X: x, Y: y}
	if !LineClear(func(px, py int) bool { // cpp:713
		return !IsMissileBlockedByTile(w.Grid, px, py)
	}, m.Position.Tile, target) {
		return false
	}

	// Deviation (documented): upstream reads dMonster[target] unguarded
	// (cpp:715), and the LineClear swap above lets an out-of-dungeon offset
	// reach here, which is an out-of-bounds read in C++. The guard skips what
	// upstream would have read as garbage.
	if !world.InDungeonBounds(target.X, target.Y) {
		return false
	}

	idx := int(w.DMonster[target.Y][target.X]) - 1 // cpp:715-717
	if idx < 0 {
		return false
	}

	if w.ProcessCallbacks.MonsterIsPlayerMinion(idx) || w.ProcessCallbacks.MonsterHasNoLife(idx) {
		return false
	}

	level := w.AddSource.playerStats(m.Source).Level
	dam := scaleSpellEffect(w.ProcessCallbacks.GenerateRnd(10)+level/2+1, m.SpellLevel) // cpp:724-726

	spellLevel := w.ProcessCallbacks.PlayerSpellLevel(m.Source, SpellGuardian)
	w.ProcessCallbacks.SpawnMissile(m.Position.Tile, target, GetDirection(m.Position.Tile, target), // cpp:728-729
		Firebolt, int(TargetMonsters), m.Source, dam, spellLevel, m)

	m.setFrameGroup(GuardianFrameAttack) // cpp:730-731
	m.Var2 = 3

	return true
}
