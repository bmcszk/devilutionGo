---
# devilGo-l9lj
title: 'persist: DeleteHero transaction + migration atomicity nits (PR11 N1+N2)'
status: completed
type: task
priority: normal
created_at: 2026-11-01T22:21:12Z
updated_at: 2026-11-02T18:54:05Z
---

Wrap DeleteHero 4 sequential deletes (store_sqlite.go:152-173) in db.BeginTx; make applyMigration DDL+version stamp atomic or pin the IF-NOT-EXISTS invariant with a comment (migrate.go:51-64). PR11 brainer nits N1+N2, robustness only, no parity impact.
