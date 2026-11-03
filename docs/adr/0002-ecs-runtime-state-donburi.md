# ADR-002: ECS runtime state via Donburi

Date: 2026-11-03

**Status:** Proposed

**Decision.**
- Player/monster/item/missile = entities; HP/stats/position/AI-state = components;
  each original game-loop step → one system in an **explicit ordered pipeline**
  mirroring original step order (reviewable, 1:1 port mapping).
- Mutating systems consume commands; presentation systems are read-only views
  (no write-back into sim state).
- Use `donburi.NewWorld()` **directly** — NOT the `ecs` convenience wrapper
  (that one imports ebiten; would break kernel purity). Donburi's core World has
  no ebiten dependency; necs (donburi networking plugin) exists when MP needs it.

**Consequences.** + natural fit to Ebiten ecosystem, systems map to original
steps; − per-system porting must resist OO-style shortcuts (law: no god structs).
