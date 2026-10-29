---
# devilGo-4xin
title: 'Corpses: dead.cpp port (Corpse records, dCorpse occupancy, corpse render)'
status: todo
type: task
priority: normal
created_at: 2026-10-29T15:44:55Z
updated_at: 2026-10-29T16:14:42Z
parent: devilGo-97mn
---

## Goal

Port `dead.cpp` (108 lines): Corpse placement — `Corpse` records (sprites per
monster/dir), `dCorpse` occupancy encoding `idx | dir<<5`
(levels/dun_tile_data.hpp:110-116), corpse render call in
engine/render/scrollrt.cpp:798-806.

## Research

Source/dead.cpp:1-108 (InitDeadAnimationFromMonster, AddCorpse, dCorpse
get/set); docs/research/dg-requirements.md §3 Entities "Dead/corpse" (line 224):
on monster death a Corpse record + dCorpse occupancy; players drop too; saves
serialize corpses (loadsave.cpp) — coordinate serializer fields with oqho.

## Scope

internal/diablo/entities/corpse.go (record + occupancy map) +
internal/diablo/entities (death hook from monster death path) + render hookup
point left for qbiy renderer (draw call interface, not pixels).

## Acceptance criteria

- [ ] Corpse struct + dCorpse map ported with exact idx|dir<<5 encoding (round-trip test)
- [ ] InitDeadAnimationFromMonster ported (sprite share, no copy) vs C++ line-order
- [ ] AddCorpse occupancy rules (max corpses, reuse policy) parity-tested vs C++ fixture
- [ ] Save/serialize fields for corpses agreed with oqho (struct tags or documented layout)
- [ ] No import-layer violations (entities may not import game/renderer)

## Skills for coder

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md

## Dependencies & blockers

None hard: entities structs (r462/dufb) landed. Render hookup waits for qbiy
iso renderer but is a one-line interface call.

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
## Triage

- Brainer beans-audit of the ledger ran 2026-10-28 (devilGo-5wsb-beansaudit session; PR #3 verdict log). Per-bean brainer pass owed before implementation dispatch; wave placement, blockers and skills list set during the 2026-10-29 reorg (d083780).
