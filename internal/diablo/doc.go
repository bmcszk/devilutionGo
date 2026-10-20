// Package diablo holds all Diablo-specific logic: data definitions, level
// formats, quest logic, monster AI, item tables, spells, town services, and
// the save-file schema.
//
// Import rule (binding, see AGENTS.md): this package may import
// internal/engine and internal/game; nothing imports this package from those
// layers.
//
// Package map:
//
//	diablo             — the Diablo layer root (this manifest).
//	diablo/animation   — AnimationInfo and the animation progress variable.
//	diablo/combat      — the combat roll pipeline (PlrHitMonst).
//	diablo/enums       — the cross-cutting constants that are not TSV-driven.
//	diablo/gamedata/*  — the TSV-driven data tables (audio, leveldat, misdat,
//	                     monstdat, objdat, quests, textdat, towners).
//	diablo/missiles    — the missile state and the per-tick missile logic.
//	diablo/save        — the .sv save codec.
//	diablo/testsupport — the Diablo fixture parsers (dun corpus, demo tapes).
//	diablo/txtdata     — the Diablo table READERS (classes, experience, mana,
//	                     spell data).
//
// Seams:
//
//	diablo/txtdata consumes engine/txtdata.Table: the parse mechanism stays
//	in the engine layer, only the table READERS are Diablo-specific.
//	diablo/testsupport consumes game/testsupport for the grid diff and the
//	state hash; it owns only the Diablo fixture formats.
package diablo
