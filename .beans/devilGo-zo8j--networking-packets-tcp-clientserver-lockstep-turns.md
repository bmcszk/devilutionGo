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
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-i56j
blocked_by:
    - devilGo-nxxo
---

# Multiplayer

## Goal
Go↔Go lockstep MP over TCP with devilutionX wire semantics: PT_* packets (secretbox), turn batching, CMD_* application layer, delta compression, join/leave.

## Research
packet.h:27-40 PT types, :45-62 ids, :56-59 turn_t; packet.cpp:225-240 secretbox_open (nonce‖MAC‖ct), key from password; base.h:19-100 SNet surface, PlayerState turnQueue, SendTurnIfReady; tcp_server.cpp:14-180 accept/connections[4]/route; frame_queue framing; msg.h:18-428 86 CMD_* + structs :432-600 (TCmdGItem :560-579), TPktHdr :694-709 + body[493]; msg.cpp:207-212 TMegaPkt 32000, :698 DeltaExportData, :2982-3066 senders, :3341 ParseCmd; multi.cpp:665 multi_msg_countdown, :729-760 ProcessGameMessagePackets; deltas msg.cpp:218-275 DLevel/DMonsterStr/DJunk; nthread.cpp:80-176 turn cadence. Study /tmp/dg-requirements.md §8, Appendix J.3.

## Scope
internal/net: packet codec, frame queue, tcp server/client, loopback, base turn engine, cmd.go manual marshal, delta.go; cut ZeroTier/UDP per /tmp/dg-requirements.md GI-5.

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
- H8 UDP LAN: post-M3 decision; H9 cross-play: same wire anyway.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK
