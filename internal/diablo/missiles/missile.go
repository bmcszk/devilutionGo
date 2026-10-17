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
}

// Add ports Missiles.emplace_back (AddMissile, Source/missiles.cpp:2801):
// the new missile is appended at the container tail.
func Add(m Missile) *Missile {
	Rows = append(Rows, m)
	return &Rows[len(Rows)-1]
}
