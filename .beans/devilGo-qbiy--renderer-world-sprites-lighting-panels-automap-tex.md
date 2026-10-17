---
# devilGo-qbiy
title: 'Renderer: world, sprites, lighting, panels, automap, text'
status: todo
type: epic
priority: normal
tags:
    - plan:epic
    - phase-1
created_at: 2026-10-15T16:37:32Z
updated_at: 2026-10-17T19:56:52Z
parent: devilGo-i56j
blocked_by:
    - devilGo-5wsb
---

# Rendering engine

## Goal
Draw the world exactly like scrollrt: isometric floor/wall pass, per-tile content order, interpolated entities, light tables, transparency, automap, panels, bitmap text, cursor — all allocation-free in Draw.

## Research
- DrawGame/DrawFloor/DrawTileContent: Source/engine/render/scrollrt.cpp:1229, :949-988 (row walk: y+=TILE_HEIGHT/2, odd rows x++/cols--), wall pre-pass :1000-1036.
- Per-tile order DrawDungeon :790-980 (missiles-pre, corpses, pre-objects, items, dead players, players with negative-id trick :830-866, monsters, post items/objects, missiles, dSpecial blended :907-952).
- Missile render interpolation :125-193; light at draw ClxDrawLight/TRN :249-258; per-pixel lightmap option :912-930.
- Blitters: Source/engine/render/dun_render.cpp:908-994; clx_render.cpp:592; text_render.cpp:291-359.
- Automap: Source/automap.cpp:980,1341,1427,1752-1874; automap.h:83,91; engine/render/automap_render.cpp:168-171.
- Panels: Source/control/control_panel.cpp:351,435-445; control_flasks; panels/*. Tests: panel_state, char_panel, spell_ui, inventory_ui, text_render_integration.
- Ebiten batching rules: https://ebitengine.org/en/documents/performancetips.html + /home/blaze/.agents/skills/ebiten/rendering/SKILL.md
- Study: docs/research/dg-requirements.md §7, Appendix D.1.

## Scope
internal/render: world renderer reading world grids + entities; sprite atlas integration; automap layer; panel framework; text renderer over CLX fonts; movie stub interface.

## Acceptance criteria
- [ ] Row-walk unit test reproduces C++ tile→screen mapping for a 5×5 map (golden screen buffer hash at progress=0)
- [ ] Per-tile draw order asserted via instrumented mock sprites (order list equality)
- [ ] Golden frames: 3 seeds × 3 view positions pixel-hash equal to C++ dumps (CI-runner only)
- [ ] Benchmarks: full 640×480 dungeon frame < 3ms on reference CI runner (port of dun_render_benchmark targets)
- [ ] Draw() allocs = 0 (testing.AllocsPerRun)
- [ ] Automap matches C++ automap_test cases
- [ ] Panels (control/char/spell/inventory/quest) render from ui_art with correct rects

## Skills for coder
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — SubImage/batching/ColorScale rules
/home/blaze/.agents/skills/go-dev/SKILL.md — hot-loop optimization without unsafe

## Dependencies & blockers
blocked_by dg-epic-02. Blocks dg-epic-05 (UI uses render), dg-epic-11 (automap qol).

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DEFERRED — question answered by recorded milestone decision; revisit point documented in the question text above.
- Per-pixel lightmap mode parity: optional flag, default table-based (decided; revisit at M2).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: World/sprite/lighting/panel/text rendering. Wired blockedBy 5wsb, blocks xih1/ftbu/41p0 ✓. 12 children (incl. lighting set kx0l/m18k/j9be/pdpc and graphics-extras gk5e/66o7/7aul/qeyu).
2 local: Cites verified: scrollrt.cpp:1229 DrawGame ✓, DrawFloor:955 (body 949-988 ✓), wall pre-pass 1000-1036 ✓, DrawDungeon def at :781 (body says :790-980 — def line drift only, function confirmed); dun_render.cpp/automap_render.cpp/text_render.cpp/control_panel.cpp all exist ✓. Skills paths ✓.
3 online: ebitengine batching/perf rules — verified in graphics wave (G10 decisions, online-checked per process log).
4 solution: none — ready. Minor cite drift (DrawDungeon 781 vs "790"); open question on per-pixel lightmap correctly records decided default. Resolution AC (640×480) matches DECIDED #5 ✓.
5 understanding: engine/render layer heart; the G10 binding decisions are retrofitted into child beans per milestone log.
6 repro: `grep -n "void DrawGame\|void DrawDungeon\|void DrawFloor" Source/engine/render/scrollrt.cpp`
7 fix: none. Optional cite refresh: `beans update devilGo-qbiy --body-replace-old "Per-tile order DrawDungeon :790-980" --body-replace-new "Per-tile order DrawDungeon :781-980"`
8 validation: body unchanged structurally; children carry G10 gates (batch-04/05 verify).
9 related: children ru6l, gk5e, 66o7, kx0l, m18k, j9be, pdpc, v1hy, zdj7, qeyu, mlv2, 7aul; blocker 5wsb; blocked xih1/ftbu/41p0.
VERDICT: TRIAGED-OK

## TRIAGE 2026-10-18 (pi wave, batch 2)

devilGo-qbiy** (Renderer epic, parent i56j, draft, phase-1)
1 Root: scrollrt-parity render pipeline — body/IMPL NOTES consistent.
2 Local: YAML `blocked_by: [devilGo-5wsb]` ✓. DrawGame:1229 ✓, DrawFloor:955 (body 949-988 ✓), **DrawDungeon def :781 vs body ":790-980"** — known drift, flagged by prior triage as optional body fix (cannot apply: READ-ONLY). All 12 children exist ✓.
3 Online: Ebiten perf/batching rules cited — already verified in graphics wave per IMPL NOTES; no new doc claim → SKIPPED.
4 Scope-fit: no dup — lighting split to kx0l/m18k/j9be/pdpc, frames to 7aul, automap to qeyu/zdj7; 7aul belongs to render not 1n0a's harness (harness *consumes* it) — wiring consistent.
5 DDD: engine/render layer reading world grids; movie stub isolated — correct.
6 Dep-order: 5wsb→qbiy→{xih1,ftbu,41p0} wired ✓.
7 AC-gaps: 7 ACs checkable; one cosmetic gap — stale cite ":790-980" should be ":781-980" (line-drift only, function confirmed).
8 Validation: 5×5 row-walk golden hash; order-list mock; 3×3 golden frames headless; AllocsPerRun=0.
9 Related: children ru6l…7aul; blocker 5wsb; blocked xih1/ftbu/41p0.
**TRIAGED-OK:devilGo-qbiy** *(deferred edit: `:790-980`→`:781-980`, blocked by READ-ONLY contract)*

BATCH-2-COMPLETE


Wave outputs: /tmp/dg-triage/out-b2.md

## Libraries

- Phase rule (DECIDED 2026-10-17): pure-Go only, no CGO. Phase-0 packages ran stdlib-only; concrete deps (modernc.org/sqlite, suprsokr/go-mpq, gopher-lua, ebiten) attach at the beans that consume them — never substituted by a worker.
