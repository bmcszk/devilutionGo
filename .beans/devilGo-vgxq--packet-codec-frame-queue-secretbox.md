---
# devilGo-vgxq
title: Packet codec + frame queue + secretbox
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - net
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:01:18Z
parent: devilGo-zo8j
---

# Wire format

## Goal
PT_* packets byte-compatible (for future cross-play) with optional plaintext mode.

## Research
packet.h/cpp as in epic; golang.org/x/crypto/nacl/secretbox.

## Scope
internal/net/packet.go, frame.go.

## Acceptance criteria
- [ ] All 10 PT types round-trip; encrypted mode passes C++-sampled vectors
- [ ] Fuzz: malformed input → error never panic

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — manual binary encode/decode

## Dependencies & blockers
None.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
