---
# devilGo-x9sv
title: 'Ledger honesty: n89r/b782 claims, gate boxes, l5fp ownership'
status: todo
type: task
priority: high
created_at: 2026-10-28T16:35:33Z
updated_at: 2026-10-28T16:35:33Z
parent: devilGo-5wsb
---

## Goal

Make the ledger tell the truth for `devilGo-n89r` and `devilGo-b782`: tighten the inflated
claims to what the tree actually pins, re-open the criteria that are not met, and record the
residuals in beans that exist rather than beans that do not.

## Evidence (verified in the parent)

- `devilGo-n89r` AC 1 reads `[x] Golden decode tests for ≥20 sampled frames **per format**`, but
  only CLX has committed goldens. CEL (`cel_test.go`), CL2 (`cl2_test.go`) and PCX
  (`cel_test.go:35`) are pinned by synthetic streams produced by helpers that mirror the
  decoder's own comments — the `0xBF` bug (filed separately) is invisible to them.
- `devilGo-n89r` PoW claims the Python re-derivation used "semantics from clx_decode.hpp:13-38",
  but `internal/testdata/gfx/README.md` documents `0x80..0xBF` fill / `>=0xC0` literal, which
  contradicts `clx_decode.hpp:25` (`ClxFillMax = 0xBE`). The "independent" check shares the
  decoder's wrong spec at that boundary.
- `devilGo-b782` AC 1 claims "asserted with seeded RNG vectors"; the tests
  (`TestLoadRndLvlPalDeterministicPerSeed`) inject counting stubs and assert draw counts, the
  `(1,4)` args, and the rv→path mapping. The behaviour is pinned; the wording is inflated.
- Both beans are `status: completed` while their binding TDD gate boxes
  ("failing test written FIRST", "1:1 parity evidence") are still `- [ ]`.
- `devilGo-l5fp` has **no** benchmark/performance acceptance criterion (verified: no
  `bench|perf|100ms` match) and **no** golden-decode criterion for CEL/CL2/PCX, so the plrgfx
  `<100ms` deferral and the CEL/CL2/PCX golden deferral both land in a bean with no room for them.

## Acceptance criteria

- [ ] `devilGo-n89r` AC 1 is re-worded to what is true (CLX ≥20 goldens MET; CEL/CL2/PCX pinned
      by synthetic tests only) and either unchecked or split into a CLX-bullet that is checked.
- [ ] `devilGo-n89r` PoW no longer implies the Python check validated the `0xBF` boundary;
      `internal/testdata/gfx/README.md` documents the correct CL2 ranges
      (`0x80..0xBE` fill, `0xBF`/`>=0xC0` literal, widths 63/65/1-64).
- [ ] `devilGo-b782` AC 1 wording matches the stub-based tests (no "seeded RNG vectors" claim
      unless a seeded vector is actually used).
- [ ] Both beans' TDD/parity gate boxes are either ticked with evidence (test name + captured
      RED output) or left unchecked with the reason stated — no `completed` bean carrying
      unexplained unchecked binding gates.
- [ ] `devilGo-l5fp` gains an explicit benchmark AC for plrgfx/CEL decode performance and an
      explicit CEL/CL2/PCX golden-decode AC, so the deferrals in PR #3 have a real owner.
- [ ] `python3 scripts/beans-lint.py .` passes.

Refs: PR #3, devilGo-n89r, devilGo-b782, devilGo-l5fp

## Libraries
- none (`.beans/` markdown only)

## Skills for coder
- `~/.agents/skills/go-dev/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules/SKILL.md` — read fully BEFORE acting
- `~/.agents/skills/coding-rules-notes/SKILL.md` — user amendments, read with coding-rules
- `~/.agents/skills/ponytail/SKILL.md` — read fully BEFORE acting

## Triage
Delegated read-only brainer audit (pi session `devilGo-5wsb-beansaudit`, report `/home/blaze/.hermes/cache/scratch/pr3-beans-audit.md`); every cited anchor re-verified in the parent against the pinned devilutionX checkout (`ca91d778f`) before filing.
