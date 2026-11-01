package game

import (
	"errors"
	"fmt"
)

// doomSpritePath is the cel the map of the stars draws (doom.cpp:25).
const doomSpritePath = `items\map\mapztown`

// DoomSpriteWidth is the LoadCel cel width for the doom sprite (doom.cpp:25).
const DoomSpriteWidth = 640

// ErrDoomNoLoader is returned by Init when no sprite loader was injected.
var ErrDoomNoLoader = errors.New("doom: no sprite loader")

// DoomSpriteLoader loads a cel sprite list and reports its frame count: the
// renderer-free stand-in for LoadCel (doom.cpp:25), which aborts on a missing
// cel.
type DoomSpriteLoader func(path string, width int) (frames int, err error)

// DoomPortal is the map of the stars easter egg state: upstream's DoomFlag
// plus the loaded sprite list (doom.cpp:13-41). Using the map of doom item
// raises the flag (items.cpp:4319-4321 IMISC_MAPOFDOOM); Escape and the chat
// log clear it (diablo.cpp:355-356, chatlog.cpp:114). Blitting DoomSprite[0]
// at the UI rectangle offset (0, 352) (doom.cpp:42-43) belongs to the renderer.
type DoomPortal struct {
	load   DoomSpriteLoader
	frames int
	active bool
}

// NewDoomPortal returns an inactive portal that loads its cel through load.
func NewDoomPortal(load DoomSpriteLoader) *DoomPortal {
	return &DoomPortal{load: load}
}

// Init loads the doom sprite and raises DoomFlag (doom.cpp:23-27). The flag
// only rises on a successful load, matching LoadCel's abort on a missing cel.
func (p *DoomPortal) Init() error {
	if p.load == nil {
		return ErrDoomNoLoader
	}

	frames, err := p.load(doomSpritePath, DoomSpriteWidth)
	if err != nil {
		return fmt.Errorf("doom: load %s: %w", doomSpritePath, err)
	}

	p.frames = frames
	p.active = true

	return nil
}

// Close lowers DoomFlag and frees the sprite (doom.cpp:29-33). It is safe to
// call when the portal was never opened.
func (p *DoomPortal) Close() {
	p.active = false
	p.frames = 0
}

// Active reports DoomFlag (doom.cpp:21): whether doom_draw draws anything.
func (p *DoomPortal) Active() bool { return p.active }

// Frames is the loaded sprite list length; 0 means the sprite is freed.
func (p *DoomPortal) Frames() int { return p.frames }
