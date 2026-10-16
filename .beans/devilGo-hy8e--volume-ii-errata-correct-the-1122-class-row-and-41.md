---
# devilGo-hy8e
title: 'Volume II errata: correct the 1122 class-row and 41-entry AiProc claims'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-778h
    - devilGo-pu3r
---

# Volume II errata pass

## Goal
Record the two arithmetic errors found while verifying `/tmp/dg-requirements-append.md` against
the repository, so no downstream bean re-litigates them and the Phase-7 ACs cite the corrected
numbers.

## Research
- `/tmp/dg-requirements-append.md` **V12.12** ("uniqueness invariant" line) and **V2.3**
  (AiProc table) — both re-checked against disk.
- Erratum 1 — class rows: V12.12 states `6 + 6*(27+27+103+13+9) = 1122`. The sum inside the
  parentheses is 179, and `179 * 6 = 1074`, so the correct total is `6 + 1074 = **1080**`.
  Verified by counting each of the 30 per-class files (`tail -n +2 f | grep -c '[^[:space:]]'`)
  under `assets/txtdata/classes/` (6 classes x attributes 27, sounds 103, animations 27,
  sprites 13, starting_loadout 9).
- Erratum 2 — AiProc size: V2.3's table has 41 rows and describes `AiProc[40]` as the
  `Invalid = -1` sentinel. `Source/monster.cpp:3089-3130` defines **40** entries (indices 0-39),
  `MonsterAIID::Invalid = -1` is not an array slot, and exactly one entry is `nullptr`
  (`FireMan`). `MonsterAIID` likewise has 40 named values + `Invalid = -1`.
- Checked and **correct** (no errata): 436 item rows, 212 monster rows, 114 missile rows, 109
  object rows, 24 quest rows, 36 spell rows, 338+139 text rows, 12+10 town rows, 814+15 audio
  rows, 50 XP rows, all 14 hf deltas, 34 shrine types, 17 themes, 13 `cursor_id`, 24 `TalkID`,
  and every one of the 63 TSV row counts.
- Related derived numbers worth recording: misdat references 51 distinct `addFn`/37 `processFn`
  in base and 74/46 in hf (V3.2's "74 Add functions" is the hf superset, not the base count).

## Scope
Append an `## Errata` block to the Volume II inventory notes (the file lives in /tmp for now; the
portable copy belongs in the Go repo's docs once epic-01 scaffolds it), and cross-link the
corrections from dg-task-114 (class rows) and dg-task-104 (AiProc) so their ACs are authoritative.

## Acceptance criteria
- [ ] TDD first: `internal/gamedata/audit/errata_test.go` asserts class-domain total == **1080**
      (not 1122) and `len(AiProc)` == **40** with exactly 1 nil — red against any code that
      assumes the inventory's original numbers.
- [ ] The errata text records both corrections with the exact command used to measure them
      (`find`/`tail -n +2 | grep -c`, `grep -c MonsterAIID` over `monster.cpp:3090-3130`).
- [ ] Cross-check pass recorded: the list of claims verified *correct* is kept so the next audit
      does not re-measure everything.
- [ ] Every Phase-7 bean AC that cites a corrected number (104, 114, 118) matches this errata.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/beans/SKILL.md — errata as a tracked bean, not a silent edit
- /home/blaze/.agents/skills/go-dev/SKILL.md — assertion tests for doc-derived constants
- /home/blaze/.agents/skills/coding-rules/SKILL.md — numbers live in exactly one place

## Dependencies & blockers
blocked_by dg-task-104, dg-task-114 (they produce the assertions). Feeds dg-task-118's expected
manifest and closes the Volume II verification loop.

## Open questions
- Where should Volume II live once the repo is scaffolded — `docs/`, `testdata/`, or a
  `gamedata/README`? Only `## Errata` needs to be durable; the rest is /tmp scratch today.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
