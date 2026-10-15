---
# devilGo-h6wn
title: 'Text + menu content: textdat (338+139=477 rows) + every menu tree as UI content spec'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-15T18:29:02Z
parent: devilGo-xih1
blocked_by:
    - devilGo-jx08
---

# Text and menus: 477 text rows + menu trees as content spec

## Goal
Ship the complete text domain (`textdat.tsv` + `textdat-unused.tsv`) and every menu tree as a
declarative UI content specification — labels verbatim from the C++ — so dg-epic-06 builds
screens from data instead of re-deriving labels.

## Research
- `/tmp/dg-requirements-append.md` **V8.1** (338 rows, loader `tables/textdat.cpp:106`, columns
  `txtstrid|txtstr|scrlltxt|sfxnr`; `_speech_id` `tables/textdat.h:20-448`), **V8.2** (139 unused
  rows, complete id list), **V8.3** (every menu tree, labels verbatim), **V10.7** (hero speech
  events), **V4.2** (`InitQTextMsg` scroll text).
- TSVs: `assets/txtdata/text/textdat.tsv` (338) + `assets/txtdata/text/textdat-unused.tsv`
  (139) = **477** rows; **no** hf override for text files (V12.10 lists none). Measured: 338 rows
  all have `sfxnr`, 324 are scroll text, 74,963 total chars; unused 139 rows all scroll text,
  30,145 chars.
- Text bodies stay in the TSV and are embedded verbatim (V8.1 note); translatable strings are
  owned by epic-10 i18n (dg-task-042), so this bean ships ids/flags/sfx + the verbatim bodies,
  not translations.
- Consumers: `InitQTextMsg`/`DrawQText` (`Source/minitext.cpp`), `interfac.cpp` `scrlltxt` rows,
  quest/towner dialog (V5.2), `GetSpeechText`, `TEXT_COW1`/`TEXT_COW2` (`textdat.h:266-267`),
  `TEXT_GARBUD*`, `TEXT_ZHAR*`, `TEXT_VILE*`, `TEXT_FARMER1..8`, `TEXT_JERSEY1..12`.
- Menu trees (V8.3): `mainmenu.cpp:53-58` = 6 items [Single Player | Multi Player | Settings |
  Support | Show Credits | Exit Diablo/Exit Hellfire by `gbIsHellfire`, compiled out under
  `NOEXIT`]; support lines `support_lines.cpp:8-24` (8 lines); credits
  `credits_lines.cpp:8+` (559-line file); settings tree `settingsmenu.cpp`; hero flow
  `hero/selhero.cpp` -> New Hero -> Choose Class -> Enter Name; MP flow
  `multi/selconn.cpp` -> `multi/selgame.cpp` -> password -> `selstart.cpp`
  [Enter Hellfire | Switch to Diablo]; shared modals `selok.cpp` (:86), `selyesno.cpp`,
  `dialogs.cpp`, `progress.cpp`, `button.cpp`, `scrollbar.cpp`, `text_input.cpp`.
- In-game menus: `gmenu.cpp:185,251` + `gmenu.h:101,106` sliders, `gamemenu.cpp:274-277`
  tick slider 20..50, `help.cpp:143,167,191`, `diablo_msg.cpp:45,125-130` (3500 ms toasts);
  store screens use the 20-line `AddSText`/`AddSLine` model (V5.3).

## Scope
`internal/gamedata/textdat/` + `internal/gamedata/menudata/`:
- `SpeechText` table (338) with `TextID` enum in `_speech_id` order, `ScrollText bool`,
  `Sfx SfxID`; `UnusedText` table (139) kept as a *documented* dead set (V8.2), not deleted.
- Menu content spec: declarative tree type (id, label string id, command id, submenu), populated
  for title/main menu/settings/support/credits/hero-select/MP-selection/selstart/yes-no/ok/
  progress/help/gamemenu, plus `panel_button_id` (8) and the store 20-line paging shape.
- No rendering, no input handling (dg-task-027/028/029/030 own those).

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Textdat` = **338** rows base **and** hf (assert no hf
      text override), `TextdatUnused` = **139** base (hf replaces text files? assert none), text
      domain total **477** rows; `TextID` enum count == 338 + sentinel, and 324 rows flagged
      `scrlltxt`.
- [ ] Verbatim fidelity: concatenating all 338 `txtstr` bodies reproduces the TSV strings byte
      for byte (74,963 chars), and every row's `sfxnr` is a valid `SfxID` present in dg-task-113.
- [ ] Menu content worked example: the main-menu tree is exactly the 6 items above in that order
      with the D/HF label switch on item 6, and the MP sub-flow reaches
      `selconn -> selgame -> password -> selstart` with `[Enter Hellfire | Switch to Diablo]`
      only when Hellfire is available.
- [ ] Every menu label in the spec resolves to a `TextID` or a literal that appears verbatim in
      the cited C++ (test greps the spec strings against `Source/DiabloUI/*` and `gmenu.cpp`) —
      zero invented labels.
- [ ] Scroll-text sequence: `InitQTextMsg(TEXT_*)` queue drains in order and each entry plays its
      own `sfxnr`; scripted test with 3 chained quest texts.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — embedded assets (`go:embed`) for verbatim text
- /home/blaze/.agents/skills/coding-rules/SKILL.md — content spec vs widget implementation split
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — menu list/paging geometry (spec shape)

## Dependencies & blockers
blocked_by dg-task-101. Needs dg-task-113 (`sfxnr` -> SfxID) for cross-validation. Blocks
dg-task-118 (audit) and existing dg-task-027/028/029/030 (UI screens build this spec), and
dg-task-042 (i18n catalogs translate these ids).

## Open questions
- Should the 139 unused rows be embedded at runtime or only kept for the audit? Recommend:
  keep in a testdata package, do not ship in the binary.
- Menu labels live in C++ string literals, not the TSV — is a one-time extracted spec acceptable,
  or must it be code-generated from the C++ each build? Recommend extracted-once + reviewed diff.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
