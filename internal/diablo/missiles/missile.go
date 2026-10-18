// Package missiles: Missile struct and container row
// (Source/missiles.h:60-160, devilGo-mwlm slice 1).
package missiles

// MissileID mirrors enum class MissileID (Source/tables/spelldat.h:98-167):
// int8_t ordinals into the misdat table — GetMissileData indexes
// MissilesData[underlying(missileId)] (Source/tables/misdat.cpp:451), so
// ordinal == base misdat.tsv row index (verified: spelldat.h names 1:1 with
// rows 1-68 of assets/txtdata/missiles/misdat.tsv; LastDiablo = DiabloApocalypse).
type MissileID int8

const (
	Arrow MissileID = iota
	Firebolt
	Guardian
	Phasing
	NovaBall
	FireWall
	Fireball
	LightningControl
	Lightning
	MagmaBallExplosion
	TownPortal
	FlashBottom
	FlashTop
	ManaShield
	FlameWave
	ChainLightning
	ChainBall // unused
	BloodHit  // unused
	BoneHit   // unused
	MetalHit  // unused
	Rhino
	MagmaBall
	ThinLightningControl
	ThinLightning
	BloodStar
	BloodStarExplosion
	Teleport
	FireArrow
	DoomSerpents // unused
	FireOnly     // unused
	StoneCurse
	BloodRitual  // unused
	Invisibility // unused
	Golem
	Etherealize
	Spurt // unused
	ApocalypseBoom
	Healing
	FireWallControl
	Infravision
	Identify
	FlameWaveControl
	Nova
	Rage
	Apocalypse
	ItemRepair
	StaffRecharge
	TrapDisarm
	Inferno
	InfernoControl
	FireMan
	Krull
	ChargedBolt
	HolyBolt
	Resurrect
	Telekinesis
	LightningArrow
	Acid
	AcidSplat
	AcidPuddle
	HealOther
	Elemental
	ResurrectBeam
	BoneSpirit
	WeaponExplosion
	RedPortal
	DiabloApocalypseBoom
	DiabloApocalypse
	LastDiablo = DiabloApocalypse // Source/tables/spelldat.h:168
	// Hellfire-region ids (spelldat.h:169-184) — explicit ordinals:
	// the enum restarts after LastDiablo upstream, so iota is broken
	// from here (verified against spelldat.h:169-184; only the
	// constants slice 3a-beta needs are named).
	Mana            MissileID = 68  // spelldat.h:169
	Magi            MissileID = 69  // spelldat.h:170
	LightningWall   MissileID = 70  // spelldat.h:171 (put-missile flag bit)
	OrangeFlare     MissileID = 98  // spelldat.h:99
	BlueFlare       MissileID = 99  // spelldat.h:100
	RedFlare        MissileID = 100 // spelldat.h:101
	YellowFlare     MissileID = 101 // spelldat.h:102
	BlueFlare2      MissileID = 102 // spelldat.h:103
	YellowExplosion MissileID = 103 // spelldat.h:104
	RedExplosion    MissileID = 104 // spelldat.h:105
	BlueExplosion   MissileID = 105 // spelldat.h:106
	BlueExplosion2  MissileID = 106 // spelldat.h:107
	OrangeExplosion MissileID = 107 // spelldat.h:108
)

// Missile is the Go port of struct Missile (Source/missiles.h:60-160),
// slice-1 field set: identity, source, del flag, animation counters and
// the 16.16 position block (MissilePosition, missiles.h:24-49).
type Missile struct {
	// Type of projectile (missiles.h:62 _mitype).
	Type MissileID
	// Position (missiles.h:92 MissilePosition position).
	Position Position
	// The direction of the missile — direction enum
	// (missiles.h:96 _mimfnum).
	MFNum int
	// Spell level (missiles.h:99 _mispllvl).
	SpellLevel int
	// Indicate whether the missile should be deleted
	// (missiles.h:100 _miDelFlag).
	DelFlag bool
	// MissileGraphicID (missiles.h:101 _miAnimType).
	AnimType int
	// Tick length of each frame in the current animation
	// (missiles.h:107 _miAnimDelay).
	AnimDelay int
	// Number of frames in current animation (missiles.h:108 _miAnimLen).
	AnimLen int
	// Increases by one each game tick, counting how close we are to
	// AnimDelay (missiles.h:114 _miAnimCnt).
	AnimCnt int
	// Added to AnimFrame each processed tick (missiles.h:115 _miAnimAdd).
	AnimAdd int
	// Current frame of animation + 1 (missiles.h:116 _miAnimFrame).
	AnimFrame int
	// Source id: player index, monster index, or -1 for traps
	// (missiles.h:121 _misource; IsTrap missiles.h:150-152).
	Source int
	// mienemy_type caster class (missiles.h:122 _micaster).
	Caster int
	// Source player, when the caster is player-side and Source != -1
	// (missiles.h:153-160 sourcePlayer).
	SourcePlayer int
	// _midist: used for arrows to measure distance travelled (increases
	// by 1 each game tick). Higher value is a penalty for accuracy
	// calculation when hitting an enemy (missiles.h:124 _midist, int).
	Dist int
	// Time to live for the missile in game ticks; once 0, the missile
	// will be marked for deletion via DelFlag (missiles.h:118-119
	// duration, int).
	Duration int
	// Hit flag (missiles.h:123 _miHitFlag, bool).
	HitFlag bool
	// PreFlag renders the missile over walls (missiles.h:116
	// _miPreFlag); PutMissile raises MissilePreFlag when set
	// (missiles.cpp:209-210).
	PreFlag bool
	// Damage of the missile, rolled at launch (missiles.h:122 _midam).
	Dam int
	// Light id for the missile's light source, NoLight = none
	// (missiles.h:125 _mlid).
	Lid int
	// Per-type scratch registers (missiles.h:127-128 var1/var2):
	// ProcessGenericProjectile tracks the light tile here
	// (missiles.cpp:3008-3010).
	Var1 int
	Var2 int
	// Hash of the dMonster^dPlayer cell the last on-tile
	// CheckMissileCol probed — the slow-missile multi-hit guard
	// (missiles.h:138 lastCollisionTargetHash, int16).
	LastCollisionTargetHash int16
	// MovementDistribution ports MissileMovementDistribution
	// (Source/tables/misdat.h:101-117) — how a hit affects movement
	// (checkTile lambda, missiles.cpp:650-652). Set from the row's
	// movementDistribution at load (devilGo-mwlm slice 3b wires the
	// table).
	MovementDistribution MovementDistribution
}

// Add ports Missiles.emplace_back (AddMissile, Source/missiles.cpp:2801):
// the new missile is appended at the container tail.
func Add(m Missile) *Missile {
	var l MissileList
	return l.AppendMissile(m)
}
