package world

import "github.com/bmcszk/devilutionGo/internal/game/kernel"

// stickPushFirm is StickPush::Firm (controller_motion.h:32-37): the A word of a
// KindStickPush command, and the only severity that walks — a gentle push only
// re-faces upstream (Movement dispatches WalkInDir on a firm push alone,
// plrctrls.cpp:1845-1850).
const stickPushFirm = 2

// faceDir is FaceDir[3][3] (plrctrls.cpp:100-105) in the Direction ordinals of
// engine/direction.hpp:10-20, indexed by the AxisDirectionX/Y ordinals a
// KindStickPush decision carries (axis_direction.h:7-16). It restates the table
// internal/diablo/controls owns because internal/game may not import the diablo
// layer (ADR-008), and the world owns the tile the push steps onto.
//
// NOTE the NONE/NONE cell is South: upstream makes an x-axis-none stick face
// south, not "no direction".
var faceDir = [3][3]int8{
	// NONE          UP          DOWN
	{0, 4, 0}, // NONE: South, North, South
	{2, 3, 1}, // LEFT: West, NorthWest, SouthWest
	{6, 5, 7}, // RIGHT: East, NorthEast, SouthEast
}

// RegisterMovement wires the walk verbs that own the grid slice onto k: the
// tile commands (a PlayerMove click and the bridge's KindMouseWalk row, both
// carrying x/y) step to the tile they name, and a firm KindStickPush steps one
// tile toward its sector (WalkInDir, plrctrls.cpp:1725-1747). Each announces
// PlayerMoved; a step the grid refuses changes nothing and emits nothing
// (ADR-018 laws 2 and 3).
func RegisterMovement(k *kernel.Kernel, s *State) {
	k.Register(kernel.KindPlayerMove, moveHandler(s))
	k.Register(kernel.KindMouseWalk, moveHandler(s))
	k.Register(kernel.KindStickPush, stickPushHandler(s))
}

// moveHandler applies one tile command: A and B are the tile to walk to.
func moveHandler(s *State) kernel.Handler {
	return func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		return step(s, Tile{X: int(c.A), Y: int(c.B)})
	}
}

// stickPushHandler applies one stick decision: a firm push walks one tile
// toward the sector its payload names. The stop-walk and stand-ground branches
// of WalkInDir act on the walkpath and the facing this slice does not carry yet,
// so a push that names no step stays silent (ADR-018 law 2).
func stickPushHandler(s *State) kernel.Handler {
	return func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		facing, ok := stickFacing(c)
		if !ok {
			return nil
		}

		deltaX, deltaY := directionDelta(facing)

		return step(s, Tile{X: s.hero.X + deltaX, Y: s.hero.Y + deltaY})
	}
}

// step walks the player onto tile and announces the move — the one write path
// every walk verb shares. A step the grid refuses (out of reach, blocked, out of
// the dungeon) moves nobody, so it emits nothing (ADR-018 law 2).
func step(s *State, tile Tile) []kernel.Event {
	if !s.StepTo(tile) {
		return nil
	}

	return []kernel.Event{{Kind: kernel.KindPlayerMoved, A: int32(tile.X), B: int32(tile.Y)}}
}

// stickFacing decodes a stick decision's payload into the facing it walks: A
// must be a firm push and B/C the AxisDirectionX/Y ordinals (axis_direction.h:7-16).
// Anything else — a gentle push, a hand-edited log word — names no step, so the
// caller stays silent instead of stepping somewhere the decision never asked for.
func stickFacing(c kernel.Command) (int8, bool) {
	if c.A != stickPushFirm || !inAxisRange(c.B) || !inAxisRange(c.C) {
		return 0, false
	}

	return faceDir[c.B][c.C], true
}

// inAxisRange reports whether word is one of the three AxisDirectionX/Y ordinals.
func inAxisRange(word int32) bool { return word >= 0 && word < 3 }
