---
# devilGo-mfds
title: 'CMD layer: 86 commands + TPkt/TMegaPkt + ParseCmd dispatch'
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - net
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:26:53Z
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

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:AC-count-wrong (86→80; add/clarify nxxo handler dep edge)


## IMPL NOTES (triage 2026-10-16)
1 root: net-epic task defining the CMD_* application layer (enum + payload structs +
   manual marshal + dispatch + gbBufferMsgs). Structurally complete (goal/scope/AC/skills/TDD
   gates present), parent devilGo-zo8j, blocked_by devilGo-vgxq (correct: codec first).
2 local: `beans show` OK. Parsed `Source/msg.h` enum `_cmd_id` with a script: **80 real
   CMD_* values** (CMD_STAND..CMD_SPAWNMONSTER, indices 0-79) + FAKE_CMD_SETID/DROPID +
   NUM_CMDS + CMD_INVALID=0xFF = 84 entries. **No 86 anywhere.** msg.h:18 enum start ✓;
   msg.cpp:3341 `ParseCmd` ✓; msg.cpp:108 `gbBufferMsgs` ✓; multi.cpp:729-760
   ProcessGameMessagePackets header validation ✓ (wCheck/wLen checks confirmed);
   msg.cpp:207 `TMegaPkt`, msg.h:711 `TPkt` ✓; Appendix L exists (/tmp/dg-requirements.md:1084).
3 online: n/a (no external claims).
4 solution: fix the AC count: "Enum count == C++ (80 real commands + 2 FAKE + NUM_CMDS +
   CMD_INVALID sentinels)". Also body says "dg-epic-05 handlers" (nxxo) is a dependency but
   nxxo is not in blocked_by — either add edge or keep vgxq-only and note handlers land with
   gameplay tasks.
5 understanding: diablo/net layer (internal/net/cmd.go) — consumes vgxq's wire codec,
   consumed by 4jo7's delta/turn apply. The dispatch table is where gameplay handlers get
   bound, so ordering after gameplay epic tasks is real.
6 repro: `python3 -c` count of `_cmd_id` enum entries in Source/msg.h (script in triage
   log); `sed -n '3341p' Source/msg.cpp`; `grep -n gbBufferMsgs Source/msg.cpp | head -3`.
7 fix: `beans update devilGo-mfds --body-replace-old "Enum count == C++ (86 + sentinels)"
   --body-replace-new "Enum count == C++ (80 real CMD_* + FAKE_CMD_SETID/FAKE_CMD_DROPID +
   NUM_CMDS + CMD_INVALID sentinels; assert 80 dispatch entries)"`. Same "86 CMD_*" error
   appears in epic devilGo-zo8j research line — parent should fix there too.
8 validation: re-run `beans show devilGo-mfds --body-only | grep -c "80 real"` ≥ 1 and
   enum-count script vs AC number match.
9 related: devilGo-vgxq (codec), devilGo-4jo7 (turn/delta), devilGo-zo8j (epic, shares the
   86 error), devilGo-h8uq (cites this as "devilGo-7be9" wrongly).
VERDICT: NEEDS-WORK:AC-count-wrong (86→80; add/clarify nxxo handler dep edge)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

AC count corrected (triage): 80 handlers, not 86; handler dep edge added.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
