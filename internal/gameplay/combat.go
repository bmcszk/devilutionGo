// Package gameplay ports the combat roll pipeline: PlrHitMonst
// (Source/player.cpp:526-660) with the exact upstream roll order and
// 1/64 fixed-point damage. Entity wiring (Player/Monster structs) lands
// with its own beans; the pipeline here takes explicit inputs so the
// roll ORDER is testable now (devilGo-7be9 roll-tape harness).
package gameplay

// ClassFlag marks the player class (PlayerClassFlag carrier set; only
// the flags PlrHitMonst reads are modeled).
type ClassFlag uint8

// Player classes (player.h ClassAttributes carriers).
const (
	Warrior ClassFlag = iota
	Rogue
	Sorcerer
)

// MonsterClass mirrors MonsterClass (monster.h): the undead/animal/demon
// triad the sword/mace damage modifiers branch on.
type MonsterClass uint8

const (
	Undead MonsterClass = iota
	Animal
	Demon
)

// WeaponType is the equipped-weapon ItemType slice PlrHitMonst reads
// (Sword/Mace; anything else is None).
type WeaponType uint8

const (
	WeaponNone WeaponType = iota
	Sword
	Mace
)

// Ref carries the inputs PlrHitMonst consumes, in one struct.
type Ref struct {
	Class          ClassFlag
	Level          int
	Adjacent       bool // ranged attack fired while adjacent (Petrify/short range)
	Petrified      bool // monster.mode == MonsterMode::Petrified
	ToHitBonus     int  // GetMeleePiercingToHit
	ArmorPierce    int  // CalculateArmorPierce(monster.armorClass, true)
	MinDam, MaxDam int
	BonusDamPct    int // _pIBonusDam
	BonusDamMod    int // _pIBonusDamMod
	DamageMod      int // _pDamageMod
	Weapon         WeaponType
	MClass         MonsterClass
	TripleDemon    bool // ItemSpecialEffect::TripleDemonDamage
}

// PlrHitMonstCore ports PlrHitMonst's roll order (player.cpp:526-613):
//
//  1. hit roll         GenerateRnd(100)                (:539)
//  2. petrified        hit = 0                          (:540-542)
//  3. hper + clamp 5..95                                (:545-547)
//  4. hit >= hper -> miss                               (:551-556)
//  5. damage roll      RandomIntBetween(min, max)      (:566)
//  6. class weapon modifiers (undead/animal)            (:592-604)
//  7. triple demon damage                               (:606-611)
//  8. adjacent damage dam >>= 2                          (:628-630)
//
// Damage returns in 1/64 fixed-point units (dam <<= 6 upstream).
// The rnd function is GenerateRnd; the tape harness pins its sequence.
func PlrHitMonstCore(r Ref, rnd func(int) int) (bool, int64) {
	hper := 0
	if r.Adjacent {
		if r.Level > 20 {
			hper -= 30
		} else {
			hper -= (35 - r.Level) * 2
		}
	}

	hit := rnd(100)
	if r.Petrified {
		hit = 0
	}

	hper += r.ToHitBonus - r.ArmorPierce
	if hper < 5 {
		hper = 5
	}
	if hper > 95 {
		hper = 95
	}

	if hit >= hper {
		return false, 0
	}

	// RandomIntBetween(min, max) = RandomIntLessThan(max-min+1) + min
	// (random.hpp:403); RandomIntLessThan(n) = GenerateRnd(n).
	dam := rnd(r.MaxDam-r.MinDam+1) + r.MinDam
	dam += dam * r.BonusDamPct / 100
	dam += r.BonusDamMod
	_ = dam << 6 // dam2 upstream: carried by the peril branch in the full port
	dam += r.DamageMod

	dam = applyMonsterClassModifiers(dam, r.MClass, r.Weapon)
	if r.MClass == Demon && r.TripleDemon {
		// ItemSpecialEffect::TripleDemonDamage (player.cpp:606-611).
		dam *= 3
	}

	if r.Adjacent {
		dam >>= 2
	}

	return true, int64(dam) << 6
}

// applyMonsterClassModifiers ports the undead/animal weapon-modifier
// switch (player.cpp:592-604): sword vs mace halves/doubles by class.
func applyMonsterClassModifiers(dam int, mc MonsterClass, weapon WeaponType) int {
	switch mc {
	case Undead:
		return applyUndeadModifier(dam, weapon)
	case Animal:
		return applyAnimalModifier(dam, weapon)
	default: // Demon and any future class: no weapon modifier.
		return dam
	}
}

// applyUndeadModifier: sword halves, mace doubles (player.cpp:593-599).
func applyUndeadModifier(dam int, weapon WeaponType) int {
	switch weapon {
	case Sword:
		return dam - dam/2
	case Mace:
		return dam + dam/2
	default: // WeaponNone
		return dam
	}
}

// applyAnimalModifier: mace halves, sword doubles (player.cpp:600-604).
func applyAnimalModifier(dam int, weapon WeaponType) int {
	switch weapon {
	case Mace:
		return dam - dam/2
	case Sword:
		return dam + dam/2
	default: // WeaponNone
		return dam
	}
}
