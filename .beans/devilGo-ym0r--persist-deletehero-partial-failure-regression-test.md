---
# devilGo-ym0r
title: 'persist: DeleteHero partial-failure regression test'
status: todo
type: task
priority: normal
created_at: 2026-11-02T14:25:47Z
updated_at: 2026-11-02T21:56:44Z
---

---
# devilGo-w3rm
title: 'persist: DeleteHero partial-failure regression test'
status: draft
type: task
created_at: 2026-11-02T00:00:00Z
updated_at: 2026-11-02T00:00:00Z
parent: devilGo-6xu2
---

# persist: DeleteHero partial-failure regression test (PR14 brainer L1)

Deleting the BeginTx wrapper around DeleteHero's four sequential deletes
(store_sqlite.go:152-173) keeps every existing test green - atomicity is
pinned only by reading the code. Add a test that injects a failing delete
(e.g. drop a dependent table mid-test or cancel ctx between deletes) and
asserts hero + all dependent rows survive. PR14 brainer finding 1 (Low).
