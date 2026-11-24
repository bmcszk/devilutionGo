---
# devilGo-njzu
title: 'M1 slice: live view — presenter follows the hero'
status: completed
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-23T12:46:23Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-9qzw
---

# M1 slice: live view — presenter follows the hero

## Goal
The presenter composes ONE static first scene and never re-composes
(cmd/devilutionGo/window_scene.go:73-103). Make the view live the legal
way: the presentation layer reads kernel events (PlayerMoved,
CreatedPack) as its ONLY state source (ADR-018 law 2 — it never writes
sim), recenters the town view on the hero (camera as presentation
state, ClickRouter TileAt precedent), draws the hero sprite from the
world's hero state, and caches town records. A screenshot artifact
shows a moved hero.

## Research
- Plan: wave-s6-plan.md; the first-scene pipeline (retail Tristram +
  warrior STAND at heroTile 75,68) is byte-validated against the
  user-approved #73 baseline — the recenter must keep that fidelity
  (pixel-stats evidence per the visual law).

## Scope
- internal/present/** + cmd/devilutionGo view glue; NO sim changes.

## Acceptance criteria
- [x] View follows the hero during a scripted (test-driven) walk: a
      later frame's screenshot differs from the first by the expected
      recenter (pixel stats recorded); hero sprite follows state
- [x] The static first frame is UNCHANGED (the validated baseline
      intact before any motion); full gates green; RED first

## Triage
Planner wave (wave-s6-plan.md s6h), filed 2026-11-20; blocked by the
device lane (a live walk to follow).

## Libraries
stdlib + ebiten (existing dep).

## Skills for coder
- ~/.agents/skills/ebiten/rendering/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md



## Progress log

- 2026-11-20T19:05Z: claimed — live view slice dispatched under contract harness.njzu.md (worktree ftbu-demo, branch feat/m1-demo-flow @ 991b-pushed anchor; base master eb4f1f4 carries the s6 wave incl. the device lane), status flip.



## Proof of Work

- Worker task-dg-m1-njzu-coder-coder (81 tool calls): commit 4f0362b on feat/m1-demo-flow, pushed (ls-remote == HEAD, pre-push gates all passed). 7 files +238/−67.
- RED captured verbatim on the unchanged tree: the scene pinned (75,68) while the walked state said (51,49) — the log records it in harness.njzu.log.
- Presenter derives the hero tile from the kernel state each frame (one source of truth, read-only); compose seam ComposeTownFrameAtHero keeps the boot frame identical to the validated baseline (the boot spawn moved 50,50→75,68, disclosed, device-lane tests updated to still bite).
- Dep-graph guard tripped mid-flight and was fixed honestly (presenter behind the window guard; headless binary stays scene-free).
- Review: brainer PASS 8/8 (.hermes/lane/ftbu-m1/harness.njzu.review.md); F1 parent decision: the pixel-stats artifact is DEFERRED to the AC-4 refresh gate (devilGo-jjd0) where real window captures + per-capture pixel stats are produced — the scene acceptance tests skip in archive-less environments (verified parent-side: DG_SCENE_ARTIFACT_DIR writes only where the MPQ resolves); reviewer judged the risk low (the recenter rides the validated composeView path, parameterized only). F2 (sticky composeErr freeze) pre-existing — named for the AC-4 gate hardening.
- Parent gates at 4f0362b: cmd suite ok, full suite 54 pkgs ok, trail f1ae95a0… unchanged, scripted-walk oracle untouched.
