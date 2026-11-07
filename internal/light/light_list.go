package light

// This file ports the runtime light inventory: the light pool and its
// mutators (Source/lighting.h:30-42, Source/lighting.cpp:240-398).
//
// Name mapping for the ported mutators, so no upstream behaviour is invented:
//   changeLightFade  -> ChangeLight (lighting.cpp:347-364): a missile fades its
//                       light by re-setting the radius in place each frame
//                       (missiles.cpp:2852, :3080).
//   changeLightOff   -> AddUnLight (lighting.cpp:276-288): upstream turns a
//                       light off by flagging it invalid, not by zeroing a
//                       radius.
//   makeLightList    -> ProcessLightList (lighting.cpp:366-398): the redraw
//                       pass, which is also the lifetime decrement (dead slots
//                       are swapped out of the live range at :387-390).

// MaxLights is MAXLIGHTS (engine/lighting_defs.hpp:7): the size of the light
// pool.
const MaxLights = 32

// NoLight is NO_LIGHT (engine/lighting_defs.hpp:9): the id of "no light".
const NoLight = -1

// Point is a world-tile coordinate (engine/point.hpp Point); it holds the same
// value as the static light passes' levels.Point (levels/geom.go).
type Point struct{ X, Y int }

// Offset is DisplacementOf<int8_t> (engine/displacement.hpp:14-30): a sub-tile
// pixel offset that picks the light cone quadrant.
type Offset struct{ DeltaX, DeltaY int8 }

// LightPosition is LightPosition (lighting.h:22-28).
type LightPosition struct {
	// Tile is the world tile the light is centred on.
	Tile Point
	// Offset is the sub-tile pixel offset from Tile.
	Offset Offset
	// Old is the tile the light occupied before its last change.
	Old Point
}

// Light is Light (lighting.h:30-36): one slot of the light pool.
type Light struct {
	Position   LightPosition
	Radius     uint8
	OldRadius  uint8
	IsInvalid  bool
	HasChanged bool
}

// LightGrid is the dLight write boundary ProcessLightList drives. The engine
// functions live outside this package (internal/diablo/levels owns dLight), so
// they are injected; every field must be set.
type LightGrid struct {
	// UnLight is DoUnLight (lighting.cpp:104-115): restores dPreLight over the
	// light's footprint.
	UnLight func(position Point, radius uint8)
	// Lighting is DoLighting (lighting.cpp:117-180): applies the light cone.
	Lighting func(position Point, radius uint8, offset Offset)
	// IsSolid is TileHasAny(position, TileProperties::Solid)
	// (lighting.cpp:392): a light inside a wall is not drawn.
	IsSolid func(position Point) bool
}

// VisionUnlighter is the DoUnVision seam (lighting.cpp:181-190): the pass that
// clears the vision a light left behind. DoVision/DoUnVision belong to the
// vision port, so this package only declares the signature it will be wired
// with and never calls it.
type VisionUnlighter func(position Point, radius uint8)

// LightList is the runtime light pool: Lights, ActiveLights and
// ActiveLightCount (lighting.h:40-42, lighting.cpp:34-36).
type LightList struct {
	// Lights is the pool; an id indexes it.
	Lights [MaxLights]Light
	// Active is ActiveLights: the ids of the lights in the pool, the live ones
	// first ([0, Count)). The slots at [Count, MaxLights) are the free ones,
	// handed out from the end.
	Active [MaxLights]uint8
	// Count is ActiveLightCount: how many ids at the front of Active are live.
	Count int
	// updateLighting is UpdateLighting (lighting.cpp:43): a mutator changed
	// the pool and ProcessLightList has not drawn it yet.
	updateLighting bool
}

// NewLightList is InitLighting (lighting.cpp:240-252): an empty pool whose free
// slots are the identity permutation, so the first MaxLights adds hand out
// 0..31 in order. Upstream also resets VisionActive and TransList there; those
// belong to the vision and automap ports.
func NewLightList() *LightList {
	list := &LightList{}
	list.reset()

	return list
}

// AddLight is AddLight (lighting.cpp:254-274): claims the next free slot and
// returns its id, or NoLight when the pool is full. The slot's stale position
// and radius are left untouched, as upstream does; only the fields it writes
// are reset.
func (l *LightList) AddLight(position Point, radius uint8) int {
	if l.Count >= MaxLights {
		return NoLight
	}

	id := int(l.Active[l.Count])
	l.Count++

	light := &l.Lights[id]
	light.Position.Tile = position
	light.Position.Offset = Offset{}
	light.Radius = radius
	light.IsInvalid = false
	light.HasChanged = false

	l.updateLighting = true

	return id
}

// AddUnLight is AddUnLight (lighting.cpp:276-288): flags a light for removal;
// ProcessLightList restores its footprint and drops it from the pool.
func (l *LightList) AddUnLight(id int) {
	if id == NoLight {
		return
	}

	l.Lights[id].IsInvalid = true

	l.updateLighting = true
}

// ChangeLightRadius is ChangeLightRadius (lighting.cpp:290-306): the light
// stays put and changes radius (the flame-trap animation, objects.cpp:1596).
func (l *LightList) ChangeLightRadius(id int, radius uint8) {
	if id == NoLight {
		return
	}

	l.markChanged(id).Radius = radius
}

// ChangeLightXY is ChangeLightXY (lighting.cpp:308-324): the light moves and
// keeps its radius (a walking player, player.cpp:427).
func (l *LightList) ChangeLightXY(id int, position Point) {
	if id == NoLight {
		return
	}

	l.markChanged(id).Position.Tile = position
}

// ChangeLightOffset is ChangeLightOffset (lighting.cpp:326-345): the light
// keeps its tile and moves its sub-tile cone offset. Upstream returns early
// when the offset is unchanged, so an unchanged offset does not dirty the pool.
func (l *LightList) ChangeLightOffset(id int, offset Offset) {
	if id == NoLight {
		return
	}

	if l.Lights[id].Position.Offset == offset {
		return
	}

	l.markChanged(id).Position.Offset = offset
}

// ChangeLight is ChangeLight (lighting.cpp:347-364): the light moves and
// changes radius in one step (every missile light, missiles.cpp:2852).
func (l *LightList) ChangeLight(id int, position Point, radius uint8) {
	if id == NoLight {
		return
	}

	light := l.markChanged(id)
	light.Position.Tile = position
	light.Radius = radius
}

// ProcessLightList is ProcessLightList (lighting.cpp:366-398): clears the cones
// of the lights that moved or died, redraws the live ones, and returns the dead
// slots to the free end of Active. It is a no-op when nothing changed.
func (l *LightList) ProcessLightList(grid LightGrid) {
	if !l.updateLighting {
		return
	}

	l.clearStaleCones(grid)
	l.redrawLiveLights(grid)

	l.updateLighting = false
}

// CloseLights drops every light. Upstream has no such call: a level change
// re-runs InitLighting (diablo.cpp:3495), which does not clear the pool. This
// resets the pool and zeroes it, so an id surviving a level change cannot
// resurrect a stale position.
func (l *LightList) CloseLights() {
	l.reset()
	l.Lights = [MaxLights]Light{}
}

// markChanged is the shared ChangeLight* prologue (lighting.cpp:300-302,
// :318-320, :339-341, :357-359): flag the light and snapshot the cone currently
// drawn, so ProcessLightList can clear that cone.
func (l *LightList) markChanged(id int) *Light {
	light := &l.Lights[id]
	light.HasChanged = true
	light.Position.Old = light.Position.Tile
	light.OldRadius = light.Radius

	l.updateLighting = true

	return light
}

// clearStaleCones is ProcessLightList's first pass (lighting.cpp:374-383): every
// un-light happens before any light is redrawn.
func (l *LightList) clearStaleCones(grid LightGrid) {
	for i := range l.Count {
		light := &l.Lights[l.Active[i]]
		if light.IsInvalid {
			grid.UnLight(light.Position.Tile, light.Radius)
		}
		if light.HasChanged {
			grid.UnLight(light.Position.Old, light.OldRadius)
			light.HasChanged = false
		}
	}
}

// redrawLiveLights is ProcessLightList's second pass (lighting.cpp:384-395):
// a dead light is swapped out of the live range and its slot becomes the next
// one AddLight hands out; a light buried in a solid tile stays dark.
func (l *LightList) redrawLiveLights(grid LightGrid) {
	for i := 0; i < l.Count; i++ {
		light := &l.Lights[l.Active[i]]
		if light.IsInvalid {
			l.Count--
			l.Active[l.Count], l.Active[i] = l.Active[i], l.Active[l.Count]
			i--
			continue
		}
		if grid.IsSolid(light.Position.Tile) {
			continue
		}
		grid.Lighting(light.Position.Tile, light.Radius, light.Position.Offset)
	}
}

// reset is InitLighting's pool half (lighting.cpp:242-249).
func (l *LightList) reset() {
	l.Count = 0
	l.updateLighting = false
	for i := range l.Active {
		l.Active[i] = uint8(i)
	}
}
