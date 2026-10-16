---
# devilGo-xwtt
title: SQLite store + legacy .sv import/export + pfile
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - save
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:16:44Z
parent: devilGo-6xu2
blocked_by:
    - devilGo-oqho
---

# Stores

## Goal
Two Store impls + hero lifecycle.

## Research
pfile.cpp:60-97,177,628; pfile.h:100-128; stash :2740-2760; schema proposal /tmp/dg-requirements.md §5.

## Scope
internal/persist/{sqlite,legacysv,pfile}.go; modernc.org/sqlite.

## Acceptance criteria
- [ ] Round-trip sqlite→StateHash equal; migration v0→v1 test
- [ ] .sv container export byte-layout matches C++ writer (hex compare on one save)
- [ ] Hero slots 0..98; UI list API returns _uiheroinfo equivalents


## Libraries
- modernc.org/sqlite — pure-Go SQLite driver (per-save .db + global stash DB); no CGO (pure-Go hard rule) pin v1.x

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — database/sql idioms, migrations
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — round-trip tests

## Dependencies & blockers
dg-task-035, dg-epic-02 (mpq writer for container).

## Open questions
- go-mpq write-side zlib-only: .sv needs STORE sectors — verify writer supports; else hand-roll container (fallback documented).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Save architecture amendment (user decision 2026-10-16)
- Per-hero SQLite DB: every save = a separate SQLite database file, mirroring C++ per-hero .sv model (e.g. `single_0.sv` → `single_0.db`).
- Additional global SQLite DB for cross-hero state: shared stash, metagame, global settings.
- Importer CLI writes into this model: legacy .sv/.sp → per-hero .db files; stash → global.db.


## Settings format amendment (user decision 2026-10-16)
- Settings = 1:1 `diablo.ini` file (INI codec port of Source/utils/ini.cpp), NOT in SQLite. Global DB = stash/metagame only. Key/value compat: diablo.ini interchangeable between C++ and Go builds.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Persistence epic store task; blocked_by oqho (matches tree dg-task-036←035). Both 2026-10-16 user amendments (per-hero SQLite DB + global stash DB; settings = diablo.ini not SQLite) are recorded IN this bean's body — decision provenance good.
2 local: pfile.cpp:60-64 passwords, :177 codec_encode hero write, :628 codec_decode read all verified; pfile.h:100-128 API verified; "stash :2740-2760" — loadsave.cpp:2740-2760 verified = stash pages/grid/itemlist writes (SaveStash ~:2706); nit: citation omits filename and could be misread as stash.cpp (which is only 737 lines) — ambiguity only, content correct; schema proposal verified /tmp/dg-requirements.md §5 (lines 303-361, Store interface + two impls + roundtrip test); modernc.org/sqlite in go.mod? Not yet — repo is scaffold-stage (go.mod has no deps); fine, deps land at implementation; epic 6xu2 blocked_by 97mn + structured chain w67z→oqho→xwtt is coherent.
3 online: n/a (go-mpq write-side capability question is flagged as open WITH documented fallback — hand-roll container; /tmp/dg-requirements.md documents zlib-only write + STORE-sector need).
4 solution: minor: label ".sv container export" AC as parity-harness-only (milestone DECIDED #1: SQLite is the only live save format; no dual-format live writes — export is the byte-compat proof, not a user-facing format). Resolve open question (go-mpq STORE sectors vs hand-rolled container) at promotion — fallback already documented. Note: milestone body's SAVE ARCHITECTURE amendment still lists "global settings" in global.db while the later SETTINGS amendment (and this bean) correctly excludes settings to diablo.ini — inconsistency lives in devilGo-i56j (batch-01 scope), not here.
5 understanding: persistence layer top — stores consume oqho serializers; per-hero .db mirrors C++ per-hero .sv model exactly per decision.
6 repro: `beans show devilGo-xwtt`; `sed -n '177,179p' .../Source/pfile.cpp`; `sed -n '2740,2744p' .../Source/loadsave.cpp`; `sed -n '303,361p' /tmp/dg-requirements.md`; `beans show devilGo-i56j --body-only | grep -A5 "SETTINGS FORMAT"`.
7 fix: optional — `beans update devilGo-xwtt --body-append "Note: .sv export AC = parity-proof harness only (DECIDED #1: SQLite is sole live save format)."`
8 validation: `beans show devilGo-xwtt` — amendments present (already verified).
9 related: oqho, w67z, dcob (hero list API), 6xu2, i56j.
VERDICT: TRIAGED-OK

---

## Verdict roll-up

devilGo-dcob: NEEDS-WORK:AC omits Bard/Barbarian class entries (selhero.cpp:437-463) + no class-stats row-exact parity criterion; stale dg-task-028 prose ref.
devilGo-glp0: NEEDS-WORK:stores-core dependency (devilGo-0qnd) not structurally wired though AC requires end-to-end store flows; stale dg-task refs.
devilGo-h6wn: TRIAGED-OK
devilGo-emdz: NEEDS-WORK:AC lacks concrete parity gate (labels must come from h6wn spec; attract timeout constant unspecified) — "Flow navigable" is not a TDD-first criterion; stale dg-task-027 ref.
devilGo-vf8q: TRIAGED-OK
devilGo-sf6k: TRIAGED-OK
devilGo-t4fz: NEEDS-WORK:JSON store conflicts with recorded diablo.ini-1:1 settings decision; INI codec ownership overlaps devilGo-0tsf (needs canonical home).
devilGo-c6wt: NEEDS-WORK:end-to-end AC (20-step session parity) not achievable at this bean's dependency level — gameplay/render deps not wired; AC needs unit-vs-e2e split.
devilGo-w67z: TRIAGED-OK
devilGo-oqho: TRIAGED-OK
devilGo-xwtt: TRIAGED-OK

Counts: TRIAGED-OK 6 · NEEDS-WORK 5 · DUPLICATE 0 · STALE 0 · DECISION-NEEDED 0

BATCH-08-COMPLETE


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
