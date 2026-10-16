---
# devilGo-cciu
title: 'Hellfire delta audit: mods/hf txtdata override mechanism + per-file row deltas (V0.2)'
status: draft
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:01:04Z
parent: devilGo-yjuv
blocked_by:
    - devilGo-jx08
    - devilGo-2j8j
    - devilGo-778h
    - devilGo-vhgd
    - devilGo-rctk
    - devilGo-aix0
    - devilGo-b8z9
---

# Hellfire delta audit and override mechanism

## Goal
Produce the authoritative Hellfire delta: prove the `hf` mod override mechanism works end to end,
and make every base-vs-hf per-file difference (appended, modified, replaced, absent) an explicit,
testable record — including the content that is *still* code-branched and cannot be expressed as
table data.

## Research
- `/tmp/dg-requirements-append.md` **V0.2** (override mechanism) and **V12.10** (authoritative
  delta list), **V12.12** (count table), plus the per-domain delta ACs in dg-task-102/104/106/
  108/111/113/114.
- Mechanism: `Source/options.cpp:71-80` `DiscoverMods()` — `modNames = {clock,
  adria_refills_mana, Floating Numbers - Damage, Floating Numbers - XP}` plus `"hf"` when
  `HaveHellfire()`; `Source/engine/assets.cpp:704-760` `LoadModArchives()` — loose overrides at
  `PrefPath/mods/<name>/` (`:729-733`) then `BasePath/mods/<name>/` (`:736-740`) into
  `OverridePaths`, packed `mods/<name>.mpq` at priority >= 10000 (`:746-758`) beating retail
  (1000) and devilutionx.mpq (9000). `mods/hf/` contains `data/`, `lua/`, `manifest.ini`,
  `nlevels/`, `txtdata/`, `ui_art/`; only `mods/hf/` exists in the repo (`ls mods/`).
- Verified deltas (all 14 hf files re-measured with `tail -n +2 | grep -c`): monstdat
  112->138 (+26, 0 modified), misdat 68->108 (+40, 0), missile_sprites 46->59 (+13, 0),
  item_prefixes 83->86 (+3, 9 modified), item_suffixes 95->98 (+3, 20 modified), unique_itemdat
  90->110 (+20, 2 modified), spelldat 36->51 (+15, 2 modified; +Null => 37/52), effects
  814->917 (+103, 10 modified), effects-unused 15->1 (replaced, -14), towners 12->15 (+3, 3
  modified), quest_dialog 10->13 (+3 rows **and a different header with 3 extra quest columns**:
  18 -> 21 cols), classdat 6->6 (2 modified), sorcerer/attributes 27->27 (1 modified),
  sorcerer/starting_loadout 9->9 (3 modified). Tables with **no** hf override: itemdat,
  objdat, questdat, textdat(+unused), Experience, unique_monstdat, and 4 of 6 classes' remaining
  per-class files.
- Still code-branched (135 `gbIsHellfire` sites in root `Source/*.cpp`): vendor caps/values
  (`NumSmithBasicItemsHf`, `MaxVendorValueHf`, `itemLevelAddHf`), shrine pool 34 vs 26
  (`objects.cpp:1302`), Wirt price sign flip (`stores.cpp:1003-1005`), premium slot count,
  `!gbIsHellfire && FlipCoin(4)` staff rule (`items.cpp:1239`), `AddSpectralArrow` damage fix,
  Jester/Devastation/Doppelganger combat branches (`player.cpp:524-660`), 25-vs-17 level count
  (`loadsave.cpp:51`), `bCowQuest`/`bTheoQuest` (dg-task-109), `hfmusic`/`hfvoice`/`hfmonk`/
  `hfbard`/`hfbarb` archives (`assets.cpp:570-586`), `selstart` mode switch, `IsTownerPresent`
  (`towners.cpp:735-750`). Diablo-only item filter in `IsItemAvailable`: `IDI_LGTFORGE`,
  `IDI_OIL..IDI_GREYSUIT`, ids 83-86, 92, 161-165, `IDI_SORCERER`, +`IDI_BARDSWORD`/
  `IDI_BARDDAGGER` under `testBard`.

## Scope
`internal/gamedata/hfdelta/`: a generated **delta manifest** (one row per TSV file: base rows, hf
rows, appended count, modified count/ids, header-diff flag, "no override" sentinel) checked
against the disk files; the Go implementation of the override/priority resolution (extending
dg-task-101 `override.go`); and a **code-branch checklist** table mapping each documented
`gbIsHellfire` behaviour to its Go feature flag + owning bean. Deliverables: `hfdelta.tsv`-style
manifest + parity tests + a written audit doc in the bean PoW.

## Acceptance criteria
- [ ] TDD first: delta-manifest parity — for all **14** hf files and all **49** base files, the
      generated manifest equals the V12.10 table exactly (per-file base rows, hf rows, appended,
      modified), and the 6 no-override tables (itemdat, objdat, questdat, textdat, Experience,
      unique_monstdat) are asserted as *absent* under `mods/hf/txtdata/`.
- [ ] Override mechanism worked example: with `hf` active, resolving `spells/spelldat.tsv`
      returns the hf file (51 rows) while `objects/objdat.tsv` still resolves to base (109 rows)
      — i.e. override is per-file, never global; and the priority order (loose PrefPath >
      loose BasePath > hf mpq >= 10000 > devilutionx.mpq 9000 > retail 1000) is asserted.
- [ ] Header-diff case: `towners/quest_dialog.tsv` is the one file whose hf header differs (18 vs
      21 columns); the loader must key columns by header name and the test asserts both shapes
      parse with no silent column shift.
- [ ] Modified-row content audit: the union of modified rows across all files = 9+20+2+2+10+3+2+1+3
      = **52** rows is enumerated by id and each differs from base in the documented columns only;
      `effects-unused.tsv` is asserted as a *replacement* (15 -> 1), not an append.
- [ ] Code-branch checklist: every behaviour in the "still code-branched" list above has a row
      (flag name + owning bean + Go site or `TODO`), and the count of `gbIsHellfire`-equivalent
      flags in Go is reconciled against the 135 C++ sites with the mapping written up.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — generated manifests, codegen discipline
- /home/blaze/.agents/skills/coding-rules/SKILL.md — one manifest, many consumers (no duplication)
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline (audit doc is the deliverable)
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — cross-file diff tests

## Dependencies & blockers
blocked_by dg-task-101 and by the domain tables it audits: dg-task-102 (items), 104 (monsters),
106 (missiles), 108 (town), 111 (spells), 113 (audio), 114 (classes). Blocks dg-task-118 (the CI
audit consumes this manifest) and feeds existing dg-task-022/023/025/026 Hellfire branches and
dg-epic-13 M4.

## Open questions
- Should the Go port support user mods (`clock`, `adria_refills_mana`, the two Floating Numbers
  mods are in `DiscoverMods`) with the same mechanism, or hard-code `hf` only for now? The
  mechanism is generic; scope is a decision.
- `mods/hf/{data,lua,nlevels,ui_art}` + `manifest.ini` are non-table HF content: does the audit
  cover them (file-level manifest) or only txtdata? Recommend a file manifest without parsing.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:add-pu3r-dep + record hf-only decision from Lua deferral
