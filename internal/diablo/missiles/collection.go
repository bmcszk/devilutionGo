// Package missiles: donburi-backed missile collection. Additive surface
// alongside MissileList (row list); this collection stores each missile as
// an ECS entity with the MissileComponent payload and keeps an explicit
// order list reproducing std::list<Missile> iteration order:
// emplace_back appends at the tail (marked rows stay in place until
// remove_if), removal preserves survivor order (missiles.h:238,
// missiles.cpp AddMissile/DeleteMissiles semantics).
package missiles

import (
	"github.com/yohamta/donburi"
)

// MissileComponent is the ECS component wrapping the Missile payload; the
// collection stores one component per entity.
var MissileComponent = donburi.NewComponentType[Missile]()

// MissileCollection is the donburi missile container: entities carry the
// Missile payload; order is the insertion-ordered id list.
type MissileCollection struct {
	world donburi.World
	// order holds live entity references in std::list iteration order.
	order []donburi.Entity
}

// NewMissileCollection creates a collection over a fresh donburi world.
func NewMissileCollection() *MissileCollection {
	return &MissileCollection{world: donburi.NewWorld()}
}

// AddMissile appends the payload at the tail, mirroring AddMissile's
// emplace_back: the missile joins after any marked-for-deletion rows.
func (c *MissileCollection) AddMissile(m Missile) donburi.Entity {
	e := c.world.Create(MissileComponent)
	donburi.SetValue(c.world.Entry(e), MissileComponent, m)
	c.order = append(c.order, e)

	return e
}

// Mark flags the missile for deletion; it stays in the
// iteration order until RemoveMarked, matching _miDelFlag semantics.
func (c *MissileCollection) Mark(e donburi.Entity) {
	m := MissileComponent.Get(c.world.Entry(e))
	m.DelFlag = true
	MissileComponent.SetValue(c.world.Entry(e), *m)
}

// Marked reports whether the missile is flagged for deletion.
func (c *MissileCollection) Marked(e donburi.Entity) bool {
	return MissileComponent.Get(c.world.Entry(e)).DelFlag
}

// RemoveMarked drops the flagged missiles and keeps survivors in their
// relative order, mirroring list::remove_if.
func (c *MissileCollection) RemoveMarked() {
	kept := c.order[:0]
	for _, e := range c.order {
		if MissileComponent.Get(c.world.Entry(e)).DelFlag {
			c.world.Remove(e)
			continue
		}

		kept = append(kept, e)
	}

	c.order = kept
}

// Missiles returns the payloads in iteration order.
func (c *MissileCollection) Missiles() []Missile {
	out := make([]Missile, len(c.order))
	for i, e := range c.order {
		out[i] = MissileComponent.GetValue(c.world.Entry(e))
	}

	return out
}

// Entities returns the live entities in iteration order.
func (c *MissileCollection) Entities() []donburi.Entity {
	out := make([]donburi.Entity, len(c.order))
	copy(out, c.order)

	return out
}

// Source returns the payload for an entity of this collection.
func (c *MissileCollection) Source(e donburi.Entity) int {
	return MissileComponent.GetValue(c.world.Entry(e)).Source
}

// Len reports the number of missiles in iteration order (including marked
// ones, like std::list::size before remove_if).
func (c *MissileCollection) Len() int { return len(c.order) }
