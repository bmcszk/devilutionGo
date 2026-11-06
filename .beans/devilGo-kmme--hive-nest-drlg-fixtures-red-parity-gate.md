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
updated_at: 2026-11-06T14:59:27Z
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
- [ ] Port pending: next unit ports the five nest branches against the RED gate

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
