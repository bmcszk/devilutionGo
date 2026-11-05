---
# devilGo-49ph
title: 'refresh-gamedata-goldens is not idempotent: committed goldens are scaffold-era row slices'
status: draft
type: bug
priority: normal
created_at: 2026-11-05T23:21:53Z
updated_at: 2026-11-05T23:22:02Z
---

## Finding (lane dg-binaries s1, during the ADR-020 gamedatagoldens move)

`just refresh-gamedata-goldens` does not reproduce the committed fixtures.

Evidence, collected without touching the repo (`git archive HEAD` export to
`/tmp`, moved generator copied in, `go run ./tools/gamedatagoldens`):

- generator output stamps `probeCommit: refresh-recipe-v1` and emits the full
  table (spelldat 36 rows/15 cols, misdat 68/8, monstdat 112/41, objdat 109/12,
  textdat 338/4, questdat 24/9, towners 12/11);
- committed `internal/testdata/gamedata/goldens/*.json` stamp
  `probeCommit: unit-fixture-slice-a` and hold a row slice (spelldat-base.json
  is 631 lines vs the generated full table), last touched in d6ef2dc (scaffold).

So the first clean-tree run of the recipe rewrites all seven goldens, breaking
the recipe's own contract (print a clean diffstat) and the bean AC "git diff on
the goldens is empty after one run".

## Why it was not fixed in that lane

The lane's scope guard forbids regenerating any committed golden (a move, not a
refresh), and the consumers (`internal/diablo/gamedata/audit`) currently pass
against the slices — so the decision to regenerate (and re-pin the row counts)
belongs to the parent, not to the move.

## Triage needed

Decide: regenerate the seven goldens from the committed TSVs (and update the
audit row-count expectations/manifest), or keep the slices and make the
generator emit them (e.g. a target row limit). Then re-run
`just refresh-gamedata-goldens` on a clean tree and require an empty diff.
