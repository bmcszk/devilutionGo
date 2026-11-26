---
# devilGo-f6zv
title: 'ebiten/rendering skill misguidance: windowed tag alone does not enable ReadPixels'
status: draft
type: task
created_at: 2026-11-26T11:51:14Z
updated_at: 2026-11-26T11:51:14Z
parent: devilGo-p1t8
---

SKILL.md lines 41+49 imply //go:build windowed alone satisfies the ebiten ReadPixels guard; it only links ebiten - running is set solely under ebiten.RunGame (ui.go:164-167). Produced the 7bij panic. Fix in bmcszk/agent-skills (public repo - no house vocabulary): state tag necessary-not-sufficient, point at lane-ops visual-verification law. Evidence: .hermes/lane/m2-followups/triage-result.md (7bij verdict).
