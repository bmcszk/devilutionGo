---
# devilGo-6z41
title: 'Town DRLG: port CreateTown to parity'
status: todo
type: task
priority: normal
created_at: 2026-11-06T00:04:32Z
updated_at: 2026-11-06T00:04:40Z
parent: devilGo-97mn
blocking:
    - devilGo-yeia
---

Subtask of the DRLG family. Red gate landed by the dg-town s2 lane: the parity
test exists and fails against HEAD (RED is the contract for this bean).

## RED state at branch dg-town

- `internal/diablo/levels/town_parity_test.go` — `TestTownParity` (per-seed
  matrix) + `TestTownCorpus` (corpus size), diffing `dPiece`, `dSpecial` and the
  `dungeon` automap tile layer of the 8 committed
  `internal/testdata/drlg/town/l0-town-main-<seed>.csv` fixtures.
- `internal/diablo/levels/town.go` — WIP plumbing only: `TownSector1..4` /
  `TownAutomap` resource paths, `TownOptions`, and a `CreateTown` stub that
  returns a zeroed `PieceGrid`.
- Observed failure (all 8 seeds, from the committed corpus):
  `dPiece[0][0] = 0, want 217`.
- Second RED finding: the five `.dun` layers `CreateTown` loads
  (`town.cpp:202-207`) are **not committed**. `TestTownParity` reports
  `inputs-committed` missing fixtures:
  `internal/testdata/drlg/town/inputs/{sector1s,sector2s,sector3s,sector4s,automap}.csv`.

## Contract

Upstream (pinned `ca91d778f`, read-only reference): `Source/levels/town.cpp`
`:191-246` (`DrlgTPass3`: 426 fill, four `FillSector` at fixed origins,
`PlaceDunTiles(automap.dun)`, closed-warp patches incl. the ten
`PickRandomlyAmong({1,2,3,4})` draws, poisoned-water fountain), `:359-387`
(`CreateTown`: view bounds + entry `ViewPosition`), `:149-186`
(`InitTownPieces`: `dPiece` micro id -> `dSpecial` code). Surface map:
`docs/research/dg-town-inventory.md`.

Entry point is already fixed by the gate: `levels.CreateTown(d *Dungeon, opts
TownOptions) PieceGrid` — it writes `d.Tiles` (`dungeon`) and returns
`PieceGrid{Piece, Special}`. Inputs are injected (`TownOptions.Files`,
`MegaTiles`, `Warps`, `Hellfire`, `Quests`; RNG is `d.Rnd`, seeded by the caller).

## Acceptance criteria

- [ ] Commit the five `.dun` inputs as setpieces-format CSVs (`width,height`
  then `width*height` uint16 tile values) under
  `internal/testdata/drlg/town/inputs/`, with a provenance note (retail
  `DIABDAT.MPQ` is not a legal test input — AGENTS.md zero-dependency rule)
- [ ] Port `DrlgTPass3` + `CreateTown` + `InitTownPieces` 1:1; every ported
  function has at least one test
- [ ] `TestTownParity` green on all 8 seeds for `dPiece`, `dSpecial` and
  `dungeon`; `dTransVal` stays all zero
- [ ] gofmt -l empty, go vet clean, `./scripts/lint.sh` 0 issues,
  `go test ./internal/... -count=1` all ok

## Libraries

- stdlib only (no third-party deps at this stage)

## Triage 2026-11-06 (dg-town s2, disk-verified, delegated triage in lieu of a brainer pass)

Census before any claim (this session's brainer step ran disk-first, per the
wave recipe, because the town surface was already researched in
docs/research/dg-town-inventory.md):

- `grep -rn "func CreateTown" internal/` finds only the WIP stub in
  `internal/diablo/levels/town.go` — no town generator exists on disk.
- The corpus is 8 files x 5 records (`awk 'END{print NR}'` = 5; row 2/3 are
  12544 cells, row 5 is 1600), matching the inventory's array shapes.
- `ls internal/testdata/drlg/town/inputs` fails: the `.dun` input corpus is
  absent, so the gate currently fails before it can diff a real generator.
- `internal/testdata/drlg/assets/town.til.csv` is committed (376 mega tiles) and
  is the `pMegaTiles` table the gate already loads — the only town input that
  exists today.
- Scope is one generator, no retry loop, one RNG consumer: the smallest DRLG
  slice in the family.

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/go-tests/SKILL.md` — read fully BEFORE acting
