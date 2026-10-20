---
# devilGo-klu5
title: coding-rules + ponytail full pass over all Go files (per-file verdicts)
status: todo
type: task
priority: high
tags:
    - quality
created_at: 2026-10-20T13:03:05Z
updated_at: 2026-10-20T13:16:53Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-7otq
---

## Goal
Full coding-rules + ponytail pass over EVERY .go file in the repo (post 3-layer restructure: engine/, game/, diablo/, cmd/), per user directive 2026-10-20. Prior sweeps touched 11/96 files; this epic-round must produce per-file verdicts.

## Acceptance criteria
- [ ] One verdict line per .go file (CLEAN or file:line finding) — count verified against `find . -name '*.go'` inventory
- [ ] All coding-rules violations fixed: narration comments gone, intent/mechanism separation, no vague names, duplication collapsed, file-structure order, black-box given/when/then tests
- [ ] All ponytail violations fixed: dead abstraction/wrapper/option removed, shortest working diff, comment only why, no features nobody asked for (deletion beats addition)
- [ ] `//nolint:revive // 1:1 port` functions stay structurally whole (1:1 fidelity mandate); comments inside them may still be cleaned
- [ ] Golden/parity test assertion semantics untouched
- [ ] Gates: go build ./... rc0; gofmt -l empty; go vet clean; golangci-lint 0 findings; go test ./... all ok; per-file net LOC delta table in PoW

## Skills for coder
- Read /home/blaze/.agents/skills/coding-rules/SKILL.md fully BEFORE editing — binding checklist.
- Read /home/blaze/.agents/skills/ponytail/SKILL.md fully BEFORE editing — smallest complete change, deletion beats addition.
- Read /home/blaze/.agents/skills/go-dev/SKILL.md fully BEFORE editing.

## Rules for coder
- GIT JAIL: no git checkout/switch/reset/commit/push; read-only git allowed.
- No new deps, no nolint/suppressions, no test weakening.
- One domain per wave; report per-file LOC delta.

## Triage (2026-10-20, research-verified)
- Live state: 96 .go files post-restructure (engine 30-ish / game / diablo incl. new combat+gamedata+testsupport+txtdata homes; cmd 2). Prior sweep verdict pass was pre-restructure; move itself added package-doc staleness (S9 sweep) — this bean re-verifies the whole tree.
- Blocked-by devilGo-7otq: `go fix` rewrites land first; sweeping before modernization churns files twice.
- Wave plan (one domain per wave, ~10-min units): W1 engine (assets, demo, txtdata, crawl/path/random/tick) → W2 game (world, testsupport, game.go) → W3 diablo (combat, gamedata 8 pkgs, txtdata, missiles, enums, save, animation, testsupport) → W4 cmd + doc.go manifests + AGENTS.md layer table.
- Per-file verdict gate: verdict lines must count 96 against `find internal cmd -name "*.go" | wc -l` — no package-level hand-waving (prior lesson).
- ponytail mandate: deletion beats addition; no new abstractions; comment only why; smallest complete diff.
- Ready: ACs complete, waves enumerated, evidence cited. → todo after 7otq completes.

## Triage correction (2026-10-20)
Prior inline triage withdrawn per user directive — triage is pi-brainer lane. Reset to draft for brainer triage pass.

## Amendment (brainer triage 2026-10-20) — wave re-split (9 waves, ≤15 files/~1.5k LOC each)
W1 engine core (11) · W2 engine assets+demo+txtdata (12) · W3 game (7) · W4 diablo missiles (15) · W5 gamedata A: misdat+objdat+leveldat+monstdat (13) · W6 gamedata B: audio+quests+textdat+towners (9) · W7 diablo txtdata (9) · W8 combat+enums+save+animation (14) · W9 testsupport (4) + cmd (2) + doc.go manifests. Per-file verdicts still count 96 via `find internal cmd -name "*.go" | wc -l`.
## Amendment 2 — precedence
Golden/parity/tape assertion semantics ALWAYS win over the style pass; 1:1 tape-port test files get verdict PORTED-EXEMPT instead of restructure.
## Amendment 3 — broaden nolint protection
ALL //nolint markers carrying 1:1-port citations (12 today, incl. nestif/gomnd) stay; their functions stay structurally whole. No new suppressions.
## Amendment 4 — bound ponytail deletions
Removals limited to non-ported, non-exported surface; ported exported functions exempt (PORTED-EXEMPT).

## Libraries
- stdlib only (no third-party deps at this stage)
