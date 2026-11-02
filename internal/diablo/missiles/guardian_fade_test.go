// Package missiles_test pins the Guardian fade tail (missiles.cpp:3558-3570).
package missiles_test

import (
	"slices"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// guardianFadeLights records the light calls of one fade tick.
type guardianFadeLights struct {
	radii   []uint8
	unlight int
}

// TestProcessGuardianFadeTailFrameFourteenFlips pins the frame-14 Start flip
// (missiles.cpp:3558-3561): the flip is checked on the tick whose decremented
// duration is 14, and the anim step turns negative.
func TestProcessGuardianFadeTailFrameFourteenFlips(t *testing.T) {
	// given
	w, lights := guardianFadeWorld()
	m := guardianFadeMissile(15, 1, -1)

	// when
	missiles.ProcessGuardian(w, &m)

	// then
	if m.Duration != 14 {
		t.Fatalf("duration = %d, want 14", m.Duration)
	}

	if m.MFNum != missiles.GuardianFrameStart {
		t.Fatalf("frame group = %d, want start/%d", m.MFNum, missiles.GuardianFrameStart)
	}

	if m.AnimFrame != 15 {
		t.Fatalf("anim frame = %d, want 15", m.AnimFrame)
	}

	if m.AnimAdd != -1 {
		t.Fatalf("anim add = %d, want -1", m.AnimAdd)
	}

	if m.Var3 != 0 {
		t.Fatalf("var3 = %d, want 0", m.Var3)
	}

	if len(lights.radii) != 0 {
		t.Fatalf("light radii = %v, want none", lights.radii)
	}
}

// TestProcessGuardianFadeTailRampsLight pins the var3 ramp
// (missiles.cpp:3564-3569): a positive var3 after the add reaches ChangeLight
// with the new radius.
func TestProcessGuardianFadeTailRampsLight(t *testing.T) {
	// given
	w, lights := guardianFadeWorld()
	m := guardianFadeMissile(14, 10, 1)

	// when
	missiles.ProcessGuardian(w, &m)

	// then
	if m.Duration != 13 {
		t.Fatalf("duration = %d, want 13", m.Duration)
	}

	if m.Var3 != 11 {
		t.Fatalf("var3 = %d, want 11", m.Var3)
	}

	if m.MFNum != missiles.GuardianFrameIdle {
		t.Fatalf("frame group = %d, want idle/%d", m.MFNum, missiles.GuardianFrameIdle)
	}

	if !slices.Equal(lights.radii, []uint8{11}) {
		t.Fatalf("light radii = %v, want [11]", lights.radii)
	}

	if m.DelFlag || lights.unlight != 0 {
		t.Fatalf("del/unlight = %v/%d, want false/0", m.DelFlag, lights.unlight)
	}
}

// TestProcessGuardianFadeTailClampsAboveFifteen pins the silent clamp
// (missiles.cpp:3566-3567): var3 above 15 is cut to 15 and no light call runs.
func TestProcessGuardianFadeTailClampsAboveFifteen(t *testing.T) {
	// given
	w, lights := guardianFadeWorld()
	m := guardianFadeMissile(14, 15, 1)

	// when
	missiles.ProcessGuardian(w, &m)

	// then
	if m.Duration != 13 {
		t.Fatalf("duration = %d, want 13", m.Duration)
	}

	if m.Var3 != 15 {
		t.Fatalf("var3 = %d, want the 15 clamp", m.Var3)
	}

	if m.MFNum != missiles.GuardianFrameIdle {
		t.Fatalf("frame group = %d, want idle/%d", m.MFNum, missiles.GuardianFrameIdle)
	}

	if len(lights.radii) != 0 {
		t.Fatalf("light radii = %v, want none", lights.radii)
	}
}

// TestProcessGuardianFadeTailZeroLeavesLightAlone pins the non-positive arm
// (missiles.cpp:3568): a var3 that lands on 0 makes no ChangeLight call.
func TestProcessGuardianFadeTailZeroLeavesLightAlone(t *testing.T) {
	// given
	w, lights := guardianFadeWorld()
	m := guardianFadeMissile(14, 0, 0)

	// when
	missiles.ProcessGuardian(w, &m)

	// then
	if m.Var3 != 0 {
		t.Fatalf("var3 = %d, want 0", m.Var3)
	}

	if len(lights.radii) != 0 {
		t.Fatalf("light radii = %v, want none", lights.radii)
	}

	if m.DelFlag || lights.unlight != 0 {
		t.Fatalf("del/unlight = %v/%d, want false/0", m.DelFlag, lights.unlight)
	}
}

// guardianFadeMissile builds an idle guardian that skips the sweep and the
// frame-group handover, so only the fade tail runs.
func guardianFadeMissile(duration, var3, animAdd int) missiles.Missile {
	m := missiles.Missile{
		Type: missiles.Guardian, Source: 0, Duration: duration,
		Var1: 120, Var2: 3, Var3: var3, Lid: 9, AnimLen: 8,
		AnimAdd: animAdd, MFNum: missiles.GuardianFrameIdle,
	}
	m.Position.Tile = missiles.TilePosition{X: 5, Y: 5}

	return m
}

// guardianFadeWorld wires the light sinks of the fade tick.
func guardianFadeWorld() (*missiles.HitWorld, *guardianFadeLights) {
	lights := &guardianFadeLights{}
	cb := missiles.NoopProcessCallbacks()
	cb.ChangeLight = func(_, _, _ int, radius uint8) {
		lights.radii = append(lights.radii, radius)
	}
	cb.AddUnLight = func(int) { lights.unlight++ }

	return &missiles.HitWorld{
		Grid:             world.NewGrid(),
		HitCallbacks:     missiles.NoopHitCallbacks(),
		ProcessCallbacks: cb,
	}, lights
}
