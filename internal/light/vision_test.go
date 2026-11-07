package light_test

import (
	"fmt"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/light"
)

// The committed goldens of this file: the probe-produced ray table and the
// three-frame reveal trace.
const (
	visionFixtureDir   = "../testdata/light"
	visionRaysFixture  = "vision_rays.tsv"
	visionTraceFixture = "vision_trace.tsv"
	// traceSide is the side of the square trace scenario: 25x25 tiles.
	traceSide = 25
	// The section names of the commit trace fixture.
	traceFlags     = "flags"
	traceExplorer  = "explorer"
	traceTransList = "translist"
	traceActive    = "active"
)

// TestVisionRaysGolden pins VisionRays and RayLenAdj against the table the
// pinned upstream translation unit emitted (vision.cpp:29-53, :61).
func TestVisionRaysGolden(t *testing.T) {
	// given
	lines := readVisionFixture(t, visionRaysFixture)
	wantRays, wantAdj, err := parseVisionRays(lines)
	if err != nil {
		t.Fatalf("parse %s: %v", visionRaysFixture, err)
	}

	// when
	gotRays := light.VisionRays
	gotAdj := light.RayLenAdj

	// then
	for ray := range wantRays {
		if gotRays[ray] != wantRays[ray] {
			t.Fatalf("ray %d = %v, want %v", ray, gotRays[ray], wantRays[ray])
		}
	}
	if gotAdj != wantAdj {
		t.Fatalf("RayLenAdj = %v, want %v", gotAdj, wantAdj)
	}
}

// TestVisionRaysCoverFullCircle pins that the table stays a one-quadrant table:
// every point is in the first quadrant and a non-vertical ray advances its x
// component, which is what makes the four-quadrant mirroring exhaustive.
func TestVisionRaysCoverFullCircle(t *testing.T) {
	for ray, points := range light.VisionRays {
		t.Run(fmt.Sprintf("ray %d", ray), func(t *testing.T) {
			// when
			fault := firstQuadrantFault(ray, points)

			// then
			if fault != "" {
				t.Fatal(fault)
			}
		})
	}
}

// TestVisionRayLengthsStayInTable pins upstream's static_assert
// (vision.cpp:62): RayLenAdj keeps every ray length at the widest radius inside
// the 15 points of its row.
func TestVisionRayLengthsStayInTable(t *testing.T) {
	const radius = light.LightsMax

	for ray := range light.RayLenAdj {
		t.Run(fmt.Sprintf("ray %d", ray), func(t *testing.T) {
			// when
			rayLen := radius - int(light.RayLenAdj[ray])

			// then
			if rayLen > len(light.VisionRays[ray]) {
				t.Fatalf("ray %d length %d exceeds the %d table points", ray, rayLen, len(light.VisionRays[ray]))
			}
		})
	}
}

// TestDoVisionStopAtWall pins the early break on a tile that does not pass the
// light: the wall tile is revealed, nothing behind it is (vision.cpp:105-112).
func TestDoVisionStopAtWall(t *testing.T) {
	// given
	origin := light.Point{X: 10, Y: 10}
	wall := light.Point{X: 12, Y: 10}
	visible := map[light.Point]int{}
	callbacks := recordingVisionCallbacks(func(position light.Point) bool {
		return position != wall
	}, visible)

	// when
	light.DoVision(origin, 4, callbacks)

	// then
	if visible[wall] == 0 {
		t.Fatal("wall tile was never revealed; the ray has to report the tile that stops it")
	}
	for _, behind := range []light.Point{{X: 13, Y: 10}, {X: 14, Y: 10}} {
		if visible[behind] != 0 {
			t.Fatalf("tile %v behind the wall revealed %d times, want 0", behind, visible[behind])
		}
	}
}

// TestDoVisionBlockedCornerIsInvisible pins the diagonal rule of
// vision.cpp:100-106: a tile whose two diagonal neighbours both block the light
// is never revealed, the case upstream's VisionTest calls "no visibility
// through adjacent tiles".
func TestDoVisionBlockedCornerIsInvisible(t *testing.T) {
	// given
	origin := light.Point{X: 10, Y: 10}
	target := light.Point{X: 11, Y: 11}
	walls := map[light.Point]bool{{X: 10, Y: 11}: true, {X: 11, Y: 10}: true}
	visible := map[light.Point]int{}
	callbacks := recordingVisionCallbacks(func(position light.Point) bool {
		return !walls[position]
	}, visible)

	// when: radius 6 is the first radius where the diagonal rays run at all
	// (RayLenAdj trims the ray 7 row to nothing at radius 2).
	light.DoVision(origin, 6, callbacks)

	// then
	if visible[target] != 0 {
		t.Fatalf("corner tile %v revealed %d times through two walls, want 0", target, visible[target])
	}
}

// TestDoVisionOnePassingCornerNeighbourReveals pins the other half of the same
// rule: one diagonal neighbour that passes the light is enough.
func TestDoVisionOnePassingCornerNeighbourReveals(t *testing.T) {
	// given
	origin := light.Point{X: 10, Y: 10}
	target := light.Point{X: 11, Y: 11}
	walls := map[light.Point]bool{{X: 10, Y: 11}: true}
	visible := map[light.Point]int{}
	callbacks := recordingVisionCallbacks(func(position light.Point) bool {
		return !walls[position]
	}, visible)

	// when
	light.DoVision(origin, 6, callbacks)

	// then
	if visible[target] == 0 {
		t.Fatalf("corner tile %v never revealed with one open neighbour", target)
	}
}

// TestDoVisionOutOfBoundsStopsRay pins the in-bounds break (vision.cpp:75-77):
// a ray leaving the grid is dropped instead of reported.
func TestDoVisionOutOfBoundsStopsRay(t *testing.T) {
	// given
	origin := light.Point{X: 0, Y: 0}
	reported := map[light.Point]int{}
	callbacks := light.VisionCallbacks{
		MarkVisible:     func(position light.Point) { reported[position]++ },
		MarkTransparent: func(light.Point) {},
		PassesLight:     func(light.Point) bool { return true },
		InBounds:        func(position light.Point) bool { return position.X >= 0 && position.Y >= 0 },
	}

	// when
	light.DoVision(origin, 3, callbacks)

	// then
	for position := range reported {
		if position.X < 0 || position.Y < 0 {
			t.Fatalf("out-of-bounds tile %v reported to the callbacks", position)
		}
	}
	if reported[origin] != 1 {
		t.Fatalf("origin reported %d times, want 1", reported[origin])
	}
}

// TestDoUnVisionClearsRadiusPlusTwoRect pins the double radius widening of
// DoUnVision (lighting.cpp:181-190): a (2*radius+5)² square loses Visible and
// Lit, and nothing outside it changes.
func TestDoUnVisionClearsRadiusPlusTwoRect(t *testing.T) {
	// given
	const radius = 4
	state := newVisionState()
	state.flags = map[light.Point]light.DungeonFlag{}
	state.inBounds = func(light.Point) bool { return true }
	state.withVisionBits()

	// when
	light.DoUnVision(light.Point{X: 10, Y: 10}, radius, visionGrid(state))

	// then
	wantSide := 2*radius + 5
	cleared := 0
	for position, flags := range state.flags {
		if flags&light.FlagVisionBits == 0 {
			cleared++
			continue
		}
		if isInsideSquare(position, light.Point{X: 10, Y: 10}, radius+2) {
			t.Fatalf("tile %v inside the rect kept %#x", position, flags)
		}
	}
	if cleared != wantSide*wantSide {
		t.Fatalf("cleared %d tiles, want %d", cleared, wantSide*wantSide)
	}
	assertUnVisionEdges(t, state)
}

// TestDoVisionFlagsExistingFlagGatesAutomap pins the guard of
// lighting.cpp:90-99: the automap only learns about a tile dFlags already
// marks, while Explored is set either way.
func TestDoVisionFlagsExistingFlagGatesAutomap(t *testing.T) {
	position := light.Point{X: 3, Y: 4}
	cases := []struct {
		name      string
		existing  light.DungeonFlag
		explorer  light.Exploration
		wantViews []automapCall
		wantFlags light.DungeonFlag
	}{
		{
			name:      "empty tile is explored without naming an explorer",
			existing:  light.FlagNone,
			explorer:  light.ExpSelf,
			wantFlags: light.FlagExplored | light.FlagVisible,
		},
		{
			name:      "flagged tile is explored for the current player",
			existing:  light.FlagPopulated,
			explorer:  light.ExpSelf,
			wantViews: []automapCall{{position: position, explorer: light.ExpSelf}},
			wantFlags: light.FlagExplored | light.FlagVisible | light.FlagPopulated,
		},
		{
			name:      "flagged tile is explored for a peer",
			existing:  light.FlagPopulated,
			explorer:  light.ExpOthers,
			wantViews: []automapCall{{position: position, explorer: light.ExpOthers}},
			wantFlags: light.FlagExplored | light.FlagVisible | light.FlagPopulated,
		},
		{
			name:      "map exp none leaves the automap alone",
			existing:  light.FlagPopulated,
			explorer:  light.ExpNone,
			wantFlags: light.FlagVisible | light.FlagPopulated,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			state := newVisionState()
			state.flags[position] = tc.existing

			// when
			light.DoVisionFlagsPeer(position, tc.explorer, visionGrid(state))

			// then
			assertAutomapGated(t, state.views, tc.wantViews)
			if got := state.flags[position]; got != tc.wantFlags {
				t.Fatalf("flags = %#x, want %#x", got, tc.wantFlags)
			}
		})
	}
}

// TestDoVisionFlagsVisibleSetsLit pins the two independent bits of
// lighting.cpp:99-102: only the current player's reveal sets Lit, every reveal
// sets Visible.
func TestDoVisionFlagsVisibleSetsLit(t *testing.T) {
	type flagWriter func(position light.Point, doAutomap light.Exploration, grid light.VisionGrid)
	cases := []struct {
		name      string
		reveal    flagWriter
		wantFlags light.DungeonFlag
	}{
		{name: "own player lights the tile", reveal: light.DoVisionFlags, wantFlags: light.FlagLit | light.FlagVisible},
		{name: "peer only makes it visible", reveal: light.DoVisionFlagsPeer, wantFlags: light.FlagVisible},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			state := newVisionState()
			position := light.Point{X: 5, Y: 6}
			reveal := tc.reveal

			// when
			reveal(position, light.ExpNone, visionGrid(state))

			// then
			if got := state.flags[position]; got != tc.wantFlags {
				t.Fatalf("flags = %#x, want %#x", got, tc.wantFlags)
			}
		})
	}
}

// TestApplyVisionMarksTransList pins that a revealed tile inside a transparency
// region raises that region and a region-0 tile raises nothing
// (lighting.cpp:198-204).
func TestApplyVisionMarksTransList(t *testing.T) {
	// given
	state := newVisionState()
	state.trans[light.Point{X: 11, Y: 10}] = 7
	state.trans[light.Point{X: 12, Y: 10}] = -3

	// when
	light.ApplyVision(light.Point{X: 10, Y: 10}, 2, light.ExpSelf, visionGrid(state))

	// then
	want := []int8{-3, 7}
	if got := sortedRegions(state.regions); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("regions raised = %v, want %v", got, want)
	}
}

// TestProcessVisionListParity pins the whole reveal pass against the trace the
// pinned upstream passes produced: dFlags, the automap explorer records,
// TransList and the VisionActive transitions all have to match
// (lighting.cpp:437-472).
func TestProcessVisionListParity(t *testing.T) {
	frames, err := parseVisionTrace(readVisionFixture(t, visionTraceFixture))
	if err != nil {
		t.Fatalf("parse %s: %v", visionTraceFixture, err)
	}
	world := newTraceWorld(t)

	t.Run("frame1", func(t *testing.T) {
		// when
		world.reveal()

		// then
		assertTraceEqual(t, frames["frame1"], world)
	})

	t.Run("frame2", func(t *testing.T) {
		// given
		world.list.ChangeXY(traceFriendlyID, light.Point{X: 8, Y: 8})
		world.players[traceOffLevelID].plractive = false

		// when
		world.reveal()

		// then
		assertTraceEqual(t, frames["frame2"], world)
	})

	t.Run("frame3", func(t *testing.T) {
		// given
		world.players[traceMyID].lvlChanging = true
		world.players[traceHostileID].plractive = false
		world.list.ChangeXY(traceMyID, light.Point{X: 13, Y: 12})

		// when
		world.reveal()

		// then
		assertTraceEqual(t, frames["frame3"], world)
	})
}

// TestProcessVisionListQuietWithoutChange pins UpdateVision (lighting.cpp:439):
// a run with no pending change must touch neither dFlags nor TransList.
func TestProcessVisionListQuietWithoutChange(t *testing.T) {
	// given
	state := newVisionState()
	list := light.NewVisionList()

	// when
	list.ProcessVisionList(nil, visionGrid(state))

	// then
	if state.writes != 0 || state.resets != 0 {
		t.Fatalf("quiet pass wrote %d flags and reset TransList %d times", state.writes, state.resets)
	}
}

// TestVisionListActivateAndChange pins the slot bookkeeping of ActivateVision,
// ChangeVisionXY and ChangeVisionRadius (lighting.cpp:406-435).
func TestVisionListActivateAndChange(t *testing.T) {
	// given
	list := light.NewVisionList()

	// when
	list.Activate(light.Point{X: 4, Y: 5}, 6, 1)
	list.ChangeXY(1, light.Point{X: 7, Y: 5})
	list.ChangeRadius(1, 8)

	// then
	slot := list.Slots[1]
	if !list.Active[1] {
		t.Fatal("slot 1 not active after Activate")
	}
	if slot.Position.Tile != (light.Point{X: 7, Y: 5}) {
		t.Fatalf("tile = %v, want {7 5}", slot.Position.Tile)
	}
	// ChangeVisionRadius snapshots the tile currently revealed, which the
	// preceding ChangeVisionXY already moved to {7 5}.
	if slot.Position.Old != (light.Point{X: 7, Y: 5}) {
		t.Fatalf("old tile = %v, want {7 5}", slot.Position.Old)
	}
	if slot.Radius != 8 || slot.OldRadius != 6 {
		t.Fatalf("radius/oldRadius = %d/%d, want 8/6", slot.Radius, slot.OldRadius)
	}
	if !slot.HasChanged {
		t.Fatal("slot not flagged as changed")
	}
}
