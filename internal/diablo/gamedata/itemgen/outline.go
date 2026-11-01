package itemgen

// Palette bases (engine/palette.h:33-40) and the +5 item-colour offset
// (inv.h:78-83).
const (
	Pal16Blue   = 176
	Pal16Yellow = 192
	Pal16Red    = 224
	Pal16Gray   = 240

	itemColorOffset = 5

	IColYellow = Pal16Yellow + itemColorOffset
	IColWhite  = Pal16Gray + itemColorOffset
	IColBlue   = Pal16Blue + itemColorOffset
	IColRed    = Pal16Red + itemColorOffset
)

// GetOutlineColor ports items.cpp:2403-2415.
//
// for gold and uniques, blue for magic, white otherwise.
//
//nolint:revive // checkReq is upstream's parameter, not control coupling.: red for unusable items, yellow
func GetOutlineColor(itemType string, quality int, statFlag, checkReq bool) int {
	if checkReq && !statFlag {
		return IColRed
	}
	if itemType == "Gold" {
		return IColYellow
	}
	if quality == QualityMagic {
		return IColBlue
	}
	if quality == QualityUnique {
		return IColYellow
	}

	return IColWhite
}
