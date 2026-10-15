---
# devilGo-sr71
title: i18n catalogs + string registry
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - i18n
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-yjuv
---

# Translations

## Goal
.po→.mo pipeline + runtime lookup.

## Research
Translations/*.po (25); language.cpp:39-42; gotext-style loading.

## Scope
internal/platform/i18n with go:generate msgfmt step (or vendored .mo).

## Acceptance criteria
- [ ] en+pl 100-key parity test
- [ ] Missing key → English fallback (never empty)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — codegen for catalogs

## Dependencies & blockers
None.

## Open questions
- bidi (sheen_bidi) deferred post-M2 (risk register).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
