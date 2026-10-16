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
updated_at: 2026-10-16T14:06:43Z
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

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: The proof machine — test families, seed matrix, timedemo oracle, state-hash cross-check, benches. Wired blockedBy 6xu2 ✓ (body "dg-epic-08 hero compare"). 7 children.
2 local: Cites verified: test/timedemo_test.cpp exists, RunTimedemo :33 + fixtures path :55 (body 28-110 ✓); /tmp/dg-requirements.md Appendix K at line 1038 ✓ (also committed in docs/requirements.md per milestone log). Skills ✓.
3 online: n/a.
4 solution: none — ready. C++ sidecar correctly gated optional (H1/H14). AC "every listed family green or documented deferral" is checkable against Appendix K.
5 understanding: cross-layer (internal/testsupport + internal/demo); feeds every milestone gate — correctly blocked on hero-compare capability.
6 repro: `grep -n "RunTimedemo" test/timedemo_test.cpp | head -2`; `grep -n "Appendix K" /tmp/dg-requirements.md`
7 fix: none.
8 validation: children (yv1r, uc44, w9or, w6qf, i2cr, hy8e, 41p0) triaged in batch-10.
9 related: blocker 6xu2; sibling zo8j (divergence), qbiy/7aul (frame dumps).
VERDICT: TRIAGED-OK
