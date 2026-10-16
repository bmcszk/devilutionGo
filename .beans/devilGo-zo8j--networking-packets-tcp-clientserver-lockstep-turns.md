---
# devilGo-zo8j
title: 'Networking: packets, TCP client/server, lockstep turns, delta sync'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - phase-3
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-16T20:39:43Z
parent: devilGo-i56j
blocked_by:
    - devilGo-nxxo
---

# Multiplayer

## Goal
Go↔Go lockstep MP over TCP with devilutionX wire semantics: PT_* packets (secretbox), turn batching, CMD_* application layer, delta compression, join/leave.

## Research
packet.h:27-40 PT types, :45-62 ids, :56-59 turn_t; packet.cpp:225-240 secretbox_open (nonce‖MAC‖ct), key from password; base.h:19-100 SNet surface, PlayerState turnQueue, SendTurnIfReady; tcp_server.cpp:14-180 accept/connections[4]/route; frame_queue framing; msg.h:18-428 86 CMD_* + structs :432-600 (TCmdGItem :560-579), TPktHdr :694-709 + body[493]; msg.cpp:207-212 TMegaPkt 32000, :698 DeltaExportData, :2982-3066 senders, :3341 ParseCmd; multi.cpp:665 multi_msg_countdown, :729-760 ProcessGameMessagePackets; deltas msg.cpp:218-275 DLevel/DMonsterStr/DJunk; nthread.cpp:80-176 turn cadence. Study docs/research/dg-requirements.md §8, Appendix J.3.

## Scope
internal/net: packet codec, frame queue, tcp server/client, loopback, base turn engine, cmd.go manual marshal, delta.go; cut ZeroTier/UDP per docs/research/dg-requirements.md GI-5.

## Acceptance criteria
- [ ] Packet marshal/unmarshal fuzz vs C++ hexdumps (100 samples)
- [ ] 2-player lockstep: 10k ticks zero divergence (StateHash per tick)
- [ ] 4-player with one forced 200ms delay: turns batch correctly (no desync)
- [ ] Delta: kill/diverge scenarios converge within 1 tick
- [ ] Join mid-... (fresh join) receives game_init_info + delta snapshot; leave broadcasts PT_DISCONNECT
- [ ] Chat CMD_STRING 80-char cap enforced

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — goroutine channel discipline (single world mutator!)
/home/blaze/.agents/skills/coding-rules/SKILL.md — net layer reads as protocol spec

## Dependencies & blockers
blocked_by dg-epic-05 (commands to send), dg-epic-08 (PlayerNetPack reuse). Blocks M3 milestone.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DEFERRED — question answered by recorded milestone decision; revisit point documented in the question text above.
- H8 UDP LAN: post-M3 decision; H9 cross-play: same wire anyway.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Go↔Go TCP lockstep MP with devilutionX wire semantics. Wired blockedBy nxxo ✓ (body "dg-epic-05"), blocks vinx ✓. 5 children.
2 local: Cites verified: msg.h:694 TPktHdr ✓; msg.cpp:209 `data[32000]` (body 207-212 TMegaPkt ✓); dvlnet/packet.cpp:233 crypto_secretbox_open_easy (body 225-240 ✓); nthread.cpp:80+ nthread_send_and_recv_turn ✓; ZeroTier/UDP cut matches docs/research/dg-requirements.md GI-5 ✓ (line 37, 607).
3 online: n/a.
4 solution: none — ready. Note: secretbox must be pure-Go (x/crypto/nacl/secretbox) per pure-Go hard rule — implementation detail, no body change needed; packet-hexdump AC is exactly the mandate's net evidence form.
5 understanding: engine/net layer; state-hash-per-tick divergence gate ties directly into 1n0a harness.
6 repro: `grep -n "TPktHdr" Source/msg.h | head -2`; `sed -n '233p' Source/dvlnet/packet.cpp`
7 fix: none.
8 validation: children (vgxq, mfds, omu6, 4jo7, h8uq) triaged in batch-09.
9 related: blocker nxxo; blocked vinx; sibling 1n0a (divergence oracle).
VERDICT: TRIAGED-OK


