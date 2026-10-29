---
# devilGo-bh4m
title: Typed classdat loaders + MP award-cap (pu3r residual)
status: todo
type: task
priority: normal
tags:
    - audit-gap
created_at: 2026-10-24T17:25:41Z
updated_at: 2026-10-29T15:41:46Z
parent: devilGo-97mn
blocked_by:
    - devilGo-mwlm
---

# Typed classdat loaders + MP award-cap

## Goal

Close the residual gaps the 2026-10-24 brainer beans-audit found in
devilGo-pu3r: the class domain ships only generic 2-column readers; the
typed per-table loaders and the MP award-cap formula are absent.

## Research

- Brainer audit 2026-10-24 (/tmp/devilgo-sweep/beans-audit.md, devilGo-pu3r
  entry): 3 AC bullets of devilGo-pu3r have no artifact in tree.
- Upstream: tables/playerdat.cpp:89-127 (attributes 27 vars x 6 classes),
  :171 (attribute column sets, md5-verified in pu3r research), :209
  (starting_loadout 9 vars x 6), :251 (animations 27 x 6), :227 (sprites
  13 x 6), :345 (classdat loader); mana.cpp award cap min(maxmana/20, 200
  x level) — verify exact terms online in the devilutionX source before
  porting (always verify online rule).
- Existing tree surface: internal/diablo/txtdata/classes.go (ClassData,
  LoadClasses, LoadClassTables), mana.go (ManaInput, SpellCosts,
  GetManaAmount).

## Scope

- internal/diablo/txtdata/classes.go (+ siblings as needed): typed loaders
  ClassAttributes (27 vars x 6), ClassLoadout (9 x 6), ClassAnimations
  (27 x 6), ClassSprites (13 x 6), each with column-set validation against
  the upstream names; errors name the offending column.
- MP award-cap: AddMana/award-side function capping at min(next/20,
  200*level) with the 1/64 fixed-point semantics matching GetManaAmount.
- Tests: per-loader happy path + short-row + unknown-column error cases;
  award-cap table test over level 1..max with upstream worked examples.
- Do NOT touch: .golangci.yml, existing GetManaAmount behavior, load-order.

## Acceptance criteria

- [ ] go test ./internal/diablo/txtdata/... green with new typed-loader
      tests + award-cap table test (>= 10 new cases).
- [ ] Each loader validates the full upstream column set (27/9/27/13 names)
      and rejects a table missing any of them with a named-column error.
- [ ] Award cap matches mana.cpp for the upstream worked example; lint 0,
      vet clean, beans-lint OK.

## Skills for coder

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/golang-testing/SKILL.md
- /home/blaze/.hermes/skills/go-dev-notes/SKILL.md
