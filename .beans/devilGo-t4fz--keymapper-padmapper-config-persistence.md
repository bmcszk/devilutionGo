---
# devilGo-t4fz
title: Keymapper + padmapper + config persistence
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - input
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-k736
---

# Binds

## Goal
Rebindable keys/pads with C++-compatible defaults.

## Research
keymapper.hpp/padmapper.hpp; diablo.ini keys (utils/ini.cpp; test ini_test.cpp).

## Scope
internal/input/{keymapper,padmapper}.go + JSON store + ini importer for migrating users.

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
