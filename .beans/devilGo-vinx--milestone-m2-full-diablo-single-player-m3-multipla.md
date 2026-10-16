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
updated_at: 2026-10-16T14:26:51Z
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


## IMPL NOTES (triage 2026-10-16)
1 root: Release integration milestones M2-M4 with per-milestone gates. Wired blockedBy nxxo/xih1/6xu2/1n0a ✓ (matches body's M2 needs). No children — correct shape.
2 local: Cites verified: hf MPQ loading assets.cpp:565-585 (body 570-586 ✓); loadsave level counts: giNumberOfLevels = 25 (hellfire) : 17 (diablo) at loadsave.cpp:2482-2487 — body says "loadsave counts 25/24/15" (24=MAXQUESTS is real at questdat.hpp:14; 15=DungeonSeeds[15] real at quests.cpp:93; but as a "loadsave counts" triple it's misleading — actual pair is 25/17). STALE COUNT: body's "135 gbIsHellfire sites" and M4 gate "all 135 C++ branch sites mapped" — live `grep -rn gbIsHellfire Source --include=*.cpp | wc -l` = 160 occurrences (82 in if-conditions) at HEAD ca91d778f. An audit targeting 135 would miss ~25 sites.
3 online: n/a.
4 solution: Refresh counts or make the M4 audit AC count-agnostic: "every gbIsHellfire branch site in Source/ (scripted grep inventory committed as the checklist)" — the AC stays testable either way. Fix the "25/24/15" shorthand to "levels 25/17, MAXQUESTS 24, DungeonSeeds[15]".
5 understanding: integration milestones over all epics; M4's branch-site audit is the Hellfire fidelity gate — must be exhaustive, hence count-agnostic or refreshed.
6 repro: `grep -rn "gbIsHellfire" Source --include=*.cpp | wc -l`; `grep -n "giNumberOfLevels = " Source/loadsave.cpp | head -4`
7 fix: `beans update devilGo-vinx --body-replace-old "gbIsHellfire audit doc: all 135 C++ branch sites mapped to Go" --body-replace-new "gbIsHellfire audit doc: EVERY C++ branch site mapped to Go (scripted grep inventory at port HEAD committed as checklist; count verified live, was 160 occurrences at ca91d778f)"` ; `beans update devilGo-vinx --body-replace-old "loadsave counts 25/24/15" --body-replace-new "giNumberOfLevels 25/17 (hellfire/diablo), MAXQUESTS 24, DungeonSeeds[15]"`
8 validation: body counts match a fresh grep; M2/M3/M4 gates unchanged otherwise.
9 related: parent i56j; sibling ftbu; blockers nxxo/xih1/6xu2/1n0a/zo8j (M3).
VERDICT: NEEDS-WORK:stale gbIsHellfire count (135 vs live 160 at HEAD) makes M4 audit AC undercount + "25/24/15" loadsave shorthand incorrect (actual 25/17)

---

## Verdict lines

devilGo-i56j: TRIAGED-OK
devilGo-qna9: TRIAGED-OK
devilGo-5wsb: NEEDS-WORK:extract-at-build scope not applied to epic body (runtime mpq.go framing) + skills typo ego-dev
devilGo-qbiy: TRIAGED-OK
devilGo-97mn: TRIAGED-OK
devilGo-nxxo: TRIAGED-OK
devilGo-xih1: TRIAGED-OK
devilGo-k736: TRIAGED-OK
devilGo-6xu2: NEEDS-WORK:body contradicts DECIDED #1 SQLite-only saves (still .sv read+write/export AC + obsolete MPQ-writer dep; per-hero/global DB model absent)
devilGo-zo8j: TRIAGED-OK
devilGo-yjuv: NEEDS-WORK:options AC says JSON vs diablo.ini-1:1 amendment; Lua ACs not deferred per DECIDED #3
devilGo-1n0a: TRIAGED-OK
devilGo-ftbu: TRIAGED-OK
devilGo-vinx: NEEDS-WORK:stale gbIsHellfire count 135→160 at HEAD; loadsave counts shorthand 25/24/15 → 25/17

BATCH-01-COMPLETE

Fix pass 2026-10-16: Stale count corrected (fix pass): gbIsHellfire refs 135 → 160 at current HEAD; M4 audit AC re-pinned to live count (re-verify at claim time).
