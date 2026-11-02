package persist

// NoLight is NO_LIGHT (Source/engine/lighting_defs.hpp:9); SaveMonster stores
// it as 0 for vanilla compatibility (loadsave.cpp:1613-1617).
const NoLight int8 = -1

// MonsterSaveSize is the SaveMonster record length (Source/loadsave.cpp:1508-1620):
// 74 fields with upstream's alignment skips.
const MonsterSaveSize = 216

// MonsterDerived carries the values SaveMonster does not read from the record
// itself but from the monster-data tables, i.e. monster.level/exp/toHit/
// toHitSpecial for the current difficulty (loadsave.cpp:1583-1610). They are
// injected so persist keeps no dependency on the data tables.
type MonsterDerived struct {
	Level        int8
	Experience   uint16
	ToHit        uint8
	ToHitSpecial uint8
}

// MonsterWalking carries the four walking displacements SaveMonster computes
// when the monster walks (loadsave.cpp:1532-1546). They are injected because
// the calculation lives on ActorPosition, not on the record.
type MonsterWalking struct {
	OffsetX   int32
	OffsetY   int32
	VelocityX int32
	VelocityY int32
	Offset2X  int32
	Offset2Y  int32
}

// MonsterRecord is the serialized form of struct Monster
// (monster.h:213-491) as SaveMonster writes it. Field names mirror the
// upstream members; the record keeps the raw int widths upstream truncates to.
type MonsterRecord struct {
	LevelType        int32
	Mode             int32
	Goal             uint8
	GoalVar1         int32
	GoalVar2         int32
	GoalVar3         int32
	PathCount        uint8
	TileX            int32
	TileY            int32
	FutureX          int32
	FutureY          int32
	OldX             int32
	OldY             int32
	Direction        int32
	Enemy            int32
	EnemyPosX        uint8
	EnemyPosY        uint8
	TicksPerFrame    int32
	TickCounter      int32
	NumberOfFrames   int32
	CurrentFrame     int32
	IsInvalid        uint32
	Var1             int32
	Var2             int32
	Var3             int32
	TempX            int32
	TempY            int32
	MaxHitPoints     int32
	HitPoints        int32
	AI               uint8
	Intelligence     uint8
	Flags            uint32
	ActiveForTicks   uint8
	LastX            int32
	LastY            int32
	RndItemSeed      uint32
	AiSeed           uint32
	UniqueType       uint8
	UniqTrans        uint8
	CorpseID         int8
	WhoHit           int8
	MinDamage        uint8
	MaxDamage        uint8
	MinDamageSpecial uint8
	MaxDamageSpecial uint8
	ArmorClass       uint8
	Resistance       uint16
	TalkMsg          int32
	Leader           uint8
	LeaderRelation   uint8
	PackSize         uint8
	LightID          int8
}

// Encode returns the SaveMonster image, mixing the derived stats and walking
// displacements into the record fields upstream writes.
func (m MonsterRecord) Encode(derived MonsterDerived, walking MonsterWalking) []byte {
	var out recordWriter

	m.encodeHead(&out)
	m.encodePosition(&out, walking)
	m.encodeAnim(&out)
	m.encodeTail(&out, derived, walking)

	return out.buf
}

// lightByte is the stored light id: NO_LIGHT maps to 0, everything else is
// written as-is (loadsave.cpp:1613-1617).
func (m MonsterRecord) lightByte() int8 {
	if m.LightID == NoLight {
		return 0
	}

	return m.LightID
}

// encodeHead writes levelType through the position block
// (loadsave.cpp:1510-1531).
func (m MonsterRecord) encodeHead(out *recordWriter) {
	out.i32(m.LevelType)
	out.i32(m.Mode)
	out.u8(m.Goal)
	out.zeroes(3)
	out.i32(m.GoalVar1)
	out.i32(m.GoalVar2)
	out.i32(m.GoalVar3)
	out.zeroes(4)
	out.u8(m.PathCount)
	out.zeroes(3)
	out.i32(m.TileX)
	out.i32(m.TileY)
	out.i32(m.FutureX)
	out.i32(m.FutureY)
	out.i32(m.OldX)
	out.i32(m.OldY)
}

// encodePosition writes the walking displacements, direction and enemy block
// (loadsave.cpp:1546-1556).
func (m MonsterRecord) encodePosition(out *recordWriter, walking MonsterWalking) {
	out.i32(walking.OffsetX)
	out.i32(walking.OffsetY)
	out.i32(walking.VelocityX)
	out.i32(walking.VelocityY)
	out.i32(m.Direction)
	out.i32(m.Enemy)
	out.u8(m.EnemyPosX)
	out.u8(m.EnemyPosY)
	out.zeroes(2)
}

// encodeAnim writes the animation block through the temp position and second
// displacement (loadsave.cpp:1558-1580).
func (m MonsterRecord) encodeAnim(out *recordWriter) {
	out.zeroes(4) // _mAnimData pointer
	out.i32(m.TicksPerFrame)
	out.i32(m.TickCounter)
	out.i32(m.NumberOfFrames)
	out.i32(m.CurrentFrame + 1)
	out.zeroes(4) // _meflag
	out.u32(m.IsInvalid)
	out.i32(m.Var1)
	out.i32(m.Var2)
	out.i32(m.Var3)
	out.i32(m.TempX)
	out.i32(m.TempY)
}

// encodeTail writes the second displacement, then the HP, AI and stat tail
// (loadsave.cpp:1578-1618).
func (m MonsterRecord) encodeTail(out *recordWriter, derived MonsterDerived, walking MonsterWalking) {
	out.i32(walking.Offset2X)
	out.i32(walking.Offset2Y)
	out.zeroes(4) // _mVar8
	out.i32(m.MaxHitPoints)
	out.i32(m.HitPoints)
	out.u8(m.AI)
	out.u8(m.Intelligence)
	out.zeroes(2) // Alignment
	out.u32(m.Flags)
	out.u8(m.ActiveForTicks)
	out.zeroes(3) // Alignment
	out.zeroes(4) // Unused
	out.i32(m.LastX)
	out.i32(m.LastY)
	out.u32(m.RndItemSeed)
	out.u32(m.AiSeed)
	out.zeroes(4) // Unused
	out.u8(m.UniqueType + 1)
	out.u8(m.UniqTrans)
	out.u8(uint8(m.CorpseID))
	out.u8(uint8(m.WhoHit))
	out.u8(uint8(derived.Level))
	out.zeroes(1) // Alignment
	out.le16(derived.Experience)
	out.u8(derived.ToHit)
	out.u8(m.MinDamage)
	out.u8(m.MaxDamage)
	out.u8(derived.ToHitSpecial)
	out.u8(m.MinDamageSpecial)
	out.u8(m.MaxDamageSpecial)
	out.u8(m.ArmorClass)
	out.zeroes(1) // Alignment
	out.le16(m.Resistance)
	out.zeroes(2) // Alignment
	out.i32(m.TalkMsg)
	out.u8(m.Leader)
	out.u8(m.LeaderRelation)
	out.u8(m.PackSize)
	out.u8(uint8(m.lightByte()))
}
