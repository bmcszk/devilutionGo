---
# devilGo-h6u0
title: Cow + Hellfire secret-level content (town cows, Nut quest, Hive/Nest, Crypt)
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T20:37:09Z
parent: devilGo-97mn
blocked_by:
    - devilGo-778h
    - devilGo-1e1b
    - devilGo-rctk
    - devilGo-fpkz
    - devilGo-1spo
---

# Cow and Hellfire secret-level content

## Goal
Implement the cow/towner novelty content and the two Hellfire secret dungeon families (Hive/Nest
and Crypt/Grave) exactly as documented — including the finding that this fork has **no** secret
cow level.

## Research
- `docs/research/dg-requirements-append.md` **V5.4** (cows, complete documentation), **V11.2** (trigger
  tiles), **V11.4** (set levels), **V11.5** (unique features), **V12.10** (still-code-branched
  `bCowQuest`/`bTheoQuest`), **V2.4** (Na-Krul/Hork death hooks).
- Town cows: objdat/towners V5.1 rows 10-12 `TOWN_COW` at (58,16) SW, (56,14) NW, (59,20) N;
  `animWidth` 128, 12 frames, delay 3, **no** `animPath` (shared `CowSprites`,
  `towners.cpp:26`); `InitCows()` `:141-175` loads one frame per cow from
  `(*CowSprites)[entry.direction]` (`:154`) and occupies **4 tiles** by writing `-cowId` into the
  three extra tiles (`:158-171`, comment explains the clipping bug it avoids).
- `TalkToCow` `towners.cpp:494-520`: ignored while a cow sound plays (`:495-497`); `CowClicks++`;
  sound `Cow1`; **4th click** => `Cow2` (spawn resets `CowClicks` to 0 so the 1->2 loop never
  resolves, `:501-505`); **8th+ click** non-spawn => reset to 4 and cycle hero speech
  `YepThatsACowAlright` -> `ImNotThirsty` -> `ImNoMilkmaid` (`CowMsg` wraps at 3, `:506-515`);
  `TownerTalk` resets `CowClicks = CowMsg = 0` (`:205-209`). Voices need `hfvoice`
  (`assets.cpp:570-586`); `Cow1`/`Cow2` = `sound_effect_enums.h:310-311`,
  `YepThatsACowAlright` = `:61`; `TEXT_COW1`/`TEXT_COW2` = `textdat.h:266-267`.
- **No secret cow level in this fork** (V5.4a: grep-verified); the only "secret level" content is
  Hellfire Hive/Nest + Crypt. Record this as a negative finding, do not invent content.
- Cow quest (`bCowQuest`, option `cowQuest` at `options.cpp:848`, flags
  `CantChangeInGame | OnlyHellfire`, default false, wire field `sgGameInitInfo.bCowQuest`
  `game_mode.hpp:27`, set `multi.cpp:612`, demo-serialised `demomode.cpp:126,164,193,228,647`,
  chat gate `control_chat_commands.cpp:159-161`): NPC swap `IsTownerPresent`
  (`towners.cpp:735-750` — `TOWN_FARMER` needs `gbIsHellfire && !bCowQuest && Q_FARMER !=
  QUEST_HIVE_DONE`, `TOWN_COWFARM` needs `gbIsHellfire && bCowQuest`, `TOWN_GIRL` needs
  `bTheoQuest && _pLvlVisited[17] && Q_GIRL != QUEST_DONE`; cow farmer short name `nut` `:722`).
- Lester path `TalkToFarmer` `:523-576` (runebomb -> `TEXT_FARMER2`/ACTIVE/`_qvar1=1`; taunts by
  visited levels; else `SpawnRuneBomb(farmer.pos + (1,0))`; DONE -> `TEXT_FARMER4` +
  `SpawnRewardItem(IDI_AURIC)` -> `QUEST_HIVE_DONE`). Nut path `TalkToCowFarmer` `:586-650`
  (Grey Suit -> `TEXT_JERSEY7`; Brown Suit -> `SpawnUnique(UITEM_BOVINE, pos+SouthEast, qlevel)`
  + `TEXT_JERSEY8` + DONE + `UpdateCowFarmerAnimAfterQuestComplete()`; HIVE_TEASE1/2/ACTIVE chain;
  taunt pool `PickRandomlyAmong({TEXT_JERSEY9..12})`). Cel path swap
  `towners\farmer\cfrmrn2` -> `mfrmrn2` on DONE (`:184-188`); `Q_JERSEY` town hook
  `quests.cpp:618`.
- Suit spawns: `bCowQuest` puts `IDI_BROWNSUIT` at tile **(25,25)** on **level 20** and
  `IDI_GREYSUIT` at **(25,25)** on **level 19** (`items.cpp:2455-2460`, in the `!setlevel` block
  next to `SpawnRock()`/`Q_ANVIL`). Reward Bovine Plate = `IDI_LGTFORGE` via `UITEM_BOVINE`.
- Hive/Nest + Crypt gating: `OpensHive` x in [79,82] y in [61,64] (`levels/town.cpp:249-254`)
  checked with `holdItem.IDidx == IDI_RUNEBOMB` (`diablo.cpp:323`) -> `OpenHive()` (`:265-272`:
  `CMD_OPENHIVE` + `Q_FARMER._qactive = QUEST_DONE`); `TownOpenHive()` rewrites
  `dPiece[78..80][60..65]` (`:283-300`); `IsWarpOpen(DTYPE_NEST)` at load (`:228-233`); town
  trigger **(80,62) -> level 17** (`trigs.cpp:136-141`); `AddOpenNest` spawns the explosion at
  those four tiles (`missiles.cpp:1238-1246`). Grave/Crypt: `OpensGrave` x in [35,38] y in
  [20,24] (`:256-261`), `TownCryptList` (`trigs.cpp:34`), trigger **(36,24) -> level 21**.
  Hive triggers `==65` PREV / `==62` NEXT / `==79` TWARPUP; Crypt `==183` TWARPUP /
  `==157` PREV / `==125` NEXT (`trigs.cpp:265-315`). Nest light radius 9 (`missiles.cpp:1338`),
  `DTYPE_NEST ? 9 : 3`.
- Payoff: with `bCowQuest`, Na-Krul's greeting SFX becomes `NaKrul6` instead of `NaKrul4`/`5`
  (`monster.cpp:931-936`, `:4282-4290`). Hork Demon death spawns Theodore when `bTheoQuest` else
  `CreateAmulet` (`monster.cpp:925-929`, `IDI_THEODORE`).
- Non-table HF content: `mods/hf/nlevels/`, `mods/hf/ui_art/`, `mods/hf/data/`, `mods/hf/lua/`,
  `manifest.ini` (V12.10).

## Scope
`internal/gamedata/secrets/` (or `internal/gameplay/cowquest/`): cow towner rows + 4-tile
occupancy convention (`-cowId`), click state machine (`CowClicks`, `CowMsg` wrap at 3, spawn
reset), `bCowQuest`/`bTheoQuest` flags threaded from SimOptions, farmer/Nut dialog tables,
suit/rune-bomb spawns at the exact tiles/levels, Hive/Nest + Grave/Crypt tile ranges and trigger
tiles, Na-Krul SFX substitution, Hork/Theodore branch. Content/spec work; level *generation*
stays in dg-task-115.

## Acceptance criteria
- [ ] TDD first: content-count parity — 3 town cow rows (positions (58,16)/(56,14)/(59,20),
      `animWidth` 128, 12 frames, delay 3, no `animPath`) and the 26 appended hf monster rows +
      40 hf missiles referenced by the Nest/Crypt flows are present; plus a negative test
      asserting **zero** secret-cow-level content exists in this fork (no such leveltype/tile).
- [ ] Cow click worked example: clicks 1,2,3 => `Cow1`; click 4 => `Cow2`; clicks 5-7 => `Cow1`
      loop; click 8 (non-spawn) => counter resets to 4 and speech cycles
      YepThatsACowAlright -> ImNotThirsty -> ImNoMilkmaid -> wraps; any NPC talk resets both
      counters. Asserted as a 12-click tape, and again with `spawn` on (never reaches speech).
- [ ] Suit/rune-bomb placement: with `bCowQuest` on, `IDI_BROWNSUIT` spawns at (25,25) on level
      20 and `IDI_GREYSUIT` at (25,25) on level 19, and no other level places them; `Q_FARMER`
      normal path spawns a rune bomb at `farmer.pos + (1,0)`.
- [ ] Hive/Crypt geometry: `OpensHive` accepts exactly x 79-82, y 61-64 and `OpensGrave` exactly
      x 35-38, y 20-24 (boundary +1/-1 cases); town triggers (80,62)->17 and (36,24)->21 fire;
      `TownOpenHive` rewrites `dPiece[78..80][60..65]`.
- [ ] Quest-state parity: Lester path ends `QUEST_HIVE_DONE` with `SpawnRewardItem(IDI_AURIC)`;
      Nut path completing with the Brown Suit spawns `UITEM_BOVINE` at `pos+SouthEast` and swaps
      the farmer cel to `mfrmrn2`; `Q_JERSEY` DONE while on level 0 refreshes town state.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — feature-flag hygiene for bCowQuest/bTheoQuest
- /home/blaze/.agents/skills/coding-rules/SKILL.md — dialog table (data) vs click machine
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline (negative findings count)
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — shared cow sprite sheet, 4-tile occupy

### game-dev skills (2026-10-16 audit)
- /home/blaze/.agents/skills/game-developer/SKILL.md
- /home/blaze/.agents/skills/game-development/game-design/SKILL.md — quest design

## Dependencies & blockers
blocked_by dg-task-104 (Na-Krul/Hork rows), dg-task-107 (objdat lever/barricade rows),
dg-task-108 (towners rows + `IsTownerPresent` data). Also needs dg-task-019 (town + setmaps) and
dg-task-026 (quest FSM) for integration, and dg-task-006 for the `mods/hf/nlevels/` data.

## Open questions
- Does the Go port load Nest/Crypt geometry from `mods/hf/nlevels/` (data) or from `.dun`
  fixtures in `test/fixtures/hellfire/`? Both exist; pick one source of truth.
- `IDI_LGTFORGE` (Bovine Plate) is gated Diablo-unavailable (V12.10 item gate list) while the cow
  quest is Hellfire-only — confirm the reward is unreachable by design in Diablo, not a bug.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-7 content task (dg-task-109): town cows, click machine, bCowQuest/bTheoQuest, Lester/Nut paths, Hive/Nest + Grave/Crypt geometry, Na-Krul/Hork branches; draft; wire blockedBy [778h, 1e1b, rctk] = exactly body's dg-task-104/107/108 ✓.
2 local: LIVE-VERIFIED: townerdat.tsv rows 11-13 TOWN_COW at (58,16)SW (56,14)NW (59,20)N, animWidth 128, 12 frames, delay 3, animPath blank ✓; CowClicks towners.cpp:28/:207/:498/:501 ✓; Cow1/Cow2 sound_effect_enums.h:310-311 ✓, YepThatsACowAlright :61 ✓; OpensHive town.cpp:249-254 (79-82/61-64) ✓ OpensGrave :256-261 (35-38/20-24) ✓; town triggers (25,29)(49,21)(17,69)(41,80)(80,62)(36,24) trigs.cpp:114-143 ✓; suits items.cpp:2457-2460 IDI_BROWNSUIT(25,25)@L20 / IDI_GREYSUIT(25,25)@L19 ✓ exact; hf deltas: +26 monster rows (112→138 ✓) and +40 missile rows (68→108 ✓) both counted on disk; portal.cpp:59 radius 15 ✓; mods/hf/nlevels/ exists ✓. Soft integration deps (yeia town, xw8k quest FSM, vof6 assets) named in body but not wired — hard blockers are, and fpkz declares the shared Hive/Crypt trigger-tile seam (fpkz→h6u0, unwired, counted under fpkz).
3 online: n/a (all claims local).
4 solution: none — ready once unblocked; optional: wire fpkz as blocker when fixing fpkz (one direction suffices).
5 understanding: content/behavior layer over epic-04 world (town/setmaps) + epic-05 quests; negative finding "no secret cow level" correctly framed as data, not invention — matches fidelity mandate.
6 repro: `grep -P "^TOWN_COW" devilutionX/assets/txtdata/towners/towners.tsv; sed -n '249,261p' devilutionX/Source/levels/town.cpp; sed -n '2457,2460p' devilutionX/Source/items.cpp; awk 'END{print NR}' devilutionX/mods/hf/txtdata/missiles/misdat.tsv`
7 fix: none (fpkz→h6u0 edge covered in fpkz's fix).
8 validation: n/a.
9 related: 778h/1e1b/rctk (deps), fpkz (tile/trigger seam), yeia/xw8k (integration).
VERDICT: TRIAGED-OK


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
