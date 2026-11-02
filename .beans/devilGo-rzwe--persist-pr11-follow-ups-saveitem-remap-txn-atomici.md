---
# devilGo-rzwe
title: 'persist PR11 follow-ups: SaveItem remap + txn atomicity + docs nits'
status: completed
type: task
priority: normal
created_at: 2026-11-02T13:57:52Z
updated_at: 2026-11-02T18:54:05Z
parent: devilGo-6xu2
---

# persist PR11 follow-ups: SaveItem remap seam + txn/atomicity + docs/test nits

[migration target: later fetched via beans CLI — placeholders fine]

## Goal

Close the three PR #11 follow-up beans (devilGo-aw1c, devilGo-l9lj,
devilGo-ep80) as one grouped PR in internal/persist.

## Background

PR #11 (persist wave, squash-merged) left three review follow-ups, all
bean-tracked. SaveItem remap is the only parity-bearing one; the other two
are robustness + documentation.

## Libraries

- modernc.org/sqlite (ADR-012) — already in go.mod, no new deps
- stdlib only otherwise

## Triage

compact read-only brainer owed on the grouped PR before merge.

## Skills for coder

~/.agents/skills/coding-rules/SKILL.md
~/.agents/skills/go-dev/SKILL.md
~/.agents/skills/ponytail/SKILL.md

## Acceptance criteria

Per member bean:
- [ ] aw1c: SaveItem.Encode applies iMappingId + RemapItemIdx per mode,
      -1 -> IDI_GOLD with iType=None; golden encode vector; seam doc.
- [ ] l9lj: DeleteHero wrapped in BeginTx; applyMigration DDL+version-stamp
      atomic or invariant pinned with comment.
- [ ] ep80: netpack_test enumerates PlayerNetPack arithmetic instead of the
      1600 constant; requirements doc records key-on-hero_slot (no FK) +
      settings-stays-ini decision.
- [x] Gates: go build / go vet / go test ./... / gofmt / just lint green;
      CI green on the PR (runs 41459467973 + 41461935162, tips 6204eae/7ad0643).

## Progress log

2026-11-02 coder wave done (372cd6a/fdfab62/6204eae), brainer SHIP (re-derived remap order + golden vectors + 1692 pack arithmetic); nit comment + ym0r follow-up bean pushed 7ad0643; CI green. PR #14 ready to merge - SHIP.
2026-11-02 grouped lane opened post-PR #13 (master 282c962); members aw1c+l9lj+ep80, branch dg-persist-saveitem-remap-txn-docs-nits.

2026-11-02 GWT audit (user-prompted): SizeGolden smuggled a SaveItemSize act in then - split into TestSaveItemSizeHelperMatchesEncode; ca56078, CI green 41464301556. Merge-ready.

## Proof of Work

- Remap parity: brainer re-derived remap order (loadsave.cpp:1164-1176), -1 passthrough through both remap tables, iType write at offset 8 (ItemType::None=-1), IDI_GOLD=0; golden vectors recomputed by hand (200->87->83; 201->83->GOLD/None; spawn 105->100->88).
- txn: DeleteHero four deletes under one BeginTx (store_sqlite.go:152-173); migration IF-NOT-EXISTS replay invariant pinned (migrate.go:51-64).
- Sizes: PlayerNetPack 1692 / ItemNetPack 26 re-derived from pack.h:90-146 member arithmetic; drift breaks the equation.
- Gates at 6204eae/7ad0643/ca56078: build+vet+test rc0, lint 0 issues; CI runs 41459467973, 41461935162, 41464301556 all success.
