---
# devilGo-7vif
title: Font ColorTranslations table belongs in internal/assets
status: completed
type: task
priority: normal
created_at: 2026-10-28T16:35:33Z
updated_at: 2026-10-28T17:11:36Z
parent: devilGo-5wsb
---

## Goal

Own the font `ColorTranslations` table in `internal/assets`, where it belongs: it is static data,
not draw-path code.

## Evidence

- `Source/engine/render/text_render.cpp:87-111` is a literal
  `std::array<const char *, 19>` of font TRN paths with three `nullptr` entries at indices 4, 5
  and 7 (`ColorDialogWhite`, `ColorDialogRed`, and an unnamed gap) — and 16 non-null paths
  (`fonts\\goldui.trn`, `grayui.trn`, `golduis.trn`, `grayuis.trn`, `yellow.trn`, `black.trn`,
  `white.trn`, `whitegold.trn`, `red.trn`, `blue.trn`, `orange.trn`, `buttonface.trn`,
  `buttonpushed.trn`, `gamedialogwhite.trn`, `gamedialogyellow.trn`, `gamedialogred.trn`).
- Only the lazy per-colour load cache (`text_render.cpp:113`, `:190-192`) is draw-path. The table
  itself is data of exactly the same kind `TRNRegistry` already owns.
- PR #3 deferred this to "the renderer/text port". That is wrong for the data half: two of the 16
  paths (`fonts/red.trn`, `fonts/whitegold.trn`) are already committed Go fixtures under
  `internal/testdata/gfx/trn/`, and `devilGo-b782`'s own Goal is to own "the *data* side of
  colour and all `.trn` files".

## Acceptance criteria

- [x] `internal/assets` exposes the 19-slot colour-translation table: 16 paths byte-identical to
      `text_render.cpp:88-110`, 3 null slots at the documented indices, and the slot order is
      asserted by index (a test names each index and its expected path or null).
- [x] A test asserts the exact count (19 slots, 16 non-null, 3 null) so a slot cannot silently
      drift.
- [x] Resolution of a slot to a loaded `TRN` reuses the existing TRN registry/loader; a null slot
      resolves to "no TRN" rather than an error.
- [x] Committed fixtures are used where they exist (at least `fonts/red.trn`,
      `fonts/whitegold.trn`), with pinned digests; no test reads the devilutionX checkout.
- [x] The bean records that only the lazy per-draw load cache remains for the renderer port.

Refs: PR #3, devilGo-b782, devilGo-qbiy

## Libraries
- none (stdlib only; data table + tests)

## Skills for coder
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules-notes/SKILL.md` — user amendments, read with coding-rules
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Triage
Delegated read-only brainer audit (pi session `devilGo-5wsb-beansaudit`, report `/home/blaze/.hermes/cache/scratch/pr3-beans-audit.md`); every cited anchor re-verified in the parent against the pinned devilutionX checkout (`ca91d778f`) before filing.


## Proof of Work (2026-10-28, PR #3)

- Worker commits `5406303` (table) — `color_translations.go` byte-exact to `text_render.cpp:87-111`
  in upstream order, nulls at 4/5/7; `ColorTranslations()`, `ColorTranslationPath`, and
  `(*TRNRegistry).ColorTranslationTRN` (null slot → no TRN, no error).
- Parent re-verification: full table diff against the C++ array (19/19 entries, order + backslash
  paths match); fixtures `fonts_red.trn.hex`, `fonts_whitegold.trn.hex` exist and are loaded with
  pinned digests (`TestColorTranslationFixtureSlots`); count test (`TestColorTranslationSlotCounts`),
  named-index test, null-slot test all PASS. Gates: gofmt/build/vet clean, 26 pkgs ok / 0 FAIL,
  `scripts/lint.sh` `0 issues.`
- Residual recorded: only the lazy per-draw load cache remains for the renderer port.
