// Package inputbridge translates the mouse click-to-walk decision into the
// kernel command it means (diablo.cpp LeftMouseCmd rows, devilGo-c6wt s2).
package inputbridge_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/diablo/inputbridge"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// mouseCommandCases pin the one-decision-one-command mapping: each upstream
// verb row becomes exactly one kernel kind with its integer payload words.
var mouseCommandCases = []struct {
	name        string
	input       controls.MousePathInput
	want        kernel.Kind
	wantA, B, C int32
}{
	{
		name:  "walk row maps to KindMouseWalk with the tile",
		input: defaultInput(controls.MousePathInput{}),
		want:  kernel.KindMouseWalk, wantA: 66, B: 68,
	},
	{
		name:  "talk row maps to KindMouseTalk with the monster id",
		input: defaultInput(controls.MousePathInput{Town: true, Monster: 7}),
		want:  kernel.KindMouseTalk, wantA: 7,
	},
	{
		name:  "far pickup maps to KindMousePickup with near=0",
		input: defaultInput(controls.MousePathInput{Town: true, Item: 3}),
		want:  kernel.KindMousePickup, wantA: 3, B: 0,
	},
	{
		name:  "near pickup maps to KindMousePickup with near=1",
		input: defaultInput(controls.MousePathInput{Town: true, Item: 3, InventoryOpen: true}),
		want:  kernel.KindMousePickup, wantA: 3, B: 1,
	},
	{
		name:  "object row maps to KindMouseObject",
		input: defaultInput(controls.MousePathInput{Object: 9}),
		want:  kernel.KindMouseObject, wantA: 9, B: 0,
	},
	{
		name:  "disarm object row maps to KindMouseObject with the disarm word",
		input: defaultInput(controls.MousePathInput{Object: 9, Cursor: controls.CursorDisarm}),
		want:  kernel.KindMouseObject, wantA: 9, B: 1,
	},
	{
		name:  "melee tile attack maps to KindMouseAttack tile=1",
		input: defaultInput(controls.MousePathInput{Shift: true}),
		want:  kernel.KindMouseAttack, wantA: 0, B: 1, C: 0,
	},
	{
		name:  "ranged tile attack maps to KindMouseAttack ranged",
		input: defaultInput(controls.MousePathInput{Shift: true, Ranged: true}),
		want:  kernel.KindMouseAttack, wantA: 1, B: 1, C: 0,
	},
	{
		name:  "monster id attack maps to KindMouseAttack",
		input: defaultInput(controls.MousePathInput{Monster: 11}),
		want:  kernel.KindMouseAttack, wantA: 0, B: 0, C: 11,
	},
	{
		name:  "ranged monster id attack maps to KindMouseAttack ranged",
		input: defaultInput(controls.MousePathInput{Monster: 11, Ranged: true}),
		want:  kernel.KindMouseAttack, wantA: 1, B: 0, C: 11,
	},
	{
		name:  "ranged talkable monster maps to the talk verb word",
		input: defaultInput(controls.MousePathInput{Monster: 6, Ranged: true, CanTalkToMonst: true}),
		want:  kernel.KindMouseAttack, wantA: 1, B: 2, C: 6,
	},
}

// TestMousePathCommand maps each row through MousePathCommand.
func TestMousePathCommand(t *testing.T) {
	for _, tc := range mouseCommandCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			cmd, ok := inputbridge.MousePathCommand(tc.input, controls.MousePathTarget{X: 66, Y: 68})
			if !ok {
				t.Fatalf("MousePathCommand(%+v) emitted nothing, want %s", tc.input, tc.want)
			}

			// then
			assertMouseCommand(t, cmd, tc.want, tc.wantA, tc.B, tc.C)
		})
	}
}

// TestMousePathSessionSubmitsThroughGate pins the Session wiring: a mouse
// decision buffers into the tick and submits through the kernel's gate at
// Tick, in decision order (ADR-018 laws 1 and 3).
func TestMousePathSessionSubmitsThroughGate(t *testing.T) {
	// given
	k := kernel.New(nil, nil)
	session := inputbridge.NewSession(k)

	// when
	session.MouseDecision(defaultInput(controls.MousePathInput{}), controls.MousePathTarget{X: 7, Y: 8})
	session.Tick(1)

	// then
	cmds := k.TraceCommandsAt(1)
	if len(cmds) != 1 || cmds[0].Kind != kernel.KindMouseWalk || cmds[0].A != 7 || cmds[0].B != 8 || cmds[0].Tick != 1 {
		t.Fatalf("TraceCommandsAt(1) = %+v, want one stamped KindMouseWalk(7,8)", cmds)
	}
}

// defaultInput fills the -1 id sentinels upstream's cursor state carries
// (plrctrls.cpp:2287-2290) and the resting hand cursor around each case.
func defaultInput(in controls.MousePathInput) controls.MousePathInput {
	if in.Cursor == 0 {
		in.Cursor = controls.CursorHand
	}
	if in.Item == 0 {
		in.Item = -1
	}
	if in.Monster == 0 {
		in.Monster = -1
	}
	if in.Player == 0 {
		in.Player = -1
	}
	if in.Object == 0 {
		in.Object = -1
	}
	return in
}

// assertMouseCommand asserts one emitted command's kind and payload words.
func assertMouseCommand(t *testing.T, cmd kernel.Command, want kernel.Kind, a, b, c int32) {
	t.Helper()
	if cmd.Kind != want || cmd.A != a || cmd.B != b || cmd.C != c {
		t.Fatalf("MousePathCommand = %+v, want %s A=%d B=%d C=%d", cmd, want, a, b, c)
	}
}
