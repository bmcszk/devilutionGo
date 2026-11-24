---
# devilGo-bh4m
title: Typed classdat loaders + MP award-cap (pu3r residual)
status: completed
type: task
priority: normal
tags:
    - audit-gap
created_at: 2026-10-24T17:25:41Z
updated_at: 2026-11-20T13:11:39Z
parent: devilGo-97mn
blocked_by:
    - devilGo-mwlm
---

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

- [x] go test ./internal/diablo/txtdata/... green with new typed-loader
      tests + award-cap table test (>= 10 new cases).
- [x] Each loader validates the full upstream column set (27/9/27/19 names — errata per review finding 1: sprites.tsv carries 19 columns at the pin playerdat.cpp:225-253, not the 13 the pu3r-era research guessed; implementation follows the pin)
      and rejects a table missing any of them with a named-column error.
- [x] Award cap matches mana.cpp for the upstream worked example; lint 0,
      vet clean, beans-lint OK.

## Skills for coder
- ~/.agents/skills/ponytail/SKILL.md

- /home/blaze/.agents/skills/coding-rules/SKILL.md
- /home/blaze/.agents/skills/go-dev/SKILL.md
- /home/blaze/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md

## Libraries

Go stdlib only — no new module dependencies to bind (repo policy; existing go.mod surface unchanged).
## Triage

- Brainer audit 2026-10-24 (/tmp/devilgo-sweep/beans-audit.md, devilGo-pu3r entry) — this bean exists as the direct residual of that audit's findings; no further brainer pass owed.

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Program-surface law (ADR-020/021 + one-tool law): no new binaries. Fixture
   instruments are subcommands of tools/goldens; user-facing surface stays
   exactly cmd/devilutionGo + cmd/devilutionGo-migrate. No `func main` outside
   the allowlist.
2. Tooling problems: fix immediately, no hacks.



## Progress log

- 2026-11-20T14:05Z: claimed — s6a dispatched under contract harness.s6a.md (worktree ftbu-s6, feat/m1-interactive-demo), status flip.


## Proof of Work

- Contract: .hermes/lane/ftbu-m1/harness.s6a.md (require_edit harness.s6.log, sentinel M1-REAL-L1-S6A-DONE). Worker: task-dg-m1-s6a-coder-coder (171 tool calls), feat/m1-interactive-demo @ 1fa79b6 — pushed; parent verified ls-remote == HEAD; street resolved (street/candidate/heads clean; trailer pi-session-2026-12-01-s6a on the commit).
- RED: captured unchanged-tree compile/test failures in the lane log (txtdata typed loaders + award-cap boundary "level 0 has a zero cap"); green only from the fix commit.
- Parent operator re-run at 1fa79b6: internal/diablo 26 pkgs ok, integration suite ok, single-walk trail f1ae95a0… unchanged; ls-remote == HEAD.
- Review: brainer PASS (approve) 8/8 (.hermes/lane/ftbu-m1/harness.s6a.review.md). Write-backs landed with this PoW: sprites.tsv column errata 13→19 (pin playerdat.cpp:225-253); Go-int typing deviation recorded (int8/16 range checks + fraction-carry guard not ported — unreachable with committed data, disclosed); lane-log cite drift (LoadClassDatFromFile :345→:325, LoadClassesAttributes :378→:359 at the pin) noted for the next log append; escape-style cosmetic named.
- Residual: sounds.tsv loader intentionally out of scope (SfxID catalog lane); nothing blocking.
