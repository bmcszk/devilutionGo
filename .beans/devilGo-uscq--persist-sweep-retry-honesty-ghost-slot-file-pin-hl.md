---
# devilGo-uscq
title: 'persist: sweep retry honesty + ghost-slot file pin (hl3 review follow-ups)'
status: completed
type: task
priority: normal
created_at: 2026-11-25T16:52:39Z
updated_at: 2026-11-26T12:15:57Z
parent: devilGo-p1t8
---

# persist: sweep retry honesty + ghost-slot file pin (hl3 review follow-ups)

## Goal
Two non-blocking review findings on the per-hero store migration (hl3 review, PR #85), resolved so they do not rot:
1. rehomeLegacySingleFile's comment claims a mid-failure sweep is retried on the next construction; the code does not implement retry. RESOLVED BY TRIAGE: correct the comment (retry is unreachable by design — the rename precedes the loop and the source rows live only inside the marker file).
2. A corrupt NON-slot ghost file in the hero directory is not pinned by a test: ListHeroes probes only single_N.db names, so a ghost file is silently ignored. RESOLVED BY TRIAGE: pin the ignore behaviour with a test (matches upstream pfile.cpp:695-703, which silently skips both failed open and failed ReadHero and never scans the directory).

## Research
- Triage verdict (task-m2fu-triage-brainer, 2026-11-26, parent spot-checked all line cites): internal/persist/hero_store_set.go:10-12 constant doc already states the one-time/rollback design; :187-189 marker short-circuit; :205 rename BEFORE the save loop; :216 the false "a retry re-runs" claim; legacyRows (:220-233) is hard-wired to single_0.db so there is nothing to resume from without new code. Implementing retry needs a failure seam that does not exist and a re-decision on user-visible half-swept dirs — a design change, not a fix. Correcting the comment is the smaller honest option.
- Upstream verified on disk (/home/blaze/github/bmcszk/devilutionX Source/pfile.cpp:693-706): pfile_ui_set_hero_infos loops slots, `if (archive)` and `if (ReadHero(...))` both silently skip, no directory walk. A non-slot ghost error would have no upstream counterpart. ADR-003 is silent on enumeration policy; the recorded "deviation 1" (slot file that fails open/read = ERROR) is unaffected.
- Test-author note: a corrupt SLOT file (garbage in single_1.db) already errors by design (OpenSQLiteStore → migrate fails, store_sqlite.go:84) — the pin test must use NON-slot names only.

## Scope
- Correct the :216 comment to the real semantics: a mid-loop failure leaves the already-written single_<slot>.db files + the marker in place, the next construction is a no-op, and the un-re-homed rows remain readable only inside single_0.db.pre-per-hero (the rename is the rollback). No code change to rehomeLegacySingleFile.
- New table-driven test (RED first) pinning ghost-file ignore.

## Acceptance criteria
- [x] `hero_store_set.go` sweep comment states the real semantics: a mid-loop failure leaves written slot files + the marker in place, the next construction is a no-op, un-re-homed rows stay only inside `single_0.db.pre-per-hero`; no retry is claimed (no code change to `rehomeLegacySingleFile`)
- [x] New table-driven test (RED first) in `hero_store_set_test.go` pins that a non-slot ghost file (>=2 names incl. one `.db`-suffixed non-slot name) in the hero dir is ignored: `ListHeroes` returns only the real slot rows and no error, `HasSlot` for every real slot is unchanged
- [x] Gates green (`gofmt -l` empty, `go vet` rc0, `golangci-lint` 0, `go test ./internal/persist/...` ok) with RED/GREEN tails pasted in the PoW; `-tags windowed` gate unaffected (no windowed files touched)

## Libraries
stdlib + modernc.org/sqlite (existing).

## Skills for coder
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md — smallest landable diff
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Filed 2026-11-25 from the hl3 independent review (APPROVE-WITH-FINDINGS; findings non-blocking, neither lost data — reviewer's own words). Deep triage DONE 2026-11-26 (task-m2fu-triage-brainer; parent-verified: every file:line cite spot-checked on the tree, upstream pfile.cpp read on disk). Verdict: PROMOTE-TODO — comment correction (no retry code) + table-driven ghost-pin test. Full verdict: .hermes/lane/m2-followups/triage-result.md


## Summary of Changes

Leg 1 of the m2-followups wave (branch fix/m2-followups, commit 07a7615). The sweep-failure comment in rehomeLegacySingleFile was corrected to the real semantics (mid-loop failure leaves written slot files + the marker; next construction is a no-op; un-re-homed rows stay only inside single_0.db.pre-per-hero; no retry exists — the rename is the rollback). No code change. A table-driven test (3 ghost names: backup_single_0.db, single_0.db.pre-per-hero, notes.txt) pins that ListHeroes ignores non-slot ghost files and HasSlot is unchanged for real slots.

## Proof of Work

- TDD sequence: test written first, run against the UNFIXED tree — GREEN (premise falsification: the ignore behaviour already exists; the test pins it). Verbatim unfixed-tree PASS output (all 3 subtests) in .hermes/lane/m2-followups/wave-result.md. NB (review NB-2): the wave log's phrase "the bean explicitly allows" overstates the bean text — the bean allowed RED-or-falsification implicitly via the triage-resolved AC wording; recorded here for accuracy.
- Parent gate re-run (CI-shaped) at d4356f1: gofmt -l empty; go build + go build -tags windowed rc0; go vet both tags rc0; scripts/lint.sh 0 issues; go test ./... -count=1 rc0; windowed pair 123 passed; check-no-raw-sql clean.
- Clean-clone oracle: git clone --local --no-hardlinks -> checkout fix/m2-followups -> go test ./internal/... -count=1 rc0 at d4356f1.
- Independent review (task-m2fu-review-brainer): APPROVE, no blocking findings. AC1/AC2/AC3 all judged MET with file:line evidence; NB-1 (frameImage third copy of the frame->NRGBA loop) filed as its own follow-up draft bean; NB-2 wording note recorded above.
- Residual: retry remains unimplemented BY DESIGN (documented in code); recovery story stays the manual rename-back rollback. No e2e impact — persist suite green.
