---
# devilGo-hw9y
title: 'First scene: Diablo 1 town in the ebiten window with the hero standing'
status: in-progress
type: epic
priority: critical
created_at: 2026-11-14T12:33:20Z
updated_at: 2026-11-14T12:36:42Z
---

## Goal

Launch devilutionGo and SEE the first Diablo 1 scene: the town rendered from
real retail archives, the warrior sprite standing in it, in a real ebiten
window. "Working" = `go run ./cmd/devilutionGo` opens a window showing town
+ hero; a screenshot command captures it without a human at the keyboard.

## Why now (user directive 2026-11-14, verbatim intent)

"Prioritize getting the first scene of Diablo 1. A real character standing
in the village, all generated in the new ebiten engine. Plan the beans and
go!" — supersedes M1 breadth until the scene is on screen.

## Verified seams (parent triage 2026-11-14, file:line)

- DRLG: internal/diablo/levels/town.go:81 CreateTown is parity-proven
  (town_parity_test.go against the committed drlg corpus) — produces the
  dPiece/dSpecial grids for the whole town.
- Rasterizer: internal/render/tile.go:83 RenderTileFrame ports the pinned
  dun_render.cpp (unclipped paths + bounds guard are in); reencode_dun_cels.go
  ports the CEL re-encode; Presenter (presenter.go:57) is the ebiten pipeline
  (index surface -> shader -> target) with Frame()/Draw().
- Archives: internal/assets OpenMPQArchive (mpqarchive.go:109) +
  mpq_smoke_test.go prove full retail decode via DG_ASSET_ARCHIVES
  (DIABDAT.MPQ present in this environment).
- Window stub: cmd/devilutionGo/main.go:74 runWindowed currently logs and
  returns — the boot seam to replace.
- Hero: real plrgfx (PL2/CL2 frames per class+armor) decode via
  assets DecodeCL2/DecodeCEL; the warrior idle (STAND) set is the target.
- Screenshot: Presenter.Target() is an *ebiten.Image — an ebiten screenshot
  key or a Frame()-dump writes PNG headless-with-display (DISPLAY=:0
  available; paritytest.Capture already proves the path).

## Missing glue (the beans below)

Town mega-tile -> pMegaTiles/SOL/CEL loading from archives (LoadSOLData
exists; the archive->level-cel block loader does not), the view/camera
driving DrawWorld over the real grid, the runWindowed replacement wiring
Game ticks + Presenter + input bridge, plrgfx selection (warrior STAND
frames) + draw-over-world, window PNG capture.

## Children (one grouped PR, review round at the end)

1. Town from real MPQ archives rendered through the ported rasterizer — 6asa
2. Windowed game boot: Presenter + ebiten RunGame replaces the runWindowed stub — v9xy
3. Warrior standing in town: real plrgfx idle frame drawn over the world — s7x0
4. Live window screenshot: Presenter target to PNG each N frames — 1awt (high)

## Acceptance criteria (epic)

- [ ] `go run ./cmd/devilutionGo` opens an ebiten window showing the town rendered from retail archives
- [ ] The warrior STAND frame stands in town, drawn over the world through Presenter
- [ ] A screenshot command/flag writes a PNG of the live window; parent verifies visually
- [ ] No regressions: full gates green, plain-clone oracle run recorded
- [ ] All four child tasks completed

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- `~/.agents/skills/go-dev/SKILL.md`
- `~/.agents/skills/go-tests/SKILL.md`
- /home/blaze/.agents/skills/ebiten/SKILL.md + rendering/SKILL.md — ebiten law
- `~/.agents/skills/ponytail/SKILL.md`

## Progress log

Claimed 2026-11-14 (parent planning, user directive): seams verified (CreateTown parity-proven town.go:81, RenderTileFrame tile.go:83, Presenter presenter.go:57, OpenMPQArchive mpqarchive.go:109, runWindowed stub main.go:74); children 6asa/v9xy/s7x0/1awt filed; grouped-PR law.
