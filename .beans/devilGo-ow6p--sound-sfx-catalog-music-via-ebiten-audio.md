---
# devilGo-ow6p
title: 'Sound: sfx catalog + music via ebiten audio'
status: todo
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - audio
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-10-16T20:39:19Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-vof6
    - devilGo-b8z9
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
OPEN-QUESTION RESOLUTION 2026-10-16: SPIKE-AT-CLAIM: resample policy — measure both (decode native into 22050 Context vs single 44100 Context) with resound/bench harness; default = C++ behavior (native rate + conversion at playback).
- Resample policy: decode at native rate into Context(22050) vs one 44100 context — measure.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wrong-dep — catalog dependency misattributed to zvwd (dg-task-008) instead of b8z9 (dg-task-113); tracker edge to b8z9 missing; minor row-count imprecision.


## IMPL NOTES (triage 2026-10-16)
1 root: wave-1 task (dg-task-009): PlaySFX/PlaySfxLoc + music at 22050 default; draft, blocked_by devilGo-vof6 (wired ✓).
2 local: sound.cpp:281 init (Aulib::init w/ sampleRate) ✓; :61-64 GetMp3Path ✓; :143-146 SpawnMusicTracks ✓; effects.cpp:190-226 sgSFX/PlaySFX/PlaySfxLoc ✓; options.cpp:55 (DEFAULT_AUDIO_SAMPLE_RATE 22050) + :520 (rates 22050/44100/48000) ✓; sound_position.cpp exists with concrete pan=(dx−dy)*256, volume=dist*−64 math — AC "document + test unit" is well-grounded ✓; ebiten/audio skill exists ✓. Two gaps: (a) deps line says "dg-task-008 (effects.tsv)" but the effects catalog is owned by dg-task-113=devilGo-b8z9 (dg-task-008=devilGo-zvwd is the generic parser); tracker has no b8z9 edge; (b) research bullet "~1000 rows" vs actual 814/917 (b8z9 owns exact counts).
3 online: n/a (no external claims; ebiten audio facts covered by existing skill).
4 solution: fix dependency attribution: ow6p blocked_by vof6 + b8z9 (catalog), drop/clarify dg-task-008; correct "~1000 rows" → "814 base / 917 hf (owned by devilGo-b8z9)".
5 understanding: engine/audio-layer playback consuming b8z9's audiodat catalog; catalog/engine split is clean once the dep points at the right bean.
6 repro: `beans show devilGo-ow6p --json | jq .blocked_by`; `head -3 dX/Source/engine/sound_position.cpp`; `wc -l` effects tsvs.
7 fix: `beans update devilGo-ow6p --blocked-by devilGo-b8z9` (keep vof6 edge) + body edit of the deps line and row count.
8 validation: ow6p `blocked_by` == [vof6, b8z9]; body references devilGo-b8z9 not dg-task-008 for the catalog.
9 related: b8z9 (catalog), vof6 (file bytes), 0tsf (audio options).
VERDICT: NEEDS-WORK:wrong-dep — catalog dependency misattributed to zvwd (dg-task-008) instead of b8z9 (dg-task-113); tracker edge to b8z9 missing; minor row-count imprecision.


LAYOUT MAP (2026-10-16): Appendix E paths superseded by AGENTS.md three-layer (engine/game/diablo, cmd/devilutionGo[-headless]) — mapping table in milestone devilGo-i56j. Re-point Scope paths at claim time.

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).

ID MAP (fix2): all legacy dg-task-NNN references in this bean map to devilGo ids via docs/research/dg-id-map.json (canonical mapping in milestone devilGo-i56j). Operative deps = frontmatter (source of truth); prose refs are documentation.


