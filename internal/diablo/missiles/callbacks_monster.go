// Package missiles: the monster-store seams ProcessCallbacks carries.
package missiles

// monsterStoreCallbacks is embedded in ProcessCallbacks so the monster rows can
// read and write the monster store without widening one file past its limits.
type monsterStoreCallbacks struct {
	// MonsterMode reports Monster::mode (monster.h:37-58).
	MonsterMode func(idx int) int
	// SetMonsterMode writes Monster::mode (monster.cpp).
	SetMonsterMode func(idx, mode int)
	// SetMonsterPetrified writes Monster::animInfo.isPetrified
	// (monster.cpp:4925-4929 Monster::petrify).
	SetMonsterPetrified func(idx int, petrified bool)
	// MonsterHitPoints reports Monster::hitPoints (missiles.cpp:3698,3711).
	MonsterHitPoints func(idx int) int
	// AddCorpse ports the whole AddCorpse(monster.position.tile, stonendx,
	// monster.direction) call of the shatter (missiles.cpp:3715,
	// dead.cpp:93-96): the tile, the direction and the stonendx corpse record
	// all live in the corpse subsystem (dead.cpp:20,77).
	AddCorpse func(monster int)
}

// monsterCallbacks is the monster-store surface the missile rows reach: the
// Stone Curse set plus the Rhino position set.
type monsterCallbacks struct {
	monsterStoreCallbacks
	monsterPositionCallbacks
}

// monsterPositionCallbacks is the second half of the monster-store seams, the
// ones the Rhino row needs: the AI and position writes plus the two monster
// subsystem entry points it calls.
type monsterPositionCallbacks struct {
	// MonsterAI reports monster.ai (monster.h:252); the Rhino add picks the
	// snake double-step on MonsterAIID::Snake (missiles.cpp:3747).
	MonsterAI func(idx int) int
	// SetMonsterTile writes monster.position.future, .old and .tile
	// (missiles.cpp:3762-3764).
	SetMonsterTile func(idx int, position TilePosition)
	// OccupyTile ports Monster::occupyTile (monster.cpp, missiles.cpp:3765).
	OccupyTile func(idx int, position TilePosition, isMoving bool)
	// MissToMonst ports MissToMonst (monster.cpp) — the collision that turns
	// the charge back into the monster (missiles.cpp:3756). It owns the
	// monster mode writes, M_StartStand/M_StartHit and the attack on whatever
	// stands on the old tile.
	MissToMonst func(m *Missile, position TilePosition)
	// MarkWalkabilityChanged ports MarkWalkabilityChanged (missiles.cpp:3746).
	MarkWalkabilityChanged func()
}
