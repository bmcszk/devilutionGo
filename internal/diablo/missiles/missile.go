// Package missiles: Missile struct and container row.
package missiles

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
	LastDiablo = DiabloApocalypse
	// Hellfire-region ids — explicit ordinals:
	// the enum restarts after LastDiablo upstream, so iota is broken
	// from here; only the
	// constants slice 3a-beta needs are named.
	Mana            MissileID = 68
	Magi            MissileID = 69
	LightningWall   MissileID = 70 // put-missile flag bit
	OrangeFlare     MissileID = 98
	BlueFlare       MissileID = 99
	RedFlare        MissileID = 100
	YellowFlare     MissileID = 101
	BlueFlare2      MissileID = 102
	YellowExplosion MissileID = 103
	RedExplosion    MissileID = 104
	BlueExplosion   MissileID = 105
	BlueExplosion2  MissileID = 106
	OrangeExplosion MissileID = 107
)

// GraphicArrow itself lives in rotation.go (untyped slice-2 const 0;
// keep both in sync).
const (
	GraphicFireball           GraphicID = 1
	GraphicLightning          GraphicID = 3
	GraphicMagmaBallExplosion GraphicID = 5
	GraphicBigExplosion       GraphicID = 19
	GraphicChargedBolt        GraphicID = 26
	GraphicHolyBolt           GraphicID = 27
	GraphicHolyBoltExplosion  GraphicID = 28
	GraphicLightningArrow     GraphicID = 29
	GraphicFireArrowExplosion GraphicID = 30
)

// MissileID mirrors enum class MissileID:
// int8_t ordinals into the misdat table — GetMissileData indexes
// MissilesData[underlying(missileId)], so
// ordinal == base misdat.tsv row index (verified: spelldat.h names 1:1 with
// rows 1-68 of assets/txtdata/missiles/misdat.tsv; LastDiablo = DiabloApocalypse).
type MissileID int8

// GraphicID mirrors enum class MissileGraphicID: uint8_t
// — the animation sheet id.
// Only the sheets slice 3a-ii's
// animation-phase switches compare against are named; ordinals are
// explicit because the enum is long.
type GraphicID uint8

// Missile is the Go port of struct Missile
// slice-1 field set: identity, source, del flag, animation counters and
// the 16.16 position block.
type Missile struct {
	// Type of projectile.
	Type MissileID
	// Position on the dungeon grid.
	Position Position
	// MFNum is the direction enum of the missile.
	MFNum int
	// Spell level.
	SpellLevel int
	// DelFlag marks the missile for deletion.
	DelFlag bool
	// MissileGraphicID: the animation sheet id (GraphicID).
	AnimType int
	// Tick length of each frame in the current animation.
	AnimDelay int
	// Number of frames in current animation.
	AnimLen int
	// Increases by one each game tick, counting how close we are to
	// AnimDelay.
	AnimCnt int
	// Added to AnimFrame each processed tick.
	AnimAdd int
	// Current frame of animation + 1.
	AnimFrame int
	// Source id: player index, monster index, or -1 for traps
	Source int
	// Caster: mienemy_type caster class.
	Caster int
	// Source player, when the caster is player-side and Source != -1
	SourcePlayer int
	// _midist: used for arrows to measure distance traveled (increases
	// by 1 each game tick). Higher value is a penalty for accuracy
	// calculation when hitting an enemy.
	Dist int
	// Duration is time to live in game ticks; at 0 the missile is
	// deleted (upstream _mirange int).
	Duration int
	// HitFlag: the missile hit something this tick (_miHitFlag).
	HitFlag bool
	// PreFlag renders the missile over walls
	// _miPreFlag); PutMissile raises the sink's MissilePreFlag
	// when set.
	PreFlag bool
	// Rolled at launch, in 1/64 fixed-point units.
	Dam int
	// Light id for the missile's light source, NoLight = none
	Lid int
	// Per-type scratch registers:
	// ProcessGenericProjectile tracks the light tile here
	// ; MissileExplosion counts its light
	// ramp in var1; ChargedBolt keeps its
	// burn-out radius in var1.
	Var1 int
	Var2 int
	// ChargedBolt steering countdown.
	Var3 int
	// Mirnd is the ChargedBolt BPath cursor (uint8 upstream).
	Mirnd uint8
	// LastCollisionTargetHash: hash of the dMonster^dPlayer cell the last
	// on-tile CheckMissileCol probed — the slow-missile multi-hit guard.
	LastCollisionTargetHash int16
	// MovementDistribution ports MissileMovementDistribution
	// — how a hit affects movement.
	// Set from the row's
	// movementDistribution at load (devilGo-mwlm slice 3b wires the
	// table).
	MovementDistribution MovementDistribution
}
