# ADR-001: Fixed-tick deterministic simulation kernel, engine-free

Date: 2026-11-03

**Status:** Proposed

**Context.** We port functionality, not code. The original netcode is per-tick
deterministic lockstep: all participants advance in lockstep on fixed ticks, wire
carries only commands, no snapshots; per-tick AI seeds keep nodes in sync
(DevilutionX). SP can be modeled as MP with one participant.

**Decision.**
- Fixed-tick (20 Hz, as the original) pure-Go simulation kernel:
  `state(t+1) = reduce(state(t), commands[t])`; kernel imports **no** engine
  packages (zero ebiten/audio/window imports — CI lint rule).
- Determinism contract, CI-enforced: no wall clock in sim (tick counts only);
  no goroutines in sim; seeded RNG streams per concern (item drops, AI per-tick
  seeds, attacks, durability) — mirroring the original's per-tick AI seeds;
  fixed-point math (int32 fixed-point, like the original) — float64 banned in
  the kernel, floats only in presentation.
- Player/AI intent → **commands**; handlers emit **events**; client-predicted
  events are presentation-only, never truth (server-authoritative paths).
- Determinism CI: same seed + command log → state checksums per N ticks,
  golden replays; optionally Ebitengine 2.10 **VM guest** (headless host drives
  ticks + input, captures render/audio frames) for golden end-to-end tests.
- `game/` = Diablo port content (data-driven); kernel + framework dirs stay
  content-free.

**Consequences.** + headless by construction; SP=MP symmetry; replay/desync test
harness; engine swap = presentation swap only. − fixed-point porting discipline;
handler ordering must be explicit (priority + registration order).
