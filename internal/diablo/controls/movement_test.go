// Package controls ports the fork gamepad stick semantics: StickPush
// classification, 8-way sectoring, the FaceDir table, and the
// Movement/TurnToDir/WalkInDir decision pipeline
// (Source/controls/plrctrls.cpp:100-105,1704-1754,1832-1893;
// Source/controls/controller_motion.cpp:210-242 @ fork ca91d778f).
package controls_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
)

var turnToDirCases = []struct {
	name       string
	setup      func(*world)
	dir        controls.Direction
	wantDir    controls.Direction
	wantStands int
}{
	{
		name:       "no-op while walking",
		setup:      func(w *world) { w.mode = controls.ModeWalk },
		dir:        controls.North,
		wantDir:    controls.South,
		wantStands: 0,
	},
	{
		name:       "no-op mid-attack below AFNum",
		setup:      func(w *world) { w.mode, w.current, w.afNum = controls.ModeAttack, 2, 6 },
		dir:        controls.North,
		wantDir:    controls.South,
		wantStands: 0,
	},
	{
		name:       "acts at attack frame >= AFNum",
		setup:      func(w *world) { w.mode, w.current, w.afNum = controls.ModeAttack, 6, 6 },
		dir:        controls.North,
		wantDir:    controls.North,
		wantStands: 1,
	},
	{
		name:       "no-op mid-spell below SFNum",
		setup:      func(w *world) { w.mode, w.current, w.sfNum = controls.ModeSpell, 1, 8 },
		dir:        controls.North,
		wantDir:    controls.South,
		wantStands: 0,
	},
	{
		name:       "acts at spell frame >= SFNum",
		setup:      func(w *world) { w.mode, w.current, w.sfNum = controls.ModeSpell, 8, 8 },
		dir:        controls.North,
		wantDir:    controls.North,
		wantStands: 1,
	},
	{
		name:       "acts when standing",
		setup:      func(*world) {},
		dir:        controls.NorthEast,
		wantDir:    controls.NorthEast,
		wantStands: 1,
	},
	{
		name:       "no-op on NoDirection",
		setup:      func(*world) {},
		dir:        controls.NoDirection,
		wantDir:    controls.South,
		wantStands: 0,
	},
}

// movementOverlayCases pins the Movement guards (plrctrls.cpp:1834-1838):
// each overlay consumer takes the stick before movement runs, so a guarded
// frame must produce no hook calls at all.
var movementOverlayCases = []struct {
	name  string
	guard func(controls.StickInput) controls.StickInput
}{
	{
		name:  "menu navigator overlay blocks everything",
		guard: func(in controls.StickInput) controls.StickInput { in.MenuNavigatorActive = true; return in },
	},
	{
		name:  "hotspell overlay blocks everything",
		guard: func(in controls.StickInput) controls.StickInput { in.HotspellMenuActive = true; return in },
	},
	{
		name:  "in-game menu blocks everything",
		guard: func(in controls.StickInput) controls.StickInput { in.InGameMenu = true; return in },
	},
	{name: "game UI handler set blocks everything", guard: func(in controls.StickInput) controls.StickInput {
		in.GameUIHandler = func(controls.AxisDirection) {}
		return in
	}},
}

// walkInDirCases pins WalkInDir's branches (plrctrls.cpp:1725-1747).
var walkInDirCases = []walkDirCase{
	{
		name:  "NONE axis stops an active walk when no dest action (CMD_WALKXY to future)",
		setup: func(w *world) { w.walkpath0 = 3; w.futureX, w.futureY = 12, 34 },
		axis:  noneAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			assertStopOnly(t, w, 12, 34)
		},
	},
	{
		name:  "NONE axis with idle walkpath emits nothing",
		setup: func(*world) {},
		axis:  noneAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			assertNoMovementOutput(t, w)
		},
	},
	{
		name:  "NONE axis with pending dest action emits nothing",
		setup: func(w *world) { w.walkpath0 = 3; w.destAction = 4 },
		axis:  noneAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			assertNoMovementOutput(t, w)
		},
	},
	{
		name:  "firm push standing: facing set, walk command to future+delta (East = {1,-1})",
		setup: func(*world) {},
		axis:  eastAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			if w.dir != controls.East {
				t.Fatalf("dir = %v, want East", w.dir)
			}
			assertWalkOnly(t, w, 11, 9)
		},
	},
	{
		name:  "facing not updated while walking",
		setup: func(w *world) { w.mode = controls.ModeWalk; w.dir = controls.West },
		axis:  eastAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			if w.dir != controls.West {
				t.Fatalf("dir = %v, want West (unchanged while walking)", w.dir)
			}
			assertWalkOnly(t, w, 11, 9)
		},
	},
	{
		name: "path blocked on a cardinal step: StartStand, no command",
		setup: func(w *world) {
			// The step tile (11,9) itself is PosOk — the upstream
			// `PosOkPlayer(delta) &&` gate — while both corner tiles are
			// solid, so the blocked branch is reached.
			w.posOkFn = func(x, y int) bool { return x == 11 && y == 9 }
			w.notSolidFn = func(int, int) bool { return false }
		},
		axis: eastAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			assertStoodFacing(t, w, controls.East)
		},
	},
	{
		name:  "path not blocked on a cardinal step: walk command emitted",
		setup: func(w *world) { w.notSolidFn = func(int, int) bool { return true } },
		axis:  eastAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			assertWalkOnly(t, w, 11, 9)
		},
	},
	{
		name: "diagonal push never runs the corner check",
		setup: func(w *world) {
			w.futureY = 21 // distinct origin so the walk target pins position.future use
			w.posOkFn = func(int, int) bool { return false }
			w.notSolidFn = func(int, int) bool { return false }
		},
		axis: northEastAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			if w.notSolCalls != 0 {
				t.Fatalf("IsTileNotSolid called %d times for a diagonal, want 0", w.notSolCalls)
			}
			assertWalkOnly(t, w, 10, 20)
		},
	},
	{
		name:  "standing ground: StartStand only, no walk command",
		setup: func(w *world) { w.standing = true },
		axis:  eastAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			assertStoodFacing(t, w, controls.East)
		},
	},
	{
		name: "standing ground while walking: no StartStand, no walk (branch returns)",
		setup: func(w *world) {
			// Upstream gates StartStand on _pmode == PM_STAND, then returns;
			// the walk command is NOT emitted (plrctrls.cpp:1741-1745).
			w.mode = controls.ModeWalk
			w.standing = true
		},
		axis: eastAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			assertNoMovementOutput(t, w)
		},
	},
	{
		name:  "step tile not PosOk: walk command emitted (fall through)",
		setup: func(w *world) { w.posOkFn = func(int, int) bool { return false } },
		axis:  eastAxis,
		assert: func(t *testing.T, w *world) {
			t.Helper()
			assertWalkOnly(t, w, 11, 9)
		},
	},
}

// world is the PlayerView + hook harness a Movement-family test runs
// against; every hook records its calls so assertions can pin emission
// order and absence.
type world struct {
	mode        controls.PlrMode
	current     int8 // AnimInfo.CurrentFrame
	afNum       int8
	sfNum       int8
	dir         controls.Direction
	futureX     int
	futureY     int
	walkpath0   int8
	destAction  int8
	standing    bool // IsStandingGround
	posOkFn     func(x, y int) bool
	notSolidFn  func(x, y int) bool
	standCalls  []controls.Direction
	walkCalls   [][2]int
	stopCalls   [][2]int
	posOkCalls  int
	notSolCalls int
}

// playerView adapts the harness world to controls.PlayerView.
type playerView struct {
	w *world
}

type walkDirCase struct {
	name   string
	setup  func(*world)
	axis   func() controls.AxisDirection
	assert func(t *testing.T, w *world)
}

// T6b — Movement overlay guards.
func TestMovementOverlayGuards(t *testing.T) {
	for _, tc := range movementOverlayCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := newWorld()
			// when
			controls.Movement(w.View(), tc.guard(controls.StickInput{X: 1, Y: 1}), w.Hooks())
			// then
			assertNoMovementOutput(t, w)
		})
	}
}

// T7 — IsStickMovementSignificant (plrctrls.cpp:1890-1893): firm left-stick
// push or any right-stick deflection counts; weak push with idle right stick
// does not.
func TestIsStickMovementSignificant(t *testing.T) {
	tests := []struct {
		name string
		in   controls.StickInput
		want bool
	}{
		{name: "idle sticks → false", in: controls.StickInput{}, want: false},
		{name: "weak left push, idle right → false", in: controls.StickInput{X: 0.5}, want: false},
		{name: "firm left push → true", in: controls.StickInput{X: 1}, want: true},
		{name: "firm diagonal push → true", in: controls.StickInput{X: 0.7, Y: 0.7}, want: true},
		{name: "right stick X deflected → true", in: controls.StickInput{RightX: 0.1}, want: true},
		{name: "right stick Y deflected → true", in: controls.StickInput{RightY: -0.2}, want: true},
		{name: "weak left + deflected right → true", in: controls.StickInput{X: 0.5, RightX: 0.1}, want: true},
	}
	// when
	got := make([]bool, 0, len(tests))
	for _, tc := range tests {
		got = append(got, controls.IsStickMovementSignificant(tc.in))
	}
	// then
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got[i] != tc.want {
				t.Fatalf("IsStickMovementSignificant(%+v) = %v, want %v", tc.in, got[i], tc.want)
			}
		})
	}
}

// T4 — TurnToDir (plrctrls.cpp:1750-1754): no-op when walking, when the
// action cannot change, or when dir is NoDirection; otherwise sets the
// facing and StartStands.
func TestTurnToDir(t *testing.T) {
	for _, tc := range turnToDirCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := newWorld()
			tc.setup(w)
			// when
			controls.TurnToDir(w.View(), tc.dir, w.Hooks())
			// then
			assertFacingAndStands(t, w, tc.wantDir, tc.wantStands)
		})
	}
}

// T5 — WalkInDir (plrctrls.cpp:1725-1747).
func TestWalkInDir(t *testing.T) {
	for _, tc := range walkInDirCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := newWorld()
			tc.setup(w)
			// when
			controls.WalkInDir(w.View(), tc.axis(), w.Hooks())
			// then
			tc.assert(t, w)
		})
	}
}

// T6 — Movement dispatch (plrctrls.cpp:1838-1853): firm push walks, weak
// push turns in place only, no push does nothing.
func TestMovement(t *testing.T) {
	t.Run("firm push walks via WalkInDir", func(t *testing.T) {
		// given — (1,1) is Firm and sectors to UP+RIGHT diagonal.
		w := newWorld()
		// when
		controls.Movement(w.View(), controls.StickInput{X: 1, Y: 1}, w.Hooks())
		// then
		if w.dir != controls.NorthEast {
			t.Fatalf("dir = %v, want NorthEast", w.dir)
		}
		assertWalkOnly(t, w, 10, 9) // future(10,10)+NE{0,-1}
	})
	t.Run("weak push turns in place only", func(t *testing.T) {
		// given — (0.5,0) is Weak and sectors RIGHT.
		w := newWorld()
		// when
		controls.Movement(w.View(), controls.StickInput{X: 0.5}, w.Hooks())
		// then
		assertFacingAndStands(t, w, controls.East, 1)
		if len(w.walkCalls) != 0 || len(w.stopCalls) != 0 {
			t.Fatalf("walk=%v stop=%v, want none (weak push never walks)", w.walkCalls, w.stopCalls)
		}
	})
	t.Run("no push emits nothing", func(t *testing.T) {
		// given
		w := newWorld()
		// when
		controls.Movement(w.View(), controls.StickInput{}, w.Hooks())
		// then
		assertNoMovementOutput(t, w)
	})
}

// newWorld starts a standing player at future (10,10) facing South with an
// idle walkpath and open ground everywhere.
func newWorld() *world {
	return &world{
		mode:       controls.ModeStand,
		dir:        controls.South,
		futureX:    10,
		futureY:    10,
		walkpath0:  controls.WalkNone,
		destAction: controls.DestActionNone,
		posOkFn:    func(int, int) bool { return true },
		notSolidFn: func(int, int) bool { return false },
	}
}

// View adapts the world to controls.PlayerView.
func (w *world) View() controls.PlayerView {
	return playerView{w}
}

// Hooks builds the StickHooks bound to this world.
func (w *world) Hooks() controls.StickHooks {
	return controls.StickHooks{
		IsStandingGround: w.isStandingGround,
		StartStand:       w.startStand,
		PosOkPlayer:      w.posOk,
		IsTileNotSolid:   w.tileNotSolid,
		SendWalkCommand:  w.sendWalkTo,
		SendStopCommand:  w.sendStopAt,
	}
}

func (w *world) isStandingGround() bool {
	return w.standing
}

func (w *world) startStand(d controls.Direction) {
	w.standCalls = append(w.standCalls, d)
}

func (w *world) posOk(x, y int) bool {
	w.posOkCalls++
	return w.posOkFn(x, y)
}

func (w *world) tileNotSolid(x, y int) bool {
	w.notSolCalls++
	return w.notSolidFn(x, y)
}

func (w *world) sendWalkTo(x, y int) {
	w.walkCalls = append(w.walkCalls, [2]int{x, y})
}

func (w *world) sendStopAt(x, y int) {
	w.stopCalls = append(w.stopCalls, [2]int{x, y})
}

func (p playerView) Mode() controls.PlrMode {
	return p.w.mode
}

func (p playerView) AnimFrame() int8 {
	return p.w.current
}

func (p playerView) AttackAnimFrames() int8 {
	return p.w.afNum
}

func (p playerView) SpellAnimFrames() int8 {
	return p.w.sfNum
}

func (p playerView) Facing() controls.Direction {
	return p.w.dir
}

func (p playerView) Future() (x, y int) {
	return p.w.futureX, p.w.futureY
}

func (p playerView) WalkpathHead() int8 {
	return p.w.walkpath0
}

func (p playerView) DestAction() int8 {
	return p.w.destAction
}

func (p playerView) SetFacing(d controls.Direction) {
	p.w.dir = d
}

// noneAxis is the NONE/NONE stick axis (stick released).
func noneAxis() controls.AxisDirection {
	return controls.AxisDirection{}
}

// eastAxis is the RIGHT/NONE stick axis (an East cardinal push).
func eastAxis() controls.AxisDirection {
	return controls.AxisDirection{X: controls.AxisXRight, Y: controls.AxisYNone}
}

// northEastAxis is the RIGHT/UP stick axis (a NorthEast diagonal push).
func northEastAxis() controls.AxisDirection {
	return controls.AxisDirection{X: controls.AxisXRight, Y: controls.AxisYUp}
}

// assertStopOnly pins exactly one stop command at (x,y) and nothing else.
func assertStopOnly(t *testing.T, w *world, x, y int) {
	t.Helper()
	if len(w.stopCalls) != 1 || w.stopCalls[0] != [2]int{x, y} {
		t.Fatalf("stopCommands = %v, want [{%d,%d}]", w.stopCalls, x, y)
	}
	if len(w.walkCalls) != 0 {
		t.Fatalf("walkCommands = %v, want none", w.walkCalls)
	}
}

// assertWalkOnly pins exactly one walk command to (x,y) and nothing else.
func assertWalkOnly(t *testing.T, w *world, x, y int) {
	t.Helper()
	if len(w.walkCalls) != 1 || w.walkCalls[0] != [2]int{x, y} {
		t.Fatalf("walkCommands = %v, want [{%d,%d}]", w.walkCalls, x, y)
	}
	if len(w.stopCalls) != 0 {
		t.Fatalf("stopCommands = %v, want none", w.stopCalls)
	}
}

// assertStoodFacing pins one StartStand call with the given facing and no
// walk/stop commands.
func assertStoodFacing(t *testing.T, w *world, want controls.Direction) {
	t.Helper()
	if len(w.standCalls) != 1 || w.standCalls[0] != want {
		t.Fatalf("standCalls = %v, want [%v]", w.standCalls, want)
	}
	if len(w.walkCalls)+len(w.stopCalls) != 0 {
		t.Fatalf("walk=%v stop=%v, want none", w.walkCalls, w.stopCalls)
	}
}

// assertNoMovementOutput fails the test when any hook fired or the facing
// moved off its initial South.
func assertNoMovementOutput(t *testing.T, w *world) {
	t.Helper()
	if len(w.standCalls)+len(w.walkCalls)+len(w.stopCalls) != 0 {
		t.Fatalf("hooks fired: stand=%v walk=%v stop=%v, want none", w.standCalls, w.walkCalls, w.stopCalls)
	}
	if w.dir != controls.South {
		t.Fatalf("facing changed to %v without any hook output", w.dir)
	}
}

// assertFacingAndStands pins the TurnToDir outcome: the facing and the
// StartStand call count.
func assertFacingAndStands(t *testing.T, w *world, wantDir controls.Direction, wantStands int) {
	t.Helper()
	if w.dir != wantDir {
		t.Fatalf("dir = %v, want %v", w.dir, wantDir)
	}
	if len(w.standCalls) != wantStands {
		t.Fatalf("standCalls = %v, want %d", w.standCalls, wantStands)
	}
}
