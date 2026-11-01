package game_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

// cheatResult is one Execute outcome: upstream's status message and error.
type cheatResult struct {
	msg string
	err error
}

type spawnRejectCase struct {
	name string
	env  game.DebugEnv
	line string
	want string
}

type jumpCase struct {
	name    string
	level   int
	current int
	want    string
	wantHop int // -1 means JumpLevel must not be called
}

type missingHookCase struct {
	name  string
	line  string
	clear func(*game.DebugEnv)
}

// debugHookSpy counts hook calls so the fail-closed tests can prove that a
// disabled router never reaches the game.
type debugHookSpy struct{ calls int }

func TestDebugRouter_SpawnMonster_PlacesInCrawlOrder(t *testing.T) {
	// given
	var spawned []game.DebugTile

	env := debugEnv(func(e *game.DebugEnv) {
		e.SpawnMonster = func(tile game.DebugTile, _, _ int) { spawned = append(spawned, tile) }
	})

	// when
	result := runCheat(env, "spawn-monster zombie 3")

	// then
	if result.err != nil {
		t.Fatalf("Execute() error = %v, want nil", result.err)
	}

	// crawl.cpp:15-45 visits (0,0), (0,1), (0,-1), (-1,0), (1,0), ... from the
	// player tile.
	want := []game.DebugTile{{X: 50, Y: 50}, {X: 50, Y: 51}, {X: 50, Y: 49}}
	assertTiles(t, spawned, want)

	if result.msg != "Spawned 3 monsters." {
		t.Fatalf("message = %q, want %q (monsters.cpp:167)", result.msg, "Spawned 3 monsters.")
	}
}

func TestDebugRouter_SpawnMonster_SkipsOccupiedAndUnwalkable(t *testing.T) {
	// given
	var spawned []game.DebugTile

	env := debugEnv(func(e *game.DebugEnv) {
		e.Occupied = func(tile game.DebugTile) bool { return tile == game.DebugTile{X: 50, Y: 50} }
		e.Walkable = func(tile game.DebugTile) bool { return tile != game.DebugTile{X: 50, Y: 51} }
		e.SpawnMonster = func(tile game.DebugTile, _, _ int) { spawned = append(spawned, tile) }
	})

	// when
	result := runCheat(env, "spawn-monster zombie 2")

	// then
	if result.err != nil {
		t.Fatalf("Execute() error = %v, want nil", result.err)
	}

	// (0,0) is occupied and (0,1) is not walkable (monsters.cpp:154-157), so
	// the next two crawl positions win.
	assertTiles(t, spawned, []game.DebugTile{{X: 50, Y: 49}, {X: 49, Y: 50}})
}

func TestDebugRouter_SpawnMonster_DefaultCountIsOne(t *testing.T) {
	// given
	spawned := 0

	env := debugEnv(func(e *game.DebugEnv) {
		e.SpawnMonster = func(game.DebugTile, int, int) { spawned++ }
	})

	// when
	result := runCheat(env, "spawn-monster zombie")

	// then
	if result.err != nil {
		t.Fatalf("Execute() error = %v, want nil", result.err)
	}

	if spawned != 1 || result.msg != "Spawned 1 monsters." {
		t.Fatalf("spawned %d with message %q, want 1 and %q (monsters.cpp:117)", spawned, result.msg, "Spawned 1 monsters.")
	}
}

func TestDebugRouter_SpawnMonster_Rejections(t *testing.T) {
	// given
	cases := []spawnRejectCase{
		{
			name: "TownRejects",
			env:  debugEnv(func(e *game.DebugEnv) { e.LevelType = func() game.LevelType { return game.LevelTypeTown } }),
			line: "spawn-monster zombie",
			want: "Can't spawn monsters in town", // monsters.cpp:26
		},
		{
			name: "UnknownName",
			env:  debugEnv(func(e *game.DebugEnv) { e.MonsterTypeID = func(string) int { return -1 } }),
			line: "spawn-monster nosuchmonster",
			want: "Monster not found", // monsters.cpp:122
		},
		{
			name: "NoFreeSlots",
			env:  debugEnv(func(e *game.DebugEnv) { e.ActiveMonsters = func() int { return game.MaxMonsters } }),
			line: "spawn-monster zombie",
			want: "Can't spawn any monsters", // monsters.cpp:151
		},
		{
			name: "CrawlRunsDry",
			env:  debugEnv(func(e *game.DebugEnv) { e.Walkable = func(game.DebugTile) bool { return false } }),
			line: "spawn-monster zombie 2",
			want: "Spawned 0 monsters.", // monsters.cpp:165 suffixes only the clamp branch
		},
	}

	// when
	results := runSpawnRejections(cases)

	// then
	for i, tt := range cases {
		t.Run(tt.name, func(t *testing.T) { assertMessage(t, tt.line, results[i], tt.want) })
	}
}

func TestDebugRouter_SpawnMonster_ClampedByFreeSlots(t *testing.T) {
	// given
	spawned := 0

	env := debugEnv(func(e *game.DebugEnv) {
		e.ActiveMonsters = func() int { return game.MaxMonsters - 1 }
		e.SpawnMonster = func(game.DebugTile, int, int) { spawned++ }
	})

	// when
	result := runCheat(env, "spawn-monster zombie 5")

	// then
	if result.err != nil {
		t.Fatalf("Execute() error = %v, want nil", result.err)
	}

	// monsters.cpp:149 clamps to MaxMonsters - ActiveMonsterCount and
	// monsters.cpp:165 reports the clamp through the message suffix.
	if spawned != 1 {
		t.Fatalf("spawned %d monsters, want 1 (one free slot)", spawned)
	}

	if result.msg != "Spawned 1 monsters. (Unable to spawn more)" {
		t.Fatalf("message = %q, want the clamped message", result.msg)
	}
}

func TestDebugRouter_SpawnMonster_KeepsLCGState(t *testing.T) {
	// given: the placement crawl (monsters.cpp:152-164) draws no random values;
	// only SpawnMonster (monster.cpp) rolls, and that is the hook's job.
	engine.SetRndSeed(0x1234_5678)
	before := engine.GetLCGEngineState()

	// when
	result := runCheat(debugEnv(), "spawn-monster zombie 3")

	// then
	if result.err != nil {
		t.Fatalf("Execute() error = %v, want nil", result.err)
	}

	if after := engine.GetLCGEngineState(); after != before {
		t.Fatalf("LCG state %#x after spawn-monster, want %#x (crawl is RNG-free)", after, before)
	}
}

func TestDebugRouter_SpawnItem_ReportsIterations(t *testing.T) {
	// given
	var gotName string

	env := debugEnv(func(e *game.DebugEnv) {
		e.SpawnItem = func(name string) (int, error) {
			gotName = name

			return 7, nil
		}
	})

	// when
	result := runCheat(env, "spawn-item short sword")

	// then
	if result.err != nil {
		t.Fatalf("Execute() error = %v, want nil", result.err)
	}

	if gotName != "short sword" {
		t.Fatalf("SpawnItem name = %q, want %q", gotName, "short sword")
	}

	if result.msg != "Item generated successfully - iterations: 7" {
		t.Fatalf("message = %q, want upstream's iteration message (items.cpp:118)", result.msg)
	}
}

func TestDebugRouter_SpawnItem_HookErrorIsWrapped(t *testing.T) {
	// given
	hookErr := errors.New("no space for another item")
	env := debugEnv(func(e *game.DebugEnv) {
		e.SpawnItem = func(string) (int, error) { return 0, hookErr }
	})

	// when
	result := runCheat(env, "spawn-item sword")

	// then
	if !errors.Is(result.err, hookErr) {
		t.Fatalf("Execute() error = %v, want it to wrap %v", result.err, hookErr)
	}

	if result.msg != "" {
		t.Fatalf("Execute() = %q, want no message", result.msg)
	}
}

func TestDebugRouter_SpawnMonster_MissingSlotHookFailsClosed(t *testing.T) {
	// given: without ActiveMonsters the router cannot know the free slots.
	env := debugEnv(func(e *game.DebugEnv) { e.ActiveMonsters = nil })

	// when
	result := runCheat(env, "spawn-monster zombie")

	// then
	assertDisabled(t, "spawn-monster zombie", result)
}

func TestDebugRouter_JumpLevel_Table(t *testing.T) {
	// given
	cases := []jumpCase{
		{name: "MovesToLevel", level: 5, current: 3, want: "Moved you to level 5.", wantHop: 5},
		{name: "MaxLevelBound", level: 16, current: 1, want: "Moved you to level 16.", wantHop: 16},
		{name: "AboveMaxLevel", level: 17, current: 3, want: "Level 17 does not exist!", wantHop: -1},
		{name: "AlreadyThere", level: 3, current: 3, want: "You are already on level 3!", wantHop: -1},
	}

	// when
	hops, results := runJumpLevels(cases)

	// then
	for i, tt := range cases {
		t.Run(tt.name, func(t *testing.T) { assertJumpLevel(t, tt, hops[i], results[i]) })
	}
}

func TestDebugRouter_Disabled_FailsClosed(t *testing.T) {
	// given: a build without a router is the shipped default (_DEBUG is off).
	var router *game.DebugRouter

	hooks := &debugHookSpy{}
	debugEnv(func(e *game.DebugEnv) {
		e.SpawnMonster = hooks.spawnMonster
		e.SpawnItem = hooks.spawnItem
		e.JumpLevel = hooks.jumpLevel
	})

	lines := []string{"spawn-monster zombie", "spawn-item short sword", "jump-level 5"}

	// when
	results := runLines(router, lines)

	// then
	for i, line := range lines {
		assertDisabled(t, line, results[i])
	}

	if hooks.calls != 0 {
		t.Fatalf("hooks called %d times while disabled, want 0", hooks.calls)
	}
}

func TestDebugRouter_MissingHook_FailsClosed(t *testing.T) {
	// given
	cases := []missingHookCase{
		{name: "SpawnMonster", line: "spawn-monster zombie", clear: func(e *game.DebugEnv) { e.SpawnMonster = nil }},
		{name: "SpawnItem", line: "spawn-item sword", clear: func(e *game.DebugEnv) { e.SpawnItem = nil }},
		{name: "JumpLevel", line: "jump-level 5", clear: func(e *game.DebugEnv) { e.JumpLevel = nil }},
	}

	// when
	results := runMissingHookCases(cases)

	// then
	for i, tt := range cases {
		t.Run(tt.name, func(t *testing.T) { assertDisabled(t, tt.line, results[i]) })
	}
}

func TestDebugRouter_MalformedCommandsError(t *testing.T) {
	// given
	lines := []string{
		"spawn-monster", "spawn-item", "jump-level", "jump-level zero", "jump-level 0", "bogus command", "",
	}
	router := game.NewDebugRouter(debugEnv())

	// when
	results := runLines(router, lines)

	// then: malformed commands are errors and never reach a hook.
	for i, line := range lines {
		if results[i].err == nil {
			t.Fatalf("Execute(%q) error = nil, want an error", line)
		}

		if results[i].msg != "" {
			t.Fatalf("Execute(%q) = %q, want no message", line, results[i].msg)
		}
	}
}

// debugEnv builds a fully wired DebugEnv over an open dungeon around tile
// (50, 50); each override mutates one hook.
func debugEnv(overrides ...func(*game.DebugEnv)) game.DebugEnv {
	env := game.DebugEnv{
		LevelType:      func() game.LevelType { return 1 }, // any non-town leveltype
		PlayerTile:     func() game.DebugTile { return game.DebugTile{X: 50, Y: 50} },
		PlayerDir:      func() int { return 0 },
		Occupied:       func(game.DebugTile) bool { return false },
		Walkable:       func(game.DebugTile) bool { return true },
		ActiveMonsters: func() int { return 0 },
		MonsterTypeID:  func(string) int { return 7 },
		SpawnMonster:   func(game.DebugTile, int, int) {},
		SpawnItem:      func(string) (int, error) { return 7, nil },
		CurrentLevel:   func() int { return 3 },
		MaxLevel:       16,
		JumpLevel:      func(int) error { return nil },
	}

	for _, override := range overrides {
		override(&env)
	}

	return env
}

// runCheat executes one command line against a router over env.
func runCheat(env game.DebugEnv, line string) cheatResult {
	return runLines(game.NewDebugRouter(env), []string{line})[0]
}

// runLines executes lines on router and collects the outcomes.
func runLines(router *game.DebugRouter, lines []string) []cheatResult {
	results := make([]cheatResult, len(lines))
	for i, line := range lines {
		msg, err := router.Execute(line)
		results[i] = cheatResult{msg: msg, err: err}
	}

	return results
}

// runSpawnRejections executes the monster-spawn rejection table.
func runSpawnRejections(cases []spawnRejectCase) []cheatResult {
	results := make([]cheatResult, len(cases))
	for i, tt := range cases {
		results[i] = runCheat(tt.env, tt.line)
	}

	return results
}

// runMissingHookCases executes the table of commands with one hook removed.
func runMissingHookCases(cases []missingHookCase) []cheatResult {
	results := make([]cheatResult, len(cases))
	for i, tt := range cases {
		results[i] = runCheat(debugEnv(tt.clear), tt.line)
	}

	return results
}

// runJumpLevels records the level each jump-level case asked the game to load
// (-1 when it asked for nothing).
func runJumpLevels(cases []jumpCase) ([]int, []cheatResult) {
	hops := make([]int, len(cases))
	results := make([]cheatResult, len(cases))

	for i, tt := range cases {
		hopped := -1
		env := debugEnv(func(e *game.DebugEnv) {
			e.CurrentLevel = func() int { return tt.current }
			e.JumpLevel = func(level int) error { hopped = level; return nil }
		})

		results[i] = runCheat(env, "jump-level "+strconv.Itoa(tt.level))
		hops[i] = hopped
	}

	return hops, results
}

func assertTiles(t *testing.T, got, want []game.DebugTile) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("spawned %d tiles %v, want %d %v", len(got), got, len(want), want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("spawned tile %d = %v, want %v", i, got[i], want[i])
		}
	}
}

func assertMessage(t *testing.T, line string, result cheatResult, want string) {
	t.Helper()

	if result.err != nil {
		t.Fatalf("Execute(%q) error = %v, want nil", line, result.err)
	}

	if result.msg != want {
		t.Fatalf("Execute(%q) = %q, want %q", line, result.msg, want)
	}
}

func assertDisabled(t *testing.T, line string, result cheatResult) {
	t.Helper()

	if !errors.Is(result.err, game.ErrDebugDisabled) {
		t.Fatalf("Execute(%q) error = %v, want ErrDebugDisabled", line, result.err)
	}

	if result.msg != "" {
		t.Fatalf("Execute(%q) = %q, want no message", line, result.msg)
	}
}

func assertJumpLevel(t *testing.T, tt jumpCase, hopped int, result cheatResult) {
	t.Helper()

	line := "jump-level " + strconv.Itoa(tt.level)
	assertMessage(t, line, result, tt.want)

	if hopped != tt.wantHop {
		t.Fatalf("JumpLevel(%d), want %d", hopped, tt.wantHop)
	}
}

func (spy *debugHookSpy) spawnMonster(game.DebugTile, int, int) { spy.calls++ }

func (spy *debugHookSpy) spawnItem(string) (int, error) { spy.calls++; return 1, nil }

func (spy *debugHookSpy) jumpLevel(int) error { spy.calls++; return nil }
