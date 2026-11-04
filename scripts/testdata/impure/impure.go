//go:build ignore

// Package impure is a deliberately impure fixture: the simpurity gate must
// report every banned line here. The ignore tag keeps the Go formatters and the
// linter out; the gate parses sources regardless of build tags.
package impure

import (
	"os"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// Tick draws a frame, reads the wall clock and the disk, all banned for sim code.
func Tick() int64 {
	render.Draw()
	ebiten.SuspendAudio()
	_ = os.Getpid

	return time.Now().Unix()
}
