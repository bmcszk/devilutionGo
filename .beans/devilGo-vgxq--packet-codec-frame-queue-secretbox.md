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
updated_at: 2026-10-16T14:16:44Z
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


## Libraries
- stdlib crypto/ed25519 + golang.org/x/crypto/nacl/secretbox (if already in body keep) — net stack starts stdlib net TCP (go-libraries.md decision)

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


## IMPL NOTES (triage 2026-10-16)
1 root: PT_* wire codec + frame queue + optional secretbox encryption. Parent devilGo-zo8j,
   no blockers — correct, it is the net-epic foundation.
2 local: `Source/dvlnet/packet.h:29-38` lists exactly **10 PT_* types**
   (PT_MESSAGE 0x01 … PT_ECHO_REPLY 0x32) — bean's "10 PT types" ✓. packet.cpp:225-240
   `crypto_secretbox_open_easy` nonce‖MAC‖ct ✓. golang.org/x/crypto/nacl/secretbox is pure
   Go ✓ (matches pure-Go hard rule). Skills path exists ✓.
3 online: n/a.
4 solution: none — ready. (Optional polish: research cites should say
   `Source/dvlnet/packet.h/cpp`; epic already cites "packet.h:27-40" so shorthand is fine.)
5 understanding: diablo/net foundation layer; everything else in the epic (omu6 transport,
   mfds CMD layer, 4jo7 turns) builds on it.
6 repro: `sed -n '29,38p' Source/dvlnet/packet.h`; `grep -n crypto_secretbox_open_easy
   Source/dvlnet/packet.cpp`.
7 fix: none.
8 validation: n/a.
9 related: devilGo-omu6, devilGo-mfds, devilGo-4jo7, devilGo-zo8j.
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
