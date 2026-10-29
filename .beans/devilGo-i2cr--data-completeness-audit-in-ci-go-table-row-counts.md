---
# devilGo-i2cr
title: 'Data-completeness audit in CI: Go table row counts vs TSV (base + hf)'
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-29T16:15:59Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-jx08
    - devilGo-2j8j
    - devilGo-baxb
    - devilGo-778h
    - devilGo-v791
    - devilGo-vhgd
    - devilGo-1e1b
    - devilGo-rctk
    - devilGo-lsz6
    - devilGo-aix0
    - devilGo-h6wn
    - devilGo-b8z9
    - devilGo-pu3r
    - devilGo-fpkz
    - devilGo-orkj
    - devilGo-cciu
    - devilGo-1spo
---

# Data-completeness audit gate ("every detail" gate)

## Goal
Make Phase 7 falsifiable: a CI job that compares every Go table's row count and first/last-row
fingerprint against the TSVs on disk, in **both** the base and `hf` variants, and fails on any
missing, extra, or drifted table. This is the user's "every detail is accounted for" gate.

## Research
- `docs/research/dg-requirements-append.md` **V12.12** (the transcription checklist: "a `go test` per table
  asserting the count plus first/last-row field equality closes this volume's acceptance
  criteria"), **V12.10** (hf deltas), **V0.1** (loader map = the authoritative table list),
  **V0.2** (override semantics).
- Reference counts (all re-measured on disk with `tail -n +2 <file> | grep -c '[^[:space:]]'`,
  base / hf): itemdat 168/—, item_prefixes 83/86, item_suffixes 95/98, unique_itemdat 90/110,
  monstdat 112/138, unique_monstdat 100/—, misdat 68/108, missile_sprites 46/59, objdat 109/—,
  questdat 24/—, spelldat 36/51 (+injected Null => 37/52), textdat 338/—, textdat-unused 139/—,
  towners 12/15, quest_dialog 10/13, effects 814/917, effects-unused 15/1, Experience 50/—,
  classdat 6/6, and per class attributes 27/27, sounds 103/103, animations 27/27, sprites 13/13,
  starting_loadout 9/9 (30 per-class files). File totals: **49** base + **14** hf `.tsv` files.
- Domain totals: items **436**, monsters **212**, missiles **114**, objects **109**, quests
  **24**, spells **36**, text **477** (338+139), town **22** (12+10), audio **829** (814+15),
  XP **50**, classes **1080** (6 + 1074) — *not* the 1122 the inventory's V12.12 invariant line
  states; devilGo-pu3r records that correction and this gate asserts 1080.
- Registry (non-TSV) counts to include: AiProc **40** entries with 1 nil (`monster.cpp:3089-3130`),
  misdat Add/Process name sets 51/37 base and 74/46 hf, `ShrineType` 34, `ThemeID` 17,
  `TalkID` 24, `_talker_id` 13 vs towners rows 12/15, `cursor_id` 13, `QoL` files 16,
  music tracks 8 (+NUM_MUSIC), movies 10, `MissileID` 111, `SpellID`, `_unique_items` 91,
  `_speech_id`.
- CI wiring: existing `devilGo-lfvi` (justfile + CI), `devilGo-uc44` (timedemo + seed matrix job),
  `devilGo-w6qf` (bench), `devilGo-l5fp` (graphics asset manifest/coverage checker — same shape,
  different domain: reuse its pattern, do not duplicate).

## Scope
`internal/gamedata/audit/`: `cmd/dataaudit` (a Go program that (1) walks `assets/txtdata` and
`mods/hf/txtdata`, (2) loads every registered Go table in both variants, (3) emits a
machine-readable report `table, base_rows_tsv, base_rows_go, hf_rows_tsv, hf_rows_go,
first_row_sha, last_row_sha, status`) and `audit_test.go` that fails CI on any `status != OK`.
Registration API so a new table that is not registered is itself a failure (closed world). Also
a `just data-audit` target and a CI step.

## Acceptance criteria
- [ ] TDD first: the audit test is written and **failing** before the tables exist — it must
      enumerate all 49 base + 14 hf files and report every unregistered table as a failure, and it
      is green only when each Go table's row count equals its TSV count **including hf deltas**
      (itemdat 168/168, item_prefixes 83/86, item_suffixes 95/98, unique_itemdat 90/110, monstdat
      112/138, unique_monstdat 100/100, misdat 68/108, missile_sprites 46/59, objdat 109/109,
      questdat 24/24, spelldat 37/52 incl. Null, textdat 338/338, towners 12/15, quest_dialog
      10/13, effects 814/917, Experience 50/50, classdat 6/6, per-class 27/103/27/13/9).
- [ ] Closed-world check: adding a `.tsv` file to `assets/txtdata/` without registering a Go table
      fails the audit (tested by a fixture directory), and dropping one fails it too.
- [ ] Fingerprint parity: for every table, first-row and last-row field hashes equal the TSV in
      both variants — proves alignment and the column-1-skip rule, not just counts (a row-shifted
      table must fail here even with the right count).
- [ ] Registry parity worked example: the audit asserts AiProc == 40 entries / 1 nil, misdat name
      sets == 51/37 base and 74/46 hf, `ShrineType` == 34, `TalkID` == 24, movies == 10, QoL
      modules == 16 — i.e. non-TSV counts are covered by the same gate.
- [ ] Report is diffable in CI: the job prints a per-table diff vs the *expected* manifest
      (devilGo-cciu's delta manifest) and fails on any change to the expected manifest without a
      recorded decision; run time budget < 60 s so it stays in the PR gate.
- [ ] `just vet test lint` and `just data-audit` are both green with `hf` toggled off **and** on,
      and the PoW records the full report output.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — codegen/registration APIs, closed-world design
- /home/blaze/.agents/skills/go-fluent-testing/SKILL.md — fixture-driven CI tests + skips
- /home/blaze/.agents/skills/coding-rules/SKILL.md — one report struct, no ad-hoc prints
- /home/blaze/.agents/skills/beans/SKILL.md — PoW must contain the actual audit output

## Dependencies & blockers
blocked_by devilGo-jx08 and by every Phase-7 table/registry bean: 102, 103, 104, 105, 106, 107,
108, 110, 111, 112, 113, 114, 115, 116, 117. (Practical note: land the audit harness early as a
red baseline and let it turn green incrementally — the gate is only useful if it exists before
the tables do.) Also needs devilGo-lfvi (CI wiring) and devilGo-cciu (expected delta manifest).
Partially overlaps devilGo-uc44 (seed matrix job) and devilGo-l5fp (graphics manifest); reuse their
CI shape, do not duplicate.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: DECIDED: TSV-only audit + first/last-row fingerprints. Probe-binary comparison rejected: probe is owned by the fixture toolchain bean for goldens, audit stays self-sufficient.
- Do we want the audit to compare against the C++ *loader* output (needs a probe binary) or only
  against the TSVs? TSV-only is self-sufficient and cheap; the first/last-row fingerprint closes
  most drift. Probe-based comparison belongs to devilGo-uc44's cross-check sidecar.
- Should `effects-unused`/`textdat-unused` be *shipped* tables or audit-only fixtures? They are
  dead data; the audit still needs to account for their rows (829/477 totals) — confirm the
  audit-only recommendation.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: The "every detail" gate — closed-world CI audit of every Go table vs TSV row counts + first/last-row fingerprints, base and hf; status `draft`, parent 1n0a, blocked_by 16 Phase-7 table beans.
2 local: Measured on disk, all match bean's claims: base/hf `.tsv` file totals **49/14** (`find assets/txtdata -name '*.tsv' | wc -l` → 49; `mods/hf/txtdata` → 14); spelldat 36/51 (+AddNullSpell at `Source/tables/spelldat.cpp:20-35`, called `:236` → 37/52); monstdat 112/138; effects 814/917; AiProc = **40 entries, 1 nullptr (FireMan)** (`sed -n '3089,3130p' Source/monster.cpp | grep -c MonsterAIID::` → 40, `grep -c nullptr` → 1); class rows 27/103/27/13/9 per file (warrior measured) → 6+6×179=**1080**; misdat distinct addFn/processFn **51/37 base, 74/46 hf** (awk measured). All 16 frontmatter blocked_by ids exist and map to Phase-7 table beans (edges.json verified). TDD-first AC explicit (audit red before tables exist).
3 online: n/a (no external claims).
4 solution: none — ready. Optional cosmetic: body "Dependencies" prose still cites stale devilGo-jx08..117 (frontmatter governs; batch-02 common finding).
5 understanding: the falsifiability backbone of the whole Phase-7 data port; supersedes V12.12's incorrect 1122 with measured 1080 (consistent with pu3r/hy8e).
6 repro: `beans show devilGo-i2cr --json`; `cd devilutionX && find assets/txtdata -name '*.tsv' | wc -l` (49); `tail -n +2 assets/txtdata/spells/spelldat.tsv | grep -c '[^[:space:]]'` (36).
7 fix: none.
8 validation: n/a (no change requested).
9 related: devilGo-yv1r (probe harness feeding it), devilGo-hy8e (errata it consumes), devilGo-cciu (hf deltas).
VERDICT: TRIAGED-OK

## Zero-dependency test constraint\n\nReplace live-corpus/upstream-TSV reads with committed Go-owned fixtures. Standard Go tests must not use DG_DEVILUTIONX, locate the devilutionX tree, or skip when upstream files are absent. Missing or corrupt committed fixtures must fail the test. C++ probes are offline fixture producers only.

\n\n## Proof of Work (audit harness baseline)\n\n- Delegated slices (pi RPC, session devilGo-i2cr-t5 lineage): audit.go (Entry/Manifest/LoadManifest), audit_manifest.json (15 entries, every WantRows measured on the committed TSVs with tail -n +2 | grep -c), audit_test.go (TestRowCountsMatchManifest, TestFingerprintsMatch with 15 subtests, TestCorruptedManifestFailsWithTableAndVariant), justfile data-audit recipe (HAS_GO_SRC guarded). ci.yml verified: ?   	github.com/bmcszk/devilutionGo/cmd/devilutionGo	[no test files]
?   	github.com/bmcszk/devilutionGo/cmd/devilutionGo-headless	[no test files]
?   	github.com/bmcszk/devilutionGo/internal/diablo	[no test files]
ok  	github.com/bmcszk/devilutionGo/internal/diablo/animation	0.005s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/combat	0.005s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/enums	0.004s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio	0.008s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audit	0.008s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat	0.006s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat	0.006s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/monstdat	0.007s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/objdat	0.006s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/quests	0.006s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/textdat	0.005s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/gamedata/towners	0.006s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/missiles	0.031s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/save	0.006s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/testsupport	0.061s
ok  	github.com/bmcszk/devilutionGo/internal/diablo/txtdata	0.005s
ok  	github.com/bmcszk/devilutionGo/internal/engine	0.198s
ok  	github.com/bmcszk/devilutionGo/internal/engine/assets	0.011s
ok  	github.com/bmcszk/devilutionGo/internal/engine/demo	0.005s
ok  	github.com/bmcszk/devilutionGo/internal/engine/txtdata	0.005s
ok  	github.com/bmcszk/devilutionGo/internal/game	0.005s
?   	github.com/bmcszk/devilutionGo/internal/game/testsupport	[no test files]
ok  	github.com/bmcszk/devilutionGo/internal/game/world	0.006s
ok  	github.com/bmcszk/devilutionGo/internal/testsupport/parity	0.005s already covers the audit package \u2014 no workflow change.\n- Zero-dependency constraint honored: loaders read committed internal/testdata/gamedata TSVs only; no DG_DEVILUTIONX, no skips for missing upstream.\n- Parent gates: gofmt clean; go vet exit 0; go test ./internal/... 518 passed / 25 pkgs; golangci-lint 0 issues; just data-audit \u2192 ok. Commit 02335a6 pushed.\n- Coverage gap (recorded, not silent): itemdat/item_prefixes/item_suffixes/unique_itemdat/experience/classdat/per-class files have no Go loaders yet \u2014 manifest covers only existing tables per contract; the audit turns green incrementally as table beans land.\n- Residual: first/last-row fingerprints are raw-TSV rows (parsed-row comparison arrives with the yv1r probe goldens); hf misdat delta covered via manifest entries 68/108.\n
