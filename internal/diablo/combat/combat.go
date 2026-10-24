// Package combat ports the combat roll pipeline (PlrHitMonst) with the
// exact upstream roll order and 1/64 fixed-point damage. Entity wiring
// (Player/Monster structs) lands with its own beans; the pipeline here
// takes explicit inputs so the roll ORDER is testable now
// (devilGo-7be9 roll-tape harness).
package combat

// Class is the hero class identity (upstream HeroClass).
type Class uint8

// The base-game classes; the Hellfire classes
// (Monk/Bard/Barbarian) land with their own bean.
const (
	Warrior Class = iota
	Rogue
	Sorcerer
)

// classFlag mirrors PlayerClassFlag; only the
// flag PlrHitMonst reads is modeled.
type classFlag uint8

const (
	FlagNone classFlag = 0
	// FlagCriticalStrike doubles the damage on a character-level roll
	// (player.cpp:574-578).
	FlagCriticalStrike classFlag = 1 << 0
)

// classFlags mirrors the classFlags row of each class's attributes.tsv
// (assets/txtdata/classes/<class>/attributes.tsv): CriticalStrike on
// the warrior, TrapSense on the rogue, none on the sorcerer. Only the
// CriticalStrike flag has a consumer in this port.
func classFlags(c Class) classFlag {
	if c == Warrior {
		return FlagCriticalStrike
	}
	return FlagNone
}

// MonsterClass mirrors MonsterClass: the
// undead/demon/animal triad the sword/mace damage modifiers branch on.
type MonsterClass uint8

const (
	Undead MonsterClass = iota
	Demon
	Animal
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
	Class          Class
	Level          int
	Adjacent       bool // upstream adjacentDamage: side-adjacent cleave hit (player.cpp:825,833)
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

// PlrHitMonstCore ports PlrHitMonst's roll order
// 1. hit roll  GenerateRnd(100)  (:542)
// 2. petrified hit = 0    (:543-545)
// 3. hper + clamp 5..95    (:547-548)
// 4. hit >= hper -> miss    (:553-558)
// 5. damage roll RandomIntBetween(min, max) (:567)
// 6. critical strike GenerateRnd(100) < level (:574-578)
// 7. class weapon modifiers (undead/animal)  (:589-602)
// 8. triple demon damage    (:603-606)
// 9. damage shift dam <<= 6   (:618)
// The adjacent penalty (dam >>= 2) lands after the shift.
// Damage returns in 1/64 fixed-point units: the shift to 1/64 happens
// before the adjacent penalty, so the word
// handed to ApplyMonsterDamage is already scaled.
// The rnd function is GenerateRnd; the tape harness pins its sequence.
func PlrHitMonstCore(r Ref, rnd func(int) int) (bool, int64) {
	// Upstream returns before any roll when the monster cannot be hit;
	// the port has no monster state, so the gate lands with the monster
	// wiring.
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

	// A gargoyle that lifts off ends the swing without a damage roll;
	// the port has no monster state to ask.
	if hit >= hper {
		return false, 0
	}

	// The hellfire SpectralArrow fires beside the melee blow; the
	// missile seam lands with the spell wiring.
	// RandomIntBetween(min, max) = RandomIntLessThan(max-min+1) + min;
	// RandomIntLessThan(n) = GenerateRnd(n).
	dam := rnd(r.MaxDam-r.MinDam+1) + r.MinDam
	dam += dam * r.BonusDamPct / 100
	dam += r.BonusDamMod
	// The hellfire Peril branch applies a second damage word; it lands
	// with the item wiring.
	dam += r.DamageMod

	// Critical strike: the class flag gates the roll, so only flagged
	// classes consume a draw here, and the roll lands after the damage
	// modifiers and before the weapon ones.
	if classFlags(r.Class)&FlagCriticalStrike != 0 {
		if rnd(100) < r.Level {
			dam *= 2
		}
	}

	dam = applyMonsterClassModifiers(dam, r.MClass, r.Weapon)
	if r.MClass == Demon && r.TripleDemon {
		// ItemSpecialEffect::TripleDemonDamage.
		dam *= 3
	}

	// The hellfire Devastation and Doppelganger effects ride the same
	// damage word; they need the item-effect flags, so they land with
	// the item wiring.

	// Damage shift to 1/64 fixed point; it lands after the class
	// modifiers and before the adjacent penalty.
	dam <<= 6

	// The hellfire Jesters effect re-rolls the damage word after the
	// shift; it lands with the item wiring.
	if r.Adjacent {
		dam >>= 2
	}

	return true, int64(dam)
}

// applyMonsterClassModifiers ports the undead/animal weapon-modifier
// switch: sword vs mace halves/doubles by class.
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

// applyUndeadModifier: sword halves, mace doubles.
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

// applyAnimalModifier: mace halves, sword doubles.
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
