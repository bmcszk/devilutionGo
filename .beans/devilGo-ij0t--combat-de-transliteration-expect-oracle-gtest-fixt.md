---
# devilGo-ij0t
title: 'Combat de-transliteration: expect() oracle + gtest fixture shape + fuzz loops (findings 4,9)'
status: todo
type: task
priority: normal
created_at: 2026-10-23T16:10:39Z
updated_at: 2026-10-23T16:11:37Z
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
