package world

import "github.com/bmcszk/devilutionGo/internal/game/kernel"

// StepReach is the furthest tile one click command may act on. The slice
// executes one grid step per command, so a click beyond an adjacent tile is out
// of range until the path-following bean turns a click into a path (ADR-018
// law 6 drops the rest).
const StepReach = 1

// Gate rules, in decision order: the first failing rule names the rejection.
const (
	RuleAccepted GateRule = ""
	// RuleSameLevel is judged first: a same-level reload is dropped before any
	// tile or cooldown rule sees the command (levelRule).
	RuleSameLevel    GateRule = "same-level"
	RuleBlockedSpawn GateRule = "blocked-spawn"
	RuleOutOfRange   GateRule = "out-of-range"
	RuleDeadTarget   GateRule = "dead-target"
	RuleCooldown     GateRule = "cooldown"
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

// LevelView is the read-only level state the descend rule consults: the level
// the session is already on. A nil LevelView makes the gate judge no descend.
type LevelView interface {
	// Level reports the level number and dungeon_type the session is on, and
	// ok=false while the session has not loaded a dungeon at all.
	Level() (level, levelType int, ok bool)
}

// SpawnView is the read-only tile state the create rule consults.
type SpawnView interface {
	// SpawnAllowed reports whether a hero may be raised on tile.
	SpawnAllowed(tile Tile) bool
}

// SessionView is the session state the gate judges beyond the player's own
// position: the level it is on and where a hero may be raised.
type SessionView interface {
	LevelView
	SpawnView
}

// GateRule names why a command was dropped. The empty rule means accepted.
type GateRule string

// Gate drops commands that cannot change anything, before dispatch (ADR-018
// law 6). It decides over one command and the state view alone: it writes no
// game state, reads no clock, and its only memory is the tick it last accepted
// per kind, which the command's own tick advances.
type Gate struct {
	view     View
	level    LevelView
	spawn    SpawnView
	cooldown uint32
	last     map[kernel.Kind]uint32
}

// NewGate builds the gate over view; cooldown is the minimum number of ticks
// between two accepted commands of the same kind, 0 for none.
func NewGate(view View, cooldown uint32) *Gate {
	return &Gate{view: view, cooldown: cooldown, last: make(map[kernel.Kind]uint32)}
}

// NewSessionGate builds the gate over view and the session state, so a descend
// onto the level the session already holds and a create on a tile no hero may
// stand on are both dropped before dispatch.
func NewSessionGate(view View, session SessionView, cooldown uint32) *Gate {
	gate := NewGate(view, cooldown)
	gate.level = session
	gate.spawn = session

	return gate
}

// Allow implements kernel.Gate: a command with a rejection rule is dropped.
func (g *Gate) Allow(c kernel.Command) bool { return g.Reject(c) == RuleAccepted }

// Reject reports the rule that drops c, or RuleAccepted when the command may be
// dispatched. An accepted command records its tick, so the cooldown rule sees
// every command that passed the gate.
func (g *Gate) Reject(c kernel.Command) GateRule {
	for _, rule := range []func(kernel.Command) GateRule{g.levelRule, g.spawnRule, g.tileRule} {
		if rejected := rule(c); rejected != RuleAccepted {
			return rejected
		}
	}

	g.last[c.Kind] = c.Tick

	return RuleAccepted
}

// tileRule applies the reach, target and cooldown rules to the commands that
// carry a tile; any other kind passes untouched.
func (g *Gate) tileRule(c kernel.Command) GateRule {
	tile, gameplay := gameplayTile(c)
	if !gameplay {
		return RuleAccepted
	}

	return g.rejection(c, tile)
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

// levelRule drops a descend onto the level the session already holds: reloading
// the same dungeon tears down and rebuilds every grid for no change (ADR-018
// law 6). A session that has not loaded a dungeon has nothing to reload, so the
// first descend always passes.
func (g *Gate) levelRule(c kernel.Command) GateRule {
	if c.Kind != kernel.KindDescend || g.level == nil {
		return RuleAccepted
	}

	level, levelType, loaded := g.level.Level()
	if loaded && level == int(c.A) && levelType == int(c.B) {
		return RuleSameLevel
	}

	return RuleAccepted
}

// spawnRule drops a create the map cannot host: a hero raised inside a wall or
// outside the dungeon cannot stand there, so the pass would change nothing
// (ADR-018 law 6). A gate with no SpawnView judges no create.
func (g *Gate) spawnRule(c kernel.Command) GateRule {
	if c.Kind != kernel.KindCreatePack || g.spawn == nil {
		return RuleAccepted
	}

	if g.spawn.SpawnAllowed(Tile{X: int(c.B), Y: int(c.C)}) {
		return RuleAccepted
	}

	return RuleBlockedSpawn
}

// gameplayTile decodes the tile payload of the commands the gate judges; any
// other kind passes untouched.
func gameplayTile(c kernel.Command) (Tile, bool) {
	if c.Kind != kernel.KindPlayerMove && c.Kind != kernel.KindPlayerAttack {
		return Tile{}, false
	}

	return Tile{X: int(c.A), Y: int(c.B)}, true
}
