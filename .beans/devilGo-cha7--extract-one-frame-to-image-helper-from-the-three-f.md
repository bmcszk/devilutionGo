---
# devilGo-cha7
title: Extract one frame-to-image helper from the three frameToImage/frameImage copies
status: completed
type: task
priority: normal
created_at: 2026-11-26T12:15:57Z
updated_at: 2026-11-26T18:12:46Z
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
- [x] `scene.FrameToImage` exists (exported, documented 1-2 lines); grep shows NO remaining frame->image loops outside it: `frameToImage|frameImage` hits only the exported function and history (paste grep)
- [x] All three former call sites (town_compose WritePNG path, ac4_demo_gate_test, screenshot_test) call `scene.FrameToImage`; per-file net line deltas reported in the PoW
- [x] Refactor gate green at head: `gofmt -l .` empty, builds both tags rc0, vet both tags rc0, `scripts/lint.sh ./...` 0 issues, `go test -count=1 ./...` ok, windowed pair ok, clean-clone oracle rc0 - tails pasted (behavior unchanged: same test counts as pre-wave head)

## Libraries
stdlib (image, image/color) + repo UI stack.

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md (duplication + refactor-gate laws govern this bean)
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Scoping DONE by the gate-parity review round 1 (task-gp-review-brainer, NB-1 finding with the three file:line cites; the reviewer is the triage-designated research subagent for this finding - re-deriving it would duplicate completed research). Parent verified all three cites on the tree during the round-1 verification. Home decision (export from scene) recorded above. PROMOTE-TODO.



## Proof of Work

- One commit 59bdb83 `refactor(scene): extract FrameToImage from the three frame-to-image copies`; 3 files, +17/-49 (net -32): town_compose.go +1 (the exported helper), ac4_demo_gate_test.go -16, screenshot_test.go -17.
- Stray check: `grep -rn "frameToImage|frameImage"` over internal/ + cmd/ (excluding FrameToImage) = zero hits (pasted in the wave log).
- Behavior parity (mechanical): `go test -count=1 -json` on both heads - 16320 PASS / 23 skip / 0 fail / 54 ok-packages IDENTICAL at 1878f02 and 59bdb83; windowed pair 114 PASS both heads. Refactor gate law satisfied with numbers, not vibes.
- Review (task-fimg-review-brainer): **APPROVE, no blocking findings** - semantics verified per-pixel (all three deleted loops were the same NRGBA SetNRGBA shape; nrgbaOf is a pure same-width field copy), PNG byte-transparency traced through the stdlib png writer (dynamic *image.NRGBA fast path, no per-pixel conversion), API name/doc/placement per the file-layout and funcorder laws.
- Parent gate re-run at 59bdb83: gofmt empty; builds + vet both tags rc0; lint 0 issues; untagged suite 54 ok; windowed pair ok (env -u form); beans-lint OK (238); clean-clone oracle rc0 (50 internal pkgs).
- Residual: none. The deleted cmd-side helper's "unexported there" comment died with the loop, as it should.
