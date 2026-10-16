---
# devilGo-qbiy
title: 'Renderer: world, sprites, lighting, panels, automap, text'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-1
created_at: 2026-10-15T16:37:32Z
updated_at: 2026-10-16T14:01:05Z
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
- Study: /tmp/dg-requirements.md §7, Appendix D.1.

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
- Per-pixel lightmap mode parity: optional flag, default table-based (decided; revisit at M2).

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
