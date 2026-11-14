---
# devilGo-6asa
title: Town from real MPQ archives rendered through the ported rasterizer
status: in-progress
type: task
priority: critical
created_at: 2026-11-14T12:33:20Z
updated_at: 2026-11-14T13:23:56Z
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

- [x] Loader reads town.pal/mega tiles + town.cel (+ SOL) from the mounted retail archives (skip-if-absent pattern)
- [x] An adapter feeds CreateTown's grids + real CEL bytes through DrawWorld/RenderTileFrame into a render Surface
- [x] Headless test composes a town frame: decodes back, asserts multiple distinct palette colors (not flat), tile-diamond structure present
- [x] The frame is written as PNG artifact for the parent's visual check
- [x] TDD red -> green recorded; gates green

## Progress log — first scene slice (2026-11-14)

- RED: internal/diablo/scene test written first; build failure proves red; suite landed at internal/present/scene (moved: scripts/simpurity bans render imports under internal/diablo, lint.sh:94; precedent internal/game/ui ad1eea4; presentation-edge home logged in the lane log).
- GREEN: TestTownSceneComposesFromRetailArchives passes with DG_ASSET_ARCHIVES=/home/blaze/Games/Heroic/Diablo:/home/blaze/Games/Heroic/Diablo/hellfire; 100 distinct palette colours in the composed 640x480 frame; PNG artifact /tmp/scene-artifacts/town.png.
- Pin/archive cites: levels\towndata\town.{cel,til,min,sol,pal} + sector1-4s.dun + automap.dun in DIABDAT.MPQ (names inventory internal/assets/names/diabdat.txt lines 369-372, 1179-1184); hellfire nlevels town.{cel,min} preferred (diablo.cpp:1374-1383, dun_tile_data.cpp:72-77) — nlevels cel carries 3803 frames vs diabdat 3547, the MIN's max frame.
- Deviation logged: townWorldView light layer is the InitLight town default (0 = fully lit, gendung.cpp:301-307); the dynamic light list is a later lane.
- Gates: lint.sh ./internal/present/scene/... rc0; go vet ./... rc0; gofmt -l clean; go test ./internal/render ./internal/diablo/levels ok.
