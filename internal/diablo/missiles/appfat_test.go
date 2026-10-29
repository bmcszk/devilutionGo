package missiles_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/missiles"
	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// Ports missiles.cpp:2920 — a non-elemental type reaching the elemental-arrow
// explosion is app_fatal("wrong missile ID N"). The process handler cannot
// return an error (its signature is the upstream behavior-table one), so the
// trip goes through the injected sink instead of a panic.
func TestProcessElementalArrow_NonElementalType_ReportsAppFatal(t *testing.T) {
	// given
	var reported *appfat.FatalError

	w := &missiles.HitWorld{Grid: world.NewGrid(), HitCallbacks: newColSpies().callbacks()}
	w.ProcessCallbacks = missiles.NoopProcessCallbacks()
	w.ProcessCallbacks.Fatal = func(fatal *appfat.FatalError) { reported = fatal }

	m := freshMissile(64, 64)
	m.Type = missiles.Firebolt
	m.Source = -1 // trap: no player/monster damage source needed
	m.Duration = 1
	m.AnimType = int(missiles.GraphicArrow)
	m.AnimLen = 15
	m.Lid = missiles.NoLight
	m.Position.Velocity = missiles.PositionDisplacement{DeltaX: 32 << 16, DeltaY: 0}

	// when
	missiles.ProcessElementalArrow(w, &m)

	// then
	if reported == nil {
		t.Fatal("non-elemental type reached the explosion without an app_fatal report")
	}

	want := fmt.Sprintf("wrong missile ID %d", int(m.Type))
	if reported.Body != want {
		t.Fatalf("fatal body = %q, want %q (missiles.cpp:2920)", reported.Body, want)
	}
}
