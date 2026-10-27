---
# devilGo-yeia
title: Town + setmaps + triggers
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - world
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-21T11:07:45Z
parent: devilGo-97mn
blocked_by:
    - devilGo-76tu
    - devilGo-1e1b
    - devilGo-fpkz
---

# Static maps

## Goal
Town.dun load + warp gating; all SL_* setmaps + their objects/triggers/palettes.

## Research
town.cpp:210-234 IsWarpOpen; setmaps.cpp:112-165 (sklkng/bonecha/foulwatr/vile + arena maps) with per-map palettes/transparency; trigs.cpp Init*Triggers; drlg_quests.

## Scope
internal/world/{town,setmaps,trigs}.go.

## Acceptance criteria
- [ ] Town loads from extracted loose assets (go:embed; extract-at-build decision 2026-10-16), warps respect quest flags
- [ ] All setmaps byte-load (dPiece grids) equal fixtures; triggers fire at documented tiles

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/go-dev/SKILL.md — data-driven map loaders

## Dependencies & blockers
devilGo-76tu.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:ac-stale-vs-decision — "loads from MPQ" contradicts recorded extract-at-build decision; "Town.dun" file does not exist (4 sector .dun files).


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-1 task (devilGo-yeia): town load + warp gating + all SL_* setmaps; draft; wire blockedBy [76tu] ✓ = body's devilGo-76tu.
2 local: Cites verified: town.cpp:210-234 IsWarpOpen gates ✓ (210/215/228-233 region); setmaps.cpp:112-165 LoadSetMap ✓ (sklkngt.dun/l1_2.pal at :123-124); Init*Triggers coords ✓ (trigs.cpp:114-143). TWO defects vs recorded decisions/live tree: (a) AC "Town loads from MPQ" conflicts with the milestone's recorded extract-at-build decision (loose assets, no runtime MPQ) — stale wording, needs "loads from extracted loose assets"; (b) Goal says "Town.dun load" but NO town.dun exists — town is assembled from 4 sector files (town.cpp:202-205 FillSector sector1s-4s.dun + automap.dun :207, verified live).
3 online: n/a.
4 solution: two one-line wording fixes; dependency and the rest of research are sound.
5 understanding: world layer static-map loader consuming 76tu's engine + fpkz's trigger tables (edge wired via fpkz fix); setmap dPiece byte-load vs fixtures is the right gate.
6 repro: `sed -n '202,207p;210,234p' devilutionX/Source/levels/town.cpp; find devilutionX/assets -name "town.dun"; sed -n '112,125p' devilutionX/Source/levels/setmaps.cpp`
7 fix: `beans update devilGo-yeia --body-replace-old "Town.dun load + warp gating; all SL_* setmaps + their objects/triggers/palettes." --body-replace-new "Town sector load (levels/towndata/sector1s-4s.dun via FillSector town.cpp:202-205 + automap.dun; no single town.dun exists) + warp gating; all SL_* setmaps + their objects/triggers/palettes."`; `beans update devilGo-yeia --body-replace-old "- [ ] Town loads from MPQ, warps respect quest flags" --body-replace-new "- [ ] Town loads from extracted loose assets (extract-at-build decision, milestone devilGo-i56j), warps respect quest flags"`
8 validation: body has no "MPQ"/"Town.dun" wording; AC references the recorded decision.
9 related: 76tu (dep), fpkz/1e1b (data deps via fixes), h6u0 (town quest hooks), 4ybb.
VERDICT: NEEDS-WORK:ac-stale-vs-decision — "loads from MPQ" contradicts recorded extract-at-build decision; "Town.dun" file does not exist (4 sector .dun files).


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Decision-alignment (triage): "loads from MPQ" → "loads from extracted loose assets (go:embed)" per extract-at-build decision.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
