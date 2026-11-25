//go:build windowed

package main

import (
	"context"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game/ui"
)

// TestWindowUpdateSettlesTheCreatedHero drives (*windowGame).Update — the REAL
// window loop — through a create: after the pick and N Updates, the store row
// EXISTS (dcob D10 / RF1). The front-flow screens must not starve the lane's
// frame, or the pending create never settles.
func TestWindowUpdateSettlesTheCreatedHero(t *testing.T) {
	boot := newWindowBoot(t)

	// walk title -> menu -> selhero and open the create flow
	boot.flow.HandleKey(ui.KeySpace)
	boot.flow.HandleKey(ui.KeyEnter)

	if err := boot.selhero.beginCreate(); err != nil {
		t.Fatalf("beginCreate: %v", err)
	}

	if err := boot.selhero.typeName("Settled Hero"); err != nil {
		t.Fatalf("typeName: %v", err)
	}

	boot.selhero.confirmName()
	boot.selhero.classMenu.Focus()
	boot.selhero.classMenu.HandleKey(ui.KeyEnter) // pick warrior

	// drive the REAL Update loop: the create settles within N frames
	const settleFrames = 10
	for range settleFrames {
		if err := boot.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	rows, err := boot.store.ListHeroes(context.Background())
	if err != nil {
		t.Fatalf("ListHeroes: %v", err)
	}

	if len(rows) != 1 || rows[0].Name != "Settled Hero" {
		t.Fatalf("store rows after create + %d Updates = %v, want 1 (Settled Hero)", settleFrames, rows)
	}
}
