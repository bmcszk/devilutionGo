---
# devilGo-ru6l
title: Isometric world renderer (floor/walls/order)
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - render
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:16:45Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
---

# World draw

## Goal
Faithful scrollrt port for static world + entity order.

## Research
scrollrt.cpp:949-1036 (DrawFloor/DrawTileContent/wall pre-pass), :790-980 (DrawDungeon order), :1229 DrawGame; displacement worldToScreen Source/engine/displacement.hpp:151-164.

## Scope
internal/render/world.go with tile iterators mirroring row walk; DrawDungeon implementing the exact order incl. negative dPlayer for moving players; dirt-tile OOB fallback (scrollrt.cpp DrawDirtTile).

## Acceptance criteria
- [ ] Order-instrumented test equals expected sequence for a scripted 3×3 scene
- [ ] Golden hash at progress 0/64/128 for one seed
- [ ] Zero Draw allocs

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — batching identical-options draws
/home/blaze/.agents/skills/go-dev/SKILL.md — index math without bounds pain

## Dependencies & blockers
dg-task-007 (sprites).

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from /tmp/dg-graphics.md G10/G11 — binding)
- [ ] Draw order matches `scrollrt.cpp:781 DrawDungeon` 12-layer order exactly; entity y-sort verified by parity frame.
- [ ] Render loop lives in Draw() only; no state mutation; per-frame cost ≤7k triangles (G10 budget).


## Resolution gate (grilling decision #5)
- [ ] 640×480 world logical surface rendered correctly, AND 2x window (1280×960) integer-scaled without artifacts — parity golden frames at BOTH sizes.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:stale-body-deps (dg-task-007) + research paths missing `engine/render/` prefix; logic/AC otherwise verified sound (order test, golden hash 0/64/128, zero-alloc — all testable; TDD gates present).


## IMPL NOTES (triage 2026-10-16)
1 root: World-draw bean for qbiy epic — faithful scrollrt port (tile walk + DrawDungeon entity order). Structurally sound; body text drifted from frontmatter.
2 local: `beans show` + probes: DrawFloor/DrawTileContent confirmed at `Source/engine/render/scrollrt.cpp:955/:992` (bean's 949-1036 ✓), DrawDirtTile `:1047` ✓, DrawGame `:1229` ✓, `worldToScreen` at `Source/engine/displacement.hpp:151-164` ✓ exact; negative-dPlayer priority logic confirmed at scrollrt.cpp:826-848 (walk-southwards/east → playerId = -playerId) ✓. Bean cites `scrollrt.cpp` without `engine/render/` prefix; gate cites "scrollrt.cpp:781 DrawDungeon" — actual `:782` (trivial off-by-one).
3 online: n/a (no external claims).
4 solution: fix research paths to `Source/engine/render/scrollrt.cpp`; replace body dep `dg-task-007` with frontmatter value `devilGo-n89r`; DrawDungeon cite 781→782.
5 understanding: Diablo-layer rendering driven by renderer-epic architecture (66o7 index framebuffer); consumes n89r sprite decode; consumed by j9be (lighting at draw time). No overlap with gk5e (gk5e explicitly excludes entity order — its body says "NOT in scope: entity order (dg-task-010)").
6 repro: `beans show devilGo-ru6l --body-only`; `sed -n '949,1036p' /home/blaze/github/bmcszk/devilutionX/Source/engine/render/scrollrt.cpp`; `sed -n '826,848p' .../scrollrt.cpp`; `sed -n '151,164p' .../engine/displacement.hpp`; `grep -A2 '^blocked_by' .beans/devilGo-ru6l--*.md`.
7 fix: `beans update devilGo-ru6l --body-replace-old "scrollrt.cpp:949-1036" --body-replace-new "Source/engine/render/scrollrt.cpp:949-1036"`; `beans update devilGo-ru6l --body-replace-old "dg-task-007 (sprites)." --body-replace-new "devilGo-n89r (CLX/CEL decoders)."`; `beans update devilGo-ru6l --body-replace-old "scrollrt.cpp:781 DrawDungeon" --body-replace-new "scrollrt.cpp:782 DrawDungeon"`.
8 validation: `beans show devilGo-ru6l --body-only | grep -c dg-task` → 0; grep for `engine/render/scrollrt` → ≥1 hit.
9 related: n89r (dep), j9be (blocked by this), gk5e (tile pipeline, complementary), pdpc (lightmap corner-blit, complementary).
VERDICT: NEEDS-WORK:stale-body-deps (dg-task-007) + research paths missing `engine/render/` prefix; logic/AC otherwise verified sound (order test, golden hash 0/64/128, zero-alloc — all testable; TDD gates present).


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
