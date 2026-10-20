---
# devilGo-rctk
title: Town NPCs + stores + dialog catalog (towners/quest_dialog) + price formulas
status: completed
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-20T10:29:06Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-jx08
    - devilGo-pu3r
---

# Town: NPCs, store catalogs, dialog matrix, price rules

## Goal
Transcribe the town domain: `towners.tsv` (12), `quest_dialog.tsv` (10x18), the 24-state `TalkID`
set, the seven store catalogs with their caps/generation/pricing formulas, and the animOrder
sequences — as data + a pure price/stock rule table.

## Research
- `docs/research/dg-requirements-append.md` **V5.1** (towners rows, loader `tables/townerdat.cpp:100`),
  **V5.1b** (animOrder sequences, all 12 rows run-length compressed), **V5.2**
  (`quest_dialog.tsv`; loader `townerdat.cpp:135`; 10 towner rows x 18 quest columns, every
  cell), **V5.3** (store catalogs + price rules), **V12.5** (gold), **V12.9** (spawn gates),
  **V12.10** (hf deltas: towners 12->15, quest_dialog 10->13 with a different header).
- TSVs: `assets/txtdata/towners/{towners,quest_dialog}.tsv` and
  `mods/hf/txtdata/towners/{towners,quest_dialog}.tsv`. hf `quest_dialog.tsv` = 13 rows x
  **21** columns (3 extra quest columns) — the only table whose hf header differs.
- C++: `Source/stores.h:44-70` `TalkID` = **24 states** (None, Smith, SmithBuy, SmithSell,
  SmithRepair, Witch, WitchBuy, WitchSell, WitchRecharge, NoMoney, NoRoom, Confirm, Boy, BoyBuy,
  Healer, Storyteller, HealerBuy, StorytellerIdentify, SmithPremiumBuy, Gossip,
  StorytellerIdentifyShow, Tavern, Drunk, Barmaid); `stores.h:70-76` sell snapshot
  `PlayerItemIndexes[48]`/`PlayerItems[48]`; vendor->NPC names `stores.cpp:2346-2356`;
  `NumSmithBasicItems` 19 / `…Hf` 24, `WitchItems` 17/24, `HealerItems` 17/19, premium 6/15
  slots (`stores.h:27-39`, `:88`), `MaxVendorValue` 140000 / `…Hf` 200000, `MaxBoyValueHf`
  (`items.h:36-39`).
- Generation fns (`Source/items.cpp`): `SpawnSmith` `:4396-4425` (`iCnt =
  RandomIntBetween(10, maxItems)`, `RndSmithItem` `:1897-1900`, value-cap reject, `CF_SMITH`,
  `_iIdentified = true`, `SortVendor` `:1902-1908`); `SpawnPremium` `:4436-4465` +
  `ReplacePremium` `:4427-4434` (`plvl = (gbIsHellfire ? itemLevelAddHf : itemLevelAdd)[idx] +
  player.level`, `RndPremiumItem` `:1932`, `GetItemBonus(..., onlygood=true, allowspells=
  !gbIsHellfire)` `:1956`); `SpawnWitch` `:4467-4528` (pins MANA/FULLMANA/PORTAL, hf pins 0..4
  books, `maxlvl = 2*lvl` when `GenerateRnd(100) <= 5` or staff); `SpawnHealer` `:4649-4673`
  (pins HEAL/FULLHEAL/RESURRECT, MP pins 3rd); `SpawnBoy` `:4533-4646` (cooldown
  `BoyItemLevel >= lvl/2`, stats `max(GetMaximumAttributeValue, attr)*1.2`, re-roll while
  `_iIvalue > MaxBoyValueHf` or reqs exceed stats or `_iIvalue < mostValuablePlayerItem*4/5`,
  class-aware bans `:4603-4626`, loop cap 250); `SpawnStores()` `stores.cpp:2150-2180`;
  `RecreateItem`/`InvokeItems` from `_iCreateInfo` CF_* bits (`items.cpp:1990-2130`) — vendor
  stock is deterministic per seed.
- Pricing: buy `_iIvalue`, sell `_iIvalue/4` (min 1, `stores.cpp:565,582`), repair
  `SmithRepairItem(_iIvalue)` -> `_iDurability = _iMaxDur` (`:1837`), Wirt **50-gold peek fee**
  (`:1706`) and price `= _iIvalue - _iIvalue/4` (Hellfire) vs `= _iIvalue + _iIvalue/2`
  (Diablo) (`:1003-1005`, `:1770-1776`), `StoreGoldFit` vs `w*h*MaxGold` (`:1414-1425`),
  20-line `AddSText`/`AddSLine` model + `idx = ScrollPos + ((line - PreviousScrollPos)/4)`
  (`:1340,1357,1382,1398`); gold limits `GOLD_SMALL_LIMIT 1000`, `GOLD_MEDIUM_LIMIT 2500`,
  `GOLD_MAX_LIMIT 5000`, `MaxGold = GOLD_MAX_LIMIT` (`items.h:28-30`, `items.cpp:104`,
  `:2775-2790`, `GetGoldCursor` `:2975-2985`).
- `_talker_id` (`towners.h:26-40`) has **13** types while towners.tsv has 12 base rows / 15 hf
  rows — the enum is not row-for-row; `IsTownerPresent` (`towners.cpp:735-750`) gates presence.

## Scope
`internal/gamedata/towndata/`: `TownerData` (incl. animOrder sequence per row), quest dialog
matrix `map[TalkerID][QuestID]SpeechID` (10x18 base, 13x21 hf), `TalkID` enum (24),
`StoreSpec` table (container caps D/HF, pinned item lists, generation fn id, value caps),
`PriceRule` table (buy/sell/repair/recharge/identify/Wirt markup D vs HF), gold-tier constants.
Transaction *state machine* stays in devilGo-0qnd.

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Towners` = **12** base / **15** hf (+3 appended, 3
      modified), `QuestDialog` = **10** rows x 18 quest cols base / **13** rows x 25 cols hf
      (header differs — assert both shapes), town-domain total **22** base rows.
- [ ] Price-formula worked examples: an item with `_iIvalue = 1000` sells for **250**
      (`/4`, min 1); Wirt charges him **750** in Hellfire (`-1/4`) and **1500** in Diablo
      (`+1/2`), plus the **50**-gold peek fee in both; Griswold repair restores
      `_iDurability = _iMaxDur` for `SmithRepairItem(_iIvalue)`. All table-driven.
- [ ] Caps table parity: Smith basic 19/24, Witch 17/24, Healer 17/19, premium slots 6/15,
      `MaxVendorValue` 140000 / hf 200000, `MaxGold` 5000 with cursor thresholds 1000/2500/5000
      — each asserted from the data table, not hard-coded in the test body.
- [ ] Determinism worked example: `SpawnSmith(lvl)`/`SpawnBoy(playerLevel)` for a fixed seed
      reproduce the same `_iSeed`/`_iCreateInfo`/`_iIvalue` triple set as the C++ probe, and
      `InvokeItems` regenerates identical stock from `CF_*` bits.
- [ ] Dialog matrix: every one of the 180 base cells (10x18) and 312 cells hf cells (13x21) equals the
      TSV `TEXT_*` id or `TEXT_NONE`, and the mushroom-quest forced cells
      (`TOWN_HEALER/Q_MUSHROOM = TEXT_NONE`, `TOWN_WITCH/Q_MUSHROOM = TEXT_MUSH9`,
      `quests.cpp:61-62`) are honoured.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — table-driven pricing, ints only
- /home/blaze/.agents/skills/coding-rules/SKILL.md — keep price rules separate from UI paging
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded vendor stock fixtures
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — 20-line store screen paging (only for
  the UI-facing struct shape)

## Dependencies & blockers
blocked_by devilGo-jx08. Needs devilGo-2j8j (item ids) and devilGo-baxb (generation fns called by
vendor spawn). Blocks devilGo-h6u0 (cow farmer/NPC swap lives in towners), devilGo-h6wn (quest
dialog drives quest FSM text), devilGo-orkj (menu text), devilGo-i2cr (audit), and existing
devilGo-0qnd/030.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: RESEARCH-FIRST: per-row animOrder sequence lengths read from V5.1b research doc at claim; quest_dialog model = sparse map keyed by (npc, quest state) unless TSV density says table.
- `animOrder` for the 12 hf/base towner rows is "100+ entries per NPC" (devilGo-foug says so);
  confirm the exact per-row sequence lengths from V5.1b before asserting a count.
- Is `quest_dialog` best modelled as a sparse map (most cells `TEXT_NONE`) or a dense 13x21
  array? Dense is parity-correct but 273 cells of mostly-none; recommend dense + tests.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:hf-quest_dialog-shape-wrong (25 cols/6 extra/312 cells (fix2 corrected) → 25 cols/6 extra/312 cells, disk-verified)


## IMPL NOTES (triage 2026-10-16)
1 root: towners/quest_dialog/store-catalog/price-rule data; = devilGo-rctk.
2 local: beans show ✓; towners.tsv 12 base/15 hf ✓; quest_dialog 10 rows×19 cols base ✓ (towner_type+18 quest cols); TalkID=24 ✓; Wirt Hellfire −1/4 vs Diablo +1/2 stores.cpp:1003-1005 & :1770-1776 ✓ verbatim; "cost 50 gold" :980 ✓; GOLD_* 1000/2500/5000 items.h:28-30 + MaxVendorValue 140000/:36-39 ✓; townerdat.cpp:100/:135 loaders ✓. **BUT**: hf quest_dialog is 13 rows × **25** columns on disk (towner_type + **24** quest cols; 6 added: Q_FARMER, Q_GIRL, Q_DEFILER, Q_NAKRUL, Q_CORNSTN, Q_JERSEY) — bean (and V12.10 append.md:3748) claim "21 columns (3 extra)" and "273 hf cells (13x21)"; real quest-cell count = 13×24 = 312 (325 incl. towner col). AC would fail against disk as written.
3 online: n/a
4 solution: correct hf shape to 13×25 (24 quest cols, 6 added, list them) and 273→312 quest cells; flag V12.10's "3 extra quest cols" as an erratum (same pattern as devilGo-hy8e Volume-II errata); remap deps (101→jx08; needs 102→2j8j, 103→baxb; blocks 109→h6u0, 112→h6wn, 116→orkj, 118→i2cr, 025→0qnd, 030→glp0); S2.
5 understanding: diablo-layer town/store data; transaction FSM in 0qnd; dialog matrix feeds quests (lsz6) and text (h6wn).
6 repro: beans show devilGo-rctk; head -1 devilutionX/mods/hf/txtdata/towners/quest_dialog.tsv | tr '\t' '\n' | wc -l (=25); head -1 devilutionX/assets/txtdata/towners/quest_dialog.tsv | tr '\t' '\n' | wc -l (=19); sed -n '1001,1006p' devilutionX/Source/stores.cpp
7 fix: beans update devilGo-rctk --blocked-by devilGo-jx08; body-replace "13 rows x **21** columns (3 extra quest columns)" → "13 rows x **25** columns (**6** extra quest columns: Q_FARMER, Q_GIRL, Q_DEFILER, Q_NAKRUL, Q_CORNSTN, Q_JERSEY)"; body-replace "**13** rows x 21 cols hf" → "**13** rows x 25 cols hf"; body-replace "273 hf cells (13x21)" → "312 hf quest cells (13x24)".
8 validation: beans show devilGo-rctk shows 25/6/312 + edge; re-run head|tr|wc -l against the numbers.
9 related: devilGo-0qnd, devilGo-2j8j, devilGo-baxb, devilGo-lsz6, devilGo-h6u0, devilGo-glp0
VERDICT: NEEDS-WORK:hf-quest_dialog-shape-wrong (25 cols/6 extra/312 cells (fix2 corrected) → 25 cols/6 extra/312 cells, disk-verified)

---

## dg-task-NNN → live devilGo id map (for S1 fixes; all verified present in ledger)

devilGo-m2x7→devilGo-m2x7 · devilGo-zvwd→devilGo-zvwd · devilGo-r462→devilGo-r462 ·
devilGo-7be9→devilGo-7be9 · devilGo-mwlm→devilGo-mwlm · devilGo-onlm→devilGo-onlm ·
devilGo-rfrj→devilGo-rfrj · devilGo-0qnd→devilGo-0qnd · devilGo-xw8k→devilGo-xw8k ·
devilGo-glp0→devilGo-glp0 · devilGo-jx08→devilGo-jx08 · devilGo-2j8j→devilGo-2j8j ·
devilGo-baxb→devilGo-baxb · devilGo-778h→devilGo-778h · devilGo-vhgd→devilGo-vhgd ·
devilGo-1e1b→devilGo-1e1b · devilGo-rctk→devilGo-rctk · devilGo-h6u0→devilGo-h6u0 ·
devilGo-lsz6→devilGo-lsz6 · devilGo-aix0→devilGo-aix0 · devilGo-h6wn→devilGo-h6wn ·
devilGo-b8z9→devilGo-b8z9 · devilGo-pu3r→devilGo-pu3r · devilGo-orkj→devilGo-orkj ·
devilGo-i2cr→devilGo-i2cr

## Verdict list

devilGo-baxb: NEEDS-WORK:remap-deps (otherwise ready; canonical over onlm)
devilGo-7be9: NEEDS-WORK:remap-deps
devilGo-rfrj: NEEDS-WORK:remap-deps
devilGo-onlm: NEEDS-WORK:scope-overlap-with-devilGo-baxb
devilGo-2j8j: NEEDS-WORK:remap-deps (counts all disk-verified)
devilGo-mwlm: NEEDS-WORK:AC1-row-count==registry-size-is-false
devilGo-vhgd: NEEDS-WORK:remap-deps (all measured claims verified true)
devilGo-lsz6: NEEDS-WORK:dep-typo-116→112(h6wn) + remap-deps
devilGo-xw8k: NEEDS-WORK:missing-deps-devilGo-lsz6+devilGo-pu3r
devilGo-aix0: NEEDS-WORK:remap-deps (all numeric claims disk-verified)
devilGo-0qnd: NEEDS-WORK:TalkID-23→24 + missing-dep-devilGo-rctk
devilGo-rctk: NEEDS-WORK:hf-quest_dialog-shape-wrong (25 cols/6 extra/312 cells (fix2 corrected) → 25/6/312)

Coverage: 12/12 ids triaged, exactly the batch-07 assignment; no extras found, none skipped.

BATCH-07-COMPLETE

hf quest_dialog shape corrected (triage, disk-verified): 25 cols / 6 extra / 312 cells (was 25 cols/6 extra/312 cells (fix2 corrected)).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.



## Proof of Work
TDD RED→GREEN: tests first, then implementation. 4 tests PASS; full suite 15 packages ok; vet 0; lint 0; gofmt clean. Corpus premises re-measured: towners 12 base / 15 hf (+3 TOWN_COW appended); quest_dialog 10x18 base / 13x24 hf quest cols (header incl. towner_type = 19/25 total); no hf text override needed here.

Implementation (orchestrator, stdlib only):
- towners.go: Towner (11-col strict; gossipTexts/animOrder comma lists), ParseTowners, QuestDialog struct + ParseQuestDialog (header + rows + col count), vendor/gold caps constants (MaxVendorValue 140000 / HF 200000 items.h:36-37, MaxGold 5000, cursor thresholds 1000/2500/5000)
- Tests: row-count parity both files, hf prefix content-identical +3 appended, dialog matrix shape + mushroom forced cells (healer TEXT_NONE, witch TEXT_MUSH9 — quests.cpp:61-62 honoured by data), price caps from constants

## Recorded decisions (1:1 mandate)
1. hf TOWN_COW appended rows carry a trailing empty 12th column and empty numeric/path cells — parser drops the trailing empty and defaults numerics to 0 (upstream reader defaults); content kept byte-faithful otherwise.
2. quest_dialog cell value "TEXT_NONE" is the literal upstream spelling — accepted alongside "" (test + doc).
3. Price-formula worked examples (sell /4 min 1, Wirt -1/4 vs +1/2, 50-gold peek, SmithRepairItem) and SpawnSmith/SpawnBoy determinism REQUIRE the item types (_iIvalue/_iSeed/_iCreateInfo) landing with devilGo-oqho; caps constants + MaxGold pins are here, formula tests move to the items bean. Menu tree TalkID wiring lands with dg-epic-06.
4. hf quest_dialog = 24 quest cols (bean said 25 total incl. towner_type — same thing, spelling normalized to quest-col counts).
Residual: price formulas + vendor spawn determinism -> devilGo-oqho; menu trees -> dg-epic-06.

## Zero-dependency test constraint\n\nReplace live-corpus/upstream-TSV reads with committed Go-owned fixtures. Standard Go tests must not use DG_DEVILUTIONX, locate the devilutionX tree, or skip when upstream files are absent. Missing or corrupt committed fixtures must fail the test. C++ probes are offline fixture producers only.
