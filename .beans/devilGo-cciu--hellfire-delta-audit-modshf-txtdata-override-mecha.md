---
# devilGo-cciu
title: 'Hellfire delta audit: mods/hf txtdata override mechanism + per-file row deltas (V0.2)'
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T20:38:06Z
parent: devilGo-yjuv
blocked_by:
    - devilGo-jx08
    - devilGo-2j8j
    - devilGo-778h
    - devilGo-vhgd
    - devilGo-rctk
    - devilGo-aix0
    - devilGo-b8z9
    - devilGo-pu3r
---

# Hellfire delta audit and override mechanism

## Goal
Produce the authoritative Hellfire delta: prove the `hf` mod override mechanism works end to end,
and make every base-vs-hf per-file difference (appended, modified, replaced, absent) an explicit,
testable record — including the content that is *still* code-branched and cannot be expressed as
table data.

## Research
- `docs/research/dg-requirements-append.md` **V0.2** (override mechanism) and **V12.10** (authoritative
  delta list), **V12.12** (count table), plus the per-domain delta ACs in devilGo-2j8j/104/106/
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
  18 -> 25 cols (fix2)), classdat 6->6 (2 modified), sorcerer/attributes 27->27 (1 modified),
  sorcerer/starting_loadout 9->9 (3 modified). Tables with **no** hf override: itemdat,
  objdat, questdat, textdat(+unused), Experience, unique_monstdat, and 4 of 6 classes' remaining
  per-class files.
- Still code-branched (135 `gbIsHellfire` sites in root `Source/*.cpp`): vendor caps/values
  (`NumSmithBasicItemsHf`, `MaxVendorValueHf`, `itemLevelAddHf`), shrine pool 34 vs 26
  (`objects.cpp:1302`), Wirt price sign flip (`stores.cpp:1003-1005`), premium slot count,
  `!gbIsHellfire && FlipCoin(4)` staff rule (`items.cpp:1239`), `AddSpectralArrow` damage fix,
  Jester/Devastation/Doppelganger combat branches (`player.cpp:524-660`), 25-vs-17 level count
  (`loadsave.cpp:51`), `bCowQuest`/`bTheoQuest` (devilGo-h6u0), `hfmusic`/`hfvoice`/`hfmonk`/
  `hfbard`/`hfbarb` archives (`assets.cpp:570-586`), `selstart` mode switch, `IsTownerPresent`
  (`towners.cpp:735-750`). Diablo-only item filter in `IsItemAvailable`: `IDI_LGTFORGE`,
  `IDI_OIL..IDI_GREYSUIT`, ids 83-86, 92, 161-165, `IDI_SORCERER`, +`IDI_BARDSWORD`/
  `IDI_BARDDAGGER` under `testBard`.

## Scope
`internal/gamedata/hfdelta/`: a generated **delta manifest** (one row per TSV file: base rows, hf
rows, appended count, modified count/ids, header-diff flag, "no override" sentinel) checked
against the disk files; the Go implementation of the override/priority resolution (extending
devilGo-jx08 `override.go`); and a **code-branch checklist** table mapping each documented
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
- [ ] Header-diff case: `towners/quest_dialog.tsv` is the one file whose hf header differs (25 cols,
      6 extra vs base — fix2 corrected per disk-verified triage); the loader must key columns by header name and the test asserts both shapes
      parse with no silent column shift.
- [ ] Modified-row content audit: the union of modified rows across all files = 9+20+2+2+10+3+2+1+3
      = **52** rows is enumerated by id and each differs from base in the documented columns only;
      `effects-unused.tsv` is asserted as a *replacement* (15 -> 1), not an append.
- [ ] Code-branch checklist: every behaviour in the "still code-branched" list above has a row
      (flag name + owning bean + Go site or `TODO`), and the count of `gbIsHellfire`-equivalent
      flags in Go is reconciled against the 135 C++ sites with the mapping written up.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — generated manifests, codegen discipline
- /home/blaze/.agents/skills/coding-rules/SKILL.md — one manifest, many consumers (no duplication)
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline (audit doc is the deliverable)
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — cross-file diff tests

## Dependencies & blockers
blocked_by devilGo-jx08 and by the domain tables it audits: devilGo-2j8j (items), 104 (monsters),
106 (missiles), 108 (town), 111 (spells), 113 (audio), 114 (classes). Blocks devilGo-i2cr (the CI
audit consumes this manifest) and feeds existing devilGo-mwlm/023/025/026 Hellfire branches and
dg-epic-13 M4.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: user mods (clock, adria_refills_mana, FF mods) = post-1.0 with Lua; this bean ports the hf override mechanism only (loose mods/<name>/ dirs load when present, no Lua).
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


## IMPL NOTES (triage 2026-10-16)
1 root: authoritative base-vs-hf delta manifest + override/priority resolution port + code-branch
   checklist. Parent devilGo-yjuv, blocked_by jx08/2j8j/778h/vhgd/rctk/aix0/b8z9 (all exist =
   devilGo-jx08/102/104/106/108/111/113 ✓ via id-map).
2 local: options.cpp:71-80 DiscoverMods verified verbatim (clock, adria_refills_mana, both
   Floating Numbers mods, +hf) ✓; `grep -rn gbIsHellfire Source/*.cpp | wc -l` → **135
   exactly** ✓; spelldat 36→51 rows ✓ (measured: base 36, mods/hf 51); mods/ contains only
   hf ✓; mods/hf/txtdata/ has classes/items/missiles/monsters/sound/spells/towners → objdat/
   itemdat/questdat/textdat/Experience/unique_monstdat absent = no-override claim ✓;
   restrict.cpp is 34 lines, ReadOnlyTest only ✓ (independent check). Body references
   devilGo-pu3r (classes = devilGo-pu3r) for classdat/attributes/starting_loadout — 3 of the
   14 audited hf files — but pu3r is NOT in blocked_by.
3 online: n/a.
4 solution: add devilGo-pu3r to blocked_by (classes hf files are part of the 14-file AC);
   resolve the mods-scope open question using the recorded Lua deferral (Floating Numbers
   mods are Lua mods → cannot run pre-1.0 → hard-code `hf` only now, generic mechanism ships
   but only hf registered; record as decision in bean).
5 understanding: gamedata layer audit feeding devilGo-i2cr (devilGo-i2cr CI audit) and every
   domain table task; override resolution extends jx08's loader — correct seam.
6 repro: `sed -n '71,80p' Source/options.cpp`; `grep -rn gbIsHellfire Source/*.cpp | wc -l`;
   `tail -n +2 mods/hf/txtdata/spells/spelldat.tsv | grep -c .`.
7 fix: `beans update devilGo-cciu --blocked-by devilGo-pu3r` (append); body-append the
   hf-only-for-now decision entry.
8 validation: blocked_by has 8 entries incl pu3r; body records the mods decision.
9 related: devilGo-jx08 (override.go), devilGo-i2cr (118 audit), devilGo-pu3r, devilGo-vinx.
VERDICT: NEEDS-WORK:add-pu3r-dep + record hf-only decision from Lua deferral

Decision-alignment (triage): pu3r dep added; hf-only scope note aligned with Lua deferral decision.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
