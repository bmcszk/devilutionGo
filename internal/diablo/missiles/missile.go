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
	Mana                 MissileID = 68
	Magi                 MissileID = 69
	LightningWall        MissileID = 70 // put-missile flag bit
	LightningWallControl MissileID = 71
	Immolation           MissileID = 72
	SpectralArrow        MissileID = 73
	FireballBow          MissileID = 74
	LightningBow         MissileID = 75
	ChargedBoltBow       MissileID = 76
	HolyBoltBow          MissileID = 77
	Warp                 MissileID = 78
	Reflect              MissileID = 79
	Berserk              MissileID = 80
	RingOfFire           MissileID = 81
	StealPotions         MissileID = 82
	StealMana            MissileID = 83
	RingOfLightning      MissileID = 84
	Search               MissileID = 85
	Aura                 MissileID = 86
	Aura2                MissileID = 87
	SpiralFireball       MissileID = 88
	RuneOfFire           MissileID = 89
	RuneOfLight          MissileID = 90
	RuneOfNova           MissileID = 91
	RuneOfImmolation     MissileID = 92
	RuneOfStone          MissileID = 93
	BigExplosion         MissileID = 94
	HorkSpawn            MissileID = 95
	Jester               MissileID = 96
	OpenNest             MissileID = 97
	OrangeFlare          MissileID = 98
	BlueFlare            MissileID = 99
	RedFlare             MissileID = 100
	YellowFlare          MissileID = 101
	BlueFlare2           MissileID = 102
	YellowExplosion      MissileID = 103
	RedExplosion         MissileID = 104
	BlueExplosion        MissileID = 105
	BlueExplosion2       MissileID = 106
	OrangeExplosion      MissileID = 107
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
	// Blood-star explosion family (missile_sprites.tsv rows 23, 41, 43, 45);
	// the four succubus tiers each carry their own palette-shifted sheet.
	GraphicBloodStarExplosion       GraphicID = 23
	GraphicBloodStarBlueExplosion   GraphicID = 41
	GraphicBloodStarYellowExplosion GraphicID = 43
	GraphicBloodStarRedExplosion    GraphicID = 45
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
	// LightFlag is _miLightFlag (missiles.h:115): the tile's light source is
	// drawn over the missile (engine/render/scrollrt.cpp:348). Only the acid
	// puddle raises it today (missiles.cpp:2357); the renderer reads it once
	// the scroll table is ported.
	LightFlag bool
	// PreFlag renders the missile over walls
	// _miPreFlag); PutMissile raises the sink's MissilePreFlag
	// when set.
	PreFlag bool
	// HitSound is the misdat row's hitSound, played at the missile tile when
	// the duration reaches 0 (missiles.cpp:549-550); empty = SfxID::None.
	HitSound string
	// LimitReached is _miLimitReached: the wall control's left arm has
	// stopped growing (missiles.h).
	LimitReached int
	// Var7 is _mivar7: the wall control's right arm has stopped growing
	// (missiles.h:134).
	Var7 int
	// DrawFlag is _miDrawFlag: isDrawn() of the misdat row — false for the
	// Invisible rows, which the renderer must skip (tables/misdat.h:185-188).
	DrawFlag bool
	// UniqTrans is _miUniqTrans: the unique monster's palette translation
	// + 1, 0 when the caster is not a unique monster (missiles.cpp:2827-2832).
	UniqTrans int
	// AnimFAmt is the sheet's animFAmt — the frame group count taken from
	// the sprite row's numFrames (misdat.cpp:152). 0 means the sprite row
	// was not loaded and graphicFrameCount stands in.
	AnimFAmt int
	// AnimDelays and AnimLengths are the row's frameDelay/frameLength
	// columns — one entry per frame group (tables/misdat.cpp:400-408).
	// setAnimation reloads AnimDelay/AnimLen from them; empty means the
	// sprite table is not loaded and the caller's values stay.
	AnimDelays  []int
	AnimLengths []int
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
	// Var4/Var5/Var6 are _mivar4..6 (missiles.h:130-132), the box cursors
	// Apocalypse walks and the wall spread positions.
	Var4 int
	Var5 int
	Var6 int
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
