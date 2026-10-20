---
# devilGo-a1n6
title: random-test-scope-trim-ponytail
status: completed
type: task
priority: normal
created_at: 2026-10-17T11:31:53Z
updated_at: 2026-10-20T00:16:31Z
parent: devilGo-qna9
---

# random_test.go scope trim (ponytail / 1:1-functionality rule)

## Goal
Apply the 2026-10-17 scope rule (milestone devilGo-i56j ## DECIDED) to internal/engine/random_test.go: 356 lines / 17 tests currently pin C++ scaffolding (double-jump statistics, reproducibility reruns) alongside genuine parity gates.

## Scope
- KEEP: golden sequence (SetRndSeed(2588) 1000 raw outputs), GenerateRnd bias/shift/mod range gates, AdvanceRndSeed INT_MIN, FlipCoin, xoshiro golden stream + Jump-changes-stream (Jump feeds GenerateSeed on the SP path)
- CUT or MERGE: statistical/reproducibility duplicates that re-prove the stdlib-free arithmetic already covered by goldens
- Target: parity evidence preserved, file <= ~200 lines

## Acceptance criteria
- [ ] go test ./internal/engine/ green after trim; every deleted test's premise either duplicated by a kept golden or game-uncalled (cite DECIDED rule)
- [ ] wc -l internal/engine/random_test.go <= 200
- [x] No narration comments introduced (coding-rules)

## Skills for coder
/home/blaze/.agents/skills/coding-rules/SKILL.md
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md

## Parent
devilGo-qna9

## TRIAGE 2026-10-18 (pi wave, batch 1)

devilGo-a1n6** — random_test.go scope trim
1. Root: test file pins C++ scaffolding; live: `internal/engine/random_test.go` = 356 lines / 17 tests (bean claims 356/17 ✓); scaffolding confirmed present: `TestXoshiro128plusplusDoubleJump:330`, `...Reproducible:297`, `...JumpDeterministic:345`.
2. Local: KEEP-list targets exist (`SetRndSeedGoldenSequence:23`, `FlipCoin:261`, `Xoshiro...JumpChangesStream:311`); `GenerateSeed`/`Jump` consumer verified `random.go:87-88`; milestone rule: i56j `## DECIDED (2026-10-17, scope rule)` names double-jump statistical proofs as cuttable.
3. Online: SKIPPED (no doc exists).
4. Scope-fit: pure test-file trim, no prod code; no overlap — no other test-trim bean found (uc44/1n0a are oracle/harness, different files).
5. DDD: engine/test layer only; no engine/game/diablo boundary impact.
6. Dep trigger: no blocked_by; prerequisite m2x7 (random.go port) = completed → dispatchable now.
7. AC gaps: KEEP/CUT list covers only 7 of 17 tests — 10 unlisted (e.g. `DiscardRandomValues:60`, `SetRndSeedStateAdvance:46`); AC's "every deleted test's premise…cited" forces per-cut justification, so workable but enumerate in PoW.
8. Validation: `go test ./internal/engine/` green; `wc -l` ≤200; diff shows no narration comments; every cut cites DECIDED rule or a kept golden.
9. Related: parent qna9; m2x7 (done); i56j (rule source); uc44/1n0a (do not touch oracle paths).
VERDICT: TRIAGED-OK:devilGo-a1n6

**devilGo-1spo** — Fixture storage, manifest, refresh flow
1. Root: fixtures have a producer (vsx3 completed, writes `tools/probe/out/`) but no owned storage/verify flow; live: `test/fixtures/` does not exist, no `fixture-refresh`/`fixture-verify` in justfile (grep empty).
2. Local: `docs/fixtures.md` exists and already names `manifest.txt → devilGo-1spo` as consumer (fixtures.md:15); probe pinned to dX commit ca91d778f; xcdz blocked_by 1spo (`devilGo-xcdz` frontmatter) so this bean gates the harness.
3. Online: SKIPPED — storage decision already recorded (committed goldens, CI verify-only; CI can't fetch retail MPQs).
4. Scope-fit: owns location+manifest+recipes; overlap seam: probe already emits manifest.txt while bean AC says "manifest format…committed" — clarify 1spo verifies/commits probe manifest rather than defining a second format; zstd note vs docs/fixtures.md format notes: no conflict (fixtures stored raw).
5. DDD: engine/testsupport + repo 

Wave outputs: /tmp/dg-triage/out-b1.md

## Libraries

- Phase rule (DECIDED 2026-10-17): pure-Go only, no CGO. Phase-0 packages ran stdlib-only; concrete deps (modernc.org/sqlite, suprsokr/go-mpq, gopher-lua, ebiten) attach at the beans that consume them — never substituted by a worker.

## Progress log
- 2026-10-20 claimed by orchestrator: lint blockers in missiles cleared by pi worker (b5edf48). Next: 3 delegated slices (classify 17 tests -> justified trim <=200 -> focused gates).

## Proof of Work

- Classification (proc_b82c6911a4e8, exit 0, marker present): /tmp/dg-a1n6-classification.md — all 17 tests inventoried, 10 unlisted enumerated, caller relevance verified live (quests.go:135, path_test.go:41, HookGenerateRnd seams). Spot-checked by parent: line ranges 188-236/248-259 exact, #11 confirmed pure statistical scaffolding.
- Trim (proc_3780d7117979, exit 0, marker present): 4 CUT (TestGenerateRndRange, TestXoshiro128plusplusReproducible, TestXoshiro128plusplusDoubleJump, TestXoshiro128plusplusJumpDeterministic — premises: statistical/reproducibility scaffolding per i56j ## DECIDED), 3 MERGE with semantics preserved (GetLCGEngineState mirror folded to golden test; GenerateRnd(1)==0 + state-advance folded to InvalidRange; INT_MIN % 2^k comment row folded to NegativeReturnValues), 10 RETAIN untouched.
- Gates: gofmt empty, go vet clean, go test ./internal/engine/ ok, go test ./internal/... 19 pkgs 0 fails, golangci-lint 0 issues, wc -l = 200.
- Parent re-gate: full diff read — no parity-row loss, no random.go changes, single-file diff.

Residual: none.
