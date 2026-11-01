// Package missiles: the missile id to behavior dispatch registry.
package missiles

// The 8 rows still falling through to the no-op stub, each blocked on a
// subsystem this package does not have yet — named with the upstream call that
// needs it so the next slice starts from the blocker, not from the row name:
//
//   - Phasing (3), Teleport (26), TownPortal (10): all three need PosOkPlayer
//     (monster.cpp:880-960) — the player/tile occupancy predicate this package
//     has no source for — plus UpdateVileMissPos, the player trans/light/vision
//     swap and the destination search over TileContainsMissile/CheckIfTrig
//     (cpp:2900-2975, 3900-3960).
//   - Rhino (20): monster Charge mode plus MissToMonst, which writes the
//     monster position and occupied tile back (cpp:3690-3730).
//   - StoneCurse (30): Monster::petrify, the Petrified mode and AddCorpse
//     (cpp:2497-2540, 3585-3620).
//   - Golem (33): monster creation from the add body (cpp:2810-2850).
//   - FireWallControl (38), FlameWaveControl (41): the
//     CanPlaceWall/PlaceWall/TryGrowWall/MoveWallToNextPosition wall subsystem
//     (cpp:3185-3260).

// Unported bodies stay no-op stubs until their slice lands.
var behaviors = map[MissileID]Behavior{
	Arrow: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddArrow(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessArrow(w, m)
		},
	},
	Firebolt: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddFirebolt(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessGenericProjectile(w, m)
		},
	},
	MagmaBall: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddMagmaBall(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessGenericProjectile(w, m)
		},
	},
	BloodStar: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddGenericMagicMissile(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessGenericProjectile(w, m)
		},
	},
	Acid: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddAcid(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessGenericProjectile(w, m)
		},
	},
	// misdat row NovaBall: addFn AddNovaBall, processFn ProcessNovaBall.
	NovaBall: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddNovaBall(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessNovaBall(w, m)
		},
	},
	// misdat row MagmaBallExplosion: addFn AddMissileExplosion, processFn
	// ProcessMissileExplosion.
	MagmaBallExplosion: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddMissileExplosion(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessMissileExplosion(w, m)
		},
	},
	// misdat row Fireball: addFn AddFireball, processFn ProcessFireball.
	Fireball: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddFireball(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessFireball(w, m)
		},
	},
	// misdat rows ChargedBolt/HolyBolt: addFn AddChargedBolt / AddHolyBolt,
	// processFn ProcessChargedBolt / ProcessHolyBolt.
	ChargedBolt: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddChargedBolt(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessChargedBolt(w, m)
		},
	},
	HolyBolt: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddHolyBolt(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessHolyBolt(w, m)
		},
	},
	// misdat rows Lightning/ThinLightning: addFn AddLightning, processFn
	// ProcessLightning (both rows share the same pair).
	Lightning: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddLightning(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessLightning(w, m)
		},
	},
	ThinLightning: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddLightning(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessLightning(w, m)
		},
	},
	// misdat row FireWall: addFn AddFireWall, processFn ProcessFireWall.
	FireWall: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddFireWall(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessFireWall(w, m)
		},
	},
	// The add-only cursor rows: their corpus process columns are empty
	// (misdat.tsv:40, 45, 46, 54, 60) because the spell direction only arms a
	// cursor, so the process side stays an explicit no-op.
	Identify: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddIdentify(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	ItemRepair: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddItemRepair(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	StaffRecharge: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddStaffRecharge(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	HealOther: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddHealOther(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	Resurrect: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddResurrect(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	TrapDisarm: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddTrapDisarm(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	Telekinesis: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddTelekinesis(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	// The corpus rows with no add and no process body (misdat.tsv lines 16-19,
	// 28-29, 31-32, 34-35, 50-51 are empty in both columns): upstream has no
	// behaviour for them, so both sides are explicit no-ops. They stay listed
	// here instead of falling through to the generic stub so a later row that
	// gains a body has an obvious place to land.
	ChainBall: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	BloodHit: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	BoneHit: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	MetalHit: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	DoomSerpents: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	FireOnly: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	BloodRitual: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	Invisibility: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	Etherealize: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	Spurt: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	FireMan: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	Krull: {
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	},
	// misdat row BoneSpirit: addFn AddBoneSpirit, processFn ProcessBoneSpirit.
	BoneSpirit: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddBoneSpirit(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessBoneSpirit(w, m)
		},
	},
	// misdat row Elemental: addFn AddElemental, processFn ProcessElemental.
	Elemental: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddElemental(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessElemental(w, m)
		},
	},
	// misdat row Apocalypse: addFn AddApocalypse, processFn ProcessApocalypse.
	Apocalypse: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddApocalypse(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessApocalypse(w, m)
		},
	},
	// misdat row DiabloApocalypse: addFn AddDiabloApocalypse; the corpus
	// process column is empty (misdat.tsv:67) because the body only spawns the
	// per-player booms and drops itself.
	DiabloApocalypse: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddDiabloApocalypse(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	// misdat row Guardian: addFn AddGuardian, processFn ProcessGuardian.
	Guardian: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddGuardian(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessGuardian(w, m)
		},
	},
	// misdat row Rage: addFn AddRage, processFn ProcessRage.
	Rage: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRage(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessRage(w, m)
		},
	},
	// misdat row Healing: addFn AddHealing; the corpus process column is empty
	// (misdat.tsv:37) because the heal resolves in the add body.
	Healing: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddHealing(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	// misdat row ManaShield: addFn AddManaShield; the corpus process column is
	// empty (misdat.tsv:13) because the spell is resolved in the add body, so
	// the process side stays a no-op.
	ManaShield: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddManaShield(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	// misdat row ChainLightning: addFn AddChainLightning, processFn
	// ProcessChainLightning.
	ChainLightning: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddChainLightning(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessChainLightning(w, m)
		},
	},
	// misdat row LightningControl: addFn AddLightningControl, processFn
	// ProcessLightningControl.
	LightningControl: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddLightningControl(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessLightningControl(w, m)
		},
	},
	// misdat row ThinLightningControl: the corpus names the same addFn and
	// processFn as LightningControl (misdat.tsv:22), so it shares the bodies.
	ThinLightningControl: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddLightningControl(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessLightningControl(w, m)
		},
	},
	// misdat row InfernoControl: addFn AddInfernoControl, processFn
	// ProcessInfernoControl.
	InfernoControl: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddInfernoControl(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessInfernoControl(w, m)
		},
	},
	// misdat row WeaponExplosion: addFn AddWeaponExplosion, processFn
	// ProcessWeaponExplosion.
	WeaponExplosion: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddWeaponExplosion(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessWeaponExplosion(w, m)
		},
	},
	// misdat row Inferno: addFn AddInferno, processFn ProcessInferno.
	Inferno: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddInferno(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessInferno(w, m)
		},
	},
	// misdat row AcidPuddle: addFn AddAcidPuddle, processFn ProcessAcidPuddle.
	AcidPuddle: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddAcidPuddle(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessAcidPuddle(w, m)
		},
	},
	// misdat row RedPortal: addFn AddRedPortal, processFn ProcessRedPortal.
	RedPortal: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRedPortal(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessRedPortal(w, m)
		},
	},
	// misdat row Nova: addFn AddNova, processFn ProcessNova.
	Nova: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddNova(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessNova(w, m)
		},
	},
	// misdat row AcidSplat: addFn AddMissileExplosion (shared with the
	// explosion rows), processFn ProcessAcidSplate.
	AcidSplat: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddMissileExplosion(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessAcidSplate(w, m)
		},
	},
	// misdat row Infravision: addFn AddInfravision, processFn
	// ProcessInfravision.
	Infravision: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddInfravision(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessInfravision(w, m)
		},
	},
	// misdat row ResurrectBeam: addFn AddResurrectBeam, processFn
	// ProcessResurrectBeam.
	ResurrectBeam: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddResurrectBeam(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessResurrectBeam(w, m)
		},
	},
	// misdat row FlashBottom: addFn AddFlashBottom, processFn
	// ProcessFlashBottom.
	FlashBottom: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddFlashBottom(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessFlashBottom(w, m)
		},
	},
	// misdat row FlashTop: addFn AddFlashTop, processFn ProcessFlashTop.
	FlashTop: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddFlashTop(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessFlashTop(w, m)
		},
	},
	// misdat row FlameWave: addFn AddFlameWave, processFn ProcessFlameWave.
	FlameWave: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddFlameWave(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessFlameWave(w, m)
		},
	},
	// misdat rows ApocalypseBoom/DiabloApocalypseBoom: addFn
	// AddApocalypseBoom, processFn ProcessApocalypseBoom.
	ApocalypseBoom: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddApocalypseBoom(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessApocalypseBoom(w, m)
		},
	},
	DiabloApocalypseBoom: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddApocalypseBoom(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessApocalypseBoom(w, m)
		},
	},
	// misdat row BloodStarExplosion: addFn AddMissileExplosion, processFn
	// ProcessMissileExplosion (same pair as MagmaBallExplosion).
	BloodStarExplosion: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddMissileExplosion(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessMissileExplosion(w, m)
		},
	},
	// misdat rows FireArrow/LightningArrow: addFn AddElementalArrow,
	// processFn ProcessElementalArrow.
	FireArrow: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddElementalArrow(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessElementalArrow(w, m)
		},
	},
	LightningArrow: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddElementalArrow(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessElementalArrow(w, m)
		},
	},
}

// Behavior is the per-MissileID function pair from one misdat.tsv row.
type Behavior struct {
	Add     addFn
	Process behaviorFn
}

// addFn is the Add* callback signature: MissileData::AddFn
// (tables/misdat.h:167) with the AddMissile parameter pack. The world arrives
// from the call site so a wired row writes into the world it was driven with
// instead of a package-level one.
type addFn func(w *HitWorld, m *Missile, p *AddParameter)

// behaviorFn is the Process* callback signature (tables/misdat.h:168).
type behaviorFn func(w *HitWorld, m *Missile)

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
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
	}
}
