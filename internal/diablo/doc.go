// Package diablo holds the Diablo-specific logic: the TSV-driven game data
// tables, the cross-cutting enums, the missile, combat and animation
// pipelines, the .sv save codec, and the committed-fixture parsers (dun
// corpus, demo tapes).
//
// Import rule (binding, see AGENTS.md): this package may import
// internal/engine and internal/game; nothing imports this package from those
// layers.
//
// Seams:
//
//	diablo/txtdata holds only the table READERS: the TSV parse MECHANISM
//	stays in engine/txtdata.
//	diablo/testsupport owns only the Diablo fixture formats: the golden-grid
//	diff and the state hash stay in game/testsupport.
//	diablo/missiles probes tile collisions against the world grid owned by
//	game/world.
package diablo
