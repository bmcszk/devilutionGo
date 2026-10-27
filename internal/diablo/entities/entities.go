// Package entities ports the upstream entity structs (player.h, monster.h,
// objects.h, items.h, missiles.h) as donburi components: plain data, one
// component per entity kind. Behavior stays in the owning packages; a
// component holds only state.
//
// Upstream layout note: this checkout has no structs.h — the structs live
// in player.h (Player), monster.h (Monster), objects.h (Object), items.h
// (Item) and missiles.h (Missile). Missile's data port already landed in
// internal/diablo/missiles.Missile; entities.Missile wraps that type so the
// ECS surface stays uniform without duplicating fields.
//
// Go struct sizes differ from the C++ ABI (no packed arrays of Item, GC
// pointers, Go string headers); the size-assertion test in size_test.go
// pins each Go size deliberately and documents the mapping — it is not a
// binary-parity claim.
package entities

import (
	"github.com/yohamta/donburi"
)

// Upstream array and enum-boundary constants, cited per use site.
const (
	// PlayerNameLength (player.h:42).
	PlayerNameLength = 32
	// NUM_INVLOC (player.h:80, inv_body_loc).
	NumInvLoc = 7
	// InventoryGridCells (player.h:38).
	InventoryGridCells = 40
	// MaxBeltItems (player.h:39).
	MaxBeltItems = 8
	// MaxPathLengthPlayer (engine/path.h:20).
	MaxPathLengthPlayer = 100
	// NumHotkeys (player.h:44).
	NumHotkeys = 12
	// Spell level array bound: _pSplLvl[64] (player.h).
	SpellLevels = 64
	// NUMLEVELS (levels/gendung_defs.hpp:11).
	NumLevels = 25
	// ItemNameLength (items.h:35).
	ItemNameLength = 64
	// MAXOBJECTS (objects.h:30).
	MaxObjects = 127
	// MaxMonsters (monster.h:38).
	MaxMonsters = 200
)

// The five entity components. Each wraps the ported struct below; the
// struct doc comment carries the single upstream citation.

// PlayerComponent is the ECS component for PlayerStruct (player.h:211).
var PlayerComponent = donburi.NewComponentType[Player]()

// MonsterComponent is the ECS component for MonsterStruct (monster.h:213).
var MonsterComponent = donburi.NewComponentType[Monster]()

// ObjectComponent is the ECS component for ObjectStruct (objects.h:32).
var ObjectComponent = donburi.NewComponentType[Object]()

// ItemComponent is the ECS component for ItemStruct (items.h:189).
var ItemComponent = donburi.NewComponentType[Item]()

// MissileComponent is the ECS component for MissileStruct (missiles.h:89).
// The field port lives in internal/diablo/missiles.Missile; this component
// references it instead of duplicating the 40+ fields.
var MissileComponent = donburi.NewComponentType[missilesMissile]()

// Point is engine.Point (point.hpp) — kept as a local alias type so the
// component field types are self-contained.
type Point = enginePoint
