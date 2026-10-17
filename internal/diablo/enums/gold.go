// Gold stack rules per items.h:28-30 and items.cpp:104.
package enums

const (
	GoldSmallLimit  = 1000 // GOLD_SMALL_LIMIT (items.h:28)
	GoldMediumLimit = 2500 // GOLD_MEDIUM_LIMIT (items.h:29)
	GoldMaxLimit    = 5000 // GOLD_MAX_LIMIT (items.h:30)
)

// MaxGold is the per-stack cap (items.cpp:104).
const MaxGold = GoldMaxLimit
