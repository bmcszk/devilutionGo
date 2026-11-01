package itemgen

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/itemdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// maxUniqueRegen caps the two force-generate loops of TryRandomUniqueItem
// (items.cpp:3342-3346, 3398-3405). Upstream spins until the RNG walk happens
// to produce the wanted uid; the cap reports a bad seed or table as an error
// instead of hanging the caller.
const maxUniqueRegen = 1000

// errUniqueRegenLimit reports a seed or table the two regen loops cannot satisfy.
var errUniqueRegenLimit = fmt.Errorf("no uid match within %d regenerations", maxUniqueRegen)

// RandomUniqueRequest is the item state TryRandomUniqueItem reads
// (items.cpp:3312-3325).
type RandomUniqueRequest struct {
	CreateInfo int
	MiscID     string
	BaseItemID string // AllItemsList[idx].iItemId
	ItemIndex  int
	Seed       uint32
	Level      int // mLevel
	Uper       int
	OnlyGood   bool
	Pregen     bool
	Spells     []txtdata.SpellData
	Env        SpellEnv
	Hero       HeroStats
}

// RandomUniqueEnv is the game state and RNG TryRandomUniqueItem reads, injected
// because upstream reaches them through globals (UniqueItemFlags, gbIsMultiplayer).
type RandomUniqueEnv struct {
	Uniques     []itemdat.UniqueItemData
	Dropped     map[int]bool
	Multiplayer bool
	Rnd         func(int32) int32
	Discard     func(int)
	// NextSeed is the DiabloGenerator(item._iSeed) walk the regen loop uses
	// (DiabloGenerator::advanceRndSeed).
	NextSeed func() uint32
	Setup    func(SetupRequest) (SetupPlan, error)
}

// RandomUniqueOutcome is the TryRandomUniqueItem result.
type RandomUniqueOutcome struct {
	Skipped       bool
	UID           int  // chosen unique row, UniqueNone on the fallback
	Offset        int  // uidOffset handed to SetupAllItems
	ReverseCompat bool // the reverse-compatible regen path ran
	Fallback      bool // no uniques left: uper forced to 1
	Regenerated   int  // SetupAllItems reruns
	Plan          SetupPlan
}

// RandomUnique ports TryRandomUniqueItem (items.cpp:3312-3419): pick a unique
// that has not dropped yet, then force SetupAllItems to produce it — through
// the reverse-compatible regen loop when the unique is reachable that way, or
// through a uid offset otherwise.
func RandomUnique(env RandomUniqueEnv, req RandomUniqueRequest) (RandomUniqueOutcome, error) {
	if req.CreateInfo&CreateInfoUnique == 0 || req.MiscID == "UNIQUE" {
		return RandomUniqueOutcome{Skipped: true}, nil
	}

	env.Discard(1) // DiscardRandomValues(1) replays GetItemAttrs (:3320)
	blvl := GetItemBLevel(req.Level, req.MiscID, req.OnlyGood, req.Uper == 15, env.Rnd)
	valid := ValidUniques(env.Uniques, req.BaseItemID, blvl)

	uids := AvailableUniques(valid, env.Dropped, env.Multiplayer)
	if len(uids) == 0 {
		return env.forceNonUnique(req)
	}

	uid := PickUnique(uids, env.Rnd)
	offset, reverse, err := UniqueOffset(valid, uid, env.Uniques)
	if err != nil {
		return RandomUniqueOutcome{}, err
	}

	out := RandomUniqueOutcome{UID: uid, Offset: offset, ReverseCompat: reverse}
	if reverse {
		out.Plan, out.Regenerated, err = env.regenerateUntilUID(req, uid)
	} else {
		out.Plan, err = env.setupAtOffset(req, offset)
	}
	if err != nil {
		return RandomUniqueOutcome{}, err
	}

	return out, nil
}

// forceNonUnique ports the uids-empty branch (:3337-3346): better odds of not
// rolling a unique, then rerun until the roll misses.
func (env RandomUniqueEnv) forceNonUnique(req RandomUniqueRequest) (RandomUniqueOutcome, error) {
	level := req.Level
	if req.Uper == 15 {
		level += 4
	}
	out := RandomUniqueOutcome{UID: UniqueNone, Fallback: true}
	for range maxUniqueRegen {
		plan, err := env.Setup(buildSetupRequest(req, env.NextSeed(), level, 1, req.OnlyGood, 0))
		if err != nil {
			return RandomUniqueOutcome{}, err
		}
		out.Regenerated++
		out.Plan = plan
		if plan.Quality != QualityUnique {
			return out, nil
		}
	}

	return RandomUniqueOutcome{}, fmt.Errorf("force non-unique: %w", errUniqueRegenLimit)
}

// regenerateUntilUID ports the reverse-compatible branch (:3387-3406): onlygood
// is forced so the base item always reaches the unique's level.
func (env RandomUniqueEnv) regenerateUntilUID(req RandomUniqueRequest, uid int) (SetupPlan, int, error) {
	uper, targetLvl := uniqueReverseLevels(env.Uniques[uid].MinLevel)
	for i := range maxUniqueRegen {
		setupReq := buildSetupRequest(req, env.NextSeed(), targetLvl, uper, true, 0)
		plan, err := env.Setup(setupReq)
		if err != nil {
			return SetupPlan{}, 0, err
		}
		if plan.UID == uid {
			return plan, i + 1, nil
		}
	}

	return SetupPlan{}, 0, fmt.Errorf("uid %d: %w", uid, errUniqueRegenLimit)
}

// setupAtOffset ports the non-reverse-compatible branch (:3407-3416): one rerun
// with the computed offset, then the CF_UIDOFFSET tag.
func (env RandomUniqueEnv) setupAtOffset(req RandomUniqueRequest, offset int) (SetupPlan, error) {
	plan, err := env.Setup(buildSetupRequest(req, req.Seed, req.Level, req.Uper, req.OnlyGood, offset))
	if err != nil {
		return SetupPlan{}, fmt.Errorf("uid offset %d: %w", offset, err)
	}
	plan.Buff |= (offset << 1) & BuffUIDOffsetMask

	return plan, nil
}

func buildSetupRequest(
	req RandomUniqueRequest,
	seed uint32,
	level, uper int,
	onlygood bool,
	uidOffset int,
) SetupRequest {
	return SetupRequest{
		ItemIndex: req.ItemIndex,
		Seed:      seed,
		Level:     level,
		Uper:      uper,
		OnlyGood:  onlygood,
		Pregen:    req.Pregen,
		Hero:      req.Hero,
		UidOffset: uidOffset,
		Spells:    req.Spells,
		Env:       req.Env,
	}
}

// uniqueReverseLevels ports the target-level choice of :3390-3398: a level 0
// target would underflow, so those uniques take the uper-1 path instead.
func uniqueReverseLevels(minLevel int) (uper, targetLvl int) {
	if minLevel-4 > 0 {
		return 15, minLevel - 4
	}

	return 1, minLevel
}
