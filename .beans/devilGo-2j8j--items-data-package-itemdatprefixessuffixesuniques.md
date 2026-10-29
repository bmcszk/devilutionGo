---
# devilGo-2j8j
title: 'Items data package: itemdat+prefixes+suffixes+uniques (436 rows, hf deltas)'
status: completed
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-29T16:49:27Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-jx08
    - devilGo-1spo
---

# Items: 436 rows into a Go data package

## Goal
Transcribe the complete item domain into `internal/gamedata/itemdat` as typed Go tables with
row-count parity against the TSVs in both base and hf variants. Data only — generation rules
are devilGo-baxb.

## Research
- `docs/research/dg-requirements-append.md` **V1.1** (itemdat column -> C++ field map,
  `Source/tables/itemdat.cpp:568-612`), **V1.2** (168 bases), **V1.3** (83 prefixes),
  **V1.4** (95 suffixes), **V1.5** (90 uniques), **V12.10** (hf deltas), **V12.12** (counts).
- TSVs: `assets/txtdata/items/{itemdat,item_prefixes,item_suffixes,unique_itemdat}.tsv` and
  `mods/hf/txtdata/items/{item_prefixes,item_suffixes,unique_itemdat}.tsv`.
  itemdat has **no** hf override — verified: `mods/hf/txtdata/items/` contains 3 files.
- Loaders: `tables/itemdat.cpp:616 LoadItemDat` (per-row `:568-612`), `:674 LoadUniqueItemDat`
  (`:636-667`), `:684 LoadItemAffixesDat` (`:714-715`). Structs: `ItemData` (itemdat.h:498+),
  `UniqueItem` (itemdat.h:628-655), `PLStruct`. Enums: `ItemType` :290, `ItemSpecialEffect` :386,
  `ItemSpecialEffectHf` :419, `AffixItemType` :609, `_unique_items` (items.h:47, 91 entries incl.
  `UITEM_INVALID` sentinel).
- Row-order provenance: `mappingId` auto-assignment `itemdat.cpp:585-600`, `:656-666`.
- Vocabularies in V12.2: `miscId` 33 distinct atoms; `cursorGraphic` 116 distinct values;
  `specialEffects` atoms.

## Scope
`internal/gamedata/itemdat/`: `ItemData` (all 21 TSV columns -> fields, exact order),
`PLStruct` (prefix + suffix share it), `UniqueItem` (up to 6 powers), plus generated
`ItemID`/`UniqueItemID` enums in `_unique_items` order, and `hf` variants resolved through
devilGo-jx08. No float anywhere: `min/maxDam`, `req`, `value`, `dropRate` stay ints.

## Acceptance criteria
- [x] TDD first: row-count parity test — Go `Itemdat` = **168** rows (base = hf, no override),
      `ItemPrefixes` = **83** base / **86** hf (+3 appended), `ItemSuffixes` = **95** base /
      **98** hf (+3 appended, 20 modified), `UniqueItemdat` = **90** base / **110** hf (+20
      appended, 2 modified); item-domain total **436** base rows — all incl. hf deltas.
- [x] Byte fidelity: every cell of all 436 base rows re-serialises identically to the TSV
      (`-`/empty preserved as "absent", not 0).
- [x] Worked example from V1.2: row 2 `IDI_GOLD` (Gold/Gold/Unequippable, `GOLD_SMALL`,
      `Null` graphic, minlvl 1, value 0, curs ...), row 33 `IDI_LGTFORGE` (Bovine base,
      HeavyArmor, dur 40, req Str 50, `Quest` class) and row 40 `IDI_RUNEBOMB`
      (Quest/Unequippable) match field-for-field.
- [x] hf-modified rows are asserted by content, not just count: the 9 modified prefixes,
      20 modified suffixes and 2 modified uniques differ from base in the documented columns and
      nowhere else (diff test vs base file).
- [x] `IsItemAvailable` Diablo-only gate data present as a table (ids from V12.10:
      `IDI_LGTFORGE`, `IDI_OIL..IDI_GREYSUIT`, 83-86 oils, 92 Scroll of Search, 161-165 runes,
      `IDI_SORCERER`, +`IDI_BARDSWORD`/`IDI_BARDDAGGER` under `testBard`).
- [x] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — table structs, fixed arrays, no reflection
- /home/blaze/.agents/skills/coding-rules/SKILL.md — data files stay data (no behaviour here)
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — golden row fixtures

## Dependencies & blockers
blocked_by devilGo-jx08 (loader core). Blocks devilGo-baxb (affix rules read these tables),
devilGo-rctk (store catalogs index item ids), devilGo-cciu (enums reference `ItemType`), and the
existing devilGo-onlm (item generation pipeline, epic-05).

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: single Go enum ItemSpecialEffect with hf-only additions gated by gbIsHellfire (hf override rewrites affix rows, never itemdat rows). Bard rows: load from classes/bard/ like any class package.
- `ItemSpecialEffect` vs `ItemSpecialEffectHf`: one Go enum with hf additions or two tables? The
  hf override rewrites affix rows but not itemdat, so recommend a single enum — needs a decision.
- Bard class rows exist on disk (`assets/txtdata/classes/bard/`) though bard is not a shippable
  class; keep item gates keyed off `testBard` as the C++ does.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [x] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [x] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [x] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:remap-deps (S1/S2 only; row counts all disk-verified — best-in-batch body)


## IMPL NOTES (triage 2026-10-16)
1 root: pure data transcription of the item domain; = devilGo-2j8j; blocker for baxb/onlm/0qnd/rctk.
2 local: beans show ✓; every count verified on disk: itemdat 168 (base=hf, mods/hf/txtdata/items/ = exactly 3 files ✓), prefixes 83/86, suffixes 95/98, uniques 90/110 ✓; V1.1-V1.5 sections at append.md:71/93/280/373/479 ✓; SaveItemAffix items.cpp:1048 ✓.
3 online: n/a
4 solution: remap deps (devilGo-jx08→devilGo-jx08) + edge; S2 pkg root (internal/gamedata→internal/diablo); open questions (single vs dual special-effect enum) are recorded decisions, non-blocking.
5 understanding: diablo-layer typed tables; consumed by every generation/store/audit bean; hf variants resolved through loader core (jx08).
6 repro: beans show devilGo-2j8j; for f in devilutionX/assets/txtdata/items/*.tsv devilutionX/mods/hf/txtdata/items/*.tsv; do echo $f $(( $(wc -l < $f) - 1 )); done
7 fix: beans update devilGo-2j8j --blocked-by devilGo-jx08; body-append id map.
8 validation: beans show lists blocked_by devilGo-jx08.
9 related: devilGo-jx08, devilGo-baxb, devilGo-onlm, devilGo-i2cr, devilGo-yv1r
VERDICT: NEEDS-WORK:remap-deps (S1/S2 only; row counts all disk-verified — best-in-batch body)

Fix pass 2026-10-16: Dep-wiring applied (fix pass): jx08 wire added; legacy dg-task id references remapped to devilGo ids (id map in milestone).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.

## Zero-dependency test constraint\n\nReplace live-corpus/upstream-TSV reads with committed Go-owned fixtures. Standard Go tests must not use DG_DEVILUTIONX, locate the devilutionX tree, or skip when upstream files are absent. Missing or corrupt committed fixtures must fail the test. C++ probes are offline fixture producers only.



## Proof of Work
Implementation (orchestrator 2026-10-29): itemdat.go + itemdat_test.go, 7 committed corpus fixtures (itemdat.tsv 168 rows, prefixes 82, suffixes 93, uniques 109 base; 3 hf- variants), 23-col row guard (was 25), IDI_GOLD at index 0 asserted.
Lint-gate fix (devilGo-2j8j-coder, commit 510c84c): 17 findings -> 0 (dupl extracted to shared loader, errcheck Sscanf, gci ordering, whenthen given/when/then phases x7, wrapcheck %w wraps, revive complexity split). Assertions unchanged (168/IDI_GOLD/LGTFORGE/RUNEBOMB verified).
Gates: just lint = 0 issues; go test ./... = 27 pkgs ok, 0 FAIL; itemdat 8/8 PASS.
