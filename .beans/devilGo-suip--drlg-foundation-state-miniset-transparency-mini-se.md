---
# devilGo-suip
title: 'DRLG foundation: state, Miniset, transparency, mini-set placement'
status: todo
type: task
priority: normal
created_at: 2026-11-04T02:35:05Z
updated_at: 2026-11-13T00:33:35Z
parent: devilGo-97mn
---

---
# devilGo-suip
title: 'DRLG foundation: state, Miniset, transparency, mini-set placement'
status: todo
type: task
priority: normal
created_at: 2026-11-04T02:35:05Z
updated_at: 2026-11-04T02:35:25Z
parent: devilGo-97mn
---

Subtask of devilGo-76tu. New package internal/diablo/levels: dungeon state arrays, Bitset2d, Miniset (match/place), DRLG transparency ops (Init/Rect/MRect/Copy/Find/Flood/Fill), PlaceMiniSet + PlaceMiniSetRandom + CanReplaceTile, PlaceDunTiles/GetDunSize. Ported 1:1 from pinned devilutionX ca91d778f (read-only reference at /tmp/dvlx-pinned). Unit tests per ported function. Scope: stdlib only, no go.mod edits.

## Contract (slice of devilGo-76tu)

Upstream reference: /tmp/dvlx-pinned (devilutionX @ ca91d778f, the fixture-producer commit).

Files (all new, package levels): internal/diablo/levels/{tile.go,state.go,miniset.go,trans.go} plus _test.go files.

Refs: utils/bitset2d.hpp; levels/gendung.h:37-72 (Miniset); levels/gendung.cpp:184-292 (IsFloor/FillTransparencyValues/FindTransparencyValues), :375-410 (DRLG_InitTrans/RectTrans/MRectTrans/CopyTrans), :459-501 (PlaceMiniSet), :503-521 (PlaceDunTiles), :576-580 (GetDunSize), :631-649 (FloodTransparencyValues); levels/drlg_l1.cpp:1252-1300 (PlaceMiniSetRandom, CanReplaceTile).

## Acceptance criteria
- [ ] Every ported function has at least one test (table-driven where multi-case)
- [ ] go build ./... and go test ./internal/diablo/levels/ -count=1 pass
- [ ] gofmt -l empty; go vet clean; lint clean

## Libraries
- stdlib only (no third-party deps at this stage)

## Triage 2026-11-05 (coder, delegated triage in lieu of a brainer pass)

This subtask was cut from devilGo-76tu after the 2026-10-16 pi-as-brainer sweep
had already run, so it never carried a triage note of its own. Scope verified on
disk before writing any claim:

- `internal/diablo/levels/{state,miniset,trans,placement,geom,bitset}.go` exist
  with the upstream refs the contract names.
- Coverage evidence: TestL1CathedralParity and TestCryptParity compare the
  dTransVal window, the placed set pieces and the view position byte for byte
  over 16 variants x 8 committed seeds; TestCanReplaceTile pins the
  PlaceMiniSetRandom screen; TestFixTransparency pins the transparency pass.
- No scope change; the acceptance criteria below are met by the landed package.

## Skills for coder
- `~/.agents/skills/coding-rules/SKILL.md` - read fully BEFORE acting
- `~/.agents/skills/ponytail/SKILL.md` - read fully BEFORE acting
- /home/blaze/.agents/skills/go-tests/SKILL.md - fixture matrix
- /home/blaze/.agents/skills/go-dev/SKILL.md - RNG-call-order discipline

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Program-surface law (ADR-020/021 + one-tool law): no new binaries. Fixture
   instruments are subcommands of tools/goldens; user-facing surface stays
   exactly cmd/devilutionGo + cmd/devilutionGo-migrate. No `func main` outside
   the allowlist.
2. Tooling problems: fix immediately, no hacks.
