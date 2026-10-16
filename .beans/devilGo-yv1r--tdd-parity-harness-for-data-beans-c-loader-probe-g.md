---
# devilGo-yv1r
title: 'TDD parity harness for data beans: C++ loader probe + golden row fixtures'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-jx08
---

# Shared parity harness for Phase-7 data beans

## Goal
Build the one testing utility every Phase-7 bean uses to prove row-level fidelity against the
C++ loaders, instead of 17 beans each hand-rolling their own comparison: a C++ probe that dumps
what the *actual* loaders parsed, and a Go comparator + golden-fixture format.

## Research
- `/tmp/dg-requirements-append.md` **V0.1** (the 21-row table->loader map is the probe's work
  list), **V12.12** (count + first/last-row field equality), **V1.6**/**V1.7** (worked examples
  the harness must be able to express), **V2.3**/**V3.2** (registry-shape assertions).
- C++ entry points to drive from the probe: `DataFile::loadOrDie`/`parseHeader`
  (`Source/data/file.hpp:37-110`), `record_reader.hpp`, and each `Load*` in the V0.1 map
  (`tables/itemdat.cpp:616`, `:674`, `:684`; `tables/monstdat.cpp:409`, `:459`;
  `tables/misdat.cpp:134`, `:373`; `tables/objdat.cpp:259`; `tables/questdat.cpp:80`;
  `tables/spelldat.cpp:232`; `tables/textdat.cpp:106`; `tables/townerdat.cpp:100`, `:135`;
  `effects.cpp:141`; `tables/playerdat.cpp:91`, `:171`, `:209`, `:227`, `:251`, `:289`, `:345`)
  including `AddNullSpell()` (`spelldat.cpp:20-35`) so the injected row is captured, and
  `AiProc[]` (`monster.cpp:3089-3130`) for the registry shape.
- Override path for the hf variant: `options.cpp:71-80` + `assets.cpp:704-760` — the probe must
  run twice (base, hf) with the same dump format.
- Existing patterns to reuse, not reinvent: dg-task-005 (golden-test harness, epic-01),
  dg-task-045 (seed matrix CI), dg-task-203 (graphics backbuffer parity harness — same idea in
  another domain), dg-task-206 (asset coverage checker).
- Build/vet gates from AGENTS.md: `make vet test lint` must stay green; probe build is optional in
  CI (deviation recorded) because the golden fixtures are checked in.

## Scope
`testsupport/dataprobe/`: a C++ sidecar (`dataprobe dump --table <tsv path> [--hf]` emitting
canonical TSV-of-parsed-fields + row counts) built from the existing tree; `testsupport/parity/`:
Go fixture format (one golden file per table per variant), comparator with field-level diff
output (table, row index, column, expected, got), and a helper `parity.AssertTable(t, got,
goldenPath)` used by dg-task-102..116. Golden fixtures live in `testdata/gamedata/`.

## Acceptance criteria
- [ ] TDD first: for `spells/spelldat.tsv`, `parity` asserts the Go table equals the C++ probe
      output field-for-field for **all 37 rows** (36 TSV + injected `Null`) in the base variant and
      **52** in the hf variant — this test is the harness's own acceptance test and is red before
      the probe exists.
- [ ] Row-count + fingerprint API: the helper reports count mismatch vs field mismatch distinctly,
      and a deliberately corrupted golden fixture produces a failure naming table/row/column.
- [ ] Probe coverage: all 21 loaders in the V0.1 table are dumpable (`--table` list asserted),
      including the 30 per-class files and both `towners`/`quest_dialog` variants, and the hf run
      reproduces the V12.10 deltas (monstdat 138, effects 917, quest_dialog 13x21).
- [ ] CI shape: goldens are checked in so `go test ./...` passes without building the C++ probe;
      a `just refresh-gamedata-goldens` target rebuilds them, and the job records which probe
      commit produced the goldens.
- [ ] Every Phase-7 data bean's fidelity AC is expressible with this helper (spot-check by wiring
      dg-task-102 and dg-task-113 to it in the same PR as the helper).
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — fixture/golden test structure
- /home/blaze/.agents/skills/go-dev/SKILL.md — comparator without reflection surprises
- /home/blaze/.agents/skills/coding-rules/SKILL.md — one harness, many consumers
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline

## Dependencies & blockers
blocked_by dg-task-101 (loader core defines the parse contract). Consumed by dg-task-102, 103,
104, 105, 106, 107, 108, 110, 111, 112, 113, 114, 115, 116, 117 and gated by dg-task-118. Needs
dg-task-001 (CI) and reuses dg-task-005. No dependency on the gameplay epics — it can land early.

## Open questions
- Is building the C++ tree in CI acceptable for `refresh-goldens`, or should goldens be generated
  once by hand and reviewed? The repo has a CMake tree; a one-off local generation is cheaper.
- `pack.h`/save-format tables are out of scope here (dg-task-035 owns them) — confirm the probe
  should not also dump entity-state structs.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
