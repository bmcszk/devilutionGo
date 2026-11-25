---
# devilGo-jjd0
title: 'M1 gate: AC-4 refresh — interactive demo smoke'
status: completed
type: task
priority: normal
created_at: 2026-11-20T12:19:10Z
updated_at: 2026-11-23T18:01:06Z
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
- [x] Smoke doc steps re-scoped per the AC-4 review: steps 1-3 + 9
      window-reachable (title/menu/selhero + the walk beats), steps 4-8
      pinned headless per the doc split — fresh per-step captures with
      pixel stats in the PoW
- [x] Scripted-walk oracle: statehash values byte-identical to the
      merged s5 pins (or the disclosed re-freeze, deterministic)
- [x] Full gates green at HEAD through pre-commit + pre-push

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



## Progress log

- 2026-11-20T21:10Z: claimed — AC-4 refresh gate dispatched under contract harness.ac4refresh.md (worktree ftbu-demo, feat/m1-demo-flow; base 7d87dc0: boot → title → menu → selhero → town walk all live).



## Proof of Work

- Worker task-dg-m1-ac4refresh-coder-coder: b86151e (18 files: doc refresh, 7 PNG captures + .stats sidecars, gate test TestAC4InteractiveDemoSmoke) → review PASS with two required doc notes → docfix task-dg-m1-ac4-docfix-coder: 8418762 (F1 no-glyph deferral named in the doc; F2 name-entry honesty — no rune reader, empty-name hero possible, k736 home).
- RE-SCOPE (per review answer 8/F6, parent-recorded): window-reachable = steps 1-3 + 9; steps 4-8 pinned headless per the doc split. The literal 'steps 2-8' AC text would have certified false — re-scoped before ticking.
- Beat proof: the interactive path is driven on the production construction surface — boot → title → menu → selhero create (CreatedPack in the tick log + store row) → town walk (PlayerMoved; spawn 75,68→78,65; walk stats within 0.02pp of the validated baseline 60.37% nonblack). Beat4's frame equals the menu (disclosed; the proof is the event + store row); beat4b roster-visible capture + glyph/text integration (h6wn/v1hy) + HUD (qbiy) are the named follow-up homes.
- Parent gates at 8418762: full suite 54 pkgs ok; trail f1ae95a0… unchanged; ls-remote == HEAD.
