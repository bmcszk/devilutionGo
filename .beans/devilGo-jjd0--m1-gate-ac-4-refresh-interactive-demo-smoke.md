---
# devilGo-jjd0
title: 'M1 gate: AC-4 refresh — interactive demo smoke'
status: in-progress
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-23T17:34:08Z
parent: devilGo-i56j
blocked_by:
    - devilGo-5fr4
    - devilGo-njzu
---

# M1 gate: AC-4 refresh — interactive demo smoke

## Goal
Refresh docs/m1-manual-smoke.md for the interactive wave: steps 2-8
(hero create, walk, descend, kill, loot, save, load) become reachable
IN-WINDOW (today the windowed boot reads only ESC/q — smoke doc:18-26),
with fresh screenshots + pixel stats per step, and the scripted-walk
pins byte-identical (oracle untouched) recorded as the CI-side proof.

## Research
- Plan: wave-s6-plan.md s6i; the s5 wave's smoke doc and captures set
  the format (pixel stats + description + root-cause per shot).
- The interactive path lands via the device-lane + selhero + live-view
  slices; this gate closes the wave.

## Scope
- docs/** + screenshots + the gate bean; no sim/render code.

## Acceptance criteria
- [ ] Smoke doc steps 2-8 marked window-reachable with the exact
      commands; fresh per-step captures with pixel analysis in the PoW
- [ ] Scripted-walk oracle: statehash values byte-identical to the
      merged s5 pins (or the disclosed re-freeze, deterministic)
- [ ] Full gates green at HEAD through pre-commit + pre-push

## Triage
Planner wave (wave-s6-plan.md s6i), filed 2026-11-20; blocked by the
selhero and live-view slices.

## Libraries
stdlib only (docs gate).

## Skills for coder
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/beans-cli-quirks-notes/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md

## AC-4 refresh progress (coder wave on feat/m1-demo-flow, head 7d87dc0)

- Gate test TestAC4InteractiveDemoSmoke (cmd/devilutionGo/ac4_demo_gate_test.go,
  windowed tag) drives all five beats GREEN on the unchanged tree — no RED was
  available or needed: the interactive path was landed by the earlier slices,
  so the gate pins rather than fixes (recorded honestly).
- Captures + pixel stats per beat committed under docs/screenshots/m1-smoke/
  (ac4-beat1-title … ac4-beat5-walk03, PNG + .stats sidecars).
- Baseline comparison: walk01 all-pixel mean 37.74 vs baseline 37.46,
  non-black 60.39% vs 60.37%, distinct 127 vs 131 — one-tile view shift,
  deltas explained in docs/m1-manual-smoke.md.
- Sim oracle untouched: statehash-run.json sha256 =
  f1ae95a0fd732040b3e5fafb17ab23d292f0abd4402d4939ec4bc63d98fc1120 (unchanged).
- Honest gap recorded: the front flow has no enter-town transition yet, so the
  walk beat drives the lane seam directly (the windowMovement-fed seam).

## Proof of Work

- Gate test run (GREEN): AC4_DEMO_ARTIFACT_DIR run, all 5 beats pass, captures
  + stats written. Pixel stats per beat recorded in the test log and the doc.
- Trail: sha256(statehash-run.json) = f1ae95a0… — unchanged.
- Gates at HEAD: gofmt empty, go vet rc 0, lint 0 issues, go test ./... rc 0
  (windowed cmd + scene run too; pre-existing TestScreenshotCaptureWriter
  ReadPixels panic at HEAD verified on the stashed tree, not from this slice).
