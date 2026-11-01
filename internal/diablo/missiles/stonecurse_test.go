// Package missiles_test: the Stone Curse row (add search + petrify life).
package missiles_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// Monster modes the row switches on (monster.h:77-99).
const (
	monsterModeStand     = 0
	monsterModeFadeIn    = 8
	monsterModeCharge    = 14
	monsterModePetrified = 15
)

// stoneCurseShatterGraphic is MissileGraphicID::StoneCurseShatter, the 19th
// entry of an enum that starts at Arrow (tables/misdat.h:36,54).
const stoneCurseShatterGraphic missiles.GraphicID = 18

// hellfireMonstdatPath is the hellfire monster table fixture.
const hellfireMonstdatPath = "../../../internal/testdata/gamedata/monstdat/hf-monstdat.tsv"

// stoneCurseAim is the aim tile of the add runs: the monster stands one row
// below the caster at (10, 10).
var stoneCurseAim = missiles.TilePosition{X: 10, Y: 12}

// stoneCurseTicks is the tick table: the petrify life, the shatter swap, the
// shatter's own last tick and a monster that left the mode.
var stoneCurseTicks = []struct {
	name string
	cfg  stoneCurseTick
}{
	{"a living monster is unpetrified on its last tick", stoneCurseTick{
		duration: 1,
		want:     stoneCurseTickWant{del: true, modes: []int{monsterModeStand}, petrify: []bool{false}},
	}},
	{"a monster that died swaps to the shatter frames", stoneCurseTick{
		duration: 2, dead: true,
		want: stoneCurseTickWant{draw: true, shatter: true, duration: 11, flagged: true},
	}},
	{"the shatter's last tick drops a corpse", stoneCurseTick{
		duration: 1, dead: true, shatter: true,
		// PutMissile no-ops on a deleted missile (cpp:198-200)
		want: stoneCurseTickWant{del: true, shatter: true, corpses: 1},
	}},
	{"a monster that left the petrified mode ends the missile", stoneCurseTick{
		duration: 2, unpetrified: true,
		want: stoneCurseTickWant{del: true, duration: 1},
	}},
}

// stoneCurseSpy records the monster-store writes one body made.
type stoneCurseSpy struct {
	modes   []int
	petrify []bool
	corpses int
	// grid is the world grid, so the PutMissile mark is observable.
	grid *world.Grid
}

// stoneCurseCase is one add-body run: the monster's mode and type, and whether
// the aim tile holds a monster at all.
type stoneCurseCase struct {
	mode int
	// monsterType is the target's type id; 0 is a common monster.
	monsterType int
	// empty leaves the aim tile without a monster, so the search fails.
	empty bool
	// wantDel marks the fizzle paths: nothing found, already petrified.
	wantDel bool
	// wantFizzled is parameter.spellFizzled after the run.
	wantFizzled bool
}

// stoneCurseTick is one process tick: the monster's state, the missile's
// remaining life, and the state the tick must leave behind.
type stoneCurseTick struct {
	// duration is the missile's life before the tick.
	duration int
	// dead gives the monster 0 hit points.
	dead bool
	// shatter starts the missile on the shatter sheet.
	shatter bool
	// unpetrified leaves the monster out of the petrified mode.
	unpetrified bool
	want        stoneCurseTickWant
}

// stoneCurseTickWant is the state one tick must leave behind.
type stoneCurseTickWant struct {
	del bool
	// draw is the _miDrawFlag the shatter swap raises (cpp:3700).
	draw bool
	// shatter is the animation sheet after the tick.
	shatter  bool
	duration int
	// modes and petrify are the store writes; nil means no write.
	modes   []int
	petrify []bool
	corpses int
	// flagged is dFlags after the tick: PutMissile ran or it did not.
	flagged bool
}

// TestAddStoneCurse pins the search, the fizzle, the petrify and the life
// (cpp:2365-2416).
func TestAddStoneCurse(t *testing.T) {
	// when
	subtests := []subtest{
		{"an idle monster is petrified", runStoneCurseAdd(stoneCurseCase{mode: monsterModeStand})},
		{"an empty aim tile fizzles the spell", runStoneCurseAdd(stoneCurseCase{
			empty: true, wantDel: true, wantFizzled: true,
		})},
		{"an already petrified monster is left alone", runStoneCurseAdd(stoneCurseCase{
			mode: monsterModePetrified, wantDel: true,
		})},
		{"a golem is refused", runStoneCurseAdd(stoneCurseCase{
			mode: monsterModeStand, monsterType: missiles.MonsterTypeGolem, wantDel: true, wantFizzled: true,
		})},
		{"a charging monster is refused", runStoneCurseAdd(stoneCurseCase{
			mode: monsterModeCharge, wantDel: true, wantFizzled: true,
		})},
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestProcessStoneCurse pins the shatter frame, the unpetrify and the corpse
// (cpp:3694-3720).
func TestProcessStoneCurse(t *testing.T) {
	// when
	subtests := make([]subtest, 0, len(stoneCurseTicks))
	for _, tick := range stoneCurseTicks {
		subtests = append(subtests, subtest{name: tick.name, run: runStoneCurseTick(tick.cfg)})
	}

	// then
	for _, st := range subtests {
		t.Run(st.name, st.run)
	}
}

// TestMonsterTypeConstsStoneCurse pins the three refused ordinals to the
// committed monstdat.tsv and hf-monstdat.tsv row order
// (tables/monstdat.h:270-271,298).
func TestMonsterTypeConstsStoneCurse(t *testing.T) {
	// given
	rows := monsterRowIndexes(t)
	hellfireRows := hellfireMonsterRowIndexes(t)

	// when
	got := map[string]int{
		"MT_GOLEM":  rows["MT_GOLEM"],
		"MT_DIABLO": rows["MT_DIABLO"],
		"MT_NAKRUL": hellfireRows["MT_NAKRUL"],
	}

	// then
	for name, want := range map[string]int{
		"MT_GOLEM":  missiles.MonsterTypeGolem,
		"MT_DIABLO": missiles.MonsterTypeDiablo,
		"MT_NAKRUL": missiles.MonsterTypeNakrul,
	} {
		if got, ok := got[name]; !ok || got != want || got == 0 {
			t.Fatalf("%s ordinal = %d (found=%v), want %d", name, got, ok, want)
		}
	}
}

// stoneCurseWorld wires the monster-store sinks and the aim monster.
func stoneCurseWorld(tc stoneCurseCase, spy *stoneCurseSpy) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.MonsterMode = func(int) int { return tc.mode }
	cb.SetMonsterMode = func(_, mode int) { spy.modes = append(spy.modes, mode) }
	cb.SetMonsterPetrified = func(_ int, petrified bool) { spy.petrify = append(spy.petrify, petrified) }
	cb.AddCorpse = func(int) { spy.corpses++ }

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		AddSource: missiles.AddSource{
			MonsterStats: func(int) missiles.AddMonsterStats {
				return missiles.AddMonsterStats{Type: tc.monsterType}
			},
		},
	}
	spy.grid = w.Grid

	if !tc.empty {
		w.DMonster[stoneCurseAim.Y][stoneCurseAim.X] = 8 // dMonster holds id + 1
	}

	return w
}

// runStoneCurseAdd drives one AddStoneCurse.
func runStoneCurseAdd(tc stoneCurseCase) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spy := &stoneCurseSpy{}
		w := stoneCurseWorld(tc, spy)
		tape := &rndTape{t: t} // the body draws no rnd

		m := missiles.Missile{Type: missiles.StoneCurse, Source: 0, SpellLevel: 5}
		m.Position.Start = missiles.TilePosition{X: 10, Y: 10}
		p := &missiles.AddParameter{Dst: stoneCurseAim}

		// when
		missiles.AddStoneCurse(w, &m, p)

		// then
		tape.assertConsumed()
		assertStoneCurseAdd(t, m, spy, p, tc)
	}
}

// assertStoneCurseAdd checks the fizzle path and the armed missile.
func assertStoneCurseAdd(t *testing.T, m missiles.Missile, spy *stoneCurseSpy, p *missiles.AddParameter,
	tc stoneCurseCase,
) {
	t.Helper()

	if m.DelFlag != tc.wantDel {
		t.Fatalf("del = %v, want %v", m.DelFlag, tc.wantDel)
	}

	if p.SpellFizzled != tc.wantFizzled {
		t.Fatalf("spellFizzled = %v, want %v", p.SpellFizzled, tc.wantFizzled)
	}

	if tc.wantDel {
		assertStoneCurseQuiet(t, spy)

		return
	}

	assertStoneCurseSaved(t, m)
	assertStoneCursePetrify(t, spy)
	assertStoneCurseLife(t, m)
}

// assertStoneCurseQuiet checks that a deleted missile wrote nothing.
func assertStoneCurseQuiet(t *testing.T, spy *stoneCurseSpy) {
	t.Helper()

	if len(spy.modes) != 0 || len(spy.petrify) != 0 {
		t.Fatalf("deleted missile wrote the monster: modes=%v petrify=%v", spy.modes, spy.petrify)
	}
}

// assertStoneCurseSaved checks the mode and the id the caster saved.
func assertStoneCurseSaved(t *testing.T, m missiles.Missile) {
	t.Helper()

	if m.Var1 != monsterModeStand { // cpp:2405
		t.Fatalf("var1 = %d, want the saved mode %d", m.Var1, monsterModeStand)
	}

	if m.Var2 != 7 { // cpp:2406, dMonster 8 is monster 7
		t.Fatalf("var2 = %d, want 7", m.Var2)
	}
}

// assertStoneCursePetrify checks the petrify writes.
func assertStoneCursePetrify(t *testing.T, spy *stoneCurseSpy) {
	t.Helper()

	if !slices.Equal(spy.modes, []int{monsterModePetrified}) { // cpp:2407
		t.Fatalf("modes = %v, want [%d]", spy.modes, monsterModePetrified)
	}

	if !slices.Equal(spy.petrify, []bool{true}) { // cpp:2407, monster.cpp:4929
		t.Fatalf("petrify = %v, want [true]", spy.petrify)
	}
}

// assertStoneCurseLife checks the landing tile and the petrified life.
func assertStoneCurseLife(t *testing.T, m missiles.Missile) {
	t.Helper()

	if m.Position.Tile != stoneCurseAim || m.Position.Start != stoneCurseAim { // cpp:2410-2411
		t.Fatalf("tile/start = %v/%v, want %v", m.Position.Tile, m.Position.Start, stoneCurseAim)
	}

	if m.Duration != 176 { // cpp:2412-2415, min(5+6, 15) << 4
		t.Fatalf("duration = %d, want 176", m.Duration)
	}
}

// runStoneCurseTick drives one ProcessStoneCurse.
func runStoneCurseTick(cfg stoneCurseTick) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		// given
		spy := &stoneCurseSpy{}
		tape := &rndTape{t: t} // the body draws no rnd
		w := stoneCurseTickWorld(cfg, spy, tape)
		missile := stoneCurseTickMissile(cfg)

		// when
		missiles.ProcessStoneCurse(w, &missile)

		// then
		tape.assertConsumed()
		assertStoneCurseTick(t, missile, spy, cfg.want)
	}
}

// stoneCurseTickWorld wires the monster store for one process tick.
func stoneCurseTickWorld(cfg stoneCurseTick, spy *stoneCurseSpy, tape *rndTape) *missiles.HitWorld {
	cb := missiles.NoopProcessCallbacks()
	cb.MonsterMode = func(int) int { return monsterModePetrified }
	cb.MonsterHitPoints = func(int) int { return 30 }
	cb.SetMonsterMode = func(_, mode int) { spy.modes = append(spy.modes, mode) }
	cb.SetMonsterPetrified = func(_ int, petrified bool) { spy.petrify = append(spy.petrify, petrified) }
	cb.AddCorpse = func(int) { spy.corpses++ }

	if cfg.unpetrified {
		cb.MonsterMode = func(int) int { return monsterModeStand }
	}

	if cfg.dead {
		cb.MonsterHitPoints = func(int) int { return 0 }
	}

	w := &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
		Rnd:              tape.draw,
	}
	spy.grid = w.Grid

	return w
}

// stoneCurseTickMissile is the petrified missile one tick starts from: var1 is
// the mode to restore and var2 the petrified monster.
func stoneCurseTickMissile(cfg stoneCurseTick) missiles.Missile {
	m := missiles.Missile{Type: missiles.StoneCurse, Source: 0, Duration: cfg.duration, Var1: monsterModeStand, Var2: 7}
	m.Position.Tile = stoneCurseAim

	if cfg.shatter {
		m.AnimType = int(stoneCurseShatterGraphic)
	}

	return m
}

// assertStoneCurseTick checks the shatter swap, the delete and the store writes.
func assertStoneCurseTick(t *testing.T, m missiles.Missile, spy *stoneCurseSpy, want stoneCurseTickWant) {
	t.Helper()

	if m.DelFlag != want.del {
		t.Fatalf("del = %v, want %v", m.DelFlag, want.del)
	}

	if m.DrawFlag != want.draw {
		t.Fatalf("drawFlag = %v, want %v", m.DrawFlag, want.draw)
	}

	if want.shatter && missiles.GraphicID(m.AnimType) != stoneCurseShatterGraphic { // cpp:3701
		t.Fatalf("animType = %d, want the shatter sheet %d", m.AnimType, stoneCurseShatterGraphic)
	}

	if m.Duration != want.duration { // cpp:3702
		t.Fatalf("duration = %d, want %d", m.Duration, want.duration)
	}

	assertStoneCurseStore(t, spy, want)
}

// assertStoneCurseStore checks the monster writes and the PutMissile mark.
func assertStoneCurseStore(t *testing.T, spy *stoneCurseSpy, want stoneCurseTickWant) {
	t.Helper()

	if !slices.Equal(spy.modes, want.modes) { // cpp:3712
		t.Fatalf("modes = %v, want %v", spy.modes, want.modes)
	}

	if !slices.Equal(spy.petrify, want.petrify) { // cpp:3713
		t.Fatalf("petrify = %v, want %v", spy.petrify, want.petrify)
	}

	if spy.corpses != want.corpses { // cpp:3715
		t.Fatalf("corpses = %d, want %d", spy.corpses, want.corpses)
	}

	flagged := spy.grid.DFlags[stoneCurseAim.Y][stoneCurseAim.X]&world.FlagMissile != 0
	if flagged != want.flagged { // cpp:3718-3719
		t.Fatalf("missile tile flagged = %v, want %v", flagged, want.flagged)
	}
}

// hellfireMonsterRowIndexes maps each hf-monstdat.tsv type name to its data row
// index; MT_NAKRUL is a hellfire-only row, so only that table carries it.
func hellfireMonsterRowIndexes(t *testing.T) map[string]int {
	t.Helper()

	data, err := os.ReadFile(hellfireMonstdatPath)
	if err != nil {
		t.Fatalf("read hellfire monstdat fixture: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	rows := make(map[string]int, len(lines)-1)

	for i, line := range lines[1:] {
		rows[strings.SplitN(line, "\t", 2)[0]] = i
	}

	return rows
}
