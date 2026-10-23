// Package missiles: Missile struct and container row.
package missiles

// missileID mirrors enum class missileID:
// int8_t ordinals into the misdat table — GetMissileData indexes
// MissilesData[underlying(missileId)], so
// ordinal == base misdat.tsv row index (verified: spelldat.h names 1:1 with
// rows 1-68 of assets/txtdata/missiles/misdat.tsv; LastDiablo = DiabloApocalypse).
type missileID int8

const (
	Arrow missileID = iota
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
	Mana            missileID = 68
	Magi            missileID = 69
	LightningWall   missileID = 70 // put-missile flag bit
	OrangeFlare     missileID = 98
	BlueFlare       missileID = 99
	RedFlare        missileID = 100
	YellowFlare     missileID = 101
	BlueFlare2      missileID = 102
	YellowExplosion missileID = 103
	RedExplosion    missileID = 104
	BlueExplosion   missileID = 105
	BlueExplosion2  missileID = 106
	OrangeExplosion missileID = 107
)

// graphicID mirrors enum class MissileGraphicID: uint8_t
// — the animation sheet id.
// Only the sheets slice 3a-ii's
// animation-phase switches compare against are named; ordinals are
// explicit because the enum is long.
type graphicID uint8

// GraphicArrow itself lives in rotation.go (untyped slice-2 const 0;
// keep both in sync).
const (
	GraphicFireball           graphicID = 1
	GraphicLightning          graphicID = 3
	GraphicMagmaBallExplosion graphicID = 5
	GraphicBigExplosion       graphicID = 19
	GraphicChargedBolt        graphicID = 26
	GraphicHolyBolt           graphicID = 27
	GraphicHolyBoltExplosion  graphicID = 28
	GraphicLightningArrow     graphicID = 29
	GraphicFireArrowExplosion graphicID = 30
)

// Missile is the Go port of struct Missile
// slice-1 field set: identity, source, del flag, animation counters and
// the 16.16 position block.
type Missile struct {
	// Type of projectile.
	Type missileID
	// Position on the dungeon grid.
	Position Position
	// MFNum is the direction enum of the missile.
	MFNum int
	// Spell level.
	SpellLevel int
	// DelFlag marks the missile for deletion.
	DelFlag bool
	// MissileGraphicID: the animation sheet id (graphicID).
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
	// mienemy_type caster class.
	caster int
	// Source player, when the caster is player-side and Source != -1
	SourcePlayer int
	// _midist: used for arrows to measure distance travelled (increases
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

// Add appends the missile at the container tail and returns it.
func Add(m Missile) *Missile {
	var l missileList
	return l.AppendMissile(m)
}
