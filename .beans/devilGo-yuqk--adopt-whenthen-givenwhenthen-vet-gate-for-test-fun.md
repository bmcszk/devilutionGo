---
# devilGo-yuqk
title: Adopt whenthen given/when/then vet gate for test functions
status: todo
type: task
priority: normal
created_at: 2026-10-26T18:03:13Z
updated_at: 2026-11-28T12:34:53Z
parent: devilGo-1n0a
---

## Goal
Adopt the whenthen given/when/then vet gate for test functions. RE-SCOPED 2026-11-27 (parity-close triage): the gate ALREADY runs on every 'just lint' via the vettool and ALREADY covers test functions - the historical 79 findings were fixed and the gate made blocking. The residual is documentation only: this bean closes as already-adopted with a fresh-run PoW.

## Acceptance criteria
- [ ] 'just lint' (which runs the whenthen vettool) exits 0; tail in PoW.
- [ ] Bean records the decision: the gate is in 'just ci' via 'just lint'; the earlier .so-plugin CI problem was resolved by the separate vettool step.
- [ ] AC text corrected (the old 'just lint-tests' command does not exist).
- [ ] Proof of Work with a fresh run (RED evidence = the historical 79-finding count in the body; GREEN = this run).
- [ ] No //nolint and no excluded files added for this purpose (wave diff clean of them).

## Libraries
None new - the gate and vettool are already wired; this bean changes documentation and records a PoW.

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Parity-close triage 2026-11-27 (task-pc-triage-brainer): PROMOTE-PARTIAL epic plan; this child = already-adopted, close with fresh PoW.
