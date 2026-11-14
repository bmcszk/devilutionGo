package main

import "time"

// frameDuration is the fixed step the window loop advances the clock by
// (gnTickDelay, diablo.cpp:146).
const frameDuration = 50 * time.Millisecond

// windowTitle is the window caption.
const windowTitle = "devilutionGo"

// keyEvent ordinal values for the quit surface (diablo.cpp:355-356).
const (
	keyEscape keyEvent = iota
	keyQ
	keySpace
)

// keyEvent is the quit surface's key identity: the window layer maps its
// input to these codes (ebiten.Key* ordinals in the boot path).
type keyEvent int
