// Package game holds the shared game systems tuned for Diablo-style games:
// the game-loop state on the engine core and the DRLG world grid.
//
// Import rule (binding, see AGENTS.md): this package may import
// internal/engine and must never import internal/diablo.
//
// Seams:
//
//	game/world owns the world-grid state and reads no Diablo data.
//	game/testsupport is test-only tooling: the golden-grid diff and the
//	canonical state hash.
//	DoomPortal (doom.go) and DebugRouter (debug.go) carry the ported easter egg
//	and cheat state; both are inert until a host wires them, so default builds
//	behave exactly as before (Game.Doom and Game.Debug stay nil).
package game
