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
package game
