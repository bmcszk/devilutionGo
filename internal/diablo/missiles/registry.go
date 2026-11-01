// Package missiles: the missile id to behavior dispatch registry.
package missiles

import "github.com/bmcszk/devilutionGo/internal/game/world"

// registryWorld is the shared HitWorld backing the wired dispatches until
// the engine owns one world per game.
var registryWorld = &HitWorld{
	Grid:             world.NewGrid(),
	HitCallbacks:     NoopHitCallbacks(),
	ProcessCallbacks: NoopProcessCallbacks(),
	AddSource:        NoopAddSource(),
}

// Unported bodies stay no-op stubs until their slice lands.
var behaviors = map[MissileID]Behavior{
	Arrow: {
		Add: func(m *Missile, p *AddParameter) {
			AddArrow(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessArrow(registryWorld, m)
		},
	},
	Firebolt: {
		Add: func(m *Missile, p *AddParameter) {
			AddFirebolt(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	MagmaBall: {
		Add: func(m *Missile, p *AddParameter) {
			AddMagmaBall(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	BloodStar: {
		Add: func(m *Missile, p *AddParameter) {
			AddGenericMagicMissile(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	Acid: {
		Add: func(m *Missile, p *AddParameter) {
			AddAcid(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessGenericProjectile(registryWorld, m)
		},
	},
	// misdat row NovaBall: addFn AddNovaBall, processFn ProcessNovaBall.
	NovaBall: {
		Add: func(m *Missile, p *AddParameter) {
			AddNovaBall(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessNovaBall(registryWorld, m)
		},
	},
	// misdat row MagmaBallExplosion: addFn AddMissileExplosion, processFn
	// ProcessMissileExplosion.
	MagmaBallExplosion: {
		Add: func(m *Missile, p *AddParameter) {
			AddMissileExplosion(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessMissileExplosion(registryWorld, m)
		},
	},
	// misdat row Fireball: addFn AddFireball, processFn ProcessFireball.
	Fireball: {
		Add: func(m *Missile, p *AddParameter) {
			AddFireball(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessFireball(registryWorld, m)
		},
	},
	// misdat rows ChargedBolt/HolyBolt: addFn AddChargedBolt / AddHolyBolt,
	// processFn ProcessChargedBolt / ProcessHolyBolt.
	ChargedBolt: {
		Add: func(m *Missile, p *AddParameter) {
			AddChargedBolt(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessChargedBolt(registryWorld, m)
		},
	},
	HolyBolt: {
		Add: func(m *Missile, p *AddParameter) {
			AddHolyBolt(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessHolyBolt(registryWorld, m)
		},
	},
	// misdat rows Lightning/ThinLightning: addFn AddLightning, processFn
	// ProcessLightning (both rows share the same pair).
	Lightning: {
		Add: func(m *Missile, p *AddParameter) {
			AddLightning(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessLightning(registryWorld, m)
		},
	},
	ThinLightning: {
		Add: func(m *Missile, p *AddParameter) {
			AddLightning(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessLightning(registryWorld, m)
		},
	},
	// misdat row FireWall: addFn AddFireWall, processFn ProcessFireWall.
	FireWall: {
		Add: func(m *Missile, p *AddParameter) {
			AddFireWall(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessFireWall(registryWorld, m)
		},
	},
	// misdat rows ApocalypseBoom/DiabloApocalypseBoom: addFn
	// AddApocalypseBoom, processFn ProcessApocalypseBoom.
	ApocalypseBoom: {
		Add: func(m *Missile, p *AddParameter) {
			AddApocalypseBoom(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessApocalypseBoom(registryWorld, m)
		},
	},
	DiabloApocalypseBoom: {
		Add: func(m *Missile, p *AddParameter) {
			AddApocalypseBoom(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessApocalypseBoom(registryWorld, m)
		},
	},
	// misdat row BloodStarExplosion: addFn AddMissileExplosion, processFn
	// ProcessMissileExplosion (same pair as MagmaBallExplosion).
	BloodStarExplosion: {
		Add: func(m *Missile, p *AddParameter) {
			AddMissileExplosion(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessMissileExplosion(registryWorld, m)
		},
	},
	// misdat rows FireArrow/LightningArrow: addFn AddElementalArrow,
	// processFn ProcessElementalArrow.
	FireArrow: {
		Add: func(m *Missile, p *AddParameter) {
			AddElementalArrow(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessElementalArrow(registryWorld, m)
		},
	},
	LightningArrow: {
		Add: func(m *Missile, p *AddParameter) {
			AddElementalArrow(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessElementalArrow(registryWorld, m)
		},
	},
}

// Behavior is the per-MissileID function pair from one misdat.tsv row.
type Behavior struct {
	Add     addFn
	Process behaviorFn
}

// addFn is the Add* callback signature: MissileData::AddFn
// (tables/misdat.h:167) with the AddMissile parameter pack.
type addFn func(m *Missile, p *AddParameter)

// behaviorFn is the Process* callback signature (tables/misdat.h:168).
type behaviorFn func(m *Missile)

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
func BehaviorFor(id MissileID) (Behavior, bool) {
	b, ok := behaviors[id]
	return b, ok
}

func stub(_ MissileID) Behavior {
	return Behavior{
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	}
}
