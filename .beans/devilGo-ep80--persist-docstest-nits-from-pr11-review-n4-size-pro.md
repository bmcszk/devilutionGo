---
# devilGo-ep80
title: 'persist: docs+test nits from PR11 review (N4 size proof enumeration, N5 requirements-doc schema decision)'
status: completed
type: task
priority: normal
created_at: 2026-11-01T22:21:12Z
updated_at: 2026-11-02T18:54:05Z
---

N4: enumerate PlayerNetPack member arithmetic in netpack_test.go instead of the self-referential 1600 constant. N5: record the games/levels/stash key-on-hero_slot (no FK) + settings-stays-ini decision in docs/research/dg-requirements.md section 5 (or an ADR note) so later sessions don't treat the gap as an error. PR11 brainer nits.

## Proof of Work

- netpack_test enumerates pack.h member arithmetic (1692/26); requirements doc records hero_slot-key/no-FK + INI-settings decision.
