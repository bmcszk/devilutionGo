// Package engine holds the project-agnostic engine core: crawl generation,
// A* path finding, the seeded RNG and the tick clock.
//
// Import rule (binding, see AGENTS.md): this package must never import
// internal/game or internal/diablo. Code here must stay reusable for any
// isometric ARPG, not just Diablo.
//
// Seams:
//
//	engine/assets and engine/demo are data-format only: they know nothing of
//	Diablo tables or of game state.
//	engine/txtdata owns the table MECHANISM (parse, override); the
//	Diablo-specific table READERS live in internal/diablo/txtdata and
//	consume engine/txtdata.Table.
package engine
