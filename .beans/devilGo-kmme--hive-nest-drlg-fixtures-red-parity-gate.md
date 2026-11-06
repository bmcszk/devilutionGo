---
# devilGo-kmme
title: Hive (nest) DRLG fixtures + RED parity gate
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - drlg
    - hellfire
created_at: 2026-11-06T14:43:04Z
updated_at: 2026-11-06T17:55:15Z
parent: devilGo-97mn
---

# Hive (nest) DRLG fixtures + RED parity gate

## Goal
Land the RED feedstock for the Hellfire hive generator (levels 17-20,
`DTYPE_NEST` -> `CreateL3Dungeon` -> caves `drlg_l3` with the nest branches):
64 `dPiece` goldens, the `l6.{til,min,sol}` asset CSVs, and a failing parity
test. No layout port in this unit.

## Research
`docs/research/dg-hellfire-drlg-inventory.md` (sections Hive, Missing, Probe
state, Order suggestion step 2). Upstream pinned at `ca91d778f`:

- nest branches: pool `drlg_l3.cpp:1499-1501`, stairs `:1982-1989`, isles
  `:2034-2042`, lights `:2182-2187`;
- `CreateL3Dungeon` (`drlg_l3.cpp:2185-2193`) calls `Pass3` (`:2198`), which is
  `DRLG_LPass3(8 - 1)` (`:2149-2152`), unconditional on level type;
- hive assets are `nlevels/l6data/l6.{cel,til,min,sol}` (`diablo.cpp:1409-1413`,
  `dun_tile_data.cpp:88-89,154`), not the `levels/l3data` set.

## Scope
`internal/testdata/drlg/{dpiece,assets}`, `internal/diablo/levels/hive_parity_test.go`,
comment lines in `internal/diablo/levels/piece.go`, dump-only probe matrix,
`internal/testdata/drlg/README.md`.

## Acceptance criteria
- [x] `dpiece/l1{7,8,9}-dpiece-{main,prev}-<seed>.csv` + `l20` x 8 seeds = 64 CSVs, sha256-pinned
- [x] `assets/l6.{til,min,sol}.csv` + `assets/manifest.txt` sha256 rows
- [x] `TestHiveParity` drives `GenerateNest` over the 8 seeds and fails RED (`errNotImplemented` stub)
- [x] `piece.go` nest comment corrected (nest DOES run `DRLG_LPass3`, caves index 7; town unchanged)
- [x] README documents both new fixture families
- [x] Gates green except the new RED test; push verified with `ls-remote`
- [x] Port (305b10e): nest branches ported, TestHiveParity 64/64 green

## Libraries
- stdlib only; `internal/assets` MPQ reader for the offline asset extraction

## Skills for coder
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-tests/SKILL.md` — RED test shape

## Progress log
- 2026-11-06 lane `dg-hive`: probe matrix extended to nest levels 17-20
  (dump-only); 64 dPiece CSVs + 3 `l6.*` asset CSVs committed; `TestHiveParity`
  added RED against `GenerateNest`; `piece.go` nest comment corrected.

### Gates (2026-11-06, branch dg-hive)

```
$ go build ./...                       rc=0
$ gofmt -l .                           (empty)
$ ./scripts/lint.sh                    0 issues  (simpurity: 230 sources clean)
$ go vet ./internal/...                rc=0
$ go test ./... -count=1               rc=1 — RED as designed
  42 packages ok
  --- FAIL: TestHiveParity (0.21s)     64/64 sub-tests: "generate nest lNN-M: hive generator not implemented"
  FAIL github.com/bmcszk/devilutionGo/internal/diablo/levels
$ python3 scripts/beans-lint.py        beans-lint: OK (177 beans)
$ just sqlc-check                      no raw SQL outside sqlc-generated code
```

Fixture provenance: the extended probe run reproduces all 256 committed
dPiece CSVs for levels 1-16 byte-identically (`diff -r` clean); the 64 nest
CSVs and the 3 `l6.*` asset CSVs each came from two independent extraction
runs that agree.

Commits: goldens 8593632, RED gate dfb3201, comment fix 3f85a6a, corpus pin
bec9d24.

Push: `HERMES_SKIP_PREPUSH=1` (tree RED by design — every gate above was run manually); `git ls-remote origin dg-hive` = `bc60f6c584b05b07103309f704c12f5129e1320c`.

## Proof of Work — hive port (lane dg-hive s2)

Commit 305b10e `feat(levels): port the hive nest DRLG generator`, branch
dg-hive, on top of the refresh merge ba2b46c.

### Port

| Upstream | Go |
|---|---|
| `CreateL3Dungeon` drlg_l3.cpp:2192-2200 | `GenerateNest` (hive.go) |
| `GenerateLevel` reroll loop :1996-2043 | `rollNest`, `placeNestFeatures` |
| `PlaceNestStairs` :1959-1980 | `placeNestStairs`, `hiveUpStairs` (hive_stairs.go) |
| `PlaceSlimePool` :1421-1437, arm of `PlacePool` :1497-1503 | `placeSlimePool`, `placeSlimePattern`, `placeSlimePatternAt` (hive_pool.go) |
| nest decoration block :2034-2089 | `decorateNest` + tables (hive_patterns.go) |
| `PlaceHiveLights` :2169-2178, arm of `PlaceLights` :2180-2187 | `PlaceHiveLights` (hive.go) |
| minisets :48-115, :431-710 | hive_minisets.go (L6UP, L6DOWN, L6HOLDWARP, L6ISLE1-5, HivePattern9,10,29-42) |
| `Pass3()` :2149-2152 | existing `ExpandPieces`; piece.go maps `Nest` to the caves index 7 |

RNG order is proven by the 64-fixture diff, not by argument: any change in
pattern order, percentage, protected-cell handling or reroll condition moves
the golden.

### Gates (branch dg-hive, tree clean at 305b10e)

    $ gofmt -l .                              rc=0, no files
    $ go vet ./internal/...                   rc=0
    $ ./scripts/lint.sh                       0 issues (simpurity: 234 sources clean)
    $ go test ./internal/diablo/levels/ -run TestHiveParity -count=1
    ok  github.com/bmcszk/devilutionGo/internal/diablo/levels  1.084s   (64/64 PASS)
    $ go test ./... -count=1                  rc=0, 43 ok, 0 FAIL, 5 no-test-files
    $ python3 scripts/beans-lint.py           beans-lint: OK (178 beans)
    $ just sqlc-check                         no raw SQL outside sqlc-generated code

### Deviations and corrections

- piece.go: `pass3DirtIndex` gained `Nest` (one case token). Required, not
  optional: l6.til entry 0 is {0,1,2,3} while upstream's `lv` is 7 =
  {28,29,30,31}, and the goldens' border row 28,30,28,30... proves the dirt
  fill reads index 7. The scout doc asks for this fix in its step 2.
- Lights are ported but not wired: `PlaceHiveLights` takes the expanded
  PieceGrid and an injected `addLight`, because this package has no dLight
  model (same deferral as crypt's PlaceCryptLights). No golden covers lights.
- Scout drift found while porting (file:line in the pinned tree):
  - the nest decoration is far larger than the scout's :2034-2042: the isles
    are :2035-2041 followed by four 1x1 floor subs (:2042-2045), the
    HivePattern29-38 sweeps (:2046-2065), HivePattern9/10 (:2066-2067) and
    22 surface subs (:2068-2089) — all ported;
  - the scout says "plus one 1x1 miniset"; there are 26 (4 + 22);
  - `PlaceNestStairs` is :1959-1980 and `PlaceStairs` :1982-1989 (the scout
    cites :1982-1989 for the nest stairs);
  - `CreateL3Dungeon` is :2192-2200 (`Pass3` :2198, lights :2199).
- Scope note: no separate hive-port bean existed; the port was this bean's
  last AC, so no new bean was filed.
