---
# devilGo-0k1m
title: persist ImportDirectory collision stash.sv/stash_spawn.sv
status: draft
type: bug
priority: high
created_at: 2026-11-09T20:58:46Z
updated_at: 2026-11-09T20:58:46Z
parent: devilGo-6xu2
---

# persist ImportDirectory collision drop between stash.sv and stash_spawn.sv

## Symptom (found by the [saves-migrate lane]([Migrate saves: legacy .sv and .sp into SQLite](.beans/devilGo-vu3u--migrate-saves-legacy-sv-into-sqlite.md)), 2026-11-09)

`internal/persist.ImportDirectory` resolves both stash.sv and stash_spawn.sv
against the same stash slot (-1) with last-wins semantics; importing a
directory that contains both silently drops one of them. Recorded verbatim in
.hermes/lane/dg-migrate-saves/dgmigratesaves.s1.log (session
task-dgmigratesaves-coder).

Impact: a retail-directory save import loses the spawn-stash rows; saves
subcommand sidesteps by importing per stash family directly, but any
caller whose input directory contains both files will take a silent loss.

## Root-cause hypothesis (needs verification)

Family-slot model: slot -1 carries a pair of stash families; either the slot
model needs a second slot for the spawn variant, or the last-wins loop needs
to keep both when the file name maps to a distinct family. Verify at
ImportDirectory before scoping.

## Triage needed

Root cause not yet verified. Do not patch first: research, scope, AC, then
promote.
