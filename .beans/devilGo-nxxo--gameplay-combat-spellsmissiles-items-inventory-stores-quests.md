---
# devilGo-nxxo
title: 'Gameplay: combat, spells/missiles, items, inventory, stores, quests'
status: todo
type: feature
priority: normal
created_at: 2026-10-29T00:28:56Z
updated_at: 2026-11-13T00:33:47Z
parent: devilGo-i56j
---

---
# devilGo-nxxo
title: 'Gameplay: combat, spells/missiles, items, inventory, stores, quests'
status: todo
type: feature
priority: normal
created_at: 2026-10-29T00:28:56Z
updated_at: 2026-11-13T00:33:34Z
parent: devilGo-i56j
---

---
# devilGo-nxxo
title: 'Gameplay: combat, spells/missiles, items, inventory, stores, quests'
status: todo
type: feature
priority: normal
created_at: 2026-10-29T00:28:56Z
updated_at: 2026-11-03T12:52:00Z
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
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md

## Progress log
- 2026-11-02: WAVE-3 GROUPED PR OPEN: https://github.com/bmcszk/devilutionGo/pull/11 (branch dg-persist-sv-sqlite-wave, 28 persistence commits, tip 0d955a9). 6xu2 CLOSED - all 6 ACs owner-verified (coder1 19 + coder2 9 commits; coder2 extra: sqlc store, migration chain v1->v2, batch importer, stash routing, demo compare). 37 pkgs ok, lint 0, sqlc-check clean. Fixed a local-vendor-dir trap in the main checkout (ignored vendor/modules.txt forced inconsistent-vendoring; rm -rf vendor, not a repo issue). Brainer dispatched (proc_2601a1434303, verdict -> /tmp/pr11-verdict.md).


2026-11-03 ADR alignment: gameplay systems stop mutating directly and become
kernel handlers; each owns a slice, wider change is requested through
commands, and events announce what changed. The missile registry sink seams
become the first handler surface. First handlers child: [Gate plus first gameplay handlers as commands](devilGo-noii--gate-plus-first-gameplay-handlers-as-commands.md);;
parent kernel: [Deterministic command/event simulation kernel (ADR-015, ADR-018)](devilGo-xc37--deterministic-commandevent-simulation-kernel-adr-0.md);

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Event sourcing (ADR-018): state changes this bean introduces go through
   kernel Commands with Gate rules and emitted Events — never direct mutations
   or a private callback bus. New Kind names live in the kernel package
   (law 5), defined per-domain WITH this wiring (no wholesale definition wave,
   no side-channel kinds). Presentation reads events only; it never writes sim
   state.
2. Layering (ADR-008): this bean's package imports only from its own layer and
   below (engine < game < diablo). Check the import direction BEFORE writing
   code; a generic mechanism never lands under internal/diablo, and diablo
   semantics (pcurs ordinals, PLR_MODE, panel flags) never leak into
   internal/game or internal/engine.
