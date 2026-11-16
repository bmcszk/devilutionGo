package render_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/render"
)

// TestDrawSpriteKeyedSkipsIndexZero pins the CeilDrawSprite keying rule
// (clx_render.cpp CeilDrawSprite: a drawn palette index 0 is skipped, the
// destination pixel survives).
func TestDrawSpriteKeyedSkipsIndexZero(t *testing.T) {
	t.Parallel()

	// given a surface with two distinct indices
	surface := render.NewSurface(2, 1)
	surface.Set(0, 0, 7)
	surface.Set(1, 0, 8)

	// when a keyed sprite with index 0 and 9 is drawn over it
	clip := render.Clip{X: 0, Y: 0, Width: 2, Height: 1, Pixels: []uint8{0, 9}}
	surface.DrawSpriteKeyed(clip)

	// then index 0 leaves destination untouched, index 9 draws
	if got := surface.At(0, 0); got != 7 {
		t.Fatalf("pixel 0 = %d, want the destination 7 to survive keyed index 0", got)
	}

	if got := surface.At(1, 0); got != 9 {
		t.Fatalf("pixel 1 = %d, want the sprite 9 drawn", got)
	}
}
