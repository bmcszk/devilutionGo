// Gold stack rules per items.h:28-30 and items.cpp:104.
package enums

const (
	// GoldSmallLimit is GOLD_SMALL_LIMIT (items.h:28): gold values at or below
	// it draw the small gold cursor (items.cpp:2980).
	GoldSmallLimit = 1000
	// GoldMediumLimit is GOLD_MEDIUM_LIMIT (items.h:29): gold values at or
	// above it draw the large gold cursor (items.cpp:2977).
	GoldMediumLimit = 2500
	// GoldMaxLimit is GOLD_MAX_LIMIT (items.h:30): the per-stack gold cap.
	GoldMaxLimit = 5000
	// GoldMaxLimitAuric is the per-stack gold cap while the hero wears the
	// Auric Amulet: upstream sets MaxGold = GOLD_MAX_LIMIT * 2 (items.cpp:2787).
	//
	// Upstream's `int MaxGold` (items.cpp:104) is mutable hero state, not a
	// constant: CalcPlrAuricBonus rewrites it on every stat recalc
	// (items.cpp:2779-2787).
	GoldMaxLimitAuric = GoldMaxLimit * 2
)
