---
# devilGo-ftbu
title: 'Milestone M1: Playable Spawn demo (town + cathedral slice)'
status: draft
type: epic
priority: normal
tags:
    - plan:epic
    - milestone
    - phase-1
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-16T14:06:44Z
parent: devilGo-i56j
blocked_by:
    - devilGo-5wsb
    - devilGo-qbiy
---

# M1 Playable Spawn slice

## Goal
A human can: launch, pick warrior, walk Tristram, enter cathedral (L1), fight zombies/skeletons with melee, pick up items, return to town, save/load via SQLite, all with sound.

## Research
M1 matrix docs/research/dg-requirements.md Appendix F.2; spawn gating (quests.cpp:96-100, monster.cpp:3159); town + L1 tasks; stores subset.

## Scope
Integration: epics 02/03 + world L1/town + entity basics + combat melee + inventory minimal + UI hero flow + persistence minimal + input kb/mouse + audio.

## Acceptance criteria
- [ ] End-to-end scripted session (boot→create→town→dungeon→kill 10→loot→save→load) runs headless with zero errors and stable StateHash on replay
- [ ] Seed parity for L1 fixtures (from devilGo-76tu) included in gate
- [ ] 60fps rendering of town+dungeon on mid hardware (no frame >16ms p95)
- [ ] Manual smoke checklist documented (screenshots attached to bean)

## Skills for coder
- /home/blaze/.agents/skills/ebiten/game-loop/SKILL.md
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md
- /home/blaze/.agents/skills/ebiten/input/SKILL.md
- /home/blaze/.agents/skills/ebiten/audio/SKILL.md — integration surfaces
/home/blaze/.agents/skills/go-dev/SKILL.md — wiring discipline

## Dependencies & blockers
blocked_by dg-epic-02, dg-epic-03 (+ tasks 017/019/021 partial from epics 04/05 — coordinate; the milestone may pull specific tasks early rather than wait for full epics).

## Open questions
None.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Integration milestone: boot→warrior→town→L1→melee→loot→save/load SQLite with sound. Wired blockedBy 5wsb+qbiy ✓ (matches body dg-epic-02/03). No children (integration epic — correct shape).
2 local: Cites verified: quests.cpp:96-100 spawn quest wipe ✓; monster.cpp:3159 retail-availability gate ✓; M1 matrix exists at docs/research/dg-requirements.md §Appendix F.2 (line 880) ✓. Body's legacy "tasks 017/019/021" resolve via docs/research/dg-id-map.json to 76tu/yeia/7be9 ✓ (correct tasks to pull early).
3 online: n/a.
4 solution: none — ready. AC set is testable (scripted headless session + StateHash replay stability + p95 frame budget + manual checklist with screenshots).
5 understanding: pure integration layer over qna9/5wsb/qbiy + partial 97mn/nxxo/6xu2/xih1/k736 — the milestone's job is pulling, not owning systems.
6 repro: `sed -n '880,887p' docs/research/dg-requirements.md`; `beans query --json '{ bean(id: "devilGo-ftbu") { blockedBy { id } } }'`
7 fix: none. Optional: swap "tasks 017/019/021" → "devilGo-76tu/yeia/7be9" for durability.
8 validation: when M1 work starts, `beans list --ready` shows the pulled tasks unblocked appropriately.
9 related: parent i56j; sibling vinx; pulled tasks 76tu/yeia/7be9.
VERDICT: TRIAGED-OK

## TRIAGE 2026-10-18 (pi wave, batch 4)

devilGo-ftbu — M1 Playable Spawn demo (milestone epic)**
1 root: Integration milestone, not system owner; body+IMPL notes from 2026-10-16 sweep present.
2 local: cites verified live — `quests.cpp:96-100` spawn quest wipe ✓, `monster.cpp:3159` Retail-availability gate ✓; blockedBy YAML `5wsb,qbiy` matches body "dg-epic-02/03" ✓; pulled tasks real: 76tu todo, yeia todo, 7be9 **completed** (early-pull model working); parent i56j ✓.
3 online: SKIPPED (no doc claim requiring it).
4 scope-fit vs children: owns boot→hero→town→L1→melee→loot→save/load+sound; pulls from 97mn/76tu (world), 7be9 (combat), rfrj (loot), xwtt/6xu2 (saves), dcob (hero flow), k736, ow6p — all exist, no orphan scope.
5 DDD fit: milestone = orchestration only, no engine/game/diablo code claimed ✓.
6 dep-order: 5wsb+qbiy both draft → M1 correctly gated behind asset+renderer epics; qna9 (todo) upstream ✓.
7 AC-gap: parent i56j global AC#5 slice includes "death/**level-up**"; ftbu goal/ACs omit both — add one AC line or record a deferral decision.
8 validation: when claimed, `beans list --ready` shows 76tu/yeia/7be9 unblocked; scripted-session StateHash gate from bean.
9 related: parent i56j, sibling vinx, tasks 76tu/yeia/7be9.
**VERDICT: NEEDS-WORK:death/level-up missing from M1 slice vs parent AC#5** (else body current)


Wave outputs: /tmp/dg-triage/out-b4.md
