package enums

const (
	// GoldSmallLimit is the gold value at or below which a gold stack uses the
	// small gold cursor.
	GoldSmallLimit = 1000
	// GoldMediumLimit is the gold value at or above which a gold stack uses the
	// large gold cursor.
	GoldMediumLimit = 2500
	// GoldMaxLimit is the maximum gold value in a single stack.
	GoldMaxLimit = 5000
	// GoldMaxLimitAuric is the maximum gold value in a single stack while the
	// hero wears the Auric Amulet.
	GoldMaxLimitAuric = GoldMaxLimit * 2
)
