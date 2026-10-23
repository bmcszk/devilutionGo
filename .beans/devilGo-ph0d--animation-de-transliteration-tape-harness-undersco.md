---
# devilGo-ph0d
title: 'Animation de-transliteration: tape harness + underscore members + C++-signature func (findings 1-3)'
status: in-progress
type: task
priority: normal
created_at: 2026-10-23T16:10:39Z
updated_at: 2026-10-23T16:11:37Z
parent: devilGo-gmtn
---

Scope (brainer audit findings 1-3, /tmp report + devilGo-zk5z audit trail):
1. animationinfo_test.go:10-60 — DELETE the step/tape harness (type kind, type step, F/T/R constructors, run interpreter); rewrite its scenarios as table-driven t.Run subtests calling SetNewAnimation + ProcessAnimation inline with named fields. TestPetrifiedDoesNotAdvance already shows the target pattern. Every existing scenario and assertion strength must survive.
2. animationinfo.go:44-47 — rename trailing-underscore members (relevantFramesForDistributing_, skippedFramesFromPreviousAnim_, tickModifier_, ticksSinceSequenceStarted_) to plain lowerCamel. Update all references repo-wide.
3. animationinfo.go:99-103 — replace the 7-positional-param signature (ignored _ *ProgressSink, dummy-0 previewShownGameTickFragments) with a named options/Config struct or split the two real call shapes. NOT a nolint-cited 1:1-port function — verify before touching.

## Acceptance criteria
- [ ] No step/tape harness remains in animationinfo_test.go; scenarios count >= prior; go test green
- [ ] grep '_' members in animation package returns 0 hits
- [ ] No caller passes dummy/ignored args; signature reads in domain words
- [ ] Gates: gofmt/vet/golangci-lint 0/test ok in internal/diablo/animation

## Libraries
- none new

## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.hermes/skills/testing/port-parity-testing/SKILL.md — read fully BEFORE acting (idiom boundary governs)


## Progress log
- 2026-10-23 claimed: pi dispatch starting (animation cluster, findings 1-3).
