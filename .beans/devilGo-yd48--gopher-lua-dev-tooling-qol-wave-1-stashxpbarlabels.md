---
# devilGo-yd48
title: gopher-lua dev tooling + QoL wave 1 (stash/xpbar/labels)
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - qol
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-16T14:26:54Z
parent: devilGo-yjuv
blocked_by:
    - devilGo-0tsf
---

# Lua + QoL

## Goal
REPL subset; three QoL modules behind flags.

## Research
lua modules list (§11); qol/stash.cpp, xpbar.cpp, itemlabels.cpp, floatingnumbers.cpp.

## Scope (fix2 split)
- Core QoL (this bean, in-scope): non-Lua quality features
- Lua-hook QoL: DEFERRED post-1.0 with gopher-lua (recorded decision) — tracked in milestone, not here
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


## IMPL NOTES (triage 2026-10-16)
1 root: two goals bundled — (a) gopher-lua REPL subset + audio/dev bindings, (b) QoL ports
   (stash, xpbar, itemlabels, floatingnumbers) behind flags. Parent devilGo-yjuv, blocked_by
   devilGo-0tsf (options — QoL needs option keys; sound ordering).
2 local: REPL claims verified against source: repl_prelude.lua loads `player`, `audio`,
   `dev` (debug-only) modules ✓; lua/modules/audio.cpp:42 `playSfx` + RegisterSfxIDEnum ✓;
   dev/player/stats.cpp exists (hp readable) ✓. Stash 100 pages ✓ (loadsave.cpp:2740-2760
   "Current stash size is 100 pages"). xpbar segment math from Experience.tsv ✓
   (xpbar.cpp:101 thresholds; assets/txtdata/Experience.tsv exists). **Conflict:** milestone
   decision log (devilGo-i56j process log, verified in bean body): "LUA: Deferred post-1.0.
   Pure-Go constraint → gopher-lua when implemented (0 VM stubs)". This phase-3 bean bundles
   Lua NOW. QoL modules themselves are C++ Source/qol/* ports (all 4 files exist) needing no
   Lua — the Lua coupling is artificial.
3 online: n/a.
4 solution: split — QoL wave 1 (stash/xpbar/itemlabels/floatingnumbers) proceeds phase-3
   with no Lua dependency; REPL/gopher-lua tooling deferred to post-1.0 per recorded
   decision (rewrite REPL AC into the deferred scope or a separate post-1.0 bean). Keep
   blocked_by 0tsf for the QoL half.
5 understanding: platform layer QoL riding on engine options (0tsf) + persistence epic 6xu2
   (stash store — matches per-hero DB + global stash DB decision ✓); Lua was always a
   post-1.0 engine extension per decision log.
6 repro: `ls Source/qol/{stash,xpbar,itemlabels,floatingnumbers}.cpp`; `grep -rn playSfx
   Source/lua/modules/audio.cpp | head -2`; `grep -n "Deferred post-1.0" .beans/devilGo-i56j*.md`.
7 fix: rewrite bean title/scope to "QoL wave 1 (no Lua)"; move REPL AC + internal/platform/lua
   scope into a new post-1.0 task bean or the epic's backlog note.
8 validation: bean body contains no phase-3 Lua deliverable; QoL ACs (stash pages, xpbar
   segments) unchanged.
9 related: devilGo-0tsf, devilGo-6xu2 (epic-08 stash store), devilGo-vinx (M4 where Lua
   could resurface), devilGo-orkj (QoL registry data).
VERDICT: NEEDS-WORK:scope-conflict-with-recorded-Lua-deferral (split QoL from gopher-lua)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Decision-alignment (triage): split — QoL Lua hooks (deferred post-1.0) separated from core QoL features (in-scope).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via /tmp/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.