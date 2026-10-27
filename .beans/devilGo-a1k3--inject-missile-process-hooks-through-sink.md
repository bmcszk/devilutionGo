---
# devilGo-a1k3
title: Inject missile process hooks through sink
status: completed
type: task
created_at: 2026-10-21T01:22:52Z
updated_at: 2026-10-21T01:22:52Z
---

Replace package-level mutable hook vars in internal/diablo/missiles/process.go with an injected ProcessCallbacks sink (HitCallbacks pattern), carried on HitWorld.

## Proof of Work
- All hook vars (MissilePreFlag, HookChangeLightOffset/ChangeLight/AddUnLight/AddLight, HookSpawnMissile, HookSourceDamage, HookElementDamage, HookIsTileAvailable, HookGenerateRnd, HookTransAt, HookCheckBlock, Currlevel) replaced by ProcessCallbacks struct + NoopProcessCallbacks(); carried as HitWorld.ProcessCallbacks.
- Signatures updated: PutMissile(w,m), UpdateMissilePos(w,m), MoveMissilePos(w,m), MoveMissile(w,m,...); elemental helpers take w. Call sites: registry.go (no-op sink), missile.go comment, process_test.go + process_ii_test.go (spies wire w.ProcessCallbacks; no global save/restore).
- Gates: gofmt -l internal/diablo/missiles empty; go build ./... ok; go vet ./internal/diablo/missiles/... ok; golangci-lint run ./internal/diablo/missiles/... 0 issues; go test ./internal/diablo/missiles/... ok.
- rg '^var (MissilePreFlag|Currlevel|Hook)' process.go → no matches.
Residual: none.
