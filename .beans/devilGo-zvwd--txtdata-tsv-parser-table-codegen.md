---
# devilGo-zvwd
title: txtdata TSV parser + table codegen
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - data
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-15T16:45:27Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-vof6
---

# Data tables

## Goal
Parse all txtdata TSVs into typed Go tables mirroring C++ structs.

## Research
Format: assets/txtdata/Readme.md EBNF (BOM optional, tab-delimited, no tabs in values); parser Source/data/file.hpp:37-110 + record/value_reader; table structs: SpellData tables/spelldat.h:225-252, MissileData tables/misdat.h:166-196, PLStruct/UniqueItem tables/itemdat.h:628-655, MonsterData tables/monstdat.cpp:52-55, Quest tables/questdat.hpp:102-129, townerdata.

## Scope
internal/assets/txtdata: parser (BOM/CR/LF tolerant, header-mapped, typed readers with error reporting parity to reportFatalFieldData), go:generate tool producing structs from a schema file (or explicit structs + tests).

## Acceptance criteria
- [ ] test/fixtures/txtdata/*.tsv edge cases all pass
- [ ] Row counts equal C++ (monstdat 113, experience 50, etc.)
- [ ] Unknown column → error naming file+field (parity)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — codegen with text/template
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — table-driven parser tests

## Dependencies & blockers
dg-task-006.

## Open questions
- Codegen vs hand structs: spike both, pick less drift (open question H3).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
