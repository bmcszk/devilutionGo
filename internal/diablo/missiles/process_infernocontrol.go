// Package missiles: the inferno control walker's bodies.
package missiles

import "github.com/bmcszk/devilutionGo/internal/game/world"

// AddInfernoControl ports AddInfernoControl (missiles.cpp:2685-2696): the walk
// heads 32 pixels per tick toward the aim tile (nudged by midir when it is the
// launch tile), remembers the launch tile in var1/var2 and lives 256 ticks.
func AddInfernoControl(_ *HitWorld, m *Missile, p *AddParameter) {
	dst := p.Dst
	if m.Position.Start == p.Dst { // cpp:2688-2690
		dx, dy := directionDeltaOf(p.Midir)
		dst = TilePosition{X: dst.X + dx, Y: dst.Y + dy}
	}

	UpdateMissileVelocity(m, dst, 32) // cpp:2691

	m.Var1 = m.Position.Start.X // cpp:2692
	m.Var2 = m.Position.Start.Y // cpp:2693
	m.Duration = 256            // cpp:2694
}

// ProcessInfernoControl ports ProcessInfernoControl (missiles.cpp:3961-3987):
// the walk advances one tick, and every tile it reaches that is not the
// remembered one forks an Inferno there — damage carried in var3, which is the
// fork counter — or stops the walk dead on a blocking tile. Three forks or an
// expired duration delete the control missile.
func ProcessInfernoControl(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3963

	m.Position.Traveled.DeltaX += m.Position.Velocity.DeltaX // cpp:3964-3965
	m.Position.Traveled.DeltaY += m.Position.Velocity.DeltaY
	UpdateMissilePos(w, m)

	if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) { // cpp:3966
		infernoControlFork(w, m)
	}

	if m.Duration == 0 || m.Var3 == 3 { // cpp:3986-3987
		m.DelFlag = true
	}
}

// infernoControlFork ports the fork-or-block branch (missiles.cpp:3967-3984).
func infernoControlFork(w *HitWorld, m *Missile) {
	if w.Grid.TileHasAny(m.Position.Tile.X, m.Position.Tile.Y, world.TileBlockMissile) {
		m.Duration = 0 // cpp:3979-3980
	} else {
		w.ProcessCallbacks.SpawnMissile(m.Position.Tile, m.Position.Start, South, Inferno, // cpp:3968-3978
			m.Caster, m.Source, m.Var3, m.SpellLevel, m)
	}

	m.Var1 = m.Position.Tile.X // cpp:3981-3983
	m.Var2 = m.Position.Tile.Y
	m.Var3++
}
