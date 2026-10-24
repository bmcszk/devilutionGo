---
# devilGo-dufb
title: Entity structs + containers (r462 residual)
status: todo
type: task
priority: normal
tags:
    - audit-gap
created_at: 2026-10-24T17:25:41Z
updated_at: 2026-10-24T17:26:11Z
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
