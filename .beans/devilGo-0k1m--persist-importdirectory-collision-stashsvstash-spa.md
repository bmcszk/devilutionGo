---
# devilGo-0k1m
title: persist ImportDirectory collision stash.sv/stash_spawn.sv
status: in-progress
type: bug
priority: high
created_at: 2026-11-09T20:58:46Z
updated_at: 2026-11-09T23:03:32Z
parent: devilGo-6xu2
---

# persist ImportDirectory collision drop between stash.sv and stash_spawn.sv

## Symptom (found by the [saves-migrate lane]([Migrate saves: legacy .sv and .sp into SQLite](.beans/devilGo-vu3u--migrate-saves-legacy-sv-into-sqlite.md)), 2026-11-09)

`internal/persist.ImportDirectory` resolves both stash.sv and stash_spawn.sv
against the same stash slot (-1) with last-wins semantics; importing a
directory that contains both silently drops one of them. Recorded verbatim in
.hermes/lane/dg-migrate-saves/dgmigratesaves.s1.log (session
task-dgmigratesaves-coder).

Impact (corrected by the 2026-11-11 triage below): the loss is NOT limited to ImportDirectory - the saves subcommand writes both stash files into the SAME global.db slot -1 (saves_import.go:446), so both call paths share the silent spawn-stash loss; fix scope covers BOTH callers plus the slot/row model.

## Root-cause hypothesis (needs verification)

Family-slot model: slot -1 carries a pair of stash families; either the slot
model needs a second slot for the spawn variant, or the last-wins loop needs
to keep both when the file name maps to a distinct family. Verify at
ImportDirectory before scoping.

## Triage needed

Root cause not yet verified. Do not patch first: research, scope, AC, then
promote.

# Triage (brainer read-only wave, 2026-11-11; corrected by parent with the brainer's verified remedy)

## Root cause (verified live, file:line)
Upstream save format keeps a FAMILY PAIR inside ONE row: stash slot -1 carries stash.sv AND stash_spawn.sv together (ca91d778f: pfile.cpp:96-104 — one stash record with cook-file variants; both files target GlobalStashSlot). internal/persist: ImportDirectory iterates ReadDir order and upserts each stash file against hero_slot -1 (importdir.go:51-70, 117-126; importstash.go:9-12); queries/store.sql:34-36 ON CONFLICT(hero_slot) UPDATE = last-wins, so the second file silently overwrites the first's stash rows. cmd/devilutionGo-migrate/saves_import.go:446 does the same (both names → GlobalStashSlot) — earlier body claim "saves sidesteps" corrected below.

## Correction of the original impact claim
The bean's "saves subcommand sidesteps" was FALSE on the live tree: the CLI shares the loss (saves_import.go:436-446 builds mode per family but writes the same slot). Corrected scope: BOTH callers + the slot/row model.

## Local research
committed goldens: persist/testdata hold single-slot fixtures only (writeLegacySaves writes stash.sv only) — the pair case is untested everywhere. The store's heroes table keys slot; a second column (or a slot -2 row for the spawn вкуs) has no schema; the .db hero item/levels tables are slot-keyed too — check what a -2 slot touches before choosing the model (child of the fix: enumerate schema keys).

## Online research
None needed — pinned local clone per the repo law.

## Solution + class prevention (remedy direction, verified + correct per the brainer)
Choose ONE model and pin it in a test: (a) slot -2 dedicated spawn-stash row (schema-parallel, keep both files), or (b) single row carrying the family pair (schema change: stashes table). The brainer's second-slot remedy (a) is the smaller move and matches the family pair semantics. Prevention: both ImportDirectory and the CLI importer must share ONE stash-load seam so family resolution lives in a single place.

## Understanding
The roster row model assumed one global stash; upstream has a family pair (retail + spawn variants). Every multi-stash importer today loses the spawn stash silently.

## Repro
Dir with stash.sv + stash_spawn.sv; persist.ImportDirectory(ctx, store, dir); LoadStash(ctx, GlobalStashSlot) contains only the LATER file's rows (seed-ordered by ReadDir). Same via `devilutionGo-migrate saves --dir <that dir>`: global.db holds the last file only.

## Fix steps
(1) RED: a two-file directory import test pinning the desired family model (both files' row contents separately addressable/retained). (2) Decide slot -2 row vs second column in the stash table; wire both importers through the ONE shared seam. (3) GREEN + CLI pin: saves_test.go corridor for the pair. (4) Document the slot model decision in ADR/bean (fidelity gate).

## Validation
persist suite + cmd suite green; gofmt/vet/lint.sh/sqlc-check clean; a rerun of the saves lane's resume tests stays green (journal fingerprints unaffected).

## Related beans
parent [Persistence](.beans/devilGo-6xu2--persistence-sv-byte-compat-sqlite-store-pfile-life.md); [Migrate saves: legacy .sv and .sp into SQLite](.beans/devilGo-vu3u--migrate-saves-legacy-sv-into-sqlite.md); [persist ImportSV leaves ClassID zero on imported heroes](.beans/devilGo-sjkj--persist-importsv-leaves-classid-zero-on-imported-h.md). No duplicate.

## Acceptance criteria

- [ ] RED-first: a directory holding stash.sv + stash_spawn.sv imports BOTH family sets (test pins the chosen slot model; currently one is silently lost).
- [ ] The two importers (ImportDirectory, migrate saves subcommand) resolve stash families through ONE shared seam (single place where family→slot/store decision lives).
- [ ] The slot/row model decision is recorded (bean + ADR/deviation note per the fidelity gate).
- [ ] Existing stash tests (single-file corpus) stay green; gates all clean.

## Libraries

- modernc.org/sqlite (already present); stdlib only.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-tests/SKILL.md — red-first family-pair test style

## Progress log

- Claimed 2026-11-11 parent (persist-fix wave, grouped lane): [fix 1] ImportSV row fill + [fix 2] stash family slot model - one PR, per the grouped-feature PR law (same imported-row mechanism). Coder dispatch pending; PR opens after the settings PR (#56) merges per the wave-serialization law.
