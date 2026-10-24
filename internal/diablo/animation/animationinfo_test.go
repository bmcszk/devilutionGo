// TODO remove nolint:testpackage // internal-state test: pokes unexported identifiers
package animation

import (
	"math"
	"testing"
)

// renderCheck pins one interpolated render probe: the frame shown at the
// given fraction of progress to the next game tick.
type renderCheck struct {
	progress  float64
	wantFrame int
}

// want is the post-tick expectation for one table row: frame and tick counter.
type want struct {
	frame int
	cnt   int
}

func stepAnimationCase(t *testing.T, a *Info, g *progressSink, frames, ticksPerFrame int, set *newAnimationOptions, w want, renders []renderCheck) {
	t.Helper()
	// when
	if set != nil {
		a.SetNewAnimation(frames, ticksPerFrame, *set)
	} else {
		a.ProcessAnimation()
	}
	// then
	if int(a.CurrentFrame) != w.frame || int(a.TickCounterOfCurrentFrame) != w.cnt {
		t.Errorf("got frame=%d cnt=%d, want frame=%d cnt=%d",
			a.CurrentFrame, a.TickCounterOfCurrentFrame, w.frame, w.cnt)
	}
	for _, rc := range renders {
		g.ProgressToNextGameTick = uint8(math.Round(rc.progress * float64(BaseValueFraction)))
		if got := a.FrameToUseForRendering(g); got != rc.wantFrame {
			t.Errorf("render p=%.2f: got frame=%d, want %d (currentFrame=%d cnt=%d)",
				rc.progress, got, rc.wantFrame, a.CurrentFrame, a.TickCounterOfCurrentFrame)
		}
	}
}

func TestAttackSwordWarrior(t *testing.T) {
	const (
		frames        = 16
		ticksPerFrame = 1
	)
	// given
	a := &Info{}
	g := &progressSink{}
	for _, tc := range []struct {
		name    string
		set     *newAnimationOptions // nil: ProcessAnimation this tick
		want    want
		renders []renderCheck
	}{
		{
			name: "set frame=16 delay=1 skipped=0 distBefore=9",
			set: &newAnimationOptions{
				Flags:            FlagProcessAnimationPending,
				NumSkippedFrames: 0, DistributeFramesBeforeFrame: 9,
			},
			want: want{frame: 0, cnt: 0},
		},
		{
			name:    "tick 1 (frame=1, cnt=0)",
			want:    want{frame: 1, cnt: 0},
			renders: []renderCheck{{0.0, 0}, {0.3, 0}, {0.6, 0}, {0.8, 0}},
		},
		{
			name:    "tick 2 (frame=2, cnt=0)",
			want:    want{frame: 2, cnt: 0},
			renders: []renderCheck{{0.0, 1}, {0.3, 1}, {0.6, 1}, {0.8, 2}},
		},
		{
			name:    "tick 3 (frame=3, cnt=0)",
			want:    want{frame: 3, cnt: 0},
			renders: []renderCheck{{0.0, 2}, {0.3, 2}, {0.6, 2}, {0.8, 3}},
		},
		{
			name:    "tick 4 (frame=4, cnt=0)",
			want:    want{frame: 4, cnt: 0},
			renders: []renderCheck{{0.0, 3}, {0.3, 3}, {0.6, 4}, {0.8, 4}},
		},
		{
			name:    "tick 5 (frame=5, cnt=0)",
			want:    want{frame: 5, cnt: 0},
			renders: []renderCheck{{0.0, 4}, {0.3, 4}, {0.6, 5}, {0.8, 5}},
		},
		{
			name:    "tick 6 (frame=6, cnt=0)",
			want:    want{frame: 6, cnt: 0},
			renders: []renderCheck{{0.0, 5}, {0.3, 6}, {0.6, 6}, {0.8, 6}},
		},
		{
			name:    "tick 7 (frame=7, cnt=0)",
			want:    want{frame: 7, cnt: 0},
			renders: []renderCheck{{0.0, 6}, {0.3, 7}, {0.6, 7}, {0.8, 7}},
		},
		// After this GameTick, the Animation Distribution Logic is disabled
		{
			name:    "tick 8 (frame=8, cnt=0)",
			want:    want{frame: 8, cnt: 0},
			renders: []renderCheck{{progress: 0.1, wantFrame: 8}},
		},
		{
			name:    "tick 9 (frame=9, cnt=0)",
			want:    want{frame: 9, cnt: 0},
			renders: []renderCheck{{progress: 0.4, wantFrame: 9}},
		},
		{
			name:    "tick 10 (frame=10, cnt=0)",
			want:    want{frame: 10, cnt: 0},
			renders: []renderCheck{{progress: 0.4, wantFrame: 10}},
		},
		{
			name:    "tick 11 (frame=11, cnt=0)",
			want:    want{frame: 11, cnt: 0},
			renders: []renderCheck{{progress: 0.3, wantFrame: 11}},
		},
		{
			name:    "tick 12 (frame=12, cnt=0)",
			want:    want{frame: 12, cnt: 0},
			renders: []renderCheck{{progress: 0.0, wantFrame: 12}},
		},
		{
			name:    "tick 13 (frame=13, cnt=0)",
			want:    want{frame: 13, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 13}},
		},
		{
			name:    "tick 14 (frame=14, cnt=0)",
			want:    want{frame: 14, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 14}},
		},
		{
			name:    "tick 15 (frame=15, cnt=0)",
			want:    want{frame: 15, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 15}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stepAnimationCase(t, a, g, frames, ticksPerFrame, tc.set, tc.want, tc.renders)
		})
	}
}

func TestAttackSwordWarriorWithFastestAttack(t *testing.T) {
	const (
		frames        = 16
		ticksPerFrame = 1
	)
	// given
	a := &Info{}
	g := &progressSink{}
	for _, tc := range []struct {
		name    string
		set     *newAnimationOptions // nil: ProcessAnimation this tick
		want    want
		renders []renderCheck
	}{
		{
			name: "set frame=16 delay=1 skipped=2 distBefore=9",
			set: &newAnimationOptions{
				Flags:            FlagProcessAnimationPending,
				NumSkippedFrames: 2, DistributeFramesBeforeFrame: 9,
			},
			want: want{frame: 2, cnt: 0},
		},
		{
			name:    "tick 1 (frame=3, cnt=0)",
			want:    want{frame: 3, cnt: 0},
			renders: []renderCheck{{0.0, 0}, {0.3, 0}, {0.6, 0}, {0.8, 1}},
		},
		{
			name:    "tick 2 (frame=4, cnt=0)",
			want:    want{frame: 4, cnt: 0},
			renders: []renderCheck{{0.0, 1}, {0.3, 2}, {0.6, 2}, {0.8, 2}},
		},
		{
			name:    "tick 3 (frame=5, cnt=0)",
			want:    want{frame: 5, cnt: 0},
			renders: []renderCheck{{0.0, 3}, {0.3, 3}, {0.6, 4}, {0.8, 4}},
		},
		{
			name:    "tick 4 (frame=6, cnt=0)",
			want:    want{frame: 6, cnt: 0},
			renders: []renderCheck{{0.0, 4}, {0.3, 5}, {0.6, 5}, {0.8, 6}},
		},
		{
			name:    "tick 5 (frame=7, cnt=0)",
			want:    want{frame: 7, cnt: 0},
			renders: []renderCheck{{0.0, 6}, {0.3, 6}, {0.6, 7}, {0.8, 7}},
		},
		{
			name:    "tick 6 (frame=8, cnt=0)",
			want:    want{frame: 8, cnt: 0},
			renders: []renderCheck{{progress: 0.1, wantFrame: 8}},
		},
		{
			name:    "tick 7 (frame=9, cnt=0)",
			want:    want{frame: 9, cnt: 0},
			renders: []renderCheck{{progress: 0.4, wantFrame: 9}},
		},
		{
			name:    "tick 8 (frame=10, cnt=0)",
			want:    want{frame: 10, cnt: 0},
			renders: []renderCheck{{progress: 0.4, wantFrame: 10}},
		},
		{
			name:    "tick 9 (frame=11, cnt=0)",
			want:    want{frame: 11, cnt: 0},
			renders: []renderCheck{{progress: 0.3, wantFrame: 11}},
		},
		{
			name:    "tick 10 (frame=12, cnt=0)",
			want:    want{frame: 12, cnt: 0},
			renders: []renderCheck{{progress: 0.0, wantFrame: 12}},
		},
		{
			name:    "tick 11 (frame=13, cnt=0)",
			want:    want{frame: 13, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 13}},
		},
		{
			name:    "tick 12 (frame=14, cnt=0)",
			want:    want{frame: 14, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 14}},
		},
		{
			name:    "tick 13 (frame=15, cnt=0)",
			want:    want{frame: 15, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 15}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stepAnimationCase(t, a, g, frames, ticksPerFrame, tc.set, tc.want, tc.renders)
		})
	}
}

func TestAttackSwordWarriorRepeated(t *testing.T) {
	const (
		frames        = 16
		ticksPerFrame = 1
	)
	// given
	a := &Info{}
	g := &progressSink{}
	for _, tc := range []struct {
		name    string
		set     *newAnimationOptions // nil: ProcessAnimation this tick
		want    want
		renders []renderCheck
	}{
		{
			name: "set frame=16 delay=1 skipped=0 distBefore=9",
			set: &newAnimationOptions{
				Flags:            FlagProcessAnimationPending,
				NumSkippedFrames: 0, DistributeFramesBeforeFrame: 9,
			},
			want: want{frame: 0, cnt: 0},
		},
		{
			name:    "tick 1 (frame=1, cnt=0)",
			want:    want{frame: 1, cnt: 0},
			renders: []renderCheck{{0.0, 0}, {0.3, 0}, {0.6, 0}, {0.8, 0}},
		},
		{
			name:    "tick 2 (frame=2, cnt=0)",
			want:    want{frame: 2, cnt: 0},
			renders: []renderCheck{{0.0, 1}, {0.3, 1}, {0.6, 1}, {0.8, 2}},
		},
		{
			name:    "tick 3 (frame=3, cnt=0)",
			want:    want{frame: 3, cnt: 0},
			renders: []renderCheck{{0.0, 2}, {0.3, 2}, {0.6, 2}, {0.8, 3}},
		},
		{
			name:    "tick 4 (frame=4, cnt=0)",
			want:    want{frame: 4, cnt: 0},
			renders: []renderCheck{{0.0, 3}, {0.3, 3}, {0.6, 4}, {0.8, 4}},
		},
		{
			name:    "tick 5 (frame=5, cnt=0)",
			want:    want{frame: 5, cnt: 0},
			renders: []renderCheck{{0.0, 4}, {0.3, 4}, {0.6, 5}, {0.8, 5}},
		},
		{
			name:    "tick 6 (frame=6, cnt=0)",
			want:    want{frame: 6, cnt: 0},
			renders: []renderCheck{{0.0, 5}, {0.3, 6}, {0.6, 6}, {0.8, 6}},
		},
		{
			name:    "tick 7 (frame=7, cnt=0)",
			want:    want{frame: 7, cnt: 0},
			renders: []renderCheck{{0.0, 6}, {0.3, 7}, {0.6, 7}, {0.8, 7}},
		},
		{
			name:    "tick 8 (frame=8, cnt=0)",
			want:    want{frame: 8, cnt: 0},
			renders: []renderCheck{{progress: 0.1, wantFrame: 8}},
		},
		{
			name:    "tick 9 (frame=9, cnt=0)",
			want:    want{frame: 9, cnt: 0},
			renders: []renderCheck{{progress: 0.3, wantFrame: 9}},
		},
		{
			name: "set2 frame=16 delay=1 skipped=0 distBefore=9",
			set: &newAnimationOptions{
				Flags:            FlagProcessAnimationPending | FlagRepeatedAction,
				NumSkippedFrames: 0, DistributeFramesBeforeFrame: 9,
			},
			want: want{frame: 0, cnt: 0},
		},
		{
			name:    "tick 10 (frame=1, cnt=0)",
			want:    want{frame: 1, cnt: 0},
			renders: []renderCheck{{0.0, 10}, {0.3, 10}, {0.6, 11}, {0.8, 11}},
		},
		{
			name:    "tick 11 (frame=2, cnt=0)",
			want:    want{frame: 2, cnt: 0},
			renders: []renderCheck{{0.0, 12}, {0.3, 12}, {0.6, 13}, {0.8, 13}},
		},
		{
			name:    "tick 12 (frame=3, cnt=0)",
			want:    want{frame: 3, cnt: 0},
			renders: []renderCheck{{0.0, 14}, {0.3, 14}, {0.6, 15}, {0.8, 15}},
		},
		{
			name:    "tick 13 (frame=4, cnt=0)",
			want:    want{frame: 4, cnt: 0},
			renders: []renderCheck{{0.0, 0}, {0.3, 0}, {0.6, 1}, {0.8, 1}},
		},
		{
			name:    "tick 14 (frame=5, cnt=0)",
			want:    want{frame: 5, cnt: 0},
			renders: []renderCheck{{0.0, 2}, {0.3, 2}, {0.6, 3}, {0.8, 3}},
		},
		{
			name:    "tick 15 (frame=6, cnt=0)",
			want:    want{frame: 6, cnt: 0},
			renders: []renderCheck{{0.0, 4}, {0.3, 4}, {0.6, 5}, {0.8, 5}},
		},
		{
			name:    "tick 16 (frame=7, cnt=0)",
			want:    want{frame: 7, cnt: 0},
			renders: []renderCheck{{0.0, 6}, {0.3, 6}, {0.6, 7}, {0.8, 7}},
		},
		{
			name:    "tick 17 (frame=8, cnt=0)",
			want:    want{frame: 8, cnt: 0},
			renders: []renderCheck{{progress: 0.1, wantFrame: 8}},
		},
		{
			name:    "tick 18 (frame=9, cnt=0)",
			want:    want{frame: 9, cnt: 0},
			renders: []renderCheck{{progress: 0.4, wantFrame: 9}},
		},
		{
			name:    "tick 19 (frame=10, cnt=0)",
			want:    want{frame: 10, cnt: 0},
			renders: []renderCheck{{progress: 0.4, wantFrame: 10}},
		},
		{
			name:    "tick 20 (frame=11, cnt=0)",
			want:    want{frame: 11, cnt: 0},
			renders: []renderCheck{{progress: 0.3, wantFrame: 11}},
		},
		{
			name:    "tick 21 (frame=12, cnt=0)",
			want:    want{frame: 12, cnt: 0},
			renders: []renderCheck{{progress: 0.0, wantFrame: 12}},
		},
		{
			name:    "tick 22 (frame=13, cnt=0)",
			want:    want{frame: 13, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 13}},
		},
		{
			name:    "tick 23 (frame=14, cnt=0)",
			want:    want{frame: 14, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 14}},
		},
		{
			name:    "tick 24 (frame=15, cnt=0)",
			want:    want{frame: 15, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 15}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stepAnimationCase(t, a, g, frames, ticksPerFrame, tc.set, tc.want, tc.renders)
		})
	}
}

func TestBlocking(t *testing.T) {
	// given
	a := &Info{}
	g := &progressSink{}
	for _, tc := range []struct {
		name          string
		frames        int
		ticksPerFrame int
		set           *newAnimationOptions // nil: ProcessAnimation this tick
		want          want
		renders       []renderCheck
	}{
		{
			name:          "set frame=2 delay=3 skipped=0 distBefore=0",
			frames:        2,
			ticksPerFrame: 3,
			set: &newAnimationOptions{
				Flags:            FlagSkipsDelayOfLastFrame,
				NumSkippedFrames: 0, DistributeFramesBeforeFrame: 0,
			},
			want:    want{frame: 0, cnt: 0},
			renders: []renderCheck{{0.0, 0}, {0.3, 0}, {0.6, 0}, {0.8, 0}},
		},
		{
			name:    "tick 1 (frame=0, cnt=1)",
			want:    want{frame: 0, cnt: 1},
			renders: []renderCheck{{0.0, 0}, {0.3, 0}, {0.6, 0}, {0.8, 0}},
		},
		{
			name:    "tick 2 (frame=0, cnt=2)",
			want:    want{frame: 0, cnt: 2},
			renders: []renderCheck{{0.0, 1}, {0.3, 1}, {0.6, 1}, {0.8, 1}},
		},
		{
			name:    "tick 3 (frame=1, cnt=0)",
			want:    want{frame: 1, cnt: 0},
			renders: []renderCheck{{0.0, 1}, {0.3, 1}, {0.6, 1}, {0.8, 1}},
		},
		{
			name:          "set frame=6 delay=3 skipped=4 distBefore=0",
			frames:        6,
			ticksPerFrame: 3,
			set: &newAnimationOptions{
				Flags:            FlagSkipsDelayOfLastFrame,
				NumSkippedFrames: 4, DistributeFramesBeforeFrame: 0,
			},
			want:    want{frame: 4, cnt: 0},
			renders: []renderCheck{{0.0, 0}, {0.3, 0}, {0.6, 0}, {0.8, 1}},
		},
		{
			name:    "tick 1 (frame=4, cnt=1)",
			want:    want{frame: 4, cnt: 1},
			renders: []renderCheck{{0.0, 1}, {0.3, 1}, {0.6, 2}, {0.8, 2}},
		},
		{
			name:    "tick 2 (frame=4, cnt=2)",
			want:    want{frame: 4, cnt: 2},
			renders: []renderCheck{{0.0, 3}, {0.3, 3}, {0.6, 3}, {0.8, 4}},
		},
		{
			name:    "tick 3 (frame=5, cnt=0)",
			want:    want{frame: 5, cnt: 0},
			renders: []renderCheck{{0.0, 4}, {0.3, 4}, {0.6, 5}, {0.8, 5}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stepAnimationCase(t, a, g, tc.frames, tc.ticksPerFrame, tc.set, tc.want, tc.renders)
		})
	}
}

func TestHitRecoverySorcererHarmony(t *testing.T) {
	const (
		frames        = 8
		ticksPerFrame = 1
	)
	// given
	a := &Info{}
	g := &progressSink{}
	for _, tc := range []struct {
		name    string
		set     *newAnimationOptions // nil: ProcessAnimation this tick
		want    want
		renders []renderCheck
	}{
		{
			name: "set frame=8 delay=1 skipped=3 distBefore=0",
			set: &newAnimationOptions{
				Flags:            FlagNone,
				NumSkippedFrames: 3, DistributeFramesBeforeFrame: 0,
			},
			want:    want{frame: 3, cnt: 0},
			renders: []renderCheck{{0.0, 0}, {0.3, 0}, {0.6, 0}, {0.8, 1}},
		},
		{
			name:    "tick 1 (frame=4, cnt=0)",
			want:    want{frame: 4, cnt: 0},
			renders: []renderCheck{{0.0, 1}, {0.3, 2}, {0.6, 2}, {0.8, 2}},
		},
		{
			name:    "tick 2 (frame=5, cnt=0)",
			want:    want{frame: 5, cnt: 0},
			renders: []renderCheck{{0.0, 3}, {0.3, 3}, {0.6, 4}, {0.8, 4}},
		},
		{
			name:    "tick 3 (frame=6, cnt=0)",
			want:    want{frame: 6, cnt: 0},
			renders: []renderCheck{{0.0, 4}, {0.3, 5}, {0.6, 5}, {0.8, 6}},
		},
		{
			name:    "tick 4 (frame=7, cnt=0)",
			want:    want{frame: 7, cnt: 0},
			renders: []renderCheck{{0.0, 6}, {0.3, 6}, {0.6, 7}, {0.8, 7}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stepAnimationCase(t, a, g, frames, ticksPerFrame, tc.set, tc.want, tc.renders)
		})
	}
}

func TestStand(t *testing.T) {
	const (
		frames        = 10
		ticksPerFrame = 4
	)
	// given
	a := &Info{}
	g := &progressSink{}
	for _, tc := range []struct {
		name    string
		set     *newAnimationOptions // nil: ProcessAnimation this tick
		want    want
		renders []renderCheck
	}{
		{
			name: "set frame=10 delay=4 skipped=0 distBefore=0",
			set: &newAnimationOptions{
				Flags:            FlagNone,
				NumSkippedFrames: 0, DistributeFramesBeforeFrame: 0,
			},
			want:    want{frame: 0, cnt: 0},
			renders: []renderCheck{{progress: 0.1, wantFrame: 0}},
		},
		{
			name:    "tick 1 (frame=0, cnt=1)",
			want:    want{frame: 0, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 0}},
		},
		{
			name:    "tick 2 (frame=0, cnt=2)",
			want:    want{frame: 0, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 0}},
		},
		{
			name:    "tick 3 (frame=0, cnt=3)",
			want:    want{frame: 0, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 0}},
		},
		{
			name:    "tick 4 (frame=1, cnt=0)",
			want:    want{frame: 1, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 1}},
		},
		{
			name:    "tick 5 (frame=1, cnt=1)",
			want:    want{frame: 1, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 1}},
		},
		{
			name:    "tick 6 (frame=1, cnt=2)",
			want:    want{frame: 1, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 1}},
		},
		{
			name:    "tick 7 (frame=1, cnt=3)",
			want:    want{frame: 1, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 1}},
		},
		{
			name:    "tick 8 (frame=2, cnt=0)",
			want:    want{frame: 2, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 2}},
		},
		{
			name:    "tick 9 (frame=2, cnt=1)",
			want:    want{frame: 2, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 2}},
		},
		{
			name:    "tick 10 (frame=2, cnt=2)",
			want:    want{frame: 2, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 2}},
		},
		{
			name:    "tick 11 (frame=2, cnt=3)",
			want:    want{frame: 2, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 2}},
		},
		{
			name:    "tick 12 (frame=3, cnt=0)",
			want:    want{frame: 3, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 3}},
		},
		{
			name:    "tick 13 (frame=3, cnt=1)",
			want:    want{frame: 3, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 3}},
		},
		{
			name:    "tick 14 (frame=3, cnt=2)",
			want:    want{frame: 3, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 3}},
		},
		{
			name:    "tick 15 (frame=3, cnt=3)",
			want:    want{frame: 3, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 3}},
		},
		{
			name:    "tick 16 (frame=4, cnt=0)",
			want:    want{frame: 4, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 4}},
		},
		{
			name:    "tick 17 (frame=4, cnt=1)",
			want:    want{frame: 4, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 4}},
		},
		{
			name:    "tick 18 (frame=4, cnt=2)",
			want:    want{frame: 4, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 4}},
		},
		{
			name:    "tick 19 (frame=4, cnt=3)",
			want:    want{frame: 4, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 4}},
		},
		{
			name:    "tick 20 (frame=5, cnt=0)",
			want:    want{frame: 5, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 5}},
		},
		{
			name:    "tick 21 (frame=5, cnt=1)",
			want:    want{frame: 5, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 5}},
		},
		{
			name:    "tick 22 (frame=5, cnt=2)",
			want:    want{frame: 5, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 5}},
		},
		{
			name:    "tick 23 (frame=5, cnt=3)",
			want:    want{frame: 5, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 5}},
		},
		{
			name:    "tick 24 (frame=6, cnt=0)",
			want:    want{frame: 6, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 6}},
		},
		{
			name:    "tick 25 (frame=6, cnt=1)",
			want:    want{frame: 6, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 6}},
		},
		{
			name:    "tick 26 (frame=6, cnt=2)",
			want:    want{frame: 6, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 6}},
		},
		{
			name:    "tick 27 (frame=6, cnt=3)",
			want:    want{frame: 6, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 6}},
		},
		{
			name:    "tick 28 (frame=7, cnt=0)",
			want:    want{frame: 7, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 7}},
		},
		{
			name:    "tick 29 (frame=7, cnt=1)",
			want:    want{frame: 7, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 7}},
		},
		{
			name:    "tick 30 (frame=7, cnt=2)",
			want:    want{frame: 7, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 7}},
		},
		{
			name:    "tick 31 (frame=7, cnt=3)",
			want:    want{frame: 7, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 7}},
		},
		{
			name:    "tick 32 (frame=8, cnt=0)",
			want:    want{frame: 8, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 8}},
		},
		{
			name:    "tick 33 (frame=8, cnt=1)",
			want:    want{frame: 8, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 8}},
		},
		{
			name:    "tick 34 (frame=8, cnt=2)",
			want:    want{frame: 8, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 8}},
		},
		{
			name:    "tick 35 (frame=8, cnt=3)",
			want:    want{frame: 8, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 8}},
		},
		{
			name:    "tick 36 (frame=9, cnt=0)",
			want:    want{frame: 9, cnt: 0},
			renders: []renderCheck{{progress: 0.6, wantFrame: 9}},
		},
		{
			name:    "tick 37 (frame=9, cnt=1)",
			want:    want{frame: 9, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 9}},
		},
		{
			name:    "tick 38 (frame=9, cnt=2)",
			want:    want{frame: 9, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 9}},
		},
		{
			name:    "tick 39 (frame=9, cnt=3)",
			want:    want{frame: 9, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 9}},
		},
		{
			name:    "tick 40 (frame=0, cnt=0)",
			want:    want{frame: 0, cnt: 0},
			renders: []renderCheck{{progress: 0.1, wantFrame: 0}},
		},
		{
			name:    "tick 41 (frame=0, cnt=1)",
			want:    want{frame: 0, cnt: 1},
			renders: []renderCheck{{progress: 0.6, wantFrame: 0}},
		},
		{
			name:    "tick 42 (frame=0, cnt=2)",
			want:    want{frame: 0, cnt: 2},
			renders: []renderCheck{{progress: 0.6, wantFrame: 0}},
		},
		{
			name:    "tick 43 (frame=0, cnt=3)",
			want:    want{frame: 0, cnt: 3},
			renders: []renderCheck{{progress: 0.6, wantFrame: 0}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stepAnimationCase(t, a, g, frames, ticksPerFrame, tc.set, tc.want, tc.renders)
		})
	}
}

func TestPetrifiedDoesNotAdvance(t *testing.T) {
	for _, tc := range []struct {
		name      string
		petrified bool
		wantFrame int
		wantProg  uint8
	}{
		{"live", false, 1, 42},
		{"petrified", true, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expectPetrifiedRender(t, tc.petrified, tc.wantFrame, tc.wantProg)
		})
	}
}

func expectPetrifiedRender(t *testing.T, petrified bool, wantFrame int, wantProg uint8) {
	t.Helper()
	// given
	a := &Info{}
	g := &progressSink{}
	a.SetNewAnimation(4, 1, newAnimationOptions{NumSkippedFrames: 1})
	g.ProgressToNextGameTick = BaseValueFraction
	a.IsPetrified = petrified
	// when
	gotFrame := a.FrameToUseForRendering(g)
	gotProg := a.AnimationProgress(g)
	// then
	if gotFrame != wantFrame {
		t.Errorf("frame: got %d, want %d", gotFrame, wantFrame)
	}
	if gotProg != wantProg {
		t.Errorf("progress: got %d, want %d", gotProg, wantProg)
	}
}
