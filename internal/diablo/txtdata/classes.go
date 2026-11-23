// Package txtdata reads the Diablo-specific TSV tables; the table mechanism
// (parsing, mod overrides) lives in internal/engine/txtdata.
package txtdata

import (
	"fmt"
	"strconv"
	"strings"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// classFixedShift is the 6-bit fixed-point scale of every decimal class field
// (utils/fixed_point.hpp:345 Fixed10_6).
const classFixedShift = 6

// classFixedMin and classFixedMax bound a parsed integer part to the Fixed10_6
// storage range before it is shifted (utils/parse_int.hpp:63-64).
const (
	classFixedMin = -512
	classFixedMax = 511
)

// classFractionDigits, classFractionRound and classFractionStep are
// ParseFixed6Fraction's fraction normalisation: seven decimal places rounded to
// the nearest 1/64 (utils/parse_int.cpp:7-32).
const (
	classFractionDigits = 7
	classFractionRound  = 78_125
	classFractionStep   = 156_250
)

// The class flag bits, in PlayerClassFlag's own order (playerdat.hpp:28-36).
const (
	ClassFlagCriticalStrike    ClassFlag = 1 << 0
	ClassFlagDualWield         ClassFlag = 1 << 1
	ClassFlagIronSkin          ClassFlag = 1 << 2
	ClassFlagNaturalResistance ClassFlag = 1 << 3
	ClassFlagTrapSense         ClassFlag = 1 << 4
)

// ClassFlag is the PlayerClassFlag bitset of one class (playerdat.hpp:28-38).
type ClassFlag uint8

// ClassData is one classdat.tsv row (classes.cpp ClassData).
type ClassData struct {
	ClassName  string
	FolderName string
}

// ClassAttributes is one attributes.tsv row: the per-class stat baselines and
// multipliers (playerdat.hpp:72-118). Every decimal field is in 1/64 units,
// matching Fixed10_6::raw().
type ClassAttributes struct {
	ClassFlags           ClassFlag
	BaseStr              int
	BaseMag              int
	BaseDex              int
	BaseVit              int
	MaxStr               int
	MaxMag               int
	MaxDex               int
	MaxVit               int
	AdjLife              int
	AdjMana              int
	LvlLife              int
	LvlMana              int
	ChrLife              int
	ChrMana              int
	ItmLife              int
	ItmMana              int
	ManaCost             int
	ItmRestoreLife       int
	ItmRestoreMana       int
	SplRestoreLife       int
	SplRestoreMana       int
	HealOtherRestoreLife int
}

// PlayerCombatData is the combat half of attributes.tsv: blockBonus and the
// three to-hit baselines (playerdat.hpp:122-133).
type PlayerCombatData struct {
	BaseToBlock     int
	BaseMagicToHit  int
	BaseMeleeToHit  int
	BaseRangedToHit int
}

// ClassLoadout is one starting_loadout.tsv row: the skill and spell a class
// starts with, five item ids and the starting gold (playerdat.hpp:146-166).
type ClassLoadout struct {
	Skill      string
	Spell      string
	SpellLevel int
	Items      [5]string
	Gold       int
}

// ClassAnimations is one animations.tsv row: the frame counts and action frames
// of every weapon animation (playerdat.hpp:180-230).
type ClassAnimations struct {
	UnarmedFrames            int
	UnarmedActionFrame       int
	UnarmedShieldFrames      int
	UnarmedShieldActionFrame int
	SwordFrames              int
	SwordActionFrame         int
	SwordShieldFrames        int
	SwordShieldActionFrame   int
	BowFrames                int
	BowActionFrame           int
	AxeFrames                int
	AxeActionFrame           int
	MaceFrames               int
	MaceActionFrame          int
	MaceShieldFrames         int
	MaceShieldActionFrame    int
	StaffFrames              int
	StaffActionFrame         int
	IdleFrames               int
	WalkingFrames            int
	BlockingFrames           int
	DeathFrames              int
	CastingFrames            int
	RecoveryFrames           int
	TownIdleFrames           int
	TownWalkingFrames        int
	CastingActionFrame       int
}

// ClassPoint is a sprite offset pair (Point, playerdat.hpp:171-178).
type ClassPoint struct {
	X, Y int
}

// ClassSprites is one sprites.tsv row: the graphic path, the class letter, the
// TRN file, the three portrait offsets and the ten sprite widths
// (playerdat.hpp:168-202).
type ClassSprites struct {
	ClassPath          string
	ClassChar          byte
	TRN                string
	PartyOffsetTown    ClassPoint
	PartyOffsetDungeon ClassPoint
	PartyOffsetDead    ClassPoint
	Stand              int
	Walk               int
	Attack             int
	Bow                int
	SwHit              int
	Block              int
	Lightning          int
	Fire               int
	Magic              int
	Death              int
}

// classValueReader walks a key/value class table positionally, the way
// ValueReader does: every read asserts the next record's key and consumes it,
// so a reordered, short or padded table cannot shift values between fields
// (data/value_reader.cpp:16-35). The first error sticks, so a loader reads its
// whole key list and checks once, like persist.recordReader.
type classValueReader struct {
	rows [][]string
	next int
	err  error
}

// LoadClasses parses classdat.tsv rows (header className,folderName,
// portrait,inv; playerdat.cpp:345-369).
func LoadClasses(t *enginetxtdata.Table) ([]ClassData, error) {
	ci, err := t.Column("className")
	if err != nil {
		return nil, fmt.Errorf("txtdata: classdat className column: %w", err)
	}

	fi, err := t.Column("folderName")
	if err != nil {
		return nil, fmt.Errorf("txtdata: classdat folderName column: %w", err)
	}

	out := make([]ClassData, 0, len(t.Rows))

	for _, row := range t.Rows {
		if ci >= len(row) || fi >= len(row) {
			return nil, fmt.Errorf("txtdata: classdat row too short: %v", row)
		}

		out = append(out, ClassData{ClassName: row[ci], FolderName: row[fi]})
	}

	return out, nil
}

// LoadClassAttributes parses one attributes.tsv record into the stat baselines
// and the combat chances (playerdat.cpp:169-205).
func LoadClassAttributes(t *enginetxtdata.Table) (ClassAttributes, PlayerCombatData, error) {
	var (
		attrs  ClassAttributes
		combat PlayerCombatData
	)

	r := &classValueReader{rows: t.Rows}

	r.flags("classFlags", &attrs.ClassFlags)
	r.int("baseStr", &attrs.BaseStr)
	r.int("baseMag", &attrs.BaseMag)
	r.int("baseDex", &attrs.BaseDex)
	r.int("baseVit", &attrs.BaseVit)
	r.int("maxStr", &attrs.MaxStr)
	r.int("maxMag", &attrs.MaxMag)
	r.int("maxDex", &attrs.MaxDex)
	r.int("maxVit", &attrs.MaxVit)
	r.int("blockBonus", &combat.BaseToBlock)
	r.decimal("adjLife", &attrs.AdjLife)
	r.decimal("adjMana", &attrs.AdjMana)
	r.decimal("lvlLife", &attrs.LvlLife)
	r.decimal("lvlMana", &attrs.LvlMana)
	r.decimal("chrLife", &attrs.ChrLife)
	r.decimal("chrMana", &attrs.ChrMana)
	r.decimal("itmLife", &attrs.ItmLife)
	r.decimal("itmMana", &attrs.ItmMana)
	r.decimal("manaCost", &attrs.ManaCost)
	r.decimal("itmRestoreLife", &attrs.ItmRestoreLife)
	r.decimal("itmRestoreMana", &attrs.ItmRestoreMana)
	r.decimal("splRestoreLife", &attrs.SplRestoreLife)
	r.decimal("splRestoreMana", &attrs.SplRestoreMana)
	r.decimal("healOtherRestoreLife", &attrs.HealOtherRestoreLife)
	r.int("baseMagicToHit", &combat.BaseMagicToHit)
	r.int("baseMeleeToHit", &combat.BaseMeleeToHit)
	r.int("baseRangedToHit", &combat.BaseRangedToHit)

	return attrs, combat, r.err
}

// LoadClassLoadout parses one starting_loadout.tsv record: skill, spell, its
// level, the five item slots and the starting gold (playerdat.cpp:207-223).
func LoadClassLoadout(t *enginetxtdata.Table) (ClassLoadout, error) {
	var loadout ClassLoadout

	r := &classValueReader{rows: t.Rows}

	r.text("skill", &loadout.Skill)
	r.text("spell", &loadout.Spell)
	r.int("spellLevel", &loadout.SpellLevel)

	for i := range loadout.Items {
		r.text("item"+strconv.Itoa(i), &loadout.Items[i])
	}

	r.int("gold", &loadout.Gold)

	return loadout, r.err
}

// LoadClassAnimations parses one animations.tsv record into the frame counts
// and action frames of every weapon animation (playerdat.cpp:255-291).
func LoadClassAnimations(t *enginetxtdata.Table) (ClassAnimations, error) {
	var anim ClassAnimations

	r := &classValueReader{rows: t.Rows}

	r.int("unarmedFrames", &anim.UnarmedFrames)
	r.int("unarmedActionFrame", &anim.UnarmedActionFrame)
	r.int("unarmedShieldFrames", &anim.UnarmedShieldFrames)
	r.int("unarmedShieldActionFrame", &anim.UnarmedShieldActionFrame)
	r.int("swordFrames", &anim.SwordFrames)
	r.int("swordActionFrame", &anim.SwordActionFrame)
	r.int("swordShieldFrames", &anim.SwordShieldFrames)
	r.int("swordShieldActionFrame", &anim.SwordShieldActionFrame)
	r.int("bowFrames", &anim.BowFrames)
	r.int("bowActionFrame", &anim.BowActionFrame)
	r.int("axeFrames", &anim.AxeFrames)
	r.int("axeActionFrame", &anim.AxeActionFrame)
	r.int("maceFrames", &anim.MaceFrames)
	r.int("maceActionFrame", &anim.MaceActionFrame)
	r.int("maceShieldFrames", &anim.MaceShieldFrames)
	r.int("maceShieldActionFrame", &anim.MaceShieldActionFrame)
	r.int("staffFrames", &anim.StaffFrames)
	r.int("staffActionFrame", &anim.StaffActionFrame)
	r.int("idleFrames", &anim.IdleFrames)
	r.int("walkingFrames", &anim.WalkingFrames)
	r.int("blockingFrames", &anim.BlockingFrames)
	r.int("deathFrames", &anim.DeathFrames)
	r.int("castingFrames", &anim.CastingFrames)
	r.int("recoveryFrames", &anim.RecoveryFrames)
	r.int("townIdleFrames", &anim.TownIdleFrames)
	r.int("townWalkingFrames", &anim.TownWalkingFrames)
	r.int("castingActionFrame", &anim.CastingActionFrame)

	return anim, r.err
}

// LoadClassSprites parses one sprites.tsv record: the graphic path, the class
// letter, the TRN file, three portrait offsets and ten sprite widths
// (playerdat.cpp:225-253).
func LoadClassSprites(t *enginetxtdata.Table) (ClassSprites, error) {
	var sprites ClassSprites

	r := &classValueReader{rows: t.Rows}

	r.text("classPath", &sprites.ClassPath)
	r.char("classChar", &sprites.ClassChar)
	r.text("trn", &sprites.TRN)
	r.int("partyOffsetTownX", &sprites.PartyOffsetTown.X)
	r.int("partyOffsetTownY", &sprites.PartyOffsetTown.Y)
	r.int("partyOffsetDungeonX", &sprites.PartyOffsetDungeon.X)
	r.int("partyOffsetDungeonY", &sprites.PartyOffsetDungeon.Y)
	r.int("partyOffsetDeadX", &sprites.PartyOffsetDead.X)
	r.int("partyOffsetDeadY", &sprites.PartyOffsetDead.Y)
	r.int("stand", &sprites.Stand)
	r.int("walk", &sprites.Walk)
	r.int("attack", &sprites.Attack)
	r.int("bow", &sprites.Bow)
	r.int("swHit", &sprites.SwHit)
	r.int("block", &sprites.Block)
	r.int("lightning", &sprites.Lightning)
	r.int("fire", &sprites.Fire)
	r.int("magic", &sprites.Magic)
	r.int("death", &sprites.Death)

	return sprites, r.err
}

// LoadClassTables parses one per-class two-column key/value table
// (Attribute,Value; SourceD1 monster.cpp-style layout).
func LoadClassTables(t *enginetxtdata.Table) ([][2]string, error) {
	out := make([][2]string, 0, len(t.Rows))

	for _, row := range t.Rows {
		if len(row) < 2 {
			return nil, fmt.Errorf("txtdata: class table row too short: %v", row)
		}

		out = append(out, [2]string{row[0], row[1]})
	}

	return out, nil
}

// value takes the next record and checks its key against want.
func (r *classValueReader) value(want string) (string, bool) {
	if r.err != nil {
		return "", false
	}

	if r.next >= len(r.rows) {
		r.err = fmt.Errorf("txtdata: missing field %q", want)
		return "", false
	}

	row := r.rows[r.next]
	r.next++

	if len(row) == 0 {
		r.err = fmt.Errorf("txtdata: missing field %q", want)
		return "", false
	}

	if row[0] != want {
		r.err = fmt.Errorf("txtdata: unexpected field %q, want %q", row[0], want)
		return "", false
	}

	if len(row) < 2 {
		r.err = fmt.Errorf("txtdata: not enough columns for field %q", want)
		return "", false
	}

	return row[1], true
}

// int reads the next record as an integer.
func (r *classValueReader) int(key string, dst *int) {
	cell, ok := r.value(key)
	if !ok {
		return
	}

	parsed, err := strconv.Atoi(cell)
	if err != nil {
		r.err = fmt.Errorf("txtdata: field %q value %q: %w", key, cell, err)
		return
	}

	*dst = parsed
}

// decimal reads the next record as a 1/64 fixed-point value.
func (r *classValueReader) decimal(key string, dst *int) {
	cell, ok := r.value(key)
	if !ok {
		return
	}

	parsed, err := parseClassDecimal(cell)
	if err != nil {
		r.err = fmt.Errorf("txtdata: field %q: %w", key, err)
		return
	}

	*dst = parsed
}

// text reads the next record as a string.
func (r *classValueReader) text(key string, dst *string) {
	cell, ok := r.value(key)
	if !ok {
		return
	}

	*dst = cell
}

// char reads the next record as a single character.
func (r *classValueReader) char(key string, dst *byte) {
	cell, ok := r.value(key)
	if !ok {
		return
	}

	if len(cell) != 1 {
		r.err = fmt.Errorf("txtdata: field %q value %q is not a single character", key, cell)
		return
	}

	*dst = cell[0]
}

// flags reads the next record as a comma-separated flag list OR'd into one
// bitset; an empty cell is no flags (data/iterators.hpp:201-212).
func (r *classValueReader) flags(key string, dst *ClassFlag) {
	cell, ok := r.value(key)
	if !ok {
		return
	}

	*dst = 0

	if cell == "" {
		return
	}

	for _, name := range strings.Split(cell, ",") {
		flag, found := classFlagByName(name)
		if !found {
			r.err = fmt.Errorf("txtdata: field %q value %q is not a class flag", key, name)
			return
		}

		*dst |= flag
	}
}

// classFlagByName maps a classFlags cell name to its bit
// (ParsePlayerClassFlag, playerdat.cpp:160-167).
func classFlagByName(name string) (ClassFlag, bool) {
	switch name {
	case "CriticalStrike":
		return ClassFlagCriticalStrike, true
	case "DualWield":
		return ClassFlagDualWield, true
	case "IronSkin":
		return ClassFlagIronSkin, true
	case "NaturalResistance":
		return ClassFlagNaturalResistance, true
	case "TrapSense":
		return ClassFlagTrapSense, true
	default:
		return 0, false
	}
}

// parseClassDecimal parses a decimal cell to 1/64 units: the integer part is
// shifted up and the fraction, rounded to the nearest step, is added to it, or
// subtracted when the cell is negative (ParseFixed6,
// utils/parse_int.hpp:47-114 and utils/parse_int.cpp:7-32).
func parseClassDecimal(cell string) (int, error) {
	if cell == "" {
		return 0, fmt.Errorf("empty value %q", cell)
	}

	negative := strings.HasPrefix(cell, "-")

	magnitude := strings.TrimPrefix(cell, "-")
	whole, rest := splitLeadingDigits(magnitude)

	if whole == "" && !strings.HasPrefix(rest, ".") {
		return 0, fmt.Errorf("value %q is not a number", cell)
	}

	integer, err := parseClassInteger(cell, whole)
	if err != nil {
		return 0, err
	}

	scaled := integer << classFixedShift

	if !strings.HasPrefix(rest, ".") {
		return scaled, nil
	}

	fraction := parseIntFraction(rest[1:])
	if negative {
		return scaled - fraction, nil
	}

	return scaled + fraction, nil
}

// parseClassInteger parses the integer digits with the cell's sign; a bare
// fraction such as ".5" has a zero integer part (utils/parse_int.hpp:106
// value_or(0)).
func parseClassInteger(cell, whole string) (int, error) {
	if whole == "" {
		return 0, nil
	}

	sign := ""

	if strings.HasPrefix(cell, "-") {
		sign = "-"
	}

	integer, err := strconv.Atoi(sign + whole)
	if err != nil {
		return 0, fmt.Errorf("value %q: %w", cell, err)
	}

	if integer > classFixedMax || integer < classFixedMin {
		return 0, fmt.Errorf("value %q out of fixed-point range", cell)
	}

	return integer, nil
}

// splitLeadingDigits splits the integer digits off a numeric cell, the way
// std::from_chars stops at the first non-digit (utils/parse_int.hpp:27-40).
func splitLeadingDigits(cell string) (digits, rest string) {
	i := 0
	for i < len(cell) && cell[i] >= '0' && cell[i] <= '9' {
		i++
	}

	return cell[:i], cell[i:]
}

// parseIntFraction normalises the fraction digits to seven places and rounds
// them to the nearest 1/64 (ParseFixed6Fraction, utils/parse_int.cpp:7-32).
func parseIntFraction(cell string) int {
	digits := 0

	var fraction uint64

	for digits < classFractionDigits && digits < len(cell) && cell[digits] >= '0' && cell[digits] <= '9' {
		fraction = fraction*10 + uint64(cell[digits]-'0')
		digits++
	}

	for digits < classFractionDigits {
		fraction *= 10
		digits++
	}

	return int((fraction + classFractionRound) / classFractionStep)
}
