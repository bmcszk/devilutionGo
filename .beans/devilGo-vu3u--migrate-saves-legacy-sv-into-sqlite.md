---
# devilGo-vu3u
title: 'Migrate saves: legacy .sv into SQLite'
status: completed
type: task
priority: normal
created_at: 2026-11-03T12:49:56Z
updated_at: 2026-11-09T21:29:01Z
parent: devilGo-o8e7
---

# Migrate saves: legacy .sv and .sp into SQLite

## Goal

ADR-014 saves subcommand on top of the ADR-003 law: `devilutionGo-migrate
saves` reads legacy .sv and .sp saves, rebuilds them through internal/persist
codecs, writes the per-hero .db files plus global.db, prints a checksum
summary per file, and stays checkpointed: a partial run resumes without
redoing converted saves.

## Research

- internal/persist: SaveGameData/ItemPack/PlayerPack codecs + importers +
  sqlite store; in-progress children there already harden this surface.
- cmd/gamedatagoldens: golden fixture pattern for verification output
  (zero-dependency policy: committed fixtures only).
- [ADR-003](../docs/adrs/adr-003.md), [ADR-014](../docs/adrs/adr-014.md).

## Scope

- saves subcommand; checkpoint journal; per-save checksum report;
  synthetic save fixtures committed as text/goldens, never binaries.

## Acceptance criteria

- [x] Given a legacy .sv or .sp, migration produces the store files that read back field-correct through the persist codecs.
- [x] Interrupted run resumes; already-converted saves are not rewritten.
- [x] go test ./... green, gofmt clean; no save binaries committed.

## Libraries

- modernc.org/sqlite (already present).

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
## Dependencies & blockers

- sqlite store work in progress inside the persistence feature; this lane
  consumes its importers unchanged.

## Open questions

- Destination path: next to the legacy file unless settings say otherwise.

## Triage

TRIAGE brainer evidence 2026-11-03: brainer REQUEST-CHANGES on the architecture
PR produced findings; this text was rewritten against the recorded law and the
verified tree (real file seams, no invented files), lint-clean.

## Summary of Changes
Worker lane feat/migrate-saves (session task-dgmigratesaves-coder, 138 calls / 0 nudges), landed at acbacdb on top of origin/master ad1eea4:
- cmd/devilutionGo-migrate/saves.go (194 lines) - usage/dispatch, legacy save classification, per-hero .db + global.db destination model
- cmd/devilutionGo-migrate/saves_import.go (463 lines) - migration pass through persist.OpenSQLiteStore/ImportSV/ImportStash (importers called unchanged), saves-scoped checkpoint journal reusing the assets journal.go pattern, sha256 per-file checksum + summary, bad-save-continues semantics (no phantom .db on failure)
- cmd/devilutionGo-migrate/saves_test.go (633 lines, 9 tests) - synthetic saves written at test time through persist codec writers; no committed binaries (MPQ containers not byte-stable upstream either, test/writehero_test.cpp:415-419)
- main.go wiring + main_test.go/cli_test.go updated; `settings` note kept for its own lane

## Proof of Work (parent verification, 2026-11-09)
- git ls-remote origin refs/heads/feat/migrate-saves = acbacdbf4643cfccdff28fb8a1854f29a63dec71 = local head (push verified by parent)
- Lane log: .hermes/lane/dg-migrate-saves/dgmigratesaves.s1.log - all seven gates green at the pushed head, in contract order (build / gofmt empty / vet / lint.sh 0 issues incl. sim purity 240 sources / go test ./... -count=1 46 pkgs 0 FAIL / just sqlc-check clean / beans-lint OK 181)
- Parent re-ran in the worktree: cmd/devilutionGo-migrate tests 119 passed (2 pkgs), beans-lint OK (181)
- AC1 evidence: TestSavesImportsLegacySavesFieldCorrect (single_/multi_/spawn saves read back field-correct via DecodePlayerPack/DecodeGameHeader/DecodeStashSave)
- AC2 evidence: TestSavesRunSkipsConvertedSavesOnRerun (migrated 0, skipped 4, byte-identical dbs), TestSavesResumesInterruptedRun, TestSavesRunReimportsChangedSave
- AC3: no save binaries in the diff (parent read the diff stat: 13 files, all .go/.md)

## Deviations recorded (fidelity gate)
- `.sp` does not exist upstream: pinned devilutionX ca91d778f pfile.* and the 1996 devilution source carry .sv families only. Implemented as .sv/.sp pair semantics per the ADR-003/014 record wording; destination remains "next to the legacy file unless settings say otherwise" (open question left for the settings lane to resolve).

## Residual
Three findings recorded (NOT patched, ownership): (1) persist ImportSV never sets ClassID; (2) persist ImportDirectory last-wins collision for stash.sv vs stash_spawn.sv at slot -1 - both filed as draft bugs under the persistence feature; (3) .sp interpretation recorded under Deviations. Brainer review pending before the PR opens; bean stays in-progress until the review verdict is handled.

## Review round + close (2026-11-09)

- Brainer review (session task-dgmigratesaves-review-brainer, verdict file .hermes/lane/dg-migrate-saves/dgmigratesaves.review.md): APPROVE (static), 9 findings minor/nit, none blocking; operator-run conditions re-run by the parent at acbacdb (cmd tests 113 pass, suite 11,456/38 pkgs, gofmt/vet/lint/sqlc/ownership diff all green).
- Fix wave (session task-dgmigratesaves-fix-coder, contract .hermes/lane/dg-migrate-saves/dgmigratesaves.s2.md): F1 help wording (1ee7677), F6 recovery hint (6a42760), F7+F8 created-ness-tracked removal both paths + RED-first tests (79f28dc), pushed to origin/feat/migrate-saves; gates re-green at head.
- Custodian sync (bb31499): the stale append-body and two draft-findings beans [persist ImportSV leaves ClassID zero on imported heroes](.beans/devilGo-sjkj--persist-importsv-leaves-classid-zero-on-imported-h.md) and [persist ImportDirectory collision stash.sv/stash_spawn.sv](.beans/devilGo-0k1m--persist-importdirectory-collision-stashsvstash-spa.md) reached the PR branch; beans-lint OK (183).
- Parent gates at the final head bb31499: `go test ./... -count=1` 11,459 pass / 38 pkgs; build+gofmt+lint+beans-lint all clean.
- PR for the user merge: https://github.com/bmcszk/devilutionGo/pull/55 (MERGEABLE at open).

Residual: none for this bean. The persist findings are tracked in their own draft beans; `_brainer triage pass owed` before they promote.

Bean references: [Migrate saves: legacy .sv and .sp into SQLite](.beans/devilGo-vu3u--migrate-saves-legacy-sv-into-sqlite.md) is the ONE bean this updater owns in this call; the two persist drafts are linked as .beans hrefs only (in the review file they also ride the PR body with ../../.beans/ hrefs).
