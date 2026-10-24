---
# devilGo-vsx3
title: C++ probe/dumper toolchain for parity fixtures
status: completed
type: task
priority: high
created_at: 2026-10-16T20:36:47Z
updated_at: 2026-10-17T10:55:33Z
parent: devilGo-qna9
---

## Goal

Build the C++-side toolchain that generates every parity fixture the Go port needs: RNG output vectors, backbuffer frame dumps, light-table binaries, seed dungeon maps, save files. This is the single owner of "how a fixture is produced"; every fixture-consuming bean cites this one.

## Research

- devilutionX builds with CMake; simplest reliable route is a small out-of-tree CMake target `dvg-probe` (patch file stored in this repo under tools/probe/) that calls the real functions (SetRndSeed/GenerateRandomNumber (random.cpp:16-96), MakeLightTable, DRLG routines, pfile save) and writes binary fixtures.
- RNG: SetRndSeed/GenerateRandomNumber (random.cpp:16-96) sequences per seed — dump 1000 outputs for a fixed seed list.
- Frames: headless mode (Source/headless_mode.cpp) WARNING (triage 2026-10-16): HeadlessMode SKIPS rendering (dx.cpp:198/:289, scrollrt.cpp:1816/:1836) — it cannot produce frames. frames subcommand = dump hook in the render path (pattern: 7aul backbuffer_dump_test.cpp) or non-headless run with SDL_VIDEODRIVER=dummy + capture.cpp; palette dumped alongside frames. vsx3 owns the generic dump mechanism; 7aul consumes it for its golden matrix.
- Light tables: call MakeLightTable + falloff init, dump byte arrays (16x256 tables, 16x128 falloffs).
- Seed dungeons: run DRLG for fixed seeds per level type, dump generated dun + tile indices.
- Saves: scripted short headless run, save via pfile, capture the .sv.

## Scope

- tools/probe/ (patch against devilutionX, stored here): CMake target + probe source, one subcommand per fixture class.
- Deterministic: fixed seed list committed (tools/probe/seeds.txt); byte-identical output on rebuild.
- docs/fixtures.md documenting every artifact, producer subcommand, consumer bean.

## Acceptance criteria

- [ ] probe rng subcommand writes RNG vector file, format documented
- [ ] probe frames subcommand writes raw backbuffer + palette pair
- [ ] probe lighttables subcommand writes tables + falloffs + cone binaries byte-identical to C++ memory layout (kx0l consumes cone.bin); leveltype set per variant to hit all four MakeLightTable branches (light_tables.cpp:20-46)
- [ ] probe dun subcommand writes generated level dump per level type
- [ ] probe save subcommand writes a real .sv from a scripted run; pin/zero dwLowDateTime/dwHighDateTime (pack.h:34-35) so byte-identical-on-rebuild holds for saves
- [ ] docs/fixtures.md lists every artifact -> producer -> consumer bean
- [ ] Same devilutionX commit ca91d778f rebuilds byte-identical fixtures (hash manifest checked)

## Libraries

- None (C++ side). Go side reads raw bytes only.

## Skills for coder

/home/blaze/.agents/skills/coding-rules/SKILL.md
/home/blaze/.agents/skills/go-dev/SKILL.md

## Dependencies & blockers

- Requires a devilutionX checkout (external repo; pinned commit recorded in docs/fixtures.md)

## Open questions

- None. Decided: out-of-tree probe target, deterministic seeds, hash manifest.

TRIAGE 2026-10-16 (pi _glm_first READ-ONLY wave): NEEDS-WORK resolved in body — frames premise corrected (HeadlessMode skips rendering; dump-hook route), rgRandom->real symbols, cone+leveltype AC, save timestamp pinning, 7aul demarcation, commit pinned ca91d778f. VERDICT: TRIAGED-OK after fixes. Consumers: 1spo, m2x7, kx0l, xcdz, 7aul.


## Proof of Work
Worker (pi _glm_flash_first): built dvg_probe against pinned devilutionX ca91d778f (libdevilutionx_so.so + 188/188 ninja targets); produced 282 artifacts (rng/lighttables/dun/frames/save) under tools/probe/out/ with manifest.txt.
Orchestrator independent re-gate:
- Determinism: /tmp/vsx3-run1/manifest.txt sha256 == /tmp/vsx3-run2/manifest.txt == committed tools/probe/out/manifest.txt (4c64a15108fe1f26...) — TREES-IDENTICAL via diff -r.
- Artifact classes present: rng (8 seed files), dun, frames, lighttables, save.
- Patch + README + seeds.txt committed; scope respected (no engine/ code touched).
Residual: DIABDAT.MPQ-dependent artifacts validated against worker's local copy; final parity binding happens in consuming beans (1spo storage, m2x7 vectors already green).
