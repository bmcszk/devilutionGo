// Package missiles: the chain lightning fork body.
package missiles

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// maxCrawlRadius ports MaxCrawlRadius (lighting.h:74), the cap on the chain
// radius.
const maxCrawlRadius = 18

// ProcessChainLightning ports ProcessChainLightning (missiles.cpp:3572-3600):
// the bolt forks one control missile at the remembered aim tile, then one more
// at every monster inside min(spell level + 3, MaxCrawlRadius) (cpp:3587-3595),
// and dies after its single tick.
func ProcessChainLightning(w *HitWorld, m *Missile) {
	id := m.Source
	position := m.Position.Tile
	dst := TilePosition{X: m.Var1, Y: m.Var2} // cpp:3577

	dir := GetDirection(position, dst) // cpp:3578
	forkChain(w, m, position, dst, dir, id)

	radius := minInt(m.SpellLevel+3, maxCrawlRadius) // cpp:3587

	engine.Crawl(1, uint(radius), func(displacement engine.Displacement) (engine.Displacement, bool) {
		target := TilePosition{X: position.X + displacement.DeltaX, Y: position.Y + displacement.DeltaY}
		if chainTargetsMonster(w, target) { // cpp:3590
			forkChain(w, m, position, target, GetDirection(position, target), id)
		}

		return engine.Displacement{}, false
	})

	m.Duration-- // cpp:3596
	if m.Duration == 0 {
		m.DelFlag = true // cpp:3598-3599
	}
}

// chainTargetsMonster ports the fork predicate (missiles.cpp:3590):
// InDungeonBounds(target) && dMonster[target.x][target.y] > 0.
func chainTargetsMonster(w *HitWorld, target TilePosition) bool {
	return world.InDungeonBounds(target.X, target.Y) && w.DMonster[target.Y][target.X] > 0
}

// forkChain issues one LightningControl missile from position toward dst
// (missiles.cpp:3579, 3592).
func forkChain(w *HitWorld, m *Missile, position, dst TilePosition, dir Direction, id int) {
	w.ProcessCallbacks.SpawnMissile(position, dst, dir, LightningControl, int(TargetMonsters),
		id, 1, m.SpellLevel, nil)
}
