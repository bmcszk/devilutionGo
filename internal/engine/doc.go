// Package engine holds the project-agnostic game engine: core loop, ECS,
// rendering primitives, asset pipeline, audio, input, networking, save/load,
// and configuration.
//
// Import rule (binding, see AGENTS.md): this package must never import
// internal/game or internal/diablo. Code here must stay reusable for any
// isometric ARPG, not just Diablo.
//
// Package map:
//
//	engine         — crawl generation, A* path finding, the seeded RNG and
//	                 the tick clock: the engine's reusable core.
//	engine/assets  — the MPQ archive reader (asset pipeline input).
//	engine/demo    — the demo-mode stream format (version 3) and its
//	                 record/replay drivers.
//	engine/txtdata — the TSV table parser and the mod-override stacking.
//
// Seams:
//
//	engine/assets and engine/demo are data-format only: they know nothing of
//	Diablo tables or of game state.
//	engine/txtdata owns the table MECHANISM (parse, override); the
//	Diablo-specific table READERS live in internal/diablo/txtdata and
//	consume engine/txtdata.Table.
package engine
