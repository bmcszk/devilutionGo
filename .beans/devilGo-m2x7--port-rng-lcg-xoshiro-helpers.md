---
# devilGo-m2x7
title: Port RNG (LCG + xoshiro + helpers)
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-0
    - engine
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T21:50:18Z
parent: devilGo-qna9
blocked_by:
    - devilGo-lfvi
    - devilGo-1spo
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

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — integer math without unwanted overflow semantics surprises
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md — property/table tests

## Dependencies & blockers
devilGo-lfvi (tree exists).

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: RESEARCH-FIRST: RNG vectors dumped from a devilutionX test build once (probe toolchain bean owns the dumper); vectors committed as testdata — non-negotiable for parity.
- Generate vectors: build devilutionX test binary once to dump 1000 outputs, commit as testdata (yes — non-negotiable for parity).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: Byte-exact Borland LCG + xoshiro128plusplus port; status `todo`, parent qna9, blocked_by lfvi. Structurally complete.
2 local: random.cpp:16-96 claims verified — GenerateRnd bias trick `(AdvanceRndSeed() >> 16) % v` at :100-103, xoshiro::next rotl(s0+s3,7)+s0 at :19-34, ReserveSeedSequence/jump at :54-56; random.hpp generateRnd doc incl. "limits 32768..65534 always <32768" vanilla bug + negative-result quirk at :55-95, RandomIntBetween(min,max,halfOpen) at :403, jump at hpp:220; test/random_test.cpp exists (312 lines, ShiftModDistributionHighBits at :184 covers the 65534-bias AC); skills paths exist.
3 online: n/a (no external claims).
4 solution: none substantive — ready. Optional: update body dep line devilGo-lfvi → devilGo-lfvi.
5 understanding: engine-layer deterministic core; every downstream system (DRLG, drops, combat) consumes this RNG, so parity here is load-bearing for the whole port.
6 repro: `beans show devilGo-m2x7 --body-only`; `sed -n '96,110p' ~/github/bmcszk/devilutionX/Source/engine/random.cpp`; `grep -n RandomIntBetween …/random.hpp`; `sed -n '184,220p' …/test/random_test.cpp`; `sed -n '13,16p' .beans/devilGo-m2x7*.md` (frontmatter blocked_by: devilGo-lfvi).
7 fix: `beans update devilGo-m2x7 --body-replace-old "devilGo-lfvi (tree exists)." --body-replace-new "devilGo-lfvi (tree exists)."`
8 validation: `beans show devilGo-m2x7 --body-only | grep dg-task` → empty; frontmatter unchanged.
9 related: lfvi (blocker), v4lb (blocked by this), ohsu (sibling engine pkg).
VERDICT: TRIAGED-OK
