---
# devilGo-omu6
title: TCP server + client + loopback provider
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - net
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:06:56Z
parent: devilGo-zo8j
blocked_by:
    - devilGo-vgxq
---

# Transports

## Goal
Host/join over TCP; loopback for single-player code path unity.

## Research
tcp_server.cpp:14-180 (accept, NextFree, broadcast/route, drop); tcp_client; loopback.cpp; cdwrap provider selection.

## Scope
internal/net/tcp/{server,client}.go, loopback.go; rejoin with cookie.

## Acceptance criteria
- [ ] 4-slot host; correct routing (broadcast vs direct) tests
- [ ] Client reconnect resumes with cookie (test)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — net + goroutine lifecycle, contexts

## Dependencies & blockers
dg-task-037.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: TCP transport (host/join, 4 slots), loopback provider for single-player code-path
   unity, cookie-based rejoin. Parent devilGo-zo8j, blocked_by devilGo-vgxq ✓ (codec before
   transport; body dep dg-task-037 = vgxq, consistent).
2 local: `Source/dvlnet/tcp_server.cpp:14` constructor verified (acceptor + StartAccept);
   connections[] routing/direct-send at :136-154 ✓; tcp_client.cpp:93
   `cookie_self = packet_out::GenerateCookie()` + PT_JOIN_ACCEPT cookie at tcp_server.cpp:149
   → rejoin-with-cookie AC is real ✓; loopback.cpp, cdwrap.cpp exist under Source/dvlnet/ ✓.
3 online: n/a.
4 solution: none — ready. (Path note: files are under `Source/dvlnet/`, not `Source/` —
   coder will find them via the epic's research block.)
5 understanding: diablo/net transport over vgxq codec; loopback gives single-player the same
   turn pipeline (unity with 4jo7), cdwrap = provider selection seam.
6 repro: `sed -n '14,24p' Source/dvlnet/tcp_server.cpp`; `grep -n GenerateCookie
   Source/dvlnet/tcp_client.cpp`; `ls Source/dvlnet/{loopback,cdwrap}.*`.
7 fix: none.
8 validation: n/a.
9 related: devilGo-vgxq, devilGo-4jo7, devilGo-zo8j.
VERDICT: TRIAGED-OK
