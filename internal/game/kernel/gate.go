package kernel

// GateFunc adapts a plain function to the Gate seam.
type GateFunc func(Command) bool

// Allow runs the wrapped rule.
func (f GateFunc) Allow(c Command) bool { return f(c) }

// DropUnnamed is the first gate rule: a command with no Kind can never reach a
// handler, so it cannot change state and it must not cost a dispatch. The full
// gameplay gate (range, dead target, mana, cooldown) is the next bean; this is
// the seam it plugs into.
func DropUnnamed() Gate {
	return GateFunc(func(c Command) bool { return c.Kind != "" })
}
