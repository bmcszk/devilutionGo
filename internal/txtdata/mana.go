// Mana-cost formula per GetManaAmount (Source/spells.cpp:104-140).
package txtdata

// manaFixedShift is the 6-bit fixed-point scale the formula runs in:
// the cost is shifted up by 6 and the class multiplier (also 1/64) is
// shifted back down (spells.cpp:132-138).
const manaFixedShift = 6

// ManaInput is the per-cast state GetManaAmount reads.
type ManaInput struct {
	SpellLevel     int // player.GetSpellLevel(sn)
	CharacterLevel int // player.getCharacterLevel() — the Healing/HealOther term
	ClassManaCost  int // GetClassAttributes(cls).manaCost, the V10.3 column
	MaxManaBase    int // _pMaxManaBase; used by the ManaCost==255 branch
}

// SpellCosts is the subset of SpellData the formula reads
// (spells.cpp:104-140 fields sManaAdj/sManaCost/sMinMana).
type SpellCosts struct {
	ManaCost int // sManaCost of the cast spell
	ManaAdj  int // sManaAdj of the cast spell
	MinMana  int // sMinMana of the cast spell
	// HealingManaCost is sManaCost of SpellData[Healing]: the
	// Healing/HealOther branch reads the Healing row, not the cast
	// spell's own cost (spells.cpp:124).
	HealingManaCost int
	IsFirebolt      bool
	IsResurrect     bool
	IsHealing       bool // Healing or HealOther
}

// GetManaAmount mirrors spells.cpp:104-140 exactly, including the <<6
// fixed-point shift and the class multiplier >>6.
func GetManaAmount(in ManaInput, c SpellCosts) int {
	sl := in.SpellLevel - 1
	if sl < 0 {
		sl = 0
	}
	adj := 0
	if sl > 0 {
		adj = sl * c.ManaAdj
	}
	if c.IsFirebolt {
		adj /= 2
	}
	if c.IsResurrect && sl > 0 {
		adj = sl * (c.ManaCost / 8)
	}

	var ma int
	switch {
	case c.IsHealing: // spells.cpp:124
		ma = c.HealingManaCost + 2*in.CharacterLevel - adj
	case c.ManaCost == 255: // spells.cpp:125-126
		ma = in.MaxManaBase>>manaFixedShift - adj
	default:
		ma = c.ManaCost - adj
	}
	if ma < 0 {
		ma = 0
	}
	ma <<= manaFixedShift
	ma = ma * in.ClassManaCost >> manaFixedShift

	if c.MinMana > ma>>manaFixedShift { // spells.cpp:137-139
		ma = c.MinMana << manaFixedShift
	}
	return ma
}
