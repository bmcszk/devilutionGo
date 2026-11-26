package game

// tickrate.go registers the tick-rate verb: the one menu action that needs a
// kernel command, because it changes the clock-to-tick ratio itself
// (X:Source/gamemenu.cpp:264-278 GamemenuSpeed -> gnTickDelay = 1000 /
// nTickRate). Pause only decides WHETHER ticks run; the rate decides how many
// run per wall-second, so two replaying clients must agree on it (ADR-018
// law 4). The handler is sim-pure: no os, no wall clock — the persistence
// write stays in the boot-injected onRate closure.

import (
	"time"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// RegisterTickRate wires the SetTickRate verb onto the kernel: the handler
// sets the game's live tick delay ONLY when the rate differs from the current
// one (ADR-018 law 2: no change, no event), emits one TickRateSet event with
// A = rate, and then calls onRate for the presentation's own bookkeeping when
// non-nil. The handler body does no I/O.
func RegisterTickRate(k *kernel.Kernel, g *Game, onRate func(int)) {
	k.Register(kernel.KindSetTickRate, func(c kernel.Command, _ *kernel.Streams) []kernel.Event {
		delay := time.Second / time.Duration(int(c.A))
		if g.tickDelay == delay {
			return nil
		}

		g.tickDelay = delay

		if onRate != nil {
			onRate(int(c.A))
		}

		return []kernel.Event{{Kind: kernel.KindTickRateSet, A: c.A}}
	})
}
