---
# devilGo-yd48
title: gopher-lua dev tooling + QoL wave 1 (stash/xpbar/labels)
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - qol
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-yjuv
blocked_by:
    - devilGo-0tsf
---

# Lua + QoL

## Goal
REPL subset; three QoL modules behind flags.

## Research
lua modules list (§11); qol/stash.cpp, xpbar.cpp, itemlabels.cpp, floatingnumbers.cpp.

## Scope
internal/platform/lua (gopher-lua, audio+dev bindings), internal/platform/qol/{stash,xpbar,itemlabels,floatingnumbers}.

## Acceptance criteria
- [ ] REPL: `dev.player.hp` read + `audio.playSfx(SfxID.Walk)` works headless
- [ ] Stash: 100 pages, grid placement rules, persists via epic-08 store
- [ ] XP bar segment math from Experience.tsv

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — embedding gopher-lua safely (sandbox)

## Dependencies & blockers
dg-task-041, dg-epic-08.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:scope-conflict-with-recorded-Lua-deferral (split QoL from gopher-lua)
