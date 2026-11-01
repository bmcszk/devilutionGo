package entities

import "github.com/bmcszk/devilutionGo/internal/diablo/animation"

const (
	GoalNone MonsterGoal = iota
	GoalNormal
	GoalRetreat
	GoalHealing
	GoalMove
	GoalAttack
	GoalInquiring
	GoalShout
	GoalGolem
)

// UniqueMonsterNone is UniqueMonsterType::None (monster.h:74), i.e. uint8 -1;
// Monster.IsUnique compares against it (monster.h:458-461).
const UniqueMonsterNone uint8 = 255

// Monster ports struct Monster (monster.h:213-491). Field order follows the
// upstream declaration; upstream methods (toHit, level, name, …) are
// behavior and stay out of the data port. Upstream carries a
// unique_ptr<uint8_t[]> TRN and an AnimationInfo; animation state ports as
// animation.Info, the TRN bytes as a slice header.
type Monster struct {
	// uniqueMonsterTRN (monster.h:214): palette remap bytes for uniques.
	UniqueMonsterTRN []uint8
	// AnimInfo (monster.h:217): current animation state.
	AnimInfo animation.Info

	MaxHitPoints int
	HitPoints    int
	Flags        uint32
	// RndItemSeed: item drops on death (monster.h:222).
	RndItemSeed uint32
	// AiSeed: AI behavior/sync in multiplayer (monster.h:224).
	AiSeed uint32

	GolemToHit uint16
	Resistance uint16
	// TalkMsg (monster.h:228, _speech_id textdat.h:18).
	TalkMsg int16

	GoalVar1 int16
	GoalVar2 int8
	GoalVar3 int8

	Var1 int16
	Var2 int16
	Var3 int8

	// Position (monster.h:237, ActorPosition engine/actor_position.hpp:11).
	Position ActorPosition

	Goal uint8
	// EnemyPosition: enemy's future position (monster.h:243).
	EnemyPosition Point
	LevelType     uint8
	Mode          uint8
	PathCount     uint8
	// Direction faced by monster (monster.h:248, direction enum).
	Direction uint8
	// Enemy: target index into players or monsters (monster.h:250).
	Enemy     uint8
	IsInvalid bool
	// Ai (monster.h:252, MonsterAIID monstdat.h:22).
	AI int8
	// Intelligence: aggression driver (monster.h:258).
	Intelligence   uint8
	ActiveForTicks uint8
	// UniqueType (monster.h:261, UniqueMonsterType monster.h:59).
	UniqueType uint8
	UniqTrans  uint8
	CorpseID   int8
	WhoHit     int8

	MinDamage             uint8
	MaxDamage             uint8
	MinDamageSpecial      uint8
	MaxDamageSpecial      uint8
	ArmorClass            uint8
	ReducePlayerStrength  uint8
	ReducePlayerMagic     uint8
	ReducePlayerDexterity uint8
	ReducePlayerVitality  uint8
	ReducePlayerMaxHP     uint8
	ReducePlayerMaxMana   uint8
	// Leader: NoLeader == 0xFF sentinel (monster.h:286-287).
	Leader         uint8
	LeaderRelation uint8
	PackSize       uint8
	LightID        int8
}

// MonsterGoal mirrors enum class MonsterGoal (monster.h:121).
type MonsterGoal uint8
