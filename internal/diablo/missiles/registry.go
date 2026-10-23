// Package missiles: the missile id to behavior dispatch registry.
package missiles

// behavior is the per-missileID function pair from one misdat.tsv row.
type behavior struct {
	Add     behaviorFn
	Process behaviorFn
}

// behaviorFn is the Add*/Process* callback signature.
type behaviorFn func(m *Missile)

// Unported bodies stay no-op stubs until their slice lands.
var behaviors = map[missileID]behavior{
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

func stub(_ missileID) behavior {
	return behavior{
		Add:     func(*Missile) {},
		Process: func(*Missile) {},
	}
}

// registryWorld is the shared hitWorld backing the wired dispatches until
// the engine owns one world per game.
var registryWorld = &hitWorld{processCallbacks: noopProcessCallbacks()}

func init() {
	for id := Arrow; id <= LastDiablo; id++ {
		if _, ok := behaviors[id]; !ok {
			behaviors[id] = stub(id)
		}
	}
}

// registryCount reports how many MissileIDs have a behavior entry.
func registryCount() int {
	return len(behaviors)
}

// missingFromRegistry returns the ids without a behavior entry.
func missingFromRegistry(ids []missileID) []missileID {
	var missing []missileID
	for _, id := range ids {
		if _, ok := behaviors[id]; !ok {
			missing = append(missing, id)
		}
	}
	return missing
}

// behaviorFor returns the registered behavior for a missileID.
func behaviorFor(id missileID) (behavior, bool) {
	b, ok := behaviors[id]
	return b, ok
}
