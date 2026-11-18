package inputbridge

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// mousePathRecorder is the MousePathHooks recorder: it watches which hook
// fired and carries the command that row means.
type mousePathRecorder struct {
	cmd   kernel.Command
	fired bool
}

// MousePathCommand translates one LeftMouseCmd decision (diablo.cpp:247-319
// @ fork ca91d778f, devilGo-c6wt s2) into the command its verb row means:
// one decision, one command, the payload words carry the decision's own
// values (integer-only per ADR-015). The bridge re-decides with the
// controls.MousePathAction machine itself and records which hook fired — it
// does not re-derive any branch, so the row order stays the upstream order.
//
// Refusal rows: upstream's town branch can legitimately fire nothing (item
// under a non-hand cursor with no towner, diablo.cpp:255-264) — such a
// decision yields no command, mirroring the stick path's nothing decision.
func MousePathCommand(in controls.MousePathInput, target controls.MousePathTarget) (kernel.Command, bool) {
	recorder := mousePathRecorder{}
	controls.MousePathAction(in, target, recorder.hooks())

	return recorder.command()
}

// command reports the recorded row's command; a silent decision (the town
// refusal rows) yields nothing.
func (m *mousePathRecorder) command() (kernel.Command, bool) {
	return m.cmd, m.fired
}

// hooks returns the recorder hooks: every hook stores the command its
// upstream row means and marks the decision fired.
func (m *mousePathRecorder) hooks() controls.MousePathHooks {
	return controls.MousePathHooks{
		SendWalk: func(x, y int) {
			m.emit(kernel.Command{Kind: kernel.KindMouseWalk, A: int32(x), B: int32(y)})
		},
		SendTalk: func(monster int) {
			m.emit(kernel.Command{Kind: kernel.KindMouseTalk, A: int32(monster)})
		},
		SendPickup:        m.pickup,
		SendOperateObject: m.object,
		SendAttack:        m.attack,
	}
}

// pickup records the pickup row (diablo.cpp:255-256): A the item id, B the
// near flag (invflag asked CMD_GOTOGETITEM).
func (m *mousePathRecorder) pickup(p controls.MousePickup) {
	near := int32(0)
	if p.Near {
		near = 1
	}
	m.emit(kernel.Command{Kind: kernel.KindMousePickup, A: int32(p.Item), B: near})
}

// object records the object row (diablo.cpp:273-275): A the object id, B the
// disarm flag (pcurs == CURSOR_DISARM asked CMD_DISARMXY).
func (m *mousePathRecorder) object(o controls.MouseObject) {
	disarm := int32(0)
	if o.Disarm {
		disarm = 1
	}
	m.emit(kernel.Command{Kind: kernel.KindMouseObject, A: int32(o.Object), B: disarm})
}

// attack records the attack row (diablo.cpp:277-316): A the ranged flag
// (weapon truth), B the verb word — 0 plain id, 1 tile, 2 talk (the
// CanTalkToMonst ATTACKID, diablo.cpp:283-284, reachable under a ranged
// weapon) — and C the id target (monster id; a player id shares the word —
// the handler disambiguates by the target registry).
func (m *mousePathRecorder) attack(a controls.MouseAttack) {
	id := 0
	if a.Monster != -1 {
		id = a.Monster
	} else if a.Player != -1 {
		id = a.Player
	}
	ranged := int32(0)
	if a.Ranged {
		ranged = 1
	}
	verb := int32(0)
	switch {
	case a.Talk:
		verb = 2
	case a.Tile:
		verb = 1
	default:
		verb = 0
	}
	m.emit(kernel.Command{
		Kind: kernel.KindMouseAttack,
		A:    ranged,
		B:    verb,
		C:    int32(id),
	})
}

// emit records one row's command.
func (m *mousePathRecorder) emit(cmd kernel.Command) {
	m.cmd = cmd
	m.fired = true
}
