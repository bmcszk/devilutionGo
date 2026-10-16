---
# devilGo-rctk
title: Town NPCs + stores + dialog catalog (towners/quest_dialog) + price formulas
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:05Z
parent: devilGo-nxxo
blocked_by:
    - devilGo-jx08
---

# Town: NPCs, store catalogs, dialog matrix, price rules

## Goal
Transcribe the town domain: `towners.tsv` (12), `quest_dialog.tsv` (10x18), the 24-state `TalkID`
set, the seven store catalogs with their caps/generation/pricing formulas, and the animOrder
sequences — as data + a pure price/stock rule table.

## Research
- `/tmp/dg-requirements-append.md` **V5.1** (towners rows, loader `tables/townerdat.cpp:100`),
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
Transaction *state machine* stays in dg-task-025.

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Towners` = **12** base / **15** hf (+3 appended, 3
      modified), `QuestDialog` = **10** rows x 18 quest cols base / **13** rows x 21 cols hf
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
- [ ] Dialog matrix: every one of the 180 base cells (10x18) and 273 hf cells (13x21) equals the
      TSV `TEXT_*` id or `TEXT_NONE`, and the mushroom-quest forced cells
      (`TOWN_HEALER/Q_MUSHROOM = TEXT_NONE`, `TOWN_WITCH/Q_MUSHROOM = TEXT_MUSH9`,
      `quests.cpp:61-62`) are honoured.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — table-driven pricing, ints only
- /home/blaze/.agents/skills/coding-rules/SKILL.md — keep price rules separate from UI paging
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — seeded vendor stock fixtures
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — 20-line store screen paging (only for
  the UI-facing struct shape)

## Dependencies & blockers
blocked_by dg-task-101. Needs dg-task-102 (item ids) and dg-task-103 (generation fns called by
vendor spawn). Blocks dg-task-109 (cow farmer/NPC swap lives in towners), dg-task-112 (quest
dialog drives quest FSM text), dg-task-116 (menu text), dg-task-118 (audit), and existing
dg-task-025/030.

## Open questions
- `animOrder` for the 12 hf/base towner rows is "100+ entries per NPC" (dg-task-020 says so);
  confirm the exact per-row sequence lengths from V5.1b before asserting a count.
- Is `quest_dialog` best modelled as a sparse map (most cells `TEXT_NONE`) or a dense 13x21
  array? Dense is parity-correct but 273 cells of mostly-none; recommend dense + tests.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:hf-quest_dialog-shape-wrong (21/3/273 → 25 cols/6 extra/312 cells, disk-verified)
