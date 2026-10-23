---
# devilGo-lu2v
title: 'coding-rules full matrix round 2: ALL rules (agnostic+Go+Go-test) x ALL 104 files'
status: completed
type: epic
priority: high
created_at: 2026-10-23T19:55:42Z
updated_at: 2026-10-23T22:06:33Z
---

User directive: find ALL rules (language-agnostic + Go + Go-test), list them, list all files, EVERY FILE x EVERY RULE. Method: (1) rule extraction from all skills (brainer, 3 parallel extractions running), (2) rule list consolidated into /tmp/devilgo-sweep/ALL-RULES.md, (3) per-package brainer batches (2 batches, ~9k lines each) judging every rule x every file, (4) fix batches via pi, (5) parent gates + residual ledger, (6) bean closes with PoW.

## Acceptance criteria
- [ ] ALL-RULES.md exists with every rule from every skill (M/J tagged)
- [ ] 104/104 files judged against every J-rule; M-rules verified mechanically
- [ ] Violations fixed via pi; gates green; CI success
- [ ] Residual ledger (legit + WHY) in bean PoW


## Progress log
- 2026-10-23 claimed: rule extraction (3 brainer extractions: code-style, testing, repo-gates) -> ALL-RULES.md (126 rules: 60 M / 66 J); file list 104 (units.json); 2 judgment batches dispatched (b0: 32 files, b1: 72 files).


## Progress log
- 2026-10-23 judgments complete: b0 (32 files) = 3 violations (process.go garbled seam comments; process_ii_test + process_test I4 self-referential oracles); b1 (72 files) = 7 violations in 6 files (quests.go B2 doc mismatch; monstdat/misdat/audio/towners _test A9/A11; testsupport/demo.go A13). 95/104 CLEAN on first pass.
- 2026-10-23 all 10 fixed by pi worker devilGo-lu2v-fix (15 files touched, commit 6b546f2): oracles re-derived from displacement.hpp citations; shared Corpus helper; demo constants deduped; doc fixes; dup nolints removed. Parent re-verified: corpus.go exists, 6 corpus users, dup nolint 1/file, TestHFMonstdatStrictSuperset renamed.

## Proof of Work
- Coverage: ALL-RULES.md = 126 rules (60 M / 66 J) consolidated from coding-rules + go-dev(+refs) + golang-{code-style,concurrency,error-handling,performance,testing} + go-fluent-testing + tdd + port-parity-testing + ponytail + repo (.golangci.yml v2.15.0 linters, just gates, AGENTS.md). File list: units.json = 104 files (57 prod 7085 lines, 47 test 10519 lines).
- Judgment: b0 32/32 files judged (29 CLEAN, 3 violations); b1 72/72 files judged (66 CLEAN, 7 violations). M-rules: lint 0 issues (v2.15.0) covers 60 M-rules mechanically; matrix v6 residuals: R36 76 = 73 outside-users + 3 parity-harness WHY exports; R40 52 nolint all WHY-carrying (verified per-line).
- Fixes: commit 6b546f2 (18 files, +592/-133); gates at 680096d: build rc0, gofmt clean, vet rc0, lint 0, 9696 tests in 29 pkgs, CI success.
- Residuals with WHY: 3 parity-harness exports (AssertTable/CompareTable/LoadFixture, repo-wide parity API); 52 nolint (41 testpackage internal-state true-pokes verified, 1:1-port cites, unused-with-bean-refs); NewGame + 3 name-collision unexport skips recorded.


## Progress log
- 2026-10-23 wave 3 (user rulings): (1) tests-first order — 32+5 test files reordered, helpers now after tests repo-wide (verified 0 remaining); (2) const/var grouping — adjacent singles grouped into ( ) blocks in 13 prod files + random.go straggler (verified 0 remaining); (3) prod order = caller-above-callee only (consts/types stay); (4) .golangci.yml confirmed untouched by sweep passes (config change predates round; user alerted, no new config edits). Commits c5c00b9 + 5373aef. CI success 5373aefd. Bean re-opened -> completed again with these rulings recorded.
