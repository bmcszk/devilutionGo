---
# devilGo-qr5a
title: Golden-fixture refresh recipe (yv1r residual)
status: todo
type: task
priority: normal
tags:
    - audit-gap
created_at: 2026-10-24T17:25:41Z
updated_at: 2026-10-24T17:26:11Z
parent: devilGo-97mn
---

# Golden-fixture refresh recipe (yv1r residual)

## Goal

Close the defensible-but-unowned residual the 2026-10-24 audit flagged on
devilGo-yv1r: nothing in the repo currently regenerates golden fixtures or
extends per-table coverage; the C++ dataprobe sidecar and `just` recipe
from yv1r's AC never landed (zero-dependency policy deferral).

## Research

- Brainer audit 2026-10-24 (/tmp/devilgo-sweep/beans-audit.md, yv1r
  entry): Go comparator + one-row unit goldens exist; full-table
  per-variant goldens and the refresh recipe do not.
- Zero-dependency test policy (devilGo-i56j, binding 2026-10): Go test
  chain must run without devilutionX checkout/C++/MPQ/network. So: the
  REFRESH path may require the C++ probe (offline producer workflow),
  but the VERIFY path (CI, go test) must not.
- Existing: internal/diablo/gamedata/* golden unit fixtures; scripts/
  beans-lint.py precedent for repo scripts.

## Scope

- `just refresh-gamedata-goldens` recipe that (a) runs the committed
  producer (Go, reading committed raw TSV inputs — NOT a C++ build) to
  regenerate golden JSON, (b) prints a diffstat, (c) refuses to run if
  the tree is dirty.
- Extend goldens from one-row samples to full-table coverage for the
  highest-traffic tables (classes, spells, experience, misdat, objdat,
  textdat) — generated output committed as fixtures.
- Do NOT touch: comparator logic semantics, .golangci.yml, CI workflow
  structure.

## Acceptance criteria

- [ ] `just refresh-gamedata-goldens` runs clean on a dirty-tree
      rejection probe and regenerates byte-identical output on a clean
      tree (idempotence check in PoW).
- [ ] Full-table goldens for >= 6 tables committed; parity tests consume
      them and stay green WITHOUT any C++/MPQ/network dependency.
- [ ] lint 0, gofmt clean, beans-lint OK, CI success.

## Libraries

- No new dependencies. Standard library + existing repo modules only
  (comments/JSON fixtures are data, not dependencies). Zero-dependency
  test policy (devilGo-i56j, binding 2026-10) applies.

## Skills for coder

- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/golang-testing/SKILL.md
- /home/blaze/.hermes/skills/go-dev-notes/SKILL.md
- /home/blaze/.agents/skills/ponytail/SKILL.md


## Triage

Delegated triage evidence: brainer beans-completion audit 2026-10-24
(/tmp/devilgo-sweep/beans-audit.md, session devilGo-lu2v-brainer) verified
the gap against the tree at HEAD c9fb7e8; parent spot-checked the cited
file:line evidence before creating this bean. Scoped per audit
recommendation; acceptance criteria are checkable as written.
