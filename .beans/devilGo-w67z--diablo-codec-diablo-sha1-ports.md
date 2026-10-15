---
# devilGo-w67z
title: Diablo-codec + Diablo-SHA1 ports
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - crypto
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:26Z
parent: devilGo-6xu2
---

# Save crypto

## Goal
codec_encode/decode + custom SHA1 exact.

## Research
codec.cpp:92-160; sha.cpp:8-25 (custom circular shift); tests codec_test.cpp.

## Scope
internal/persist/{codec,diablosha}.go.

## Acceptance criteria
- [ ] codec_test port green incl. wrong-password failure path
- [ ] Block size 32 + signature layout asserted

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — crypto without crypto pkg (custom algorithm)

## Dependencies & blockers
None.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
