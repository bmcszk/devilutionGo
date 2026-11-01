package itemgen

// DrawItemPowerPrefixAndSuffix is GetItemPowerPrefixAndSuffix plus the RNG the
// prefix callback consumes (items.cpp:2336-2358), so a caller can replay the
// exact draw sequence of SetupAllItems.
func DrawItemPowerPrefixAndSuffix(
	prefixes, suffixes []Affix,
	minlvl, maxlvl int,
	flgs AffixItemType,
	onlygood bool,
	rnd func(int32) int32,
	discard func(int),
) AffixLoadout {
	out := GetItemPowerPrefixAndSuffix(prefixes, suffixes, minlvl, maxlvl, flgs, onlygood, rnd)
	if out.Prefix != nil {
		discard(prefixDiscards(out.Prefix.Power))
	}

	return out
}

// MagicItemName ports the name selection of SetupAllItems (:2360-2363): the
// long name, retried with the short name when it does not fit the panel.
func MagicItemName(longName, shortName string, prefix, suffix *Affix, fits func(string) bool) string {
	name := GenerateMagicItemName(longName, prefix, suffix)
	if fits != nil && !fits(name) {
		return GenerateMagicItemName(shortName, prefix, suffix)
	}

	return name
}

// prefixDiscards is the DiscardRandomValues count the SetupAllItems prefix
// callback burns after a prefix pick (items.cpp:2341-2351): one for the
// GenerateRnd(param2-param2+1) call plus the power-type extras.
func prefixDiscards(power string) int {
	switch power {
	case "DOPPELGANGER", "TOHIT_DAMP":
		return 1 + 2
	case "TOHIT_DAMP_CURSE":
		return 1 + 1
	default:
		return 1
	}
}
