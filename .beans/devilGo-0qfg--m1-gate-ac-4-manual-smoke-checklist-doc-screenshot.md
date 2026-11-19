---
# devilGo-0qfg
title: 'M1 gate: AC-4 manual smoke — checklist doc + screenshots'
status: draft
type: task
priority: normal
created_at: 2026-11-19T22:32:40Z
updated_at: 2026-11-19T22:33:55Z
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
- [ ] Smoke checklist committed under docs/ (steps runnable from the
      repo root, each step naming the expected screen state)
- [ ] Screenshots captured for each milestone step, each annotated in
      the bean with pixel stats + visual description
- [ ] Checklist cross-checked against the M1 goal line (town + cathedral
      slice covered end to end by a human following the doc)

## Triage
Planner wave: delegated brainer plan pass + parent verification, 2026-11-19.
Plan: .hermes/lane/ftbu-m1/wave-s5-plan.md. Blocked-by: the parity and
fps gates (a smoke run lands only after the gates they pin are in).

## Libraries
stdlib only (docs gate).

## Skills for coder
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md — visual PoW laws
- ~/.agents/skills/ebiten/deployment/SKILL.md — windowed build/run
- ~/.hermes/skills/beans-cli-quirks-notes/SKILL.md — visual bean evidence
