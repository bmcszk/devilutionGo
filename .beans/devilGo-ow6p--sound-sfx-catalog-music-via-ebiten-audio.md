---
# devilGo-ow6p
title: 'Sound: sfx catalog + music via ebiten audio'
status: draft
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - audio
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T14:01:05Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-vof6
---

# Audio

## Goal
PlaySFX/PlaySfxLoc + music tracks at 22050 default with volume/positional attenuation.

## Research
Source/engine/sound.cpp:281 init, :61-64 mp3 fallback, :143-146 track tables; effects.cpp:190-226 sgSFX; sfx catalog txtdata/sound/effects.tsv (~1000 rows); options.cpp:55,520 rates; Ebiten /home/blaze/.agents/skills/ebiten/audio/SKILL.md (Context at file rate, wav/mp3 decode, players pooled).

## Scope
internal/platform/audio: Context(22050), SfxPlayer pool, MusicPlayer with track table + fade; hooks PlaySFX(id)/PlaySfxLoc(id, pos).

## Acceptance criteria
- [ ] 50 concurrent sfx without underrun (test on CI with null output where possible)
- [ ] Track switches with fade < 100ms click
- [ ] Positional attenuation formula matches C++ sound_position.cpp behavior (document + test unit)

## Skills for coder
/home/blaze/.agents/skills/ebiten/audio/SKILL.md — decode + player lifecycle

## Dependencies & blockers
dg-task-006, dg-task-008 (effects.tsv).

## Open questions
- Resample policy: decode at native rate into Context(22050) vs one 44100 context — measure.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wrong-dep — catalog dependency misattributed to zvwd (dg-task-008) instead of b8z9 (dg-task-113); tracker edge to b8z9 missing; minor row-count imprecision.
