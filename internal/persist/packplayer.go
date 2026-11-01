package persist

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
	"github.com/bmcszk/devilutionGo/internal/engine"
)

// PackPlayer minifies a player into the hero-file record
// (Source/pack.cpp:152-203). charLevel is player.getCharacterLevel(): the
// character level lives on the player-construction layer, which
// entities.Player does not yet carry.
func PackPlayer(player entities.Player, charLevel uint8, env PackEnv) PlayerPack {
	var packed PlayerPack

	packPlayerHead(&packed, player, env)
	packPlayerStats(&packed, player, charLevel)
	packPlayerSlots(&packed, player, env)

	return packed
}

// UnpackPlayer expands a hero-file record into player fields
// (Source/pack.cpp:350-420). The record-level half is ported here; the
// clamps, InitPlayer and the GetSpellBookLevel gating stay on the player and
// itemgen layers, so the caller re-applies them.
func UnpackPlayer(packed PlayerPack, env PackEnv) entities.Player {
	var player entities.Player

	player.MaxHPBase = int(packed.MaxHP)
	player.HPBase = int(packed.HPBase)
	player.MaxHP = player.MaxHPBase
	player.HitPoints = player.HPBase
	player.Position.Tile = engine.Point{X: int(packed.Px), Y: int(packed.Py)}
	player.Position.Future = player.Position.Tile
	player.PlrLevel = packed.PlrLevel
	player.Class = packed.Class
	copy(player.Name[:], packed.Name[:])
	player.BaseStr = int(packed.BaseStr)
	player.BaseMag = int(packed.BaseMag)
	player.BaseDex = int(packed.BaseDex)
	player.BaseVit = int(packed.BaseVit)
	player.StatPts = int(packed.StatPts)
	player.Experience = packed.Exp
	player.Gold = int(packed.Gold)
	player.MaxManaBase = int(packed.MaxMana)
	player.ManaBase = int(packed.ManaBase)
	player.MemSpells = packed.MemSpells
	unpackPlayerSpells(&player, packed, env)
	unpackPlayerSlots(&player, packed, env)
	player.Reflections = packed.Reflections
	player.DamAcFlags = uint8(packed.DamAcFlags)

	return player
}

// packPlayerHead stores the destination, position and name block
// (pack.cpp:155-171).
func packPlayerHead(packed *PlayerPack, player entities.Player, env PackEnv) {
	packed.DestAction = player.DestAction
	packed.DestParam1 = int8(player.DestParam1)
	packed.DestParam2 = int8(player.DestParam2)
	packed.PlrLevel = player.PlrLevel
	packed.Px = uint8(player.Position.Tile.X)
	packed.Py = uint8(player.Position.Tile.Y)

	if env.Vanilla {
		packed.TargX = packed.Px
		packed.TargY = packed.Py
	}

	copy(packed.Name[:], player.Name[:])
}

// packPlayerStats stores the class, stat and spell block (pack.cpp:172-199).
func packPlayerStats(packed *PlayerPack, player entities.Player, charLevel uint8) {
	packed.Class = player.Class
	packed.BaseStr = uint8(player.BaseStr)
	packed.BaseMag = uint8(player.BaseMag)
	packed.BaseDex = uint8(player.BaseDex)
	packed.BaseVit = uint8(player.BaseVit)
	packed.Level = charLevel
	packed.StatPts = uint8(player.StatPts)
	packed.Exp = player.Experience
	packed.Gold = int32(player.Gold)
	packed.HPBase = int32(player.HPBase)
	packed.MaxHP = int32(player.MaxHPBase)
	packed.ManaBase = int32(player.ManaBase)
	packed.MaxMana = int32(player.MaxManaBase)
	packed.MemSpells = player.MemSpells

	for i := range packed.SplLvl {
		packed.SplLvl[i] = player.SplLvl[i]
	}

	for i := range packed.SplLvl2 {
		packed.SplLvl2[i] = player.SplLvl[SpellLevelsPacked+i]
	}
}

// packPlayerSlots stores the item block (pack.cpp:201-216).
func packPlayerSlots(packed *PlayerPack, player entities.Player, env PackEnv) {
	for i := range packed.InvBody {
		packed.InvBody[i] = PackItem(player.InvBody[i], env)
	}

	packed.NumInv = uint8(player.NumInv)
	for i := 0; i < int(packed.NumInv) && i < len(packed.InvList); i++ {
		packed.InvList[i] = PackItem(player.InvList[i], env)
	}

	packed.InvGrid = player.InvGrid

	for i := range packed.SpdList {
		packed.SpdList[i] = PackItem(player.SpdList[i], env)
	}

	packed.Reflections = player.Reflections
	packed.DamAcFlags = uint32(player.DamAcFlags)
}

// unpackPlayerSpells fills the 47 packed spell-level slots (pack.cpp:396-415).
func unpackPlayerSpells(player *entities.Player, packed PlayerPack, env PackEnv) {
	copy(player.SplLvl[:SpellLevelsPacked], packed.SplLvl[:])
	copy(player.SplLvl[SpellLevelsPacked:], packed.SplLvl2[:])

	// Apocalypse and Nova are unavailable in Diablo (pack.cpp:410-414).
	if !env.Hellfire {
		player.SplLvl[int8(SpellApocalypse)] = 0
		player.SplLvl[int8(SpellNova)] = 0
	}
}

// unpackPlayerSlots fills the item block (pack.cpp:417-440).
func unpackPlayerSlots(player *entities.Player, packed PlayerPack, env PackEnv) {
	for i := range packed.InvBody {
		player.InvBody[i] = unpackOrEmpty(packed.InvBody[i], env)
	}

	player.NumInv = int(packed.NumInv)
	for i := 0; i < player.NumInv && i < len(packed.InvList); i++ {
		player.InvList[i] = unpackOrEmpty(packed.InvList[i], env)
	}

	player.InvGrid = packed.InvGrid

	for i := range packed.SpdList {
		player.SpdList[i] = unpackOrEmpty(packed.SpdList[i], env)
	}
}

// unpackOrEmpty unpacks one slot, keeping the empty slot when the record or
// the itemdat table rejects it.
func unpackOrEmpty(packed ItemPack, env PackEnv) entities.Item {
	item, err := UnpackItem(packed, env)
	if err != nil {
		return inventory.EmptyItem()
	}

	return item
}
