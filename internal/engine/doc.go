// Package engine holds the project-agnostic game engine: core loop, ECS,
// rendering primitives, asset pipeline, audio, input, networking, save/load,
// and configuration.
//
// Import rule (binding, see AGENTS.md): this package must never import
// internal/game or internal/diablo. Code here must stay reusable for any
// isometric ARPG, not just Diablo.
package engine
