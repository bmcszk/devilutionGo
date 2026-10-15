---
# devilGo-xwtt
title: SQLite store + legacy .sv import/export + pfile
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - save
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:27Z
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
