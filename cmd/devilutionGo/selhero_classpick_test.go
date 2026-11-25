//go:build windowed

package main

import (
	"context"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/persist"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// TestSelHeroClassPickRendersClassRows pins the class-step draw site: while
// the create flow sits at the class pick, the composed front frame draws one
// band row per class-pick row — and ZERO hero-list/action rows (xoap X3).
func TestSelHeroClassPickRendersClassRows(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	if err := boot.store.SaveHero(context.Background(), persist.HeroRecord{
		Slot: 0, Name: "Hero One", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	// walk into the class step of the create flow
	boot.handleKey(ui.KeyEnter) // hero row -> create flow
	boot.handleKey(ui.KeyEnter) // confirm the name -> class pick

	if boot.selhero.creating == nil || boot.selhero.creating.step != selheroClassStep {
		t.Fatalf("create flow not at the class step (creating=%v)", boot.selhero.creating)
	}

	// when the front frame composes at the class step
	surface := &render.Surface{Width: 640, Height: 480, Pixels: make([]uint8, 640*480)}
	if err := scene.ComposeFrontFrame(surface, boot.flow); err != nil {
		t.Fatalf("ComposeFrontFrame: %v", err)
	}

	// then the band carries the 6 class rows — NOT the 4 hero/action rows
	if rows := countFrontLabelRows(surface, 200, 440); rows != 6 {
		t.Fatalf("class-step band rows = %d, want the 6 class-pick rows (0 hero/action rows)", rows)
	}
}

// TestSelHeroRefusedCreateSurfacesAnErrorRow pins the error surface: once a
// create is refused, the composed front frame draws an extra band row for the
// refusal — the lastCreateError reader (xoap X4).
func TestSelHeroRefusedCreateSurfacesAnErrorRow(t *testing.T) {
	boot := newSelHeroTestBoot(t)

	boot.handleKey(ui.KeySpace)
	boot.handleKey(ui.KeyEnter)

	// a create the kernel refuses: the data-built source serves every committed
	// class (dcob D7), so the refusal is forced with an ordinal outside the
	// committed classdat rows (see the pickClass call below)
	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("Ghost Hero"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()
	boot.selhero.classMenu.Focus()

	// the data-built source serves every committed class (dcob D7), so the
	// refusal is forced with an ordinal no committed classdat row has
	boot.selhero.pickClass(len(classDirs))
	boot.frame()

	if boot.selhero.lastCreateError == nil {
		t.Fatalf("refused create left no lastCreateError, want the refusal recorded")
	}

	// when the front frame composes after the refusal (0 heroes: 3 action rows)
	surface := &render.Surface{Width: 640, Height: 480, Pixels: make([]uint8, 640*480)}
	if err := scene.ComposeFrontFrame(surface, boot.flow); err != nil {
		t.Fatalf("ComposeFrontFrame: %v", err)
	}

	// then the band carries the 3 action rows PLUS one error row
	if rows := countFrontLabelRows(surface, 200, 440); rows != 4 {
		t.Fatalf("band rows after refused create = %d, want 4 (3 actions + 1 error row)", rows)
	}
}
