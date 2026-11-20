---
# devilGo-9ukg
title: 'M1 slice: retail town as world LevelLayout (CreateTown adapter)'
status: completed
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-20T14:05:57Z
parent: devilGo-97mn
---

# M1 slice: retail town as world LevelLayout (CreateTown adapter)

## Goal
The world's town walks today run on an EMPTY grid (world.NewState,
spawn 0,0; integration session_stage_test.go) while the town parity
corpus + CreateTown live in internal/diablo/levels (town.go:7-16 sector
consts, town_parity_test.go, present/scene/town.go record loading).
Fill the world's LevelSource seam (ADR-008: diablo fills world seams)
with a retail-town LevelLayout adapter: committed town .dun corpus →
grid + SOL + spawn tile, so a walked hero stands on the real Tristram.

## Research
- Town data committed and parity-tested in the diablo layer; the world
  LevelSource seam is the documented consumer-defined boundary.
- The scripted-walk oracle is untouched by this slice (the CI walk keeps
  its empty-grid town; a new test exercises the retail layout).

## Scope
- internal/diablo (the adapter — it OWNS the town data), world seam
  types only if a field is genuinely missing (smallest diff, disclose).

## Acceptance criteria
- [x] A LevelSource adapter serves the retail town layout from the
      committed corpus through the world's existing seam
- [x] A world-level test walks/queries the retail town (spawn, SOL
      blocking, dimensions) with parity asserts vs the corpus
- [x] The scripted-walk oracle's trail UNCHANGED (no CI-walk edits);
      full gates green; RED first on the adapter test

## Triage
Planner wave (wave-s6-plan.md s6c), filed 2026-11-20; parent-verified
against master ce342c7.

## Libraries
stdlib only.

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md

## Progress log

- Claimed 2026-11-20 (s6c coder wave, feat/m1-interactive-demo @ 7a87cdb).

## Proof of Work (s6c, 2026-11-20)

- RED: townlayout_test.go on unchanged tree 7a87cdb — `undefined:
  levels.NewTownLayoutSource`, package build failed. GREEN after adapter.
- Adapter: internal/diablo/levels/townlayout.go (TownLayoutSource); ZERO
  world-layer edits — the LevelSource seam fit as-is (disclosed). SOL record
  is a constructor input so diablo stays render-free (sim purity).
- Test: 8 corpus seeds, spawn == ViewPosition (town.cpp:364-384 branch),
  full dPiece parity cell-for-cell vs committed goldens, spawn walkable,
  off-case rejection.
- Mutation proof: townRockTile+1 → test FAILs (piece[0][96]=427, want 426);
  reverted.
- Trail: single-walk statehash f1ae95a0fd73… byte-identical before/after
  (3 runs at daf0ac4). CI walk town path untouched.
- Gates at daf0ac4: gofmt empty, vet rc0, lint.sh 0 issues (255 sources),
  go test 54 pkgs 0 FAIL, beans-lint OK (222 beans).
- Push: pre-push all gates passed; ls-remote daf0ac44 == HEAD.

## Status

Completed (s6c coder wave).
