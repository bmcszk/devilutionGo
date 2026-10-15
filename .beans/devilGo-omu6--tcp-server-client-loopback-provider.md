---
# devilGo-omu6
title: TCP server + client + loopback provider
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - net
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:27Z
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
