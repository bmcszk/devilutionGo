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
