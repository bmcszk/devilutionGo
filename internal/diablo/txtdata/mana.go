package txtdata

// manaFixedShift is the 6-bit fixed-point scale the formula runs in: the cost
// is shifted up by 6 and the class multiplier (also 1/64) is shifted back down.
const manaFixedShift = 6

// ManaInput is the per-cast state GetManaAmount reads.
type ManaInput struct {
	SpellLevel     int
	CharacterLevel int // the Healing/HealOther term
	ClassManaCost  int // the class mana multiplier in 1/64 units
	MaxManaBase    int // base mana in 1/64 units; used by the ManaCost==255 branch
}

// SpellCosts is the subset of SpellData the formula reads.
type SpellCosts struct {
	ManaCost int
	ManaAdj  int
	MinMana  int
	// HealingManaCost is the mana cost of the Healing spell: the
	// Healing/HealOther branch reads the Healing row, not the cast spell's
	// own cost.
	HealingManaCost int
	IsFirebolt      bool
	IsResurrect     bool
	IsHealing       bool // Healing or HealOther
}

// GetManaAmount returns the mana cost of a cast, scaled by manaFixedShift.
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
	case c.IsHealing:
		ma = c.HealingManaCost + 2*in.CharacterLevel - adj
	case c.ManaCost == 255:
		ma = in.MaxManaBase>>manaFixedShift - adj
	default:
		ma = c.ManaCost - adj
	}
	if ma < 0 {
		ma = 0
	}
	ma <<= manaFixedShift
	ma = ma * in.ClassManaCost >> manaFixedShift

	if c.MinMana > ma>>manaFixedShift {
		ma = c.MinMana << manaFixedShift
	}
	return ma
}
