---
# devilGo-41p0
title: Lighting parity test harness + dLight.dmp fixtures (L8 + G11)
status: draft
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-16T14:06:43Z
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


## IMPL NOTES (triage 2026-10-16)
1 root: End-to-end lighting pixel-parity harness (dLight.dmp + seed-dungeon pixel diff with tolerance list); status `draft`, parent 1n0a, blocked_by = {qbiy, **7be9**, kx0l, bodd, m18k}.
2 local: `test/fixtures/light_render_benchmark/dLight.dmp` EXISTS (only file in dir) ✓; reader is `test/light_render_benchmark.cpp:20-26` ✓ — but body says "extend test/fixtures/light_render_benchmark/main.cpp": **no main.cpp exists there** (bad cite; target is test/light_render_benchmark.cpp or a new gtest, mirroring 7aul's backbuffer_dump_test.cpp pattern). Frontmatter dep **devilGo-7be9 = "Combat resolution pipeline" (child of nxxo)** — unrelated to lighting; body actually names devilGo-7aul (correct) + gk5e/66o7/j9be, none of which are in frontmatter. Body's "All 251-257 tasks" and "(already in task 252)" are stale pre-migration refs (252 ≈ kx0l).
3 online: n/a (no external claims).
4 solution: fix frontmatter deps (drop 7be9, add 7aul + gk5e/66o7/j9be), fix main.cpp cite, refresh 251-257 refs to devilGo ids; add one sentence drawing the boundary vs devilGo-7aul (7aul = general backbuffer/dumper/comparator; 41p0 = lighting-specific fixtures+tolerances on top) to prevent CI-harness duplication.
5 understanding: lighting-domain specialization of the 7aul pixel-parity harness — it must build on, not re-implement, the dumper/comparator.
6 repro: `ls devilutionX/test/fixtures/light_render_benchmark/` (dLight.dmp only, no main.cpp); `grep -m2 '^title\|^parent' .beans/devilGo-7be9*.md` (Combat resolution, parent nxxo); `beans show devilGo-41p0 --json` blockedByIds.
7 fix: `beans update devilGo-41p0` → remove blocked-by 7be9, add blocked-by 7aul, gk5e, 66o7, j9be; body replace "main.cpp" cite with "test/light_render_benchmark.cpp (or a new test/backbuffer-style gtest per devilGo-7aul)"; replace "251-257"/"task 252" with devilGo ids (kx0l, bodd, m18k, pdpc, j9be).
8 validation: `beans query --json '{ bean(id:"devilGo-41p0") { blockedByIds } }'` no 7be9, has 7aul; body has no "main.cpp" / "task 252".
9 related: devilGo-7aul (parent harness), devilGo-kx0l/bodd/m18k (unit fixtures), devilGo-b782 (palette cycling).
VERDICT: NEEDS-WORK:wrong-frontmatter-dep-7be9(should-be-7aul);main.cpp-cite-nonexistent;stale-251-257-refs
