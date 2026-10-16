---
# devilGo-vsx3
title: C++ probe/dumper toolchain for parity fixtures
status: draft
type: task
priority: high
created_at: 2026-10-16T20:36:47Z
updated_at: 2026-10-16T20:36:47Z
parent: devilGo-qna9
---

## Goal

Build the C++-side toolchain that generates every parity fixture the Go port needs: RNG output vectors, backbuffer frame dumps, light-table binaries, seed dungeon maps, save files. This is the single owner of "how a fixture is produced"; every fixture-consuming bean cites this one.

## Research

- devilutionX builds with CMake; simplest reliable route is a small out-of-tree CMake target `dvg-probe` (patch file stored in this repo under tools/probe/) that calls the real functions (rgRandom, MakeLightTable, DRLG routines, pfile save) and writes binary fixtures.
- RNG: rgRandom sequences per seed — dump 1000 outputs for a fixed seed list.
- Frames: headless mode (Source/headless_mode.cpp) renders real backbuffers without a window; capture via capture.cpp path (PCX) or direct framebuffer memcpy + palette.
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
- [ ] probe lighttables subcommand writes tables + falloffs binaries byte-identical to C++ memory layout
- [ ] probe dun subcommand writes generated level dump per level type
- [ ] probe save subcommand writes a real .sv from a scripted run
- [ ] docs/fixtures.md lists every artifact -> producer -> consumer bean
- [ ] Same devilutionX commit rebuilds byte-identical fixtures (hash manifest checked)

## Libraries

- None (C++ side). Go side reads raw bytes only.

## Skills for coder

/home/blaze/.agents/skills/coding-rules/SKILL.md
/home/blaze/.agents/skills/go-dev/SKILL.md

## Dependencies & blockers

- Requires a devilutionX checkout (external repo; pinned commit recorded in docs/fixtures.md)

## Open questions

- None. Decided: out-of-tree probe target, deterministic seeds, hash manifest.
