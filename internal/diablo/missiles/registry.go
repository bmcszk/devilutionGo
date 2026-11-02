// Package missiles: the missile id to behavior dispatch registry.
package missiles

// Every misdat row dispatches a ported body: the last stub, Rhino (20), landed
// with process_rhino.go. The stub() fallback in init stays as the guard for a
// row added upstream before its body is ported.

// A row without an entry here still lands on the no-op stub.
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
	// misdat row Rhino: addFn AddRhino, processFn ProcessRhino.
	Rhino: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRhino(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessRhino(w, m)
		},
	},
	// misdat row StoneCurse: addFn AddStoneCurse, processFn ProcessStoneCurse.
	StoneCurse: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddStoneCurse(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessStoneCurse(w, m)
		},
	},
	// misdat row TownPortal: addFn AddTownPortal, processFn ProcessTownPortal.
	TownPortal: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddTownPortal(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessTownPortal(w, m)
		},
	},
	// misdat row Phasing: addFn AddPhasing, processFn ProcessTeleport.
	Phasing: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddPhasing(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessTeleport(w, m)
		},
	},
	// misdat row Teleport: addFn AddTeleport, processFn ProcessTeleport.
	Teleport: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddTeleport(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessTeleport(w, m)
		},
	},
	// misdat row Golem: addFn AddGolem, empty process column (misdat.tsv:35).
	Golem: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddGolem(w, m, p)
		},
		Process: func(*HitWorld, *Missile) {},
	},
	// misdat row FireWallControl: addFn AddWallControl, processFn ProcessWallControl.
	FireWallControl: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddWallControl(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessWallControl(w, m)
		},
	},
	// misdat row FlameWaveControl: addFn AddFlameWaveControl, processFn ProcessFlameWaveControl.
	FlameWaveControl: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddFlameWaveControl(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessFlameWaveControl(w, m)
		},
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

// hellfireBehaviors holds the Hellfire-region rows (hf-misdat.tsv data rows
// 68-107, missile ids 68+). They live apart from behaviors because the base
// corpus ledger gates RegistryCount against misdat.tsv's row count.
var hellfireBehaviors = map[MissileID]Behavior{
	// hf-misdat.tsv:68 Mana — addFn AddMana; the row has no processFn.
	Mana: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddMana(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:69 Magi — addFn AddMagi; the row has no processFn.
	Magi: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddMagi(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:79 Reflect — addFn AddReflect; the row has no processFn.
	Reflect: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddReflect(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:83 StealMana — addFn AddStealMana; the row has no
	// processFn.
	StealMana: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddStealMana(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:73 SpectralArrow — addFn AddSpectralArrow; processFn
	// ProcessSpectralArrow is not ported yet.
	SpectralArrow: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddSpectralArrow(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:74 FireballBow — addFn AddImmolation; processFn
	// ProcessFireball is ported.
	FireballBow: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddImmolation(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessFireball(w, m)
		},
	},
	// hf-misdat.tsv:75 LightningBow — addFn AddLightningBow; processFn
	// ProcessLightningBow is not ported yet.
	LightningBow: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddLightningBow(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:76 ChargedBoltBow — addFn AddChargedBoltBow; processFn
	// ProcessChargedBolt is ported.
	ChargedBoltBow: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddChargedBoltBow(w, m, p)
		},
		Process: func(w *HitWorld, m *Missile) {
			ProcessChargedBolt(w, m)
		},
	},
	// hf-misdat.tsv:89 RuneOfFire — addFn AddRuneOfFire; processFn ProcessRune
	// is not ported yet.
	RuneOfFire: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRuneOfFire(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:90 RuneOfLight — addFn AddRuneOfLight; processFn
	// ProcessRune is not ported yet.
	RuneOfLight: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRuneOfLight(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:92 RuneOfImmolation — addFn AddRuneOfImmolation; processFn
	// ProcessRune is not ported yet.
	RuneOfImmolation: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRuneOfImmolation(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:93 RuneOfStone — addFn AddRuneOfStone; processFn ProcessRune
	// is not ported yet.
	RuneOfStone: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRuneOfStone(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:94 BigExplosion — addFn AddBigExplosion; processFn
	// ProcessBigExplosion is not ported yet.
	BigExplosion: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddBigExplosion(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:81 RingOfFire — addFn AddRingOfFire; processFn
	// ProcessRingOfFire is not ported yet.
	RingOfFire: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRingOfFire(w, m, p)
		},
		Process: unportedProcess,
	},
	// hf-misdat.tsv:91 RuneOfNova — addFn AddRuneOfNova; processFn ProcessRune
	// is not ported yet.
	RuneOfNova: {
		Add: func(w *HitWorld, m *Missile, p *AddParameter) {
			AddRuneOfNova(w, m, p)
		},
		Process: unportedProcess,
	},
}

// Behavior is the per-MissileID function pair from one misdat.tsv row.
type Behavior struct {
	Add     addFn
	Process behaviorFn
	// Stub marks the no-op fallback init installs for a row without a ported
	// body, so a ledger test can tell a ported row from a stubbed one.
	Stub bool
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

// BehaviorFor returns the registered behavior for a MissileID, base corpus
// rows first and the Hellfire-region rows second.
func BehaviorFor(id MissileID) (Behavior, bool) {
	if b, ok := behaviors[id]; ok {
		return b, true
	}

	b, ok := hellfireBehaviors[id]

	return b, ok
}

func stub(_ MissileID) Behavior {
	return Behavior{
		Add:     func(*HitWorld, *Missile, *AddParameter) {},
		Process: func(*HitWorld, *Missile) {},
		Stub:    true,
	}
}

// unportedProcess is the placeholder for a Hellfire row whose upstream
// processFn has no port yet: the row's add body lands in slice 3c, the
// process body in the following slice.
func unportedProcess(*HitWorld, *Missile) {}
