---
# devilGo-t4fz
title: Keymapper + padmapper + config persistence
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:38:07Z
parent: devilGo-k736
---

# Binds

## Goal
Rebindable keys/pads with C++-compatible defaults.

## Research
keymapper.hpp/padmapper.hpp; diablo.ini keys (utils/ini.cpp; test ini_test.cpp).

## Scope
internal/input/{keymapper,padmapper}.go + diablo.ini INI (fix2: JSON store dropped per decision) store + ini importer for migrating users.

## Acceptance criteria
- [ ] Defaults table equals C++ Keymapper defaults (fixture)
- [ ] Import: existing diablo.ini binds load correctly (test with sample ini)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — config structs + validation

## Dependencies & blockers
None.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:JSON store conflicts with recorded diablo.ini-1:1 settings decision; INI codec ownership overlaps devilGo-0tsf (needs canonical home).


## IMPL NOTES (triage 2026-10-16)
1 root: Input epic task — rebindable keys/pads; no blockers (matches tree dg-task-033).
2 local: controls/keymapper.hpp, controls/padmapper.hpp, utils/ini.cpp, test/ini_test.cpp all exist; C++ persists binds via diablo.ini (options.cpp:1203 LoadFromIni / :1229 SaveToIni for KeymapperOptions::Action).
3 online: n/a.
4 solution: scope conflict with recorded decision: "JSON store + ini importer" vs SETTINGS FORMAT AMENDMENT (2026-10-16, milestone i56j + xwtt body): settings are 1:1 diablo.ini, NOT SQLite/JSON, INI codec = port of utils/ini.cpp. JSON store for binds must be dropped or re-decided; binds belong in diablo.ini sections like C++. Also overlaps devilGo-0tsf (Options + ini import, owns ini_import.go + ini_test port): the shared INI codec needs one canonical home — recommend 0tsf owns codec, t4fz consumes it.
5 understanding: input layer config; must ride the platform-layer INI codec per decision.
6 repro: `beans show devilGo-t4fz`; `grep -n "SaveToIni\|LoadFromIni" .../Source/options.cpp`; `beans show devilGo-i56j --body-only | grep -A3 "SETTINGS FORMAT"`; `beans show devilGo-0tsf`.
7 fix: `beans update devilGo-t4fz --body-replace-old "internal/input/{keymapper,padmapper}.go + JSON store + ini importer for migrating users." --body-replace-new "internal/input/{keymapper,padmapper}.go persisting to diablo.ini via the shared INI codec (devilGo-0tsf owns codec; 1:1 with C++ KeymapperOptions::Action LoadFromIni/SaveToIni). No JSON store (superseded by SETTINGS FORMAT AMENDMENT 2026-10-16)."`
8 validation: `beans show devilGo-t4fz` shows ini-based scope; AC unchanged (defaults fixture + ini import test both still valid).
9 related: 0tsf, sf6k, c6wt, k736, i56j.
VERDICT: NEEDS-WORK:JSON store conflicts with recorded diablo.ini-1:1 settings decision; INI codec ownership overlaps devilGo-0tsf (needs canonical home).


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Decision-alignment (triage): JSON store replaced by diablo.ini 1:1 INI codec per recorded decision.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
