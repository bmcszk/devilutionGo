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
		})
	}
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
	m := list.AddMissile(missiles.AddMissileArgs{
		Src:     missiles.TilePosition{X: 5, Y: 5},
		Dst:     missiles.TilePosition{X: 5, Y: 9},
		Midir:   missiles.West,
		Type:    missiles.Healing,
		Caster:  int(missiles.TargetMonsters),
		Source:  0,
		Graphic: missiles.GraphicID(ordinal),
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
