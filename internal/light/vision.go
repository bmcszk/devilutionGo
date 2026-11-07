package light

// This file ports the vision half of the lighting subsystem: the level-agnostic
// ray walk (Source/vision.cpp:51-120), the dFlags/TransList writers
// (Source/lighting.cpp:84-212) and the per-player reveal pass
// (Source/lighting.cpp:437-472). The arrays those passes write stay in the
// level layer; this file declares the boundary and never allocates them.

// FlagVisionBits is `DungeonFlag::Visible | DungeonFlag::Lit`, the pair
// DoUnVision clears (lighting.cpp:189).
const FlagVisionBits = FlagVisible | FlagLit

// VisionCallbacks is the four function_ref callbacks DoVision takes
// (vision.hpp:11-17).
type VisionCallbacks struct {
	// MarkVisible is markVisibleFn: the tile was hit by a ray.
	MarkVisible func(position Point)
	// MarkTransparent is markTransparentFn: the tile was hit and still passes
	// the light on.
	MarkTransparent func(position Point)
	// PassesLight is passesLightFn, TileAllowsLight (lighting.cpp:81-88).
	PassesLight func(position Point) bool
	// InBounds is inBoundsFn, InDungeonBounds (levels/dun_tile_data.hpp:143-146).
	InBounds func(position Point) bool
}

// VisionGrid is the dungeon-state boundary the vision passes read and write:
// dFlags (levels/dun_tile_data.hpp:109), dTransVal (:103), TransList (:97) and
// the automap hook the explored bits call (automap.cpp:1883-1891). The level
// layer owns every array.
type VisionGrid struct {
	// Flags returns dFlags[position].
	Flags func(position Point) DungeonFlag
	// SetFlags stores dFlags[position].
	SetFlags func(position Point, flags DungeonFlag)
	// PassesLight is TileAllowsLight: in bounds and no BlockLight property
	// (lighting.cpp:81-88).
	PassesLight func(position Point) bool
	// InBounds is InDungeonBounds (levels/dun_tile_data.hpp:143-146).
	InBounds func(position Point) bool
	// Trans returns dTransVal[position].
	Trans func(position Point) int8
	// MarkTransparent raises TransList[region] (lighting.cpp:199-203). Upstream
	// indexes the 256-entry array with the int8_t region; the level layer
	// reinterprets it as uint8 so a region above 127 stays in range.
	MarkTransparent func(region int8)
	// ResetTransList is `TransList = {}` (lighting.cpp:441).
	ResetTransList func()
	// SetAutomapView is SetAutomapView (automap.cpp:1883-1891).
	SetAutomapView func(position Point, explorer Exploration)
}

// PlayerView is the per-player slice of state ProcessVisionList reads
// (lighting.cpp:444-470). The player layer implements it; the pass never
// mutates a player.
type PlayerView interface {
	// ID is player.getId(): the VisionList slot and VisionActive index.
	ID() int
	// MyPlayer reports whether this player is MyPlayer.
	MyPlayer() bool
	// Plractive is player.plractive.
	Plractive() bool
	// OnActiveLevel is player.isOnActiveLevel().
	OnActiveLevel() bool
	// LvlChanging is player._pLvlChanging.
	LvlChanging() bool
	// FriendlyMode is player.friendlyMode, the automap bit a peer's reveal
	// sets (:466).
	FriendlyMode() bool
}

// VisionList is the vision state InitLighting resets (lighting.cpp:244-251):
// VisionList, VisionActive (lighting.h:35,39) and UpdateVision (:44). One value
// is one game's vision state.
type VisionList struct {
	// Slots is VisionList[MAXVISION]: one Light per player id.
	Slots [MaxVision]Light
	// Active is VisionActive: the ids ProcessVisionList still reveals for.
	Active [MaxVision]bool
	// update is UpdateVision: a mutator changed the list and ProcessVisionList
	// has not revealed it yet.
	update bool
}

// NewVisionList is InitLighting's vision half (lighting.cpp:244-251): no slot
// active and nothing to reveal.
func NewVisionList() *VisionList {
	return &VisionList{}
}

// Activate is ActivateVision (lighting.cpp:406-415): the player's vision slot
// starts at position with radius, and a reveal is pending.
func (v *VisionList) Activate(position Point, radius uint8, id int) {
	slot := &v.Slots[id]
	slot.Position.Tile = position
	slot.Radius = radius
	slot.IsInvalid = false
	slot.HasChanged = false
	v.Active[id] = true

	v.update = true
}

// ChangeRadius is ChangeVisionRadius (lighting.cpp:417-425): the slot's
// previous position and radius stay until the next reveal clears them.
func (v *VisionList) ChangeRadius(id int, radius uint8) {
	slot := v.markChanged(id)
	slot.Radius = radius
}

// ChangeXY is ChangeVisionXY (lighting.cpp:427-435): the player moved, the old
// tile keeps its revealed tiles until the next reveal clears them.
func (v *VisionList) ChangeXY(id int, position Point) {
	slot := v.markChanged(id)
	slot.Position.Tile = position
}

// ProcessVisionList is ProcessVisionList (lighting.cpp:437-472): the two-pass
// reveal that un-vision every dropped, moved or off-level player before it
// reveals anything, so no ray sees a tile another player just left.
func (v *VisionList) ProcessVisionList(players []PlayerView, grid VisionGrid) {
	if !v.update {
		return
	}

	grid.ResetTransList()
	v.clearDroppedVision(players, grid)
	v.revealPlayerVision(players, grid)

	v.update = false
}

// markChanged is the shared ChangeVision* prologue (lighting.cpp:419-422,
// :429-432): flag the slot and snapshot the footprint currently revealed.
func (v *VisionList) markChanged(id int) *Light {
	slot := &v.Slots[id]
	slot.HasChanged = true
	slot.Position.Old = slot.Position.Tile
	slot.OldRadius = slot.Radius

	v.update = true

	return slot
}

// clearDroppedVision is ProcessVisionList's first pass (lighting.cpp:444-459):
// the active slots that must stop revealing, and the stale footprint of every
// slot that moved.
func (v *VisionList) clearDroppedVision(players []PlayerView, grid VisionGrid) {
	for _, player := range players {
		id := player.ID()
		if !v.Active[id] {
			continue
		}

		slot := &v.Slots[id]
		if playerDroppedVision(player) {
			DoUnVision(slot.Position.Tile, slot.Radius, grid)
			v.Active[id] = false
			continue
		}
		if slot.HasChanged {
			DoUnVision(slot.Position.Old, slot.OldRadius, grid)
			slot.HasChanged = false
		}
	}
}

// DoVision is DoVision (vision.cpp:51-120): casts VisionRays over four
// mirrored quadrants from position and reports every point to the callbacks.
//
// Upstream overloads this name in lighting.cpp:194; Go cannot, so that wrapper
// is ApplyVision.
func DoVision(position Point, radius uint8, callbacks VisionCallbacks) {
	callbacks.MarkVisible(position)

	for _, quadrant := range visionQuadrants {
		for ray := range VisionRays {
			castVisionRay(position, int(radius)-int(RayLenAdj[ray]), VisionRays[ray], quadrant, callbacks)
		}
	}
}

// ApplyVision is DoVision (lighting.cpp:194-212) for MyPlayer: the reveal that
// writes dFlags, TransList and the automap, lighting every tile it reaches.
func ApplyVision(position Point, radius uint8, doAutomap Exploration, grid VisionGrid) {
	DoVision(position, radius, visionCallbacks(doAutomap, grid, DoVisionFlags))
}

// ApplyVisionPeer is DoVision (lighting.cpp:194-212) for another player: the
// same reveal, but a peer only makes a tile Visible, never Lit (:204-209).
func ApplyVisionPeer(position Point, radius uint8, doAutomap Exploration, grid VisionGrid) {
	DoVision(position, radius, visionCallbacks(doAutomap, grid, DoVisionFlagsPeer))
}

// DoVisionFlags is DoVisionFlags (lighting.cpp:90-94) for MyPlayer's reveal:
// the per-tile bits one ray point contributes, Lit included.
//
// The automap only learns about a tile the level already flags, so an
// unexplored tile is marked Explored without naming an explorer.
func DoVisionFlags(position Point, doAutomap Exploration, grid VisionGrid) {
	writeVisionFlags(position, doAutomap, FlagLit, grid)
}

// DoVisionFlagsPeer is DoVisionFlags for another player's reveal: the tile
// stays unlit, only Visible is raised (lighting.cpp:98-101).
func DoVisionFlagsPeer(position Point, doAutomap Exploration, grid VisionGrid) {
	writeVisionFlags(position, doAutomap, FlagNone, grid)
}

// DoUnVision is DoUnVision (lighting.cpp:181-190): clears Visible and Lit over
// the (2*radius+5)² square around position. Upstream widens the radius twice so
// no stray vision tile survives the move; the reveal that follows runs in the
// same tick.
func DoUnVision(position Point, radius uint8, grid VisionGrid) {
	radius += 2

	for x := position.X - int(radius); x <= position.X+int(radius); x++ {
		for y := position.Y - int(radius); y <= position.Y+int(radius); y++ {
			tile := Point{X: x, Y: y}
			if grid.InBounds(tile) {
				grid.SetFlags(tile, grid.Flags(tile)&^FlagVisionBits)
			}
		}
	}
}

// writeVisionFlags is the shared DoVisionFlags body: the automap bits, the bit
// only the current player adds, and the Visible every reveal sets.
func writeVisionFlags(position Point, doAutomap Exploration, lit DungeonFlag, grid VisionGrid) {
	flags := grid.Flags(position)
	if doAutomap != ExpNone {
		if flags != FlagNone {
			grid.SetAutomapView(position, doAutomap)
		}
		flags |= FlagExplored
	}
	grid.SetFlags(position, flags|lit|FlagVisible)
}

// visionCallbacks is the four lambdas DoVision is called with in
// lighting.cpp:196-210, wired to the grid boundary; mark writes the per-tile
// bits of the tile a ray reached.
func visionCallbacks(
	doAutomap Exploration,
	grid VisionGrid,
	mark func(position Point, doAutomap Exploration, grid VisionGrid),
) VisionCallbacks {
	return VisionCallbacks{
		MarkVisible: func(rayPoint Point) {
			mark(rayPoint, doAutomap, grid)
		},
		MarkTransparent: func(rayPoint Point) {
			if region := grid.Trans(rayPoint); region != 0 {
				grid.MarkTransparent(region)
			}
		},
		PassesLight: grid.PassesLight,
		InBounds:    grid.InBounds,
	}
}

// castVisionRay walks one of the 23 rays in one quadrant (vision.cpp:58-116): a
// ray stops at the first point that is blocked, out of bounds or opaque.
func castVisionRay(position Point, rayLen int, ray [15]Offset, quadrant Offset, callbacks VisionCallbacks) {
	for step := range rayLen {
		rayPoint, blocked := visionRayPoint(position, ray[step], quadrant, callbacks)
		if blocked {
			return
		}

		callbacks.MarkVisible(rayPoint)
		if !callbacks.PassesLight(rayPoint) {
			return
		}
		callbacks.MarkTransparent(rayPoint)
	}
}

// revealPlayerVision is ProcessVisionList's second pass (lighting.cpp:461-470).
func (v *VisionList) revealPlayerVision(players []PlayerView, grid VisionGrid) {
	for _, player := range players {
		id := player.ID()
		if !v.Active[id] {
			continue
		}

		slot := &v.Slots[id]
		if player.MyPlayer() {
			ApplyVision(slot.Position.Tile, slot.Radius, ExpSelf, grid)
			continue
		}
		ApplyVisionPeer(slot.Position.Tile, slot.Radius, explorationOfPeer(player), grid)
	}
}

// playerDroppedVision is the first-pass guard of lighting.cpp:449: a dead, off
// level or level-changing player stops revealing. MyPlayer keeps its vision
// while the level loads.
func playerDroppedVision(player PlayerView) bool {
	if !player.Plractive() || !player.OnActiveLevel() {
		return true
	}

	return player.LvlChanging() && !player.MyPlayer()
}

// explorationOfPeer is `player.friendlyMode ? MAP_EXP_OTHERS : MAP_EXP_NONE`
// (lighting.cpp:466): a hostile peer's reveal is not recorded on the automap.
func explorationOfPeer(player PlayerView) Exploration {
	if player.FriendlyMode() {
		return ExpOthers
	}

	return ExpNone
}

// visionRayPoint resolves the next point of one ray step and reports whether
// the ray ends before it is reported (vision.cpp:66-113).
//
// A ray that is not parallel to an axis may only reach a tile when one of the
// two diagonally adjacent tiles still passes the light, so a wall corner is not
// visible through its corner: the ray ends without revealing the tile.
func visionRayPoint(
	position Point,
	rayPoint Offset,
	quadrant Offset,
	callbacks VisionCallbacks,
) (point Point, blocked bool) {
	point = displacedBy(position, mirroredRayOffset(rayPoint, quadrant))
	if !callbacks.InBounds(point) {
		return point, true
	}
	if rayPoint.DeltaX <= 0 || rayPoint.DeltaY <= 0 {
		return point, false
	}

	adjacent1 := Offset{DeltaX: -quadrant.DeltaX}
	adjacent2 := Offset{DeltaY: -quadrant.DeltaY}
	if callbacks.PassesLight(displacedBy(point, adjacent1)) {
		return point, false
	}

	return point, !callbacks.PassesLight(displacedBy(point, adjacent2))
}

// mirroredRayOffset is `relRayPoint * quadrant` (vision.cpp:68): the component
// product of a one-quadrant ray point and a sign quadrant.
func mirroredRayOffset(rayPoint, quadrant Offset) Offset {
	return Offset{
		DeltaX: rayPoint.DeltaX * quadrant.DeltaX,
		DeltaY: rayPoint.DeltaY * quadrant.DeltaY,
	}
}

// displacedBy is `position + displacement` (engine/point.hpp:193-197).
func displacedBy(position Point, offset Offset) Point {
	return Point{X: position.X + int(offset.DeltaX), Y: position.Y + int(offset.DeltaY)}
}
