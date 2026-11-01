// Package missiles: the Rhino (charging monster) bodies.
package missiles

// MonsterGraphic sheet ids AddRhino picks (monster.h:113-120: Stand, Walk,
// Attack, GotHit, Death, Special).
const (
	monsterGraphicWalk    = 1
	monsterGraphicAttack  = 2
	monsterGraphicSpecial = 5
)

// Monster type ordinals AddRhino switches on (tables/monstdat.h): enum
// _monster_id starts at MT_NZOMBIE = 0, so the ordinal is the monstdat.tsv data
// row index.
const (
	MonsterTypeHorned = 64
	MonsterTypeMudRun = 65
	MonsterTypeFrostC = 66
	MonsterTypeObLord = 67
	MonsterTypeNSnake = 89
	MonsterTypeRSnake = 90
	MonsterTypeBSnake = 91
	MonsterTypeGSnake = 92
)

// monsterAISnake is MonsterAIID::Snake (tables/monstdat.h:47) — the charge that
// steps twice per tick (missiles.cpp:3747-3752).
const monsterAISnake = 24

// rhinoVelocity is the 18-pixel velocity AddRhino sets (missiles.cpp:2274).
const rhinoVelocity = 18

// rhinoDuration is the 256-tick life InitMissileAnimationFromMonster sets
// (missiles.cpp:2260).
const rhinoDuration = 256

// rhinoSnakeFrame is the attack frame a snake charge starts on
// (missiles.cpp:2277).
const rhinoSnakeFrame = 7

// AddRhino ports AddRhino (missiles.cpp:2264-2282): the sheet follows the
// monster type (Special for the horned family, Attack for the snakes, Walk
// otherwise, cpp:2268-2273), the velocity is 18, the animation comes from the
// monster's own sheet (cpp:2275), snakes start on frame 7 (cpp:2276-2277) and
// a unique carries its light id (cpp:2278-2280).
func AddRhino(w *HitWorld, m *Missile, p *AddParameter) {
	monster := m.Source                                   // cpp:2266
	monsterType := w.AddSource.monsterStats(monster).Type // cpp:2266-2273

	UpdateMissileVelocity(m, p.Dst, rhinoVelocity)                                     // cpp:2274
	initMissileAnimationFromMonster(w, m, p.Midir, monster, rhinoGraphic(monsterType)) // cpp:2275

	if isSnakeType(monsterType) { // cpp:2276-2277
		m.AnimFrame = rhinoSnakeFrame
	}

	if w.AddSource.monsterStats(monster).IsUnique { // cpp:2278-2280
		m.Lid = w.AddSource.monsterStats(monster).LightID
	}

	PutMissile(w, m) // cpp:2281
}

// ProcessRhino ports ProcessRhino (missiles.cpp:3734-3770): a monster that left
// the charge ends the missile (cpp:3739-3742); otherwise the missile moves, the
// snake variant steps twice (cpp:3747-3752) and a blocked landing (or a snake
// that ran out of attack frames) converts the missile back into the monster
// (cpp:3755-3759). A clear landing writes the monster's position and occupy
// flag and moves the light of a unique (cpp:3762-3769).
func ProcessRhino(w *HitWorld, m *Missile) {
	monster := m.Source                                               // cpp:3736-3737
	if w.ProcessCallbacks.MonsterMode(monster) != monsterModeCharge { // cpp:3739-3742
		m.DelFlag = true

		return
	}

	UpdateMissilePos(w, m)     // cpp:3743
	prevPos := m.Position.Tile // cpp:3744

	w.DMonster[prevPos.Y][prevPos.X] = 0        // cpp:3745
	w.ProcessCallbacks.MarkWalkabilityChanged() // cpp:3746

	snakePos, snake := rhinoStep(w, m, monster) // cpp:3747-3753
	newPos := m.Position.Tile                   // cpp:3754

	blocked := rhinoLandingBlocked(w, monster, newPos) // cpp:3755
	if !blocked && snake {                             // cpp:3755-3759 snake half
		blocked = snakeChargeStops(w, m, monster, snakePos)
	}

	if blocked {
		w.ProcessCallbacks.MissToMonst(m, prevPos) // cpp:3756
		m.DelFlag = true

		return
	}

	w.ProcessCallbacks.SetMonsterTile(monster, newPos)   // cpp:3762-3764
	w.ProcessCallbacks.OccupyTile(monster, newPos, true) // cpp:3765
	if w.AddSource.monsterStats(monster).IsUnique {      // cpp:3766-3767
		w.ProcessCallbacks.ChangeLightXY(m.Lid, newPos)
	}

	MoveMissilePos(w, m) // cpp:3768
	PutMissile(w, m)     // cpp:3769
}

// rhinoStep ports the advance of one tick (missiles.cpp:3747-3753) and reports
// the snake's twice-stepped tile.
func rhinoStep(w *HitWorld, m *Missile, monster int) (TilePosition, bool) {
	if w.ProcessCallbacks.MonsterAI(monster) != monsterAISnake { // cpp:3747
		m.Position.Traveled = addTraveled(m.Position.Traveled, m.Position.Velocity, 1) // cpp:3752
		UpdateMissilePos(w, m)

		return TilePosition{}, false
	}

	// The snake steps twice: +2 velocity, read the tile, then -1 back so the
	// next tick starts from the normal advance (cpp:3748-3751).
	m.Position.Traveled = addTraveled(m.Position.Traveled, m.Position.Velocity, 2)
	UpdateMissilePos(w, m)
	snakePos := m.Position.Tile
	m.Position.Traveled = subVelocity(m.Position.Traveled, m.Position.Velocity)
	UpdateMissilePos(w, m)

	return snakePos, true
}

// rhinoLandingBlocked is the landing-tile test of the charge (missiles.cpp:3755).
func rhinoLandingBlocked(w *HitWorld, monster int, newPos TilePosition) bool {
	return !w.ProcessCallbacks.IsTileAvailable(monster, newPos.X, newPos.Y)
}

// snakeChargeStops is the snake half of the landing test (missiles.cpp:3755-3759).
func snakeChargeStops(w *HitWorld, m *Missile, monster int, snakePos TilePosition) bool {
	if !w.ProcessCallbacks.IsTileAvailable(monster, snakePos.X, snakePos.Y) {
		return true
	}

	return m.AnimFrame >= m.AnimLen
}

// initMissileAnimationFromMonster ports InitMissileAnimationFromMonster
// (missiles.cpp:2237-2261): the direction, the monster sheet's rate and frame
// count, the frame step, a zeroed var1/var2, the light flag and 256 ticks.
// Upstream also copies the sprite width and width2 (cpp:2246-2253) from the
// decoded CLX list; those are renderer state this struct does not carry
// (headless builds leave them 0 anyway, cpp:2242).
func initMissileAnimationFromMonster(w *HitWorld, m *Missile, midir Direction, monster, graphic int) {
	anim := w.AddSource.monsterAnimation(monster, graphic) // cpp:2239

	setDirection(m, midir)     // cpp:2240
	m.AnimDelay = anim.Rate    // cpp:2254
	m.AnimLen = anim.Frames    // cpp:2255
	m.AnimAdd = 1              // cpp:2256
	m.Var1 = 0                 // cpp:2257
	m.Var2 = 0                 // cpp:2258
	m.LightFlag = true         // cpp:2259
	m.Duration = rhinoDuration // cpp:2260
}

// rhinoGraphic ports the two type sets of AddRhino (missiles.cpp:2268-2273).
func rhinoGraphic(monsterType int) int {
	switch {
	case isAnyMonsterType(monsterType, MonsterTypeHorned, MonsterTypeMudRun, MonsterTypeFrostC, MonsterTypeObLord):
		return monsterGraphicSpecial // cpp:2269-2270
	case isSnakeType(monsterType):
		return monsterGraphicAttack // cpp:2271-2272
	}

	return monsterGraphicWalk // cpp:2268
}

// isSnakeType is the snake family test (missiles.cpp:2271,2276,3757).
func isSnakeType(monsterType int) bool {
	return isAnyMonsterType(monsterType, MonsterTypeNSnake, MonsterTypeRSnake, MonsterTypeBSnake, MonsterTypeGSnake)
}

// isAnyMonsterType is upstream's IsAnyOf over a type id list.
func isAnyMonsterType(monsterType int, set ...int) bool {
	for _, candidate := range set {
		if monsterType == candidate {
			return true
		}
	}

	return false
}

// addTraveled is traveled += velocity * factor (missiles.cpp:3748,3752).
func addTraveled(traveled, velocity PositionDisplacement, factor int32) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: traveled.DeltaX + velocity.DeltaX*factor,
		DeltaY: traveled.DeltaY + velocity.DeltaY*factor,
	}
}

// subVelocity is traveled -= velocity (missiles.cpp:3751).
func subVelocity(traveled, velocity PositionDisplacement) PositionDisplacement {
	return PositionDisplacement{
		DeltaX: traveled.DeltaX - velocity.DeltaX,
		DeltaY: traveled.DeltaY - velocity.DeltaY,
	}
}
