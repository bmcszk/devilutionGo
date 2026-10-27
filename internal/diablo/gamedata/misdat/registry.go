package misdat

// BaseAddNames are the 50 AddFn names referenced by base misdat.tsv.
var BaseAddNames = []string{
	"AddAcid",
	"AddAcidPuddle",
	"AddApocalypse",
	"AddApocalypseBoom",
	"AddArrow",
	"AddBoneSpirit",
	"AddChainLightning",
	"AddChargedBolt",
	"AddDiabloApocalypse",
	"AddElemental",
	"AddElementalArrow",
	"AddFireball",
	"AddFirebolt",
	"AddFireWall",
	"AddFlameWave",
	"AddFlameWaveControl",
	"AddFlashBottom",
	"AddFlashTop",
	"AddGenericMagicMissile",
	"AddGolem",
	"AddGuardian",
	"AddHealing",
	"AddHealOther",
	"AddHolyBolt",
	"AddIdentify",
	"AddInferno",
	"AddInfernoControl",
	"AddInfravision",
	"AddItemRepair",
	"AddLightning",
	"AddLightningControl",
	"AddMagmaBall",
	"AddManaShield",
	"AddMissileExplosion",
	"AddNova",
	"AddNovaBall",
	"AddPhasing",
	"AddRage",
	"AddRedPortal",
	"AddResurrect",
	"AddResurrectBeam",
	"AddRhino",
	"AddStaffRecharge",
	"AddStoneCurse",
	"AddTelekinesis",
	"AddTeleport",
	"AddTownPortal",
	"AddTrapDisarm",
	"AddWallControl",
	"AddWeaponExplosion",
}

// BaseProcessNames is the 36-name set referenced by base misdat.tsv (excl. empty).
var BaseProcessNames = []string{
	"ProcessAcidPuddle",
	"ProcessAcidSplate",
	"ProcessApocalypse",
	"ProcessApocalypseBoom",
	"ProcessArrow",
	"ProcessBoneSpirit",
	"ProcessChainLightning",
	"ProcessChargedBolt",
	"ProcessElemental",
	"ProcessElementalArrow",
	"ProcessFireball",
	"ProcessFireWall",
	"ProcessFlameWave",
	"ProcessFlameWaveControl",
	"ProcessFlashBottom",
	"ProcessFlashTop",
	"ProcessGenericProjectile",
	"ProcessGuardian",
	"ProcessHolyBolt",
	"ProcessInferno",
	"ProcessInfernoControl",
	"ProcessInfravision",
	"ProcessLightning",
	"ProcessLightningControl",
	"ProcessMissileExplosion",
	"ProcessNova",
	"ProcessNovaBall",
	"ProcessRage",
	"ProcessRedPortal",
	"ProcessResurrectBeam",
	"ProcessRhino",
	"ProcessStoneCurse",
	"ProcessTeleport",
	"ProcessTownPortal",
	"ProcessWallControl",
	"ProcessWeaponExplosion",
}

// HFOnlyAddNames is the hf-superset AddFn delta (23 names; 73 hf named − 50 base).
var HFOnlyAddNames = []string{
	"AddBerserk",
	"AddBigExplosion",
	"AddChargedBoltBow",
	"AddHorkSpawn",
	"AddImmolation",
	"AddJester",
	"AddLightningBow",
	"AddLightningWall",
	"AddMagi",
	"AddMana",
	"AddOpenNest",
	"AddReflect",
	"AddRingOfFire",
	"AddRuneOfFire",
	"AddRuneOfImmolation",
	"AddRuneOfLight",
	"AddRuneOfNova",
	"AddRuneOfStone",
	"AddSearch",
	"AddSpectralArrow",
	"AddStealMana",
	"AddStealPotions",
	"AddWarp",
}

// HFOnlyProcessNames is the hf-superset ProcessFn delta (9 names; 45 hf named − 36 base).
var HFOnlyProcessNames = []string{
	"ProcessBigExplosion",
	"ProcessHorkSpawn",
	"ProcessImmolation",
	"ProcessLightningBow",
	"ProcessLightningWall",
	"ProcessRingOfFire",
	"ProcessRune",
	"ProcessSearch",
	"ProcessSpectralArrow",
}

// HFAddRegistry holds implementations for hf-only AddFn names.
var HFAddRegistry = func() map[string]AddFn {
	m := map[string]AddFn{
		"AddOpenNest": AddOpenNest,
	}

	for _, name := range HFOnlyAddNames {
		if name == "AddOpenNest" {
			continue
		}

		m[name] = stubFn
	}

	return m
}()

func init() {
	registerMissingAdd(AddRegistry, BaseAddNames)
	registerMissingProcess(ProcessRegistry, BaseProcessNames)
	registerMissingProcess(ProcessRegistry, HFOnlyProcessNames)
}

// stubFn stands in for a behavior port that has not landed yet.
func stubFn(_ World, _ *MissileRef) {}

// registerMissingAdd populates an AddFn registry with stubFn for names not already present.
func registerMissingAdd(reg map[string]AddFn, names []string) {
	for _, name := range names {
		if _, ok := reg[name]; !ok {
			reg[name] = stubFn
		}
	}
}

// registerMissingProcess populates a ProcessFn registry with stubFn for names not already present.
func registerMissingProcess(reg map[string]ProcessFn, names []string) {
	for _, name := range names {
		if _, ok := reg[name]; !ok {
			reg[name] = stubFn
		}
	}
}
