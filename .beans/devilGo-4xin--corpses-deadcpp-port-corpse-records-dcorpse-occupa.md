---
# devilGo-4xin
title: 'Corpses: dead.cpp port (Corpse records, dCorpse occupancy, corpse render)'
status: completed
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-11-03T00:54:02Z
parent: devilGo-97mn
---

## Goal

Port `dead.cpp` (108 lines): Corpse placement — `Corpse` records (sprites per
monster/dir), `dCorpse` occupancy encoding `idx | dir<<5`
(levels/dun_tile_data.hpp:110-116), corpse render call in
engine/render/scrollrt.cpp:798-806.

## Research

Source/dead.cpp:1-108 (InitDeadAnimationFromMonster, AddCorpse, dCorpse
get/set); docs/research/dg-requirements.md §3 Entities "Dead/corpse" (line 224):
on monster death a Corpse record + dCorpse occupancy; players drop too; saves
serialize corpses (loadsave.cpp) — coordinate serializer fields with oqho.

## Scope

internal/diablo/entities/corpse.go (record + occupancy map) +
internal/diablo/entities (death hook from monster death path) + render hookup
point left for qbiy renderer (draw call interface, not pixels).

## Acceptance criteria

- [x] Corpse struct + dCorpse map ported with exact idx|dir<<5 encoding (round-trip test) — TestCorpses_TilePacking, TestCorpses_CorpseAtDecodesReference, TestCorpses_GobRoundTrip (internal/diablo/entities/corpse_test.go)
- [x] InitDeadAnimationFromMonster ported (sprite share, no copy) vs C++ line-order — TestCorpse_InitFromMonsterAnimSharesSprites, TestCorpse_SpritesForDirectionPicksSheetOrList, TestCorpse_SpritesForDirectionOutOfRange
- [x] AddCorpse occupancy rules (max corpses, reuse policy) parity-tested vs C++ fixture — TestCorpses_AddCorpseReusesOnlyItsOwnTile, TestCorpses_AddCorpseFixtureParity, TestCorpses_DecodeFixtureParity, TestCorpses_RecordBound
- [x] Save/serialize fields for corpses agreed with oqho (struct tags or documented layout) — TestCorpses_SaveLayoutIsDenseTileBytes + TestCorpses_JSONRoundTrip; the dense tile block is the same layout internal/persist writes via TestSaveLevelRoundTrip
- [x] No import-layer violations (entities may not import game/renderer) — TestEntitiesImportLaw

## Skills for coder

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md

## Dependencies & blockers

None hard: entities structs (r462/dufb) landed. Render hookup waits for qbiy
iso renderer but is a one-line interface call.

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
## Triage

- Brainer beans-audit of the ledger ran 2026-10-28 (devilGo-5wsb-beansaudit session; PR #3 verdict log). Per-bean brainer pass owed before implementation dispatch; wave placement, blockers and skills list set during the 2026-10-29 reorg (d083780).

## Progress log
- 2026-10-29: claimed in-progress, coder dispatched (session devilGo-4xin-coder, micro-session chain)
- 2026-10-31: dead.cpp fully ported - 13 commits on dg-corpses-deadcpp-port, PR #6 open, CI green, brainer verdict SHIP (3 minors closed by follow-up commits)


## Summary of Changes

Verified the corpse port against the C++ behavior with no source edits (gate
sweep only): tile packing (idx|dir<<5), corpse-at decode against the reference
fixture, AddCorpse occupancy/reuse, sprite sharing in
InitDeadAnimationFromMonster, the dense-tile save layout shared with
internal/persist, and the import-layer law. 28 tests in
internal/diablo/entities cover the five acceptance criteria; each criterion is
ticked with its test names above.

## Proof of Work

```
go build ./...                                   rc0
gofmt -l internal/ cmd/                          (empty)
go vet ./...                                     rc0
scripts/lint.sh ./internal/... ./cmd/...          0 issues.
go test ./internal/diablo/entities/... -count=1   ok
```

Residual: none. Corpse rendering itself is the renderer wave's job (this bean's
"corpse render" AC is DrawTileCorpse dispatch, pinned by
TestCorpses_DrawTileCorpseDispatchesDraw).
