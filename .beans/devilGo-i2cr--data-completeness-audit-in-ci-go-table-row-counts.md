---
# devilGo-i2cr
title: 'Data-completeness audit in CI: Go table row counts vs TSV (base + hf)'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-jx08
    - devilGo-2j8j
    - devilGo-baxb
    - devilGo-778h
    - devilGo-v791
    - devilGo-vhgd
    - devilGo-1e1b
    - devilGo-rctk
    - devilGo-lsz6
    - devilGo-aix0
    - devilGo-h6wn
    - devilGo-b8z9
    - devilGo-pu3r
    - devilGo-fpkz
    - devilGo-orkj
    - devilGo-cciu
---

# Data-completeness audit gate ("every detail" gate)

## Goal
Make Phase 7 falsifiable: a CI job that compares every Go table's row count and first/last-row
fingerprint against the TSVs on disk, in **both** the base and `hf` variants, and fails on any
missing, extra, or drifted table. This is the user's "every detail is accounted for" gate.

## Research
- `/tmp/dg-requirements-append.md` **V12.12** (the transcription checklist: "a `go test` per table
  asserting the count plus first/last-row field equality closes this volume's acceptance
  criteria"), **V12.10** (hf deltas), **V0.1** (loader map = the authoritative table list),
  **V0.2** (override semantics).
- Reference counts (all re-measured on disk with `tail -n +2 <file> | grep -c '[^[:space:]]'`,
  base / hf): itemdat 168/—, item_prefixes 83/86, item_suffixes 95/98, unique_itemdat 90/110,
  monstdat 112/138, unique_monstdat 100/—, misdat 68/108, missile_sprites 46/59, objdat 109/—,
  questdat 24/—, spelldat 36/51 (+injected Null => 37/52), textdat 338/—, textdat-unused 139/—,
  towners 12/15, quest_dialog 10/13, effects 814/917, effects-unused 15/1, Experience 50/—,
  classdat 6/6, and per class attributes 27/27, sounds 103/103, animations 27/27, sprites 13/13,
  starting_loadout 9/9 (30 per-class files). File totals: **49** base + **14** hf `.tsv` files.
- Domain totals: items **436**, monsters **212**, missiles **114**, objects **109**, quests
  **24**, spells **36**, text **477** (338+139), town **22** (12+10), audio **829** (814+15),
  XP **50**, classes **1080** (6 + 1074) — *not* the 1122 the inventory's V12.12 invariant line
  states; dg-task-114 records that correction and this gate asserts 1080.
- Registry (non-TSV) counts to include: AiProc **40** entries with 1 nil (`monster.cpp:3089-3130`),
  misdat Add/Process name sets 51/37 base and 74/46 hf, `ShrineType` 34, `ThemeID` 17,
  `TalkID` 24, `_talker_id` 13 vs towners rows 12/15, `cursor_id` 13, `QoL` files 16,
  music tracks 8 (+NUM_MUSIC), movies 10, `MissileID` 111, `SpellID`, `_unique_items` 91,
  `_speech_id`.
- CI wiring: existing `dg-task-001` (justfile + CI), `dg-task-045` (timedemo + seed matrix job),
  `dg-task-046` (bench), `dg-task-206` (graphics asset manifest/coverage checker — same shape,
  different domain: reuse its pattern, do not duplicate).

## Scope
`internal/gamedata/audit/`: `cmd/dataaudit` (a Go program that (1) walks `assets/txtdata` and
`mods/hf/txtdata`, (2) loads every registered Go table in both variants, (3) emits a
machine-readable report `table, base_rows_tsv, base_rows_go, hf_rows_tsv, hf_rows_go,
first_row_sha, last_row_sha, status`) and `audit_test.go` that fails CI on any `status != OK`.
Registration API so a new table that is not registered is itself a failure (closed world). Also
a `just data-audit` target and a CI step.

## Acceptance criteria
- [ ] TDD first: the audit test is written and **failing** before the tables exist — it must
      enumerate all 49 base + 14 hf files and report every unregistered table as a failure, and it
      is green only when each Go table's row count equals its TSV count **including hf deltas**
      (itemdat 168/168, item_prefixes 83/86, item_suffixes 95/98, unique_itemdat 90/110, monstdat
      112/138, unique_monstdat 100/100, misdat 68/108, missile_sprites 46/59, objdat 109/109,
      questdat 24/24, spelldat 37/52 incl. Null, textdat 338/338, towners 12/15, quest_dialog
      10/13, effects 814/917, Experience 50/50, classdat 6/6, per-class 27/103/27/13/9).
- [ ] Closed-world check: adding a `.tsv` file to `assets/txtdata/` without registering a Go table
      fails the audit (tested by a fixture directory), and dropping one fails it too.
- [ ] Fingerprint parity: for every table, first-row and last-row field hashes equal the TSV in
      both variants — proves alignment and the column-1-skip rule, not just counts (a row-shifted
      table must fail here even with the right count).
- [ ] Registry parity worked example: the audit asserts AiProc == 40 entries / 1 nil, misdat name
      sets == 51/37 base and 74/46 hf, `ShrineType` == 34, `TalkID` == 24, movies == 10, QoL
      modules == 16 — i.e. non-TSV counts are covered by the same gate.
- [ ] Report is diffable in CI: the job prints a per-table diff vs the *expected* manifest
      (dg-task-117's delta manifest) and fails on any change to the expected manifest without a
      recorded decision; run time budget < 60 s so it stays in the PR gate.
- [ ] `just vet test lint` and `just data-audit` are both green with `hf` toggled off **and** on,
      and the PoW records the full report output.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — codegen/registration APIs, closed-world design
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — fixture-driven CI tests + skips
- /home/blaze/.agents/skills/coding-rules/SKILL.md — one report struct, no ad-hoc prints
- /home/blaze/.agents/skills/beans/SKILL.md — PoW must contain the actual audit output

## Dependencies & blockers
blocked_by dg-task-101 and by every Phase-7 table/registry bean: 102, 103, 104, 105, 106, 107,
108, 110, 111, 112, 113, 114, 115, 116, 117. (Practical note: land the audit harness early as a
red baseline and let it turn green incrementally — the gate is only useful if it exists before
the tables do.) Also needs dg-task-001 (CI wiring) and dg-task-117 (expected delta manifest).
Partially overlaps dg-task-045 (seed matrix job) and dg-task-206 (graphics manifest); reuse their
CI shape, do not duplicate.

## Open questions
- Do we want the audit to compare against the C++ *loader* output (needs a probe binary) or only
  against the TSVs? TSV-only is self-sufficient and cheap; the first/last-row fingerprint closes
  most drift. Probe-based comparison belongs to dg-task-045's cross-check sidecar.
- Should `effects-unused`/`textdat-unused` be *shipped* tables or audit-only fixtures? They are
  dead data; the audit still needs to account for their rows (829/477 totals) — confirm the
  audit-only recommendation.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
