package game_test

import (
	"errors"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/game"
)

// doomCelPath is the cel upstream loads for the map of the stars (doom.cpp:25).
const doomCelPath = `items\map\mapztown`

func TestDoomPortal_Init_LoadsMapOfTheStarsSprite(t *testing.T) {
	// given
	var gotPath string

	var gotWidth int

	portal := game.NewDoomPortal(func(path string, width int) (int, error) {
		gotPath, gotWidth = path, width

		return 1, nil
	})

	// when
	err := portal.Init()
	// then
	if err != nil {
		t.Fatalf("Init() error = %v, want nil", err)
	}

	if gotPath != doomCelPath || gotWidth != 640 {
		t.Fatalf("LoadCel(%q, %d), want (%q, 640) per doom.cpp:25", gotPath, gotWidth, doomCelPath)
	}

	if !portal.Active() {
		t.Fatal("Active() = false after Init, want true (DoomFlag = true, doom.cpp:26)")
	}

	if portal.Frames() != 1 {
		t.Fatalf("Frames() = %d after Init, want 1 (doom.cpp:42 draws frame [0])", portal.Frames())
	}
}

func TestDoomPortal_LoadFailure_StaysInactive(t *testing.T) {
	// given
	loadErr := errors.New("mapztown.pcx missing")
	portal := game.NewDoomPortal(func(string, int) (int, error) { return 0, loadErr })

	// when
	err := portal.Init()

	// then
	if !errors.Is(err, loadErr) {
		t.Fatalf("Init() error = %v, want the loader error %v", err, loadErr)
	}

	// doom.cpp:25-26 sets DoomFlag only behind a successful LoadCel, which
	// aborts on a missing asset.
	if portal.Active() || portal.Frames() != 0 {
		t.Fatalf("Active()=%v Frames()=%d after a failed load, want false/0", portal.Active(), portal.Frames())
	}
}

func TestDoomPortal_Close_FreesSpriteAndFlag(t *testing.T) {
	// given
	portal := game.NewDoomPortal(func(string, int) (int, error) { return 1, nil })
	if err := portal.Init(); err != nil {
		t.Fatalf("Init() error = %v, want nil", err)
	}

	// when
	portal.Close()
	portal.Close() // doom_close is idempotent: DoomFlag = false, sprite = nullopt

	// then
	if portal.Active() {
		t.Fatal("Active() = true after Close, want false (doom.cpp:31)")
	}

	if portal.Frames() != 0 {
		t.Fatalf("Frames() = %d after Close, want 0 (sprite dropped, doom.cpp:32)", portal.Frames())
	}
}

func TestDoomPortal_InitAfterClose_ReloadsTheSprite(t *testing.T) {
	// given: using the map of doom again after Escape closes it re-runs doom_init
	// (items.cpp:4319-4321), which reassigns the cel and raises DoomFlag.
	loads := 0

	portal := game.NewDoomPortal(func(string, int) (int, error) {
		loads++

		return 1, nil
	})

	// when
	firstErr := portal.Init()
	portal.Close()
	secondErr := portal.Init()

	// then
	if firstErr != nil || secondErr != nil {
		t.Fatalf("Init() errors = %v, %v, want nil", firstErr, secondErr)
	}

	if loads != 2 {
		t.Fatalf("loader called %d times, want 2 (reopening reloads, doom.cpp:23-27)", loads)
	}

	if !portal.Active() || portal.Frames() != 1 {
		t.Fatalf("Active()=%v Frames()=%d after reopening, want true/1", portal.Active(), portal.Frames())
	}
}
