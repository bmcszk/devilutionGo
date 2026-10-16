---
# devilGo-dcob
title: Hero select/create/delete + class stats
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - ui
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:38:06Z
parent: devilGo-xih1
blocked_by:
    - devilGo-emdz
    - devilGo-pu3r
---

# Hero screens

## Goal
selhero over persistence (list/create/delete), class stat preview.

## Research
DiabloUI/hero/selhero.cpp; _uiheroinfo diabloui.h:60-72; callbacks :81,93; pfile_ui_* API pfile.h:120-125; MAX_CHARACTERS 99 pfile.h:24.

## Scope
internal/ui/selhero.go wired to internal/persist hero store interface (stub until epic-08).

## Acceptance criteria (fix2 additions)
- [ ] Bard + Barbarian class entries included (selhero.cpp:437-463 Hellfire classes)
- [ ] class-stats table parity vs C++ (all classes, row-exact)
- [ ] Create warrior/rogue/sorcerer(+monk HF) with defaults from playerdat
- [ ] Delete confirm dialog; 99-slot cap enforced

## Skills for coder
/home/blaze/.agents/skills/ebiten/input/SKILL.md — text input for name (16 chars, _uiheroinfo.name)

## Dependencies & blockers
dg-task-028.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC omits Bard/Barbarian class entries (selhero.cpp:437-463) + no class-stats row-exact parity criterion; stale dg-task-028 prose ref.


## IMPL NOTES (triage 2026-10-16)
1 root: UI epic (xih1) task — selhero screens over persistence; structurally parented + blocked_by emdz (front screens), mirroring tree dg-task-029←028.
2 local: `beans show` OK; `_uiheroinfo` struct verified DiabloUI/diabloui.h:60-72 (name[16] etc.); callbacks :81/:93 (gfnHeroInfo, UiSelHeroSingDialog) verified; `MAX_CHARACTERS 99` pfile.h:24 verified; `pfile_ui_*` cluster pfile.h:120-128 verified; hero/selhero.cpp exists (22,830 bytes); `playerdat` = tables/playerdat.cpp exists.
3 online: n/a (no external claims).
4 solution: AC incomplete vs totality directive: selhero.cpp:437-463 also defines Bard (asset/test-gated) and Barbarian (test-gated) class entries — AC "warrior/rogue/sorcerer(+monk HF)" silently drops them; record decision or add them. No domain-specific parity AC (e.g. name-length validation = UiValidPlayerName, class-stat defaults row-exact vs playerdat). Prose dep "dg-task-028" stale → emdz.
5 understanding: UI layer (internal/ui/selhero.go) consuming persistence store interface — correct seam; stub-until-epic-6xu2 noted in scope.
6 repro: `beans show devilGo-dcob`; `sed -n '60,72p' ~/github/bmcszk/devilutionX/Source/DiabloUI/diabloui.h`; `sed -n '437,463p' .../DiabloUI/hero/selhero.cpp`; `grep -n MAX_CHARACTERS .../pfile.h`.
7 fix: `beans update devilGo-dcob --body-replace-old "dg-task-028." --body-replace-new "devilGo-emdz (structured blocked_by already set)."` + append AC line for Bard/Barbarian handling + UiValidPlayerName parity.
8 validation: `beans show devilGo-dcob` shows updated dep + new AC item; re-grep selhero.cpp:183-187 gates.
9 related: emdz, vf8q, h6wn (menu spec), xwtt (pfile API), 0tsf.
VERDICT: NEEDS-WORK:AC omits Bard/Barbarian class entries (selhero.cpp:437-463) + no class-stats row-exact parity criterion; stale dg-task-028 prose ref.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
