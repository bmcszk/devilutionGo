// Package leveldat ports the level-content tables: dungeon/set/theme enums
// (levels/gendung_defs.hpp), trigger piece lists (levels/trigs.cpp:31-59),
// town triggers (trigs.cpp:109-149), the L4 betrayer gate (trigs.cpp:253-261),
// and the setmap manifest (levels/setmaps.cpp:112-165).
package leveldat

import "fmt"

// DungeonType mirrors dungeon_type (gendung_defs.hpp:15-24).
type DungeonType int8

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

// SetLevelID mirrors _setlevels (gendung_defs.hpp:47-60).
type SetLevelID int8

const (
	SLNone SetLevelID = iota
	SLSkelking
	SLBonechamb
	SLMaze
	SLPoisonwater
	SLVilebetrayer
	SLArenaChurch
	SLArenaHell
	SLArenaCircleOfLife

	SLFirstArena SetLevelID = SLArenaChurch
	SLLast       SetLevelID = SLArenaCircleOfLife
)

// ThemeID count (objdat.h:18-37): 17 themes + THEME_NONE.
const (
	ThemeCount int8 = 17
	ThemeNone  int8 = -1
)

// pieceLists mirrors trigs.cpp:31-59 (uint16 piece ids).
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

// PieceList returns a copy of the named trigger piece list.
func PieceList(name string) []uint16 {
	src := pieceLists[name]
	out := make([]uint16, len(src))
	copy(out, src)
	return out
}

// TrigEffect mirrors the WM_DIAB* trigger messages.
type TrigEffect string

const (
	TrigNextLevel TrigEffect = "WM_DIABNEXTLVL"
	TrigPrevLevel TrigEffect = "WM_DIABPREVLVL"
	TrigTownWarp  TrigEffect = "WM_DIABTOWNWARP"
	TrigSetLevel  TrigEffect = "WM_DIABSETLVL"
	TrigRetLevel  TrigEffect = "WM_DIABRTNLVL"
)

// Trigger is one entry of trigs[] (trigs.cpp:27-29 block): position,
// effect, optional target level, and the town-warp gate.
type Trigger struct {
	X, Y              int
	Effect            TrigEffect
	TargetLevel       int
	RequiresWarpOpen  DungeonType // DTNone = unconditional
	RequiresQuestDone bool        // L4 369: Q_BETRAYER == QUEST_DONE
}

// TownTriggers ports InitTownTriggers (trigs.cpp:109-149): the cathedral
// stairs at (25,29) plus the five IsWarpOpen-gated town warps.
func TownTriggers() []Trigger {
	tg := []Trigger{
		{X: 25, Y: 29, Effect: TrigNextLevel},
	}
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

// PieceRule is the per-piece dispatch rule within one dungeon level.
type PieceRule struct {
	Effect            TrigEffect
	RequiresQuestDone bool // Q_BETRAYER == QUEST_DONE gate
}

// L4PieceRules ports InitHellFireTrigger (trigs.cpp:253-261): piece 369 is
// NEXT only when Q_BETRAYER is QUEST_DONE; the stair pieces (82 up/PREV,
// 119/129-132 down/NEXT) dispatch unconditionally by list membership.
func L4PieceRules() map[uint16]PieceRule {
	rules := map[uint16]PieceRule{
		369: {Effect: TrigNextLevel, RequiresQuestDone: true},
	}
	for _, p := range pieceLists["L4UpList"] {
		if _, ok := rules[p]; !ok {
			rules[p] = PieceRule{Effect: TrigPrevLevel}
		}
	}
	for _, p := range pieceLists["L4DownList"] {
		if _, ok := rules[p]; !ok {
			rules[p] = PieceRule{Effect: TrigNextLevel}
		}
	}
	for _, p := range pieceLists["L4TWarpUpList"] {
		if _, ok := rules[p]; !ok {
			rules[p] = PieceRule{Effect: TrigTownWarp}
		}
	}
	return rules
}

// Setmap is one LoadSetMap case (setmaps.cpp:112-165).
type Setmap struct {
	Files   []string
	EntryX  int
	EntryY  int
	VisualX int // 0 when entry == visual
	VisualY int
	Palette string
}

// setmapManifest mirrors the LoadSetMap switch entries.
var setmapManifest = map[SetLevelID]Setmap{
	SLSkelking: {
		Files:  []string{"levels/l1data/sklkng1.dun", "levels/l1data/sklkng2.dun"},
		EntryX: 83, EntryY: 44,
		Palette: "levels/l1data/sklkngt.dun", // transparency file (SetMapTransparency)
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

// SetmapManifest returns the manifest keyed by SetLevelID (copies).
func SetmapManifest() map[SetLevelID]Setmap {
	out := make(map[SetLevelID]Setmap, len(setmapManifest))
	for k, v := range setmapManifest {
		out[k] = v
	}
	return out
}

// ArenaLevelType ports GetArenaLevelType (setmaps.h:19-30).
func ArenaLevelType(id SetLevelID) (DungeonType, error) {
	switch id {
	case SLArenaChurch:
		return DTCathedral, nil
	case SLArenaHell, SLArenaCircleOfLife:
		return DTHell, nil
	default:
		return DTNone, fmt.Errorf("leveldat: %v is not an arena level", id)
	}
}
