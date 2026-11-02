---
# devilGo-rzwe
title: 'persist PR11 follow-ups: SaveItem remap + txn atomicity + docs nits'
status: draft
parent: devilGo-6xu2
type: task
created_at: 2026-11-02T13:57:52Z
updated_at: 2026-11-02T13:57:52Z
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
- [ ] Gates: go build / go vet / go test ./... / gofmt / just lint green;
      CI green on the PR.

## Progress log

2026-11-02 grouped lane opened post-PR #13 (master 282c962); members aw1c+l9lj+ep80, branch dg-persist-saveitem-remap-txn-docs-nits.
