//go:build windowed

package main

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// classdatPackSource builds every committed class's starting pack from the
// committed per-class attributes.tsv row (dcob D7): baseStr/baseMag/baseDex/
// baseVit through txtdata.LoadClassAttributes, and the level-1 vitals the
// calculateBaseLife/calculateBaseMana transcription (player.cpp:1976-1980
// @e1ad0c2b6: adjLife + lvlLife*1 + chrLife*baseVit) — the loader's decimal
// columns already arrive in the 1/64 raw scale upstream stores
// (player.h:236-243), so the sum IS the fixed-point value (review R1 F1:
// never shifted down; pack.cpp compares .sv bytes against .raw()).
// NOTE the real Hellfire asset gating of Bard/Barbarian (selhero.cpp:179/:183/
// :187 @e1ad0c2b6) is a recorded deviation: this source serves all six
// committed classes; gating by MPQ assets is a follow-up bean, not here.
type classdatPackSource struct{}

// StartingPack implements world.PackSource for all six committed classes.
func (classdatPackSource) StartingPack(class int) (world.StartingPack, bool) {
	if class < 0 || class >= len(classDirs) {
		return world.StartingPack{}, false
	}

	data, err := scene.ReadCommittedFixture(classDirs[class], "attributes.tsv")
	if err != nil {
		return world.StartingPack{}, false
	}

	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return world.StartingPack{}, false
	}

	attrs, _, err := txtdata.LoadClassAttributes(tbl)
	if err != nil {
		return world.StartingPack{}, false
	}

	return world.StartingPack{
		Class:       class,
		BaseStr:     attrs.BaseStr,
		BaseMag:     attrs.BaseMag,
		BaseDex:     attrs.BaseDex,
		BaseVit:     attrs.BaseVit,
		HPBase:      classVital(attrs.AdjLife, attrs.LvlLife, attrs.ChrLife, attrs.BaseVit),
		MaxHPBase:   classVital(attrs.AdjLife, attrs.LvlLife, attrs.ChrLife, attrs.BaseVit),
		ManaBase:    classVital(attrs.AdjMana, attrs.LvlMana, attrs.ChrMana, attrs.BaseMag),
		MaxManaBase: classVital(attrs.AdjMana, attrs.LvlMana, attrs.ChrMana, attrs.BaseMag),
	}, true
}

// classVital computes one level-1 vital: the life/mana formula
// (calculateBaseLife, player.cpp:1976-1980 @e1ad0c2b6) with characterLevel 1:
// adj + lvl*1 + chr*base. The loader's decimal columns are already 1/64 raw
// fixed-point ints, so the sum is kept raw — no shift in either direction
// (review R1 F1: upstream stores the Fixed26_6 raw value, pack.cpp compares
// the .sv bytes against .raw()).
func classVital(adj, lvl, chr, base int) int {
	return adj + lvl + chr*base
}

// The windowed boot serves the data-built packs (dcob D7).
func init() { lanePackSource = classdatPackSource{} }
