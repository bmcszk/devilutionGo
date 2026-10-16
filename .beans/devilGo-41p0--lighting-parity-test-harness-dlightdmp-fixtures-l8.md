---
# devilGo-41p0
title: Lighting parity test harness + dLight.dmp fixtures (L8 + G11)
status: draft
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-16T14:01:03Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-qbiy
    - devilGo-7be9
    - devilGo-kx0l
    - devilGo-bodd
    - devilGo-m18k
---

# Lighting parity test harness + dLight.dmp fixtures

## Goal
End-to-end lighting parity: C++ backbuffer dumper (devilGo-7aul) vs Go renderer on seeded dungeon frames — pixel diff allowed only documented palette-stage differences (G11). Plus unit fixtures for every L2 table.

## Research
- G11 parity plan: CPU-side index rasterizer as reference; seed-dungeon CI goldens; allowed palette diffs documented
- C++ fixture: `test/fixtures/light_render_benchmark/dLight.dmp` (binary dump of dLight array after ProcessLightList on a known seed)
- G5: LightFalloffs/cone tables have `test/fixtures/light_render_benchmark/` directory
- Graphics G10: batching rules — light application must not break draw-call merging
- `lighting_color_cycling` rotate(1,2,32) every frame — parity across frames

## Scope
Package: `internal/light/parity_test.go` + `cmd/lightparity/` CI binary
- `func DumpDLight() []byte` in C++ (extend test/fixtures/light_render_benchmark/main.cpp) + Go equivalent
- `func CompareDLight(cpp, go []byte) (diffCount, firstDiffIndex, diffValues)`
- CI job: seed dungeon (cathedral/crypt/caves/hell/nest), run C++ dumper + Go renderer headless, pixel diff with tolerance list (palette animation frames, TRN states)
- Unit fixtures: `TestFalloffsGolden`, `TestConeGolden`, `TestLightTablesGolden` (already in task 252)
- `TestColorCyclingFrames` compare frame N and N+31 identical
- `TestApocalypseFlash` if flash overlay exists → golden frame at flash tick

## Acceptance criteria
- [ ] All L1-L7 tasks pass unit parity tests
- [ ] CI job `lighting-parity` runs on every PR, fails on any pixel diff outside tolerance list
- [ ] Tolerance list documented in `docs/lighting-parity-tolerance.md` (palette-cycle frame, TRN hero states, hell blood-wall animation)
- [ ] dLight.dmp fixture updated if C++ algorithm changes (versioned in repo)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md

## Dependencies & blockers
- devilGo-7aul (C++ backbuffer dumper must exist first)
- devilGo-gk5e (dungeon generation for seed)
- devilGo-66o7 (index framebuffer + palette shader)
- devilGo-j9be (transparency + light blend)
- All 251-257 tasks (unit parity)

## Open questions
- Apocalypse flash path (missiles.cpp) not fully traced — add if exists

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wrong-frontmatter-dep-7be9(should-be-7aul);main.cpp-cite-nonexistent;stale-251-257-refs
