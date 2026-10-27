---
# devilGo-dufb
title: Entity structs + containers (r462 residual)
status: completed
type: task
priority: normal
tags:
    - audit-gap
created_at: 2026-10-24T17:25:41Z
updated_at: 2026-10-24T18:52:20Z
parent: devilGo-97mn
---

# Entity structs + containers (r462 residual)

## Goal

Close the residual the 2026-10-24 brainer beans-audit found in devilGo-r462:
the bean's headline AC — full entity structs and containers — was unshipped;
only AnimationInfo, Missile and MissileList slices landed.

## Research

- Brainer audit 2026-10-24 (/tmp/devilgo-sweep/beans-audit.md, devilGo-r462
  entry): headline scope absent, no named successor.
- Upstream: structs.h (PlayerStruct, MonsterStruct, ObjectStruct,
  ItemStruct, MissileStruct), monster.cpp monster helpers, missiles.cpp
  list semantics; control.cpp for item/gold overlaps where relevant.
- Existing tree: internal/diablo has Missile + MissileList (r462 shipped
  slice), internal/game/world grid; ECS container precedent: donburi is the
  approved library (devilGo-r462 ## Libraries).
-devilGo-mwlm (Missile engine + behavior registry, in-progress) consumes
  missile iteration order — coordinate field additions with that bean.

## Scope

- Entity structs (player, monster, object, item, missile) as donburi
  components with field parity to upstream structs.h (order documented,
  size-assertion table per the r462 AC pattern).
- Containers: donburi collections with std::list-equivalent iteration
  order for missiles (the 1000-random-add/del parity probe from the r462
  AC) and stable-index semantics for monsters/objects.
- Tests: iteration-order parity probe (1000 sequences vs reference
  fixture), struct size-assertions compile.
- Do NOT touch: missile behavior registry (devilGo-mwlm's surface),
  .golangci.yml, save/stream code.

## Acceptance criteria

- [ ] donburi components for the 5 entity kinds with documented upstream
      field mapping (structs.h line refs) compile with size assertions.
- [ ] Missile container reproduces std::list add/del order for the 1000-
      sequence probe fixture (test green).
- [ ] devilGo-mwlm surface unbroken: go build ./... green; lint 0.

## Skills for coder

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/golang-testing/SKILL.md
- /home/blaze/.agents/skills/game-developer/SKILL.md — ECS component design
- /home/blaze/.hermes/skills/go-dev-notes/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md


## Triage

Delegated triage evidence: brainer beans-completion audit 2026-10-24
(/tmp/devilgo-sweep/beans-audit.md, session devilGo-lu2v-brainer) verified
the gap against the tree at HEAD c9fb7e8; parent spot-checked the cited
file:line evidence before creating this bean. Scoped per audit
recommendation; acceptance criteria are checkable as written.



## Proof of Work

Commands + results (2026-10-24, HEAD 37ea750):
- go build ./... → rc0
- go vet ./... → rc0
- gofmt -l internal/ cmd/ → empty
- golangci-lint run → 0 issues (config untouched, diff vs f34643d empty)
- go test -count=1 ./... → 9705 passed in 31 packages, 0 FAIL
- Probe TestCollectionRandomSequencesVsListModel: deterministic at seed
  20261024 (fixed-seed repeat run identical)
- CI: success at 37ea750d68963a8a54c5ce19c520e148769f0807

Commits:
- 23edc2f entities package: Player/Monster/Object/Item donburi components
  + Missile wrapper, upstream citations (player.h:211, monster.h:213,
  objects.h:32, items.h:189 — checkout has no structs.h), size pins
  (19712/224/144/336/80/16)
- b0ae187 revert of parent's inline test commit (USER RULING 2026-10-24:
  ALL-CODE-DELEGATED — code changes only via pi workers)
- 37ea750 donburi MissileCollection + std::list order probe (per-op
  RemoveMarked), collection.go + collection_test.go

Report: /tmp/devilgo-sweep/report-dufb.md
Residual: none — all AC met; entities surface ready for devilGo-mwlm and
quest/item consumers.

Claimed 2026-10-24: dispatched to pi coder worker devilGo-dufb-coder (omni/_worker, standard mode, contract /tmp/devilgo-sweep/contract-dufb.md, log dufb-coder.log).


Proof of Work (2026-10-24):
- Commit 23edc2f: entities package (Player/Monster/Object/Item donburi components + Missile wrapper, upstream citations from real files player.h/monster.h/objects.h/items.h/missiles.h, size pins 19712/224/144/336/80/16)
- Commit b0ae187: revert of parent's inline test commit per USER RULING (ALL-CODE-DELEGATED)
- Commit 37ea750: donburi-backed MissileCollection with std::list order probe (per-op RemoveMarked, seed 20261024)
Gates: go build rc0, go vet rc0, gofmt (internal/) clean, golangci-lint 0 issues, go test -count=1 ./... 9705 pass / 31 packages
Report: /tmp/devilgo-sweep/report-dufb.md
Residual: none — all AC met; entities package ready for devilGo-mwlm / quest consumers.
