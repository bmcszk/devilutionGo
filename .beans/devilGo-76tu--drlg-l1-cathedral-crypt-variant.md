---
# devilGo-76tu
title: DRLG L1 Cathedral + Crypt variant
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - drlg
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:16:44Z
parent: devilGo-97mn
blocked_by:
    - devilGo-vti7
---

# L1 generator

## Goal
CreateL5Dungeon parity incl. crypt branch.

## Research
drlg_l1.cpp:1298-1312 entry; internals :1200-1294 (Fix*, Substitution, lamps, FillFloor, SelectChamber); crypt :1214,1221; Pass3/DRLG_LPass3 gendung.h:88; fixtures test/fixtures/diablo + levels/l1data.

## Scope
internal/world/drlg_l1.go + crypt.go sharing engine.

## Acceptance criteria
- [ ] All diablo L1 fixtures byte-equal
- [ ] Hellfire crypt fixtures byte-equal
- [ ] Quest chamber selection matches SelectChamber logic for all availability combos

## Skills for coder
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — fixture matrix
/home/blaze/.agents/skills/go-dev/SKILL.md — RNG-call-order discipline

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md
- /home/blaze/.agents/skills/game-development/game-design/SKILL.md

## Dependencies & blockers
dg-task-015, dg-task-002.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-1 generator task (dg-task-017): CreateL5Dungeon parity incl. crypt; draft; wire blockedBy [vti7] ✓; body's dg-task-002 (m2x7 RNG) is transitively covered (vti7→m2x7 chain verified).
2 local: drlg_l1.cpp:1298-1312 CreateL5Dungeon ✓ exact (sed match, incl. crypt branch `if (leveltype == DTYPE_CRYPT)`); crypt internals :1214 (FixCryptDirtTiles/CryptSubstitution) ✓; SelectChamber at :1270 ✓ (body ":1268-1294" ✓); fixtures live: test/fixtures/diablo 27 .dun + hellfire 17 .dun ✓, levels/l1data ✓, drlg_l1_test.cpp ✓. One cite drift: `DRLG_LPass3 gendung.h:88` — actual gendung.h:95.
3 online: n/a.
4 solution: none blocking — refresh one line number.
5 understanding: world layer core; consumes vti7 grids/miniset + m2x7 RNG call-order; feeds 4ybb/yeia/foug; fixture-byte-equality is exactly the epic's parity gate.
6 repro: `sed -n '1298,1312p' devilutionX/Source/levels/drlg_l1.cpp; grep -n "DRLG_LPass3" devilutionX/Source/levels/gendung.h; ls devilutionX/test/fixtures/diablo/*.dun | wc -l`
7 fix: optional `beans update devilGo-76tu --body-replace-old "Pass3/DRLG_LPass3 gendung.h:88" --body-replace-new "Pass3/DRLG_LPass3 gendung.h:95"`
8 validation: grep shows :95 after edit.
9 related: vti7 (dep), 4ybb/yeia (dependents), foug (population).
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
