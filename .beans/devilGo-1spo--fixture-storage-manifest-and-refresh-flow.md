---
# devilGo-1spo
title: Fixture storage, manifest and refresh flow
status: draft
type: task
priority: high
created_at: 2026-10-16T20:36:59Z
updated_at: 2026-10-16T20:36:59Z
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
