// Mana-cost formula per GetManaAmount (Source/spells.cpp:104-140).
package txtdata

// ManaInput is the per-cast state GetManaAmount reads.
type ManaInput struct {
	SpellLevel    int
	ClassManaCost int // GetClassAttributes(cls).manaCost, the V10.3 column
	MaxManaBase   int // _pMaxManaBase; used by the ManaCost==255 branch
}

// SpellCosts is the subset of SpellData the formula reads
// (spells.cpp:104-140 fields sManaAdj/sManaCost).
type SpellCosts struct {
	ManaCost    int
	ManaAdj     int // sManaAdj, the manaMultiplier column
	IsFirebolt  bool
	IsResurrect bool
	IsHealing   bool // Healing or HealOther
}

// GetManaAmount mirrors spells.cpp:104-140 exactly, including the <<6
// fixed-point shift and the class multiplier >>6.
func GetManaAmount(in ManaInput, c SpellCosts) int {
	sl := in.SpellLevel - 1
	if sl < 0 {
		sl = 0
	}
	adj := sl * c.ManaAdj
	switch {
	case c.IsFirebolt: // spells.cpp:117-119
		adj = sl * (c.ManaAdj / 2)
	case c.IsResurrect: // spells.cpp:120-122
		adj = sl * (c.ManaCost / 8)
	default:
	}
	var ma int
	switch {
	case c.IsHealing: // spells.cpp:124
		ma = c.ManaCost + 2*in.SpellLevel - adj
	case c.ManaCost == 255: // spells.cpp:125-126
		ma = in.MaxManaBase>>6 - adj
	default:
		ma = c.ManaCost - adj
	}
	if ma < 0 {
		ma = 0
	}
	ma <<= 6
	return ma * in.ClassManaCost >> 6 // spells.cpp:136-138
}
