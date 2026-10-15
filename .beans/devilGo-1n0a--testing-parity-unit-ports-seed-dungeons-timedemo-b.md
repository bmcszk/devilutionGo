---
# devilGo-1n0a
title: 'Testing & parity: unit ports, seed dungeons, timedemo, benchmarks'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - cross-cutting
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-15T16:37:34Z
parent: devilGo-i56j
blocked_by:
    - devilGo-6xu2
---

# Parity harness

## Goal
The proof machine: everything in /tmp/dg-requirements.md §12 + Appendix K/F: ported test families, seed-dungeon matrix, timedemo replay vs reference save, optional C++ cross-check, benchmarks.

## Research
Test inventory §12; timedemo test/timedemo_test.cpp:28-110 (oracle demo_0_reference_spawn_0.sv); fixtures trees; benchmarks (clx/dun/light render, crawl, path, palette); MPQ skip pattern :47-49; demo format demomode.cpp:39-58,595-599.

## Scope
internal/testsupport extensions + internal/demo (record/replay cmd stream) + parity CI job + bench suite; state-hash corpus generator (optional C++ sidecar).

## Acceptance criteria
- [ ] Appendix K matrix: every listed C++ test family has a green Go counterpart (or documented deferral)
- [ ] Timedemo: WarriorLevel1to2 replay → hero bytes equal reference .sv (with MPQs; skip otherwise)
- [ ] Seed matrix: all diablo+hellfire .dun fixtures green in one `just parity` run
- [ ] State-hash diff vs C++ headless on 3 seeds × 2000 ticks: 0 divergences
- [ ] Benchmarks recorded in CI with regression budget (frame <3ms, path 10k<50ms)

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — the whole epic
/home/blaze/.agents/skills/go-dev/SKILL.md — benchmark hygiene

## Dependencies & blockers
blocked_by dg-epic-08 (hero compare), needs epics 02-05 progressively. Feeds every milestone gate.

## Open questions
- C++ cross-check sidecar in CI: keep optional/gated (H1/H14).
