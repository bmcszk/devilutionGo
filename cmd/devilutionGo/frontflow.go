//go:build windowed

package main

import (
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/ui"
	"github.com/bmcszk/devilutionGo/internal/present/scene"
)

// frontBoot is the front-screen half of the windowed boot: the boot lane the
// window drives plus the front-screen flow installed in front of the town
// scene (devilGo-uuyh). It lives behind the windowed build guard: the scene
// import must stay out of the headless binary's dependency graph
// (TestHeadlessMode_HasNoDisplayDependency). Ebiten-free itself, so the
// transitions are black-box testable windowless.
type frontBoot struct {
	lane *deviceLane
	flow *scene.FrontFlow
}

// newFrontBoot wires the front-screen flow exactly as runWindowed boots it:
// the same device lane over a scripted tick source, the flow in front of the
// town scene. Windowless and deterministic.
func newFrontBoot() *frontBoot {
	src := &engine.MockTickSource{}
	g := game.NewGame(src)

	return &frontBoot{lane: bootDeviceLane(src, g), flow: scene.NewFrontFlow()}
}

// handleKey feeds one key press to the flow: the input seam the window layer
// maps its raw key codes onto (windowKey precedent).
func (b *frontBoot) handleKey(key ui.Key) {
	b.flow.HandleKey(key)
}
