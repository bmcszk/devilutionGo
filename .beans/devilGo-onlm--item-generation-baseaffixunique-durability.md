---
# devilGo-onlm
title: 'Item generation: base/affix/unique + durability'
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - items
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-r462
    - devilGo-baxb
---

---
# devilGo-onlm
title: 'Item generation: base/affix/unique + durability'
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-2
    - items
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-21T11:07:45Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-r462
    - devilGo-baxb
---

# Itemgen

## Goal
SpawnItem/SpawnUnique/affix rolls parity.

## Research
items.cpp:1048,1138,1218-1224,1475,3195,3422; itemdat.h:628-655; T_MASK treasure decode; tests items_test.cpp.

## Scope
internal/gameplay/itemgen.go + affix application + price calc + durability/break flow (player.cpp:505-522).

## Acceptance criteria
- [ ] 10000-seed histogram parity (epic criterion) for 3 classes
- [ ] Unique availability + min-level gates tested
- [ ] Durability break removes equipment + CalcPlrInv call verified

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/go-tests/SKILL.md — distribution tests

## Dependencies & blockers
devilGo-r462, devilGo-zvwd.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:scope-overlap-with-devilGo-baxb (canonical = baxb; narrow or merge)


## IMPL NOTES (triage 2026-10-16)
1 root: coarse itemgen task (= devilGo-onlm) predating the phase-7 split; scope overlaps devilGo-baxb on SpawnItem/SpawnUnique/affix rolls.
2 local: beans show ✓; SpawnUnique items.cpp:3195 ✓; durability break + RemoveEquipment + CalcPlrInv player.cpp:505-510 ✓; T_MASK/T_NODROP/T_UNIQ monstdat.h:86-88 ✓; deps devilGo-r462→r462 ✓ live, devilGo-zvwd→devilGo-zvwd ✓ live.
3 online: n/a
4 solution: deconflict with baxb — narrow this bean to durability/break flow + price-calc wiring + integration (baxb owns the SpawnItem decision tree + affix algorithms, and says it blocks devilGo-onlm); add blocked_by devilGo-baxb + devilGo-2j8j; remap; S2.
5 understanding: game-layer glue consuming diablo-layer tables (2j8j) and algorithms (baxb); its unique AC overlaps baxb's sharper distribution tests.
6 repro: beans show devilGo-onlm; beans show devilGo-baxb (compare Scope sections); sed -n '505,522p' devilutionX/Source/player.cpp
7 fix: beans update devilGo-onlm --blocked-by devilGo-baxb; then body-replace Scope sentence "internal/gameplay/itemgen.go + affix application + price calc + durability/break flow" → "durability/break flow (player.cpp:505-522), price-calc wiring and integration; SpawnItem/affix algorithms owned by devilGo-baxb".
8 validation: beans show devilGo-onlm shows new edge + narrowed scope; no double-ownership with baxb.
9 related: devilGo-baxb, devilGo-2j8j, devilGo-rfrj, devilGo-0qnd
VERDICT: NEEDS-WORK:scope-overlap-with-devilGo-baxb (canonical = baxb; narrow or merge)


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Seam resolution (triage): scope narrowed to inventory/Talker subsetting ONLY (NOT stores — stores core lives in devilGo-baxb); canonical stores core = devilGo-baxb (single-source pointer).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Event sourcing (ADR-018): state changes this bean introduces go through
   kernel Commands with Gate rules and emitted Events — never direct mutations
   or a private callback bus. New Kind names live in the kernel package
   (law 5), defined per-domain WITH this wiring (no wholesale definition wave,
   no side-channel kinds). Presentation reads events only; it never writes sim
   state.
2. Layering (ADR-008): this bean's package imports only from its own layer and
   below (engine < game < diablo). Check the import direction BEFORE writing
   code; a generic mechanism never lands under internal/diablo, and diablo
   semantics (pcurs ordinals, PLR_MODE, panel flags) never leak into
   internal/game or internal/engine.
