---
# devilGo-ftbu
title: 'Milestone M1: Playable Spawn demo (town + cathedral slice)'
status: todo
type: milestone
priority: normal
tags:
    - plan:epic
    - milestone
    - phase-1
created_at: 2026-10-15T16:37:33Z
updated_at: 2026-10-29T16:01:30Z
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
## M1 INTEGRATION MANIFEST (2026-10-29)

Milestones hold no children by design (features stay under GOAL epic devilGo-i56j;
beans allows one parent). This manifest lists the exact beans M1 ACCEPTS at gate time.
Status shown is as of 2026-10-29. Gate: all below completed -> run F.2 acceptance matrix.
- Asset pipeline finishers:
  - [ ] devilGo-l5fp [todo] Graphics asset manifest, coverage checker and CLX conversion tooling
  - [ ] devilGo-r7qf [todo] CL2/CLX 0xBF is a 65-byte literal run, not a fill
  - [ ] devilGo-ow6p [todo] Sound: sfx catalog + music via ebiten audio
- Renderer core:
  - [ ] devilGo-gk5e [todo] Dungeon tile pipeline: SOL/MIN/TIL load + CEL re-encode + micro render
  - [ ] devilGo-ru6l [todo] Isometric world renderer (floor/walls/order)
  - [ ] devilGo-66o7 [todo] Index framebuffer + palette-tran shader (light tables, TRNs, blend LUT
  - [ ] devilGo-m18k [todo] Kage light application shader + TRN hero states (L4b + G4 TRN)
  - [ ] devilGo-mlv2 [todo] Panels, flasks, belt UI rendering
  - [ ] devilGo-v1hy [todo] Bitmap text renderer (CLX fonts)
  - [ ] devilGo-pdpc [todo] Dun_render corner-blit lightmap selection parity (L4)
- World:
  - [ ] devilGo-yeia [todo] Town + setmaps + triggers
  - [ ] devilGo-76tu [todo] DRLG L1 Cathedral + Crypt variant
  - [ ] devilGo-v791 [todo] Monster pack placement + behaviour constants (PlaceGroup parity)
  - [ ] devilGo-foug [todo] Themes + monster/item population + towners
  - [ ] devilGo-bodd [todo] Light source inventory data package — every emitter radius/lifetime/fl
  - [ ] devilGo-9o2h [todo] Vision system — VisionRays, DoVision, TransList, ProcessVisionList (L5
  - [ ] devilGo-4xin [todo] Corpses: dead.cpp port (Corpse records, dCorpse occupancy, corpse rend
- Gameplay minimal:
  - [ ] devilGo-2j8j [todo] Items data package: itemdat+prefixes+suffixes+uniques (436 rows, hf de
  - [ ] devilGo-baxb [todo] Affix selection + drop-rule parity (V1.6/V1.7 worked examples as tests
  - [ ] devilGo-rfrj [todo] Inventory grid + belt + gold
  - [ ] devilGo-onlm [todo] Item generation: base/affix/unique + durability
  - [ ] devilGo-0qnd [todo] Stores: TalkID machine + transactions
- UI + input:
  - [ ] devilGo-dcob [todo] Hero select/create/delete + class stats
  - [ ] devilGo-emdz [todo] Title, main menu, credits, dialogs, progress
  - [ ] devilGo-c6wt [todo] Virtual cursor + plrctrls action machine
- Persistence:
  - [ ] devilGo-oqho [todo] ItemPack/PlayerPack + SaveGameData/SaveLevel streams
  - [ ] devilGo-xwtt [todo] SQLite store + legacy .sv import/export + pfile
- Gates:
  - [ ] devilGo-41p0 [todo] Lighting parity test harness + dLight.dmp fixtures (L8 + G11)
  - [ ] devilGo-uc44 [todo] Timedemo oracle test + seed matrix CI job
  - [ ] devilGo-i2cr [in-progress] Data-completeness audit in CI: Go table row counts vs TSV (base + hf)
