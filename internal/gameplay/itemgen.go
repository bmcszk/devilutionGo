// Item generation entry points the bean contract names for this tree. The port
// itself lives in internal/diablo/gamedata/itemgen (AGENTS.md layer 3: Diablo
// item data and generation), so this package re-exports it the same way
// inventory.go re-exports internal/diablo/inventory — type aliases and function
// values only, no logic here.
package gameplay

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemgen"
)

// Item generation constants, in upstream value order.
const (
	// LevelNotGiven is SpawnUnique's std::nullopt level (items.cpp:3197).
	LevelNotGiven = itemgen.LevelNotGiven
	// DifficultyNormal is DIFF_NORMAL (items.cpp:3203).
	DifficultyNormal = itemgen.DifficultyNormal
	// ItemNone is IDI_NONE, the "no item" index (items.cpp:3475).
	ItemNone = itemgen.ItemNone
)

// Item generation entry points, re-exported from internal/diablo/gamedata/itemgen.
var (
	// SpawnItem is SpawnItem (items.cpp:3422-3481).
	SpawnItem = itemgen.SpawnItem
	// SpawnUnique is SpawnUnique (items.cpp:3195-3228).
	SpawnUnique = itemgen.SpawnUnique
	// SpawnMonsterItem is the SetupAllItems + TryRandomUniqueItem tail.
	SpawnMonsterItem = itemgen.SpawnMonsterItem
	// PlanDrop is the T_MASK/T_NODROP/T_UNIQ treasure decode.
	PlanDrop = itemgen.PlanDrop
	// SetupAllItems is SetupAllItems (items.cpp:3269-3312).
	SetupAllItems = itemgen.SetupAllItems
	// RandomUnique is TryRandomUniqueItem (items.cpp:3312-3419).
	RandomUnique = itemgen.RandomUnique
	// AvailableUniques is GetValidUniques (items.cpp:1138-1150).
	AvailableUniques = itemgen.AvailableUniques
	// PickUnique is the GenerateRnd draw over the valid uid list.
	PickUnique = itemgen.PickUnique
	// DamageWeapon is DamageWeapon (player.cpp:457-524).
	DamageWeapon = itemgen.DamageWeapon
	// ApplyUniqueItem is GetUniqueItem (items.cpp:1436-1493).
	ApplyUniqueItem = itemgen.ApplyUniqueItem
	// SaveItemPower is SaveItemPower (items.cpp:701-728).
	SaveItemPower = itemgen.SaveItemPower
	// CalcItemValue is CalcItemValue (items.cpp:600-610).
	CalcItemValue = itemgen.CalcItemValue
)

// Item generation types, aliased so both import paths name the same types.
type (
	// SpawnDeps is the game state the generation tail reads.
	SpawnDeps = itemgen.SpawnDeps
	// MonsterDrop is the monstdat row a spawn starts from.
	MonsterDrop = itemgen.MonsterDrop
	// BrainQuest is the Q_MUSHROOM state SpawnItem branches on.
	BrainQuest = itemgen.BrainQuest
	// DropEnv is the game-mode state the drop decode reads.
	DropEnv = itemgen.DropEnv
	// DropPlan is the decoded treasure plan.
	DropPlan = itemgen.DropPlan
	// MonsterItem is one spawned item.
	MonsterItem = itemgen.MonsterItem
	// UniqueSpawn is the SpawnUnique outcome.
	UniqueSpawn = itemgen.UniqueSpawn
	// SetupRequest is one SetupAllItems call.
	SetupRequest = itemgen.SetupRequest
	// SetupPlan is the item state SetupAllItems produces.
	SetupPlan = itemgen.SetupPlan
	// HeroStats is the player stat slice the conversion powers read.
	HeroStats = itemgen.HeroStats
	// AffixBonus is one applied power (items.cpp:701-728).
	AffixBonus = itemgen.AffixBonus
	// Affix is one prefix/suffix row.
	Affix = itemgen.Affix
	// WeaponState is the equipment pair DamageWeapon inspects.
	WeaponState = itemgen.WeaponState
	// WearResult is the ordered action list of one DamageWeapon pass.
	WearResult = itemgen.WearResult
	// WearAction is one equipment mutation (player.cpp:457-524).
	WearAction = itemgen.WearAction
)
