---
# devilGo-sjkj
title: persist ImportSV leaves ClassID zero on imported heroes
status: draft
type: bug
priority: high
created_at: 2026-11-09T20:58:46Z
updated_at: 2026-11-09T20:58:46Z
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
