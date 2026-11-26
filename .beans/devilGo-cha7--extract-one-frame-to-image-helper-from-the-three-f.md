---
# devilGo-cha7
title: Extract one frame-to-image helper from the three frameToImage/frameImage copies
status: draft
type: task
created_at: 2026-11-26T12:15:57Z
updated_at: 2026-11-26T12:15:57Z
parent: devilGo-qbiy
---

Coding-rules extract-on-3rd-occurrence finding (m2-followups review NB-1, APPROVE round): the frame->NRGBA loop now exists three times - internal/present/scene/town_compose.go:223-233 (frameToImage, unexported), cmd/devilutionGo/ac4_demo_gate_test.go:252-262 (frameToImage), internal/present/scene/screenshot_test.go:78-91 (frameImage, added by the 7bij fix). Fix shape per reviewer: export one scene.FrameToImage(frame render.Frame) image.Image (or an internal/testsupport/img helper) and collapse all three copies onto it. Not fixable inside the wave's scope jail (needs an export or a shared test-support package). Draft; triage owed (pick the home, check for more copies).
