---
# devilGo-zvwd
title: txtdata TSV parser + table codegen
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - data
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T20:39:19Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-vof6
---

# Data tables

## Goal
Generate code + structs from txtdata TSVs (parser canonical at devilGo-jx08; fix2 seam) into typed Go tables mirroring C++ structs.

## Research
Format: assets/txtdata/Readme.md EBNF (BOM optional, tab-delimited, no tabs in values); parser Source/data/file.hpp:37-110 + record/value_reader; table structs: SpellData tables/spelldat.h:225-252, MissileData tables/misdat.h:166-196, PLStruct/UniqueItem tables/itemdat.h:628-655, MonsterData tables/monstdat.cpp:52-55, Quest tables/questdat.hpp:102-129, townerdata.

## Scope
internal/assets/txtdata: parser (BOM/CR/LF tolerant, header-mapped, typed readers with error reporting parity to reportFatalFieldData), go:generate tool producing structs from a schema file (or explicit structs + tests).

## Acceptance criteria
- [ ] test/fixtures/txtdata/*.tsv edge cases all pass
- [ ] Row counts equal C++ (monstdat 113, experience 50, etc.)
- [ ] Unknown column → error naming file+field (parity)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — codegen with text/template
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — table-driven parser tests

## Dependencies & blockers
dg-task-006.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: SPIKE-AT-CLAIM: codegen vs hand structs — time-boxed spike (1 day), pick less drift, decision recorded here before the data-package beans consume it.
- Codegen vs hand structs: spike both, pick less drift (open question H3).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:duplication-seam — parser scope overlaps jx08 (different pkg paths, pre-V0.2 body); monstdat AC count wrong (113 vs 112); re-scope to codegen+structs only.


## IMPL NOTES (triage 2026-10-16)
1 root: wave-1 task (dg-task-008): parse all txtdata TSVs into typed Go tables; draft, blocked_by devilGo-vof6 (wired ✓); blocks jx08 (wired ✓).
2 local: Readme.md format ✓; parser cites file.hpp ✓; struct cites verified: spelldat.h:225+ SpellData ✓, misdat.h:166+ MissileData ✓, itemdat.h:628+ PLStruct ✓, questdat.hpp:102+ Quest ✓; monstdat.cpp:52-55 shows the MonstersData vectors, not the MonsterData struct (struct is in monstdat.h) — minor cite nit. Three real gaps: (a) AC "Row counts equal C++ (monstdat 113...)" is wrong — 112 data rows (V12.12 append.md:3738 says 112; awk says 112; 113 = physical lines incl. header); (b) duplication seam with jx08: zvwd scope = parser (BOM/CR/LF, header-mapped, typed readers) + codegen in `internal/assets/txtdata`, jx08 scope = same parser job + V0.1/V0.2 semantics in `internal/txtdata` — two packages, overlapping parser ownership; jx08's body declares the split ("dg-task-008 owns the codegen scaffold; this task owns the semantics") but zvwd's body doesn't reflect it (written 2h earlier, pre-V0.2 discovery); (c) zvwd lacks the hf/override dimension entirely (pre-dates V0.2).
3 online: n/a (no external claims).
4 solution: re-scope zvwd to the codegen/typed-struct layer ONLY (schema→struct generation, per-domain row structs mirroring the cited C++ structs, row-count/field parity tests), consume jx08's loader for parsing; drop its own parser/typed-reader scope and the `internal/assets/txtdata` path (use jx08's `internal/txtdata` or `internal/gamedata` per Phase-7 convention); fix monstdat count 113→112; add hf-variant counts from V12.12 to its row-count AC.
5 understanding: diablo-layer table structs; the generator/consumer of jx08's mechanism — the pair covers what each alone claims.
6 repro: `beans show devilGo-zvwd --body-only | head -30` vs `beans show devilGo-jx08 --body-only | sed -n '/## Scope/,/## Acceptance/p'`; `awk 'NR>1' dX/assets/txtdata/monsters/monstdat.tsv | wc -l`.
7 fix: `beans update devilGo-zvwd --body-replace-old "Parse all txtdata TSVs into typed Go tables mirroring C++ structs." --body-replace-new "Generate typed Go table structs mirroring C++ structs (schema->struct codegen + parity tests); parsing/V0.1/V0.2 semantics owned by devilGo-jx08 (dg-task-101)."` + replace "monstdat 113" with "monstdat 112 (V12.12)" + retarget scope path.
8 validation: zvwd body names jx08 as parser owner; no parser implementation in its scope; count reads 112.
9 related: jx08 (canonical parser — see verdict there), vof6 (dep), 2j8j..b8z9 (per-domain tables that consume both).
VERDICT: NEEDS-WORK:duplication-seam — parser scope overlaps jx08 (different pkg paths, pre-V0.2 body); monstdat AC count wrong (113 vs 112); re-scope to codegen+structs only.

---

devilGo-b8z9: NEEDS-WORK:wire-deps (4 missing tracker edges; content verified excellent)
devilGo-n89r: TRIAGED-OK
devilGo-l5fp: NEEDS-WORK:AC-count-errors+wire-deps (113→112 monstdat; font 76→64; 3 missing edges)
devilGo-vof6: NEEDS-WORK:external-gap-unrecorded (go-mpq v0.3.1 Huffman/ADPCM ❌ vs extract-every-file AC; DECIDED #2 reconciliation)
devilGo-b782: NEEDS-WORK:wire-deps (1 missing edge b782→66o7)
devilGo-ow6p: NEEDS-WORK:wrong-dep (catalog dep should be b8z9 not zvwd; missing tracker edge; row count nit)
devilGo-jx08: TRIAGED-OK
devilGo-zvwd: NEEDS-WORK:duplication-seam (parser overlaps jx08; monstdat 113→112; missing hf dimension)

BATCH-03-COMPLETE


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Seam resolution (triage): parser scope overlaps devilGo-jx08 — canonical parser = jx08; zvwd = codegen+structs only (pre-V0.2 body superseded).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.


