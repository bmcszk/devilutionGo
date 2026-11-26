---
# devilGo-cha7
title: Extract one frame-to-image helper from the three frameToImage/frameImage copies
status: todo
type: task
priority: normal
created_at: 2026-11-26T12:15:57Z
updated_at: 2026-11-26T17:59:33Z
parent: devilGo-qbiy
---

# Extract scene.FrameToImage from the three frame-to-image copies

## Goal
One frame-to-image conversion exists instead of three: export `scene.FrameToImage(frame render.Frame) image.Image` and collapse the three copies onto it. Pure consolidation - behavior must not change.

## Research (triage source: gate-parity review round 1, NB-1 - the scoping subagent already did this research with file:line cites)
- Copy 1: internal/present/scene/town_compose.go:223-233 (`frameToImage`, unexported) - feeds WritePNG's PNG artifacts.
- Copy 2: cmd/devilutionGo/ac4_demo_gate_test.go:252-262 (`frameToImage`) - carries a comment saying the same "unexported there" thing.
- Copy 3: internal/present/scene/screenshot_test.go:78-91 (`frameImage`, added by the 7bij fix) - the NRGBA loop over frame.Pixels.
- Home decision: export from scene (smallest diff, no new package; scene already exports the frame-facing helpers DistinctColors/MeanBrightness/WritePNG/ComposeTownFrame). A internal/testsupport/img package was the alternative - rejected as more surface for the same collapse.
- Why not in the gate-parity wave: the extraction needs an export from scene - outside that wave's scope jail (reviewer's own words); filed as its own bean.

## Scope
- Add `func FrameToImage(frame render.Frame) image.Image` in internal/present/scene (next to the existing frame helpers), with the one-line what/why doc.
- Rewire all three copies to call it; delete the three local loops (net line count reported per file in the PoW).
- No test edits beyond the deleted helpers' call sites; existing tests bound the design (refactor gate: build, lint, FULL suite green before push; report per-file net line deltas).

## Acceptance criteria
- [ ] `scene.FrameToImage` exists (exported, documented 1-2 lines); grep shows NO remaining frame->image loops outside it: `frameToImage|frameImage` hits only the exported function and history (paste grep)
- [ ] All three former call sites (town_compose WritePNG path, ac4_demo_gate_test, screenshot_test) call `scene.FrameToImage`; per-file net line deltas reported in the PoW
- [ ] Refactor gate green at head: `gofmt -l .` empty, builds both tags rc0, vet both tags rc0, `scripts/lint.sh ./...` 0 issues, `go test -count=1 ./...` ok, windowed pair ok, clean-clone oracle rc0 - tails pasted (behavior unchanged: same test counts as pre-wave head)

## Libraries
stdlib (image, image/color) + repo UI stack.

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md (duplication + refactor-gate laws govern this bean)
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Scoping DONE by the gate-parity review round 1 (task-gp-review-brainer, NB-1 finding with the three file:line cites; the reviewer is the triage-designated research subagent for this finding - re-deriving it would duplicate completed research). Parent verified all three cites on the tree during the round-1 verification. Home decision (export from scene) recorded above. PROMOTE-TODO.
