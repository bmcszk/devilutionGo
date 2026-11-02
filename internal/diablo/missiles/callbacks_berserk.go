// Package missiles: the monster-store seams the Berserk add body needs.
package missiles

// MonsterDamage is the monster damage quad AddBerserk raises
// (missiles.cpp:1334-1337).
type MonsterDamage struct {
	Min        int
	Max        int
	MinSpecial int
	MaxSpecial int
}

// berserkCallbacks is embedded in ProcessCallbacks so AddBerserk can read and
// write the monster store without widening a file past its limits
// (missiles.cpp:1303-1344).
type berserkCallbacks struct {
	// MonsterFlags reports Monster::flags (monster.h:44-60,
	// missiles.cpp:1313-1314).
	MonsterFlags func(idx int) int
	// SetMonsterFlags writes Monster::flags (missiles.cpp:1333).
	SetMonsterFlags func(idx, flags int)
	// MonsterIsUnique reports Monster::isUnique (missiles.cpp:1316).
	MonsterIsUnique func(idx int) bool
	// MonsterResistance reports Monster::resistance
	// (tables/monstdat.h:73-79, missiles.cpp:1320-1323).
	MonsterResistance func(idx int) int
	// MonsterDamage reads the damage quad (missiles.cpp:1334-1337).
	MonsterDamage func(idx int) MonsterDamage
	// SetMonsterDamage writes the damage quad (missiles.cpp:1334-1337).
	SetMonsterDamage func(idx int, damage MonsterDamage)
	// SetMonsterLight writes Monster::lightId (missiles.cpp:1341).
	SetMonsterLight func(idx, lightID int)
}
