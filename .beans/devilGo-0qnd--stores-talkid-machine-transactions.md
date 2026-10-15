---
# devilGo-0qnd
title: 'Stores: TalkID machine + transactions'
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - stores
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-15T16:45:27Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-onlm
---

# Stores

## Goal
All 23 TalkID states; buy/sell/repair/recharge/identify/premium/boy/healer; gossip texts.

## Research
stores.h:44-70; stores.cpp flows (StartStore re-entry :504,911-929,1192,1293); tests store_transaction/vendor/visual_store.

## Scope
internal/gameplay/stores: pure transaction core (prices, stock, premium refresh timers, gold checks) + UI-facing state struct the UI epic consumes.

## Acceptance criteria
- [ ] store_transaction_test port green
- [ ] Premium item slots 6 (D) / 15 (HF) refresh per C++ timers

## Skills for coder
/home/blaze/.agents/skills/coding-rules/SKILL.md — separate intent (transaction) from detail (UI refresh)
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — scripted store sessions

## Dependencies & blockers
dg-task-023.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
