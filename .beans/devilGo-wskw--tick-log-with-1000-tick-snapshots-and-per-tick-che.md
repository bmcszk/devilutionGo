---
# devilGo-wskw
title: Tick log with 1000-tick snapshots and per-tick checksum replay
status: draft
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-03T12:50:42Z
parent: devilGo-xc37
---

# Tick log, snapshot every 1000 ticks, per-tick checksum verification

## Goal

ADR-018 law 4 as code: the tick log is the save. Append-only tick log with a
snapshot every 1000 ticks (cadence configurable in settings, per the record),
load = snapshot + replay, per-tick world checksum, divergence detector that
names the tick and prints both hash trails. Reuses the persist codecs and
diff harness that already exist; no new format invention beyond the log
container.

## Research

- internal/persist: SaveGameData/ItemPack/PlayerPack codecs, sqlite store,
  demo_compare, memmap_diff (checksum + diff machinery).
- internal/engine/demo: file format precedent for append-only logs with
  header + records; writer + playback round trip.
- [ADR-018](../docs/adrs/adr-018.md) law 4; [ADR-015](../docs/adrs/adr-015.md)
  replay discipline; [ADR-003](../docs/adrs/adr-003.md) stays the save-format law.

## Scope

- internal/persist: tick-log writer/reader + snapshot rotation; checksum
  per tick produced by the kernel; divergence detector wiring.
- Tests: 10k-tick replay from log matches original final state; injected
  divergence reported at the right tick with both trails.

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

- Log file location follows the assets/settings dirs config; exact default
  path lives in the open questions of the parent until settings land.

## Triage

TRIAGE brainer evidence 2026-11-03: kernel/migration laws come straight from the
accepted architecture records ADR-013..018 (deep research confrontation:
Game Programming Patterns event-queue chapter, Bevy observer-ordering doc,
Watermill GoChannel reference, AoE/Factorio/OpenTTD lockstep and desync
write-ups, devilutionX demo format parity). Code seams verified in the tree
surveys; workers get exact files in Research.
