---
# devilGo-4jo7
title: Turn engine (nthread port) + delta sync
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - net
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:26:52Z
parent: devilGo-zo8j
blocked_by:
    - devilGo-omu6
    - devilGo-mfds
---

# Lockstep

## Goal
Turn cadence, in-transit batching, delta export/import.

## Research
nthread.cpp:80-176; base.h PlayerState; msg.cpp:218-275 deltas, :698; multi.cpp:665.

## Scope
internal/net/turn.go, delta.go integrated with engine tick.

## Acceptance criteria
- [ ] Epic's 10k-tick zero-divergence and converge tests green
- [ ] Delta payload byte-size within C++ norms for same scenarios

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — channel-based turn pump, deterministic apply order

## Dependencies & blockers
dg-task-038, dg-task-039.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:missing-blocked-by-mfds + wrong line cite (698→2774)


## IMPL NOTES (triage 2026-10-16)
1 root: lockstep turn cadence (nthread), in-transit turn batching, delta export/import.
   Parent devilGo-zo8j, blocked_by devilGo-omu6 only.
2 local: nthread.cpp:80 `nthread_send_and_recv_turn` + turns-in-transit loop verified ✓
   (research range 80-176 ✓); msg.cpp:218-275 DLevel/DMonsterStr/DObjectStr/DJunk delta
   structs verified ✓; multi.cpp:665 `multi_msg_countdown` ✓; dvlnet/base.h exists ✓.
   **msg.cpp:698 is `return src;`** — DeltaExportData is at **msg.cpp:2774** (line-cite wrong
   here and in epic). Body deps list dg-task-038 (omu6) AND dg-task-039 (**mfds** — delta
   apply goes through cmd dispatch) but blocked_by omits mfds.
3 online: n/a.
4 solution: add devilGo-mfds to blocked_by (delta import runs through ParseCmd handlers);
   fix research cite msg.cpp:698 → msg.cpp:2774 (and same fix in epic zo8j body).
5 understanding: engine/net turn pump integrated with the engine tick (Appendix L
   Has500msPassed); deltas are the authoritative level-state sync for mid-game joins.
6 repro: `sed -n '80,101p' Source/nthread.cpp`; `grep -n "DeltaExportData" Source/msg.cpp`
   → 2774; `sed -n '218,260p' Source/msg.cpp`.
7 fix: `beans update devilGo-4jo7 --blocked-by devilGo-mfds` (append; keep omu6), plus
   `--body-replace-old "msg.cpp:218-275 deltas, :698" --body-replace-new "msg.cpp:218-275
   deltas, DeltaExportData :2774"`.
8 validation: `beans show devilGo-4jo7 --json | jq .blocked_by` contains both omu6 and mfds.
9 related: devilGo-omu6, devilGo-mfds, devilGo-vgxq, devilGo-zo8j.
VERDICT: NEEDS-WORK:missing-blocked-by-mfds + wrong line cite (698→2774)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
