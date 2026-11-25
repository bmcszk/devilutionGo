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

			want := dataBuiltPack(t, ordinal)
			if pack != want {
				t.Fatalf("raised pack = %+v, want the data-built %+v", pack, want)
			}
		})
	}
}

// TestSelHeroPreviewMatchesPlayerdatBases pins the preview parity law: the
// screen's preview for each class reads the level-1 baseStr/baseMag/baseDex/
// baseVit of the committed loaders — the pfile_ui_set_class_stats
// transcription (pfile.cpp:724-732 @e1ad0c2b6). 6x5 table.
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

			want := dataBuiltPack(t, ordinal)

			if attrs.BaseStr != want.BaseStr || attrs.BaseMag != want.BaseMag ||
				attrs.BaseDex != want.BaseDex || attrs.BaseVit != want.BaseVit {
				t.Fatalf("preview %s = str %d mag %d dex %d vit %d, want the committed row %d/%d/%d/%d",
					name, attrs.BaseStr, attrs.BaseMag, attrs.BaseDex, attrs.BaseVit,
					want.BaseStr, want.BaseMag, want.BaseDex, want.BaseVit)
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

// dataBuiltPack builds the class's starting pack the D8 pack source builds:
// the committed attributes row's bases and the level-1 vitals the
// calculateBaseLife/calculateBaseMana transcription gives (adjLife +
// lvlLife*1 + chrLife*baseVit, Fixed10_6 raw >> classFixedShift).
func dataBuiltPack(t *testing.T, ordinal int) world.StartingPack {
	t.Helper()

	pack, ok := (classdatPackSource{}).StartingPack(ordinal)
	if !ok {
		t.Fatalf("classdatPackSource refused ordinal %d, want all six committed classes", ordinal)
	}

	return pack
}
