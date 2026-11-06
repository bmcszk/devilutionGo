---
# devilGo-t1lp
title: 'Town DRLG goldens: probe town subcommand + inventory doc'
status: completed
type: task
priority: high
created_at: 2026-11-05T23:48:01Z
updated_at: 2026-11-06T00:52:30Z
parent: devilGo-97mn
---

Lane dg-town s1 (goldens only, no Go port).

## Goal
RED-gate feedstock for CreateTown (town.cpp:359): committed CSV goldens plus a
linked inventory of the upstream town surface.

## Deliverables
- [x] docs/research/dg-town-inventory.md
- [x] dvg_probe town subcommand (dump-only, additive patch)
- [x] internal/testdata/drlg/town/*.csv (8 seeds, main only: the probe dumps ENTRY_PREV in-process, fails if it differs from ENTRY_MAIN, and writes no prev CSV)
- [x] README correction: committed seeds are seeds.txt, not {0,1,2,77,...}
- [x] gates + push

## Out of scope
Go port of CreateTown, parity test (later sessions).


## Triage (2026-11)

Disk-verified findings that change the contract:

1. Seed set: the lane contract named `{0,1,2,77,2588,5193,7777,13131333}` but
   the committed corpus convention is `tools/probe/seeds.txt` =
   `2588,743271966,1383137027,844660068,19770182,1522546307,125121312,1511478689`.
   Use seeds.txt (8 seeds) — the same set the 256 dPiece goldens use.
2. Town RNG: `CreateTown` (town.cpp:359) takes no rseed, and
   `SetRndSeedForDungeonLevel()` (diablo.cpp:3149) runs *after* `CreateLevel`
   (diablo.cpp:3392-3395), so at the CreateTown call site upstream never pins
   the RNG. The only RNG consumer is `PickRandomlyAmong({1,2,3,4})` in the
   `!IsWarpOpen(DTYPE_HELL)` branch (town.cpp:220-227, 10 calls). The probe must
   pin `SetRndSeed(seed)` itself; that pin is a probe decision, documented.
3. dTransVal is all-zero for town: `DRLG_InitTrans()` (gendung.cpp:385) memsets
   it inside `InitGlobals` and no town path writes it (town's `PlaceDunTiles`
   call at town.cpp:208 uses the 2-arg overload, which does not touch it).
   Dump it anyway as a zero-check gate.
4. Promoted to todo: scope is fully specified, no open questions.

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-tests/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Progress log

- 2026-11-06 inventory: docs/research/dg-town-inventory.md (2aea670)
- 2026-11-06 probe `town` subcommand + `dunconv town` encoder; 8 dumps, entry-invariant
- 2026-11-06 8 CSVs under internal/testdata/drlg/town + assets/town.til.csv

## Proof of Work

No Go code under test was added this session (goldens-only contract), so there
is no RED→GREEN run; the gate is the fixture set plus the repo gates.

Probe run (pinned tree, `tools/probe/seeds.txt`):

```
$ ./dvg_probe town .hermes/lane/dg-town/town-out tools/probe/seeds.txt
rc=0, 8 files, 51784 bytes each, 8 distinct sha256
```
Sanity decode of `l0-town-main-2588.bin`:
`dTransVal` non-zero cells = 0; `dSpecial` non-zero = 226; `dPiece[96][0] = 426`
and `piece[111][111] = 426` (the 16-row/16-col fill no sector overwrites);
tile layer 40x40, 1600 non-zero cells.

Encoder: `go run ./tools/probe/dunconv town .hermes/lane/dg-town/town-out/town internal/testdata/drlg/town`
→ 8 CSVs, 5 records each: `[2 12544 12544 2 1600]`.

Gates (post-commit, HEAD c90a989):

```
go build ./...                  -> BUILD OK
gofmt -l .                      -> 0 files
go vet ./internal/...           -> VET OK
./scripts/lint.sh               -> 0 issues.
go test ./... -count=1          -> 43 packages ok, 0 failures
python3 scripts/beans-lint.py   -> OK (174 beans)
just sqlc-check                 -> no raw SQL outside sqlc-generated code
```

Commits: 2aea670 (inventory), c90a989 (probe + goldens + assets + READMEs).
Pushed: origin dg-town = c90a9898b4d1bbcd138943130c261eb3453520d9.

Deviations from the dispatch, recorded:
- seed set is `tools/probe/seeds.txt`, not the illustrative `{0,1,2,77,...}`;
- 8 CSVs, not 16: the entry reaches `ViewPosition` only, and the probe asserts
  entry invariance instead of committing a duplicate file;
- no new `tools/probe/patches/*.patch`: all four town layers are exported
  globals (`dun_tile_data.hpp:66,99,103,136`), so the dump needs no game-source
  hook — the additive-dump-only rule is satisfied by adding nothing upstream.

PoW addendum: review F3: checklist reconciled with delivered corpus (8 main CSVs committed, ENTRY_PREV compared in-probe only).
