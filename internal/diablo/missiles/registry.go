// Package missiles: the missile id to behavior dispatch registry.
package missiles

// Behavior is the per-MissileID function pair from one misdat.tsv row.
type Behavior struct {
	Add     BehaviorFn
	Process BehaviorFn
}

// BehaviorFn is the Add*/Process* callback signature.
type BehaviorFn func(m *Missile)

// behaviors maps MissileID to Behavior; bodies that are not ported yet are
// no-op stubs.
var behaviors = map[MissileID]Behavior{
	Arrow: {
		Add: func(*Missile) {},
		Process: func(m *Missile) {
			ProcessArrow(registryWorld, m)
		},
	},
	Firebolt: {
		Add: func(*Missile) {},
		Process: func(m *Missile) {
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	MagmaBall: {
		Add: func(*Missile) {},
		Process: func(m *Missile) {
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	BloodStar: {
		Add: func(*Missile) {},
		Process: func(m *Missile) {
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	Acid: {
		Add: func(*Missile) {},
		Process: func(m *Missile) {
			ProcessGenericProjectile(registryWorld, m)
		},
	},
}

func stub(_ MissileID) Behavior {
	return Behavior{
		Add:     func(*Missile) {},
		Process: func(*Missile) {},
	}
}

// registryWorld is the shared HitWorld backing the wired dispatches until
// the engine owns one world per game.
var registryWorld = &HitWorld{ProcessCallbacks: NoopProcessCallbacks()}

func init() {
	for id := Arrow; id <= LastDiablo; id++ {
		if _, ok := behaviors[id]; !ok {
			behaviors[id] = stub(id)
		}
	}
}

// RegistryCount reports how many MissileIDs have a Behavior entry.
func RegistryCount() int {
	return len(behaviors)
}

// MissingFromRegistry returns the ids without a Behavior entry.
func MissingFromRegistry(ids []MissileID) []MissileID {
	var missing []MissileID
	for _, id := range ids {
		if _, ok := behaviors[id]; !ok {
			missing = append(missing, id)
		}
	}
	return missing
}

// BehaviorFor returns the registered Behavior for a MissileID.
func BehaviorFor(id MissileID) (Behavior, bool) {
	b, ok := behaviors[id]
	return b, ok
}
