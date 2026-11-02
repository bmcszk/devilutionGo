// Package missiles: the Hellfire spawn add bodies — the nest chain, the hork
// spawn, the jester's random cast, the lightning wall and the Search timer.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// math_maxInt is std::numeric_limits<int>::max() for the AddSearch guard
// (missiles.cpp:1710).
const math_maxInt = int(^uint(0) >> 1)

// openNestSpawnTiles are the four tiles AddOpenNest detonates
// (missiles.cpp:1240-1242), aiming them all at the nest mouth.
var openNestSpawnTiles = []TilePosition{{X: 80, Y: 62}, {X: 81, Y: 62}, {X: 80, Y: 63}, {X: 81, Y: 63}}

// openNestAim is the AddMissile dst every nest explosion shares
// (missiles.cpp:1241: { 80, 62 }).
var openNestAim = TilePosition{X: 80, Y: 62}

// AddOpenNest ports AddOpenNest (missiles.cpp:1240-1247): four BigExplosion
// rows at the nest mouth, each casting on the nest's behalf, then the row
// deletes itself.
func AddOpenNest(w *HitWorld, m *Missile, p *AddParameter) {
	for _, tile := range openNestSpawnTiles {
		w.ProcessCallbacks.SpawnMissileRow(AddMissileArgs{ // cpp:1241
			Src: tile, Dst: openNestAim, Midir: p.Midir, Type: BigExplosion,
			Caster: m.Caster, Source: m.Source, Dam: m.Dam,
		})
	}

	m.DelFlag = true // cpp:1245
}

// AddHorkSpawn ports AddHorkSpawn (missiles.cpp:1346-1352): a corpse row flying
// at 8 px per tick for 9 ticks, remembering its direction in var1, and placed
// on the map so it is visible from the first tick.
func AddHorkSpawn(w *HitWorld, m *Missile, p *AddParameter) {
	UpdateMissileVelocity(m, p.Dst, 8) // cpp:1348
	m.Duration = 9                     // cpp:1349
	m.Var1 = int(p.Midir)              // cpp:1350
	PutMissile(w, m)                   // cpp:1351
}

// AddJester ports AddJester (missiles.cpp:1354-1390): one of ten spells picked
// by a d10 is cast from the jester's tile at the aim; the jester's row then
// deletes itself and reports whether the random cast fizzled.
func AddJester(w *HitWorld, m *Missile, p *AddParameter) {
	spell := jesterSpell(w) // cpp:1356-1384

	child := w.ProcessCallbacks.SpawnMissileRow(AddMissileArgs{ // cpp:1385-1386
		Src: m.Position.Start, Dst: p.Dst, Midir: p.Midir, Type: spell,
		Caster: m.Caster, Source: m.Source, SpellLevel: m.SpellLevel, Parent: m,
	})

	p.SpellFizzled = child == nil // cpp:1387
	m.DelFlag = true              // cpp:1388
}

// AddLightningWall ports AddLightningWall (missiles.cpp:1582-1601): a wall
// moving at 16 px per tick for 255 ticks per spell level, starting on a random
// frame, launched from the cast tile — the trap's own tile or the player's.
func AddLightningWall(w *HitWorld, m *Missile, p *AddParameter) {
	UpdateMissileVelocity(m, p.Dst, 16)     // cpp:1584
	m.AnimFrame = randomIntBetween(w, 1, 8) // cpp:1585
	m.Duration = 255 * (m.SpellLevel + 1)   // cpp:1586

	switch m.SourceOf() { // cpp:1587-1600
	case SourcePlayer:
		player := w.AddSource.playerStats(m.Source)
		m.Var1 = player.Tile.X
		m.Var2 = player.Tile.Y
	default:
		// Trap and monster sources both keep the launch tile; a monster
		// source is unreachable upstream (asserted).
		m.Var1 = m.Position.Start.X
		m.Var2 = m.Position.Start.Y
	}
}

// AddSearch ports AddSearch (missiles.cpp:1693-1714): the row's lifetime is
// 245 + 10 per spell level + twice the caster's level and it merges into an
// equal-source Search row that is already flying, deleting itself; a player
// caster also turns on the auto-map item display.
func AddSearch(w *HitWorld, m *Missile, _ *AddParameter) {
	if w.ProcessCallbacks.IsMyPlayer(m.Source) { // cpp:1697-1698
		w.ProcessCallbacks.ShowAutoMapItems()
	}

	lvl := 2 // cpp:1699-1701
	if m.Source >= 0 {
		lvl = w.AddSource.playerStats(m.Source).Level * 2
	}

	m.Duration = lvl + 10*m.SpellLevel + 245 // cpp:1703

	mergeIntoRunningSearch(w, m) // cpp:1705-1713
}

// mergeIntoRunningSearch folds this row's timer into the first equal-source
// Search row already flying, deleting itself (missiles.cpp:1705-1713).
func mergeIntoRunningSearch(w *HitWorld, m *Missile) {
	for _, other := range w.ProcessCallbacks.OtherMissiles() {
		if other == m || other.Type != Search || !sameSource(m, other) {
			continue
		}

		if other.Duration < math_maxInt-m.Duration { // cpp:1709-1711
			other.Duration += m.Duration
		}

		m.DelFlag = true

		return
	}
}

// sameSource ports Missile::isSameSource (missiles.h:160-163): equal source
// kind and equal source index.
func sameSource(a, b *Missile) bool {
	return a.SourceOf() == b.SourceOf() && a.Source == b.Source
}

// findTrappedTileAt ports the FindClosestValidPosition call shape the trap rows
// share (missiles.cpp:1303, 1460): an in-dungeon tile within radius of start
// that the predicate accepts.
func findTrappedTileAt(_ *HitWorld, start TilePosition, radius int, ok func(x, y int) bool) (TilePosition, bool) {
	spot, found := engine.FindClosestValidPosition(
		func(target engine.Point) bool {
			return world.InDungeonBounds(target.X, target.Y) && ok(target.X, target.Y)
		},
		engine.Point{X: start.X, Y: start.Y}, 0, uint(radius),
	)
	if !found {
		return TilePosition{}, false
	}

	return TilePosition{X: spot.X, Y: spot.Y}, true
}

// jesterSpell maps the d10 roll to the spell the jester casts
// (missiles.cpp:1356-1384).
func jesterSpell(w *HitWorld) MissileID {
	switch randomIntLessThan(w, 10) {
	case 0, 1:
		return Firebolt
	case 2:
		return Fireball
	case 3:
		return FireWallControl
	case 4:
		return Guardian
	case 5:
		return ChainLightning
	case 6:
		return TownPortal
	case 7:
		return Teleport
	case 8:
		return Apocalypse
	default:
		return StoneCurse // roll 9
	}
}
