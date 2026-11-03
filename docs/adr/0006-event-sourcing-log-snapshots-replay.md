# ADR-006: Event sourcing: command log, snapshots, replay

Date: 2026-11-03

**Status:** Proposed

**Decision.**
- Event log is not just persistence: it is the netcode (commands on the wire),
  the save format (commands + periodic snapshots in SQLite), the replay format
  (seed + commands → deterministic replay), and the desync detector (state
  checksums). One mechanism, four uses — this is why determinism (ADR-001) is
  a hard requirement, not a nice-to-have.
- Snapshots: per-level, periodic (every N ticks) + on demand (save).
- Replay = re-apply command log after snapshot; golden replays in CI.

**Consequences.** + saves become portable/replayable/debuggable; MP desyncs
become diagnosable from logs; − strict discipline (all mutation through
commands, no hidden state).
