---
# devilGo-a1n6
title: random-test-scope-trim-ponytail
status: draft
type: task
priority: normal
created_at: 2026-10-17T11:31:53Z
updated_at: 2026-10-17T11:32:07Z
parent: devilGo-qna9
---

# random_test.go scope trim (ponytail / 1:1-functionality rule)

## Goal
Apply the 2026-10-17 scope rule (milestone devilGo-i56j ## DECIDED) to internal/engine/random_test.go: 356 lines / 17 tests currently pin C++ scaffolding (double-jump statistics, reproducibility reruns) alongside genuine parity gates.

## Scope
- KEEP: golden sequence (SetRndSeed(2588) 1000 raw outputs), GenerateRnd bias/shift/mod range gates, AdvanceRndSeed INT_MIN, FlipCoin, xoshiro golden stream + Jump-changes-stream (Jump feeds GenerateSeed on the SP path)
- CUT or MERGE: statistical/reproducibility duplicates that re-prove the stdlib-free arithmetic already covered by goldens
- Target: parity evidence preserved, file <= ~200 lines

## Acceptance criteria
- [ ] go test ./internal/engine/ green after trim; every deleted test's premise either duplicated by a kept golden or game-uncalled (cite DECIDED rule)
- [ ] wc -l internal/engine/random_test.go <= 200
- [ ] No narration comments introduced (coding-rules)

## Skills for coder
/home/blaze/.agents/skills/coding-rules/SKILL.md
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md

## Parent
devilGo-qna9
