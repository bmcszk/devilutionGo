package main

import (
	"context"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
	"github.com/bmcszk/devilutionGo/internal/persist"
)

// refusingLevelSource serves no level image: a town record never calls the
// source (world restoreSession loads only when record.Loaded), so the save/load
// tests run asset-free.
type refusingLevelSource struct{}

// LevelImage implements world.LevelSource.
func (refusingLevelSource) LevelImage(int, int) (*world.LevelLayout, world.Tile, bool) {
	return nil, world.Tile{}, false
}

// TestDeviceLane_AttachSaveLoadLetsLoadSlotRestoreARecord drives the lane's
// save/load wiring: with the archive attached, a frozen session restores and
// the lane reports the load outcome once through its own hook.
func TestDeviceLane_AttachSaveLoadLetsLoadSlotRestoreARecord(t *testing.T) {
	// given the lane over a temp-dir store with the production archive
	// attached and the load hook registered
	lane, _ := newSaveLoadLane(t)

	var loads []kernel.Event
	lane.onLoadEvent = func(ev kernel.Event) { loads = append(loads, ev) }

	// when the session freezes into slot 0 and restores from it
	trace := saveThenLoad(t, lane, 0xC0FFEE)

	// then the trace holds the save and the load announcements, in order
	kinds := kindsOf(trace)
	if len(kinds) != 2 || kinds[0] != kernel.KindSavedSlot || kinds[1] != kernel.KindLoadedSlot {
		t.Fatalf("trace kinds = %v, want [SavedSlot LoadedSlot]", kinds)
	}

	if trace[0].A != 0 {
		t.Fatalf("SavedSlot slot = %d, want 0", trace[0].A)
	}

	// and the load hook fired once with the loaded slot
	if len(loads) != 1 || loads[0].Kind != kernel.KindLoadedSlot || loads[0].A != 0 {
		t.Fatalf("load hook events = %v, want one LoadedSlot for slot 0", loads)
	}
}

// TestDeviceLane_LoadSlotWithoutAttachIsSilent pins the default state: with
// no archive attached, a LoadSlot decision dispatches and announces nothing
// — the unregistered verb stays silent (ADR-018 law 2).
func TestDeviceLane_LoadSlotWithoutAttachIsSilent(t *testing.T) {
	// given the plain boot lane, no archive attached
	src := &engine.MockTickSource{}
	lane := bootDeviceLane(src, game.NewGame(src))

	// when a load decision dispatches and the frame settles
	lane.input.SubmitDecision(kernel.Command{
		Kind: kernel.KindLoadSlot, Tick: lane.game.TickCount, A: 0, B: 1,
	})

	lane.frame(func() {})

	// then the trace holds nothing: the verb never registered a handler
	if kinds := kindsOf(lane.game.Kernel.Trace()); len(kinds) != 0 {
		t.Fatalf("trace kinds after unattached load = %v, want none", kinds)
	}
}

// newSaveLoadLane boots the lane over a temp-dir store with the production
// archive attached: the wiring slice 2 pins.
func newSaveLoadLane(t *testing.T) (*deviceLane, *persist.HeroStoreSet) {
	t.Helper()

	set := persist.NewHeroStoreSet(t.TempDir(), persist.SaveNamespace{})

	src := &engine.MockTickSource{}
	lane := bootDeviceLane(src, game.NewGame(src))
	lane.attachSaveLoad(persist.NewSessionArchive(set, bootPackSource{}), refusingLevelSource{})

	// the create flow writes the hero row before any save: the archive keeps
	// the row's own name, so it must exist (production order, selhero.go)
	hero := persist.HeroRecord{
		Slot: 0, Name: "Lane Hero", ClassID: 0, Pack: make([]byte, persist.PlayerPackSize),
	}

	if err := set.SaveHero(context.Background(), hero); err != nil {
		t.Fatalf("SaveHero: %v", err)
	}

	return lane, set
}

// saveThenLoad freezes the lane's session into slot 0 and restores it,
// reporting the kernel trace after both settles.
func saveThenLoad(t *testing.T, lane *deviceLane, seed uint32) []kernel.Event {
	t.Helper()

	save := kernel.Command{Kind: kernel.KindSaveSlot, Tick: lane.game.TickCount, A: 0, B: int32(seed)}
	accepted := lane.input.SubmitDecision(save)
	if !accepted {
		t.Fatal("gate refused SaveSlot")
	}

	lane.frame(func() {})

	load := kernel.Command{Kind: kernel.KindLoadSlot, Tick: lane.game.TickCount, A: 0, B: int32(seed)}
	accepted = lane.input.SubmitDecision(load)
	if !accepted {
		t.Fatal("gate refused LoadSlot")
	}

	lane.pendingLoad = true // the screen sets the flag when it submits (selhero.go)

	lane.frame(func() {})

	return lane.game.Kernel.Trace()
}

// kindsOf projects a trace down to its event kinds.
func kindsOf(trace []kernel.Event) []kernel.Kind {
	kinds := make([]kernel.Kind, 0, len(trace))

	for _, ev := range trace {
		kinds = append(kinds, ev.Kind)
	}

	return kinds
}
