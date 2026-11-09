---
# devilGo-bqcf
title: persist ExportStash spawn family roundtrip unpinned
status: draft
type: bug
priority: high
created_at: 2026-11-09T23:39:19Z
updated_at: 2026-11-09T23:39:19Z
parent: devilGo-6xu2
---

# persist ExportStash spawn-family roundtrip unpinned

## Symptom (imported-rows review F3, 2026-11-11)

`persist.ExportStash(path, mode, slot)` takes caller-supplied mode+slot while
import resolves family by NAME in one seam. Only the retail slot has a test
(importdir_test.go:142); nothing pins an export round-trip for the spawn
family (`ExportStash(..., GameMode{Spawn: true}, SpawnStashSlot)`).
Export re-encodes the blob with the family's own item size on import, so a
wrong-mode export cannot be caught today.

## Root-cause hypothesis (needs verification)

Asymmetric test gap, not a code defect: import's stashFamilyFor decides the
family; export truststhe caller. A wrong-mode/slot slip at a future call site
corrupts exports silently. Verify by reading stash.go's ItemSizeFor branch
(stash.go:52-61 - branches only on Hellfire; spawn shares Diablo size).

## Triage needed

Root cause not yet verified. Do not patch first: research the export call
sites, pin whether a wrong-mode export is reachable, write the AC, then
promote.
