---
# devilGo-kefy
title: 'Coding-rules + ponytail full pass round 2: every .go file one by one'
status: completed
type: epic
priority: normal
created_at: 2026-10-21T11:33:02Z
updated_at: 2026-10-23T13:03:03Z
---

# Coding-rules full pass round 2 — EVERY .go file, rules from 2026-10-21 revision

User directive (verbatim): "reload the 'coding-rules' skill, there were changes there. Apply ALL these rules to ALL the source files in the project. If this is overwhelming task, use beans tasks to split it into smaller ones."

## Triage
User binding quote above. coding-rules skill RELOADED 2026-10-21; governing revision points beyond the prior sweep's comment focus:
1. **Comments**: no comments by default; delete narration + upstream citations; doc comments restating name/signature are narration; **test phase markers (given/when/then, arrange/act/assert) are REQUIRED structure — never delete, add when missing** (explicitly outranks no-comments).
2. **Intent/mechanism separation**: top-level function = business steps in domain words; mechanisms make no business decisions; no vague utils/manager/process names.
3. **Visibility**: start every symbol at tightest visibility; widen only for real outside callers.
4. **File structure**: imports → constants → vars → structs → constructors → methods; **call order (caller above callee) ranks within category**; public before private only when call order doesn't decide; newer toward bottom.
5. **Duplication**: one-value-sibling getters collapse to chooser + one-liners; shared preambles hoist.
6. **Tests**: black box, public API only, no mocks (thin only for network/clock/randomness/hardware).
7. **Gates unchanged**: no nolint additions, no lint-config edits, never skip tests; five 1:1-port exempt functions keep citations; PORTED-EXEMPT files keep citation blocks.

Split into 4 countable batch tasks (parent-owned closure, per coding-rules "commit one file/batch per commit, gates green between"):
- devilGo-xc4a: batch 1 — cmd/ + internal/diablo/animation, doc.go, enums/ (files 1-26 of /tmp/dg-sweep2-files.txt)
- devilGo-xc4b: batch 2 — internal/diablo/gamedata, light, sound (files 27-52)
- devilGo-xc4c: batch 3 — internal/diablo rest + internal/game (files 53-79)
- devilGo-xc4d: batch 4 — remaining internal/ + root go files incl. all _test.go (files 80-104)

(Exact per-batch file lists live at /tmp/dg-sweep2-batch{1..4}.txt; batch sizes sum to 104 — audit command in PoW.)

## Acceptance criteria
- [ ] 4 batch tasks completed, each with per-file verdict audit: batch file count == verdict count (CLEAN or FIXED, exact marker per file)
- [ ] New-rule coverage: batch contracts explicitly include phase-marker preservation/addition in _test.go, tightest-visibility check, call-order layout check, duplication check (not comments-only)
- [ ] Every FIXED file re-gated green; parent full re-run (gofmt/build/vet/golangci-lint/go test ./...)
- [ ] Numeric audit in epic PoW: `git ls-files '*.go' | wc -l` == 104 == total unique verdicts across batches
- [ ] CI green after push; PR body re-synced with ready marker

## Libraries
- none new (style pass; no dependency additions allowed)

## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting on every file (2026-10-21 revision governs)
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting on every file (outranked by coding-rules on conflict)
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/golang-testing/SKILL.md — read fully BEFORE acting (test files: phase markers, no mocks, public API)

## Progress log
- 2026-10-21: reloaded changed coding-rules; split into 4 batch children; stale v2 worker stopped; prior partial edits (7 files) preserved in worktree as batch-1 starting state.

 (completed: all gates green)


## Proof of Work
- Parent epic devilGo-kefy closed via its 4 batch children, each with per-file verdicts: devilGo-iw3r (batch 1/4), devilGo-9wpf (2/4), devilGo-ej38 (3/4), devilGo-ar2y (4/4) — commits 9844429..219ad07, all five gates green between commits (build rc0, gofmt tracked-tree clean, vet clean, golangci-lint 0, go test ./... ok).
- Local HEAD 219ad07: go build ./... rc0 verified 2026-10-23; go vet ./... clean; gofmt -l clean on tracked tree (vendor/ excluded from repo style scope).
- CI green on pushed 219ad07: ci SUCCESS + bench SUCCESS (gh run list --commit 219ad07, 2026-10-22).
