package light_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/light"
)

// The labels recordGrid records for the two LightGrid calls.
const (
	callUnLight  = "unlight"
	callLighting = "light"
)

// recordGrid is the LightGrid recording stub: it remembers every call the
// process passes make, in order.
type recordGrid struct {
	calls []string
}

// TestAddLightFirstSlot pins the id the first add returns: InitLighting's
// identity permutation hands out slot 0 first (lighting.cpp:249, :263).
func TestAddLightFirstSlot(t *testing.T) {
	// given
	list := light.NewLightList()

	// when
	id := list.AddLight(light.Point{X: 4, Y: 7}, 10)

	// then
	if id != 0 {
		t.Fatalf("first AddLight id = %d, want 0", id)
	}
	if list.Count != 1 {
		t.Fatalf("ActiveLightCount = %d, want 1", list.Count)
	}
}

// TestAddLightStoresRadiusAndTile pins what AddLight writes into the slot
// (lighting.cpp:265-269): the tile, the radius, a zero offset and both flags
// cleared.
func TestAddLightStoresRadiusAndTile(t *testing.T) {
	// given
	list := light.NewLightList()

	// when
	id := list.AddLight(light.Point{X: 4, Y: 7}, 10)

	// then
	got := list.Lights[id]
	want := light.Light{Position: light.LightPosition{Tile: light.Point{X: 4, Y: 7}}, Radius: 10}
	if got != want {
		t.Fatalf("Lights[%d] = %+v, want %+v", id, got, want)
	}
}

// TestAddLightSlotExhaustion pins the NO_LIGHT return at the MAXLIGHTS cap
// (lighting.cpp:260-261) and that a freed slot becomes available again.
func TestAddLightSlotExhaustion(t *testing.T) {
	// given: a full pool.
	list := light.NewLightList()
	for i := range light.MaxLights {
		if id := list.AddLight(light.Point{X: i, Y: 0}, 1); id != i {
			t.Fatalf("AddLight #%d id = %d, want %d", i, id, i)
		}
	}

	// when: one add too many.
	overflow := list.AddLight(light.Point{X: 1, Y: 1}, 1)

	// then: no light, and the count is untouched.
	if overflow != light.NoLight {
		t.Fatalf("AddLight on a full pool = %d, want NoLight (%d)", overflow, light.NoLight)
	}
	if list.Count != light.MaxLights {
		t.Fatalf("ActiveLightCount = %d, want %d", list.Count, light.MaxLights)
	}
}

// TestFreedSlotIsReused pins the swap in ProcessLightList (lighting.cpp:387-389):
// the dead id leaves the live range and the next AddLight hands it back.
func TestFreedSlotIsReused(t *testing.T) {
	// given: three lights, the middle one un-lit.
	list := light.NewLightList()
	list.AddLight(light.Point{X: 0, Y: 0}, 1)
	doomed := list.AddLight(light.Point{X: 1, Y: 0}, 1)
	list.AddLight(light.Point{X: 2, Y: 0}, 1)
	list.AddUnLight(doomed)
	grid := &recordGrid{}

	// when
	list.ProcessLightList(grid.grid())

	// then: the dead id is out of the live range and is the next one handed out.
	if list.Count != 2 {
		t.Fatalf("ActiveLightCount = %d, want 2", list.Count)
	}
	if got := order(list); len(got) != 2 || got[0] == doomed || got[1] == doomed {
		t.Fatalf("live ids = %v, want %d dropped", got, doomed)
	}
	if id := list.AddLight(light.Point{X: 3, Y: 0}, 1); id != doomed {
		t.Fatalf("reused id = %d, want the freed %d", id, doomed)
	}
}

// TestAddUnLightClearsFootprintOnce pins the invalid pass
// (lighting.cpp:376-378): an un-lit light restores its cone exactly once, not
// once per process.
func TestAddUnLightClearsFootprintOnce(t *testing.T) {
	// given
	list := light.NewLightList()
	id := list.AddLight(light.Point{X: 5, Y: 5}, 4)
	list.AddUnLight(id)
	grid := &recordGrid{}

	// when: two process passes back to back.
	list.ProcessLightList(grid.grid())
	first := len(grid.calls)
	list.ProcessLightList(grid.grid())

	// then: the first pass un-lit and never drew; the second was a no-op.
	if first != 1 || grid.calls[0] != callUnLight {
		t.Fatalf("first pass calls = %v, want one unlight", grid.calls)
	}
	if len(grid.calls) != first {
		t.Fatalf("second pass added calls %v, want the pool to be clean", grid.calls[first:])
	}
}

// TestAddUnLightIgnoresNoLight pins the NO_LIGHT guard (lighting.cpp:282-283).
func TestAddUnLightIgnoresNoLight(t *testing.T) {
	// given
	list := light.NewLightList()

	// when
	list.AddUnLight(light.NoLight)

	// then
	if list.Count != 0 {
		t.Fatalf("ActiveLightCount = %d, want 0", list.Count)
	}
}

// TestChangeLightRadiusRoundTrip pins the radius change and its restore
// (lighting.cpp:299-303): the old radius is snapshotted, the tile is kept.
func TestChangeLightRadiusRoundTrip(t *testing.T) {
	// given
	list := light.NewLightList()
	id := list.AddLight(light.Point{X: 5, Y: 5}, 1)

	// when
	list.ChangeLightRadius(id, 9)

	// then
	got := list.Lights[id]
	if got.Radius != 9 || got.OldRadius != 1 {
		t.Fatalf("radius/oldRadius = %d/%d, want 9/1", got.Radius, got.OldRadius)
	}
	if got.Position.Tile != (light.Point{X: 5, Y: 5}) || got.Position.Old != (light.Point{X: 5, Y: 5}) {
		t.Fatalf("tile/old = %+v/%+v, want (5,5)/(5,5)", got.Position.Tile, got.Position.Old)
	}
}

// TestChangeLightXYKeepsRadius pins the move (lighting.cpp:308-324): the tile
// moves, the radius does not.
func TestChangeLightXYKeepsRadius(t *testing.T) {
	// given
	list := light.NewLightList()
	id := list.AddLight(light.Point{X: 5, Y: 5}, 6)

	// when
	list.ChangeLightXY(id, light.Point{X: 8, Y: 9})

	// then
	got := list.Lights[id]
	if got.Position.Tile != (light.Point{X: 8, Y: 9}) || got.Position.Old != (light.Point{X: 5, Y: 5}) {
		t.Fatalf("tile/old = %+v/%+v, want (8,9)/(5,5)", got.Position.Tile, got.Position.Old)
	}
	if got.Radius != 6 || got.OldRadius != 6 {
		t.Fatalf("radius/oldRadius = %d/%d, want 6/6", got.Radius, got.OldRadius)
	}
}

// TestChangeLightSetsBoth pins the combined mutator (lighting.cpp:347-364).
func TestChangeLightSetsBoth(t *testing.T) {
	// given
	list := light.NewLightList()
	id := list.AddLight(light.Point{X: 5, Y: 5}, 1)

	// when
	list.ChangeLight(id, light.Point{X: 2, Y: 3}, 12)

	// then
	got := list.Lights[id]
	if got.Position.Tile != (light.Point{X: 2, Y: 3}) || got.Radius != 12 {
		t.Fatalf("tile/radius = %+v/%d, want (2,3)/12", got.Position.Tile, got.Radius)
	}
}

// TestChangeLightOffsetSkipOnUnchanged pins the early return upstream takes
// when the offset equals the current one (lighting.cpp:330-332).
func TestChangeLightOffsetSkipOnUnchanged(t *testing.T) {
	// given: a light already processed, so the pool is clean.
	list := light.NewLightList()
	id := list.AddLight(light.Point{X: 5, Y: 5}, 6)
	grid := &recordGrid{}
	list.ProcessLightList(grid.grid())
	grid.calls = nil

	// when: the same offset is set again.
	list.ChangeLightOffset(id, light.Offset{})
	list.ProcessLightList(grid.grid())

	// then: nothing was drawn, because nothing changed.
	if len(grid.calls) != 0 {
		t.Fatalf("calls = %v, want none for an unchanged offset", grid.calls)
	}
}

// TestChangeLightOffsetMovesCone pins the offset change (lighting.cpp:334-341):
// the cone moves, the tile does not.
func TestChangeLightOffsetMovesCone(t *testing.T) {
	// given
	list := light.NewLightList()
	id := list.AddLight(light.Point{X: 5, Y: 5}, 6)

	// when
	list.ChangeLightOffset(id, light.Offset{DeltaX: 3, DeltaY: -2})

	// then
	if got := list.Lights[id].Position.Offset; got != (light.Offset{DeltaX: 3, DeltaY: -2}) {
		t.Fatalf("offset = %+v, want {3 -2}", got)
	}
}

// TestChangeMutatorsIgnoreNoLight pins the NO_LIGHT guard every ChangeLight*
// starts with (lighting.cpp:296-297, :314-315, :329-330, :352-353).
func TestChangeMutatorsIgnoreNoLight(t *testing.T) {
	// given
	list := light.NewLightList()
	grid := &recordGrid{}

	// when
	list.ChangeLightRadius(light.NoLight, 9)
	list.ChangeLightXY(light.NoLight, light.Point{X: 1, Y: 1})
	list.ChangeLightOffset(light.NoLight, light.Offset{DeltaX: 1})
	list.ChangeLight(light.NoLight, light.Point{X: 1, Y: 1}, 9)
	list.ProcessLightList(grid.grid())

	// then
	if list.Count != 0 || len(grid.calls) != 0 {
		t.Fatalf("count/calls = %d/%v, want 0 and no calls", list.Count, grid.calls)
	}
}

// TestProcessLightListQuietWhenUnchanged pins the UpdateLighting guard
// (lighting.cpp:372-373): a fresh pool draws nothing.
func TestProcessLightListQuietWhenUnchanged(t *testing.T) {
	// given
	list := light.NewLightList()
	grid := &recordGrid{}

	// when
	list.ProcessLightList(grid.grid())

	// then
	if len(grid.calls) != 0 {
		t.Fatalf("calls = %v, want none on a clean pool", grid.calls)
	}
}

// TestProcessLightListClearsOldConeThenDrawsNew pins the two-pass order
// (lighting.cpp:374-395): every clear happens before any draw, and a moved
// light clears its old cone before drawing the new one.
func TestProcessLightListClearsOldConeThenDrawsNew(t *testing.T) {
	// given: two lights, both moved, one of them also re-radiied.
	list := light.NewLightList()
	first := list.AddLight(light.Point{X: 1, Y: 1}, 3)
	second := list.AddLight(light.Point{X: 2, Y: 2}, 3)
	grid := &recordGrid{}
	list.ProcessLightList(grid.grid())
	grid.calls = nil
	list.ChangeLightXY(first, light.Point{X: 5, Y: 5})
	list.ChangeLightRadius(second, 7)

	// when
	list.ProcessLightList(grid.grid())

	// then
	want := []string{callUnLight, callUnLight, callLighting, callLighting}
	assertCalls(t, grid.calls, want)
}

// TestProcessLightListDrawsSwappedInLight pins the re-scan at lighting.cpp:390:
// the id swapped into the dead id's place is processed in the same pass, so the
// live range is walked to completion.
func TestProcessLightListDrawsSwappedInLight(t *testing.T) {
	// given: three lights, the first one un-lit, all already drawn once.
	list := light.NewLightList()
	doomed := list.AddLight(light.Point{X: 1, Y: 1}, 3)
	list.AddLight(light.Point{X: 2, Y: 1}, 3)
	list.AddLight(light.Point{X: 3, Y: 1}, 3)
	grid := &recordGrid{}
	list.ProcessLightList(grid.grid())
	grid.calls = nil
	list.AddUnLight(doomed)

	// when
	list.ProcessLightList(grid.grid())

	// then: one clear for the dead light and two draws for the survivors.
	want := []string{callUnLight, callLighting, callLighting}
	assertCalls(t, grid.calls, want)
	if list.Count != 2 {
		t.Fatalf("ActiveLightCount = %d, want 2", list.Count)
	}
}

// TestProcessLightListShortCircuitsSolidTile pins the solid-tile skip
// (lighting.cpp:391-394): a light inside a wall is cleared but never drawn.
func TestProcessLightListShortCircuitsSolidTile(t *testing.T) {
	// given: a light on a solid tile.
	list := light.NewLightList()
	list.AddLight(light.Point{X: 5, Y: 5}, 4)
	grid := &recordGrid{}

	// when
	list.ProcessLightList(light.LightGrid{
		UnLight:  grid.UnLight,
		Lighting: grid.Lighting,
		IsSolid:  func(light.Point) bool { return true },
	})

	// then
	if len(grid.calls) != 0 {
		t.Fatalf("calls = %v, want none for a light inside a wall", grid.calls)
	}
}

// TestCloseLightsEmptiesPool pins CloseLights: after it every slot is free and
// the ids restart at 0, so an id held across a level change cannot resurrect a
// stale position.
func TestCloseLightsEmptiesPool(t *testing.T) {
	// given
	list := light.NewLightList()
	list.AddLight(light.Point{X: 5, Y: 5}, 4)
	list.AddLight(light.Point{X: 6, Y: 5}, 4)

	// when
	list.CloseLights()

	// then
	if list.Count != 0 {
		t.Fatalf("ActiveLightCount = %d, want 0", list.Count)
	}
	if list.Lights[0] != (light.Light{}) {
		t.Fatalf("Lights[0] = %+v, want the zero light", list.Lights[0])
	}
	if id := list.AddLight(light.Point{X: 1, Y: 1}, 2); id != 0 {
		t.Fatalf("id after CloseLights = %d, want 0", id)
	}
}

// TestVisionSeamIsDeclarable pins the DoUnVision seam type: the wiring the
// vision port will use compiles against this package without that port
// existing yet (DoVision is a separate bean).
func TestVisionSeamIsDeclarable(t *testing.T) {
	// given: the seam the vision port will be wired through.
	var seen light.Point
	var hook light.VisionUnlighter = func(position light.Point, _ uint8) { seen = position }

	// when: the seam is called the way the vision port will call it.
	hook(light.Point{X: 1, Y: 1}, 10)

	// then: the seam takes the (position, radius) pair DoUnVision takes.
	if seen != (light.Point{X: 1, Y: 1}) {
		t.Fatalf("hook saw %+v, want {1 1}", seen)
	}
}

// UnLight records a DoUnLight call.
func (g *recordGrid) UnLight(_ light.Point, _ uint8) {
	g.calls = append(g.calls, callUnLight)
}

// Lighting records a DoLighting call.
func (g *recordGrid) Lighting(_ light.Point, _ uint8, _ light.Offset) {
	g.calls = append(g.calls, callLighting)
}

// IsSolid reports no solid tiles.
func (*recordGrid) IsSolid(_ light.Point) bool { return false }

// grid builds the LightGrid these methods implement.
func (g *recordGrid) grid() light.LightGrid {
	return light.LightGrid{UnLight: g.UnLight, Lighting: g.Lighting, IsSolid: g.IsSolid}
}

// order returns the live ids of a pool, front to back.
func order(list *light.LightList) []int {
	ids := make([]int, 0, list.Count)
	for i := range list.Count {
		ids = append(ids, int(list.Active[i]))
	}

	return ids
}

// assertCalls fails when the recorded grid calls differ from want.
func assertCalls(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("calls = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("calls = %v, want %v", got, want)
		}
	}
}
