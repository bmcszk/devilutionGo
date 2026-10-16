---
# devilGo-h8uq
title: Lighting network commands — CMD_LIGHT/V/R + loadsave sync (L6)
status: draft
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-zo8j
blocked_by:
    - devilGo-gk5e
    - devilGo-kx0l
    - devilGo-bodd
    - devilGo-m18k
---

# Lighting network commands + loadsave sync

## Goal
Implement the exact CMD payloads and handlers for light synchronization across clients (L6), plus level-load forced refresh from loadsave.cpp.

## Research
- msg.cpp CMD_* list (verify exact CMD numbers: CMD_LIGHT=?, CMD_LIGHTV=?, CMD_LIGHTR=?) — grep pending
- Typical payload: {x,y,radius} or {light_id, x,y,radius,offset} for Add/Change/Remove
- Receiving side calls same state funcs: AddLight/ChangeLightXY/ChangeLightRadius/ChangeLightOffset/AddUnLight
- loadsave.cpp:2090/2647 forced ChangeLightXY(MyPlayer) on level load (single+multi)
- Determinism: all clients run ProcessLightList on same tick after applying net commands

## Scope
Package: `internal/net/lighting.go` (in networking epic)
- `func HandleLightCmd(cmd Cmd, payload []byte) error` — decode + call state funcs
- `func EncodeLightAdd(lid int, pos Point, radius uint8) []byte`
- `func EncodeLightChange(lid int, pos Point, radius uint8, offset DisplacementI8) []byte`
- `func EncodeLightRemove(lid int) []byte`
- Net command enum mirroring C++ values (grep msg.cpp for exact IDs)
- LoadSave path: `func ForcePlayerLightRefresh()` → ChangeLightXY(MyPlayerId, myPlayer.position.tile)

## Acceptance criteria
- [ ] `TestNetLightParity` encode/decode roundtrip matches C++ hexdump (fixture `test/fixtures/net/light_cmds.bin`)
- [ ] `TestNetLightApply` apply decoded cmd → Lights[] state identical to C++ after same sequence
- [ ] `TestLoadSaveRefresh` level load triggers ChangeLightXY on MyPlayer → light position synced
- [ ] `TestDeterminism` 2 clients receive same cmd sequence → same Lights[] after ProcessLightList

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md

## Dependencies & blockers
- devilGo-qbiy (state funcs must exist)
- devilGo-7be9 (CMD layer protocol)

## Open questions
- Exact CMD enum values + struct layouts from msg.cpp — targeted grep required

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:premise-false (CMD_LIGHT/V/R do not exist in devilutionX; rewrite to
