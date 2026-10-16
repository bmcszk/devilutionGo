---
# devilGo-oqho
title: ItemPack/PlayerPack + SaveGameData/SaveLevel streams
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - save
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:05Z
parent: devilGo-6xu2
blocked_by:
    - devilGo-w67z
---

# Serializers

## Goal
All pack/unpack + game/level stream writers with exact endianness.

## Research
pack.h structs; loadsave.cpp:1503,1902-2900; MaxMissilesForSaveGame 125 ordering rules; padding to MaxMonsters.

## Scope
internal/persist/{pack,loadsave}.go with SaveHelper-equivalent (WriteLE/WriteBE/Skip).

## Acceptance criteria
- [ ] pack_test.cpp:441-868 all cases green
- [ ] 5 scripted states hex-equal SaveGameData
- [ ] LoadLevel regenerates identical StateHash (round-trip)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — explicit byte order per field, no binary.Read shortcuts

## Dependencies & blockers
dg-task-034, dg-epic-04 structs.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
