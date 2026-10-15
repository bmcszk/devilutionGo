---
# devilGo-mfds
title: 'CMD layer: 86 commands + TPkt/TMegaPkt + ParseCmd dispatch'
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - net
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-zo8j
blocked_by:
    - devilGo-vgxq
---

# Commands

## Goal
Full command enum + payload structs + dispatch to gameplay handlers.

## Research
msg.h:18-600; msg.cpp senders/dispatch; header validation multi.cpp:729-760.

## Scope
internal/net/cmd.go (manual marshal, exact widths), handlers bound to gameplay; NetSendCmd* API mirroring signatures (Appendix L).

## Acceptance criteria
- [ ] Enum count == C++ (86 + sentinels); payload size table asserted
- [ ] Round-trip all commands via fuzz vs hexdumps
- [ ] Buffered-message mode (gbBufferMsgs) honored during join

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — exhaustive switches, no reflection hot path

## Dependencies & blockers
dg-task-037, dg-epic-05 handlers.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
