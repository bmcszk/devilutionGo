//go:build windowed

package main

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// TestSelHeroCreatesEveryCommittedClass pins the pack law: every committed
// classdat row creates — exactly one CreatedPack with the right class
// ordinal and the data-built pack (dcob D7).
func TestSelHeroCreatesEveryCommittedClass(t *testing.T) {
	for ordinal, name := range classNamesOfCommittedClasses(t) {
		t.Run(name, func(t *testing.T) {
			boot := newSelHeroTestBoot(t)
			boot.handleKey(ui.KeySpace)
			boot.handleKey(ui.KeyEnter)

			if err := boot.selhero.beginCreate(); err != nil {
				t.Fatalf("beginCreate: %v", err)
			}

			if err := boot.selhero.typeName("Hero " + name); err != nil {
				t.Fatalf("typeName: %v", err)
			}

			boot.selhero.confirmName()

			// walk the class menu to the ordinal's row and pick it
			boot.selhero.classMenu.Focus()
			for range ordinal {
				boot.selhero.classMenu.HandleKey(ui.KeyArrowDown)
			}

			boot.selhero.classMenu.HandleKey(ui.KeyEnter)
			boot.frame()

			events := boot.createdEvents()
			if len(events) != 1 {
				t.Fatalf("CreatedPack events = %d (%v), want exactly one", len(events), events)
			}

			if events[0].A != int32(ordinal) {
				t.Fatalf("CreatedPack class = %d, want the classdat ordinal %d", events[0].A, ordinal)
			}

			pack, _, ok := boot.lane.state.RaisedPack()
			if !ok {
				t.Fatalf("kernel raised no pack for %s", name)
			}

			want := transcribedClassTable[ordinal]
			if pack != want.pack(ordinal) {
				t.Fatalf("raised pack = %+v, want the transcribed %+v", pack, want.pack(ordinal))
			}
		})
	}
}

// transcribedClassRow is one hand-transcribed class row: the level-1 bases
// and the raw vitals computed by hand from the committed numbers (D8 parity:
// expected values are literals, not loader output). The vitals are in the
// 1/64 raw scale upstream stores (player.h:236-243; pack.cpp compares the
// .sv bytes against the raw fixed-point), each computed as
// adj + lvl*1 + chr*base from the row's own adjMana/adjLife, lvlMana/lvlLife,
// chrMana/chrLife and baseVit/baseMag — e.g. warrior HP
// 18+2+2*25 = 70 -> 4480, sorcerer mana -2+2+2*35 = 70 -> 4480.
type transcribedClassRow struct {
	baseStr, baseMag, baseDex, baseVit int
	hpRaw, manaRaw                     int
}

// pack folds the row into the StartingPack shape the pack source serves
// (class = the classdat ordinal).
func (r transcribedClassRow) pack(class int) world.StartingPack {
	return world.StartingPack{
		Class: class, BaseStr: r.baseStr, BaseMag: r.baseMag,
		BaseDex: r.baseDex, BaseVit: r.baseVit,
		HPBase: r.hpRaw, MaxHPBase: r.hpRaw,
		ManaBase: r.manaRaw, MaxManaBase: r.manaRaw,
	}
}

// transcribedClassTable is the six committed classes transcribed by hand from
// internal/testdata/assets/txtdata/upstream/classes/<dir>/attributes.tsv, in
// the committed classdat.tsv row order. Level is pinned at 1 by the vitals
// formula above (the preview struct carries no level field — the screen
// always previews a level-1 hero).
var transcribedClassTable = [...]transcribedClassRow{
	{30, 10, 20, 25, 4480, 640},  // attributes.tsv: warrior:3-6,12-17
	{20, 15, 30, 20, 2880, 1440}, // attributes.tsv: rogue:3-6,12-17
	{15, 35, 15, 20, 1920, 4480}, // attributes.tsv: sorcerer:3-6,12-17
	{25, 15, 25, 20, 2880, 1440}, // attributes.tsv: monk:3-6,12-17
	{20, 20, 25, 20, 2880, 2240}, // attributes.tsv: bard:3-6,12-17
	{40, 0, 20, 25, 4480, 0},     // attributes.tsv: barbarian:3-6,12-17
}

// TestSelHeroPreviewMatchesPlayerdatBases pins the preview parity law: the
// screen's preview for each class reads the level-1 baseStr/baseMag/baseDex/
// baseVit of the committed loaders — the pfile_ui_set_class_stats
// transcription (pfile.cpp:724-732 @e1ad0c2b6) against the hand-transcribed
// transcribedClassTable. 6x6 table (bases + raw vitals).
func TestSelHeroPreviewMatchesPlayerdatBases(t *testing.T) {
	boot := newSelHeroTestBoot(t)
	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("Preview"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()

	for ordinal, name := range classNamesOfCommittedClasses(t) {
		t.Run(name, func(t *testing.T) {
			attrs, err := boot.selhero.classPreview(ordinal)
			if err != nil {
				t.Fatalf("classPreview %s: %v", name, err)
			}

			want := transcribedClassTable[ordinal]

			if attrs.BaseStr != want.baseStr || attrs.BaseMag != want.baseMag ||
				attrs.BaseDex != want.baseDex || attrs.BaseVit != want.baseVit {
				t.Fatalf("preview %s = str %d mag %d dex %d vit %d, want the transcribed row %d/%d/%d/%d",
					name, attrs.BaseStr, attrs.BaseMag, attrs.BaseDex, attrs.BaseVit,
					want.baseStr, want.baseMag, want.baseDex, want.baseVit)
			}
		})
	}
}

// classNamesOfCommittedClasses reads the committed classdat.tsv row order: the
// six class names, index = class ordinal.
func classNamesOfCommittedClasses(t *testing.T) []string {
	t.Helper()

	rows := committedClassTable(t, "classes", "classdat.tsv")
	classes, err := txtdata.LoadClasses(rows)
	if err != nil {
		t.Fatalf("LoadClasses: %v", err)
	}

	names := make([]string, 0, len(classes))
	for _, c := range classes {
		names = append(names, c.ClassName)
	}

	return names
}
