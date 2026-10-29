---
# devilGo-nxxo
title: 'Gameplay: combat, spells/missiles, items, inventory, stores, quests'
status: todo
type: feature
priority: normal
created_at: 2026-10-29T00:28:56Z
updated_at: 2026-10-29T15:28:41Z
parent: devilGo-i56j
---

## Triage (2026-10-29, reorganization wave)

Type corrected task->feature (CLI type hierarchy: the epic slot is reserved
for THE GOAL devilGo-i56j; this bean is the gameplay use-case layer). Legacy
child-ID typos in the old body (devilsGo-*) corrected to devilGo-*. Child
beans carry their own Research/AC/Skills-for-coder triage — verified
2026-10-29: every open child has full triage sections.

## Scope

Combat resolution + missiles + items data + affix selection + item
generation + inventory/belt/gold + stores + quests FSM/XP. Byte-exact to
devilutionX; pi-delegated workers; TDD with golden fixtures.

## Acceptance criteria

- [ ] All 12 child beans completed with PoW (failing->passing gates + brainer audit)
- [ ] Item drop path E2E parity-tested: kill -> drop roll -> affix select -> ground item -> pickup -> inventory
- [ ] Quest state-machine transitions byte-match devilutionX questdat matrix
- [ ] Full test suite green in CI including data-completeness audit (devilGo-i2cr)

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
## Triage

- Brainer beans-audit of the ledger ran 2026-10-28 (devilGo-5wsb-beansaudit; PR #3 verdict log) — ledger-level brainer evidence; per-bean brainer pass owed before implementation dispatch. Wave placement, blockers and skills list set during the 2026-10-29 reorg (d083780).
## Skills for coder

- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
