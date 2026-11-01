package stores

// The gossip table. The towner gossip ids are the `gossipTexts` column of
// towners.tsv, committed at internal/testdata/gamedata/towners/towners.tsv
// (upstream devilutionX c20b2f2456 assets/txtdata/towners/towners.tsv); the
// strings behind those ids are the textdat.tsv rows committed at
// internal/testdata/gamedata/textdat/textdat.tsv. This file is the mapping and
// the pick; the string tables themselves stay in the data loaders.

// GossipTextIDs maps a towner type to its gossipTexts column, in TSV order
// (towners.tsv column 10). TOWN_DEADGUY has no gossip.
var GossipTextIDs = map[string][]string{
	"TOWN_SMITH": {
		"TEXT_GRISWOLD2", "TEXT_GRISWOLD3", "TEXT_GRISWOLD4", "TEXT_GRISWOLD5",
		"TEXT_GRISWOLD6", "TEXT_GRISWOLD7", "TEXT_GRISWOLD8", "TEXT_GRISWOLD9",
		"TEXT_GRISWOLD10", "TEXT_GRISWOLD12", "TEXT_GRISWOLD13",
	},
	"TOWN_HEALER": {
		"TEXT_PEPIN2", "TEXT_PEPIN3", "TEXT_PEPIN4", "TEXT_PEPIN5",
		"TEXT_PEPIN6", "TEXT_PEPIN7", "TEXT_PEPIN9", "TEXT_PEPIN10", "TEXT_PEPIN11",
	},
	"TOWN_TAVERN": {
		"TEXT_OGDEN2", "TEXT_OGDEN3", "TEXT_OGDEN4", "TEXT_OGDEN5",
		"TEXT_OGDEN6", "TEXT_OGDEN8", "TEXT_OGDEN9", "TEXT_OGDEN10",
	},
	"TOWN_STORY": {
		"TEXT_STORY2", "TEXT_STORY3", "TEXT_STORY4", "TEXT_STORY5",
		"TEXT_STORY6", "TEXT_STORY7", "TEXT_STORY9", "TEXT_STORY10", "TEXT_STORY11",
	},
	"TOWN_DRUNK": {
		"TEXT_FARNHAM2", "TEXT_FARNHAM3", "TEXT_FARNHAM4", "TEXT_FARNHAM5",
		"TEXT_FARNHAM6", "TEXT_FARNHAM8", "TEXT_FARNHAM9", "TEXT_FARNHAM10",
		"TEXT_FARNHAM11", "TEXT_FARNHAM12", "TEXT_FARNHAM13",
	},
	"TOWN_WITCH": {
		"TEXT_ADRIA2", "TEXT_ADRIA3", "TEXT_ADRIA4", "TEXT_ADRIA5",
		"TEXT_ADRIA6", "TEXT_ADRIA7", "TEXT_ADRIA8", "TEXT_ADRIA9",
		"TEXT_ADRIA10", "TEXT_ADRIA12", "TEXT_ADRIA13",
	},
	"TOWN_BMAID": {
		"TEXT_GILLIAN2", "TEXT_GILLIAN3", "TEXT_GILLIAN4", "TEXT_GILLIAN5",
		"TEXT_GILLIAN6", "TEXT_GILLIAN7", "TEXT_GILLIAN9", "TEXT_GILLIAN10",
	},
	"TOWN_PEGBOY": {
		"TEXT_WIRT2", "TEXT_WIRT3", "TEXT_WIRT4", "TEXT_WIRT5", "TEXT_WIRT6",
		"TEXT_WIRT7", "TEXT_WIRT8", "TEXT_WIRT9", "TEXT_WIRT11", "TEXT_WIRT12",
	},
}

// Rnd is the GenerateRnd source the gossip pick draws from, so a test can
// script the tape (engine.GenerateRnd is the production one).
type Rnd func(n int32) int32

// PickGossip ports the gossip draw of InitTownerFromData
// (towners.cpp:133-136): GenerateRnd(count) clamped to zero picks the text id.
// The second result is false for a towner with no gossip.
func PickGossip(townerType string, rnd Rnd) (string, int, bool) {
	ids := GossipTextIDs[townerType]
	if len(ids) == 0 {
		return "", -1, false
	}

	idx := int(rnd(int32(len(ids))))
	if idx < 0 {
		idx = 0
	}

	if idx >= len(ids) {
		idx = len(ids) - 1
	}

	return ids[idx], idx, true
}

// GossipText resolves a gossip id through the textdat table the caller loaded,
// reporting whether the id is present.
func GossipText(id string, table map[string]string) (string, bool) {
	text, ok := table[id]

	return text, ok
}
