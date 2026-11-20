package txtdata_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// The typed class tables the loaders read, named so a table test can pick one.
const (
	classAttributes = "attributes"
	classLoadout    = "starting_loadout"
	classAnimations = "animations"
	classSprites    = "sprites"
)

// committedClasses are the six classdat.tsv folder names the typed loaders are
// pinned against (assets/txtdata/classes/, playerdat.cpp:378-396).
var committedClasses = []string{"warrior", "rogue", "sorcerer", "monk", "bard", "barbarian"}

// classLoaders maps a table name to its loader, so one table test can drive any
// of the four typed loaders.
var classLoaders = map[string]func(*enginetxtdata.Table) error{
	classAttributes: loadAttributes,
	classLoadout:    loadLoadout,
	classAnimations: loadAnimations,
	classSprites:    loadSprites,
}

// loaderError labels one typed loader's failure with the table it came from.
type loaderError struct {
	kind string
	err  error
}

func TestLoadClasses(t *testing.T) {
	// given
	const classdatTSV = "className\tfolderName\tportrait\tinv\r\n" +
		"Warrior	warr\t0\tinv\r\n" +
		"Rogue	rogue\t1\tinv_rog\r\n"

	tbl := mustParse(t, []byte(classdatTSV))

	// when
	classes, err := txtdata.LoadClasses(tbl)
	if err != nil {
		t.Fatalf("txtdata.LoadClasses: %v", err)
	}

	// then
	if len(classes) != 2 {
		t.Fatalf("classes = %d, want 2", len(classes))
	}

	if classes[0].ClassName != "Warrior" || classes[0].FolderName != "warr" {
		t.Fatalf("class0 = %+v", classes[0])
	}
}

func TestLoadClassTables(t *testing.T) {
	// given
	const classTableTSV = "Attribute	Value\r\nclassFlags\tCriticalStrike\r\n"

	tbl := mustParse(t, []byte(classTableTSV))

	// when
	tables, err := txtdata.LoadClassTables(tbl)
	if err != nil {
		t.Fatalf("txtdata.LoadClassTables: %v", err)
	}

	// then
	if len(tables) != 1 || tables[0][0] != "classFlags" || tables[0][1] != "CriticalStrike" {
		t.Fatalf("tables = %v", tables)
	}
}

// The warrior's attributes.tsv row: the stat baselines and the four combat
// chances come from one file (playerdat.cpp:169-205), with the decimals in
// 1/64 units.
func TestLoadClassAttributesReadsWarriorRow(t *testing.T) {
	// given
	tbl := classFixture(t, "warrior", "attributes.tsv")
	// when
	attrs, combat, err := txtdata.LoadClassAttributes(tbl)
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadClassAttributes: %v", err)
	}

	want := txtdata.ClassAttributes{
		ClassFlags: txtdata.ClassFlagCriticalStrike,
		BaseStr:    30, BaseMag: 10, BaseDex: 20, BaseVit: 25,
		MaxStr: 250, MaxMag: 50, MaxDex: 60, MaxVit: 100,
		AdjLife: 1152, AdjMana: -64, LvlLife: 128, LvlMana: 64,
		ChrLife: 128, ChrMana: 64, ItmLife: 128, ItmMana: 64,
		ManaCost: 64, ItmRestoreLife: 128, ItmRestoreMana: 64,
		SplRestoreLife: 128, SplRestoreMana: 64, HealOtherRestoreLife: 128,
	}
	if attrs != want {
		t.Fatalf("attributes = %+v, want %+v", attrs, want)
	}

	wantCombat := txtdata.PlayerCombatData{
		BaseToBlock: 30, BaseMagicToHit: 50, BaseMeleeToHit: 70, BaseRangedToHit: 60,
	}
	if combat != wantCombat {
		t.Fatalf("combat = %+v, want %+v", combat, wantCombat)
	}
}

// classFlags is a comma-separated PlayerClassFlag list OR'd into one bitset
// (playerdat.cpp:178, data/iterators.hpp:201-212); an empty cell is no flags.
func TestLoadClassAttributesParsesClassFlags(t *testing.T) {
	tests := []struct {
		name  string
		class string
		want  txtdata.ClassFlag
	}{
		{name: "warriorCriticalStrike", class: "warrior", want: txtdata.ClassFlagCriticalStrike},
		{name: "rogueTrapSense", class: "rogue", want: txtdata.ClassFlagTrapSense},
		{name: "bardDualWield", class: "bard", want: txtdata.ClassFlagDualWield},
		{
			name:  "barbarianThreeFlags",
			class: "barbarian",
			want:  txtdata.ClassFlagCriticalStrike | txtdata.ClassFlagIronSkin | txtdata.ClassFlagNaturalResistance,
		},
		{name: "sorcererNone", class: "sorcerer", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			tbl := classFixture(t, tt.class, "attributes.tsv")
			// when
			got := assertClassFlags(t, tbl)
			// then
			if got != tt.want {
				t.Fatalf("classFlags = %d, want %d", got, tt.want)
			}
		})
	}
}

// Decimal columns parse to 1/64 units rounded to the nearest step
// (utils/parse_int.hpp:47-114); the bard carries 1.5, 1.75 and 0.75.
func TestLoadClassAttributesParsesFractionalValues(t *testing.T) {
	// given
	tbl := classFixture(t, "bard", "attributes.tsv")
	// when
	attrs, _, err := txtdata.LoadClassAttributes(tbl)
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadClassAttributes: %v", err)
	}

	want := txtdata.ClassAttributes{
		ClassFlags: txtdata.ClassFlagDualWield,
		BaseStr:    20, BaseMag: 20, BaseDex: 25, BaseVit: 20,
		MaxStr: 120, MaxMag: 120, MaxDex: 120, MaxVit: 100,
		AdjLife: 1472, AdjMana: 192, LvlLife: 128, LvlMana: 128,
		ChrLife: 64, ChrMana: 96, ItmLife: 96, ItmMana: 112,
		ManaCost: 48, ItmRestoreLife: 96, ItmRestoreMana: 96,
		SplRestoreLife: 96, SplRestoreMana: 96, HealOtherRestoreLife: 96,
	}
	if attrs != want {
		t.Fatalf("attributes = %+v, want %+v", attrs, want)
	}
}

// The warrior's starting_loadout.tsv: skill, spell, five items and the gold
// stack (playerdat.cpp:207-223).
func TestLoadClassLoadoutReadsWarriorRow(t *testing.T) {
	// given
	tbl := classFixture(t, "warrior", "starting_loadout.tsv")
	// when
	loadout, err := txtdata.LoadClassLoadout(tbl)
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadClassLoadout: %v", err)
	}

	want := txtdata.ClassLoadout{
		Skill: "ItemRepair", Spell: "Null", SpellLevel: 0,
		Items: [5]string{"IDI_WARRIOR", "IDI_WARRSHLD", "IDI_WARRCLUB", "IDI_HEAL", "IDI_HEAL"},
		Gold:  100,
	}
	if loadout != want {
		t.Fatalf("loadout = %+v, want %+v", loadout, want)
	}
}

// A class with a starting spell carries its name and level, and the unused item
// slots keep the IDI_NONE sentinel (sorcerer loadout).
func TestLoadClassLoadoutReadsSorcererStartingSpell(t *testing.T) {
	// given
	tbl := classFixture(t, "sorcerer", "starting_loadout.tsv")
	// when
	loadout, err := txtdata.LoadClassLoadout(tbl)
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadClassLoadout: %v", err)
	}

	want := txtdata.ClassLoadout{
		Skill: "StaffRecharge", Spell: "Firebolt", SpellLevel: 2,
		Items: [5]string{"IDI_SORCERER_DIABLO", "IDI_MANA", "IDI_MANA", "IDI_NONE", "IDI_NONE"},
		Gold:  100,
	}
	if loadout != want {
		t.Fatalf("loadout = %+v, want %+v", loadout, want)
	}
}

// The warrior's animations.tsv: 27 frame counts and action frames in the
// reader's own order (playerdat.cpp:255-291).
func TestLoadClassAnimationsReadsWarriorRow(t *testing.T) {
	// given
	tbl := classFixture(t, "warrior", "animations.tsv")
	// when
	anim, err := txtdata.LoadClassAnimations(tbl)
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadClassAnimations: %v", err)
	}

	want := txtdata.ClassAnimations{
		UnarmedFrames: 16, UnarmedActionFrame: 9, UnarmedShieldFrames: 16, UnarmedShieldActionFrame: 9,
		SwordFrames: 16, SwordActionFrame: 9, SwordShieldFrames: 16, SwordShieldActionFrame: 9,
		BowFrames: 16, BowActionFrame: 11, AxeFrames: 20, AxeActionFrame: 10,
		MaceFrames: 16, MaceActionFrame: 9, MaceShieldFrames: 16, MaceShieldActionFrame: 9,
		StaffFrames: 16, StaffActionFrame: 11, IdleFrames: 10, WalkingFrames: 8,
		BlockingFrames: 2, DeathFrames: 20, CastingFrames: 20, RecoveryFrames: 6,
		TownIdleFrames: 20, TownWalkingFrames: 8, CastingActionFrame: 14,
	}
	if anim != want {
		t.Fatalf("animations = %+v, want %+v", anim, want)
	}
}

// The warrior's sprites.tsv: the class path, letter, TRN, three portrait
// offsets and the ten sprite widths (playerdat.cpp:225-253).
func TestLoadClassSpritesReadsWarriorRow(t *testing.T) {
	// given
	tbl := classFixture(t, "warrior", "sprites.tsv")
	// when
	sprites, err := txtdata.LoadClassSprites(tbl)
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadClassSprites: %v", err)
	}

	want := txtdata.ClassSprites{
		ClassPath: "warrior", ClassChar: 'w', TRN: "warrior",
		PartyOffsetTown:    txtdata.ClassPoint{X: -4, Y: -18},
		PartyOffsetDungeon: txtdata.ClassPoint{X: 6, Y: -21},
		PartyOffsetDead:    txtdata.ClassPoint{X: -6, Y: -50},
		Stand:              96, Walk: 96, Attack: 128, Bow: 96, SwHit: 96, Block: 96,
		Lightning: 96, Fire: 96, Magic: 96, Death: 128,
	}
	if sprites != want {
		t.Fatalf("sprites = %+v, want %+v", sprites, want)
	}
}

// The monk row proves the loader reads a wider sprite set than the warrior's
// defaults: 112/130/160 px (playerdat.cpp:225-253).
func TestLoadClassSpritesReadsMonkWidths(t *testing.T) {
	// given
	tbl := classFixture(t, "monk", "sprites.tsv")
	// when
	sprites, err := txtdata.LoadClassSprites(tbl)
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadClassSprites: %v", err)
	}

	want := txtdata.ClassSprites{
		ClassPath: "monk", ClassChar: 'm', TRN: "monk",
		PartyOffsetTown:    txtdata.ClassPoint{X: -2, Y: -19},
		PartyOffsetDungeon: txtdata.ClassPoint{X: 1, Y: -19},
		PartyOffsetDead:    txtdata.ClassPoint{X: 28, Y: -60},
		Stand:              112, Walk: 112, Attack: 130, Bow: 130, SwHit: 98, Block: 98,
		Lightning: 114, Fire: 114, Magic: 114, Death: 160,
	}
	if sprites != want {
		t.Fatalf("sprites = %+v, want %+v", sprites, want)
	}
}

// Every combat-field loader reads all six committed classes: the per-class files
// carry the same key set, so a loader that reads one class reads all six
// (playerdat.cpp:378-396 loops the classdat rows).
func TestLoadClassAttributesReadsEveryCommittedClass(t *testing.T) {
	for _, class := range committedClasses {
		t.Run(class, func(t *testing.T) {
			// given
			tbl := classFixture(t, class, "attributes.tsv")
			// when
			err := assertAttributesUsable(t, tbl)
			// then the row carries non-zero baselines
			if err != nil {
				t.Fatalf("attributes: %v", err)
			}
		})
	}
}

// Every loadout loader reads all six committed classes, each with a skill, a
// first item and gold.
func TestLoadClassLoadoutReadsEveryCommittedClass(t *testing.T) {
	for _, class := range committedClasses {
		t.Run(class, func(t *testing.T) {
			// given
			tbl := classFixture(t, class, "starting_loadout.tsv")
			// when
			err := assertLoadoutUsable(t, tbl)
			// then the row carries a skill, a first item and gold
			if err != nil {
				t.Fatalf("loadout: %v", err)
			}
		})
	}
}

// Every animation loader reads all six committed classes.
func TestLoadClassAnimationsReadsEveryCommittedClass(t *testing.T) {
	for _, class := range committedClasses {
		t.Run(class, func(t *testing.T) {
			// given
			tbl := classFixture(t, class, "animations.tsv")
			// when
			err := assertAnimationsUsable(t, tbl)
			// then the row carries non-zero frame counts
			if err != nil {
				t.Fatalf("animations: %v", err)
			}
		})
	}
}

// Every sprite loader reads all six committed classes.
func TestLoadClassSpritesReadsEveryCommittedClass(t *testing.T) {
	for _, class := range committedClasses {
		t.Run(class, func(t *testing.T) {
			// given
			tbl := classFixture(t, class, "sprites.tsv")
			// when
			err := assertSpritesUsable(t, tbl)
			// then the row carries a path, a letter and widths
			if err != nil {
				t.Fatalf("sprites: %v", err)
			}
		})
	}
}

// Every loader walks the full upstream key set: dropping any record leaves the
// walk one field short and fails with the missing column's name
// (value_reader.cpp:16-20 "Missing field").
func TestLoadersRejectMissingColumn(t *testing.T) {
	tests := []struct {
		name  string
		class string
		file  string
		drop  string
		kind  string
	}{
		{
			name: "attributesFirstKey", class: "warrior", file: "attributes.tsv", drop: "classFlags",
			kind: classAttributes,
		},
		{
			name: "attributesLastKey", class: "warrior", file: "attributes.tsv", drop: "baseRangedToHit",
			kind: classAttributes,
		},
		{
			name: "loadoutMiddleKey", class: "warrior", file: "starting_loadout.tsv", drop: "spell",
			kind: classLoadout,
		},
		{
			name: "loadoutLastKey", class: "warrior", file: "starting_loadout.tsv", drop: "gold",
			kind: classLoadout,
		},
		{
			name: "animationsLastKey", class: "warrior", file: "animations.tsv", drop: "castingActionFrame",
			kind: classAnimations,
		},
		{
			name: "spritesMiddleKey", class: "warrior", file: "sprites.tsv", drop: "trn",
			kind: classSprites,
		},
		{
			name: "spritesLastKey", class: "warrior", file: "sprites.tsv", drop: "death",
			kind: classSprites,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			tbl := withoutRow(classFixture(t, tt.class, tt.file), tt.drop)
			// when
			err := loadClassTable(tt.kind, tbl)
			// then
			requireErrorNames(t, err, tt.drop)
		})
	}
}

// A renamed key is an unexpected field, not a shifted value: the error names
// both the column found and the column expected (value_reader.cpp:21-24).
func TestLoadClassAttributesRejectsUnknownColumn(t *testing.T) {
	// given
	tbl := withRowKey(classFixture(t, "warrior", "attributes.tsv"), "maxDex", "maxDexterity")
	// when
	err := loadClassTable(classAttributes, tbl)
	// then
	if err == nil {
		t.Fatal("load succeeded, want an unknown-column error")
	}

	if !strings.Contains(err.Error(), "maxDexterity") || !strings.Contains(err.Error(), "maxDex") {
		t.Fatalf("error %q does not name the found and expected columns", err)
	}
}

// A record without a value field is DataFile::Error::NotEnoughColumns
// (value_reader.cpp:26-28); the error names the column.
func TestLoadClassLoadoutRejectsShortRow(t *testing.T) {
	// given
	tbl := withRowFields(classFixture(t, "warrior", "starting_loadout.tsv"), "gold", []string{"gold"})
	// when
	err := loadClassTable(classLoadout, tbl)
	// then
	if err == nil {
		t.Fatal("load succeeded, want a short-row error")
	}

	if !strings.Contains(err.Error(), "gold") {
		t.Fatalf("error %q does not name the short column", err)
	}
}

// An unknown PlayerClassFlag name fails the load and names the value:
// ParsePlayerClassFlag rejects anything magic_enum cannot map
// (playerdat.cpp:160-167).
func TestLoadClassAttributesRejectsUnknownClassFlag(t *testing.T) {
	// given
	tbl := withRowValue(classFixture(t, "warrior", "attributes.tsv"), "classFlags", "CriticalStrike,SpellSlinger")
	// when
	err := loadClassTable(classAttributes, tbl)
	// then
	if err == nil {
		t.Fatal("load succeeded, want an unknown-flag error")
	}

	if !strings.Contains(err.Error(), "SpellSlinger") {
		t.Fatalf("error %q does not name the unknown flag", err)
	}
}

// Trailing records past the last expected key are ignored: upstream reads the
// fixed key list and never looks at what follows (value_reader.cpp:16-28).
func TestLoadClassLoadoutIgnoresTrailingRows(t *testing.T) {
	// given
	tbl := withAppendedRow(classFixture(t, "warrior", "starting_loadout.tsv"), "futureKey", "futureValue")
	// when
	loadout, err := txtdata.LoadClassLoadout(tbl)
	// then
	if err != nil {
		t.Fatalf("txtdata.LoadClassLoadout: %v", err)
	}

	if loadout.Gold != 100 {
		t.Fatalf("gold = %d, want 100", loadout.Gold)
	}
}

// The decimal column rounds a fraction to the nearest 1/64: upstream's own
// worked example parses 6.34 as 6<<6 + 22 = 406 (test/data_file_test.cpp:266-270).
func TestLoadClassAttributesRoundsFractionToNearestStep(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{name: "thirdsRoundsUp", value: "6.34", want: 406},
		{name: "exactHalf", value: "1.5", want: 96},
		{name: "exactQuarter", value: "0.75", want: 48},
		{name: "roundsDown", value: "2.007", want: 128},
		{name: "integer", value: "3", want: 192},
		{name: "negativeFraction", value: "-1.5", want: -96},
		{name: "bareFraction", value: ".5", want: 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given adjLife carries the value under test
			tbl := withRowValue(classFixture(t, "warrior", "attributes.tsv"), "adjLife", tt.value)
			// when
			got, err := assertAdjLife(t, tbl)
			// then
			requireAdjLife(t, got, err, tt.value, tt.want)
		})
	}
}

// A decimal cell the fixed-point parser cannot read fails the load and names
// the offending column (DataFile::reportFatalFieldError, value_reader.cpp:30).
func TestLoadClassAttributesRejectsBadDecimal(t *testing.T) {
	// given
	tbl := withRowValue(classFixture(t, "warrior", "attributes.tsv"), "manaCost", "abc")
	// when
	err := loadClassTable(classAttributes, tbl)
	// then
	if err == nil {
		t.Fatal("load succeeded, want a parse error")
	}

	if !strings.Contains(err.Error(), "manaCost") {
		t.Fatalf("error %q does not name the offending column", err)
	}
}

// An integer column rejects a non-numeric cell and names the column.
func TestLoadClassAnimationsRejectsBadInteger(t *testing.T) {
	// given
	tbl := withRowValue(classFixture(t, "warrior", "animations.tsv"), "idleFrames", "ten")
	// when
	err := loadClassTable(classAnimations, tbl)
	// then
	if err == nil {
		t.Fatal("load succeeded, want a parse error")
	}

	if !strings.Contains(err.Error(), "idleFrames") {
		t.Fatalf("error %q does not name the offending column", err)
	}
}

// A classChar cell must hold exactly one character
// (ValueReader::readChar, value_reader.hpp:103-111).
func TestLoadClassSpritesRejectsMultiCharClassChar(t *testing.T) {
	// given
	tbl := withRowValue(classFixture(t, "warrior", "sprites.tsv"), "classChar", "wa")
	// when
	err := loadClassTable(classSprites, tbl)
	// then
	if err == nil {
		t.Fatal("load succeeded, want a single-character error")
	}

	if !strings.Contains(err.Error(), "classChar") {
		t.Fatalf("error %q does not name the offending column", err)
	}
}

// The class flag bitset matches PlayerClassFlag's bit assignments
// (playerdat.hpp:28-36).
func TestClassFlagValuesMatchUpstream(t *testing.T) {
	tests := []struct {
		name string
		flag txtdata.ClassFlag
		want uint8
	}{
		{name: "CriticalStrike", flag: txtdata.ClassFlagCriticalStrike, want: 1},
		{name: "DualWield", flag: txtdata.ClassFlagDualWield, want: 2},
		{name: "IronSkin", flag: txtdata.ClassFlagIronSkin, want: 4},
		{name: "NaturalResistance", flag: txtdata.ClassFlagNaturalResistance, want: 8},
		{name: "TrapSense", flag: txtdata.ClassFlagTrapSense, want: 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			got := uint8(tt.flag)
			// then
			if got != tt.want {
				t.Fatalf("flag = %d, want %d", got, tt.want)
			}
		})
	}
}

// loadClassTable runs the loader named kind over tbl.
func loadClassTable(kind string, tbl *enginetxtdata.Table) error {
	load, ok := classLoaders[kind]
	if !ok {
		return &loaderError{kind: kind, err: errors.New("unknown class table")}
	}

	return load(tbl)
}

// loadAttributes runs LoadClassAttributes and keeps its error.
func loadAttributes(tbl *enginetxtdata.Table) error {
	_, _, err := txtdata.LoadClassAttributes(tbl)
	if err != nil {
		return fmt.Errorf("%s: %w", classAttributes, err)
	}

	return nil
}

// loadLoadout runs LoadClassLoadout and keeps its error.
func loadLoadout(tbl *enginetxtdata.Table) error {
	_, err := txtdata.LoadClassLoadout(tbl)
	if err != nil {
		return fmt.Errorf("%s: %w", classLoadout, err)
	}

	return nil
}

// loadAnimations runs LoadClassAnimations and keeps its error.
func loadAnimations(tbl *enginetxtdata.Table) error {
	_, err := txtdata.LoadClassAnimations(tbl)
	if err != nil {
		return fmt.Errorf("%s: %w", classAnimations, err)
	}

	return nil
}

// loadSprites runs LoadClassSprites and keeps its error.
func loadSprites(tbl *enginetxtdata.Table) error {
	_, err := txtdata.LoadClassSprites(tbl)
	if err != nil {
		return fmt.Errorf("%s: %w", classSprites, err)
	}

	return nil
}

// requireErrorNames requires err to be non-nil and to name want.
func requireErrorNames(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("load succeeded, want an error naming %q", want)
	}

	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q does not name %q", err, want)
	}
}

// requireAdjLife requires the parsed adjLife to equal want.
func requireAdjLife(t *testing.T, got int, err error, cell string, want int) {
	t.Helper()

	if err != nil {
		t.Fatalf("load attributes: %v", err)
	}

	if got != want {
		t.Fatalf("adjLife %q = %d, want %d", cell, got, want)
	}
}

// assertClassFlags loads tbl and returns the row's class flag bitset.
func assertClassFlags(t *testing.T, tbl *enginetxtdata.Table) txtdata.ClassFlag {
	t.Helper()

	attrs, _, err := txtdata.LoadClassAttributes(tbl)
	if err != nil {
		t.Fatalf("txtdata.LoadClassAttributes: %v", err)
	}

	return attrs.ClassFlags
}

// assertAdjLife loads tbl and returns the row's parsed adjLife.
func assertAdjLife(t *testing.T, tbl *enginetxtdata.Table) (int, error) {
	t.Helper()

	attrs, _, err := txtdata.LoadClassAttributes(tbl)
	if err != nil {
		return 0, fmt.Errorf("load attributes: %w", err)
	}

	return attrs.AdjLife, nil
}

// assertAttributesUsable requires the row to carry non-zero stat baselines.
func assertAttributesUsable(t *testing.T, tbl *enginetxtdata.Table) error {
	t.Helper()

	attrs, combat, err := txtdata.LoadClassAttributes(tbl)
	if err != nil {
		return fmt.Errorf("load attributes: %w", err)
	}

	if attrs.BaseStr == 0 || combat.BaseMeleeToHit == 0 {
		return fmt.Errorf("attributes = %+v / %+v, want non-zero baselines", attrs, combat)
	}

	return nil
}

// assertLoadoutUsable requires the row to carry a skill, a first item and gold.
func assertLoadoutUsable(t *testing.T, tbl *enginetxtdata.Table) error {
	t.Helper()

	loadout, err := txtdata.LoadClassLoadout(tbl)
	if err != nil {
		return fmt.Errorf("load loadout: %w", err)
	}

	if loadout.Skill == "" || loadout.Items[0] == "" || loadout.Gold == 0 {
		return fmt.Errorf("loadout = %+v, want a skill, a first item and gold", loadout)
	}

	return nil
}

// assertAnimationsUsable requires the row to carry non-zero frame counts.
func assertAnimationsUsable(t *testing.T, tbl *enginetxtdata.Table) error {
	t.Helper()

	anim, err := txtdata.LoadClassAnimations(tbl)
	if err != nil {
		return fmt.Errorf("load animations: %w", err)
	}

	if anim.UnarmedFrames == 0 || anim.CastingActionFrame == 0 || anim.TownWalkingFrames == 0 {
		return fmt.Errorf("animations = %+v, want non-zero frame counts", anim)
	}

	return nil
}

// assertSpritesUsable requires the row to carry a path, a letter and widths.
func assertSpritesUsable(t *testing.T, tbl *enginetxtdata.Table) error {
	t.Helper()

	sprites, err := txtdata.LoadClassSprites(tbl)
	if err != nil {
		return fmt.Errorf("load sprites: %w", err)
	}

	if sprites.ClassPath == "" || sprites.ClassChar == 0 || sprites.Stand == 0 || sprites.Death == 0 {
		return fmt.Errorf("sprites = %+v, want a path, a letter and widths", sprites)
	}

	return nil
}

// Error renders the table name and the underlying loader message.
func (e *loaderError) Error() string {
	return e.kind + ": " + e.err.Error()
}

// Unwrap exposes the loader's own error for errors.Is and errors.As.
func (e *loaderError) Unwrap() error {
	return e.err
}

// classFixture loads one committed upstream class table
// (internal/testdata/assets/txtdata/upstream/classes/<class>/<name>.tsv). A
// missing fixture fails the test: the zero-dependency contract never skips.
func classFixture(t *testing.T, class, name string) *enginetxtdata.Table {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(
		"..", "..", "..", "internal", "testdata", "assets", "txtdata", "upstream", "classes", class, name,
	))
	if err != nil {
		t.Fatalf("read committed %s/%s: %v", class, name, err)
	}

	return mustParse(t, data)
}

// withoutRow returns a copy of tbl with the record whose key is key removed.
func withoutRow(tbl *enginetxtdata.Table, key string) *enginetxtdata.Table {
	out := &enginetxtdata.Table{Header: tbl.Header}

	for _, row := range tbl.Rows {
		if len(row) > 0 && row[0] == key {
			continue
		}

		out.Rows = append(out.Rows, row)
	}

	return out
}

// withRowValue returns a copy of tbl with the keyed record's value replaced.
func withRowValue(tbl *enginetxtdata.Table, key, value string) *enginetxtdata.Table {
	return withRowFields(tbl, key, []string{key, value})
}

// withRowKey returns a copy of tbl with the keyed record's key renamed.
func withRowKey(tbl *enginetxtdata.Table, from, to string) *enginetxtdata.Table {
	out := &enginetxtdata.Table{Header: tbl.Header, Rows: append([][]string(nil), tbl.Rows...)}

	for i, row := range out.Rows {
		if len(row) > 0 && row[0] == from {
			out.Rows[i] = []string{to, row[1]}
		}
	}

	return out
}

// withRowFields replaces the record keyed key with the given fields.
func withRowFields(tbl *enginetxtdata.Table, key string, fields []string) *enginetxtdata.Table {
	out := &enginetxtdata.Table{Header: tbl.Header, Rows: append([][]string(nil), tbl.Rows...)}

	for i, row := range out.Rows {
		if len(row) > 0 && row[0] == key {
			out.Rows[i] = fields
		}
	}

	return out
}

// withAppendedRow appends a key/value record to the table.
func withAppendedRow(tbl *enginetxtdata.Table, key, value string) *enginetxtdata.Table {
	out := &enginetxtdata.Table{Header: tbl.Header, Rows: append([][]string(nil), tbl.Rows...)}
	out.Rows = append(out.Rows, []string{key, value})

	return out
}
