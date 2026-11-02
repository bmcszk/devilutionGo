// Package missiles: the Hellfire Warp add body.
package missiles

import "github.com/bmcszk/devilutionGo/internal/engine"

// Level, set-level and interface-message constants of AddWarp
// (levels/gendung_defs.hpp:15-25,48-60; interfac.h:24-33).
const (
	dtypeTown      = 0
	dtypeCathedral = 1
	dtypeCatacombs = 2
	dtypeCaves     = 3
	dtypeHell      = 4
	dtypeNest      = 5
	dtypeCrypt     = 6
	slVileBetrayer = 5
	wmDiabNextLvl  = 0
	wmDiabPrevLvl  = 1
	wmDiabRtnLvl   = 2
	wmDiabTwarpUp  = 6
	maxTriggers    = 8
	warpSearchMax  = 5
	warpDuration   = 2
	// maxDistance is std::numeric_limits<int>::max(), the AddWarp start value
	// (missiles.cpp:1509).
	maxDistance = int(^uint(0) >> 1)
)

// AddWarp ports AddWarp (missiles.cpp:1507-1580): the player teleports to the
// tile in front of the level-stair trigger closest to it — the offset in front
// depends on the dungeon type and on which stair the trigger represents — and
// the row lasts 2 ticks; when no free destination exists the row deletes itself
// and the spell fizzles.
func AddWarp(w *HitWorld, m *Missile, p *AddParameter) {
	player := m.Source
	tile := w.AddSource.playerStats(player).Tile // cpp:1515
	tile = closestStairTile(w, tile)             // cpp:1517-1568

	m.Duration = warpDuration // cpp:1569

	destination, found := engine.FindClosestValidPosition(
		func(target engine.Point) bool { return warpFreeTile(w, player)(target.X, target.Y) },
		engine.Point{X: tile.X, Y: tile.Y}, 0, warpSearchMax,
	)
	if !found { // cpp:1577-1580
		m.DelFlag = true
		p.SpellFizzled = true

		return
	}

	m.Position.Tile = TilePosition{X: destination.X, Y: destination.Y} // cpp:1575
}

// closestStairTile walks the trigger table and returns the tile in front of the
// closest level stair (missiles.cpp:1517-1568).
func closestStairTile(w *HitWorld, from TilePosition) TilePosition {
	tile := from
	minDistanceSq := maxDistance

	count := w.ProcessCallbacks.TriggerCount()

	for i := 0; i < count && i < maxTriggers; i++ {
		message := w.ProcessCallbacks.TriggerMessage(i)
		if !isLevelStair(message) {
			continue
		}

		candidate := w.ProcessCallbacks.TriggerPosition(i)
		offset := w.ProcessCallbacks.triggerOffset(message)
		candidate = TilePosition{X: candidate.X + offset.DeltaX, Y: candidate.Y + offset.DeltaY}

		deltaY := from.Y - candidate.Y
		deltaX := from.X - candidate.X
		if distanceSq := deltaY*deltaY + deltaX*deltaX; distanceSq < minDistanceSq {
			minDistanceSq = distanceSq
			tile = candidate
		}
	}

	return tile
}

// isLevelStair reports whether tmsg is one of the four stair messages AddWarp
// considers (missiles.cpp:1519).
func isLevelStair(message int) bool {
	switch message {
	case wmDiabTwarpUp, wmDiabPrevLvl, wmDiabNextLvl, wmDiabRtnLvl:
		return true
	default:
		return false
	}
}

// triggerOffset ports the getTriggerOffset lambda of AddWarp
// (missiles.cpp:1522-1563): the tile in front of a stair depends on the dungeon
// type and on which stair the message represents.
func (cb *ProcessCallbacks) triggerOffset(message int) engine.Displacement {
	switch cb.LevelType {
	case dtypeCathedral:
		return cb.cathedralOffset(message)
	case dtypeCatacombs:
		return messageOffset(message, isUpOrPrevious,
			engine.Displacement{DeltaX: 1, DeltaY: 1}, engine.Displacement{DeltaX: 0, DeltaY: 1})
	case dtypeCaves:
		return messageOffset(message, isUpOrPrevious,
			engine.Displacement{DeltaX: 0, DeltaY: 1}, engine.Displacement{DeltaX: 1, DeltaY: 0})
	case dtypeHell:
		return engine.Displacement{DeltaX: 1, DeltaY: 0}
	case dtypeNest:
		return messageOffset(message, isUpOrPreviousOrReturn,
			engine.Displacement{DeltaX: 0, DeltaY: 1}, engine.Displacement{DeltaX: 1, DeltaY: 0})
	default: // DTYPE_CRYPT, and TOWN/NONE are unreachable upstream
		return messageOffset(message, isUpOrPreviousOrReturn,
			engine.Displacement{DeltaX: 1, DeltaY: 1}, engine.Displacement{DeltaX: 0, DeltaY: 1})
	}
}

// cathedralOffset is the cathedral branch of getTriggerOffset
// (missiles.cpp:1523-1530), including the Vile Betrayer portal corner.
func (cb *ProcessCallbacks) cathedralOffset(message int) engine.Displacement {
	if cb.IsSetLevel() && cb.SetLevelNum() == slVileBetrayer {
		return engine.Displacement{DeltaX: 1, DeltaY: 1} // Portal
	}

	return messageOffset(message, isUpOrPrevious,
		engine.Displacement{DeltaX: 1, DeltaY: 2}, engine.Displacement{DeltaX: 0, DeltaY: 1})
}

// messageOffset picks the first displacement when message matches, else the
// second — the two-way shape every dungeon-type branch shares
// (missiles.cpp:1525-1562).
func messageOffset(message int, matches func(int) bool, when, otherwise engine.Displacement) engine.Displacement {
	if matches(message) {
		return when
	}

	return otherwise
}

// isUpOrPrevious is the twarp-up/previous-level test of the cathedral and
// catacombs branches (missiles.cpp:1525, 1538).
func isUpOrPrevious(message int) bool {
	return message == wmDiabTwarpUp || message == wmDiabPrevLvl
}

// isUpOrPreviousOrReturn adds the return-level message (missiles.cpp:1550,
// 1557).
func isUpOrPreviousOrReturn(message int) bool {
	return isUpOrPrevious(message) || message == wmDiabRtnLvl
}

// warpFreeTile is the PosOkPlayer-and-not-a-trigger predicate AddWarp places
// with (missiles.cpp:1571-1575).
func warpFreeTile(w *HitWorld, player int) func(int, int) bool {
	return func(x, y int) bool {
		for i := range w.ProcessCallbacks.TriggerCount() {
			if w.ProcessCallbacks.TriggerPosition(i) == (TilePosition{X: x, Y: y}) {
				return false
			}
		}

		return w.ProcessCallbacks.PosOkPlayer(player, x, y)
	}
}
