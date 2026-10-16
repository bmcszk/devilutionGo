---
# devilGo-vof6
title: MPQ archive stack over go-mpq
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - mpq
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T20:39:19Z
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

## Acceptance criteria (fix2 carve-out)
- [ ] PKWARE/ADPCM codec gap recorded as explicit deliverable (port from mpqfs C); go-mpq used only for what it supports
- [ ] Extract-at-build v1 scope; runtime MPQ explicitly out of scope post-1.0
- [ ] Smoke test extracts every file of user-provided spawn/diabdat/hellfire (skips politely when absent, like test/timedemo_test.cpp:47-49)
- [ ] Priority override unit test with two synthetic archives
- [ ] Unpacked-dir mode reads same names with forward/backslash normalization


## Libraries
- github.com/suprsokr/go-mpq — MPQ v1 reader for build-time extraction; PKWARE/ADPCM gaps filled by our port (see go-libraries.md gap list)

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — io.Reader abstractions and error wrapping

## Dependencies & blockers
dg-task-001.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: RESEARCH-FIRST: go-mpq stream vs buffer handling verified against its API at claim; bytes.Buffer wrapper only if needed (no speculative LRU).
- go-mpq returns io streams vs buffers — wrap with bytes.Buffer caching LRU (size cap).


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:external-gap-unrecorded — go-mpq codec limits (verified online) make the extract-everything AC false as written; scope must be reconciled with the recorded extract-at-build decision.


## IMPL NOTES (triage 2026-10-16)
1 root: wave-1 task (dg-task-006): OpenAsset priority semantics, unpacked dirs + retail archives via go-mpq; draft, blocked_by devilGo-lfvi (wired ✓).
2 local: dX cites verified — assets.cpp:385-400 LoadMPQ/FindMPQ ✓, :505-516 LoadCoreArchives ✓, :568-590 hf archives ✓, :704-760 LoadModArchives ✓; assets.hpp:307-311 priorities (1000/9000/…) ✓; `Source/mpq/mpq_reader.cpp` exists (bean omits `mpq/` dir — trivial) ✓; test/timedemo_test.cpp:47-49 polite-skip ✓. Milestone DECIDED #2: v1 = build-time extraction to loose assets; runtime MPQ reader deferred post-1.0.
3 online: pkg.go.dev/github.com/suprsokr/go-mpq fetched (v0.3.1, 2026-02-09): Open/HasFile/ExtractFile(mpqPath, destPath) as claimed, BUT codec table: PKWare DCL(0x08) read ✅, **Huffman(0x01) ❌, ADPCM mono(0x40)/stereo(0x80) ❌ — "Audio files only, not implemented"**; ExtractFile writes to disk path, not bytes (bean's "adapt to bytes" acknowledges).
4 solution: record the codec gap in the bean — the AC "extracts every file of spawn/diabdat/hellfire" is unachievable with go-mpq alone because Diablo audio uses Huffman+ADPCM (devilutionGo AGENTS.md `assets/extract.sh` already plans custom PKWARE/Huffman/ADPCM, so the gap has an owner — this bean or a child must name it). Also reconcile scope with DECIDED #2: unpacked-dir mode is the v1 requirement; retail-archive reading is the deferred post-1.0 capability — scope/AC should say which parts are v1.
5 understanding: engine-layer asset source; in v1 primarily serves build-time extraction + loose-asset lookup with devilutionX priority semantics.
6 repro: `sed -n '385,400p' dX/Source/engine/assets.cpp`; keenable_fetch https://pkg.go.dev/github.com/suprsokr/go-mpq → "Compression Support" table.
7 fix: body edit — add research bullet: "go-mpq v0.3.1: DCL read OK; Huffman/ADPCM ❌ → custom codecs required for audio (extract.sh plan); ExtractFile→disk, wrap for bytes"; adjust smoke-test AC to "every file except Huffman/ADPCM-compressed audio (skipped politely) until custom codecs land (child bean)"; mark retail-MPX read path as post-1.0 per DECIDED #2.
8 validation: bean body mentions Huffman/ADPCM gap + DECIDED #2; AC no longer promises all-files extraction unconditionally.
9 related: lfvi (scaffold), n89r/zvwd/jx08 (blocked by it), l5fp (manifest consumer).
VERDICT: NEEDS-WORK:external-gap-unrecorded — go-mpq codec limits (verified online) make the extract-everything AC false as written; scope must be reconciled with the recorded extract-at-build decision.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.


