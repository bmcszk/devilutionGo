package world

import "github.com/bmcszk/devilutionGo/internal/game/kernel"

// StepReach is the furthest tile one click command may act on. The slice
// executes one grid step per command, so a click beyond an adjacent tile is out
// of range until the path-following bean turns a click into a path (ADR-018
// law 6 drops the rest).
const StepReach = 1

// Gate rules, in decision order: the first failing rule names the rejection.
const (
	RuleAccepted   GateRule = ""
	RuleOutOfRange GateRule = "out-of-range"
	RuleDeadTarget GateRule = "dead-target"
	RuleCooldown   GateRule = "cooldown"
)

// View is the read-only state the gate and the input mapping consult. The grid
// slice implements it; tests script their own view so each rule is provable
// without a level.
type View interface {
	// Range is the chebyshev distance from the player to tile.
	Range(tile Tile) int
	// TargetAt reports the target on a tile: ok=false means the tile is
	// empty, alive=false means a dead body lies there.
	TargetAt(tile Tile) (alive, ok bool)
}

// GateRule names why a command was dropped. The empty rule means accepted.
type GateRule string

// Gate drops commands that cannot change anything, before dispatch (ADR-018
// law 6). It decides over one command and the state view alone: it writes no
// game state, reads no clock, and its only memory is the tick it last accepted
// per kind, which the command's own tick advances.
type Gate struct {
	view     View
	cooldown uint32
	last     map[kernel.Kind]uint32
}

// NewGate builds the gate over view; cooldown is the minimum number of ticks
// between two accepted commands of the same kind, 0 for none.
func NewGate(view View, cooldown uint32) *Gate {
	return &Gate{view: view, cooldown: cooldown, last: make(map[kernel.Kind]uint32)}
}

// Allow implements kernel.Gate: a command with a rejection rule is dropped.
func (g *Gate) Allow(c kernel.Command) bool { return g.Reject(c) == RuleAccepted }

// Reject reports the rule that drops c, or RuleAccepted when the command may be
// dispatched. An accepted command records its tick, so the cooldown rule sees
// every command that passed the gate.
func (g *Gate) Reject(c kernel.Command) GateRule {
	tile, gameplay := gameplayTile(c)
	if !gameplay {
		return RuleAccepted
	}

	rule := g.rejection(c, tile)
	if rule != RuleAccepted {
		return rule
	}

	g.last[c.Kind] = c.Tick

	return RuleAccepted
}

// rejection applies the three rules in order: reach, target, cooldown.
func (g *Gate) rejection(c kernel.Command, tile Tile) GateRule {
	if g.view.Range(tile) > StepReach {
		return RuleOutOfRange
	}

	if c.Kind == kernel.KindPlayerAttack {
		if alive, ok := g.view.TargetAt(tile); !ok || !alive {
			return RuleDeadTarget
		}
	}

	return g.cooldownRule(c)
}

// cooldownRule drops a command that arrives before the previous one of its kind
// has cooled down.
func (g *Gate) cooldownRule(c kernel.Command) GateRule {
	if g.cooldown == 0 {
		return RuleAccepted
	}

	if last, seen := g.last[c.Kind]; seen && c.Tick < last+g.cooldown {
		return RuleCooldown
	}

	return RuleAccepted
}

// gameplayTile decodes the tile payload of the commands the gate judges; any
// other kind passes untouched.
func gameplayTile(c kernel.Command) (Tile, bool) {
	if c.Kind != kernel.KindPlayerMove && c.Kind != kernel.KindPlayerAttack {
		return Tile{}, false
	}

	return Tile{X: int(c.A), Y: int(c.B)}, true
}
