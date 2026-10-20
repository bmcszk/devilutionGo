// Registry of all base misdat addFn/processFn names. Full behaviour ports
// land with devilGo-mwlm; each entry cites its upstream definition. A stub
// registration keeps the data↔behaviour name-set contract checkable now
// (completeness test) without inventing behaviour.
package misdat

// BaseAddNames is the 50-name set referenced by base misdat.tsv (measured;
// 12 control rows carry an empty addFn = nullptr upstream).
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

// HFOnlyAddNames is the hf-superset addFn delta (23 names; 73 hf named − 50 base).
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

// HFOnlyProcessNames is the hf-superset processFn delta (9 names; 45 hf named − 36 base).
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

// stubFn stands in for a behavior port that has not landed yet.
func stubFn(_ World, _ *MissileRef) {}

// HFAddRegistry holds implementations for hf-only addFn names.
var HFAddRegistry = func() map[string]AddFn {
	m := map[string]AddFn{
		"AddOpenNest": AddOpenNest, // missiles.cpp:1238 (implemented, worked example)
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
	for _, name := range BaseAddNames {
		if _, ok := AddRegistry[name]; !ok {
			AddRegistry[name] = stubFn
		}
	}
	for _, name := range BaseProcessNames {
		if _, ok := ProcessRegistry[name]; !ok {
			ProcessRegistry[name] = stubFn
		}
	}
	for _, name := range HFOnlyProcessNames {
		if _, ok := ProcessRegistry[name]; !ok {
			ProcessRegistry[name] = stubFn
		}
	}
}
