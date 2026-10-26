// Package missiles: the missile id to behavior dispatch registry.
package missiles

// behavior is the per-MissileID function pair from one misdat.tsv row.
type behavior struct {
	Add     behaviorFn
	Process behaviorFn
}

// behaviorFn is the Add*/Process* callback signature.
type behaviorFn func(m *Missile)

// registryWorld is the shared HitWorld backing the wired dispatches until
// the engine owns one world per game.
var registryWorld = &HitWorld{ProcessCallbacks: NoopProcessCallbacks()}

// Unported bodies stay no-op stubs until their slice lands.
var behaviors = map[MissileID]behavior{
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

func stub(_ MissileID) behavior {
	return behavior{
		Add:     func(*Missile) {},
		Process: func(*Missile) {},
	}
}

func init() {
	for id := Arrow; id <= LastDiablo; id++ {
		if _, ok := behaviors[id]; !ok {
			behaviors[id] = stub(id)
		}
	}
}

// RegistryCount reports how many MissileIDs have a behavior entry.
func RegistryCount() int {
	return len(behaviors)
}

// MissingFromRegistry returns the ids without a behavior entry.
func MissingFromRegistry(ids []MissileID) []MissileID {
	var missing []MissileID
	for _, id := range ids {
		if _, ok := behaviors[id]; !ok {
			missing = append(missing, id)
		}
	}
	return missing
}

// BehaviorFor returns the registered behavior for a MissileID.
func BehaviorFor(id MissileID) (behavior, bool) {
	b, ok := behaviors[id]
	return b, ok
}
