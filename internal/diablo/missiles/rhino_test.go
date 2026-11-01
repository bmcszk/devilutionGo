// Package missiles_test: the Rhino (charging monster) row.
package missiles_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

const (
	monsterModeChargeTest = 14
	monsterAISnakeTest    = 24
)

var addRhinoCases = []struct {
	name string
	cfg  rhinoAddCase
}{
	{"a common monster charges on its walk sheet", rhinoAddCase{
		monsterType: 0, wantGraphic: 1, wantFrame: 1,
	}},
	{"the horned family charges on its special sheet", rhinoAddCase{
		monsterType: missiles.MonsterTypeHorned, wantGraphic: 5, wantFrame: 1,
	}},
	{"the mud run charges on its special sheet", rhinoAddCase{
		monsterType: missiles.MonsterTypeMudRun, wantGraphic: 5, wantFrame: 1,
	}},
	{"a snake charges on its attack sheet, from frame 7", rhinoAddCase{
		monsterType: missiles.MonsterTypeNSnake, wantGraphic: 2, wantFrame: 7,
	}},
	{"a unique carries its light id", rhinoAddCase{
		monsterType: 0, unique: true, lightID: 42, wantGraphic: 1, wantFrame: 1,
	}},
}

var processRhinoCases = []struct {
	name string
	cfg  rhinoStepCase
}{
	{"a monster that left the charge ends the missile", rhinoStepCase{mode: 0, wantDel: true}},
	{"a clear landing writes the monster position", rhinoStepCase{
		mode: monsterModeChargeTest, velocityY: 32 << 16, animLen: 8, wantTile: &[2]int{11, 11},
	}},
	{"a blocked landing converts the missile back", rhinoStepCase{
		mode: monsterModeChargeTest, velocityY: 32 << 16, animLen: 8,
		blocked: [][2]int{{11, 11}}, missToMonst: &[2]int{10, 10}, wantDel: true,
	}},
	{"a snake steps twice and stops on the doubled tile", rhinoStepCase{
		mode: monsterModeChargeTest, ai: monsterAISnakeTest, velocityY: 32 << 16, animLen: 8,
		blocked: [][2]int{{12, 12}}, missToMonst: &[2]int{10, 10}, wantDel: true,
	}},
	{"a snake past its last attack frame stops", rhinoStepCase{
		mode: monsterModeChargeTest, ai: monsterAISnakeTest, velocityY: 32 << 16,
		animFrame: 8, animLen: 8, missToMonst: &[2]int{10, 10}, wantDel: true,
	}},
	{"a snake with a clear doubled tile charges on", rhinoStepCase{
		mode: monsterModeChargeTest, ai: monsterAISnakeTest, velocityY: 32 << 16, animLen: 8,
		wantTile: &[2]int{11, 11},
	}},
}

type rhinoSpy struct {
	// animGraphic is the graphic the add asked the monster sheet for.
	animGraphic int
	// monsterTile is the tile SetMonsterTile wrote, nil when it was not called.
	monsterTile    *missiles.TilePosition
	occupyTiles    int
	missToMonstPos *missiles.TilePosition
	walkability    int
	grid           *world.Grid
}

type rhinoAddCase struct {
	monsterType int
	unique      bool
	lightID     int
	wantGraphic int
	wantFrame   int
}

type rhinoStepCase struct {
	mode int
	ai   int
	// velocityY is the 16.16 screen velocity; 32 << 16 is one tile south.
	velocityY int
	// animFrame/animLen drive the snake's out-of-frames test.
	animFrame int
	animLen   int
	// blocked holds the tiles IsTileAvailable refuses.
	blocked [][2]int
	// missToMonst is the tile MissToMonst must be handed.
	missToMonst *[2]int
	wantDel     bool
	// wantTile is the monster tile written on a clear landing.
	wantTile *[2]int
}

func TestAddRhino(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(addRhinoCases))
	for _, tc := range addRhinoCases {
		subtests = append(subtests, subtest{name: tc.name, run: runRhinoAdd(tc.cfg)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestProcessRhino(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(processRhinoCases))
	for _, tc := range processRhinoCases {
		subtests = append(subtests, subtest{name: tc.name, run: runRhinoStep(tc.cfg)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

func TestMonsterTypeConstsRhino(t *testing.T) {
	// given
	rows := monsterRowIndexes(t)

	// when
	got := map[string]int{
		"MT_HORNED": rows["MT_HORNED"],
		"MT_MUDRUN": rows["MT_MUDRUN"],
		"MT_FROSTC": rows["MT_FROSTC"],
		"MT_OBLORD": rows["MT_OBLORD"],
		"MT_NSNAKE": rows["MT_NSNAKE"],
		"MT_RSNAKE": rows["MT_RSNAKE"],
		"MT_BSNAKE": rows["MT_BSNAKE"],
		"MT_GSNAKE": rows["MT_GSNAKE"],
	}

	// then
	for name, want := range map[string]int{
		"MT_HORNED": missiles.MonsterTypeHorned,
		"MT_MUDRUN": missiles.MonsterTypeMudRun,
		"MT_FROSTC": missiles.MonsterTypeFrostC,
		"MT_OBLORD": missiles.MonsterTypeObLord,
		"MT_NSNAKE": missiles.MonsterTypeNSnake,
		"MT_RSNAKE": missiles.MonsterTypeRSnake,
		"MT_BSNAKE": missiles.MonsterTypeBSnake,
		"MT_GSNAKE": missiles.MonsterTypeGSnake,
	} {
		if got, ok := got[name]; !ok || got != want || got == 0 {
			t.Fatalf("%s ordinal = %d (found=%v), want %d", name, got, ok, want)
		}
	}
}

func rhinoWorld(spy *rhinoSpy, stats missiles.AddMonsterStats, mode, ai int,
	blocked [][2]int,
) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.MonsterMode = func(int) int { return mode }
	cb.MonsterAI = func(int) int { return ai }
	cb.SetMonsterTile = func(_ int, position missiles.TilePosition) { spy.monsterTile = &position }
	cb.OccupyTile = func(int, missiles.TilePosition, bool) { spy.occupyTiles++ }
	cb.MissToMonst = func(_ *missiles.Missile, position missiles.TilePosition) {
		spy.missToMonstPos = &position
	}
	cb.MarkWalkabilityChanged = func() { spy.walkability++ }
	cb.IsTileAvailable = func(_, x, y int) bool { return !isBlockedRhinoTile(blocked, x, y) }

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			MonsterStats: func(int) missiles.AddMonsterStats { return stats },
			MonsterAnimation: func(_, graphic int) missiles.AddMonsterAnimation {
				spy.animGraphic = graphic

				return missiles.AddMonsterAnimation{Rate: 3, Frames: 8}
			},
		},
	}
	spy.grid = w.Grid

	return w
}

func isBlockedRhinoTile(blocked [][2]int, x, y int) bool {
	for _, tile := range blocked {
		if tile[0] == x && tile[1] == y {
			return true
		}
	}

	return false
}

func runRhinoAdd(cfg rhinoAddCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spy := &rhinoSpy{}
		stats := missiles.AddMonsterStats{
			Type: cfg.monsterType, IsUnique: cfg.unique, LightID: cfg.lightID,
		}
		w := rhinoWorld(spy, stats, monsterModeChargeTest, 0, nil)
		tape := &rndTape{t: t} // the body draws no rnd
		w.Rnd = tape.draw

		m := missiles.Missile{Type: missiles.Rhino, Source: 3, Lid: missiles.NoLight}
		m.SetTraveledFromTile(10, 10)
		p := &missiles.AddParameter{Dst: missiles.TilePosition{X: 10, Y: 12}, Midir: missiles.South}

		// when
		missiles.AddRhino(w, &m, p)

		// then
		tape.assertConsumed()
		assertRhinoAdd(t, m, spy, cfg)
	}
}

func assertRhinoAdd(t *testing.T, m missiles.Missile, spy *rhinoSpy, cfg rhinoAddCase) {
	t.Helper()

	if spy.animGraphic != cfg.wantGraphic { // cpp:2268-2273
		t.Fatalf("anim graphic = %d, want %d", spy.animGraphic, cfg.wantGraphic)
	}

	if m.AnimDelay != 3 || m.AnimLen != 8 || m.AnimAdd != 1 { // cpp:2254-2256
		t.Fatalf("anim delay/len/add = %d/%d/%d, want 3/8/1", m.AnimDelay, m.AnimLen, m.AnimAdd)
	}

	if m.AnimFrame != cfg.wantFrame { // cpp:2276-2277
		t.Fatalf("animFrame = %d, want %d", m.AnimFrame, cfg.wantFrame)
	}

	if m.Duration != 256 || m.Var1 != 0 || m.Var2 != 0 || !m.LightFlag { // cpp:2257-2260
		t.Fatalf("duration/vars/lightFlag = %d/%d/%d/%v, want 256/0/0/true",
			m.Duration, m.Var1, m.Var2, m.LightFlag)
	}

	assertRhinoLight(t, m, cfg)
}

func assertRhinoLight(t *testing.T, m missiles.Missile, cfg rhinoAddCase) {
	t.Helper()

	want := missiles.NoLight
	if cfg.unique {
		want = cfg.lightID
	}

	if m.Lid != want {
		t.Fatalf("lid = %d, want %d", m.Lid, want)
	}
}

func runRhinoStep(cfg rhinoStepCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spy := &rhinoSpy{}
		w := rhinoWorld(spy, missiles.AddMonsterStats{Type: 0}, cfg.mode, cfg.ai, cfg.blocked)
		tape := &rndTape{t: t}
		w.Rnd = tape.draw

		m := missiles.Missile{Type: missiles.Rhino, Source: 3, AnimFrame: cfg.animFrame, AnimLen: cfg.animLen}
		m.SetTraveledFromTile(10, 10)
		m.Position.Velocity = missiles.PositionDisplacement{DeltaY: int32(cfg.velocityY)}
		w.DMonster[10][10] = 4 // the charge's own cell (id + 1)

		// when
		missiles.ProcessRhino(w, &m)

		// then
		tape.assertConsumed()
		assertRhinoStep(t, m, spy, cfg)
	}
}

func assertRhinoStep(t *testing.T, m missiles.Missile, spy *rhinoSpy, cfg rhinoStepCase) {
	t.Helper()

	if m.DelFlag != cfg.wantDel { // cpp:3739-3742,3757
		t.Fatalf("del = %v, want %v", m.DelFlag, cfg.wantDel)
	}

	if cfg.wantDel && cfg.missToMonst == nil {
		assertRhinoQuiet(t, spy)

		return
	}

	assertRhinoLanding(t, spy, cfg)
}

func assertRhinoQuiet(t *testing.T, spy *rhinoSpy) {
	t.Helper()

	if spy.walkability != 0 || spy.occupyTiles != 0 || spy.monsterTile != nil {
		t.Fatalf("early exit wrote the monster: walkability=%d occupy=%d tile=%v",
			spy.walkability, spy.occupyTiles, spy.monsterTile)
	}
}

func assertRhinoLanding(t *testing.T, spy *rhinoSpy, cfg rhinoStepCase) {
	t.Helper()

	if spy.walkability != 1 { // cpp:3745-3746
		t.Fatalf("walkability marks = %d, want 1", spy.walkability)
	}

	assertRhinoConversion(t, spy, cfg)

	if cfg.wantTile == nil {
		return
	}

	if spy.monsterTile == nil || *spy.monsterTile != (missiles.TilePosition{
		X: cfg.wantTile[0], Y: cfg.wantTile[1],
	}) { // cpp:3762-3764
		t.Fatalf("monster tile = %v, want %v", spy.monsterTile, *cfg.wantTile)
	}

	if spy.occupyTiles != 1 { // cpp:3765
		t.Fatalf("occupy calls = %d, want 1", spy.occupyTiles)
	}
}

func assertRhinoConversion(t *testing.T, spy *rhinoSpy, cfg rhinoStepCase) {
	t.Helper()

	if cfg.missToMonst == nil {
		if spy.missToMonstPos != nil {
			t.Fatalf("MissToMonst ran at %v, want no conversion", spy.missToMonstPos)
		}

		return
	}

	want := missiles.TilePosition{X: cfg.missToMonst[0], Y: cfg.missToMonst[1]}
	if spy.missToMonstPos == nil || *spy.missToMonstPos != want { // cpp:3756
		t.Fatalf("MissToMonst tile = %v, want %v", spy.missToMonstPos, want)
	}
}
