---
# devilGo-vinx
title: 'Milestone M2: Full Diablo single-player; M3: Multiplayer; M4: Hellfire'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - milestone
    - phase-2
    - phase-3
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-i56j
blocked_by:
    - devilGo-nxxo
    - devilGo-xih1
    - devilGo-6xu2
    - devilGo-1n0a
---

# Release milestones M2-M4

## Goal
M2: complete Diablo SP (16 quests, 4 DRLGs, all classes, stores, saves, timedemo gate). M3: TCP lockstep MP (2-4p). M4: Hellfire (25 levels, Monk, 8 quests, crypt/nest, hf MPQs).

## Research
Matrices /tmp/dg-requirements.md Appendix F.3-F.5; Hellfire branch inventory (135 gbIsHellfire sites; loadsave counts 25/24/15); mpq hellfire loading assets.cpp:570-586.

## Scope
Milestone integration epics pulling from all prior epics; gates defined below.

## Acceptance criteria
M2:
- [ ] All 16 quests completable in scripted runs; setmaps functional
- [ ] Combat/itemgen/store parity suites green (epic-05 criteria)
- [ ] .sv round-trip + SQLite persistence green; timedemo oracle green
- [ ] 3-seed × 2000-tick C++ cross-check zero divergence (if sidecar available) else state-hash stability
M3:
- [ ] Epic-09 lockstep criteria green; 4-player session 30min stable
M4:
- [ ] hellfire.mpq+hf* load; Monk playable; crypt/nest generators fixture-green; 24 quests; runes/oils items
- [ ] gbIsHellfire audit doc: all 135 C++ branch sites mapped to Go

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — integration and feature-flag hygiene
/home/blaze/.agents/skills/beans/SKILL.md — milestone tracking discipline

## Dependencies & blockers
As listed; M3 also needs dg-epic-09; M4 needs assets hf support (epic-02 done).

## Open questions
- M4 movies (H11): stills+text default; SMK decode optional stretch.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:stale gbIsHellfire count (135 vs live 160 at HEAD) makes M4 audit AC undercount + "25/24/15" loadsave shorthand incorrect (actual 25/17)
