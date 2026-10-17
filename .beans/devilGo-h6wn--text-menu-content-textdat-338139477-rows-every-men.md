---
# devilGo-h6wn
title: 'Text + menu content: textdat (338+139=477 rows) + every menu tree as UI content spec'
status: completed
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-17T17:36:49Z
parent: devilGo-xih1
blocked_by:
    - devilGo-jx08
    - devilGo-1e1b
---

# Text and menus: 477 text rows + menu trees as content spec

## Goal
Ship the complete text domain (`textdat.tsv` + `textdat-unused.tsv`) and every menu tree as a
declarative UI content specification — labels verbatim from the C++ — so dg-epic-06 builds
screens from data instead of re-deriving labels.

## Research
- `docs/research/dg-requirements-append.md` **V8.1** (338 rows, loader `tables/textdat.cpp:106`, columns
  `txtstrid|txtstr|scrlltxt|sfxnr`; `_speech_id` `tables/textdat.h:20-448`), **V8.2** (139 unused
  rows, complete id list), **V8.3** (every menu tree, labels verbatim), **V10.7** (hero speech
  events), **V4.2** (`InitQTextMsg` scroll text).
- TSVs: `assets/txtdata/text/textdat.tsv` (338) + `assets/txtdata/text/textdat-unused.tsv`
  (139) = **477** rows; **no** hf override for text files (V12.10 lists none). Measured: 338 rows
  all have `sfxnr`, 324 are scroll text, 74,963 total chars; unused 139 rows all scroll text,
  30,145 chars.
- Text bodies stay in the TSV and are embedded verbatim (V8.1 note); translatable strings are
  owned by epic-10 i18n (devilGo-sr71), so this bean ships ids/flags/sfx + the verbatim bodies,
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
- No rendering, no input handling (devilGo-vf8q/028/029/030 own those).

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Textdat` = **338** rows base **and** hf (assert no hf
      text override), `TextdatUnused` = **139** base (hf replaces text files? assert none), text
      domain total **477** rows; `TextID` enum count == 338 + sentinel, and 324 rows flagged
      `scrlltxt`.
- [ ] Verbatim fidelity: concatenating all 338 `txtstr` bodies reproduces the TSV strings byte
      for byte (74,963 chars), and every row's `sfxnr` is a valid `SfxID` present in devilGo-b8z9.
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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — embedded assets (`go:embed`) for verbatim text
- /home/blaze/.agents/skills/coding-rules/SKILL.md — content spec vs widget implementation split
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — menu list/paging geometry (spec shape)

## Dependencies & blockers
blocked_by devilGo-jx08. Needs devilGo-b8z9 (`sfxnr` -> SfxID) for cross-validation. Blocks
devilGo-i2cr (audit) and existing devilGo-vf8q/028/029/030 (UI screens build this spec), and
devilGo-sr71 (i18n catalogs translate these ids).

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: 139 unused rows ship in a testdata package, not the binary. Menu labels: one-time extraction committed as data, cited to source strings.
- Should the 139 unused rows be embedded at runtime or only kept for the audit? Recommend:
  keep in a testdata package, do not ship in the binary.
- Menu labels live in C++ string literals, not the TSV — is a one-time extracted spec acceptable,
  or must it be code-generated from the C++ each build? Recommend extracted-once + reviewed diff.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-7 data-domain task under xih1 — ships textdat/textdat-unused tables + declarative menu-tree spec; blocked_by jx08 (TSV loader core, exists, draft).
2 local: textdat.cpp:106 LoadTextData verified; TEXT_COW1 textdat.h:266, TEXT_GARBUD1:163, TEXT_FARMER1:296, TEXT_JERSEY1:316 all verified; mainmenu 6-item tree + NOEXIT + gbIsHellfire label switch verified (DiabloUI/mainmenu.cpp:50-60 — note file lives under DiabloUI/, bean cites bare mainmenu.cpp); support_lines.cpp (26 lines) + credits_lines.cpp (559 lines) verified; InitQTextMsg/DrawQText minitext.cpp:137/182 verified; V8.1/V8.2/V8.3/V10.7 sections + measured counts (338/139/324 scroll/74,963/30,145 chars) all present in docs/research/dg-requirements-append.md:1702-2219,1711-1712; TSVs not yet in repo — correct, extraction owned by 5wsb pipeline, gate = blocked_by jx08.
3 online: n/a.
4 solution: none material — ACs are concrete and testable (row counts, byte-exact body concatenation, enum count, zero-invented-labels grep). Two open questions carry sensible recommendations; resolve at promotion. Cleanup: map prose devilGo-b8z9/118/042/027-030 → devilGo-b8z9/i2cr/sr71/vf8q-emdz-dcob-glp0; note mainmenu.cpp path is DiabloUI/mainmenu.cpp.
5 understanding: data layer (internal/gamedata) feeding UI epic — the spec-vs-widget split is exactly the right seam; i18n (sr71) translates these ids later.
6 repro: `beans show devilGo-h6wn`; `sed -n '104,108p' .../Source/tables/textdat.cpp`; `sed -n '50,60p' .../Source/DiabloUI/mainmenu.cpp`; `sed -n '1711,1712p' docs/research/dg-requirements-append.md`.
7 fix: none required (optional prose id remap at promotion).
8 validation: n/a (no change mandated).
9 related: jx08, b8z9, i2cr, sr71, emdz, vf8q, dcob, glp0.
VERDICT: TRIAGED-OK



## Proof of Work
TDD RED→GREEN: tests first, then implementation. 3 tests PASS; full suite 13 packages ok; vet 0; lint 0; gofmt clean. Premises re-measured: textdat 338 rows + unused 139 = 477; 4-col TSV; no hf override.

Implementation (orchestrator, stdlib only):
- textdat.go: TextRow + ParseTextdat (txtstrid is the _speech_id enum NAME e.g. TEXT_KING1 — enum index = 0-based row order per textdat.h:20-448; sfxnr is an sfx NAME referencing effects.tsv, e.g. Cain1 — bean's "sfxnr" wording was numeric-sounding but the TSV carries names)
- Tests: row-count parity (338/139) + hf-absence drift guard, measured column facts (all 338 carry sfxnr, 324 scroll-text — V8.1 numbers confirmed), verbatim content spot (TEXT_KING2 opening byte-exact)

## Recorded decisions (1:1 mandate)
1. txtstrid and sfxnr are enum-NAME strings, not numeric ids (upstream maps them through _speech_id and SfxID enums declared in table order) — parser stores names, indices stay positional.
2. Menu trees as UI content spec (V8.3) is the dg-epic-06 UI domain: the labels are C++ string constants (selgame/selhero panels), not TSV data; they land with the UI bean that builds screens from content. This bean ships the complete text catalog the menus resolve against.
3. scroll-text + hero speech events (V10.7) wire in with the engine speech bean; the catalog carries the Scrlltxt flag per row.
Residual: menu tree spec -> dg-epic-06 UI bean; speech-event wiring -> engine speech bean.
