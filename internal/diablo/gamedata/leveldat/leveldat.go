// Package leveldat holds the level-content tables: dungeon, set and theme
// enums, trigger piece lists, town triggers, the L4 betrayer gate and the
// setmap manifest.
package leveldat

import (
	"maps"
)

const (
	DTTown      DungeonType = 0
	DTCathedral DungeonType = 1
	DTCatacombs DungeonType = 2
	DTCaves     DungeonType = 3
	DTHell      DungeonType = 4
	DTNest      DungeonType = 5
	DTCrypt     DungeonType = 6
	DTLast      DungeonType = DTCrypt
	DTNone      DungeonType = -1
)

const (
	SLNone setLevelID = iota
	SLSkelking
	SLBonechamb
	SLMaze
	SLPoisonwater
	SLVilebetrayer
	SLArenaChurch
	SLArenaHell
	SLArenaCircleOfLife

	SLFirstArena setLevelID = SLArenaChurch
	SLLast       setLevelID = SLArenaCircleOfLife
)

// ThemeCount is the number of themes; ThemeNone marks the absence of a theme.
const (
	ThemeCount int8 = 17
	ThemeNone  int8 = -1
)

const (
	TrigNextLevel trigEffect = "WM_DIABNEXTLVL"
	TrigPrevLevel trigEffect = "WM_DIABPREVLVL"
	TrigTownWarp  trigEffect = "WM_DIABTOWNWARP"
	TrigSetLevel  trigEffect = "WM_DIABSETLVL"
	TrigRetLevel  trigEffect = "WM_DIABRTNLVL"
)

var pieceLists = map[string][]uint16{
	"TownDownList":  {715, 714, 718, 719, 720, 722, 723, 724, 725, 726},
	"TownWarp1List": {1170, 1171, 1172, 1173, 1174, 1175, 1176, 1177, 1178, 1180, 1182, 1184},
	"TownCryptList": {1330, 1331, 1332, 1333, 1334, 1335, 1336, 1337},
	"TownHiveList":  {1306, 1307, 1308, 1309},
	"L1UpList":      {126, 128, 129, 130, 131, 132, 134, 136, 137, 138, 139},
	"L1DownList":    {105, 106, 107, 108, 109, 111, 113, 114, 117},
	"L2UpList":      {265, 266},
	"L2DownList":    {268, 269, 270, 271},
	"L2TWarpUpList": {557, 558},
	"L3UpList":      {169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182},
	"L3DownList":    {161, 162, 163, 164, 165, 166, 167, 168},
	"L3TWarpUpList": {181, 547, 548, 549, 550, 551, 552, 553, 554, 555, 556, 557, 558, 559},
	"L4UpList":      {81, 82, 89},
	"L4DownList":    {119, 129, 130, 131, 132},
	"L4TWarpUpList": {420, 421, 428},
}

var setmapManifest = map[setLevelID]setmap{
	SLSkelking: {
		Files:  []string{"levels/l1data/sklkng1.dun", "levels/l1data/sklkng2.dun"},
		EntryX: 83, EntryY: 44,
		Palette: "levels/l1data/sklkngt.dun", // transparency file
	},
	SLBonechamb: {
		Files:  []string{"levels/l2data/bonecha2.dun", "levels/l2data/bonecha1.dun"},
		EntryX: 70, EntryY: 40,
	},
	SLPoisonwater: {
		Files:  []string{"levels/l3data/foulwatr.dun"},
		EntryX: 31, EntryY: 83,
	},
	SLVilebetrayer: {
		Files:  []string{"levels/l1data/vile1.dun", "levels/l1data/vile2.dun"},
		EntryX: 35, EntryY: 36,
	},
	SLArenaChurch: {
		Files: []string{"arena/church.dun"}, EntryX: 29, EntryY: 22, VisualX: 28, VisualY: 20,
	},
	SLArenaHell: {
		Files: []string{"arena/hell.dun"}, EntryX: 34, EntryY: 26, VisualX: 33, VisualY: 26,
	},
	SLArenaCircleOfLife: {
		Files: []string{"arena/circle_of_death.dun"}, EntryX: 30, EntryY: 26, VisualX: 29, VisualY: 26,
	},
}

// DungeonType identifies a dungeon level family; it is the canonical
// dungeon_type port — audio and light alias it.
type DungeonType int8

// SetLevelID is the exported themed-level-set identifier
// (upstream setlvl_t / setlevel enum, leveldat.h).
type SetLevelID = setLevelID

// Trigger is one trigger entry: position, effect, target level and gate
// conditions.
type Trigger struct {
	X, Y              int
	Effect            trigEffect
	TargetLevel       int
	RequiresWarpOpen  DungeonType // DTNone = unconditional
	RequiresQuestDone bool        // Q_BETRAYER == QUEST_DONE
}

// PieceRule is the dispatch rule for one piece within a dungeon level
// (upstream l4 trigger rules struct, leveldat.cpp).
type PieceRule struct {
	Effect            trigEffect
	RequiresQuestDone bool // Q_BETRAYER == QUEST_DONE
}

// Setmap is the exported themed-level manifest entry
// (upstream setlvl struct, leveldat.h).
type Setmap = setmap

// setLevelID identifies a themed level set.
type setLevelID int8

// trigEffect names the trigger message; the values are the upstream
// WM_DIAB* message names.
type trigEffect string

// setmap describes a themed level's files and entry position.

// setmap describes a themed level's files and entry position.
type setmap struct {
	Files   []string
	EntryX  int
	EntryY  int
	VisualX int // 0 when entry == visual
	VisualY int
	Palette string
}

// PieceList is the exported read access to one trigger piece list
// (upstream leveltrigger piece lists, leveldat.cpp); returns a copy.
func PieceList(name string) []uint16 { return pieceListCopy(name) }

// TownTriggers is the exported read access to the town trigger table
// (upstream town trigger init, leveldat.cpp); returns the full slice.
func TownTriggers() []Trigger { return townTriggersList() }

// L4PieceRules is the exported read access to the L4 trigger rule table
// (upstream l4 trigger dispatch, leveldat.cpp).
func L4PieceRules() map[uint16]PieceRule { return l4PieceRulesMap() }

// SetmapManifestCopy is the exported read access to the setmap manifest
// (upstream setlvldir entries, leveldat.cpp); returns a copy.
func SetmapManifestCopy() map[SetLevelID]Setmap { return setmapManifestCloned() }

// pieceListCopy returns a copy so callers cannot mutate the shared table.
func pieceListCopy(name string) []uint16 {
	src := pieceLists[name]
	out := make([]uint16, len(src))
	copy(out, src)

	return out
}

// townTriggersList returns the cathedral stairs plus the five warp-gated town
// warps.
func townTriggersList() []Trigger {
	gated := []struct {
		pos  [2]int
		tlvl int
		dt   DungeonType
	}{
		{pos: [2]int{49, 21}, tlvl: 5, dt: DTCatacombs},
		{pos: [2]int{17, 69}, tlvl: 9, dt: DTCaves},
		{pos: [2]int{41, 80}, tlvl: 13, dt: DTHell},
		{pos: [2]int{80, 62}, tlvl: 17, dt: DTNest},
		{pos: [2]int{36, 24}, tlvl: 21, dt: DTCrypt},
	}
	tg := make([]Trigger, 0, 1+len(gated))
	tg = append(tg, Trigger{X: 25, Y: 29, Effect: TrigNextLevel})

	for _, g := range gated {
		tg = append(tg, Trigger{
			X: g.pos[0], Y: g.pos[1],
			Effect:           TrigTownWarp,
			TargetLevel:      g.tlvl,
			RequiresWarpOpen: g.dt,
		})
	}

	return tg
}

// l4PieceRulesMap returns the Hellfire trigger rules; only piece 369 is gated
// on the betrayer quest.
func l4PieceRulesMap() map[uint16]PieceRule {
	rules := map[uint16]PieceRule{
		369: {Effect: TrigNextLevel, RequiresQuestDone: true},
	}
	addRules(rules, pieceLists["L4UpList"], PieceRule{Effect: TrigPrevLevel})
	addRules(rules, pieceLists["L4DownList"], PieceRule{Effect: TrigNextLevel})
	addRules(rules, pieceLists["L4TWarpUpList"], PieceRule{Effect: TrigTownWarp})

	return rules
}

// addRules adds a rule for each piece in the list if not already present.
// Mirrors the piece-rule loops in l4PieceRules (leveldat.cpp).
func addRules(rules map[uint16]PieceRule, pieces []uint16, rule PieceRule) {
	for _, p := range pieces {
		if _, ok := rules[p]; !ok {
			rules[p] = rule
		}
	}
}

// setmapManifestCloned returns a copy of the manifest keyed by setLevelID.
func setmapManifestCloned() map[setLevelID]setmap {
	out := make(map[setLevelID]setmap, len(setmapManifest))
	maps.Copy(out, setmapManifest)

	return out
}
