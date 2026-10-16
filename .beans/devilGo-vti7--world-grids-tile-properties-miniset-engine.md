---
# devilGo-vti7
title: World grids + tile properties + miniset engine
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - world
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T14:06:56Z
parent: devilGo-97mn
blocked_by:
    - devilGo-m2x7
---

# Grids foundation

## Goal
The 12 world arrays + SOL loading + Miniset match/place + transparency bookkeeping.

## Research
dun_tile_data.hpp:26-120; gendung.h Miniset :35-70; tile_properties.cpp:11-41 + CanStep rules tile_properties.hpp:28-37; SOL loading LoadLevelSOLData.

## Scope
internal/world/grid.go, tileprops.go, miniset.go, trans.go (DRLG_InitTrans/MRectTrans/RectTrans/CopyTrans).

## Acceptance criteria
- [ ] tile_properties_test.cpp port passes
- [ ] Miniset matches/place unit tests incl. Protected flag and crypt respectProtected=false quirk
- [ ] TransList semantics (TransVal) identical on scripted ops

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — fixed arrays, no slices in hot paths

## Dependencies & blockers
dg-task-002.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-1 foundation task (dg-task-015): 12 world arrays + SOL loading + Miniset + trans bookkeeping; draft; wire blockedBy [m2x7] ✓ = body's dg-task-002; dependents gk5e/r462/76tu wired ✓.
2 local: Verified: dPiece/dTransVal/dLight/dFlags/dPlayer/dMonster/dCorpse/dObject/dSpecial externs at dun_tile_data.hpp:99-136 (+TransList :97) ✓; SOLData dun_tile_data.cpp:41 + LoadLevelSOLData :99 ✓ exact (town.sol at :103); Miniset gendung.h:37 ✓ (body ":35-70" ✓) with the Crypt `respectProtected=false` bug documented in-source at gendung.h:45-53 ✓ ("Match bug from Crypt levels if false"); tile_properties.cpp:11 IsTileNotSolid ✓; CanStep tile_properties.hpp:28-37 ✓; tile_properties_test.cpp ✓ on disk. Seam with gk5e (SOL load vs re-encode) is ORDERED by the existing gk5e-blockedBy-vti7 edge and ACs are disjoint (vti7: loader+semantics tests; gk5e: hand-patch index list, verified in batch-04) — acceptable, one-line declaration would be nice.
3 online: n/a.
4 solution: none — ready; optional body note "owns LoadLevelSOLData + SOLData; gk5e consumes for re-encode".
5 understanding: the root of the whole epic — every generator, entity container, and the render tile pipeline hangs off these arrays; correctly first in every chain.
6 repro: `grep -n "extern.*d\(Piece\|TransVal\|Light\|Flags\|Player\|Monster\|Corpse\|Object\|Special\)" devilutionX/Source/levels/dun_tile_data.hpp | head; grep -n "LoadLevelSOLData\|SOLData\[" devilutionX/Source/levels/dun_tile_data.cpp | head -3; sed -n '45,53p' devilutionX/Source/levels/gendung.h`
7 fix: none (optional seam note above).
8 validation: n/a.
9 related: m2x7 (dep), r462/76tu/gk5e/9o2h (dependents incl. via fixes), 4ybb/yeia/foug (transitive).
VERDICT: TRIAGED-OK

---

## Verdict summary
- devilGo-pu3r: NEEDS-WORK:wire-deps (4 missing blockedBy edges: aix0/rctk/dcob/xw8k; playerdat.h→playerdat.cpp cite)
- devilGo-h6u0: TRIAGED-OK
- devilGo-76tu: TRIAGED-OK (1 optional cite: gendung.h:88→:95)
- devilGo-4ybb: TRIAGED-OK
- devilGo-r462: TRIAGED-OK
- devilGo-fpkz: NEEDS-WORK:wire-deps (h6u0/yeia/foug/l5fp edges) + wrong cite (InitThemes :349-449→:826) + L3UpList OQ resolvable
- devilGo-bodd: NEEDS-WORK:content-defect (AddLight sites ≈30 not 13/14; dead.cpp mislabel; ±2-tile tolerance AC; body-vs-wire deps)
- devilGo-v791: NEEDS-WORK:wire-deps (vti7, foug edges; 3 minor cite drifts)
- devilGo-778h: NEEDS-WORK:wire-deps (1e1b/b8z9/foug edges; 2 minor cite drifts)
- devilGo-1e1b: NEEDS-WORK:wire-deps (fpkz/h6wn/yeia/foug edges; trailing-newline + ObjectData cite errors)
- devilGo-foug: NEEDS-WORK:wire-deps (r462 edge; themes.cpp:349-449→:826 cite)
- devilGo-yeia: NEEDS-WORK:ac-stale-vs-decision ("loads from MPQ" vs extract-at-build; town.dun misnomer)
- devilGo-9o2h: NEEDS-WORK:wrong-deps (wire 5 render beans vs body gk5e vs correct vti7; 2 cite drifts)
- devilGo-vti7: TRIAGED-OK

Counts: TRIAGED-OK 5 · NEEDS-WORK 9 · DUPLICATE-OF 0 · STALE 0 · DECISION-NEEDED 0.
Coverage: 14/14 assigned ids triaged (batch-06-ids.txt match); these are exactly the 14 children of devilGo-97mn, so no epic child is missed and no foreign id appeared.
Cross-batch seams for the parent: (1) fpkz↔h6u0 Hive/Crypt tile sharing — wired via fpkz fix; (2) 9o2h dep rewire touches batch-04/05 lighting beans (kx0l/m18k/pdpc wires include bodd — recheck after bodd re-scope); (3) batch-02's package-layout DECISION still gates all Scope paths here.

BATCH-06-COMPLETE
