---
# devilGo-78pw
title: 'coding-rules full matrix: every rule x every file (user directive 2026-10-23)'
status: completed
type: task
priority: high
created_at: 2026-10-23T17:07:43Z
updated_at: 2026-10-29T15:21:06Z
parent: devilGo-1n0a
---

User directive: EVERY ONE rule from coding-rules applied to EVERY ONE source file; fix and verify. Method: mechanical matrix (/tmp/devilgo-sweep/rule-matrix.py) + brainer judgment sweep for non-mechanical rules. Mechanical findings: 145 exported syms w/o outside callers (R36), narration comment process_test.go:513, scope TODO process.go:574, test func in path.go:347. nolint hits legal (cited WHY).

## Acceptance criteria
- [ ] All files carry a per-rule verdict (matrix + brainer), coverage reconciled 104/104
- [ ] Mechanical findings fixed via pi; gates green; go test ./... green
- [ ] CI success on push

## Libraries
- none new

## Skills for coder
- /home/blaze/.agents/skills/coding-rules/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/ponytail/SKILL.md — read fully BEFORE acting
- /home/blaze/.agents/skills/go-dev/SKILL.md — read fully BEFORE acting

## Triage
Delegated survey: mechanical rule matrix ran over all 104 files (git ls-files), findings deduped by hand. Brainer judgment sweep pending.


## Progress log
- 2026-10-23 matrix built (rule-matrix.py: narration/TODO/orphan-markers/visibility/vague-names/nolint-why per file) + brainer judgment audit (devilGo-78pw-judge: rules A/B/C/E/F, 1:1-port + infra calibration).
- Fixes via pi: 78pw-a (batch A unexport, da3935c), 78pw-b/b2 (batch B + comment hygiene, 5bfd1bb), 78pw-d (lint zero: unused triaged/SA4016/unparam/35 testpkg conversions, f61e370), 78pw-e (parity exports justified, 8222654), 78pw-jfix (judgment findings 1-8: enginetxtdata consolidation, dead code, one parity table, canonical DungeonType, 29b7c8d).

## Proof of Work
- Coverage: 104/104 files passed through the matrix every round (git ls-files minus vendor); judgment audit covered all 104 with explicit calibration.
- Final state: golangci-lint ./... = 0 issues; go vet = rc0; gofmt = clean; go test ./... = 9681 passed in 28 pkgs; go build rc0; CI success at 29b7c8df.
- Residual accounting: R36 76 raw decls = 73 with verified outside-package users + 3 parity-harness exports carrying WHY doc lines (intentional API). R40 57 nolint markers all carry WHY comments (34 testpackage internal-state + unused-with-bean-refs + 1:1-port cites). R14 narration/TODO = 0. Phase markers clean. Judgment findings: 8/8 fixed (net -144 lines).
- Skips with reasons: NewGame (headless cmd user), Movies/SetmapManifest/ShrineGametype (name collisions), testpackage x6 (genuine internal-state tests).

## AUDIT (brainer 2026-10-24, beans-completion audit vs tree)
STALE (brainer audit 2026-10-24): ledger describes tree states (57 nolint
markers, export set) fully superseded by later waves (kefy, kc11, lu2v);
today tree has zero //nolint and .golangci.yml is at revert state.
Treat this bean as superseded-by devilGo-kefy/devilGo-lu2v — accounting
here no longer reflects the tree. No action needed beyond this note.
