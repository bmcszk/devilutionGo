---
# devilGo-7aul
title: 'Full-frame parity harness: C++ backbuffer dumper + Go comparator + seed-dungeon goldens'
status: todo
type: task
priority: normal
tags:
    - plan:task
    - graphics-extras
created_at: 2026-10-15T17:58:19Z
updated_at: 2026-10-16T14:26:52Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-xcdz
    - devilGo-ru6l
    - devilGo-gk5e
    - devilGo-66o7
    - devilGo-zdj7
---

# Pixel-parity harness

## Goal
Make 'pixel-exact' falsifiable: dump C++ indexed backbuffers for deterministic scenes, assert Go produces byte-identical index buffers (and colour-identical PNGs modulo a documented allow-list), runnable in CI without a GPU.

## Research
- docs/research/dg-graphics.md G11 (plan P1-P4, allowed palette-stage differences a-g), G1 (frame order), G8 (capture.cpp flow).
- Source/capture.cpp:84 CaptureScreen, :97 RedPalette; Source/utils/surface_to_pcx.cpp:45/70/92/150; Source/utils/surface_to_png.hpp
- Source/engine/render/scrollrt.cpp:1814 `scrollrt_draw_game_screen` (early-return at `:1816`) and `:1834`
  `DrawAndBlit` (early-return at `:1836`) — this is why a dump hook is needed rather than a screenshot loop
- Source/engine/palette.cpp:271/324 (short-circuits at `:273`/`:275` and `:326`/`:328`), `:363`/`:373`/`:391` cycling
- Source/diablo.cpp:1031-1033 (--record/--demo/--timedemo), :1127 parse, :1204 InitPlayBack
- test/timedemo_test.cpp:28-110 (drives the same code path with only SDL_INIT_EVENTS), :89 AdjustToScreenGeometry
- Fixtures to reuse: test/fixtures/diablo/*.dun, test/fixtures/hellfire/*.dun, test/fixtures/levels/l{1..4}data/*.dun, test/fixtures/light_render_benchmark/dLight.dmp, test/fixtures/text_render_integration_test/*.png, assets/arena/*.dun
- tools/linux_reduced_cpu_variance_run.sh, tools/measure_timedemo_performance.py
- docs/research/dg-requirements.md Appendix H H14 (no GPU variance rule), Appendix K (C++ test → Go test matrix)

## Scope
(a) a gtest in the devilutionX tree (`test/backbuffer_dump_test.cpp`, documented in devilutionGo/docs/parity.md, NOT committed to this repo by this task) writing `<frame>.idx.bin` + `<frame>.pal.bin` + `<frame>.ppm`;
(b) internal/render/parity: loader for the dumps, `Diff`, per-region allow-list, `cmd/dgo-parity` CLI (`-record`, `-compare`, exit code);
(c) 3 seeds × 3 view positions × progress∈{0,64,127} golden frames generated from .dun fixtures with a scripted minimal world (player, monster, item, object, arch, automap on/off);
(d) `just parity` target and CI wiring (ubuntu/macos/windows, no display).

## Acceptance criteria
- [ ] `just parity` passes on a headless runner and fails if any golden index byte changes (diff report names the first differing (x,y) and the owning tile/sprite)
- [ ] Index-exact gate present AND separate colour gate present, with the allow-list encoded as data (brightness=0, cycling off, fades settled, QoL overlays off, debug overlays off)
- [ ] At least one golden frame each for: town (trees deferred-one-row rule), L1 cathedral (arches + transparency + foliage), L4 hell (light-table override), crypt (nest/crypt override + lava brightness)
- [ ] Automap-on golden frame compares (dg-task-013 dependency)
- [ ] Text goldens reused from test/fixtures/text_render_integration_test/*.png via the same comparator (proves the pipeline composes)
- [ ] Documentation: how to regenerate the C++ dumps (commands, SDL_VIDEODRIVER=dummy requirement, MPQ location) verified by running them once

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden-file + fixture matrix
/home/blaze/.agents/skills/ebiten/rendering/SKILL.md — ReadPixels/WritePixels only inside test helpers
/home/blaze/.agents/skills/go-dev/SKILL.md — CLI structure, no CGO, build tags for headless
/home/blaze/.agents/skills/beans/SKILL.md — keep this bean's todo list current

## Dependencies & blockers
- devilGo-v1hy (fix2: missing edge wired)
blocked_by dg-task-010, dg-task-201, dg-task-202, dg-task-005. Blocks the dg-epic-03 'golden frames' acceptance criterion and milestone M2 sign-off.

## Open questions
- Legal/CI question: can the job fetch retail MPQs? If no, all goldens must come from the committed fixtures + bundled assets path.
- Do we need Windows/macOS goldens (fonts/driver differences) or is one platform + 'CPU variance' rule enough?


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.


## Ebiten graphics-migration gates (from docs/research/dg-graphics.md G10/G11 — binding)
- [ ] Full-frame pixel-diff harness: C++ backbuffer dumper (via `capture.cpp` headless mode) vs Go `screen` dump; diff tool allows only documented palette-stage deltas (G11 allowed list).
- [ ] Seed-dungeon golden frames in CI (GPU-variance-free: CPU-side index rasteriser as reference per G11).
- [ ] `Layout()` returns fixed 640×480; FilterNearest integer scaling per G10 D1.


## Resolution gate (grilling decision #5)
- [ ] 640×480 world logical surface rendered correctly, AND 2x window (1280×960) integer-scaled without artifacts — parity golden frames at BOTH sizes.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps (automap-on + text-compose ACs reference zdj7/v1hy, not wired; 2 one-line cite drifts)


## IMPL NOTES (triage 2026-10-16)
1 root: Graphics-wave task (dg-task-203): makes "pixel-exact" falsifiable in CI. Wire: blockedBy [xcdz, ru6l, gk5e, 66o7] — EXACT match to body (dg-task-005/010/201/202) ✓. 41p0 (lighting harness) correctly blockedBy 7aul ✓ — complementary, not duplicate.
2 local: Cites verified precisely: scrollrt.cpp:1814 `scrollrt_draw_game_screen` + early-return :1816, :1834 `DrawAndBlit` + :1836 ✓ (this is the dump-hook justification); palette.cpp:271-276/324-328 fade short-circuits (HeadlessMode/demo) ✓; diablo.cpp:1031-1033 --record/--demo/--timedemo ✓, :1204 InitPlayBack ✓; timedemo_test.cpp AdjustToScreenGeometry :90 (bean :89, drift 1) ✓; capture.cpp CaptureScreen :85 (bean :84, drift 1) ✓. ALL fixture claims exist: test/fixtures/diablo/*.dun ✓, hellfire/*.dun ✓, levels/l1data/*.dun ✓, light_render_benchmark/dLight.dmp ✓, text_render_integration_test/*.png ✓, assets/arena/{church,hell,circle_of_death}.dun ✓; tools/linux_reduced_cpu_variance_run.sh + measure_timedemo_performance.py ✓.
3 online: n/a.
4 solution: Two ACs name deps that aren't wired: "Automap-on golden frame (dg-task-013 dependency)" needs zdj7 (+qeyu) and "Text goldens reused via same comparator (proves pipeline composes)" needs v1hy. Either wire them or move those ACs to a follow-up. Open questions (retail-MPQ fetch in CI; per-platform goldens) are operational with recorded defaults — fine.
5 understanding: testing layer for the render epic; C++ side gtest intentionally not committed here; `just parity` target aligns with existing justfile (test-parity present at :116).
6 repro: `sed -n '1814,1816p;1834,1836p' Source/engine/render/scrollrt.cpp; ls assets/arena/; beans query --json '{ bean(id:"devilGo-7aul") { blockedByIds } }'; grep -n test-parity justfile`
7 fix: `beans update devilGo-7aul --blocked-by devilGo-zdj7 --blocked-by devilGo-v1hy` (verify additivity) — or rescope: drop those two ACs into a "phase 2 goldens" follow-up bean. Minor: `cmd/dgo-parity` naming vs repo `cmd/devilutionGo-*` convention.
8 validation: blockedByIds includes zdj7+v1hy (or follow-up bean exists carrying the two ACs); all cited fixture paths still present.
9 related: xcdz (fixture loader dep), 41p0 (lighting harness, blockedBy this), gk5e/66o7/ru6l (deps), uc44 (timedemo oracle).
VERDICT: NEEDS-WORK:wire-deps (automap-on + text-compose ACs reference zdj7/v1hy, not wired; 2 one-line cite drifts)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.