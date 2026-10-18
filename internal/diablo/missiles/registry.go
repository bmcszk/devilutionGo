// Package missiles: the Go replacement for the missiledata[] function
// pointer table (Source/missiles.cpp:60-620). Dispatch mirrors
// ProcessMissiles (missiles.cpp:4210-4212): per live missile,
// GetMissileData(_mitype).processFn(missile).
package missiles

// Behavior is the per-MissileID function pair from one misdat.tsv row
// (MissileData addFn/processFn, Source/tables/misdat.h:166-196;
// missiledata[] rows missiles.cpp:60-620).
type Behavior struct {
	// Add ports the row's Add* function (invoked from AddMissile,
	// missiles.cpp:2828 missileData.addFn(missile, parameter)).
	Add BehaviorFn
	// Process ports the row's Process* function (invoked from
	// ProcessMissiles, missiles.cpp:4212 missileData.processFn(missile)).
	Process BehaviorFn
}

// BehaviorFn is the Add*/Process* signature: fn(Missile&) upstream.
type BehaviorFn func(m *Missile)

// behaviors maps MissileID -> Behavior. Full coverage of the base
// misdat.tsv name set (68 rows) with slice-1 stubs; real ported bodies
// land in later slices. Stub shape: empty op + TODO upstream cite (panic
// bodies are banned by contract).
var behaviors = map[MissileID]Behavior{
	Arrow: {
		Add: func(m *Missile) {
			// TODO(devilGo-mwlm slice 2+): port AddArrow
			// (Source/missiles.cpp:1770-1796) — UpdateMissileVelocity +
			// GetDirection16 frame + duration = 256.
			_ = m
		},
		Process: func(m *Missile) {
			// missiledata[] row 1: ProcessArrow (missiles.cpp:69-70);
			// the world seam lands with the ProcessMissiles bean.
			ProcessArrow(registryWorld, m)
		},
	},
	Firebolt: {
		Add: func(m *Missile) {
			// TODO(devilGo-mwlm slice 2+): port AddFirebolt
			// (Source/missiles.cpp:1858) — speed by caster, then
			// SetMissileVelocity (missiles.cpp:730-750).
			_ = m
		},
		Process: func(m *Missile) {
			// missiledata[] row 2: ProcessGenericProjectile
			// (missiles.cpp:65-66).
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	MagmaBall: {
		Add: func(m *Missile) {
			// TODO(devilGo-mwlm slice 2+): port AddMagmaBall
			// (Source/missiles.cpp:1978) — SetMissileVelocity +
			// duration by speed (row 22).
			_ = m
		},
		Process: func(m *Missile) {
			// missiledata[] row 23: ProcessGenericProjectile
			// (missiles.cpp:91-92).
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	BloodStar: {
		Add: func(m *Missile) {
			// TODO(devilGo-mwlm slice 2+): port AddBloodStar
			// (row 25) — SetMissileVelocity + AddLight 8.
			_ = m
		},
		Process: func(m *Missile) {
			// missiledata[] row 26: ProcessGenericProjectile
			// (missiles.cpp:107-108).
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	Acid: {
		Add: func(m *Missile) {
			// TODO(devilGo-mwlm slice 2+): port AddAcid (row 58) —
			// velocity + duration by distance.
			_ = m
		},
		Process: func(m *Missile) {
			// missiledata[] row 59: ProcessGenericProjectile
			// (missiles.cpp:329-330).
			ProcessGenericProjectile(registryWorld, m)
		},
	},
}

func stub(id MissileID) Behavior {
	return Behavior{
		Add: func(m *Missile) {
			// TODO(devilGo-mwlm later slice): port the Add* body for
			// this MissileID from Source/missiles.cpp addFn region.
			_ = id
			_ = m
		},
		Process: func(m *Missile) {
			// TODO(devilGo-mwlm later slice): port the Process* body
			// for this MissileID from Source/missiles.cpp processFn
			// region (missiledata[] missiles.cpp:60-620).
			_ = id
			_ = m
		},
	}
}

// registryWorld is the shared HitWorld backing the five wired registry
// dispatches until the engine owns one world per game (TODO
// devilGo-mwlm ProcessMissiles bean: per-game world instance; the
// behavior-map signature grows then too — upstream processFn reads the
// globals directly, missiles.cpp:4212).
var registryWorld = &HitWorld{}

func init() {
	for id := Arrow; id <= LastDiablo; id++ {
		if _, ok := behaviors[id]; !ok {
			behaviors[id] = stub(id)
		}
	}
}

// RegistryCount reports how many MissileIDs have a Behavior entry
// (missiledata[] row coverage; gate: == base misdat.tsv row count,
// asserted from the on-disk TSV at test runtime).
func RegistryCount() int {
	return len(behaviors)
}

// MissingFromRegistry returns the ids without a Behavior entry — the Go
// shape of upstream dispatching into an out-of-range MissilesData slot
// (GetMissileData, Source/tables/misdat.cpp:448-452).
func MissingFromRegistry(ids []MissileID) []MissileID {
	var missing []MissileID
	for _, id := range ids {
		if _, ok := behaviors[id]; !ok {
			missing = append(missing, id)
		}
	}
	return missing
}

// BehaviorFor returns the registered Behavior for a MissileID
// (GetMissileData port, Source/tables/misdat.cpp:448-452).
func BehaviorFor(id MissileID) (Behavior, bool) {
	b, ok := behaviors[id]
	return b, ok
}
