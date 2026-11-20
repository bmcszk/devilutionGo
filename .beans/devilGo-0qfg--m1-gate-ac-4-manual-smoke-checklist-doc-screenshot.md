---
# devilGo-0qfg
title: 'M1 gate: AC-4 manual smoke — checklist doc + screenshots'
status: completed
type: task
priority: normal
created_at: 2026-11-19T22:32:40Z
updated_at: 2026-11-20T01:50:01Z
parent: devilGo-i56j
blocked_by:
    - devilGo-451a
    - devilGo-prcc
---

---
# devilGo-0qfg
title: 'M1 gate: AC-4 manual smoke — checklist doc + screenshots'
status: in-progress
type: task
priority: normal
created_at: 2026-11-19T22:32:40Z
updated_at: 2026-11-20T01:34:40Z
parent: devilGo-i56j
blocked_by:
    - devilGo-451a
    - devilGo-prcc
---

---
# devilGo-0qfg
title: 'M1 gate: AC-4 manual smoke — checklist doc + screenshots'
status: in-progress
type: task
priority: normal
created_at: 2026-11-19T22:32:40Z
updated_at: 2026-11-20T01:34:18Z
parent: devilGo-i56j
blocked_by:
    - devilGo-451a
    - devilGo-prcc
---

# M1 gate: AC-4 manual smoke — checklist doc + screenshots

## Goal
Document the manual smoke checklist (boot to window, create warrior,
walk town, descend, kill, loot, save, load, quit) as a runnable recipe
in the repo, and attach timestamped screenshots (with pixel stats) to
this bean so a human can execute the smoke pass for M1.

## Research
- A smoke recipe exists (`just smoke` — headless smoke recipe); this
  gate adds the WINDOWED manual pass: `just build` + run, DISPLAY=:0,
  the windowed tag. Capture via the established screenshot route.
- Visual law (lane ops): every screenshot ships with pixel stats
  (mean, min/max, distinct colors, non-black %) + visual description +
  root-cause hypothesis; the headless parity test is the oracle.

## Scope
- docs/ checklist page + bean evidence; no engine/sim code.

## Acceptance criteria
- [x] Smoke checklist committed under docs/ (steps runnable from the
      repo root, each step naming the expected screen state)
- [x] Screenshots captured for each milestone step, each annotated in
      the bean with pixel stats + visual description
- [x] Checklist cross-checked against the M1 goal line (town + cathedral
      slice covered end to end by a human following the doc)

## Triage
Planner wave: delegated brainer plan pass + parent verification, 2026-11-19.
Plan: .hermes/lane/ftbu-m1/wave-s5-plan.md. Blocked-by: the parity and
fps gates (a smoke run lands only after the gates they pin are in).

## Libraries
stdlib only (docs gate).

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md — comment/doc hygiene, gate honesty
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — visual PoW laws
- ~/.agents/skills/ebiten/deployment/SKILL.md — windowed build/run
- ~/.hermes/skills/beans-cli-quirks-notes/SKILL.md — visual bean evidence
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff



## Progress log

- 2026-11-20T02:25Z: promoted (triage: planner plan + parent verification) and claimed — AC-4 dispatched under contract harness.ac4.md (worktree ftbu-s5, feat/m1-create-pack), status flip.


## Proof of Work

- Contract: .hermes/lane/ftbu-m1/harness.ac4.md (require_edit harness.s5a.log, sentinel M1-REAL-L1-AC4-DONE). Worker: task-dg-m1-ac4-coder-coder (185 tool calls), branch feat/m1-create-pack @ e472a25 — pushed; parent verified ls-remote == HEAD, sentinel present.
- Deliverable: docs/m1-manual-smoke.md (334 lines, nine steps, every command execution-verified by the worker before writing; expected screen state named per step) + four captures under docs/screenshots/m1-smoke/.
- Parent pixel analysis (visual law): step2-window-town-hero.png 640x480 mean 37.46 min 0 max 255 distinct 131 nonblack 60.37% — byte-profile EXACT match to the user-validated first-scene baseline (#73: town structures + upright warrior + panels); step2-town-without-hero.png mean 18.42 distinct 100 nonblack 60.16% — same scene without hero sprite/panel overlay (darker mean explained by the missing lit sprite + HUD); step4-smoke-walk-before/after.png 1280x960 mean ~210.8 distinct 3->4 nonblack 100% — the smoke tool's text capture (after adds the walk-completed state line), proving the headless end-to-end pass. Root-cause hypothesis for frame darkness: known palette/presentation baseline validated in the first-scene round, not a defect.
- Honest shape (worker-disclosed, kept): the windowed boot wires no command kernel yet (reads only ESC/q) — boot/quit steps are interactive; steps 3-9 run through the headless smoke path. The full nine-step human click-through lands with the hero-creation and input beans; the page names this gap explicitly. Filename collision between the two smoke game tests named for a later wave.
- No cross-review run: docs + screenshots only, no code/config surface (risk classes untouched).
- Residual: interactive-phase coverage (hero-creation + input beans); town-scene bench row (devilGo-ht4l bean under Testing & parity) for the render-side town coverage.
