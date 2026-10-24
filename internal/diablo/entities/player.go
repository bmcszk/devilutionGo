package entities

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/animation"
	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

type enginePoint = struct{ X, Y int }

type missilesMissile = missiles.Missile

// Player ports struct Player (player.h:211-914). Field order follows the
// upstream declaration; upstream methods and sprite pointers are behavior
// or asset surface and stay out of the data port. Upstream stores Items by
// value in fixed arrays; Go ports them as slice headers over Item values.
type Player struct {
	// _pName[PlayerNameLength] (player.h:216).
	Name [PlayerNameLength]byte
	// InvBody[NUM_INVLOC] (player.h:217).
	InvBody [NumInvLoc]Item
	// InvList[InventoryGridCells] (player.h:218).
	InvList [InventoryGridCells]Item
	// SpdList[MaxBeltItems] (player.h:219).
	SpdList [MaxBeltItems]Item
	// HoldItem (player.h:220) — cursor-held item.
	HoldItem Item

	LightID int

	NumInv       int
	Strength     int
	BaseStr      int
	Magic        int
	BaseMag      int
	Dexterity    int
	BaseDex      int
	Vitality     int
	BaseVit      int
	StatPts      int
	DamageMod    int
	HPBase       int
	MaxHPBase    int
	HitPoints    int
	MaxHP        int
	HPPer        int
	ManaBase     int
	MaxManaBase  int
	Mana         int
	MaxMana      int
	ManaPer      int
	IMinDam      int
	IMaxDam      int
	IAC          int
	IBonusDam    int
	IBonusToHit  int
	IBonusAC     int
	IBonusDamMod int
	IGetHit      int
	IEnAc        int
	IFMinDam     int
	IFMaxDam     int
	ILMinDam     int
	ILMaxDam     int
	Experience   uint32
	Mode         PlrMode
	// Walkpath[MaxPathLengthPlayer] (player.h:261).
	Walkpath  [MaxPathLengthPlayer]int8
	PlrActive bool
	// DestAction (player.h:263, action_id).
	DestAction int8
	DestParam1 int
	DestParam2 int
	DestParam3 int
	DestParam4 int
	Gold       int

	// AnimInfo (player.h:262): current animation state.
	AnimInfo animation.Info
	// AnimationFrameCounts ports AnimationData[enum_size<player_graphic>]
	// (player.h:286): frame counts per player_graphic (Stand..Block, 9).
	AnimationFrameCounts [9]int8
	// InvGrid[InventoryGridCells] (player.h:298).
	InvGrid [InventoryGridCells]int8

	PlrLevel        uint8
	PlrIsOnSetLevel bool
	// Position (player.h:302, ActorPosition, engine/actor_position.hpp:11).
	Position ActorPosition
	// Direction faced by player (player.h:303, direction enum).
	Direction uint8
	// HeroClass (player.h:305 via tables/playerdat.hpp:19).
	Class uint8

	// ExecutedSpell (player.h:318, SpellCastInfo player.h:202).
	ExecutedSpell  SpellCastInfo
	InventorySpell int32
	SpellFrom      int8
	RSpell         int32
	RSplType       uint8
	SBkSpell       int32
	// SplLvl[64] (player.h:326).
	SplLvl [SpellLevels]uint8
	// ISpells/MemSpells/AblSpells/ScrlSpells bitmasks (player.h:328-334).
	ISpells    uint64
	MemSpells  uint64
	AblSpells  uint64
	ScrlSpells uint64
	SpellFlags uint8
	// SplHotKey[NumHotkeys] (player.h:336).
	SplHotKey [NumHotkeys]int32
	// SplTHotKey[NumHotkeys] (player.h:337).
	SplTHotKey [NumHotkeys]uint8

	BlockFlag   bool
	Invincible  bool
	LightRad    int8
	LvlChanging bool

	ArmorClass    int8
	MagResist     int8
	FireResist    int8
	LghtResist    int8
	InfraFlag     bool
	TempDirection uint8

	// LvlVisited[NUMLEVELS] (player.h:352).
	LvlVisited  [NumLevels]bool
	SLvlVisited [NumLevels]bool

	OilType           int8
	TownWarps         uint8
	DungMsgs          uint8
	LvlLoad           uint8
	ManaShield        bool
	DungMsgs2         uint8
	OriginalCathedral bool
	DiabloKillLevel   uint8
	Reflections       uint16
	// DamAcFlags (player.h:364, ItemSpecialEffectHf itemdat.h:419).
	DamAcFlags uint8
	// IFlags (player.h:272, ItemSpecialEffect itemdat.h:386).
	IFlags uint32
}

// ActorPosition ports struct ActorPosition (engine/actor_position.hpp:11-32):
// five world tile positions.
type ActorPosition struct {
	Tile   Point
	Future Point
	Last   Point
	Old    Point
	Temp   Point
}

// SpellCastInfo ports struct SpellCastInfo (player.h:202-209).
type SpellCastInfo struct {
	SpellID    int32
	SpellType  uint8
	SpellFrom  int8
	SpellLevel int
}

// PlrMode mirrors enum PLR_MODE (player.h:109).
type PlrMode uint8

const (
	ModeStand PlrMode = iota
	ModeWalk
	ModeAttack
	ModeHitRecovery
	ModeBlock
	ModeSpell
	ModeCast
	ModeDeath
	ModeDead
)
