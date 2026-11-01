// Package missiles_test pins the five wired rows against the committed misdat
// corpus: the addFn/processFn columns must name the ported bodies, and the
// row's graphic must produce the frame group AddMissile picks
// (missiles.cpp:2834-2837, animFAmt = the sprite's numFrames, misdat.cpp:152).
package missiles_test

import (
	"os"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
)

const (
	misdatPath  = "../../../internal/testdata/gamedata/misdat/misdat.tsv"
	spritesPath = "../../../internal/testdata/gamedata/misdat/missile_sprites.tsv"
)

// wiredRows is the slice-1 scope in misdat column terms.
var wiredRows = []wiredRow{
	{missiles.Arrow, "Arrow", "AddArrow", "ProcessArrow", "Arrow"},
	{missiles.Firebolt, "Firebolt", "AddFirebolt", "ProcessGenericProjectile", "Fireball"},
	{missiles.MagmaBall, "MagmaBall", "AddMagmaBall", "ProcessGenericProjectile", "MagmaBall"},
	{missiles.BloodStar, "BloodStar", "AddGenericMagicMissile", "ProcessGenericProjectile", "BloodStar"},
	{missiles.Acid, "Acid", "AddAcid", "ProcessGenericProjectile", "Acid"},
	{missiles.ChargedBolt, "ChargedBolt", "AddChargedBolt", "ProcessChargedBolt", "ChargedBolt"},
	{missiles.HolyBolt, "HolyBolt", "AddHolyBolt", "ProcessHolyBolt", "HolyBolt"},
}

// wiredRow pins one misdat row against the port.
type wiredRow struct {
	id        missiles.MissileID
	name      string
	addFn     string
	processFn string
	graphic   string
}

// TestMisdatRowsMatchWiredBodies: every wired id's corpus row names the ported
// addFn/processFn, and its graphic drives the frame group AddMissile picks.
func TestMisdatRowsMatchWiredBodies(t *testing.T) {
	// given
	rows := loadMisdat(t)
	sprites := loadSprites(t)

	// when
	observed := make([]int, 0, len(wiredRows))
	for _, row := range wiredRows {
		observed = append(observed, frameGroupOf(t, rows, sprites, row))
	}

	// then
	for i, row := range wiredRows {
		t.Run(row.name, func(t *testing.T) {
			assertRowColumns(t, rows, row)
			assertFrameGroupMatchesCorpus(t, sprites, row, observed[i])
			assertDrawFlagMatchesCorpus(t, rows, row)
		})
	}
}

// TestAddMissileInvisibleRowIsNotDrawn: a row flagged Invisible seeds
// DrawFlag false while a drawn row seeds true — the corpus rows ManaShield
// and Arrow pin both sides (tables/misdat.h:185-188).
func TestAddMissileInvisibleRowIsNotDrawn(t *testing.T) {
	// given
	rows := loadMisdat(t)
	sprites := loadSprites(t)

	// when
	invisible := addCorpusRow(t, rows, sprites, "ManaShield")
	drawn := addCorpusRow(t, rows, sprites, "Arrow")

	// then
	if invisible.DrawFlag {
		t.Fatal("ManaShield DrawFlag = true, want false (Invisible row)")
	}

	if !drawn.DrawFlag {
		t.Fatal("Arrow DrawFlag = false, want true")
	}
}

// TestAddMissileSeedsAnimationTiming: the row's frameDelay/frameLength land in
// AnimDelay/AnimLen at the frame group AddMissile picks (misdat.cpp:400-408).
// HolyBolt's delays differ between group 0 and the midir group, so the group
// actually indexes the columns.
func TestAddMissileSeedsAnimationTiming(t *testing.T) {
	// given
	rows := loadMisdat(t)
	sprites := loadSprites(t)
	row := misdatRow(t, rows, "HolyBolt")
	sprite := spriteRow(t, sprites, row.Graphic)
	ordinal, _ := spriteOrdinal(t, sprites, row.Graphic)
	group := int(missiles.West)

	list := &missiles.MissileList{}
	args := missiles.AddMissileArgs{
		Src:             missiles.TilePosition{X: 5, Y: 5},
		Dst:             missiles.TilePosition{X: 5, Y: 9},
		Midir:           missiles.West,
		Type:            missiles.HolyBolt,
		Caster:          int(missiles.TargetMonsters),
		Source:          0,
		Graphic:         missiles.GraphicID(ordinal),
		Flags:           row.Flags,
		AnimFrameAmount: sprite.NumFrames,
		FrameDelay:      sprite.FrameDelay,
		FrameLength:     sprite.FrameLength,
	}

	// when
	m := list.AddMissile(newSilentWorld(), args)

	// then
	if m == nil {
		t.Fatal("AddMissile returned nil for the HolyBolt row")
	}

	if sprite.FrameDelay[0] == sprite.FrameDelay[group] {
		t.Fatalf("corpus row keeps frameDelay %d across groups; pick another row",
			sprite.FrameDelay[0])
	}

	if m.MFNum != group {
		t.Fatalf("frame group = %d, want %d", m.MFNum, group)
	}

	if m.AnimDelay != sprite.FrameDelay[group] {
		t.Fatalf("AnimDelay = %d, want %d for group %d",
			m.AnimDelay, sprite.FrameDelay[group], group)
	}

	if m.AnimLen != sprite.FrameLength[group] {
		t.Fatalf("AnimLen = %d, want %d", m.AnimLen, sprite.FrameLength[group])
	}
}

// addCorpusRow appends one corpus row through AddMissile.
func addCorpusRow(t *testing.T, rows []misdat.MissileData,
	sprites []misdat.MissileSprite, name string,
) *missiles.Missile {
	t.Helper()

	row := misdatRow(t, rows, name)
	ordinal, _ := spriteOrdinal(t, sprites, row.Graphic)

	list := &missiles.MissileList{}
	m := list.AddMissile(newSilentWorld(), missiles.AddMissileArgs{
		Src:     missiles.TilePosition{X: 5, Y: 5},
		Dst:     missiles.TilePosition{X: 5, Y: 9},
		Midir:   missiles.West,
		Type:    missiles.Phasing,
		Caster:  int(missiles.TargetMonsters),
		Source:  0,
		Graphic: missiles.GraphicID(ordinal),
		Flags:   row.Flags,
	})
	if m == nil {
		t.Fatalf("%s: AddMissile returned nil", name)
	}

	return m
}

// spriteRow finds a sprite row by id.
func spriteRow(t *testing.T, sprites []misdat.MissileSprite, name string) misdat.MissileSprite {
	t.Helper()

	for _, sprite := range sprites {
		if sprite.ID == name {
			return sprite
		}
	}

	t.Fatalf("missile_sprites has no %s row", name)

	return misdat.MissileSprite{}
}

// loadMisdat parses the committed base misdat corpus.
func loadMisdat(t *testing.T) []misdat.MissileData {
	t.Helper()

	data, err := os.ReadFile(misdatPath)
	if err != nil {
		t.Fatalf("read %s: %v", misdatPath, err)
	}

	rows, err := misdat.ParseMisdat(data)
	if err != nil {
		t.Fatalf("misdat.ParseMisdat: %v", err)
	}

	return rows
}

// loadSprites parses the committed base missile_sprites corpus.
func loadSprites(t *testing.T) []misdat.MissileSprite {
	t.Helper()

	data, err := os.ReadFile(spritesPath)
	if err != nil {
		t.Fatalf("read %s: %v", spritesPath, err)
	}

	rows, err := misdat.ParseMissileSprites(data)
	if err != nil {
		t.Fatalf("misdat.ParseMissileSprites: %v", err)
	}

	return rows
}

// misdatRow finds a corpus row by id.
func misdatRow(t *testing.T, rows []misdat.MissileData, id string) misdat.MissileData {
	t.Helper()

	for _, row := range rows {
		if row.ID == id {
			return row
		}
	}

	t.Fatalf("misdat has no %s row", id)

	return misdat.MissileData{}
}

// spriteOrdinal returns the sprite's data row index — the MissileGraphicID
// ordinal — and its animFAmt (numFrames).
func spriteOrdinal(t *testing.T, sprites []misdat.MissileSprite,
	name string,
) (ordinal, frames int) {
	t.Helper()

	for i, sprite := range sprites {
		if sprite.ID == name {
			return i, sprite.NumFrames
		}
	}

	t.Fatalf("missile_sprites has no %s row", name)

	return 0, 0
}

// frameGroupOf runs one row's graphic through AddMissile on an unwired id so
// the stub addFn leaves the frame-group pick in place, and checks it against
// the corpus animFAmt.
func frameGroupOf(t *testing.T, rows []misdat.MissileData,
	sprites []misdat.MissileSprite, row wiredRow,
) int {
	t.Helper()

	graphic := misdatRow(t, rows, row.name).Graphic
	ordinal, frames := spriteOrdinal(t, sprites, graphic)

	list := &missiles.MissileList{}
	m := list.AddMissile(newSilentWorld(), missiles.AddMissileArgs{
		Src:             missiles.TilePosition{X: 5, Y: 5},
		Dst:             missiles.TilePosition{X: 5, Y: 9},
		Midir:           missiles.West,
		Type:            missiles.Healing,
		Caster:          int(missiles.TargetMonsters),
		Source:          0,
		Graphic:         missiles.GraphicID(ordinal),
		AnimFrameAmount: frames,
	})
	if m == nil {
		t.Fatalf("%s: AddMissile returned nil", row.name)
	}

	if frames >= 8 && m.MFNum != int(missiles.West) {
		t.Fatalf("%s: %d frame groups kept group %d, want %d",
			row.name, frames, m.MFNum, int(missiles.West))
	}

	if frames < 8 && m.MFNum != 0 {
		t.Fatalf("%s: %d frame groups followed midir to %d, want group 0",
			row.name, frames, m.MFNum)
	}

	return m.MFNum
}

// assertRowColumns checks the addFn/processFn/graphic of one corpus row.
func assertRowColumns(t *testing.T, rows []misdat.MissileData, row wiredRow) {
	t.Helper()

	got := misdatRow(t, rows, row.name)

	if got.AddFn != row.addFn {
		t.Fatalf("%s addFn = %s, want %s", row.name, got.AddFn, row.addFn)
	}

	if got.ProcessFn != row.processFn {
		t.Fatalf("%s processFn = %s, want %s", row.name, got.ProcessFn, row.processFn)
	}

	if got.Graphic != row.graphic {
		t.Fatalf("%s graphic = %s, want %s", row.name, got.Graphic, row.graphic)
	}
}

// assertFrameGroupMatchesCorpus re-derives the expectation from the corpus so
// the observed pick cannot drift from the committed data.
func assertFrameGroupMatchesCorpus(t *testing.T, sprites []misdat.MissileSprite,
	row wiredRow, observed int,
) {
	t.Helper()

	_, frames := spriteOrdinal(t, sprites, row.graphic)

	want := int(missiles.West)
	if frames < 8 {
		want = 0
	}

	if observed != want {
		t.Fatalf("%s frame group = %d, want %d for %d frame groups",
			row.name, observed, want, frames)
	}
}

// assertDrawFlagMatchesCorpus: AddMissile's seeded DrawFlag is the row's
// isDrawn() — false only for rows flagged Invisible (tables/misdat.h:185-188).
func assertDrawFlagMatchesCorpus(t *testing.T, rows []misdat.MissileData, row wiredRow) {
	t.Helper()

	misdatRow := misdatRow(t, rows, row.name)

	_, frames := spriteOrdinal(t, loadSprites(t), misdatRow.Graphic)

	list := &missiles.MissileList{}
	m := list.AddMissile(newSilentWorld(), missiles.AddMissileArgs{
		Src:     missiles.TilePosition{X: 5, Y: 5},
		Dst:     missiles.TilePosition{X: 5, Y: 9},
		Midir:   missiles.West,
		Type:    row.id,
		Caster:  int(missiles.TargetMonsters),
		Source:  0,
		Graphic: missiles.GraphicID(graphicOf(t, misdatRow.Graphic)),
		Flags:   misdatRow.Flags,
	})
	if m == nil {
		t.Fatalf("%s: AddMissile returned nil", row.name)
	}

	want := true
	for _, flag := range misdatRow.Flags {
		if flag == "Invisible" {
			want = false
		}
	}

	if m.DrawFlag != want {
		t.Fatalf("%s DrawFlag = %v, want %v (flags %v, %d frame groups)",
			row.name, m.DrawFlag, want, misdatRow.Flags, frames)
	}
}

// graphicOf resolves a sprite name to its MissileGraphicID ordinal.
func graphicOf(t *testing.T, name string) int {
	t.Helper()

	ordinal, _ := spriteOrdinal(t, loadSprites(t), name)

	return ordinal
}
