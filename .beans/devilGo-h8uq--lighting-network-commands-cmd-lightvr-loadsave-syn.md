---
# devilGo-h8uq
title: Lighting network sync — DeltaExportData light-state mechanism (CMD_LIGHT/V/R do not exist) + loadsave sync (L6)
status: todo
type: task
priority: normal
tags:
    - phase-lighting
created_at: 2026-10-15T22:41:05Z
updated_at: 2026-10-16T20:37:10Z
parent: devilGo-zo8j
blocked_by:
    - devilGo-gk5e
    - devilGo-kx0l
    - devilGo-bodd
    - devilGo-m18k
    - devilGo-1spo
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


## IMPL NOTES (triage 2026-10-16)
1 root: lighting-phase task proposing dedicated CMD_LIGHT/CMD_LIGHTV/CMD_LIGHTR network
   commands + handlers for cross-client light sync. Parent devilGo-zo8j, blocked_by
   gk5e/kx0l/bodd/m18k (all exist in ledger, all lighting-prereq — logically ordered).
2 local: **`grep -rn "CMD_LIGHT" Source/` → 0 hits.** devilutionX has NO light network
   commands. Real mechanism: lighting is deterministically re-derived per client from synced
   entity state; the only net-path light mutation is inside existing handlers —
   msg.cpp:870-871 (`if (monster.lightId != NO_LIGHT) ChangeLightXY(monster.lightId, position)`
   in monster-position sync). loadsave claims ARE correct: loadsave.cpp:2090 & :2647
   `ChangeLightXY(...position.tile); // forces player light refresh` verified verbatim.
   Its own research admits "verify exact CMD numbers — grep pending" — grep now done: none exist.
3 online: n/a.
4 solution: rewrite scope: (a) port the ChangeLightXY call sites inside existing CMD
   handlers (msg.cpp:870-871), (b) port ForcePlayerLightRefresh semantics from
   loadsave.cpp:2090/2647, (c) replace TestNetLightParity/EncodeLightAdd/Change/Remove with a
   determinism test: 2 clients, same CMD feed → identical Lights[] after ProcessLightList
   (that AC already exists — keep it as THE AC). Delete invented CMD_LIGHT/V/R enum/payloads.
   Also fix dep text "devilGo-7be9 (CMD layer protocol)" — 7be9 is the combat pipeline;
   the CMD layer is devilGo-mfds.
5 understanding: engine lighting state (epic qbiy children kx0l/m18k) + diablo net sync via
   mfds dispatch. Correct seam: lighting needs NO new protocol — it rides existing commands;
   inventing commands breaks wire parity with C++ (cross-play goal in epic AC).
6 repro: `grep -rn "CMD_LIGHT" <devilutionX-repo>/Source/ | wc -l` → 0;
   `sed -n '870,871p' Source/msg.cpp`; `sed -n '2090p;2647p' Source/loadsave.cpp`.
7 fix: rewrite Research/Scope/AC blocks as above; `beans update devilGo-h8uq --body-replace-old
   "devilGo-7be9 (CMD layer protocol)" --body-replace-new "devilGo-mfds (CMD layer dispatch)"`;
   add devilGo-mfds to blocked_by.
8 validation: `grep -c "CMD_LIGHT" <bean body>` → 0; blocked_by contains mfds; AC list has
   determinism test but no encode/decode-of-nonexistent-commands test.
9 related: devilGo-mfds, devilGo-kx0l, devilGo-bodd, devilGo-m18k, devilGo-gk5e.
VERDICT: NEEDS-WORK:premise-false (CMD_LIGHT/V/R do not exist in devilutionX; rewrite to
handler call-sites + loadsave refresh + determinism test; fix 7be9→mfds dep)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

PREMISE-FALSE corrected (triage): CMD_LIGHT/V/R do NOT exist in devilutionX. Rewrite scope: actual light-state delta mechanism = dLight[] sync via DeltaExportData + CalcLight/ChangeLight events; body re-scope required before claim.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
