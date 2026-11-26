---
# devilGo-7bij
title: 'scene: ScreenshotCaptureWriter panics (ReadPixels before game start) under windowed tests with assets'
status: todo
type: bug
priority: normal
created_at: 2026-11-25T13:04:19Z
updated_at: 2026-11-26T11:51:13Z
parent: devilGo-p1t8
---

# TestScreenshotCaptureWriter panics under `-tags windowed` with assets: ReadPixels before game start

## Goal
Stop `TestScreenshotCaptureWriter` (internal/present/scene/screenshot_test.go:40) from panicking
`ui: ReadPixels cannot be called before the game starts` when the windowed-tagged suite runs with
DG_ASSET_ARCHIVES present. The panic kills the whole scene package run, masking any later test
results in that package.

## Research
- Found 2026-11-25 during the hl1 wave's parent gate re-run (not wave-introduced); reproduces on master 770df80 and on 408f8f7.
- Root cause (triage-verified): the test calls `presenter.Target().ReadPixels` at screenshot_test.go:40 after a `presenter.Draw`; ebiten v2.10.6 sets `running` only inside `ebiten.RunGame` (internal/ui/ui.go:164-167 guards ReadPixels), and the `//go:build windowed` tag merely links ebiten — it never starts a loop, so the tag cannot satisfy the guard.
- The test does not test the real capture writer: the production path (cmd/devilutionGo/window_scene.go:187-210) reads the screen via `screen.At(x, y)` inside draw(), never `Target().ReadPixels`. The ReadPixels leg covers a path production never takes.
- A game-loop harness would be the larger, worse diff: DISPLAY-gated (skips on CI and headless boxes, so a weaker gate than the headless version) and it collides with the one-loop-per-process law (internal/render/paritytest/capture.go:62-64).
- Origin of the bug: ~/.agents/skills/ebiten/rendering/SKILL.md:41,49 says the windowed tag is what ReadPixels tests need — necessary but NOT sufficient. The correct, tighter law lives in ~/.hermes/skills/dev/devilutiongo-lane-ops/references/visual-verification.md:34-38. The skill-text defect is filed as its own draft; NOT fixed under this bean.

## Scope
Re-scope the test to the headless-safe capture seam (triage decision (a)):
- Keep the asset-gated town compose; drop `NewPresenter` + `Draw` + `ReadPixels`.
- Feed the PNG encode→decode→bounds round-trip from the headless expansion `composed.Surface.ExpandTile(&composed.State.Variants, 0)` (the same oracle WritePNG uses, town_compose.go:32-47; index_blit.go:111).
- Assert the PNG decodes to non-flat content (DistinctColors/MeanBrightness) so the test still proves a frame was written.
- Rewrite the test's comment to cite the ReadPixels-needs-RunGame law (ebiten ui.go:164-167) and the lane-ops visual-verification reference.
- Do NOT skip-with-assets: skipping would delete coverage to silence the panic.

## Acceptance criteria
- [ ] `internal/present/scene/screenshot_test.go` no longer calls `ReadPixels`/`NewPresenter`; it composes the real town frame (asset-gated skip only when `DG_ASSET_ARCHIVES` is unset), expands it headlessly via `Surface.ExpandTile`, and asserts the PNG encodes, decodes, has matching bounds and non-flat content; its comment cites the ReadPixels-needs-RunGame law (ebiten ui.go:164-167) and the lane-ops visual-verification reference
- [ ] `DG_ASSET_ARCHIVES=<real archives> go test -count=1 -tags windowed ./internal/present/scene` passes with no panic (RED tail = the current panic captured before the edit; GREEN tail after); the full `-tags windowed` suite (`./cmd/devilutionGo` + `./internal/present/scene`) stays green
- [ ] Gates green (`gofmt -l` empty, `go vet` rc0, `golangci-lint` 0, `go test ./internal/...` ok); the `ebiten/rendering/SKILL.md:41,49` misguidance draft exists and is referenced from the PoW (not fixed under this bean)

## Libraries
stdlib + ebiten v2.10.6.

## Skills for coder
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.agents/skills/ebiten/rendering/SKILL.md — CONTEXT ONLY: its lines 41 and 49 are the misguidance that produced this bug; the coder follows the lane-ops reference, not those two lines
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Filed 2026-11-25 by the parent session from live reproduction at master 770df80 (evidence above). Deep triage DONE 2026-11-26 (task-m2fu-triage-brainer; parent spot-checked the capture path, the ExpandTile oracle, and the skill text on the tree). Verdict: PROMOTE-TODO — headless ExpandTile→PNG seam, loop harness rejected. Full verdict: .hermes/lane/m2-followups/triage-result.md
