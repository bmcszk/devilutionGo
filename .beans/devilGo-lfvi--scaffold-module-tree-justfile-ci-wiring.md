---
# devilGo-lfvi
title: Scaffold module tree + justfile + CI wiring
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - infra
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T20:38:06Z
parent: devilGo-qna9
---

# Scaffold module tree + justfile + CI wiring

## Goal
Create the package tree from docs/research/dg-requirements.md Appendix E, wire just + CI to vet/lint/test, document layout in README.

## Research
- go-setup skill layout guidance (/home/blaze/.agents/skills/go-setup/SKILL.md)
- Repo already has .golangci.yml, justfile, go.mod (module github.com/bmcszk/devilutionGo, go 1.27.1)

## Scope
cmd/, internal/{engine,world,entity,gameplay,net,assets,render,ui,persist,input,platform,demo,game}, testdata/. Empty package doc.go files. just targets: build, test, vet, lint, bench, parity.

## Acceptance criteria
- [ ] `just build` produces cmd/dgo and cmd/dgoheadless binaries
- [ ] `just vet lint test` all green from empty packages
- [ ] README documents the tree and the one-way import rule (only internal/game imports ebiten)

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
/home/blaze/.agents/skills/go-setup/SKILL.md — canonical Go project structure for a multi-package game engine port
/home/blaze/.agents/skills/coding-rules/SKILL.md — file layout and comment conventions

## Dependencies & blockers
None.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): DECISION-NEEDED: which package layout governs — repo AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) vs docs/research/dg-requirements.md Appendix E 14-package tree (internal/{world,entity,gameplay,render,ui,…}, cmd/dgo)? Cascades into all downstream Scope paths + binary names.


## IMPL NOTES (triage 2026-10-16)
1 root: Create package tree + wire just/CI; status `todo`, parent qna9, no blockers. Body embeds docs/research/dg-requirements.md Appendix E 14-package layout.
2 local: Appendix E exists (docs/research/dg-requirements.md:798+: cmd/dgo, internal/{engine,world,entity,gameplay,net,assets,render,ui,persist,input,platform,demo,game}). CONFLICT 1 — repo AGENTS.md:19-23 (newer, self-declared single source of truth) mandates three-layer internal/{engine,game,diablo} + import rules + cmd/devilutionGo & cmd/devilutionGo-headless (AGENTS.md tree), and justfile already wires BIN_NAME="devilutionGo", build-gui → ./cmd/devilutionGo (justfile:21,29,45,50). CONFLICT 2 — bean AC says "cmd/dgo and cmd/dgoheadless binaries". Also partially done already: .golangci.yml/go.mod/justfile exist (verified ls), and all six wanted just targets already defined (build:54, test:96, vet:80, lint:76, bench:121, test-parity:117); cmd/ and internal/ trees do NOT exist (ls: no such dir).
3 online: n/a (no external claims).
4 solution: user must pick governing layout (AGENTS.md three-layer vs Appendix E 14-package) — the choice cascades into Scope paths of ~every downstream bean (ohsu internal/game/game.go, m2x7/v4lb internal/engine/…, xcdz internal/testsupport) and binary names; then rewrite Scope/AC to that layout, shrink scope to "create tree + README + CI" (just wiring already done).
5 understanding: the root structural bean of the port — every layer boundary decision lands here first; wrong tree now = mass re-pointing later.
6 repro: `beans show devilGo-lfvi --body-only`; `sed -n '798,830p' docs/research/dg-requirements.md`; `sed -n '17,23p' AGENTS.md`; `grep -n 'BIN_NAME\|build-gui\|CMD_DIR' justfile`; `ls cmd internal` (absent).
7 fix: none until decision; after: `beans update devilGo-lfvi --body-replace-old "<Appendix E scope line>" --body-replace-new "<governing layout>"` + AC binary names updated to match justfile/AGENTS.
8 validation: bean Scope tree == governing doc; AC `just build` binary names match justfile BIN_NAME/BIN_HEADLESS.
9 related: qna9 (parent), ohsu+m2x7 (blocked by this), every epic-02+ bean (inherits tree).
VERDICT: DECISION-NEEDED: which package layout governs — repo AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) vs docs/research/dg-requirements.md Appendix E 14-package tree (internal/{world,entity,gameplay,render,ui,…}, cmd/dgo)? Cascades into all downstream Scope paths + binary names.


## LAYOUT RULING (user decision 2026-10-16, supersedes Appendix E above)
- Governing layout = repo AGENTS.md three-layer: `cmd/devilutionGo`, `cmd/devilutionGo-headless`, `internal/{engine,game,diablo}`, `shaders/`, `test/{fixtures,parity}`, `assets/`.
- Import rules (binding): engine imports NOTHING from game/diablo; game may import engine; diablo may import both.
- Scope narrowed: existing .golangci.yml/go.mod/justfile stay (already complete — triage verified all 6 just targets exist); remaining work = create package tree with doc.go stubs, README layout section, CI workflow wiring gates.
## Acceptance criteria (corrected)
- [ ] `just build-gui` → bin/devilutionGo; `just build-headless` → bin/devilutionGo-headless (names match justfile BIN_NAME/BIN_HEADLESS)
- [ ] just vet lint test green from stub packages
- [ ] README documents three-layer tree + import rules
- [ ] (TDD mandate n/a: pure scaffold — no game logic; parity n/a)

TRIAGE resolution 2026-10-16: DECISION-NEEDED resolved by user — AGENTS.md three-layer governs. Verdict now TRIAGED-OK.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
