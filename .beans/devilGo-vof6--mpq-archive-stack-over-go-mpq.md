---
# devilGo-vof6
title: MPQ archive stack over go-mpq
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - mpq
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:01:06Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-lfvi
---

# MPQ stack

## Goal
OpenAsset(name) with devilutionX's priority semantics; support unpacked dirs (UNPACKED_MPQS equivalent) and retail archives.

## Research
Source/engine/assets.cpp:385-437 LoadMPQ/FindMPQ, :505-594 loaders, :704-760 mods; assets.hpp:309-312,344 HaveMainData; mpq_reader.cpp wrapper over mpqfs; go-mpq API (Open/HasFile/ExtractFile→ adapt to bytes) https://pkg.go.dev/github.com/suprsokr/go-mpq

## Scope
internal/assets/mpq.go: Stack []Archive sorted by priority desc; Open(path) probing devilutionX search-path order; ReadFile(name) []byte; CLI subcommand to extract/dump for smoke tests.

## Acceptance criteria
- [ ] Smoke test extracts every file of user-provided spawn/diabdat/hellfire (skips politely when absent, like test/timedemo_test.cpp:47-49)
- [ ] Priority override unit test with two synthetic archives
- [ ] Unpacked-dir mode reads same names with forward/backslash normalization

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — io.Reader abstractions and error wrapping

## Dependencies & blockers
dg-task-001.

## Open questions
- go-mpq returns io streams vs buffers — wrap with bytes.Buffer caching LRU (size cap).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:external-gap-unrecorded — go-mpq codec limits (verified online) make the extract-everything AC false as written; scope must be reconciled with the recorded extract-at-build decision.
