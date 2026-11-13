package inputbridge

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// Session is the headless decision source's submit site: it buffers the
// commands one tick's decisions produced and submits them at the tick
// boundary, in decision order, through the kernel's gate — the same boundary
// DemoReplay submits at (internal/game/demo.go, ADR-018 laws 1 and 3). The
// device lane later replaces the headless decision calls with real axis and
// key reading; the submit contract stays this one.
type Session struct {
	kernel  *kernel.Kernel
	pending []kernel.Command
}

// NewSession binds the submit site to the kernel that runs the passes.
func NewSession(k *kernel.Kernel) *Session {
	return &Session{kernel: k}
}

// StickDecision buffers one classified left-stick decision for the current
// tick. The nothing decision buffers nothing: a quiet gamepad queues no
// command.
func (s *Session) StickDecision(push controls.StickPush, dir controls.AxisDirection) {
	if cmd, ok := StickCommand(push, dir); ok {
		s.pending = append(s.pending, cmd)
	}
}

// PrimaryDecision buffers one primary-action dispatch decision.
func (s *Session) PrimaryDecision(view controls.ActionView) {
	if cmd, ok := PrimaryActionCommand(view); ok {
		s.pending = append(s.pending, cmd)
	}
}

// ReleaseDecision buffers one primary-action release decision.
func (s *Session) ReleaseDecision(view controls.ActionView) {
	if cmd, ok := ReleaseCommand(view); ok {
		s.pending = append(s.pending, cmd)
	}
}

// SubmitDecision queues one already-built command through the kernel's gate
// and reports whether the gate accepted it; a refused command never reaches a
// handler (ADR-018 law 6).
func (s *Session) SubmitDecision(cmd kernel.Command) bool {
	return s.kernel.Submit(cmd)
}

// Tick closes the tick boundary: the buffered decisions submit at tick in
// decision order, stamped with the tick they belong to (ADR-018 law 1, the
// demo.go buffer() discipline), then the kernel runs its single ordered pass.
func (s *Session) Tick(tick uint32) {
	for _, cmd := range s.pending {
		cmd.Tick = tick
		s.kernel.Submit(cmd)
	}

	s.pending = nil

	s.kernel.Pass(tick)
}
