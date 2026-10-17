---
# devilGo-1n0a
title: 'Testing & parity: unit ports, seed dungeons, timedemo, benchmarks'
status: todo
type: epic
priority: normal
tags:
    - plan:epic
    - cross-cutting
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-17T19:56:52Z
parent: devilGo-i56j
blocked_by:
    - devilGo-6xu2
---

# Parity harness

## Goal
The proof machine: everything in docs/research/dg-requirements.md §12 + Appendix K/F: ported test families, seed-dungeon matrix, timedemo replay vs reference save, optional C++ cross-check, benchmarks.

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
2 local: Cites verified: test/timedemo_test.cpp exists, RunTimedemo :33 + fixtures path :55 (body 28-110 ✓); docs/research/dg-requirements.md Appendix K at line 1038 ✓ (also committed in docs/requirements.md per milestone log). Skills ✓.
3 online: n/a.
4 solution: none — ready. C++ sidecar correctly gated optional (H1/H14). AC "every listed family green or documented deferral" is checkable against Appendix K.
5 understanding: cross-layer (internal/testsupport + internal/demo); feeds every milestone gate — correctly blocked on hero-compare capability.
6 repro: `grep -n "RunTimedemo" test/timedemo_test.cpp | head -2`; `grep -n "Appendix K" docs/research/dg-requirements.md`
7 fix: none.
8 validation: children (yv1r, uc44, w9or, w6qf, i2cr, hy8e, 41p0) triaged in batch-10.
9 related: blocker 6xu2; sibling zo8j (divergence), qbiy/7aul (frame dumps).
VERDICT: TRIAGED-OK

## TRIAGE 2026-10-18 (pi wave, batch 2)

devilGo-1n0a** (Testing & parity epic, parent i56j, status draft)
1 Root: proof machine (test ports, seed matrix, timedemo oracle, state-hash, benches) — matches body §Goal + IMPL NOTES.
2 Local: YAML `blocked_by: [devilGo-6xu2]` ✓ matches "dg-epic-08 hero compare". `RunTimedemo` at test/timedemo_test.cpp:33, `WarriorLevel1to2` :107, fixtures path :55 (body range 28-110 ✓). Appendix K at docs/research/dg-requirements.md:1038 ✓. All 7 children (yv1r/uc44/w9or/w6qf/i2cr/hy8e/41p0) exist ✓.
3 Online: SKIPPED — no doc claims needing web (Ebiten perf cited in qbiy only).
4 Scope-fit: cross-cutting internal/testsupport+demo; no overlap leakage — 7aul/41p0 correctly children, uc44/w9or/w6qf split into child beans, not duplicates.
5 DDD: engine/test layer feeding game+diablo gates; correct as cross-cutting epic.
6 Dep-order: hard blocker 6xu2 wired; soft "needs 02-05 progressively" consistent with sibling wiring.
7 AC-gaps: all 5 ACs binary-checkable (Appendix K sweep, .sv byte equality, just-parity, 0-divergence hash, bench budgets) — no gap.
8 Validation: `grep -n RunTimedemo test/timedemo_test.cpp`; `just parity`; bench CI job output vs <3ms/<50ms budgets.
9 Related: blocker 6xu2; children batch-10; siblings zo8j, qbiy/7aul.
**TRIAGED-OK:devilGo-1n0a**


Wave outputs: /tmp/dg-triage/out-b2.md

## Libraries

- Phase rule (DECIDED 2026-10-17): pure-Go only, no CGO. Phase-0 packages ran stdlib-only; concrete deps (modernc.org/sqlite, suprsokr/go-mpq, gopher-lua, ebiten) attach at the beans that consume them — never substituted by a worker.
