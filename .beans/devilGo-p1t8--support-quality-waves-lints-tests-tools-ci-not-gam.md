---
# devilGo-p1t8
title: 'Support: quality waves, lints, tests, tools, CI — NOT game features'
status: todo
type: epic
created_at: 2026-10-29T16:04:42Z
updated_at: 2026-10-29T16:04:42Z
---

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
## Triage

- Brainer beans-audit of the ledger ran 2026-10-28 (devilGo-5wsb-beansaudit session; PR #3 verdict log). Per-bean brainer pass owed before implementation dispatch; wave placement, blockers and skills list set during the 2026-10-29 reorg (d083780).
## Acceptance criteria

- [ ] No feature bean dispatch blocked on missing support infra at any milestone gate
- [ ] Wave simulation (topo scan) clean across all 8 waves
- [ ] Support tasks closed with PoW before their dependents enter review
