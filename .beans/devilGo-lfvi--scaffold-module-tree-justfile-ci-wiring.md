---
# devilGo-lfvi
title: Scaffold module tree + justfile + CI wiring
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - infra
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-15T16:45:27Z
parent: devilGo-qna9
---

# Scaffold module tree + justfile + CI wiring

## Goal
Create the package tree from /tmp/dg-requirements.md Appendix E, wire just + CI to vet/lint/test, document layout in README.

## Research
- go-setup skill layout guidance (/home/blaze/.agents/skills/go-setup/SKILL.md)
- Repo already has .golangci.yml, justfile, go.mod (module github.com/bmcszk/devilutionGo, go 1.27.1)

## Scope
cmd/, internal/{engine,world,entity,gameplay,net,assets,render,ui,persist,input,platform,demo,game}, testdata/. Empty package doc.go files. just targets: build, test, vet, lint, bench, parity.

## Acceptance criteria
- [ ] `just build` produces cmd/dgo and cmd/dgoheadless binaries
- [ ] `just vet lint test` all green from empty packages
- [ ] README documents the tree and the one-way import rule (only internal/game imports ebiten)

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
