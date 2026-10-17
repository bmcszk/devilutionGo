---
# devilGo-778h
title: 'Monsters data package: monstdat+unique_monstdat (212 rows) + AiProc registry'
status: completed
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-17T17:17:59Z
parent: devilGo-97mn
blocked_by:
    - devilGo-jx08
---

# Monsters: 212 rows + AI family registry

## Goal
Transcribe `monstdat` (112) and `unique_monstdat` (100) into Go tables **and** the `AiProc[]`
function-pointer table as a Go registry keyed by `MonsterAIID`, with the exact null/non-null
shape of the C++.

## Research
- `docs/research/dg-requirements-append.md` **V2.1** (41 columns; loader `Source/tables/monstdat.cpp:409`;
  `MonsterData` monstdat.h:100-160), **V2.2** (100 unique rows; loader `:459`), **V2.3**
  (AiProc table), **V2.4** (pack/spawn rules), **V12.2** (enums), **V12.12** (counts).
- TSVs: `assets/txtdata/monsters/{monstdat,unique_monstdat}.tsv`,
  `mods/hf/txtdata/monsters/monstdat.tsv` (unique_monstdat has **no** hf override).
- Enums (`Source/tables/monstdat.h`): `MonsterAIID` `:22-63` = **40 named entries** +
  `Invalid = -1`; `MonsterClass` `:66-70` (Undead, Demon, Animal); `MonsterAvailability`
  `:92-95`; treasure bits `:86-88` (`T_MASK=0xFFF`, `T_NODROP=0x4000`, `T_UNIQ=0x8000`);
  `Uniq(...)` written by `monstdat.cpp:39`.
- **Correction to V2.3**: `Source/monster.cpp:3089-3130` `AiProc[]` has **40 entries**
  (`grep -c MonsterAIID` on that range = 40), 39 real function pointers plus **one `nullptr`**
  at `MonsterAIID::FireMan` (`:3107` — elemental behaviour lives in the missile code). The
  inventory's V2.3 table lists 41 rows because it counts the `Invalid = -1` sentinel as
  `AiProc[40]`; there is no index 40.
- Distinct `ai` values actually referenced: **24** in base monstdat, **31** in hf monstdat.
- Ranged missile selection `GetMissileType(MonsterAIID)` (`monster.cpp:1941-1972`) and
  `toHitSpecial` / `min/maxDamageSpecial` columns.

## Scope
`internal/gamedata/monstdat/`: `MonsterData` (all parsed columns, exact order),
`UniqueMonsterData`, `MonsterAIID`/`MonsterClass`/`MonsterAvailability` enums, treasure-bit
helpers, and `var AiRegistry = map[MonsterAIID]AIRuntime` where the `FireMan` key maps to a
**nil** implementation (so "no AI" is data, not an accident) and every registry entry has
`Name()` for diagnostics. Behaviour implementations themselves land in devilGo-v791/epic-04;
this bean ships the registry *shape* + names.

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Monstdat` = **112** base / **138** hf (+26 appended,
      0 modified), `UniqueMonstdat` = **100** base = hf (no override); monster-domain total
      **212** base rows; the 26 hf rows are exactly the V12.10 list (`MT_HELLBOAR` ...
      `MT_NAKRUL`).
- [ ] AI registry parity: `AiProc` has **40** entries with exactly **one** nil (`FireMan`), and
      every distinct `ai` value used by base+hf monstdat (31) resolves to a registry key; a test
      enumerates `MonsterAIID` and fails on any unmapped or extra entry.
- [ ] Worked example: the two monstdat rows carrying `Uniq(...)` treasure (e.g. `MT_CLEAVER`)
      decode to `T_UNIQ | id` and `dropsSpecialTreasure` is true; `T_NODROP` rows produce no
      loot — asserted from the table, not from generation code.
- [ ] hf diff content test: hf monstdat == base monstdat for the first 112 rows (0 modified) and
      appends 26, verified by a byte-level diff against both files.
- [ ] `GetMissileType` mapping table asserted for all documented pairs (GoatRanged->Arrow,
      Succubus/LazarusSuccubus->BloodStar, Acid/AcidUnique->Acid, FireBat->Firebolt,
      Torchant->Fireball, Lich->OrangeFlare, ArchLich->YellowFlare, Psychorb->BlueFlare,
      Necromorb->RedFlare, Magma->MagmaBall, Storm->ThinLightningControl,
      Diablo->DiabloApocalypse, BoneDemon->BlueFlare2, default->Arrow).
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — registry pattern replacing fn pointers
- /home/blaze/.agents/skills/coding-rules/SKILL.md — registry (data) vs AI (behaviour) separation
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — table-driven enum tests

## Dependencies & blockers
blocked_by devilGo-jx08. Blocks devilGo-v791 (pack placement consumes these tables), devilGo-1e1b
(missile ids referenced by monster AI must exist), devilGo-b8z9 (monster voices -> audio),
devilGo-i2cr (audit), and existing devilGo-foug (population).

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: no hf variant of unique_monstdat exists or will be invented (disk shows 100=100); Hellfire uniques are base-table rows. Availability band naming: follow monster.cpp enum names verbatim (1:1).
- Should `unique_monstdat` get an hf variant later? None exists on disk (100 = 100) — confirm
  Hellfire uniques are all already in the base table rather than missing data.
- `MonsterAvailability` Hellfire band naming in V12.2 ("(Hellfire band)") is vague; read
  monstdat.h:92-95 and pin the exact identifiers.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps — 3 declared blocks edges missing; row counts and AiProc shape verified exact.


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-7 data task (devilGo-778h); draft; wire blockedBy [jx08] ✓ = body's devilGo-jx08.
2 local: Row counts LIVE-VERIFIED: monstdat.tsv 113 lines=112 data ✓, unique_monstdat 101=100 ✓, hf monstdat 139=138 (+26 appended) ✓, hf has NO unique_monstdat override ✓ (ls shows only monstdat.tsv) — "212 base rows" and "100 base = hf" both hold. AiProc: `void (*AiProc[])(Monster&)` monster.cpp:3090, range 3089-3130 grep = 40 MonsterAIID entries ✓ with exactly ONE nullptr at FireMan (offset :3109; body ":3107" — 2-line drift); MonsterAIID enum has 40 named + `Invalid = -1` ✓ (bean's V2.3 correction is right). Loader cites monstdat.cpp:409 and :459 ✓ exact (sed match). GetMissileType at :1940 (body ":1941-1972" ✓ ~1-line); MonsterData struct monstdat.h:98 (body ":100-160" ~2-line drift). Wire gap: body "Blocks devilGo-v791 (✓ v791 wired), 107, 113, 118 (✓ i2cr wired), 020" — 1e1b(107), b8z9(113), foug(020) have no 778h edge (all blockedBy jx08 / 4ybb only).
3 online: n/a.
4 solution: wire 3 missing edges; optional cite refreshes (3107→3109, 100-160→98-160).
5 understanding: the monster-domain root table for v791 placement, 1e1b missile refs, b8z9 voices, foug population, i2cr audit; the nil-FireMan-as-data design decision is a good Go translation of the C++ fn-pointer table.
6 repro: `awk 'END{print NR}' devilutionX/{assets/txtdata,mods/hf/txtdata}/monsters/monstdat.tsv devilutionX/assets/txtdata/monsters/unique_monstdat.tsv; ls devilutionX/mods/hf/txtdata/monsters/; sed -n '3089,3131p' devilutionX/Source/monster.cpp | grep -c MonsterAIID; sed -n '3089,3131p' devilutionX/Source/monster.cpp | grep -n nullptr`
7 fix: `beans update devilGo-1e1b --blocked-by devilGo-778h`; `beans update devilGo-b8z9 --blocked-by devilGo-778h`; `beans update devilGo-foug --blocked-by devilGo-778h`
8 validation: each dependent's blockedByIds includes 778h.
9 related: jx08 (dep), v791/1e1b/b8z9/foug/i2cr (dependents), h6u0 (26 hf rows consumer).
VERDICT: NEEDS-WORK:wire-deps — 3 declared blocks edges missing; row counts and AiProc shape verified exact.

Fix pass 2026-10-16: Dep-wiring applied (fix pass): 3 declared blocks edges now wired in tracker.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.



## Proof of Work
TDD RED→GREEN: tests first (undefined symbols), then implementation. 5 tests PASS; full suite 9 packages ok; vet 0; lint 0; gofmt clean. Corpus premises re-verified on disk before coding: monstdat 112 base / 138 hf (+26 appended), unique 100, 41 cols, AiProc 40 entries with FireMan nil, 31 distinct non-empty ai values.

Implementation (orchestrator, stdlib only):
- monstdat.go: MonsterData (41-col strict TSV, Raw() byte-parity accessor), MonsterAIID enum verbatim (monstdat.h:22-63, Invalid=-1), treasure flags T_MASK/T_NODROP/T_UNIQ (:86-88), parseTreasure exact port of ParseMonsterTreasure (monstdat.cpp:299-310: ""→0, "None"→T_NODROP, Uniq(SKCROWN)/Uniq(CLEAVER)→T_UNIQ|UITEM_x, else error), ParseUniqueMonstdat (raw rows, no hf override), AiProc registry 40 entries with exactly one nil at FireMan (monster.cpp:3106), AiProcByName (nil-resolving for FireMan), GetMissileType mapping (all 14 documented pairs + Arrow default)
- Tests: row-count parity vs live corpus; hf strict-superset (shared-prefix row-identical via Raw(), +26 appended, no base-id dupes); AI registry parity (40 entries, 1 nil at FireMan, 31 distinct data ai all resolve); treasure decode (2 T_UNIQ rows = MT_SKING+MT_CLEAVER, 4 T_NODROP = Fiend/Blink/Gloom/Familiar, MT_CLEAVER = T_UNIQ|UITEM_CLEAVER bit-exact); GetMissileType all pairs

## Recorded decisions (1:1 mandate)
1. AI behaviour fns are named placeholders (registry SHAPE + null-ness is this bean's contract; real AI ports belong to the monster-engine bean, matching vhgd's stub precedent).
2. GetMissileType hardcoded map: upstream is a switch in monster.cpp (same shape); the hf-only monsters resolve via the default arm exactly as upstream.
3. ParseMonsterTreasure ported as upstream's incomplete parser (only SKCROWN/CLEAVER Uniq cases; anything else errors) — extending it would be inventing behaviour; upstream TODO noted at monstdat.cpp:301.
Residual: unique_monstdat rows stay raw (decode lands with the unique-monster consumer); AI fns are placeholders by scope.
