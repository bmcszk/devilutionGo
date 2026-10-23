// Package monstdat loads the monster data tables (monstdat.tsv,
// unique_monstdat.tsv) and holds the AiProc behaviour registry keyed by
// monsterAIID.
package monstdat

import (
	"errors"
	"fmt"
	"strings"
)

// monsterAIID identifies a monster behaviour.
type monsterAIID int8

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

// Unique-item ids used by parseTreasure.
const (
	UITEMCLEAVER uint16 = iota
	UITEMSKCROWN
)

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

// Raw returns the byte-exact source columns for parity assertions.
func (m monsterData) Raw() []string { return m.raw }

// ParseMonstdat parses monstdat.tsv (41 columns) into one monsterData per row.
func ParseMonstdat(data []byte) ([]monsterData, error) {
	rows, err := splitTSV(data, 41)
	if err != nil {
		return nil, err
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
	return splitTSV(data, 0)
}

func splitTSV(data []byte, wantCols int) ([][]string, error) {
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var rows [][]string
	var header []string
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		cols := strings.Split(ln, "\t")
		if header == nil {
			header = cols
			continue
		}
		if wantCols != 0 && len(cols) != wantCols {
			return nil, fmt.Errorf("monstdat: row has %d cols, want %d: %q", len(cols), wantCols, ln)
		}
		rows = append(rows, cols)
	}
	if header == nil {
		return nil, errors.New("monstdat: empty table")
	}
	return rows, nil
}

// AiProc maps monsterAIID → behaviour; it holds 40 entries and exactly one
// nil, at FireMan.
var AiProc = map[monsterAIID]func(){}

// AiProcByName keys the same table by the TSV ai column spelling.
var AiProcByName = map[string]func(){}

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

// Placeholder behaviours: the registry shape and its null-ness are the
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

// getMissileType maps a monster id to the missile its ranged attack fires;
// unknown ids fall back to Arrow.
func getMissileType(id string) string {
	switch id {
	case "Succubus", "LazarusSuccubus":
		return "BloodStar"
	case "Acid", "AcidUnique":
		return "Acid"
	case "FireBat":
		return "Firebolt"
	case "Torchant":
		return "Fireball"
	case "Lich":
		return "OrangeFlare"
	case "ArchLich":
		return "YellowFlare"
	case "Psychorb":
		return "BlueFlare"
	case "Necromorb":
		return "RedFlare"
	case "Magma":
		return "MagmaBall"
	case "Storm":
		return "ThinLightningControl"
	case "Diablo":
		return "DiabloApocalypse"
	case "BoneDemon":
		return "BlueFlare2"
	default:
		return "Arrow"
	}
}
