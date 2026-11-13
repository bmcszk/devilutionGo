---
# devilGo-bqcf
title: persist ExportStash spawn family roundtrip unpinned
status: draft
type: bug
priority: high
created_at: 2026-11-09T23:39:19Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-6xu2
---

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

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Save story (ADR-003 + triage TD-4): SQLite rows (SaveHero/SaveGame/
   SaveLevel) are THE session save; the tick log is the replay/desync oracle
   validated against rows (upstream pfile.cpp:657/:665 model) — never a
   competing loader.
2. Data access (repo law): every DB read/write goes through sqlc-generated
   queries; raw SQL is a contract breach. New schema file => added to
   sqlc.yaml in the same commit; `just sqlc-check` in the gate list.
