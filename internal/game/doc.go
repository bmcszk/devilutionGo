// Package game holds shared game systems tuned for Diablo-style games:
// isometric grid, line-of-sight, path-finding, inventory UI, stats, and combat
// formulas.
//
// Import rule (binding, see AGENTS.md): this package may import
// internal/engine and must never import internal/diablo.
//
// Package map:
//
//	game             — the shared game state that sits on the engine core.
//	game/world       — the DRLG world-grid state and the tile properties
//	                   (isometric grid, walkability).
//	game/testsupport — the golden-grid diff reporter and the canonical
//	                   state hashing shared by the parity tests.
//
// Seams:
//
//	game/world owns grid state and reads no Diablo data.
//	game/testsupport is test-only tooling: it imports nothing from
//	internal/diablo.
package game
