---
# devilGo-23dr
title: CLX sheets (multi-group) + explicit monster ordinal mapping
status: completed
type: task
priority: low
created_at: 2026-10-28T16:35:33Z
updated_at: 2026-10-28T17:11:36Z
parent: devilGo-5wsb
---

## Goal

Make multi-group CLX (sheets) decodable, and make the local monster-ordinal mapping explicit.
Both are latent today, but both turn into silent misdecodes the moment sheet assets or a
C++-ordinal monster id are fed in.

## Evidence

1. **Sheets rejected.** Go `internal/assets/cl2.go` `parseFrameOffsets` errors when the last frame
   offset differs from the buffer size, and `internal/assets/clx.go:27-35` documents rejecting
   sheets. C++ decodes them: `Source/engine/clx_sprite.hpp` `FromBuffer` and
   `Source/utils/cl2_to_clx.cpp:23-28` walk a group-offset table the same way the CEL path does.
   **Verified scope**: all 100 `.clx` files shipped in devilutionX `assets/` are plain frame lists
   (last offset == file size for every one), so nothing in the repo is currently broken by this.
   The risk is retail/monster sheets, which are not in the vendored asset set.
2. **Monster ordinals are local.** `internal/assets/trn.go` defines `MonsterType` values 0..3 for
   councilor/magistrate/cabalist/advocate, while C++ `Source/tables/monstdat.h:266-269`
   `MT_COUNSLR..MT_ADVOCATE` are mid-`_monster_id` enum. A future caller casting a C++ ordinal
   into `MonsterType` gets silently wrong `MonsterTRNSkipAnim` results — the mapping must be an
   explicit conversion, not a cast.

## Acceptance criteria

- [x] `DecodeCLX`/`DecodeCL2` decode a multi-group (sheet) buffer into per-group frame lists,
      matching `clx_sprite.hpp`/`cl2_to_clx.cpp:23-28` semantics, with a test built from a
      multi-group buffer that the current code rejects.
- [x] A single-group buffer still decodes exactly as today (all committed digests unchanged).
- [x] `MonsterType` has an explicit, tested conversion from the C++ `_monster_id` ordinals
      (`monstdat.h:266-269`) — no bare cast is possible without a named function, and an
      out-of-range id is an error.
- [x] A test pins the four types' skip-anim behaviour through the explicit conversion.
- [x] RED first for the sheet path (capture the current rejection), and a comment on the local
      ordinals cites `monstdat.h:266-269` as the mapping source.

Refs: PR #3, devilGo-n89r

## Libraries
- none (stdlib only)

## Skills for coder
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules-notes/SKILL.md` — user amendments, read with coding-rules
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Triage
Delegated read-only brainer audit (pi session `devilGo-5wsb-beansaudit`, report `/home/blaze/.hermes/cache/scratch/pr3-beans-audit.md`); every cited anchor re-verified in the parent against the pinned devilutionX checkout (`ca91d778f`) before filing.


## Proof of Work (2026-10-28, PR #3)

- Worker commits `fc131c9` (sheets) + `034637d` (ordinals); README spec corrected in `eb23ee8`
  (`0x80..0xBE` fill / `0xBF` 65-byte literal / `0xC0..0xFF` literal, citing `clx_decode.hpp:25`).
- Sheets: single-group lists unchanged (all 20 committed CLX golden digests still pass);
  hand-built multi-group buffer now decodes per-group (`TestDecodeCLX/sheet_of_lists`).
- Ordinals: parent enumerated `_monster_id` (`monstdat.h:161..`) — `MT_COUNSLR = 105` confirmed,
  `105..108` pinned in `trn.go` with `MonsterTypeFromMonsterID`; mapping + rejection tests PASS
  (`TestMonsterTypeFromMonsterIDMapsCouncilorFamily`, `...RejectsOthers`).
- Parent gates: gofmt/build/vet clean, 26 pkgs ok / 0 FAIL, `scripts/lint.sh` → `0 issues.`


## Known limitation (brainer check 2026-10-28, Finding 3)

`DecodeCL2(data, width, height)` applies ONE width/height to every frame of every group; C++
`Cl2ToClx` takes per-frame widths (`cl2_to_clx.cpp:66` `PointerOrValue`) and derives heights by
line counting. Sheet support for CL2 is therefore structure-only — real per-frame-dims CL2
sheets cannot be decoded faithfully through this API yet. Recorded in code on `DecodeCL2` by
`devilGo-260w`; per-frame dims tracked for the renderer port.
