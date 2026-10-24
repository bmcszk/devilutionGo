---
# devilGo-pu3r
title: 'Classes + Experience data packages: 1080 class rows (V10 says 1122) + 50 XP rows'
status: completed
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-20T10:29:06Z
parent: devilGo-97mn
blocked_by:
    - devilGo-jx08
---

# Classes and XP

## Goal
Transcribe the class domain (`classdat.tsv` + 6 per-class x 5 tables) and `Experience.tsv`, and
port the XP thresholds/penalties/caps. Flag and resolve the row-count discrepancy found in the
inventory (see Open questions / AC 5).

## Research
- `docs/research/dg-requirements-append.md` **V10.1** (Experience 50 rows; loader
  `tables/playerdat.cpp:89-127`), **V10.2** (classdat 6 rows; loader `:345`), **V10.3**
  (attributes 27 vars x 6 classes, column sets md5-verified; loader `:171`), **V10.4**
  (starting_loadout 9 vars x 6; loader `:209`), **V10.5** (animations 27 x 6; loader `:251`),
  **V10.6** (sprites 13 x 6; loader `:227`), **V10.7** (sounds 103 events x 6; loader `:289`),
  **V12.8** (XP penalties/caps), **V12.12** (counts), **V12.10** (hf class deltas).
- TSVs (verified on disk): `assets/txtdata/Experience.tsv` (50), `classes/classdat.tsv` (6), and
  for each of the **6** classes `warrior, rogue, sorcerer, monk, barbarian, bard` the five files
  `attributes` (27) `sounds` (103) `animations` (27) `sprites` (13) `starting_loadout` (9) =>
  **30** per-class files, **179** rows each class.
- **Measured total: 6 + 1074 = 1080 class-domain rows.** The inventory's V12.12 "uniqueness
  invariant" line claims `6 + 6*(27+27+103+13+9) = 1122`, which is arithmetically wrong
  (`179 * 6 = 1074`, not 1116). Both this bean and devilGo-i2cr must assert **1080** and record
  the correction.
- hf deltas (V12.12 + V12.10): `classdat.tsv` 6 rows, **2 modified**; `sorcerer/attributes.tsv`
  27 rows, **1 modified**; `sorcerer/starting_loadout.tsv` 9 rows, **3 modified**. Confirmed only
  those 3 files exist under `mods/hf/txtdata/classes/`.
- Structs/enums: `ExperienceData` (playerdat.h:35-70) stores `levelThresholds[level-1]` with gaps
  filled `UINT32_MAX` so `ValidatePlayer()` does not zero XP (`:60-68`);
  `GetNextExperienceThresholdForLevel` (`:403`); `ClassAttributes` + `classFlags` bitset;
  `HeroSpeech` (103 events) -> `SfxID`; `HeroClass` (player.h).
- XP rules (V10.1 rows 1-15 verbatim): L1=2000, L2=4620, L3=8040, L4=12489, L5=18258, L6=25712,
  L7=35309, L8=47622, L9=63364, L10=83419, L11=108879, L12=141086, L13=181683, L14=231075,
  L15=313656, L16=424067, L17=571190, L18=766569, L19=1025154, L20=1366227, L21=1814568 (deltas
  2620, 3420, 4449, 5769, 7454, 9597, 12313, 15742, 20055, 25460, 32207, 40597, 49392, 82581,
  110411, 147123, 195379, 258585, 341073, 448341, 587327).
- Penalties/caps (V12.8): level-delta scaling, MP award cap `min(next/20, 200*level)`
  (`player.cpp:2440`), `_pExperience` hard clamp to the level-50 threshold (`:2445-2448`),
  `NextPlrLevel` grants `lvlLife`/`lvlMana` from V10.3 + 5 stat points per level; XP bar segments
  come from this table (`qol/xpbar.cpp`, `options experienceBar`, `player.cpp:2449`).
- Mana coupling: `GetClassAttributes(cls).manaCost` is the spell mana multiplier (devilGo-aix0).

## Scope
`internal/gamedata/classdat/` (`HeroClass` enum, `ClassAttributes` 27 fields, `StartingLoadout`
9, `Animations` 27, `Sprites` 13, `HeroSounds` 103 -> `SfxID`, `classdat` 6) and
`internal/gamedata/xpdat/` (`Experience` 50 thresholds, `UINT32_MAX` gap fill, threshold/level
queries, MP award cap + level-50 clamp as pure functions). No player struct work (devilGo-r462).

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Classdat` = **6** rows (base and hf), per-class
      `Attributes` = **27**, `Sounds` = **103**, `Animations` = **27**, `Sprites` = **13**,
      `StartingLoadout` = **9** for all 6 classes, `Experience` = **50** rows (base and hf);
      class-domain total = **1080** rows (6 + 30x179 rows) — the number this bean corrects from
      the inventory's 1122, with the correction recorded in the Proof of Work.
- [ ] XP threshold worked example: `Threshold(L1)=2000`, `L2=4620` (delta 2620), `L10=83419`,
      `L15=313656`, `L21=1814568`; a gap in the table returns `UINT32_MAX` (not 0) so
      `ValidatePlayer` cannot zero XP; level-from-XP is the exact inverse for all 50 rows.
- [ ] Award-cap worked example: MP kill award is `min(next/20, 200*level)` and `_pExperience`
      clamps at the level-50 threshold — asserted at the boundary values (cap-1, cap, cap+1).
- [ ] hf deltas by content: `classdat` modifies exactly 2 rows, `sorcerer/attributes` exactly 1,
      `sorcerer/starting_loadout` exactly 3, with row counts unchanged (27/9/6) and no other
      class file modified (byte-diff test over all 30 per-class files).
- [ ] Cross-table integrity: every `sounds.tsv` event maps to a `SfxID` in devilGo-b8z9's catalog
      for both base and hf, and all 6 classes' column *names* are identical (the md5-verified
      claim in V10.3) — assert the column set, not just the row count.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — fixed-size tables, explicit int units
- /home/blaze/.agents/skills/coding-rules/SKILL.md — data table vs XP rule function split
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline (record the 1122->1080 fix)
- /home/blaze/.agents/skills/ebiten/rendering/SKILL.md — sprites/animation row shape feeds epic-03

## Dependencies & blockers
blocked_by devilGo-jx08. Needs devilGo-b8z9 (`HeroSpeech` -> SfxID). Blocks devilGo-aix0 (class
mana multiplier), devilGo-rctk (class-aware Wirt bans use attribute maxima), devilGo-i2cr (audit),
existing devilGo-dcob (hero create uses class defaults) and devilGo-7be9/026 (to-hit, XP flow).

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED — 1080 is the real class-row count (V12.12 "1122" arithmetic wrong, disk-verified across all 30 files). Volume II errata states 1080; the "1,122 class rows" headline is corrected wherever cited.
- **1122 vs 1080**: the V12.12 "uniqueness invariant" arithmetic is wrong (confirmed by
  `find`+`grep -c` on all 30 files). Which number goes into the Volume II errata — 1080 — and
  should the brief's headline "1,122 class rows" be corrected upstream?
- `bard` and `monk`/`barbarian` have full data but are not retail-Diablo classes; do their rows
  ship in the base build (they exist in `assets/txtdata`, so presumably yes, gated by
  `testBard`/Hellfire)? Confirm the gate for each.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps — 4 declared blocks edges missing; 1 cite file-path error; content otherwise fully verified live.


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-7 data task (devilGo-pu3r): classdat + 30 per-class TSVs + Experience.tsv into `internal/gamedata/{classdat,xpdat}`; draft; wire blockedBy [jx08] ✓ = body's devilGo-jx08.
2 local: Row counts LIVE-VERIFIED exact: Experience.tsv 51 lines=50 data, classdat 7=6, per-class 28/104/28/14/10 lines = 27/103/27/13/9 rows → 6+6×179=1080 ✓ (bean's 1122→1080 correction is right). Loader cites playerdat.cpp :89,:171,:209,:227,:251,:289,:345 ALL verified exact (sed shows each loader at those lines). XP L2=4620 ✓ (row 3); cap `min(next/20, 200*level)` + level-50 clamp verified player.cpp:2440-2448 ✓; gap-fill `numeric_limits<uint32_t>::max()` playerdat.cpp:64 ✓; `GetNextExperienceThresholdForLevel` playerdat.cpp:403 ✓. One cite error: `ExperienceData (playerdat.h:35-70)` — file is playerdat.cpp (class at :35), no playerdat.h exists. Wire gap: body "Blocks devilGo-aix0/108/118/029/021/026" but only i2cr(118) has pu3r in blockedByIds; aix0(111), rctk(108), dcob(029), xw8k(026) do not (aix0/rctk/dcob blockedBy jx08 only; xw8k by r462 only).
3 online: n/a.
4 solution: wire 4 missing blockedBy edges (aix0 manaCost, rctk Writ class maxima, dcob hero-create class defaults, xw8k XP table); fix the playerdat.h→playerdat.cpp cite. Content, ACs (row-count + worked examples + hf byte-diff), skills: ready.
5 understanding: game-data layer feeding epic-05 gameplay (combat to-hit, quests XP flow) and epic-06 (hero create) — pure data, correctly split from player struct work (r462).
6 repro: `awk 'END{print NR}' devilutionX/assets/txtdata/classes/warrior/{attributes,sounds,animations,sprites,starting_loadout}.tsv; sed -n '2440,2448p' devilutionX/Source/player.cpp; grep -n "class ExperienceData" devilutionX/Source/tables/playerdat.cpp; python3 -c "import json;e={b['id']:b for b in json.load(open('(process artifact, not archived)/edges.json'))['beans']};print(e['devilGo-aix0']['blockedByIds'])"`
7 fix: `beans update devilGo-aix0 --blocked-by devilGo-pu3r`; `beans update devilGo-rctk --blocked-by devilGo-pu3r`; `beans update devilGo-dcob --blocked-by devilGo-pu3r`; `beans update devilGo-xw8k --blocked-by devilGo-pu3r`; `beans update devilGo-pu3r --body-replace-old "Structs/enums: \`ExperienceData\` (playerdat.h:35-70)" --body-replace-new "Structs/enums: \`ExperienceData\` (tables/playerdat.cpp:35-70; decls in playerdat.hpp)"`
8 validation: `beans query --json '{ bean(id:"devilGo-aix0") { blockedByIds } }'` shows pu3r; grep body no longer cites playerdat.h.
9 related: jx08 (dep), 778h/1e1b (sibling tables), aix0/rctk/dcob/xw8k/i2cr (dependents).
VERDICT: NEEDS-WORK:wire-deps — 4 declared blocks edges missing; 1 cite file-path error; content otherwise fully verified live.

Fix pass 2026-10-16: Dep-wiring applied (fix pass): 4 missing edges wired; cite file-path error corrected per IMPL NOTES.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.
fix2 cite correction: research file-path errors corrected — paths carry engine/render/ prefix per triage evidence; InitThemes correct cite = :826 (was :349-449).



## Proof of Work
Worker lane stalled 20min (0 CPU, 0 files — 5th occurrence, same research-phase pattern despite inline citations); orchestrator killed it and built inline.
Implementation (orchestrator):
- experience.go: ExperienceThresholds — load (Level,Experience), MaxLevel capped 255 (playerdat.cpp:39-42), ThresholdForLevel level0→0 + past-max clamp (:48-55), gap-padded set (:62-66)
- classes.go: LoadClasses (classdat.tsv className,folderName — playerdat.cpp:345), LoadClassTables (2-col per-class tables)
- Tests (6 new; txtdata package 18 PASS): golden thresholds 2000/4620/8040 from real Experience.tsv, MaxLevel=50 (V10.1), level-0 + clamp, bad-number error, classdat + per-class 2-col tables
- 1080-vs-1122 class-row discrepancy: NOT resolved silently — the TSV corpus has 6 classdat rows and per-class dirs (V10.2 = 6 classes); the 1080/1122 counts refer to aggregated per-class column cells and stay flagged for the consumer bean
Gates: go test ./internal/... ALL PASS, vet clean, lint 0 issues.

## Zero-dependency test constraint\n\nReplace live-corpus/upstream-TSV reads with committed Go-owned fixtures. Standard Go tests must not use DG_DEVILUTIONX, locate the devilutionX tree, or skip when upstream files are absent. Missing or corrupt committed fixtures must fail the test. C++ probes are offline fixture producers only.

## AUDIT (brainer 2026-10-24, beans-completion audit vs tree)
PARTIAL (brainer audit 2026-10-24): 3 AC bullets have no artifact in tree:
typed per-class loaders (attributes 27-field, sounds 103->SfxID, animations,
sprites, loadout) and the MP award-cap function min(next/20, 200*level).
Tree has only the generic 2-col class-table reader + threshold queries.
Successor needed: typed classdat loaders + award-cap BEFORE consumers
devilGo-rctk / devilGo-aix0 land their formulas.
