# ADR-007: Game-agnostic framework, Diablo as first content

Date: 2026-11-03

**Status:** Proposed

**Decision.**
- Layout: `framework/` (kernel, ECS glue, buses, transports, persistence,
  migration) — content-free; `game/` — Diablo content as data + systems;
  `tools/migrate` — the migration binary.
- Framework never imports `game/`; `game/` never imports ebiten directly — it
  registers into framework interfaces (Renderer/Mixer/input mapping).
- Registry format (YAML/CSV per ADR-005) doubles as the "other games" hook:
  a different ARPG = different data + different `game/` content, same kernel.

**Consequences.** + two goals (Diablo 1 port, ARPG framework) share one
architecture instead of diverging; − framework APIs must resist Diablo-specific
shortcuts (review law).

## Explicit non-decisions (parking lot)
- Networking protocol details (turn protocol, prediction, rollback) — separate
  ADR when MP slices start; ADR-003 only fixes the seam (transport interface).
- Renderer/audio engine alternatives — ADR-004 fixes the seam only.
- Bean rework — after your acceptance, per your instruction.
