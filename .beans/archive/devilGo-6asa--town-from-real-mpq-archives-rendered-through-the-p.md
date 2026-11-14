---
# devilGo-6asa
title: Town from real MPQ archives rendered through the ported rasterizer
status: completed
type: task
priority: critical
created_at: 2026-11-14T12:33:20Z
updated_at: 2026-11-14T23:46:38Z
parent: devilGo-hw9y
---

## Goal

Load the town's render inputs from the real retail archives and draw the
scene through RenderTileFrame into a Presenter index surface: the first
REAL frame, headless-assertable (no window yet — the bytes matter first).

## Research (parent, verified 2026-11-14)

- Inputs per upstream town flow: pMegaTiles (levels/towndata/towns.tfi
  equivalent), SOL data (LoadSOLData already ports tile_properties.cpp),
  the town CEL blocks (levels/towndata/town.cel re-encoded through the
  committed reencode_dun_cels.go path), palette town.pal (pal assets exist).
- CreateTown (town.go:81) + its parity corpus give the dPiece/dSpecial
  grids; DrawWorld/DrawFloor (world.go:190/198) consume a World interface —
  an adapter from PieceGrid+dPiece to render.World is the glue.
- Archive read proven: assets.OpenMPQArchive + findAssetArchives pattern
  (mpq_smoke_test.go); DG_ASSET_ARCHIVES reachable in this environment.
- Acceptance is BYTE-level first: compose the frame headless (index
  surface), decode, assert non-uniform content + palette colors present;
  visual check rides the window task.

## Acceptance criteria

- [ ] Loader reads towns.tfi/mega tiles + town.pal + town.cel (+ SOL) from the mounted retail archives (skip-if-absent pattern like mpq_smoke_test)
- [ ] An adapter feeds CreateTown's grids + real CEL bytes through DrawWorld/RenderTileFrame into a render Surface
- [ ] Headless test composes a town frame: decodes back, asserts multiple distinct palette colors (not flat), tile-diamond structure present
- [ ] The frame is written as PNG artifact for the parent's visual check
- [ ] TDD red -> green recorded; gates green

## Libraries

- stdlib + existing internal deps only (image/png, assets, render)

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md
- `~/.agents/skills/go-tests/SKILL.md`

## Triage

Parent mini-triage 2026-11-14 (session _chat, user priority directive): seams verified in-tree at file:line (see Research); scope split = town render inputs (6asa), window boot (v9xy), hero sprite (s7x0), live capture (1awt); grouped PR, brainer re-review rides the PR review round per lane law. Skills paths stat-verified.

## Proof of Work

**Files implemented/modified:**
- `internal/present/scene/town.go` — `ComposeTownFrame` loads town.pal, town.cel, town.til, town.min, town.sol + sector .dun files from retail MPQs; builds grid via `levels.CreateTown`; renders through `render.DrawWorld` into 640×480 index surface
- `internal/present/scene/town_compose.go` — `WritePNG`, `MeanBrightness`, `DistinctColors`, `HasDiamondEdge` helpers
- `internal/present/scene/town_test.go` — `TestTownSceneComposesFromRetailArchives` acceptance test (skip-if-absent on `DG_ASSET_ARCHIVES`)

**Acceptance verified:**
- Loader reads all town records from retail archives (DIABDAT.MPQ + devilutionx.mpq)
- Adapter feeds CreateTown grids + CEL bytes through DrawWorld/RenderTileFrame → render.Surface
- Headless test composes frame, decodes, asserts ≥64 distinct palette colors (got 100), tile-diamond edge present
- PNG artifact written to `DG_SCENE_ARTIFACT_DIR/town.png`
- Test passes: `DG_ASSET_ARCHIVES=/home/blaze/Games/Heroic/Diablo:/home/blaze/Games/Heroic/Diablo/hellfire go test ./internal/present/scene -run TestTownSceneComposesFromRetailArchives -v`

**Gates:** go build ✅, gofmt ✅, go vet ✅, scripts/lint.sh ✅, go test ./... ✅ (51 packages)
