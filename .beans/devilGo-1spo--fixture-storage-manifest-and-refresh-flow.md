---
# devilGo-1spo
title: Fixture storage, manifest and refresh flow
status: todo
type: task
priority: high
created_at: 2026-10-16T20:36:59Z
updated_at: 2026-10-17T19:56:52Z
parent: devilGo-1n0a
blocked_by:
    - devilGo-vsx3
---

## Goal

Own fixture storage and refresh in this repo: canonical location, hash manifest, refresh command, and the rule for when goldens regenerate. Consumer beans compare against these files; this bean decides where they live and how they stay honest.

## Research

- Fixture classes (produced by the C++ probe toolchain bean): RNG vectors, backbuffer frames + palettes, light tables + falloffs, seed dungeons, .sv saves.
- Git handling: binary fixtures need stable storage; raw backbuffers compress well enough with zstd at repo scale (per docs/fixtures.md manifest). Large media stays out; generated levels are small.
- Refresh flow: probe toolchain writes to test/fixtures/generated/, manifest (sha256 per artifact + devilutionX commit) committed alongside.

## Scope

- test/fixtures/ tree layout (generated/ + static/), README per class
- just fixture-refresh recipe: run probe (against pinned devilutionX), verify hash manifest, commit
- CI mode: verify-only (recompute hashes, never regenerate)

## Acceptance criteria

- [ ] test/fixtures/ layout exists with per-class READMEs
- [ ] manifest format: devilutionX commit + per-file sha256, committed
- [ ] just fixture-refresh regenerates and reports drift; just fixture-verify checks hashes only
- [ ] docs/fixtures.md updated by the refresh flow automatically (artifact list)

## Libraries

- None beyond stdlib (hashing) at this stage.

## Skills for coder

/home/blaze/.agents/skills/go-dev/SKILL.md
/home/blaze/.agents/skills/go-fluent-testing/SKILL.md

## Dependencies & blockers

- devilGo-vsx3 (probe toolchain must exist to generate anything)

## Open questions

- None. Decided: committed goldens (CI cannot fetch retail MPQs legally); verify-only in CI.

## TRIAGE 2026-10-18 (pi wave, batch 1)

devilGo-1spo** — Fixture storage, manifest, refresh flow
1. Root: fixtures have a producer (vsx3 completed, writes `tools/probe/out/`) but no owned storage/verify flow; live: `test/fixtures/` does not exist, no `fixture-refresh`/`fixture-verify` in justfile (grep empty).
2. Local: `docs/fixtures.md` exists and already names `manifest.txt → devilGo-1spo` as consumer (fixtures.md:15); probe pinned to dX commit ca91d778f; xcdz blocked_by 1spo (`devilGo-xcdz` frontmatter) so this bean gates the harness.
3. Online: SKIPPED — storage decision already recorded (committed goldens, CI verify-only; CI can't fetch retail MPQs).
4. Scope-fit: owns location+manifest+recipes; overlap seam: probe already emits manifest.txt while bean AC says "manifest format…committed" — clarify 1spo verifies/commits probe manifest rather than defining a second format; zstd note vs docs/fixtures.md format notes: no conflict (fixtures stored raw).
5. DDD: engine/testsupport + repo tooling; no diablo-layer coupling.
6. Dep trigger: blocked_by vsx3 = completed → released; but 1spo is draft while dependent xcdz is todo — inversion; promote 1spo to todo before xcdz dispatch (ledger said all impl tasks todo; this one missed).
7. AC gaps: all 4 ACs still open (layout, manifest commit, just recipes, docs/fixtures.md auto-update) — none Done; AC4 vs existing fixtures.md artifact table = additive edit, fine.
8. Validation: `just fixture-refresh` regenerates + reports drift; `just fixture-verify` hash-only green in CI; committed manifest matches `sha256sum` of test/fixtures tree.
9. Related: vsx3 (blocker, done), xcdz/7aul/uc44/kx0l/m2x7/6xu2-oqho (consumers), parent 1n0a.
VERDICT: TRIAGED-OK:devilGo-1spo (promote draft→todo; xcdz depends on it)

**devilGo-6xu2** — Persistence epic (.sv/SQLite/pfile)
1. Root: prior triage found body contradicts recorded save decisions; live evidence the fix was NEVER applied: goal line 20 still "Read AND write vanilla-compatible hero saves", scope line 26 still `LegacySVStore (.sv import/export)`, AC line 32 still ".sv export readable by devilutionX", dep line 41 still "dg-epic-02 (MPQ writer)".
2. Local: decisions at i56j `DECIDED #1` (SQLite-only live, one-time importer) + SAVE ARCHITECTURE AMENDMENT (per-hero .db + global stash); children already aligned: xwtt:36,57-59 (per-hero db, importer CLI), w67z completed (codec/sha) — only the epic body lags.
3. Online: SK

Wave outputs: /tmp/dg-triage/out-b1.md
