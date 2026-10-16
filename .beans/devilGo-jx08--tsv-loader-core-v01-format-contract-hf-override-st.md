---
# devilGo-jx08
title: 'TSV loader core: V0.1 format contract + hf override stack (shared data pkg)'
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T14:06:55Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-lfvi
    - devilGo-zvwd
---

# TSV loader core

## Goal
Build the single shared Go package every Phase-7 data table reads through: a TSV reader that
implements the V0.1 format contract exactly, the `id`/`type` column-skip rule, and the V0.2
Hellfire mod **override + append** semantics. No per-domain row data lives here — only the
mechanism. This is the blocker for all of dg-task-102..118.

## Research
- `/tmp/dg-requirements-append.md` **V0.1** (format contract, verbatim from
  `assets/txtdata/Readme.md`) and **V0.2** (the mod-override mechanism Volume I missed).
- C++ parser: `Source/data/file.hpp:37-110` (`DataFile::loadOrDie`, `parseHeader`,
  `DataFileRecord`) and `Source/data/record_reader.hpp` (`readInt`/`readBool`/`readString`/
  `read`/`readEnumList`/`readEnumArray`).
- Column-skip rule: **every** loader calls `reader.advance()` to skip column 1 (`id`/`type` are
  documentation/provenance, never parsed); positions come from row order + auto-assigned
  `mappingId` (`Source/tables/itemdat.cpp:585-600`, `:656-666`).
- Override stack: `Source/options.cpp:71-80` `DiscoverMods()` inserts `"hf"` when
  `HaveHellfire()`; `Source/engine/assets.cpp:704-760` `LoadModArchives()` pushes loose
  `PrefPath/mods/<name>/` (`:729-733`) then `BasePath/mods/<name>/` (`:736-740`) into
  `OverridePaths` and loads packed `mods/<name>.mpq` at priority >= 10000, above retail (1000)
  and devilutionx.mpq (9000).
- On-disk inventory (verified with `find`): `assets/txtdata/` = **49** `.tsv` + `Readme.md` +
  `classes/README.md`; `mods/hf/txtdata/` = **14** `.tsv`. Only `mods/hf/` exists under `mods/`.
- Table->loader map: the 21-row table in V0.1.

## Scope
`internal/txtdata/` (shared pkg, Go 1.22+):
- `tsv.go`: tab-delimited UTF-8, optional leading BOM, exactly one header row, no tabs inside
  values, `-` or empty cell == absent; strict error on arity mismatch.
- `record.go`: typed readers mirroring `record_reader.hpp` incl. `readEnumList` /
  `readEnumArray` (fixed-width arrays such as `frameDelay[16]`).
- `override.go`: resolution order base -> mod override, with **per-file whole-file replacement**
  (that is what the C++ does: an overriding TSV replaces the base file; deltas in V12.10 are
  emergent, not merged). Expose `Load(table string, hf bool) ([]Row, error)`.
- `paths.go`: the 49 base paths + 14 hf paths as one generated manifest table.
- Cross-reference (do not duplicate): `dg-task-008` (epic-02 "txtdata TSV parser + table
  codegen") owns the codegen scaffold; this task owns the V0.1/V0.2 **semantics** and the
  base/hf resolution used by every Phase-7 table.

## Acceptance criteria
- [ ] TDD first: `internal/txtdata/manifest_test.go` asserts the file manifest == 49 base + 14
      hf `.tsv` files found by `find` on disk, and asserts the aggregate base data-row count
      equals the V12.12 table (49 files, per-file counts as listed in this bean tree) — red
      before any loader code exists.
- [ ] Round-trip: re-encoding every parsed row reproduces the source file byte-for-byte
      (BOM, header, `-` cells preserved) for all 63 TSVs.
- [ ] Column-skip parity: for each of the 21 loaders in the V0.1 table, the Go reader consumes
      columns 2..N in the same order as the cited C++ and ignores column 1 (verified against a
      probe binary).
- [ ] Override semantics: with `hf` enabled, `Load("monsters/monstdat.tsv", true)` returns the
      hf file's 138 rows (not a 112-row base plus merge), and with `hf` disabled returns 112 —
      the V0.2 whole-file-replacement rule.
- [ ] `readEnumArray` decodes `missile_sprites.tsv` `frameDelay[16]`/`frameLength[16]` into
      fixed-width arrays with identical de-dup indices to `misdat.cpp:47-55` `ToIndex`.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — package layout, no reflection in the hot path
- /home/blaze/.agents/skills/coding-rules/SKILL.md — reader (intent) vs typed getters (detail)
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — round-trip fixture tests

## Dependencies & blockers
blocked_by dg-task-001 (module scaffold), dg-task-008 (TSV parser + table codegen scaffold).
Blocks dg-task-102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 118
(every table-loading task).

## Open questions
- Does the Go port need `PrefPath/mods/<name>/` loose overrides at all, or only the built-in
  `hf` set? V0.2 implies both; decide before implementing `override.go` ordering.
- effects-unused.tsv under hf is **1 row** vs base **15** (V12.10: "replaced") — confirm the
  intended semantics is "no unused rows to report", not a truncated file.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK.


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-7 shared-mechanism task (dg-task-101) for all table tasks; draft, blocked_by [lfvi, zvwd] (wired ✓); dependents wired: 2j8j/778h/pu3r/h6wn/i2cr blocked_by jx08 ✓ (103→via 102, 105→via 104: transitively ordered, acceptable).
2 local: outstanding verification pass — V0.1 (append.md:24, 21-row loader table = 23 grep rows incl. header/sep ✓) & V0.2 (:56) exist; file.hpp:37+ DataFile BOM handling ✓; record_reader.hpp has readInt/readBool/readString/read/readEnumList/readEnumArray ✓; itemdat.cpp:585-600 col-2..N reads + :656-666 mappingId auto-assign ✓; options.cpp:71-80 DiscoverMods inserts "hf" ✓; assets.cpp:725-745 PrefPath/BasePath mods loose overrides ✓; on-disk: 49 base + 14 hf .tsv, mods/ contains only hf ✓; monstdat 112/138 data rows on disk ✓ (AC matches); missile_sprites header frameDelay/frameLength 16-elem lists ✓ + misdat.cpp:47-55 ToIndex de-dup ✓; Readme.md EBNF (BOM optional, trailing-newline flexibility) ✓ — round-trip AC must preserve the no-trailing-newline case (base effects.tsv lacks it; hf has it) — worth one test case. All 4 skills exist ✓. Open questions (loose-mods need; hf effects-unused 1-row semantics) are correctly scoped as decisions.
3 online: n/a (no external claims).
4 solution: none blocking; optionally add "round-trip preserves absent trailing newline (base effects.tsv) and present one (hf)" as an explicit test case.
5 understanding: engine-layer shared data mechanism; correct whole-file-replacement reading of V0.2 (override shadows base, deltas emergent) matches C++ OverridePaths+archive-priority behavior.
6 repro: `find dX/assets/txtdata -name '*.tsv' | wc -l` (=49); `find dX/mods -name '*.tsv' | wc -l` (=14); `awk 'NR>1' dX/mods/hf/txtdata/monsters/monstdat.tsv | wc -l` (=138); `sed -n '47,55p' dX/Source/tables/misdat.cpp`.
7 fix: none (optional test-case note above).
8 validation: n/a.
9 related: zvwd (codegen counterpart — see seam issue there), 2j8j..i2cr (consumers), b8z9 (downstream).
VERDICT: TRIAGED-OK.
