---
# devilGo-41p0
title: Lighting parity test harness + dLight.dmp fixtures (L8 + G11)
status: completed
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-11-07T11:55:04Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-qbiy
    - devilGo-7be9
    - devilGo-kx0l
    - devilGo-bodd
    - devilGo-m18k
    - devilGo-1spo
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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
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
OPEN-QUESTION RESOLUTION 2026-10-16: RESEARCH-FIRST: Apocalypse flash path (missiles.cpp) traced as first implementation step; add to scope if present (expected).
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

Fix pass 2026-10-16: Dep-wiring applied (fix pass): removed stale 7be9 blocker; wired 7aul/gk5e/66o7/j9be. Cite drifts corrected per IMPL NOTES (main.cpp → test/light_render_benchmark.cpp; devilGo-kx0l..257 → devilGo ids kx0l/bodd/m18k/pdpc/j9be).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.

## Zero-dependency test constraint\n\nParity tests compare committed Go-owned fixtures only. C++ dumpers/probes may be used once in an offline producer workflow, but must never run in go test, CI, production code, or ordinary fixture verification. No DG_DEVILUTIONX or external source-tree access; missing fixtures are hard failures.


## Proof of Work

Landed as PR #47 (branch `dg-lightparity` off `origin/master` 742bb08), 2 commits:
`d634da4` probe+goldens, `6a1ba22` harness+corpus.

### AC tick (scoped to this lane)

The bean's original AC list is renderer-side (pixel diff, tolerance list,
`cmd/lightparity`). The lane contract scopes this PR to the **dLight state grid**
— renderer/palette/TRN is a separate lane. ACs re-scoped accordingly:

- [x] **dLight fixture exists for a known seed grid.** 128 CSVs under
  `internal/testdata/drlg/light/` (nest 17-20 + crypt 21-24, both entries × the
  8 pinned seeds of `tools/probe/seeds.txt`), produced from devilutionX
  `ca91d778fba92a1f680afb1fb2cdf09337ff1b42` by `dvg_probe light` and
  `goldens light`. Replaces the "update dLight.dmp if the C++ algorithm
  changes" AC with the committed-text corpus shape (repo decision: no raw
  binaries).
- [x] **Producer is the dumped upstream binary, not a reimplementation.**
  `dLight` is an exported global (`Source/levels/dun_tile_data.hpp:105`), so no
  generator hook was needed: `DumpLightGrids` in `tools/probe/dvg_probe.cpp`
  runs the real `CreateDungeon` and copies the array out. Verified
  deterministic: two clean runs `diff -r` byte-identical (128/128).
- [x] **Go side reproduces it exactly.** `TestLightParity`
  (`internal/diablo/levels/light_parity_test.go`) rebuilds each variant with the
  port's generator, runs Pass3 + the static light pass, and diffs all 12544
  cells. 128 sub-tests PASS.
- [x] **Regeneration is ONE command** (ADR-021 fold): `goldens light
  <artifact-dir> <corpus-dir>`. No new tool; `tools/goldens/main.go` gained one
  table row, and `main_test.go` pins it.
- [x] **Corpus pinned by sha256.** `internal/testdata/drlg/manifest.txt` carries
  128 new rows; `sha256sum -c` over the whole corpus = 779 rows, 0 mismatches.
- [x] **Non-vacuous.** Mutating one golden cell fails the test
  (`dLight[44][72] = 13, want 14`, "1 of 12544 dLight cells differ"); restoring
  it goes green. Test does not read MPQs, the C++ tree, or the network.
- [x] **Upstream behaviour deviation documented, not ported silently.**
  `MakeLightTable` runs before `CreateDungeon` in the probe instead of after
  (`diablo.cpp:3467` vs `:3373`), because upstream's order makes the first
  level's radii depend on session history. Recorded in
  `internal/testdata/drlg/README.md` "Known deviation" and in the probe table
  comment on `DumpLightGrids`.
- [ ] **PlaceCaveLights levels 9-12** — out of scope: that pass is not ported
  yet (`internal/diablo/levels` has no caves source collection). Follow-up when
  it lands; the matrix is one table row away.
- [ ] **Renderer parity / tolerance list / `cmd/lightparity`** — out of scope,
  devilGo-66o7 lane (renderer, palette, TRN).

### Commands and observed results

```
$ go build ./...                      → BUILD_OK
$ gofmt -l .                           → (empty)
$ ./scripts/lint.sh                    → 0 issues.
$ go vet ./...                         → (clean)
$ go test ./... -count=1               → rc=0, 44 ok, 0 FAIL
$ python3 scripts/beans-lint.py        → beans-lint: OK (180 beans)
$ just sqlc-check                      → clean (no raw SQL outside sqlc)
$ cmake --build build-probe --target dvg_probe -j24 → [5/6] Linking CXX executable dvg_probe
$ (cd internal/testdata/drlg && sha256sum -c manifest.txt) → 779 rows OK, 0 mismatches
$ dvg_probe light ... twice; diff -r → identical
$ go run ./tools/goldens light ... twice vs committed → identical
```

### Files

| File | Change |
|---|---|
| `tools/probe/dvg_probe.cpp` | +`DumpLightGrids`/`DumpSeedLightGrids`, `LightVariants` matrix, `light` subcommand |
| `tools/probe/README.md` | subcommand row, refresher/gate table, goldens snippet |
| `tools/goldens/light.go` | new: `runLight` + artifact parser/encoder |
| `tools/goldens/light_test.go` | new: parser tests |
| `tools/goldens/main.go`, `main_test.go` | one `refresher` row + its pins |
| `internal/diablo/levels/light_parity_test.go` | new: `TestLightParity` |
| `internal/testdata/drlg/light/*.csv` | new: 128 committed goldens |
| `internal/testdata/drlg/manifest.txt` | +128 sha256 rows, +1 header line |
| `internal/testdata/drlg/README.md` | "dLight corpus" section incl. the deviation |

Residual: caves levels 9-12 (`PlaceCaveLights`) and all renderer-side parity
remain unported / in the renderer lane.
