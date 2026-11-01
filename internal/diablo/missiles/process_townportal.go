// Package missiles: the Town Portal bodies.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// townPortalIdleGroup is PortalFrame::Idle (missiles.h:79-82).
const townPortalIdleGroup = 1

// townPortalLife is the 100-tick life AddTownPortal sets (missiles.cpp:2126).
const townPortalLife = 100

// townPortalExpLight is the expLight ramp of ProcessTownPortal (cpp:4072).
var townPortalExpLight = [17]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 15, 15}

// AddTownPortal ports AddTownPortal (missiles.cpp:2066-2115). In town the
// portal simply lands on the aim tile; in the dungeon it searches within 5 for
// a tile that is in bounds, free of objects, players and missiles, not solid or
// missile-blocking, and not beside a trigger (cpp:2106-2124). A failed search
// deletes the portal. Either way it lives 100 ticks, flips to the idle frames
// when the countdown reaches 100 - AnimLen, zeroes the caster's older portals
// (cpp:2129-2133) and — for the local player in the dungeon — tells the level
// to activate the portal (cpp:2135-2142).
func AddTownPortal(w *HitWorld, m *Missile, p *AddParameter) {
	if w.ProcessCallbacks.LevelType == levelTypeTown { // cpp:2102-2105
		m.Position.Tile = p.Dst
		m.Position.Start = p.Dst
	} else {
		spot, ok := townPortalSpot(w, p.Dst)
		if !ok { // cpp:2121-2124
			m.DelFlag = true
		} else {
			m.Position.Tile = spot
			m.Position.Start = spot
			m.DelFlag = false
		}
	}

	m.Duration = townPortalLife                    // cpp:2126
	m.Var1 = townPortalLife - m.AnimLen            // cpp:2127
	w.ProcessCallbacks.ExpireSiblingTownPortals(m) // cpp:2128-2133
	PutMissile(w, m)

	if w.ProcessCallbacks.LevelType == levelTypeTown || m.DelFlag { // cpp:2135
		return
	}

	if w.ProcessCallbacks.IsMyPlayer(m.Source) { // cpp:2135-2142
		w.ProcessCallbacks.ActivatePortal(m.Position.Tile)
	}
}

// ProcessTownPortal ports ProcessTownPortal (missiles.cpp:4070-4098): the
// countdown stops at 1, the frames flip to idle at var1, the light ramps one
// expLight entry per tick outside town, every standing player on the portal
// tile has its path cleared and the local one is sent the warp command; the
// last tick deletes the portal and its light.
func ProcessTownPortal(w *HitWorld, m *Missile) {
	if m.Duration > 1 { // cpp:4074-4075
		m.Duration--
	}

	if m.Duration == m.Var1 { // cpp:4076-4077
		m.setFrameGroup(townPortalIdleGroup)
	}

	townPortalLightTick(w, m) // cpp:4078-4083

	for _, player := range w.ProcessCallbacks.ActivePlayers() { // cpp:4085-4094
		if w.ProcessCallbacks.PortalReadyPlayer(player, m.Position.Tile) {
			warpStandingPlayer(w, m, player)
		}
	}

	if m.Duration == 0 { // cpp:4096-4099
		m.DelFlag = true
		w.ProcessCallbacks.AddUnLight(m.Lid)
	}

	PutMissile(w, m)
}

// townPortalSpot ports the FindClosestValidPosition predicate (missiles.cpp:2107-2120).
func townPortalSpot(w *HitWorld, dst TilePosition) (TilePosition, bool) {
	posOk := func(candidate engine.Point) bool {
		return townPortalTileFree(w, TilePosition{X: candidate.X, Y: candidate.Y})
	}

	spot, found := engine.FindClosestValidPosition(posOk, engine.Point{X: dst.X, Y: dst.Y}, 0, 5)
	if !found {
		return TilePosition{}, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, true
}

// townPortalTileFree ports the per-tile half of the search (missiles.cpp:2108-2119).
func townPortalTileFree(w *HitWorld, target TilePosition) bool {
	if !inDungeon(target) { // cpp:2109-2110
		return false
	}

	if w.Grid.FindObjectAtPosition(target.X, target.Y) != nil { // cpp:2112-2113
		return false
	}

	if w.ProcessCallbacks.PlayerAtPosition(target) >= 0 { // cpp:2115-2116
		return false
	}

	if w.ProcessCallbacks.TileContainsMissile(target) { // cpp:2117-2118
		return false
	}

	if w.Grid.TileHasAny(target.X, target.Y, world.TileSolid|world.TileBlockMissile) { // cpp:2119
		return false
	}

	return !w.ProcessCallbacks.TrigNear(target) // cpp:2120
}

// sweepTownPortalLight ports the ramp body (missiles.cpp:4079-4083).
func sweepTownPortalLight(w *HitWorld, m *Missile) {
	if m.Var2 == 0 {
		m.Lid = w.ProcessCallbacks.AddLight(m.Position.Tile.X, m.Position.Tile.Y, 1)
	}

	w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y,
		uint8(townPortalExpLight[minInt(m.Var2, len(townPortalExpLight)-1)]))
	m.Var2++
}

// townPortalLightTick ports the ramp guard (missiles.cpp:4078-4083): outside
// town, before the idle frames and while the portal lives.
func townPortalLightTick(w *HitWorld, m *Missile) {
	if w.ProcessCallbacks.LevelType == levelTypeTown || m.MFNum == townPortalIdleGroup || m.Duration == 0 {
		return
	}

	sweepTownPortalLight(w, m)
}

// warpStandingPlayer ports the warp of one standing player (missiles.cpp:4086-4093).
func warpStandingPlayer(w *HitWorld, m *Missile, player int) {
	w.ProcessCallbacks.ClearPlayerPath(player)

	if w.ProcessCallbacks.IsMyPlayer(player) {
		w.ProcessCallbacks.SendWarpCmd(m.Source)
	}
}
