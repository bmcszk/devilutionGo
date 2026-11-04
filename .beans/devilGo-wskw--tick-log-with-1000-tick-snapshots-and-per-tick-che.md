---
# devilGo-wskw
title: Tick log with 1000-tick snapshots and per-tick checksum replay
status: completed
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T18:21:50Z
parent: devilGo-xc37
---

# Tick log with 1000-tick snapshots and per-tick checksum replay

## Goal

ADR-018 law 4 as code: the tick log is the save. Append-only tick log with a
snapshot every 1000 ticks (cadence configurable in settings), where the log
and its snapshots live inside the SQLite store per the record (one column
family of the same store, not a second save format). Load = snapshot +
replay. Per-tick world checksum; the divergence detector names the tick and
prints both hash trails.

## Research

- internal/persist: SaveGameData/ItemPack/PlayerPack codecs, sqlite store,
  demo_compare + memmap_diff (checksum and diff machinery).
- internal/engine/demo: format.go + playback round trips prove the
  append-only-with-records container pattern that the log container mirrors.
- [ADR-018](../docs/adrs/adr-018.md) law 4 (container = SQLite store per
  ADR-003); [ADR-015](../docs/adrs/adr-015.md) replay discipline.

## Scope

- internal/persist: tick-log writer/reader + snapshot rotation inside the
  store; checksum per tick produced by the kernel; divergence detector wiring.
- Tests: 10k-tick replay from the log reproduces the original final state;
  injected divergence reported at the right tick with both trails.

## Acceptance criteria

- [x] Given a 10k-tick log, replay reproduces the original run's final hash.
- [x] Given two diverged runs, the detector names the first differing tick with both checksums.
- [x] Schema v3 migrates v1/v2 stores additively without data loss.

## Libraries

- modernc.org/sqlite already present; nothing new.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- Kernel core; helps if gate/handlers exist so the log carries real events.

## Open questions

- None; container law settled the storage question.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.

Claim evidence 2026-11-03: lane dispatched on branch dg-tick-log, base 4b902f8 (PR #23 merge). Deliverable per contract /home/blaze/github/bmcszk/dg-lane-wskw; parallel lane file-ownership law active.

Claimed 2026-11-03: lane dispatched per contract /home/blaze/github/dg-lane-wskw.contract.md, worktree /home/blaze/github/bmcszk/dg-lane-wskw, base 4b902f8 (PR #23 merge).

## Commits (branch dg-tick-log, base 4b902f8)

- 4762ab2 devilGo-wskw: tick-log column family schema (v3 migration)
- d283faa devilGo-wskw: tick-log store layer (append/load ticks, checksums, snapshots)
- d80f2f4 devilGo-wskw: tick-log writer, replay reader, divergence detector
- dadef76 devilGo-wskw: lint-gate persist package
- HEAD of dg-tick-log devilGo-wskw: fixup round 2 — direct v2‑stamped → v3 upgrade test
  (listed as HEAD, not a hash: the hash of the commit that carries this line cannot
  be written inside it; `git log -1 --format='%h %s'` names it)

## Proof of Work (at head 7a39884 + this fixup commit, 2026-11-03T18:26Z)

Commands run from the lane worktree, tails verbatim:

- `go build ./...` -> rc 0
- `gofmt -l .` -> no output
- `./scripts/lint.sh` -> `0 issues.` rc 0
- `go vet ./internal/...` -> rc 0
- `go test ./... -count=1` -> rc 0, 38 `ok ` package lines, 0 `FAIL` lines
  (5 further lines are `?   ... [no test files]`)
- `go run ./scripts/beans-lint` -> fails: `stat .../scripts/beans-lint: directory
  not found` (the vetted command does not exist in this tree). The real gate is
  `just beans-lint` = `python3 scripts/beans-lint.py .` -> `beans-lint: OK
  (160 beans)` rc 0.

## AC evidence

- AC1: `TestTickLogReplayTenThousandTicks` records 10,000 ticks and asserts the
  replayed world hash equals the recorded run's final hash.
- AC2: `TestTickLogReplayNamesMismatchedTick` names the corrupted tick;
  `TestCompareLogsDivergence` (4 cases) asserts both checksums plus the
  `divergence at tick N: left=... right=...` report text.
- AC3: `schema_v3.sql` only issues `CREATE TABLE IF NOT EXISTS` for the four
  tick_* tables — no ALTER/DROP on v1/v2 data; `TestMigrationReplaysUnstampedV1` walks
  the whole v1 -> v3 chain and keeps its rows.

## Review round 1 findings closed here

- F1: this record (ACs, commit list, PoW).
- F3: `ticklog_store.go` preamble now states the hand-written-SQL exception and
  that drift protection is the append/load roundtrip plus the NOT NULL schema
  test, not sqlc.
- F4/F5 (brief wording "gap-fill NOT NULL", "late-callback replay rejection"):
  `TestTickLogNullGapFillRowRejected` proves a NULL column in
  tick_commands/tick_events/tick_checksums is refused and leaves the tick tables
  empty; `TestTickLogReplayRejectsLateCallback` drives a gate that has closed
  tick 0 and asserts `persist.Replay` returns `ErrTickLogReplay` instead of a
  hash.
  Disclosure: the brief's F5 wording ("after the schema step has begun replay")
  does not map to this tree — replay has no callback or schema-step seam; the
  gate is the only rejection seam, and the test pins that branch.

## Fixup round 2

- F5 (direct v2-stamped -> v3 upgrade test) landed in this commit:
  `TestMigrationUpgradesV2Store` seeds a v2 file (v1/v2 rows, v3 tables dropped
  back out, `user_version = 2` via the `stampVersion` helper), reopens it through
  `OpenSQLiteStore`, and asserts the stamp is 3, the four tick-log tables exist,
  and the seeded hero/level rows survived. RED evidence: the same table check
  run against the pre-upgrade fixture failed with
  `version-3 table tick_commands missing: sql: no rows in result set`.
