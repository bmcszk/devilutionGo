---
# devilGo-uc44
title: Timedemo oracle test + seed matrix CI job
status: todo
type: task
priority: normal
tags:
    - plan:task
    - parity
    - ci
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-11-28T13:05:02Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-w9or
    - devilGo-xc37
---

---
# devilGo-uc44
title: Timedemo oracle test + seed matrix CI job
status: todo
type: task
priority: normal
tags:
    - plan:task
    - parity
    - ci
created_at: 2026-10-15T16:35:18Z
updated_at: 2026-11-03T12:52:18Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-w9or
    - devilGo-6xu2
    - devilGo-1spo
---

# The oracle

## Goal
Wire reference-save comparison + fixture matrix into CI.

## Research
timedemo_test.cpp:47-49,96-100; fixtures listing; pfile_compare_hero_demo.

## Scope
internal/demo/timedemo_test.go; justfile parity target; CI workflow.

## Acceptance criteria
- [ ] Green with spawn.mpq present; graceful skip otherwise
- [x] Matrix run < 5min CI

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- \`~/.agents/skills/coding-rules/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/ponytail/SKILL.md\` — read fully BEFORE acting
- \`~/.agents/skills/go-dev/SKILL.md\` — read fully BEFORE acting
/home/blaze/.agents/skills/go-tests/SKILL.md — CI-friendly skips and reports

## Dependencies & blockers
- devilGo-6xu2 (fix2: verified wired; body refs refreshed)
devilGo-w9or, dg-epic-08.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:frontmatter-missing-persistence-dep;stale-devilGo-w9or/dg-epic-08;no-explicit-TDD-first-AC


## IMPL NOTES (triage 2026-10-16)
1 root: CI oracle wiring — run demo fixtures, compare hero saves via pfile_compare parity, seed matrix <5min; status `draft`, parent 1n0a, blocked_by w9or ✓ (demo engine first).
2 local: `test/timedemo_test.cpp` — MPQ-skip guard ~:47-52 (HaveMainData → GTEST_SKIP, cited :47-49, regionally correct), `pfile_compare_hero_demo` at :94 with `ASSERT_EQ(result.status, Same)` ~:96-100 ✓; implementation `Source/pfile.cpp:665` ✓; fixture matrix on disk ✓. Body deps cite stale `devilGo-w9or, dg-epic-08` (044 = demo engine → w9or, correctly in frontmatter; epic-08 = persistence → devilGo-6xu2, MISSING from frontmatter despite the oracle requiring the .sv compare path). AC list has no explicit TDD-first item.
3 online: n/a (no external claims).
4 solution: add devilGo-6xu2 (or its .sv-compare child) to frontmatter deps; refresh stale ids; make TDD-first explicit (e.g. "- [ ] TDD: oracle red against a deliberately-corrupted spawn_0.sv copy before wiring CI").
5 understanding: the end-state oracle of the whole program — everything upstream must make this green; consumes w9or playback + persistence compare.
6 repro: `sed -n '47,52p;92,100p' devilutionX/test/timedemo_test.cpp`; `beans show devilGo-uc44 --json` (blockedByIds: [w9or] only).
7 fix: `beans update devilGo-uc44 --blocked-by devilGo-6xu2`; body replace "devilGo-w9or, dg-epic-08." → "devilGo-w9or, devilGo-6xu2."; append explicit TDD-first AC as above.
8 validation: `beans query --json '{ bean(id:"devilGo-uc44") { blockedByIds } }'` includes 6xu2; body grep TDD.
9 related: devilGo-w9or (dep), devilGo-6xu2 (persistence), devilGo-i2cr (sibling CI gate).
VERDICT: NEEDS-WORK:frontmatter-missing-persistence-dep;stale-devilGo-w9or/dg-epic-08;no-explicit-TDD-first-AC


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.

## Zero-dependency test constraint\n\nParity tests compare committed Go-owned fixtures only. C++ dumpers/probes may be used once in an offline producer workflow, but must never run in go test, CI, production code, or ordinary fixture verification. No DG_DEVILUTIONX or external source-tree access; missing fixtures are hard failures.


2026-11-03 ADR alignment: the timedemo file becomes the command file through
the kernel once the dispatcher child lands, and the per-tick checksum trail
arrives with the tick-log child so parity runs report the first diverging
tick with both hash trails ([ADR-018](../docs/adrs/adr-018.md) law 4,
[ADR-015](../docs/adrs/adr-015.md)). Kernel parent: [Deterministic command/event simulation kernel (ADR-015, ADR-018)](devilGo-xc37--deterministic-commandevent-simulation-kernel-adr-0.md);
tick-log child: [Tick log with 1000-tick snapshots and per-tick checksum replay](devilGo-wskw--tick-log-with-1000-tick-snapshots-and-per-tick-che.md)

## Architecture guardrails (tech-debt triage 2026-11-12, binding for this bean's lane)

Source: the ok51 tech-debt triage (task-techdeptTriage-brainer) — these are the
lessons from the mistakes already made, applied BEFORE this bean starts. The
contract for this bean restates whichever apply; violating one is a review
REQUEST-CHANGES, not a debate.

1. Program-surface law (ADR-020/021 + one-tool law): no new binaries. Fixture
   instruments are subcommands of tools/goldens; user-facing surface stays
   exactly cmd/devilutionGo + cmd/devilutionGo-migrate. No `func main` outside
   the allowlist.
2. Tooling problems: fix immediately, no hacks.



Split 2026-11-27 (parity-close triage): the .sv writer (LegacySVStore.Write, ExportSV) and the comparator (CompareSaves, CompareHeroDemo) are LANDED and tested - the premise that this waits on the persistence epic was false. The real blocker: NO production simulation path can produce a level-up (XP exists only as a pure function + test handlers), so the WarriorLevel1to2 replay cannot produce a state to compare. Blocker re-pointed to the Deterministic-command/event-simulation-kernel epic (xc37). Upstream pin e1ad0c2b6c1721686928e9ef75fc0447037bd3f2; oracle = whole-entry BYTE EQUALITY of decoded archive entries (pfile.cpp:476-524 CompareSaves), not hashes.
- [ ] Reference save committed as text (demo_0_reference_spawn_0.sv.hex), sha256 recorded, provenance = pinned upstream commit.
- [ ] Test: the reference archive loads, its hero entry decodes, entry names/sizes match a committed expectation.
- [ ] Test: comparator yields Same for reference-vs-reference and Difference (naming the entry) for a one-byte flip.
- [ ] Test: the real demo_0.dmo.hex replays through DemoReplay; record histogram + tick count equal committed values.
- [ ] (BLOCKED xc37) WarriorLevel1to2 replay produces an actual .sv whose decoded entries are byte-equal to the reference, via persist.CompareHeroDemo with Status == DemoSame.



## Proof of Work (seed-matrix half, parity-close wave 2026-11-27)

- `just parity` EXISTS: runs every committed seed-matrix test (1131 subtests incl. subtests, 0 SKIP), exits non-zero on failure. MEASURED 18-22s (levels 17.7-21.6s + integration ~5.8s) - the '< 5 min' AC met with ~14x headroom; the measurement replaces the claim.
- D1: one source of truth - the justfile recipe defines the command; ci.yml's new `parity` job and scripts/git-hooks/pre-push both call `just parity` (single place to drift).
- Asset-free: the job sets no DG_ASSET_ARCHIVES; committed text goldens only, no skips in the run.
- Timedemo half: OPEN on the kernel epic (xc37) - the five new boxes above; the .sv writer + comparator are landed; the missing piece is a simulation that can produce the level-up.
