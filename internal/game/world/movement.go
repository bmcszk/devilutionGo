package world

import "github.com/bmcszk/devilutionGo/internal/game/kernel"

// RegisterMovement wires the handlers that own the grid slice onto k: a
// PlayerMove command walks the player and announces PlayerMoved, and a move the
// grid refuses changes nothing and emits nothing (ADR-018 laws 2 and 3).
func RegisterMovement(k *kernel.Kernel, s *State) {
	k.Register(kernel.KindPlayerMove, moveHandler(s))
}

// moveHandler applies one PlayerMove command to the slice.
func moveHandler(s *State) kernel.Handler {
	return func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		dst := Tile{X: int(c.A), Y: int(c.B)}

		if !s.StepTo(dst) {
			return nil
		}

		return []kernel.Event{{Kind: kernel.KindPlayerMoved, A: c.A, B: c.B}}
	}
}
