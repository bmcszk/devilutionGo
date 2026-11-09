---
# devilGo-sjkj
title: persist ImportSV leaves ClassID zero on imported heroes
status: completed
type: bug
priority: high
created_at: 2026-11-09T20:58:46Z
updated_at: 2026-11-09T23:39:49Z
parent: devilGo-6xu2
---

# persist ImportSV leaves ClassID zero on imported heroes

## Symptom (found by the [saves-migrate lane]([Migrate saves: legacy .sv and .sp into SQLite](.beans/devilGo-vu3u--migrate-saves-legacy-sv-into-sqlite.md)), 2026-11-09)

`internal/persist.ImportSV` never sets ClassID on the imported hero. Recorded
verbatim in .hermes/lane/dg-migrate-saves/dgmigratesaves.s1.log ("Recorded"
section, session task-dgmigratesaves-coder).

Impact: any caller importing through ImportSV gets a hero row whose class
fields read wrong/zero in the store until it also goes through
DecodePlayerPack; other future callers will silently store the wrong class.

## Root-cause hypothesis (needs verification)

Internal seam: internal/persist/importsv.go hero mapping. Verify against the
codec fields (PlayerPack.Encode) and the hero-table insert before scoping.

## Triage needed

Root cause not yet verified. Do not patch first: research, scope, AC, then
promote.

# Triage (brainer read-only wave, 2026-11-11; verified against the live tree)

## Root cause (verified live, file:line)
internal/persist/importsv.go:43-77 — ImportSV fills hero.Slot (:51) and calls fillHeroName(&hero) (:56), whose body (:127-141) decodes the pack but copies ONLY pack.Name into the record; hero.ClassID is never assigned on the import path. LegacySVStore.Read (internal/persist/store_legacy.go:108-140) builds HeroRecord{Pack: pack} setting only Spawn/Hellfire/Multiplayer. The zero flows through heroToParams (store_sqlite_map.go:60-73) into heroes.class_id (schema.sql NOT NULL DEFAULT 0). The pack is the sole authoritative source: PlayerPack.Class at internal/persist/playerpack_decode.go:39 (pack.h:46-59). Upstream sets the class from the pack on the same seam (ca91d778f: pfile.cpp:215 via Game2UiPlayer; pfile.cpp:694-727 pfile_ui_set_hero_infos).

## Local research
grep ClassID internal/persist → store/sqlc map + schema only, no importer writes it; cmd matches zero. queries/store.sql UpsertHero already writes class_id on insert + update (store side ready); the gap is purely the record fill. validateHero (store_sqlite_map.go:76-90) has no class gate; test gap confirmed: importdir_test.go:147-173 and saves_test.go:675-690 assert pack.Class via decode, never row.ClassID — the field-correct AC cannot see this defect. Export surface unaffected (ExportSV writes hero.Pack verbatim, importsv.go:103-125).

## Online research
None needed — the pinned devilutionX clone (~/github/bmcszk/devilutionX @ ca91d778f) is the authoritative source per the repo law; pfile.cpp/pack.h verified on disk.

## Solution + class prevention
In ImportSV hoist the single DecodePlayerPack and fill both row columns: rename fillHeroName → fillHeroColumns(hero), set hero.ClassID = int(pack.Class) alongside hero.Name (cites pfile.cpp:215, pack.h). Prevention: every HeroRecord column kept "for list reads without decoding the pack" is filled by the importer at the same seam; a RED-first row/pack pair test makes future column additions fail until filled.

## Understanding
The heroes row duplicates two pack fields (name, class) so the character-select list can read rows without decoding every encrypted pack. ImportSV decoded the pack once for Name and stopped one field short: every imported row reads class 0 (= Warrior) regardless of the real class. Latent today (no live consumer), wrong-class list the moment the UI port lands.

## Repro
pack := goldenPlayerPack() with pack.Class = 2; write single_0.sv via LegacySVStore.Write (pattern importdir_test.go:224-247); persist.ImportSV(ctx, store, path, GameMode{}, 0); store.LoadHero(ctx, 0).ClassID == 0 while DecodePlayerPack(hero.Pack).Class == 2 (SQL: SELECT class_id → 0).

## Fix steps
(1) RED: TestImportSVFillsClassID in internal/persist (importsv_test.go) per the repro; run to see it fail. (2) importsv.go: fillHeroColumns sets Name + ClassID; keep the pfile.cpp cites. (3) GREEN + optional CLI pin: extend TestSavesImportsLegacySavesFieldCorrect (saves_test.go:675-690) with a row-ClassID assertion on the class-2 fixture. (4) No store/schema/sqlc/caller changes (minimal-change law).

## Validation
go test ./internal/persist/... -count=1 (red→green, TestImportExportSV/RGS suite stays green); go test ./cmd/devilutionGo-migrate/... -count=1; gofmt empty, go vet clean, ./scripts/lint.sh, just sqlc-check clean (no query change); ExportSV round-trips via CompareSaves semantics.

## Related beans
parent [Persistence: .sv byte-compat + SQLite store + pfile lifecycle](.beans/devilGo-6xu2--persistence-sv-byte-compat-sqlite-store-pfile-life.md); found by [Migrate saves: legacy .sv and .sp into SQLite](.beans/devilGo-vu3u--migrate-saves-legacy-sv-into-sqlite.md) (its Residual names this finding); family: xwtt, 0k1m, ym0r. No duplicate in the ledger (scanned; only this bean names ClassID).

## Acceptance criteria

- [x] RED-first row test: TestImportSVFillsClassID fails before the fix on a class-2 fixture (pack.Class=2, row class_id 0), passes after.
- [x] ImportSV fills hero.ClassID = int(pack.Class) at the same decode it already performs for Name (pfile.cpp:215 cite in the comment).
- [x] CLI pin: the saves-field-correct test asserts the ROW class for a class-2 fixture.
- [x] Gates green: persist suite + cmd suite + gofmt + vet + lint.sh + sqlc-check (no schema/query change).

## Libraries

- modernc.org/sqlite (already present, unchanged); stdlib only.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-tests/SKILL.md — red-first row/pack pair test style

## Progress log

- Claimed 2026-11-11 parent (persist-fix wave, grouped lane): [fix 1] ImportSV row fill + [fix 2] stash family slot model - one PR, per the grouped-feature PR law (same imported-row mechanism). Coder dispatch pending; PR opens after the settings PR (#56) merges per the wave-serialization law.

/home/blaze/.hermes/cache/<append slurped the scratch path literally; content restored below>

## Progress log

- 2026-11-11 lane feat/imported-rows (task-importedrows-coder): RED run logged (row ClassID = 0, want 2), fix at the named seam, GREEN + CLI pin; pushed 9738abd.

## Summary of Changes

- internal/persist/importsv.go: fillHeroName → fillHeroColumns(hero); hero.ClassID = int(pack.Class) filled from the same decode (comment cites pfile.cpp:215). No store/schema/sqlc/caller edits.
- internal/persist/importsv_test.go: TestImportSVFillsClassID (RED-first evidence in the lane log: row ClassID = 0, want 2).
- cmd/devilutionGo-migrate/saves_test.go: the field-correct assertion pins the ROW class for the class-2 fixture.

## Proof of Work (parent verification, 2026-11-11)

- git ls-remote origin refs/heads/feat/imported-rows == 2875aa84499ea8162f1c34688c1f45bba65796f4 == local head (push verified by parent).
- Parent re-ran at 2875aa8: `go test ./internal/persist/... ./cmd/devilutionGo-migrate/... -count=1` 1117 pass / 3 pkgs; `go test ./... -count=1` 11,456 pass / 38 pkgs; gofmt empty; lint.sh 0 issues; beans-lint OK (183).
- Brainer review (task-importedrows-review-brainer): APPROVE (static), 6 minor/nit findings, none on this bean's semantics; fix-not needed; the two comment nits (F1/F2/F6) ride the s2 comment unit.
- Reviewer + parent confirm the RED evidence predates the fix and the fix commit 9738abd turns it green.

## Deviations recorded (fidelity gate)

- The row-truth design: heroes row duplicates two pack fields (name, class) so a character-select list can read rows without decoding packs - mirrors upstream's Game2UiPlayer (pfile.cpp:215). Deviation: none; the fix restores 1:1 with upstream's row fill.

Residual: none for this bean.
