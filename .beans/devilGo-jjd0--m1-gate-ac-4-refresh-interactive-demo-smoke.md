---
# devilGo-jjd0
title: 'M1 gate: AC-4 refresh — interactive demo smoke'
status: draft
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-20T12:19:16Z
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
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.hermes/skills/beans-cli-quirks-notes/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
