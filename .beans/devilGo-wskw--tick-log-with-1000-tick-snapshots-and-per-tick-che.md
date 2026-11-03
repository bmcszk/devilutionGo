---
# devilGo-wskw
title: Tick log with 1000-tick snapshots and per-tick checksum replay
status: in-progress
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T15:56:11Z
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

- [ ] Given a 10k-tick log, replay reproduces the original final state hash.
- [ ] Given two diverged runs, the detector names the first diverging tick plus both checksums.
- [ ] Whole go test ./... green, gofmt clean.

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
