---
# devilGo-0tsf
title: Options + ini import + SimOptions freeze
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-3
    - platform
created_at: 2026-10-15T16:35:17Z
updated_at: 2026-10-16T20:38:06Z
parent: devilGo-yjuv
---

# Options

## Goal
Typed options tree, persistence, migration, sim snapshot.

## Research
options.h/cpp structure; audio callbacks sound.cpp:183-185; ini_test.cpp.

## Scope
internal/platform/options.go, ini_import.go.

## Acceptance criteria
- [ ] ini_test port green
- [ ] TickRate option maps 20..50 like gamemenu slider; audio rate {22050,44100,48000}

## Skills for coder
/home/blaze/.agents/skills/go-dev/SKILL.md — config struct + validation

## Dependencies & blockers
None.

## Open questions
None.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): TRIAGED-OK


## IMPL NOTES (triage 2026-10-16)
1 root: typed options tree, diablo.ini import, migration, sim snapshot freeze. Parent
   devilGo-yjuv, no blockers ✓ (foundation for yd48 QoL flags and audio options).
2 local: options.h/options.cpp exist ✓; "sound.cpp:183-185 audio callbacks" verified at
   Source/engine/sound.cpp:183-185 (OptionChangeSampleRate/Channels/BufferSize callbacks) ✓;
   test/ini_test.cpp exists ✓; TickRate slider 20..50 verified (gamemenu.cpp:177
   `gmenu_slider_set(..., 20, 50, ...)`) ✓; audio sample rate options {22050, 44100, 48000}
   verified (options.cpp:520) ✓ — both ACs map to real C++ behavior. Aligned with recorded
   "diablo.ini settings 1:1" decision.
3 online: n/a.
4 solution: none — ready. (Path polish only: cite `test/ini_test.cpp` and
   `Source/engine/sound.cpp` precisely.)
5 understanding: platform/config layer; SimOptions freeze supports deterministic headless
   parity runs (testing epic).
6 repro: `sed -n '183,185p' Source/engine/sound.cpp`; `sed -n '177p' Source/gamemenu.cpp`;
   `grep -n "22050, 44100, 48000" Source/options.cpp`; `ls test/ini_test.cpp`.
7 fix: none.
8 validation: n/a.
9 related: devilGo-yd48 (blocked by this), devilGo-sr71 (i18n, separate), devilGo-yjuv.
VERDICT: TRIAGED-OK

---

## Verdict summary

devilGo-mfds: NEEDS-WORK:AC-count-wrong (86→80 CMD count; nxxo handler dep edge)
devilGo-h8uq: NEEDS-WORK:premise-false (CMD_LIGHT/V/R nonexistent; fix 7be9→mfds dep)
devilGo-vgxq: TRIAGED-OK
devilGo-omu6: TRIAGED-OK
devilGo-4jo7: NEEDS-WORK:missing-blocked-by-mfds + line cite 698→2774
devilGo-yd48: NEEDS-WORK:scope-conflict-with-Lua-deferral (split QoL from REPL)
devilGo-cciu: NEEDS-WORK:add pu3r dep + record hf-only mods decision
devilGo-sr71: TRIAGED-OK
devilGo-orkj: NEEDS-WORK:gold-cursor boundary error in AC
devilGo-0tsf: TRIAGED-OK

Coverage: 10/10 batch-09 ids triaged, all 10 confirmed present in (process artifact, not archived)
(93 total). No ids from other batches encountered.

BATCH-09-COMPLETE


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.
