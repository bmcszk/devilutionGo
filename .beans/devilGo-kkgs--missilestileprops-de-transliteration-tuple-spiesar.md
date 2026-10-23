---
# devilGo-kkgs
title: 'Missiles+tileprops de-transliteration: tuple spies/args (findings 5-8)'
status: todo
type: task
priority: normal
created_at: 2026-10-23T16:10:39Z
updated_at: 2026-10-23T16:11:37Z
parent: devilGo-gmtn
---

Scope (brainer audit findings 5-8):
1. missiles/collision_test.go:63-78,147-172 — rollArgs mirrors the C++ arg pack; put named want fields (toHit, minD, maxD, ...) on the case struct.
2. game/world/tileprops_test.go:180,199,226-253 — opaque [4]int tuples + gtest-fixture-named tests; use struct{from, to Point}-style named cases; keep upstream citations.
3. missiles/process_ii_test.go:28-29 and process_test.go:58-59 — [][4]int call-record spies; record named call structs (method + fields). Multi-tick tape rounds in process_test keep their when/then-per-tick labels (kc11 rule).

## Acceptance criteria
- [ ] No [4]int arg packs or [][4]int spies remain in the named files (grep 0)
- [ ] Case structs read in domain words; citations preserved
- [ ] All prior scenarios still covered; go test green in touched packages
- [ ] Gates: gofmt/vet/golangci-lint 0/test ok

## Libraries
- none new

## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.hermes/skills/testing/port-parity-testing/SKILL.md — read fully BEFORE acting (idiom boundary governs)


## Triage
Brainer audit devilGo-cppaudit-1 findings 5-8 assigned here; parent grepped the tuple patterns live before flipping to todo. Named-struct replacements must preserve every recorded call and citation.
