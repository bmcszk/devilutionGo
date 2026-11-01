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
