---
# devilGo-zo8j
title: 'Networking: packets, TCP client/server, lockstep turns, delta sync'
status: todo
type: feature
priority: normal
tags:
    - plan:epic
    - phase-3
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-29T16:13:18Z
parent: devilGo-i56j
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
- ~/.agents/skills/ponytail/SKILL.md
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

## TRIAGE 2026-10-18 (pi wave, batch 3)

devilGo-zo8j** (Networking epic, parent i56j)
1 root: Go↔Go TCP lockstep w/ devilutionX wire semantics; gate nxxo ✓ wired; blocks M3 (vinx prose).
2 local: `msg.h:694 TPktHdr` ✓; `msg.cpp:207-212 TMegaPkt data[32000]` ✓; `packet.cpp:233 crypto_secretbox_open_easy` (body 225-240 ✓); `nthread.cpp:80 nthread_send_and_recv_turn` ✓; PT_* table :29-40 ✓; `tcp_server.cpp:14-15` ctor ✓. 5 children exist (vgxq,mfds,omu6,4jo7,h8uq).
3 online: SKIPPED.
4 solution: **dep-wiring gap**: prose declares "blocked_by dg-epic-08 (PlayerNetPack reuse)" = 6xu2, but frontmatter lists only nxxo. Fix-pass claimed all 41 unwired edges created — this one was missed. Wire devilGo-6xu2.
5 DDD fit: engine/net (codec/transport) + diablo/net (cmd/delta application) ✓ matches split of children.
6 dep-order: nxxo (cmds) + 6xu2 (netpack) → zo8j → M3.
7 AC-gap: fuzz-vs-hexdump 100 samples ✓ concrete; StateHash-per-tick ties to 1n0a oracle ✓; secretbox pure-Go (x/crypto) noted — fine.
8 validation: 10k-tick zero-divergence; 4P/200ms batch test; delta converge ≤1 tick.
9 related: nxxo, 6xu2, 1n0a, vgxq/mfds/omu6/4jo7/h8uq.

**NEEDS-WORK:declared dg-epic-08 → devilGo-6xu2 (PlayerNetPack) dep not wired in frontmatter (blocked_by lists only nxxo)**


Wave outputs: /tmp/dg-triage/out-b3.md

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
