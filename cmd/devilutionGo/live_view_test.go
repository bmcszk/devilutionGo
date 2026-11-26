//go:build windowed

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// TestLiveViewPresenterFollowsTheHeroTile pins the live-view seam (devilGo-njzu):
// the presenter composes the town scene at the kernel state's hero tile —
// the same state the device lane reads — so a walked hero drags the view with
// it. No shadow position, no view-side prediction: the presenter's only hero
// source is the state read the window boot injects.
func TestLiveViewPresenterFollowsTheHeroTile(t *testing.T) {
	// given the windowed boot's wiring (the s6e device lane) and a presenter
	// reading the same kernel state the window injects
	dirs := requireArchives(t)
	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	lane := bootDeviceLane(src, g)
	start := lane.state.Hero()

	present := newScenePresenter("", nil, func() render.TilePos {
		return toSceneTile(lane.state.Hero())
	}, func() int { return engine.BrightnessMin }, func(tile render.TilePos, brightness int) (scene.TownScene, error) {
		return scene.ComposeTownFrameAtBrightness(dirs, tile, brightness)
	})

	// when the boot's first frame composes at the spawn tile
	present.syncScene()

	if got := present.composedHeroTile(); got != toSceneTile(start) {
		t.Fatalf("first compose at %v, want the spawn tile %v", got, toSceneTile(start))
	}

	// and a firm east push walks the hero through the settled pass
	if !lane.PushStick(controls.PushFirm, eastPush()) {
		t.Fatal("gate refused the firm east push, want it accepted")
	}

	lane.frame(func() {})

	moved := lane.state.Hero()
	if moved == start {
		t.Fatalf("hero never moved from %v, want the east step", start)
	}

	// then the presenter recomposes at the moved tile — not the pinned
	// first-scene tile
	present.syncScene()

	if got := present.composedHeroTile(); got != toSceneTile(moved) {
		t.Fatalf("presenter hero tile = %v, want the state hero tile %v", got, toSceneTile(moved))
	}
}

// TestLiveViewSceneComposesAtTheGivenTile pins the scene seam (devilGo-njzu):
// ComposeTownFrameAtHero composes at an arbitrary hero tile and reports the
// tile it composed at — the scene has no tile of its own.
func TestLiveViewSceneComposesAtTheGivenTile(t *testing.T) {
	dirs := requireArchives(t)

	// given a tile away from the pinned first-scene tile
	tile := render.TilePos{X: 60, Y: 55}

	// when the scene composes at it
	composed, err := scene.ComposeTownFrameAtHero(dirs, tile)
	if err != nil {
		t.Fatalf("ComposeTownFrameAtHero: %v", err)
	}

	// then it reports the tile it composed at
	if got := scene.ComposedHeroTile(composed); got != tile {
		t.Fatalf("composed hero tile = %v, want %v", got, tile)
	}
}

// toSceneTile maps the kernel state's hero tile onto the render tile the
// scene composes at: the one conversion the window layer owns.
func toSceneTile(t world.Tile) render.TilePos {
	return render.TilePos{X: t.X, Y: t.Y}
}

// requireArchives skips when the retail archives are unreachable: the compose
// reads the real MPQ records (the warrior_test.go contract).
func requireArchives(t *testing.T) []string {
	t.Helper()

	dirs := filepath.SplitList(os.Getenv("DG_ASSET_ARCHIVES"))
	if len(dirs) == 0 || !scene.ArchivesPresent(dirs) {
		t.Skip("no retail MPQ found; set DG_ASSET_ARCHIVES to the Diablo install directories")
	}

	return dirs
}
