// Package monstdat loads the monster data tables (monstdat.tsv,
// unique_monstdat.tsv) and holds the AiProc behavior registry keyed by
// monsterAIID.
package monstdat

import (
	"fmt"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// AI ids in declaration order; AIInvalid marks the absence of an id.
const (
	AIZombie monsterAIID = iota
	AIFat
	AISkeletonMelee
	AISkeletonRanged
	AIScavenger
	AIRhino
	AIGoatMelee
	AIGoatRanged
	AIFallen
	AIMagma
	AISkeletonKing
	AIBat
	AIGargoyle
	AIButcher
	AISuccubus
	AISneak
	AIStorm
	AIFireMan
	AIGharbad
	AIAcid
	AIAcidUnique
	AIGolem
	AIZhar
	AISnotspill
	AISnake
	AICounselor
	AIMega
	AIDiablo
	AILazarus
	AILazarusSuccubus
	AILachdanan
	AIWarlord
	AIFireBat
	AITorchant
	AIHorkDemon
	AILich
	AIArchLich
	AIPsychorb
	AINecromorb
	AIBoneDemon
	AIInvalid monsterAIID = -1
)

const (
	T_MASK   uint16 = 0xFFF  // masks the unique item id
	T_NODROP uint16 = 0x4000 // monster doesn't drop any loot
	T_UNIQ   uint16 = 0x8000 // combined with unique item's ID
)

// Unique-item ids used by parseTreasure.
const (
	UITEMCLEAVER uint16 = iota
	UITEMSKCROWN
)

// AiProc maps monsterAIID → behavior; it holds 40 entries and exactly one
// nil, at FireMan.
// AiProcByName keys the same table by the TSV ai column spelling.
var (
	AiProc       = map[monsterAIID]func(){}
	AiProcByName = map[string]func(){}
)

// monsterMissileType maps monster IDs to their ranged missile type.
// Mirrors the switch in MissileTypeFor (monstdat.cpp).
var monsterMissileType = map[string]string{
	"Succubus":        "BloodStar",
	"LazarusSuccubus": "BloodStar",
	"Acid":            "Acid",
	"AcidUnique":      "Acid",
	"FireBat":         "Firebolt",
	"Torchant":        "Fireball",
	"Lich":            "OrangeFlare",
	"ArchLich":        "YellowFlare",
	"Psychorb":        "BlueFlare",
	"Necromorb":       "RedFlare",
	"Magma":           "MagmaBall",
	"Storm":           "ThinLightningControl",
	"Diablo":          "DiabloApocalypse",
	"BoneDemon":       "BlueFlare2",
}

// monsterAIID identifies a monster behavior.
type monsterAIID int8

// monsterData is one monstdat.tsv row; raw preserves the byte-exact source
// columns.
type monsterData struct {
	ID           string // _monster_id column (TSV key)
	Name         string
	AI           string // ai column, enum name as written in the TSV
	Treasure     uint16
	DropsSpecial bool
	raw          []string
}

// MonsterData is the exported monstdat row type (upstream MonsterData
// struct, monsters.cpp / monstdat.txt columns).
type MonsterData = monsterData

func init() {
	entries := []struct {
		id   monsterAIID
		name string
		fn   func()
	}{
		{AIZombie, "Zombie", zombieAI},
		{AIFat, "Fat", overlordAI},
		{AISkeletonMelee, "SkeletonMelee", skeletonAI},
		{AISkeletonRanged, "SkeletonRanged", skeletonBowAI},
		{AIScavenger, "Scavenger", scavengerAI},
		{AIRhino, "Rhino", rhinoAI},
		{AIGoatMelee, "GoatMelee", aiAvoidance},
		{AIGoatRanged, "GoatRanged", aiRanged},
		{AIFallen, "Fallen", fallenAI},
		{AIMagma, "Magma", aiRangedAvoidance},
		{AISkeletonKing, "SkeletonKing", leoricAI},
		{AIBat, "Bat", batAI},
		{AIGargoyle, "Gargoyle", gargoyleAI},
		{AIButcher, "Butcher", butcherAI},
		{AISuccubus, "Succubus", aiRanged},
		{AISneak, "Sneak", sneakAI},
		{AIStorm, "Storm", aiRangedAvoidance},
		{AIFireMan, "FireMan", nil},
		{AIGharbad, "Gharbad", gharbadAI},
		{AIAcid, "Acid", aiRangedAvoidance},
		{AIAcidUnique, "AcidUnique", aiRanged},
		{AIGolem, "Golem", golumAI},
		{AIZhar, "Zhar", zharAI},
		{AISnotspill, "Snotspill", snotSpilAI},
		{AISnake, "Snake", snakeAI},
		{AICounselor, "Counselor", counselorAI},
		{AIMega, "Mega", megaAI},
		{AIDiablo, "Diablo", aiRangedAvoidance},
		{AILazarus, "Lazarus", lazarusAI},
		{AILazarusSuccubus, "LazarusSuccubus", lazarusMinionAI},
		{AILachdanan, "Lachdanan", lachdananAI},
		{AIWarlord, "Warlord", warlordAI},
		{AIFireBat, "FireBat", aiRanged},
		{AITorchant, "Torchant", aiRanged},
		{AIHorkDemon, "HorkDemon", horkDemonAI},
		{AILich, "Lich", aiRanged},
		{AIArchLich, "ArchLich", aiRanged},
		{AIPsychorb, "Psychorb", psychorbAI},
		{AINecromorb, "Necromorb", necromorbAI},
		{AIBoneDemon, "BoneDemon", boneDemonAI},
	}
	if len(entries) != 40 {
		panic("monstdat: AiProc table must hold exactly 40 entries")
	}

	for _, e := range entries {
		AiProc[e.id] = e.fn
		AiProcByName[e.name] = e.fn
	}
}

// Raw returns the byte-exact source columns for parity assertions.
func (m monsterData) Raw() []string { return m.raw }

// ParseMonstdat parses monstdat.tsv (41 columns) into one monsterData per row.
func ParseMonstdat(data []byte) ([]monsterData, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("monstdat: parse monstdat.tsv: %w", err)
	}

	rows := tbl.Rows
	if len(tbl.Header) != 41 {
		return nil, fmt.Errorf("monstdat: header has %d cols, want 41", len(tbl.Header))
	}

	out := make([]monsterData, 0, len(rows))
	for _, r := range rows {
		treasure, terr := parseTreasure(r[39])
		if terr != nil {
			return nil, fmt.Errorf("monstdat: row %s: %w", r[0], terr)
		}

		m := monsterData{
			ID:       r[0],
			Name:     r[1],
			AI:       r[17],
			Treasure: treasure,
			raw:      r,
		}
		m.DropsSpecial = m.Treasure&T_UNIQ != 0 && m.Treasure&T_NODROP == 0
		out = append(out, m)
	}

	return out, nil
}

// ParseUniqueMonstdat parses unique_monstdat.tsv and returns its rows raw.
func ParseUniqueMonstdat(data []byte) ([][]string, error) {
	tbl, err := enginetxtdata.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("monstdat: parse unique_monstdat.tsv: %w", err)
	}

	return tbl.Rows, nil
}

// MissileTypeFor maps a monster id to the missile its ranged attack fires;
// unknown ids fall back to Arrow.
func MissileTypeFor(id string) string {
	if m, ok := monsterMissileType[id]; ok {
		return m
	}

	return "Arrow"
}

// parseTreasure decodes a treasure column: "" → 0, "None" → T_NODROP,
// Uniq(NAME) → T_UNIQ|item id.
func parseTreasure(value string) (uint16, error) {
	if value == "" {
		return 0, nil
	}

	if value == "None" {
		return T_NODROP, nil
	}

	switch value {
	case "Uniq(SKCROWN)":
		id := UITEMSKCROWN
		return T_UNIQ | id, nil
	case "Uniq(CLEAVER)":
		id := UITEMCLEAVER
		return T_UNIQ | id, nil
	}

	return 0, fmt.Errorf("monstdat: invalid treasure %q", value)
}

// Placeholder behaviors: the registry shape and its null-ness are the
// contract until the monster engine ports the real AI.
func zombieAI()          {}
func overlordAI()        {}
func skeletonAI()        {}
func skeletonBowAI()     {}
func scavengerAI()       {}
func rhinoAI()           {}
func aiAvoidance()       {}
func aiRanged()          {}
func fallenAI()          {}
func aiRangedAvoidance() {}
func leoricAI()          {}
func batAI()             {}
func gargoyleAI()        {}
func butcherAI()         {}
func sneakAI()           {}
func gharbadAI()         {}
func golumAI()           {}
func zharAI()            {}
func snotSpilAI()        {}
func snakeAI()           {}
func counselorAI()       {}
func megaAI()            {}
func lazarusAI()         {}
func lazarusMinionAI()   {}
func lachdananAI()       {}
func warlordAI()         {}
func horkDemonAI()       {}
func psychorbAI()        {}
func necromorbAI()       {}
func boneDemonAI()       {}
