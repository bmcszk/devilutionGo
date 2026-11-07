---
# devilGo-q5sv
title: Port PlaceCaveLights — caves L3 static light pass
status: completed
type: feature
priority: normal
tags:
    - phase-lighting
created_at: 2026-11-07T12:22:58Z
updated_at: 2026-11-07T12:39:23Z
---

## Scope

Port `PlaceCaveLights` — the caves (`DTYPE_CAVES`, levels 9-12) static light pass
(`levels/drlg_l3.cpp:2154-2168`) — and add its dLight parity goldens. This is the
one static light pass `devilGo-0yqm` ("Level light state — dLight array +
crypt/hive dPiece light sources") left unported; `devilGo-41p0` ("Lighting parity
test harness + dLight.dmp fixtures") recorded it as "one table row away".

Upstream shape (`drlg_l3.cpp:2154-2168`): for every world tile, if dPiece is in
55-146, or in 153-160, or is 149 or 151, `DoLighting({i, j}, 7, {})`.

Call site: `CreateL3Dungeon` (`drlg_l3.cpp:2192-2200`) → `Pass3()` (ExpandPieces)
→ `PlaceLights()` (`:2180-2188`) → `PlaceCaveLights()` (`:2187`) for
`leveltype != DTYPE_NEST`; the nest branch calls `PlaceHiveLights`.
`LoadL3Dungeon` (`:2213-2217`) repeats the same `Pass3(); PlaceLights();` pair.

## Out of scope

- vision (`devilGo-9o2h`), renderer (`devilGo-66o7`/`pdpc`/`m18k`), light-table
  generation (`devilGo-kx0l`, done), network light sync (`devilGo-h8uq`)
- no probe source change: `dvg_probe light` already dumps dLight after
  `CreateDungeon` for any level; only the variant matrix row is added

## Todo

- [x] Port PlaceCaveLights into internal/diablo/levels/cave_lights.go (collector shape of crypt_lights.go)
- [x] Dispatch Caves in levelLightSources (light.go), same as Crypt/Nest
- [x] Unit tests: cave tile classification + collection rule
- [x] Probe: add caves 9-12 rows to LightVariants, rebuild, dump light artifacts
- [x] Regenerate light/l9..l12-*.csv (64 CSVs: 4 levels x 2 entries x 8 seeds) + manifest rows + README table
- [x] light_parity_test.go: caves variants read the new goldens (16 -> 24 variants)
- [x] Non-vacuity: flip one caves golden cell -> RED with cell-exact diff -> restore -> GREEN
- [x] Gates: build, gofmt, lint, vet, go test ./..., beans-lint, sqlc-check, sha256sum -c manifest

## Skills for coder
- `~/.agents/skills/beans/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-tests/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Upstream evidence (pinned devilutionX ca91d778f)

- `Source/levels/drlg_l3.cpp:2154-2168` — `PlaceCaveLights`: nested x/y loops over
  MAXDUNX/MAXDUNY; `dPiece[i][j] >= 55 && <= 146` → `DoLighting({i, j}, 7, {})`;
  else `>= 153 && <= 160` → same; else `IsAnyOf(dPiece[i][j], 149, 151)` → same.
- `Source/levels/drlg_l3.cpp:2180-2188` — `PlaceLights`: `DTYPE_NEST` returns after
  `PlaceHiveLights()`; otherwise falls through to `PlaceCaveLights()`.
- `Source/levels/drlg_l3.cpp:2192-2200` — `CreateL3Dungeon`: `SetRndSeed` →
  `GenerateLevel(entry)` → `Pass3()` → `PlaceLights()`.
- `Source/levels/drlg_l3.cpp:2213-2217` — `LoadL3Dungeon` repeats `Pass3(); PlaceLights();`.
- `Source/gendung.cpp`-equivalent `Source/levels/gendung.cpp:349-372` —
  `CreateDungeon` dispatches `DTYPE_CAVES` and `DTYPE_NEST` to `CreateL3Dungeon`.
- `Source/lighting.cpp:117-180` — `DoLighting` (already ported in light.go):
  for CAVES neither NEST nor CRYPT, so the cone centre is set to 0
  (`source/lighting.cpp:153-159`), and LightFalloffs is the linear lean falloff
  (`light_tables.cpp:91-98`) — both already modelled by `internal/light`.
- `Source/levels/gendung.cpp:301-316` — `InitGlobals` seeds `dLight` with 15 for
  every non-town level; the caves pass is what departs from it.

No unported subsystem is involved: `DoLighting`, `LightFalloffs` and the dPiece
grid are all in tree.

## Progress log
- 2026-11-07: claimed. Read skills + pinned sources; creating the RED tests.


## Proof of Work (2026-11-07, branch dg-cavelights off origin/master 95aba5e)

### RED (before the port existed)
```
cave_lights_test.go:48:20: undefined: levels.PlaceCaveLights
cave_lights_test.go:86:21: undefined: levels.PlaceCaveLights
cave_lights_test.go:115:20: undefined: levels.PlaceCaveLights
FAIL github.com/bmcszk/devilutionGo/internal/diablo/levels [build failed]
```
and, once the collector existed but the dispatch did not:
```
--- FAIL: TestCavesApplyLightsShadesLava (0.05s)
    cave_lights_test.go:128: no tile brightened by the caves light pass
```

### GREEN
- `go build ./...` rc0
- `gofmt -l .` → empty
- `go vet ./...` rc0
- `./scripts/lint.sh` → `simpurity: internal/game internal/diablo clean (237 sources)`
  / `0 issues.`
- `go test ./... -count=1` → 44 packages ok, 0 FAIL
- `TestLightParity` → 192 variant/seed sub-tests PASS (24 variants × 8 seeds,
  caves 9-12 included), 0 FAIL
- `python3 scripts/beans-lint.py` → `OK (181 beans)`
- `just sqlc-check` → clean (`no raw SQL outside sqlc-generated code`)
- `sha256sum -c internal/testdata/drlg/manifest.txt` → 0 mismatches

### Non-vacuity (goldens really gate the port)
Flipped one lit cell of `light/l9-light-main-2588.csv` (index 1517 = tile
13,61, value 14 → 15):
```
--- FAIL: TestLightParity (0.08s)
    --- FAIL: TestLightParity/l9-light-main-seed-2588 (0.08s)
        light_parity_test.go:123: ../../testdata/drlg/light/l9-light-main-2588.csv: dLight[13][61] = 14, want 15
        light_parity_test.go:123: ../../testdata/drlg/light/l9-light-main-2588.csv: 1 of 12544 dLight cells differ
FAIL
```
Restored the file (sha256 back to
`6e19cc018b4519a608a11b1c42b9273cfa82157eda0aa167dd197798aae48efe`) → suite green.

### Fixture provenance / reproducibility
- probe rebuilt from this branch's `tools/probe` sources against pinned
  devilutionX `ca91d778f` (`cmake --build build-probe --target dvg_probe` rc0)
- `dvg_probe light` run twice → 192 artifacts, byte-identical between runs
- `go run ./tools/goldens light` twice → 192 CSVs, byte-identical
- `diff -rq` probe-then-goldens output vs `internal/testdata/drlg/light` →
  no differences (committed corpus is exactly what the probe produces)
- the 128 pre-existing nest/crypt CSVs came out unchanged after the rebuild
  (`cmp` loop, 0 mismatches), so this change adds only the 64 caves rows

## Summary of Changes

- `internal/diablo/levels/cave_lights.go` (new) — `PlaceCaveLights(PieceGrid) []Light`
  ports `levels/drlg_l3.cpp:2154-2168`: three tile predicates (55-146, 153-160,
  149/151) at radius 7 (`:2159,2161,2163`), grid order, min-write via the
  existing `DoLighting`.
- `internal/diablo/levels/light.go` — `levelLightSources` dispatches `Caves` to
  the new pass beside `Crypt`/`Nest`; `ApplyLights` doc now names the three
  families and the `PlaceLights` dispatch site (`drlg_l3.cpp:2180-2188`).
- `internal/diablo/levels/l3.go` — `GenerateL3` doc records the third
  `CreateL3Dungeon` step (`:2192-2200`): caller runs `ExpandPieces` then
  `ApplyLights`.
- Tests (all in `package levels_test`, no skips): collection rule over all 102
  matching + 7 near-miss ids; source count/positions over all 64 committed caves
  dPiece goldens; a real caves-9 generate→expand→light run added to
  `TestApplyLightsAfterRealGeneration`; `TestLightParity` grew the 8 caves
  variants (192 pairs).
- `tools/probe/dvg_probe.cpp` — `LightVariants` adds levels 9-12 both entries;
  comments name the caves pass. No other probe change: `light` already dumps
  dLight for any level post-`CreateDungeon`.
- Fixtures: 64 new `internal/testdata/drlg/light/l{9,10,11,12}-light-{main,prev}-*.csv`,
  64 manifest sha256 rows + updated `light/` header, README dLight section
  (24 variants / 192 files, caves table row, producer note).

### Notes / deviations
- None new. The pre-existing light-table build-order deviation
  (`internal/testdata/drlg/README.md`, "Known deviation") is unchanged and now
  also covers the caves rows.
- No unported subsystem is needed: `DoLighting`, `LightFalloffs`, `ExpandPieces`
  and the dPiece grid are all in tree; `PlaceCaveLights` reads nothing else.
- Scope untouched: vision (devilGo-9o2h), renderer (devilGo-66o7/pdpc/m18k),
  light-table generation (devilGo-kx0l), net lighting (devilGo-h8uq).
