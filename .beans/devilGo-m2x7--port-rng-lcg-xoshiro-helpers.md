---
# devilGo-m2x7
title: Port RNG (LCG + xoshiro + helpers)
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - engine
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-15T16:45:27Z
parent: devilGo-qna9
blocked_by:
    - devilGo-lfvi
---

# Port devilutionX RNG exactly

## Goal
Byte-exact Borland LCG port so every downstream roll sequence matches C++.

## Research
- Source/engine/random.cpp:16-96: xoshiro128plusplus::next (rotl(s0+s3,7)+s0), SetRndSeed, GenerateRandomNumber, AdvanceRndSeed, GenerateRnd, ReserveSeedSequence/jump.
- Source/engine/random.hpp:55-105: generateRnd high-bits trick (v<=0x7FFF → (advance>>16)%v), flipCoin, pickRandomlyAmong, RandomIntBetween(:391-405).
- Tests to port: test/random_test.cpp.

## Scope
internal/engine/random.go with: DiabloGenerator (uint32 state, Next() = (0x015A4E35*x+1) mod 2^32), AdvanceRndSeed int32 (abs with INT_MIN), GenerateRnd(v int32) int32 (negative-result quirk preserved), FlipCoin, RandomIntBetween(min,max,halfOpen), Xoshiro with Jump; global sglGameSeed semantics via a *Generator on GameState (no package-level mutable global in exported API; internal var allowed for port fidelity).

## Acceptance criteria
- [ ] Known-sequence test: SetRndSeed(2588) reproduces C++ first 1000 raw outputs (vectors generated from C++ test or hand-derived from LCG)
- [ ] GenerateRnd(65534) reproduces the <32768 vanilla bias (property test)
- [ ] AdvanceRndSeed never returns value outside [0,2^31) ∪ {INT_MIN} handling
- [ ] Xoshiro Jump doubles the stream (statistical test)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — integer math without unwanted overflow semantics surprises
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — property/table tests

## Dependencies & blockers
dg-task-001 (tree exists).

## Open questions
- Generate vectors: build devilutionX test binary once to dump 1000 outputs, commit as testdata (yes — non-negotiable for parity).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.
