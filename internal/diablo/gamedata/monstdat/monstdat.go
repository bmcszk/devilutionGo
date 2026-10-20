// Package monstdat ports the monster data tables: monstdat.tsv
// (Source/tables/monstdat.cpp:409 loader), unique_monstdat.tsv (:459), and
// the AiProc function-pointer table (monster.cpp:3089-3130) as a Go
// registry keyed by MonsterAIID.
package monstdat

import (
	"errors"
	"fmt"
	"strings"
)

// MonsterAIID mirrors MonsterAIID (monstdat.h:22-63) in declaration order.
type MonsterAIID int8

// AI ids, verbatim order (Invalid = -1).
const (
	AIZombie MonsterAIID = iota
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
	AIInvalid MonsterAIID = -1
)

// Treasure bit flags (monstdat.h:86-88).
const (
	T_MASK   uint16 = 0xFFF
	T_NODROP uint16 = 0x4000 // monster doesn't drop any loot
	T_UNIQ   uint16 = 0x8000 // combined with unique item's ID
)

// MonsterData is one monstdat.tsv row. Numeric/flag columns are decoded;
// raw preserves the byte-exact source line for parity assertions.
type MonsterData struct {
	ID           string // _monster_id column (TSV key)
	Name         string
	AI           string // ai column, enum name as written in the TSV
	Treasure     uint16
	DropsSpecial bool // treasure != 0 && T_NODROP not set
	raw          []string
}

// Unique-item ids used by ParseMonsterTreasure (items.h:46-48).
const (
	UITEMCLEAVER uint16 = iota // UITEM_CLEAVER
	UITEMSKCROWN               // UITEM_SKCROWN
)

// parseTreasure ports ParseMonsterTreasure (monstdat.cpp:299-310): ""
// → 0, "None" → T_NODROP, Uniq(NAME) → T_UNIQ|UITEM_NAME.
func parseTreasure(value string) (uint16, error) {
	if value == "" {
		return 0, nil
	}
	if value == "None" {
		return T_NODROP, nil
	}
	// The Uniq(NAME) table mirrors monstdat.cpp:305-306; both arms OR the
	// item id (CLEAVER is 0 — the OR is still upstream-faithful there).
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
func (m MonsterData) Raw() []string { return m.raw }

// ParseMonstdat parses monstdat.tsv (41 columns) in row order.
func ParseMonstdat(data []byte) ([]MonsterData, error) {
	rows, err := splitTSV(data, 41)
	if err != nil {
		return nil, err
	}
	out := make([]MonsterData, 0, len(rows))
	for _, r := range rows {
		treasure, terr := parseTreasure(r[39])
		if terr != nil {
			return nil, fmt.Errorf("monstdat: row %s: %w", r[0], terr)
		}
		m := MonsterData{
			ID:       r[0],
			Name:     r[1],
			AI:       r[17],
			Treasure: treasure,
			raw:      r,
		}
		// Uniq(...) writes T_UNIQ | id (monstdat.cpp:39); T_NODROP suppresses
		// loot (monstdat.h:87).
		m.DropsSpecial = m.Treasure&T_UNIQ != 0 && m.Treasure&T_NODROP == 0
		out = append(out, m)
	}
	return out, nil
}

// ParseUniqueMonstdat parses unique_monstdat.tsv (no hf override exists);
// rows stay raw until the unique-monster bean decodes them.
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

// AiProc maps MonsterAIID → behaviour, exactly the C++ AiProc[] shape
// (monster.cpp:3089-3130): 40 entries, exactly one nil at FireMan.
var AiProc = map[MonsterAIID]func(){}

// AiProcByName keys the same table by the TSV ai column spelling.
var AiProcByName = map[string]func(){}

func init() {
	entries := []struct {
		id   MonsterAIID
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
		{AIFireMan, "FireMan", nil}, // monster.cpp:3106 — the only nil
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
		AiProcByName[e.name] = e.fn // nil at FireMan resolves by name, upstream shape
	}
}

// Behaviour placeholders: the AI implementations port with the monster
// engine (devilGo-mwlm); the registry shape and null-ness are this bean's
// contract, so each entry holds a distinct named placeholder.
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

// GetMissileType maps monster id → the missile its ranged attack fires
// (monster.cpp GetMissileType region); default Arrow.
func GetMissileType(id string) string {
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
