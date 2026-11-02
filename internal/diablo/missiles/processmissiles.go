// Package missiles: the collection-level ProcessMissiles driver — the tick
// entry point that runs the process pass and the DeleteMissiles sweeps.
package missiles

import "github.com/bmcszk/devilutionGo/internal/game/world"

// ProcessMissiles ports the collection-level ProcessMissiles driver
// (missiles.cpp:4218-4256) over the donburi missile container:
//
//  1. the bounds pass flags every row that left the dungeon and clears the
//     missile tile flags of the rest (cpp:4220-4227),
//  2. DeleteMissiles sweeps the flagged rows (cpp:4194, call site cpp:4229),
//  3. MissilePreFlag is reset for the frame (cpp:4231),
//  4. every surviving row runs its processFn in list order (cpp:4233-4248),
//  5. DeleteMissiles sweeps again so a row the pass flagged goes away
//     (cpp:4194, call site cpp:4255).
//
// Iteration is index-based over the order list so a row appended by a
// processFn joins the tail and is visited by the same pass — the std::list
// range-for behaviour of the C++ loop. Rows flagged during the pass stay in
// the list until the trailing sweep, matching _miDelFlag semantics.
//
// Deviation: the animation advance of the loop (cpp:4235-4247) is not ported
// here — it keys off the misdat NotAnimated flags column, which Missile does
// not carry yet. ProcessManaShield (cpp:4254) is a player-state tick the
// engine owns.
func (c *MissileCollection) ProcessMissiles(w *HitWorld) {
	for i := range c.order {
		m := MissileComponent.Get(c.world.Entry(c.order[i]))
		if !world.InDungeonBounds(m.Position.Tile.X, m.Position.Tile.Y) {
			m.DelFlag = true // cpp:4225

			continue
		}

		clearMissileTileFlags(w, m) // cpp:4223
	}

	c.RemoveMarked() // DeleteMissiles, cpp:4229

	w.ProcessCallbacks.MissilePreFlag = false // cpp:4231

	// len is re-read per step so a row spawned by a processFn joins this pass.
	for i := 0; i < len(c.order); i++ { //nolint:intrange // dynamic bound, std::list range-for
		m := MissileComponent.Get(c.world.Entry(c.order[i]))

		if behavior, ok := BehaviorFor(m.Type); ok {
			behavior.Process(w, m) // cpp:4235
		}
	}

	c.RemoveMarked() // DeleteMissiles, cpp:4255
}

// clearMissileTileFlags ports the dFlags clear of the bounds pass
// (missiles.cpp:4222-4223): the row no longer blocks its tile as a missile,
// firewall or lightning wall. PutMissile raises the same three bits.
func clearMissileTileFlags(w *HitWorld, m *Missile) {
	flags := &w.Grid.DFlags[m.Position.Tile.Y][m.Position.Tile.X]
	*flags &= ^(world.FlagMissile | world.FlagMissileFire | world.FlagMissileLight)
}
