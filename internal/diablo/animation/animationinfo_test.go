// Package animation_test ports test/animationinfo_test.cpp (549 lines,
// 8 scenarios) case-for-case against the Go AnimationInfo port.
package animation_test

import (
	"math"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/animation"
)

// step kinds mirror the C++ TestData hierarchy.
type kind int

const (
	set kind = iota
	tick
	render
)

type step struct {
	kind     kind // "set" | "tick" | "render"
	frames   int  // set: numberOfFrames
	delay    int  // set: ticksPerFrame
	flags    animation.DistFlags
	skipped  int     // set: numSkippedFrames
	distBef  int     // set: distributeFramesBeforeFrame
	expFrame int     // tick: expected frame; render: expected frame
	expCnt   int     // tick: expected tickCounterOfCurrentFrame
	progress float64 // render: fraction 0..1
}

func run(t *testing.T, name string, steps []step) {
	t.Helper()
	a := &animation.Info{}
	g := &animation.ProgressSink{}
	for i, s := range steps {
		switch s.kind {
		case set:
			a.SetNewAnimation(s.frames, s.delay, s.flags, s.skipped, s.distBef, 0, g)
		case tick:
			a.ProcessAnimation()
			if int(a.CurrentFrame) != s.expFrame || int(a.TickCounterOfCurrentFrame) != s.expCnt {
				t.Errorf("%s step %d (tick): got frame=%d cnt=%d, want frame=%d cnt=%d",
					name, i, a.CurrentFrame, a.TickCounterOfCurrentFrame, s.expFrame, s.expCnt)
			}
		case render:
			g.ProgressToNextGameTick = uint8(math.Round(s.progress * float64(animation.BaseValueFraction)))
			if got := a.FrameToUseForRendering(g); got != s.expFrame {
				t.Errorf("%s step %d (render p=%.2f): got frame=%d, want %d (currentFrame=%d cnt=%d)",
					name, i, s.progress, got, s.expFrame, a.CurrentFrame, a.TickCounterOfCurrentFrame)
			}
		default:
			t.Fatalf("%s step %d: unknown step kind %v", name, i, s.kind)
		}
	}
}

func F(frames, delay int, flags animation.DistFlags, skipped, distBef int) step {
	return step{kind: set, frames: frames, delay: delay, flags: flags, skipped: skipped, distBef: distBef}
}
func T(frame, cnt int) step       { return step{kind: tick, expFrame: frame, expCnt: cnt} }
func R(p float64, frame int) step { return step{kind: render, progress: p, expFrame: frame} }

// AnimationInfo.AttackSwordWarrior
func TestAttackSwordWarrior(t *testing.T) {
	run(t, "AttackSwordWarrior", []step{
		F(16, 1, animation.FlagProcessAnimationPending, 0, 9),
		T(1, 0),
		R(0.0, 0), R(0.3, 0), R(0.6, 0), R(0.8, 0),
		T(2, 0),
		R(0.0, 1), R(0.3, 1), R(0.6, 1), R(0.8, 2),
		T(3, 0),
		R(0.0, 2), R(0.3, 2), R(0.6, 2), R(0.8, 3),
		T(4, 0),
		R(0.0, 3), R(0.3, 3), R(0.6, 4), R(0.8, 4),
		T(5, 0),
		R(0.0, 4), R(0.3, 4), R(0.6, 5), R(0.8, 5),
		T(6, 0),
		R(0.0, 5), R(0.3, 6), R(0.6, 6), R(0.8, 6),
		T(7, 0),
		R(0.0, 6), R(0.3, 7), R(0.6, 7), R(0.8, 7),
		// After this GameTick, the Animation Distribution Logic is disabled
		T(8, 0), R(0.1, 8),
		T(9, 0), R(0.4, 9),
		T(10, 0), R(0.4, 10),
		T(11, 0), R(0.3, 11),
		T(12, 0), R(0.0, 12),
		T(13, 0), R(0.6, 13),
		T(14, 0), R(0.6, 14),
		T(15, 0), R(0.6, 15),
	})
}

// AnimationInfo.AttackSwordWarriorWithFastestAttack
func TestAttackSwordWarriorWithFastestAttack(t *testing.T) {
	run(t, "FastestAttack", []step{
		F(16, 1, animation.FlagProcessAnimationPending, 2, 9),
		T(3, 0),
		R(0.0, 0), R(0.3, 0), R(0.6, 0), R(0.8, 1),
		T(4, 0),
		R(0.0, 1), R(0.3, 2), R(0.6, 2), R(0.8, 2),
		T(5, 0),
		R(0.0, 3), R(0.3, 3), R(0.6, 4), R(0.8, 4),
		T(6, 0),
		R(0.0, 4), R(0.3, 5), R(0.6, 5), R(0.8, 6),
		T(7, 0),
		R(0.0, 6), R(0.3, 6), R(0.6, 7), R(0.8, 7),
		T(8, 0), R(0.1, 8),
		T(9, 0), R(0.4, 9),
		T(10, 0), R(0.4, 10),
		T(11, 0), R(0.3, 11),
		T(12, 0), R(0.0, 12),
		T(13, 0), R(0.6, 13),
		T(14, 0), R(0.6, 14),
		T(15, 0), R(0.6, 15),
	})
}

// AnimationInfo.AttackSwordWarriorRepeated
func TestAttackSwordWarriorRepeated(t *testing.T) {
	run(t, "Repeated", []step{
		F(16, 1, animation.FlagProcessAnimationPending, 0, 9),
		T(1, 0),
		R(0.0, 0), R(0.3, 0), R(0.6, 0), R(0.8, 0),
		T(2, 0),
		R(0.0, 1), R(0.3, 1), R(0.6, 1), R(0.8, 2),
		T(3, 0),
		R(0.0, 2), R(0.3, 2), R(0.6, 2), R(0.8, 3),
		T(4, 0),
		R(0.0, 3), R(0.3, 3), R(0.6, 4), R(0.8, 4),
		T(5, 0),
		R(0.0, 4), R(0.3, 4), R(0.6, 5), R(0.8, 5),
		T(6, 0),
		R(0.0, 5), R(0.3, 6), R(0.6, 6), R(0.8, 6),
		T(7, 0),
		R(0.0, 6), R(0.3, 7), R(0.6, 7), R(0.8, 7),
		T(8, 0), R(0.1, 8),
		T(9, 0), R(0.3, 9),
		// Start of repeated attack
		F(16, 1, animation.FlagProcessAnimationPending|animation.FlagRepeatedAction, 0, 9),
		T(1, 0),
		R(0.0, 10), R(0.3, 10), R(0.6, 11), R(0.8, 11),
		T(2, 0),
		R(0.0, 12), R(0.3, 12), R(0.6, 13), R(0.8, 13),
		T(3, 0),
		R(0.0, 14), R(0.3, 14), R(0.6, 15), R(0.8, 15),
		T(4, 0),
		R(0.0, 0), R(0.3, 0), R(0.6, 1), R(0.8, 1),
		T(5, 0),
		R(0.0, 2), R(0.3, 2), R(0.6, 3), R(0.8, 3),
		T(6, 0),
		R(0.0, 4), R(0.3, 4), R(0.6, 5), R(0.8, 5),
		T(7, 0),
		R(0.0, 6), R(0.3, 6), R(0.6, 7), R(0.8, 7),
		T(8, 0), R(0.1, 8),
		T(9, 0), R(0.4, 9),
		T(10, 0), R(0.4, 10),
		T(11, 0), R(0.3, 11),
		T(12, 0), R(0.0, 12),
		T(13, 0), R(0.6, 13),
		T(14, 0), R(0.6, 14),
		T(15, 0), R(0.6, 15),
	})
}

// AnimationInfo.BlockingWarriorNormal
func TestBlockingWarriorNormal(t *testing.T) {
	run(t, "BlockWarrior", []step{
		F(2, 3, animation.FlagSkipsDelayOfLastFrame, 0, 0),
		R(0.0, 0), R(0.3, 0), R(0.6, 0), R(0.8, 0),
		T(0, 1),
		R(0.0, 0), R(0.3, 0), R(0.6, 0), R(0.8, 0),
		T(0, 2),
		R(0.0, 1), R(0.3, 1), R(0.6, 1), R(0.8, 1),
		T(1, 0),
		R(0.0, 1), R(0.3, 1), R(0.6, 1), R(0.8, 1),
	})
}

// AnimationInfo.BlockingSorcererWithFastBlock
func TestBlockingSorcererWithFastBlock(t *testing.T) {
	run(t, "BlockSorcerer", []step{
		F(6, 3, animation.FlagSkipsDelayOfLastFrame, 4, 0),
		R(0.0, 0), R(0.3, 0), R(0.6, 0), R(0.8, 1),
		T(4, 1),
		R(0.0, 1), R(0.3, 1), R(0.6, 2), R(0.8, 2),
		T(4, 2),
		R(0.0, 3), R(0.3, 3), R(0.6, 3), R(0.8, 4),
		T(5, 0),
		R(0.0, 4), R(0.3, 4), R(0.6, 5), R(0.8, 5),
	})
}

// AnimationInfo.HitRecoverySorcererHarmony
func TestHitRecoverySorcererHarmony(t *testing.T) {
	run(t, "HitRecovery", []step{
		F(8, 1, animation.FlagNone, 3, 0),
		R(0.0, 0), R(0.3, 0), R(0.6, 0), R(0.8, 1),
		T(4, 0),
		R(0.0, 1), R(0.3, 2), R(0.6, 2), R(0.8, 2),
		T(5, 0),
		R(0.0, 3), R(0.3, 3), R(0.6, 4), R(0.8, 4),
		T(6, 0),
		R(0.0, 4), R(0.3, 5), R(0.6, 5), R(0.8, 6),
		T(7, 0),
		R(0.0, 6), R(0.3, 6), R(0.6, 7), R(0.8, 7),
	})
}

// AnimationInfo.Stand
func TestStand(t *testing.T) {
	run(t, "Stand", []step{
		F(10, 4, animation.FlagNone, 0, 0),
		R(0.1, 0),
		T(0, 1), R(0.6, 0),
		T(0, 2), R(0.6, 0),
		T(0, 3), R(0.6, 0),
		T(1, 0), R(0.6, 1),
		T(1, 1), R(0.6, 1),
		T(1, 2), R(0.6, 1),
		T(1, 3), R(0.6, 1),
		T(2, 0), R(0.6, 2),
		T(2, 1), R(0.6, 2),
		T(2, 2), R(0.6, 2),
		T(2, 3), R(0.6, 2),
		T(3, 0), R(0.6, 3),
		T(3, 1), R(0.6, 3),
		T(3, 2), R(0.6, 3),
		T(3, 3), R(0.6, 3),
		T(4, 0), R(0.6, 4),
		T(4, 1), R(0.6, 4),
		T(4, 2), R(0.6, 4),
		T(4, 3), R(0.6, 4),
		T(5, 0), R(0.6, 5),
		T(5, 1), R(0.6, 5),
		T(5, 2), R(0.6, 5),
		T(5, 3), R(0.6, 5),
		T(6, 0), R(0.6, 6),
		T(6, 1), R(0.6, 6),
		T(6, 2), R(0.6, 6),
		T(6, 3), R(0.6, 6),
		T(7, 0), R(0.6, 7),
		T(7, 1), R(0.6, 7),
		T(7, 2), R(0.6, 7),
		T(7, 3), R(0.6, 7),
		T(8, 0), R(0.6, 8),
		T(8, 1), R(0.6, 8),
		T(8, 2), R(0.6, 8),
		T(8, 3), R(0.6, 8),
		T(9, 0), R(0.6, 9),
		T(9, 1), R(0.6, 9),
		T(9, 2), R(0.6, 9),
		T(9, 3), R(0.6, 9),
		// Animation starts again
		T(0, 0), R(0.1, 0),
		T(0, 1), R(0.6, 0),
		T(0, 2), R(0.6, 0),
		T(0, 3), R(0.6, 0),
	})
}
