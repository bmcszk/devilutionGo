---
# devilGo-orkj
title: 'Misc enumerations package: cursors, gold, durability, XP caps, restrictions, QoL (V12)'
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:26:53Z
parent: devilGo-yjuv
blocked_by:
    - devilGo-jx08
---

# Misc enumerations and cross-cutting constants

## Goal
Define, in one place, every enum and constant the port needs that is *not* in a TSV: cursors,
dungeon enums not covered elsewhere, gold limits, durability rules, spawn gating, and the QoL
module inventory — so no other bean invents its own constant.

## Research
- `docs/research/dg-requirements-append.md` **V12.1** (cursors), **V12.2** (enum master list), **V12.5**
  (gold), **V12.6** (auto-pickup), **V12.7** (durability/repair), **V12.8** (XP caps), **V12.9**
  (spawn vs full gate list), **V12.11** (QoL modules), **V0.1** (format contract).
- Cursors (`Source/cursor_defs.hpp:14-37`): `SelectionRegion` `:14-19` = None=0,
  Bottom=1<<0, Middle=1<<1, Top=1<<2; `cursor_id` `:22-36` = **all 13 verbatim**: `CURSOR_NONE,
  CURSOR_HAND, CURSOR_IDENTIFY, CURSOR_REPAIR, CURSOR_RECHARGE, CURSOR_DISARM, CURSOR_OIL,
  CURSOR_TELEKINESIS, CURSOR_RESURRECT, CURSOR_TELEPORT, CURSOR_HEALOTHER, CURSOR_HOURGLASS,
  CURSOR_FIRSTITEM`; plus `item_cursor_id` (`items.h`, from itemdat `cursorGraphic` — 116
  distinct values in V1.2 — with the `UICurs` unique override) and gold cursors
  `GOLD_SMALL/MEDIUM/LARGE` via `GetGoldCursor(value)` (`items.cpp:2975-2985`).
  Plumbing: `Source/cursor.cpp` (`SetICursor`/`SetYCursor`/`ClearCursor`), `CheckCursMove`,
  `hwcursor.cpp:112`, software blit `scrollrt.cpp:214`, visibility rule `:236-247`.
- Gold (`V12.5`): `GOLD_SMALL_LIMIT 1000`, `GOLD_MEDIUM_LIMIT 2500`, `GOLD_MAX_LIMIT 5000`
  (`items.h:28-30`); `int MaxGold = GOLD_MAX_LIMIT` (`items.cpp:104`) re-read on split/merge
  (`:2775-2790`, `player.h:973`); `MakeGoldStack` (`~:4676`); `StoreGoldFit` (`stores.cpp:1414-1425`).
- Durability (`V12.7`): `iDurability` from itemdat; jitter `ItemRndDur` (`items.cpp:1475-1478`);
  break checks `player.cpp:474-522` (weapon after hit, shield after block; `DUR_INDESTRUCTIBLE`
  and 0 skip; at 0 => `RemoveEquipment` + `CalcPlrInv`); affixes `IPL_DUR`/`IPL_MODDUR`; repair
  `SpellID::ItemRepair` + `SmithRepairItem`; recharge `SpellID::StaffRecharge`/`WitchRecharge`/
  ShrineStone; hf oils use the same IPL powers (V4.4).
- Spawn gating (`V12.9`, **correction included**): `Source/restrict.cpp` (34 lines) is **not**
  content gating — it only implements `ReadOnlyTest()` (`:19-28`, write+delete
  `Diablo1ReadOnlyTest.foo`, `DirErrorDlg` on failure). All gating is `gbIsSpawn` branches:
  quests (`quests.cpp:94-99` wipe, `:139-141`, `:229`, `:203-205`), monsters
  (`monster.cpp:3159-3167` rejects `availability == Retail`), items (`IsItemAvailable`
  ~`items.cpp:2360-2390`: rejects ids **62-70** medium+heavy armour and **105,107,108,110,111,113**
  unavailable scrolls; `GetBookSpell` clamps book level to 5 `items.cpp:620`; staff spell level
  clamp 10 `:1249`; `SetupAllUseful` branch), spells (`spells.cpp:280-345` skip list, `:74+`),
  heroes (`selhero.cpp:294` locks Rogue/Sorcerer with a purchase message), missiles/music
  (`menu.cpp:36-38` forces INTRO), town warps (`town.cpp:210-234`), saves (`spawn_%d.sv` prefix +
  passwords `adslhfb1`/`lshbkfg1`, `pfile.cpp:60-67`; `pack.cpp` strip rules), cows
  (`towners.cpp:501-505`).
- QoL modules (`Source/qol/`, 16 files, V12.11): autopickup (option radius, gold toggle, belt
  top-up, `autoPickupInTown` default false `options.cpp:862`, `CanUseItem` rule), stash (100
  pages, `GetStashSavePath`, `loadsave.cpp:2740-2760`, `stash_test.cpp`), visual_store
  (`StoreUi::VisualGrid`, `stores.cpp:1701-1705` dispatch), itemlabels, floatingnumbers,
  monhealthbar, xpbar (segments from `Experience.tsv`, `experienceBar` option,
  `player.cpp:2449`), chatlog (+ `control_chat_commands.cpp`), and the rest of the 16 files.
- `panel_button_id` (`control_panel.cpp:131-141`) = Charinfo, Qlog, Automap, Mainmenu, Inventory,
  Spellbook, Sendmsg, Friendly.

## Scope
`internal/gamedata/enums/` + `internal/gamedata/consts/`: cursor enums + `SelectionRegion`
bitset + gold-cursor thresholds; `MaxGold`/gold limits; durability constants + `DUR_INDESTRUCTIBLE`
+ break-rule table; the **spawn gate table** (item id ranges, scroll ids, spell skip list, monster
availability rule, hero locks, save prefixes/passwords) as one auditable structure; QoL module
registry (name, option key, default, logic-affecting y/n); `panel_button_id`. Pure data +
predicates, no behaviour.

## Acceptance criteria
- [ ] TDD first: enumeration parity — `cursor_id` = **13** values in the cited order,
      `SelectionRegion` = 4 (None/Bottom/Middle/Top bit values 0/1/2/4), gold limits =
      {1000, 2500, 5000} with `MaxGold == 5000`, `panel_button_id` = 8, and the QoL registry =
      **16** module files in `Source/qol/` (asserted against the directory listing).
- [ ] Worked example `GetGoldCursor`: value 999 -> `GOLD_SMALL`, 1000 -> `GOLD_MEDIUM`,
      2500 -> `GOLD_LARGE`, 5000 -> `GOLD_LARGE`, 5001 clamps per `MaxGold`; boundary test at
      999/1000/2499/2500/4999/5000.
- [ ] Spawn gate table parity: with `spawn` on, item ids 62-70 and 105,107,108,110,111,113 are
      unavailable, book level clamps to 5, staff spell level to 10, Rogue/Sorcerer heroes are
      locked with the purchase message, and the V12.9 spell skip list is honoured — each asserted
      from this package's table (single source of truth for other beans).
- [ ] Durability rule test: `ItemRndDur` jitter stays within the C++ bounds; an item at 0
      durability triggers `RemoveEquipment` + `CalcPlrInv`, while `DUR_INDESTRUCTIBLE` and
      durability 0-at-source never break (worked example: weapon after hit vs shield after block).
- [ ] `restrict.cpp` is asserted to contain only `ReadOnlyTest` (no content gate) and the port
      must not implement gating there — recorded as a corrected assumption with a test that the
      gate table, not a `restrict` analogue, drives spawn behaviour.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — one enum per C++ enum, exhaustive switches
- /home/blaze/.agents/skills/coding-rules/SKILL.md — constants file reads as a spec
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/ebiten/input/SKILL.md — cursor/hwcursor vs software blit choice

## Dependencies & blockers
blocked_by dg-task-101. Referenced by dg-task-102/103/104/106/107/110/111/113/114/115 (all use
these constants) and blocks dg-task-118 (audit). Options themselves live in dg-task-041; QoL
implementations in dg-task-043 — this bean defines only their data/registry shape.

## Open questions
- Do cursor *graphics* belong here or to dg-task-206 (graphics asset manifest)? Recommend:
  ids/thresholds here, art files there.
- `item_cursor_id` has 116 distinct `cursorGraphic` values in itemdat — ship the full enum or a
  sparse map keyed by item? Sparse map is smaller; full enum matches C++ exhaustiveness.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:GetGoldCursor-boundary-wrong-in-AC (1000→SMALL not MEDIUM)


## IMPL NOTES (triage 2026-10-16)
1 root: single-source-of-truth constants/enums package (cursors, gold limits, durability,
   spawn gates, QoL registry, panel buttons). Parent devilGo-yjuv, blocked_by devilGo-jx08
   (dg-task-101 TSV loader — item_cursor_id comes from itemdat cursorGraphic) ✓ consistent.
2 local: cursor_defs.hpp:14-19 SelectionRegion (None/Bottom/Middle/Top = 0/1/2/4) ✓ and
   :22-36 all 13 cursor_id values in cited order ✓ (verified verbatim); items.h:28-30
   GOLD_SMALL/MEDIUM/MAX = 1000/2500/5000 ✓; restrict.cpp = ReadOnlyTest only ✓; spawn gates
   items.cpp:2378-2383 ids 62-70 + scrolls 105,107,108,110,111,113 ✓ exact; selhero Rogue/
   Sorcerer lock at DiabloUI/hero/selhero.cpp:292 (bean says :294 — trivial drift) ✓;
   Source/qol/ = 16 files / 8 modules ✓ matches AC phrasing. **AC boundary error:**
   GetGoldCursor (items.cpp:2975-2985): `value >= 2500 → LARGE; value <= 1000 → SMALL; else
   MEDIUM`. Bean AC says "1000 -> GOLD_MEDIUM" — wrong, C++ gives 1000 → GOLD_SMALL.
3 online: n/a.
4 solution: fix worked example: 999→SMALL, **1000→SMALL**, 2499→MEDIUM, 2500→LARGE,
   5000→LARGE, 5001 clamps per MaxGold. Keep everything else — research is unusually
   accurate.
5 understanding: gamedata/consts layer consumed by ~9 domain tasks (per body) — pure data +
   predicates, no behaviour; correct anti-duplication design.
6 repro: `sed -n '2975,2985p' Source/items.cpp`; `sed -n '14,37p' Source/cursor_defs.hpp`;
   `sed -n '2378,2383p' Source/items.cpp`.
7 fix: `beans update devilGo-orkj --body-replace-old "999 -> \`GOLD_SMALL\`, 1000 ->
   \`GOLD_MEDIUM\`" --body-replace-new "999 -> \`GOLD_SMALL\`, 1000 -> \`GOLD_SMALL\` (C++
   \`<= GOLD_SMALL_LIMIT\`)". (Match exact bean text when applying.)
8 validation: boundary test in AC covers 999/1000/2499/2500/4999/5000 — after fix the
   expected values match items.cpp branches.
9 related: devilGo-jx08, devilGo-0tsf (option keys), devilGo-yd48 (QoL impls vs registry),
   devilGo-cciu (spawn/HF gating overlaps are complementary, not duplicate).
VERDICT: NEEDS-WORK:GetGoldCursor-boundary-wrong-in-AC (1000→SMALL not MEDIUM)

GetGoldCursor boundary corrected (triage): 1000 = SMALL bucket, not MEDIUM.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).
