---
# devilGo-jlrz
title: Screenshot from a live game frame + smoke recipe
status: in-progress
type: task
priority: critical
created_at: 2026-11-14T11:13:02Z
updated_at: 2026-11-14T12:29:22Z
parent: devilGo-wq2s
---

## Goal

From the same running-game e2e: compose one live frame (world -> index
surface) and write it out as a PNG artifact, plus a `just smoke` recipe that
runs the whole loop (boot -> input -> logs -> screenshot) with one command.

## Research (parent, 2026-11-14, verified)

- internal/render/index_blit.go:53 NewSurface allocates the index buffer
  headless — no GPU needed. The palette path (palette rows + TRN) is
  committed already (paritytest fixtures: l1_1.pal.hex etc.) — expand index
  -> RGB with the palette and encode PNG with stdlib image/png.
- paritytest.Capture (capture.go:84) is the GPU/DISPLAY=:0 path — NOT
  required here; this bean is the cheap headless artifact. GPU parity
  capture stays with the harness bean (devilGo-7aul lineage).
- just recipes live in the justfile; lint runs scripts/lint.sh.

## Acceptance criteria

- [ ] The e2e writes <artifact dir>/frame.png: a composed frame of the world after the injected move (hero visible or at least world tiles rendered through the committed palette)
- [ ] PNG is valid (decoded back in the test; dimensions match the surface)
- [ ] `just smoke` runs boot -> ticks -> injected click -> logs -> PNG and exits 0; recipe documented in the justfile
- [ ] Artifact path printed by the recipe output
- [ ] TDD red -> green recorded

## Skills for coder

- `~/.agents/skills/coding-rules/SKILL.md` — read fully before editing
- `~/.agents/skills/go-tests/SKILL.md`
- `~/.agents/skills/go-dev/SKILL.md`

- `~/.agents/skills/ponytail/SKILL.md` — smallest change that fully solves the task

## Libraries

- stdlib only (slog, image/png, testing); no new third-party deps

## Triage

Parent mini-triage 2026-11-14 (session _chat, planning mode per user directive): seams verified against the tree at file:line (see Goal/Research sections); scope split = boot/logs (u1hd), injection (judg), screenshot+recipe (jlrz); brainer re-review rides the PR review round per lane law (read-first verdict + operator checks). Skills paths stat-verified on disk.

## Progress log — smoke slice (2026-11-14)

- RED: `h.WriteFramePNG` undefined (vet failure).
- GREEN (commit 932ac0d): WriteFramePNG delegates to
  testsupport.ComposeFramePNG — the composer lives in
  internal/game/testsupport (sim purity gate class: file io + rendering
  stay out of the simulation sources; lint.sh simpurity now reports
  "internal/game internal/diablo clean (249 sources)"). The composer
  paints a 20x15-tile 64px viewport into render.NewSurface (floor 0xC0,
  hero 0x40), expands through the committed
  internal/testdata/render/pal/l1_1.pal.hex (missing fixture fails
  loudly, no gray-ramp fallback), encodes PNG via stdlib image/png into
  SMOKE_ARTIFACT_DIR, echoes the path to stderr.
- TestSmokeFramePNGArtifact decodes the PNG back: dimensions 1280x960,
  path under SMOKE_ARTIFACT_DIR, floor/hero colors distinct. Verified
  artifact: `file frame.png` -> "PNG image data, 1280 x 960, 8-bit/color
  RGB".
- `just smoke` recipe: SMOKE_ARTIFACT_DIR (default $TMPDIR/
  devilutionGo-smoke) -> go test ./internal/game -run 'TestSmoke' -v ->
  prints "smoke artifact: <path>/frame.png". Runs exit 0.
- ACs satisfied: frame.png after injected move through the committed
  palette; PNG valid (decoded back, dimensions match); just smoke exits
  0 and is documented in the justfile; artifact path printed; TDD red ->
  green recorded.

## s2 upgrade — the frame proves the input->movement

The artifact is now a causal-chain proof, not a flat screenshot: two
frames per smoke run (frame-1-before.png / frame-2-after.png). The user
sees where the hero started, where the click landed (target cell in a
third distinct shade), where the hero ended; the floor is readable
(walkable / blocked / 1px grid lines) and the camera clamps to world
bounds so the hero always renders fully inside the frame. Camera offset
is exported (testsupport.CameraOffset) so tests map world tiles to
frame pixels. just smoke prints both artifact paths.

Units: 2e4565e (Y1 camera clamp, RED: hero absent at edge spawns),
8ecee2f (Y2 readable floor, RED: flat color), 37df290 (Y3 story
frames, RED: second file missing). PR body 'What you see' section
added (5705 chars).
