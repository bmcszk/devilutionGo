// Package integration holds the headless, kernel-level integration tests: a
// whole scripted session (boot → create → town → dungeon → kills → loot →
// save → load) wired from the landed surfaces (world slice, input bridge,
// combat rolls, persist store) and frozen into per-step StateHashes — the
// M1 replayability gate (devilGo-ftbu AC-1, ADR-015/018).
package integration
