---
# devilGo-s7x0
title: 'Warrior standing in town: real plrgfx idle frame drawn over the world'
status: completed
type: task
priority: critical
created_at: 2026-11-14T12:33:20Z
updated_at: 2026-11-14T23:46:41Z
parent: devilGo-hw9y
---

## Goal

The warrior, standing, in the town scene: load the real plrgfx warrior
STAND (idle) frames from the archives, pick frame 0, and draw it over the
world at the hero's tile through the Presenter pipeline — the "real
character standing in the village" the user asked for.

## Research (parent, verified 2026-11-14)

- plrgfx layout at the pin: plrgfx/common.cel + per-class files; the
  warrior STAND set for town armor is one of the plrgfx archives
  (beapst/CL2-family — verify exact record names against the pin's
  plr.cpp/plr gfx tables and the archive listfile; cite in PoW).
- Decode path exists: assets DecodeCL2/DecodeCEL produce sprite lists;
  Presenter draws Surface content through the palette+shader pipeline —
  the hero frame must go through the same palette mapping as the world
  (no independent RGB side-channel).
- Direction/animation discipline: STAND has 8 directions x N frames; this
  bean pins direction = the caller's facing (default S) and frame = 0
  only — animation cycling is a later bean.
- The hero's tile comes from the smoke lane's world state (world.State
  Hero()) — draw at WorldToScreen(heroTile) with the standard sprite
  anchor (feet at tile center-bottom, upstream's offset rule).

## Acceptance criteria

- [ ] Warrior STAND frame 0 (one direction) decodes from the real archives (skip-if-absent guard)
- [ ] Drawn into the composed town scene at the hero tile with upstream's sprite anchor; headless test asserts hero pixels present over the world at the expected screen rect
- [ ] Palette-consistent: hero pixels go through the same Presenter palette mapping (no raw RGB bypass)
- [ ] PNG artifact shows the figure visibly standing in town (parent visual check)
- [ ] TDD red -> green; gates green

## Libraries

- existing internal deps (assets/render) + stdlib; no new third-party

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md
- `~/.agents/skills/go-tests/SKILL.md`

- `~/.agents/skills/ponytail/SKILL.md`

## Triage

Parent mini-triage 2026-11-14 (session _chat, user priority directive): seams verified in-tree at file:line (see Research); scope split = town render inputs (6asa), window boot (v9xy), hero sprite (s7x0), live capture (1awt); grouped PR, brainer re-review rides the PR review round per lane law. Skills paths stat-verified.


## Summary of Changes

**RED RECTANGLE (solid red hero) — FIXED:**
- WarriorStandFrame0 now calls  on  (168562 bytes, 8 groups × 20 frames = 160 frames; STAND frames are 96×96 — 96×128 overruns the row walk)
- Takes group 0 frame 0 (South direction per upstream LoadPlrGFX: szCel="st", classChar='w', ArmourChar='l', WepChar='n')
- Anchor corrected: clipX = 320 - heroW/2, clipY = 240 - heroH + TileHeight/2 (matches scrollrt.cpp DrawPlayer)
- DrawSpriteKeyed (keyed blit, index 0 = hole) draws hero over town surface

**DIM TOWN (too dark) — FIXED:**
- Ported brightness tone-map from devilutionX ApplyGlobalBrightness (palette.cpp:96-110): a = -(brightness/100)*2.0; f(x) = a*x² + (1-a)*x over x = c/255; out = uint8(clamp(f,0,1)*255 + 0.5)
- PaletteState.WithBrightness(percent int) applies tone-map to palette then re-derives 16 light variants
- Brightness 0 = byte-identity (parity oracle stays green)
- First-scene boot uses brightness 60 (readable; devilutionX players run gamma 30-60)

**CL2 SHEET REGRESSION TEST:**
- groupRanges fixed to use header-relative offsets per cl2_to_clx.cpp GetGroupRanges (last pointer = group's frame-data END; plrgfx sheets lay all headers first, so slicing to next group's header truncates)
- Pinned by cl2ExpectGroupHeaderRelativeRanges test

**ACCEPTANCE TESTS ALL PASS:**
- TestWarriorStandFrameOverTown: hero pixels at hero tile + brightness gate (non-black mean ≥ 55 at brightness 60; without tone-map was 30.6)
- Headless scene test + windowed visual oracle both pass
- Screenshot artifact: /tmp/scene-shots/town_hero_fixed.png shows warrior in readable town

**GATES GREEN:** go build, gofmt, go vet, scripts/lint.sh, go test ./... (with archives env)


## Summary of Changes

**RED RECTANGLE (solid red hero) — FIXED:**
- WarriorStandFrame0 now calls assets.DecodeCL2 on plrgfx/warrior/wln/wlnst.cl2 (168562 bytes, 8 groups x 20 frames = 160 frames; STAND frames are 96x96)
- Takes group 0 frame 0 (South direction per upstream LoadPlrGFX: szCel=st, classChar=w, ArmourChar=l, WepChar=n)
- Anchor corrected: clipX = 320 - heroW/2, clipY = 240 - heroH + TileHeight/2
- DrawSpriteKeyed (keyed blit, index 0 = hole) draws hero over town surface

**DIM TOWN (too dark) — FIXED:**
- Ported brightness tone-map from devilutionX ApplyGlobalBrightness: a = -(brightness/100)*2.0; f(x) = a*x^2 + (1-a)*x
- PaletteState.WithBrightness applies tone-map then re-derives 16 light variants
- Brightness 0 = byte-identity (parity oracle stays green)
- First-scene boot uses brightness 60 (readable; devilutionX players run gamma 30-60)

**CL2 SHEET REGRESSION TEST:**
- groupRanges fixed to use header-relative offsets per cl2_to_clx.cpp GetGroupRanges
- Pinned by cl2ExpectGroupHeaderRelativeRanges test

**ACCEPTANCE TESTS ALL PASS:**
- TestWarriorStandFrameOverTown: hero pixels at hero tile + brightness gate (non-black mean >= 55 at brightness 60; without tone-map was 30.6)
- Headless scene test + windowed visual oracle both pass
- Screenshot artifact: /tmp/scene-shots/town_hero_fixed.png shows warrior in readable town

**GATES GREEN:** go build, gofmt, go vet, scripts/lint.sh, go test ./... (with archives env)

## Proof of Work

**Files implemented/modified:**
- `internal/present/scene/warrior.go` — `WarriorStandFrame0` decodes real CL2 from `plrgfx/warrior/wln/wlnst.cl2` (168562 bytes, 8 groups × 20 frames = 160 frames; STAND frames 96×96) via `assets.DecodeCL2(data, 96, 96)`. Anchor: clipX = 320 - w/2, clipY = 240 - h + TileHeight/2. Drawn via `Surface.DrawSpriteKeyed` (skips index 0).
- `internal/assets/cl2.go` — Fixed `groupRanges` to use header-relative offsets per `cl2_to_clx.cpp GetGroupRanges` (last pointer = group's frame-data END).
- `internal/render/lightvariants.go` — `PaletteState.WithBrightness(percent int)` ports devilutionX `ApplyGlobalBrightness` (palette.cpp:96-110): `a = -(b/100)*2.0; f(x)=a·x²+(1-a)·x` over `x=c/255`; brightness 0 = byte-identity.
- `internal/present/scene/warrior_test.go` — `TestWarriorStandFrameOverTown` asserts hero pixels at hero tile + brightness gate (non-black mean ≥ 55 at brightness 60).

**Acceptance verified:**
- Warrior STAND frame 0 (South direction) decodes from real archives (skip-if-absent)
- Drawn at hero tile with upstream sprite anchor; headless test asserts hero pixels present over world at expected screen rect
- Palette-consistent: hero pixels go through same Presenter palette mapping (no raw RGB bypass)
- PNG artifact shows figure standing in town: `/tmp/scene-shots/town_hero_fixed.png`
- Brightness gate: non-black mean 62.04 at brightness 60 (was 30.6 without tone-map)
- Warrior upright: head region mean 63.50 > feet region mean 38.33 (y grows down)

**Gates:** go build ✅, gofmt ✅, go vet ✅, scripts/lint.sh ✅, go test ./... ✅ (51 packages)
