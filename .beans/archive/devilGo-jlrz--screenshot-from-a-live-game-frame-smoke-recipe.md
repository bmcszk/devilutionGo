---
# devilGo-jlrz
title: Screenshot from a live game frame + smoke recipe
status: scrapped
type: task
priority: critical
created_at: 2026-11-14T11:13:02Z
updated_at: 2026-11-14T11:16:58Z
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
