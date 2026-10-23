---
# devilGo-ij0t
title: 'Combat de-transliteration: expect() oracle + gtest fixture shape + fuzz loops (findings 4,9)'
status: completed
type: task
priority: normal
created_at: 2026-10-23T16:10:39Z
updated_at: 2026-10-23T16:45:06Z
parent: devilGo-gmtn
---

Scope (brainer audit findings 4,9):
1. combat_test.go:14,69,119 — the fixture type + runCase + expect() oracle re-derive the upstream player.cpp arithmetic; the test must PIN literal expected values per table case (computed once, offline, cited to the upstream line) instead of re-implementing the body under test. Keep the RNG tape seam (tap) — that is parity infrastructure, not the problem. All existing roll scenarios survive with identical coverage.
2. combat_fuzz_test.go:16-20 — triple-nested C++-style param loops under rand.New are neither table tests nor real fuzzing; convert to table-driven t.Run subtests over the same coverage space (or a genuine F.Fuzz target if the vars are fuzz-compatible — state which and why).

## Acceptance criteria
- [ ] No expect() oracle re-implementing upstream arithmetic; wants are literals traceable to citations
- [ ] Fuzz file is either a real F.Fuzz target or honest table subtests; no rand.New seeding
- [ ] All prior scenarios still covered; go test green in internal/diablo/combat
- [ ] Gates: gofmt/vet/golangci-lint 0/test ok

## Libraries
- none new

## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.hermes/skills/testing/port-parity-testing/SKILL.md — read fully BEFORE acting (idiom boundary governs)

## Triage
Brainer audit devilGo-cppaudit-1 findings 4+9 assigned to this task; parent read combat_test.go fixture/expect() + combat_fuzz_test.go loops against the report before flipping to todo. Conversion constraints: literals must stay traceable to upstream citations; tape seam stays.
EOF
## Triage
Brainer audit devilGo-cppaudit-1 findings 5-8 assigned here; parent grepped the tuple patterns live ([4]int arg packs, [][4]int spies) before flipping to todo. Named-struct replacements must preserve every recorded call and citation.


## Progress log
- 2026-10-23 claimed: pi dispatch starting (combat cluster, findings 4+9).


## Progress log
- 2026-10-23 three pi sessions (ij0t x3 resumes stalled on narration; ij0t-f1 fuzz file; ij0t-t1/t2 test file): table cases pin literal wantDam/wantHit per player.cpp-cited derivation, expect() kept ONLY as differential oracle for the randomized harness, fuzz triple loop -> deterministic table over same 1000-seed space, expectHper/weaponMod extracted for revive (complexity 7, defaults added; interim double-RNG-consumption caught by the differential harness and fixed). Parent gates: gofmt 0, vet rc0, golangci-lint 0, go test 9006 passed, build rc0. Commit cdefa65.

## Proof of Work
- gofmt -l internal/diablo/combat -> empty; go vet -> rc0; golangci-lint run ./internal/diablo/combat/... -> 0 issues; go test -> 9006 passed; go build ./... -> rc0.
- Parent diff re-read: literals carry inline player.cpp citations (527-533, 542-558, 567, 574-578, 589-602, 603-606, 618, 626-627); RNG tape seam untouched (6 tap() refs); every prior scenario present (t.Run ladder over classes/mclasses/tapes).
Residual: none.
