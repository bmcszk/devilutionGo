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
	// The add-only cursor rows: their corpus process columns are empty
	// (misdat.tsv:40, 45, 46, 54, 60) because the spell direction only arms a
	// cursor, so the process side stays an explicit no-op.
	Identify: {
		Add: func(m *Missile, p *AddParameter) {
			AddIdentify(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	ItemRepair: {
		Add: func(m *Missile, p *AddParameter) {
			AddItemRepair(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	StaffRecharge: {
		Add: func(m *Missile, p *AddParameter) {
			AddStaffRecharge(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	HealOther: {
		Add: func(m *Missile, p *AddParameter) {
			AddHealOther(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	Resurrect: {
		Add: func(m *Missile, p *AddParameter) {
			AddResurrect(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	TrapDisarm: {
		Add: func(m *Missile, p *AddParameter) {
			AddTrapDisarm(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	Telekinesis: {
		Add: func(m *Missile, p *AddParameter) {
			AddTelekinesis(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	// The corpus rows with no add and no process body (misdat.tsv lines 16-19,
	// 28-29, 31-32, 34-35, 50-51 are empty in both columns): upstream has no
	// behaviour for them, so both sides are explicit no-ops. They stay listed
	// here instead of falling through to the generic stub so a later row that
	// gains a body has an obvious place to land.
	ChainBall: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	BloodHit: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	BoneHit: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	MetalHit: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	DoomSerpents: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	FireOnly: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	BloodRitual: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	Invisibility: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	Etherealize: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	Spurt: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	FireMan: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	Krull: {
		Add:     func(*Missile, *AddParameter) {},
		Process: func(*Missile) {},
	},
	// misdat row Healing: addFn AddHealing; the corpus process column is empty
	// (misdat.tsv:37) because the heal resolves in the add body.
	Healing: {
		Add: func(m *Missile, p *AddParameter) {
			AddHealing(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	// misdat row ManaShield: addFn AddManaShield; the corpus process column is
	// empty (misdat.tsv:13) because the spell is resolved in the add body, so
	// the process side stays a no-op.
	ManaShield: {
		Add: func(m *Missile, p *AddParameter) {
			AddManaShield(registryWorld, m, p)
		},
		Process: func(*Missile) {},
	},
	// misdat row ChainLightning: addFn AddChainLightning, processFn
	// ProcessChainLightning.
	ChainLightning: {
		Add: func(m *Missile, p *AddParameter) {
			AddChainLightning(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessChainLightning(registryWorld, m)
		},
	},
	// misdat row LightningControl: addFn AddLightningControl, processFn
	// ProcessLightningControl.
	LightningControl: {
		Add: func(m *Missile, p *AddParameter) {
			AddLightningControl(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessLightningControl(registryWorld, m)
		},
	},
	// misdat row ThinLightningControl: the corpus names the same addFn and
	// processFn as LightningControl (misdat.tsv:22), so it shares the bodies.
	ThinLightningControl: {
		Add: func(m *Missile, p *AddParameter) {
			AddLightningControl(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessLightningControl(registryWorld, m)
		},
	},
	// misdat row InfernoControl: addFn AddInfernoControl, processFn
	// ProcessInfernoControl.
	InfernoControl: {
		Add: func(m *Missile, p *AddParameter) {
			AddInfernoControl(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessInfernoControl(registryWorld, m)
		},
	},
	// misdat row WeaponExplosion: addFn AddWeaponExplosion, processFn
	// ProcessWeaponExplosion.
	WeaponExplosion: {
		Add: func(m *Missile, p *AddParameter) {
			AddWeaponExplosion(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessWeaponExplosion(registryWorld, m)
		},
	},
	// misdat row Inferno: addFn AddInferno, processFn ProcessInferno.
	Inferno: {
		Add: func(m *Missile, p *AddParameter) {
			AddInferno(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessInferno(registryWorld, m)
		},
	},
	// misdat row AcidPuddle: addFn AddAcidPuddle, processFn ProcessAcidPuddle.
	AcidPuddle: {
		Add: func(m *Missile, p *AddParameter) {
			AddAcidPuddle(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessAcidPuddle(registryWorld, m)
		},
	},
	// misdat row RedPortal: addFn AddRedPortal, processFn ProcessRedPortal.
	RedPortal: {
		Add: func(m *Missile, p *AddParameter) {
			AddRedPortal(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessRedPortal(registryWorld, m)
		},
	},
	// misdat row Nova: addFn AddNova, processFn ProcessNova.
	Nova: {
		Add: func(m *Missile, p *AddParameter) {
			AddNova(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessNova(registryWorld, m)
		},
	},
	// misdat row AcidSplat: addFn AddMissileExplosion (shared with the
	// explosion rows), processFn ProcessAcidSplate.
	AcidSplat: {
		Add: func(m *Missile, p *AddParameter) {
			AddMissileExplosion(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessAcidSplate(registryWorld, m)
		},
	},
	// misdat row Infravision: addFn AddInfravision, processFn
	// ProcessInfravision.
	Infravision: {
		Add: func(m *Missile, p *AddParameter) {
			AddInfravision(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessInfravision(registryWorld, m)
		},
	},
	// misdat row ResurrectBeam: addFn AddResurrectBeam, processFn
	// ProcessResurrectBeam.
	ResurrectBeam: {
		Add: func(m *Missile, p *AddParameter) {
			AddResurrectBeam(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessResurrectBeam(registryWorld, m)
		},
	},
	// misdat row FlashBottom: addFn AddFlashBottom, processFn
	// ProcessFlashBottom.
	FlashBottom: {
		Add: func(m *Missile, p *AddParameter) {
			AddFlashBottom(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessFlashBottom(registryWorld, m)
		},
	},
	// misdat row FlashTop: addFn AddFlashTop, processFn ProcessFlashTop.
	FlashTop: {
		Add: func(m *Missile, p *AddParameter) {
			AddFlashTop(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessFlashTop(registryWorld, m)
		},
	},
	// misdat row FlameWave: addFn AddFlameWave, processFn ProcessFlameWave.
	FlameWave: {
		Add: func(m *Missile, p *AddParameter) {
			AddFlameWave(registryWorld, m, p)
		},
		Process: func(m *Missile) {
			ProcessFlameWave(registryWorld, m)
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
