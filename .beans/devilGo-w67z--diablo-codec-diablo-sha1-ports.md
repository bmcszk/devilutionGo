---
# devilGo-w67z
title: Diablo-codec + Diablo-SHA1 ports
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - crypto
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:06:45Z
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

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Persistence epic leaf — codec_encode/decode + custom SHA1; no blockers (matches tree dg-task-034 unblocked; correct: crypto has no deps).
2 local: codec.cpp:92 codec_decode verified; codec.cpp:15-21 CodecSignature{checksum,error,lastChunkSize} + SignatureSize 8 + BlockSizeBytes layout verified (AC "block size 32 + signature layout" assertable); sha.cpp:8-25 custom arithmetic-shift SHA1 comment + SHA1CircularShift verified; test/codec_test.cpp exists.
3 online: n/a.
4 solution: none — cleanest bean in batch: C++ test suite portable as golden vectors first (TDD-first trivially satisfiable), wrong-password path is in codec_test.
5 understanding: persistence-layer crypto leaf feeding oqho/xwtt (hero pack encryption via pfile passwords, pfile.cpp:60-64 verified).
6 repro: `beans show devilGo-w67z`; `sed -n '92,96p' .../Source/codec.cpp`; `sed -n '8,25p' .../Source/sha.cpp`; `ls ~/github/bmcszk/devilutionX/test/codec_test.cpp`.
7 fix: none.
8 validation: n/a.
9 related: oqho, xwtt, 6xu2.
VERDICT: TRIAGED-OK
