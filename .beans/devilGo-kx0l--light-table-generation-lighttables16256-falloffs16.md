---
# devilGo-kx0l
title: Light table generation — LightTables[16][256], falloffs[16][128], cone[8][8][16][16] (L2)
status: todo
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:04Z
updated_at: 2026-10-16T20:39:43Z
parent: devilGo-qbiy
blocked_by:
    - devilGo-n89r
    - devilGo-b782
    - devilGo-bodd
    - devilGo-pdpc
    - devilGo-1spo
---

# Light table generation — byte-exact port of MakeLightTable()

## Goal
Generate all 4 tables exactly as C++ light_tables.cpp:24-119: 16x256 shade tables with leveltype branches (normal/hell/nest+crypt), falloffs with quadratic vs linear formulas (float32 arithmetic), cone interpolations (sqrt double promotion), TRN loads, color cycling rotate.

## Research
- Source/engine/light_tables.cpp:24-119 (full function)
- Source/engine/light_tables.hpp:10-17 (extern declarations)
- Source/engine/lighting_defs.hpp:14-17 (LightTableSize=256, NumLightingLevels=16, NumLightRadiuses=16)
- Constants: VolumeScale=3321.9281, MillibelMin=-10000, StereoSeparation=6000 (from soundsample.cpp, also used in light? No - that's audio; lighting uses LightFalloffs only)
- LightFalloffs formula: factor=d/maxDist; NEST/CRYPT: brightness=radius*1.25; scaled=factor^2*brightness + (15-brightness); else scaled=factor*15; scaled+=0.5; uint8
- LightConeInterpolations: a=8*x-offX; b=8*y-offY; C=uint8(sqrt(double(a*a+b*b))) — double promotion!
- Hell branch: color = ((15<<4)/15)*(15-i)/16*(j+1); color=1+(color>>4); symm j+1 & 31-j
- NEST/CRYPT branch: iota(0..15) on first 16 entries; table[15][0]=0; table[15][1..15]=1
- Color cycling: rotate(begin+1, begin+2, begin+32) per table, called every frame

## Scope
Package: `internal/light/tables.go`
- `func MakeLightTables(levelType DungeonType) (LightTables [16][256]uint8, Falloffs [16][128]uint8, Cone [8][8][16][16]uint8, FullyLit, FullyDark *[256]uint8)`
- Exact int8/uint8 arithmetic; int divisions first, then float32 operations
- sqrt uses math.Sqrt(float64()) matching C++ double promotion
- levelType switch for three branches (normal/hell/nest+crypt)
- Return nil for FullyLit/FullyDark when disabled (hell/nest+crypt)
- ColorCycle(offset int) rotates slice[1:32] of each table

## Acceptance criteria
- [ ] `TestLightTablesIdentity` shade 0 table = identity (0..255); table15 all zeros (unless hell/nest+crypt branch mutates)
- [ ] `TestFalloffsGolden` all 2048 bytes match C++ fixture (`test/fixtures/light_render_benchmark/falloffs.bin`)
- [ ] `TestConeGolden` all 16384 bytes match C++ fixture (`test/fixtures/light_render_benchmark/cone.bin`)
- [ ] `TestHellBranchColors` leveltype=HELL → tables 0..14 have blood-wall color override, FullyLit=nil, table[15] unchanged
- [ ] `TestNestCryptBranch` leveltype=NEST/CRYPT → first 16 entries = 0..15; table15[0]=0, [1..15]=1; FullyDark=nil
- [ ] `TestColorCycling` rotate(1,2,32) applied each tick → 31-step period, byte-dump after N frames matches C++

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md

## Dependencies & blockers
- devilGo-n89r (palette/TRN decoder for table index→color pipeline)
- devilGo-b782 (palette registry level-specific)

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: fixture dependency resolved — falloffs/cone binaries are produced by devilGo-vsx3 (probe toolchain) and stored via devilGo-1spo; this bean is blocked-by devilGo-1spo in the tracker.
- None - formulas transcribed verbatim from light_tables.cpp:24-119 with integer/float32 order.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:missing-fixture-source (falloffs.bin/cone.bin don't exist; no task explicitly creates them) + inverted/bogus dependencies (blocked_by m18k and bodd) + MakeLightTable ownership duplicated with 66o7.


## IMPL NOTES (triage 2026-10-16)
1 root: Pure-data bean — byte-exact port of MakeLightTable() into internal/light/tables.go. Formulas verified accurate; fixture and dependency problems.
2 local: `beans show` + probes against `Source/engine/light_tables.cpp:18-119` (MakeLightTable at :20): hell branch verified at :55-70 (`color=((Range-1)<<4)/shades*(shades-i)/Range*(j+1); color=1+(color>>4); idx=j+1 & 31-j`; FullyLitLightTable=nullptr) ✓; nest/crypt branch :66-72 (iota 0..15, table[15][0]=0, [1..15]=1, FullyDarkLightTable=nullptr) ✓; falloffs :79-101 (factor²·brightness+(15−brightness) quadratic vs linear, `scaled += 0.5F` float32) ✓; cone :104-113 (`uint8_t sqrt((a*a)+(b*b))`, double promotion) ✓; cycling `Source/lighting.cpp:481` rotate(1,2,32) ✓; dims `LightFalloffs[NumLightRadiuses=16][128]` (light_tables.hpp:13,:23) ✓. BUT: AC fixtures `test/fixtures/light_render_benchmark/falloffs.bin` and `cone.bin` DO NOT EXIST anywhere (C++ dir has only dLight.dmp; `grep -n falloff test/light_render_benchmark.cpp` → empty — no dumper emits them). Frontmatter blocked_by: n89r, b782, bodd, m18k — inverted: m18k's shader CONSUMES these tables, so kx0l should block m18k, not vice versa; bodd (emitter data, runtime lighting L3) is not a table-generation dependency. Minor: bean says "TRN loads" happen in MakeLightTable — they don't (TRNs load in lighting.cpp:214 LoadTrns); audio constants paragraph (VolumeScale etc.) is acknowledged-as-irrelevant noise.
3 online: n/a (no external claims).
4 solution: add a fixture-generation step to scope (extend light_render_benchmark dumper to emit falloffs.bin/cone.bin, or fold into 41p0's dumper and make that an explicit blocked_by); invert the m18k edge; justify or drop bodd dep; drop the TRN-loads and audio-constants lines; demarcate MakeLightTable ownership vs 66o7 (66o7 scope includes "internal/render/lighttable.go (port MakeLightTable verbatim)" and AC `TestMakeLightTable` 256×16×6 — duplicate of this bean's entire purpose; make 66o7 consume kx0l's internal/light/tables.go instead).
5 understanding: Diablo-layer pure data feeding every renderer consumer (66o7 palette rows, m18k shader, pdpc lightmap, 41p0 parity tests). Correctly engine-agnostic, no GPU needed — ideal CI-first bean once fixtures exist.
6 repro: `beans show devilGo-kx0l --body-only`; `sed -n '55,72p' <devilutionX-repo>/Source/engine/light_tables.cpp`; `ls <devilutionX-repo>/test/fixtures/light_render_benchmark/`; `grep -rn falloff <devilutionX-repo>/test/light_render_benchmark.cpp`; `sed -n '/^blocked_by/,/^---/p' .beans/devilGo-kx0l--*.md` (or frontmatter block).
7 fix: `beans update devilGo-kx0l --body-replace-old "- \`TestFalloffsGolden\` all 2048 bytes match C++ fixture (\`test/fixtures/light_render_benchmark/falloffs.bin\`)" --body-replace-new "- Fixture-generation FIRST: extend the C++ light_render_benchmark (or 41p0 dumper) to emit falloffs.bin/cone.bin; then \`TestFalloffsGolden\` all 2048 bytes match that fixture"`; dependency edge fix: `beans update devilGo-kx0l --blocked-by devilGo-n89r,devilGo-b782,devilGo-41p0` (drop m18k+bodd, add fixture owner) AND `beans update devilGo-m18k --blocked-by devilGo-gk5e,devilGo-66o7,devilGo-j9be,devilGo-kx0l`; ownership note: `beans update devilGo-66o7 --body-append "MakeLightTable port is owned by devilGo-kx0l (internal/light/tables.go); 66o7 consumes it — remove the duplicated lighttable.go/TestMakeLightTable scope line here."`
8 validation: `beans show devilGo-kx0l --json` blockedBy contains no m18k; falloffs.bin generation step present in body; 66o7 body no longer claims the MakeLightTable port.
9 related: 66o7 (duplicate AC — canonical data owner should be kx0l), m18k (inverted edge), 41p0 (fixture harness; its body already says "TestFalloffsGolden/TestConeGolden already in task 252" i.e. assumes this bean), pdpc/bodd (phase-lighting siblings).
VERDICT: NEEDS-WORK:missing-fixture-source (falloffs.bin/cone.bin don't exist; no task explicitly creates them) + inverted/bogus dependencies (blocked_by m18k and bodd) + MakeLightTable ownership duplicated with 66o7.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).


