package game

import (
	"github.com/bmcszk/devilutionGo/internal/engine/demo"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
	"github.com/bmcszk/devilutionGo/internal/game/world"
)

// leftButton is the recorded left mouse button (SDL_BUTTON_LEFT).
const leftButton = 1

// TileAt resolves a recorded pointer position to the dungeon tile under it. The
// camera is presentation state, so the host owns this mapping and the replay
// stays free of scroll offsets.
type TileAt func(x, y uint16) world.Tile

// ClickRouter turns a recorded demo stream into the gameplay commands its clicks
// raise: a left press on a live target attacks that tile, any other left press
// walks to it. The gate decides whether the command may run.
type ClickRouter struct {
	view   world.View
	tileAt TileAt
}

// NewClickRouter binds the mapping to the state view it targets and to the
// host's pointer-to-tile resolver.
func NewClickRouter(view world.View, tileAt TileAt) *ClickRouter {
	return &ClickRouter{view: view, tileAt: tileAt}
}

// Command maps one demo record to the command it raises. Records that are not a
// left button press raise no gameplay command; the raw input path already
// carries those (DemoReplay).
func (r *ClickRouter) Command(rec demo.Record) (kernel.Command, bool) {
	click, ok := rec.(demo.MouseButtonRecord)
	if !ok || !click.Down || click.Button != leftButton {
		return kernel.Command{}, false
	}

	tile := r.tileAt(click.X, click.Y)

	return kernel.Command{Kind: r.clickKind(tile), A: int32(tile.X), B: int32(tile.Y)}, true
}

// clickKind is the verb the click means: an attack when a live target stands on
// the tile, a walk otherwise (diablo.cpp click handling).
func (r *ClickRouter) clickKind(tile world.Tile) kernel.Kind {
	if alive, ok := r.view.TargetAt(tile); ok && alive {
		return kernel.KindPlayerAttack
	}

	return kernel.KindPlayerMove
}
