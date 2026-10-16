---
# devilGo-b8z9
title: 'Audio data package: effects catalog (814+15=829) + music tracks + movie list'
status: todo
type: task
priority: normal
tags:
    - phase-7
created_at: 2026-10-15T18:28:59Z
updated_at: 2026-10-16T20:39:19Z
parent: devilGo-5wsb
blocked_by:
    - devilGo-jx08
    - devilGo-778h
---

# Audio: effects catalog, music map, movie list

## Goal
Transcribe the sound domain (`effects.tsv` 814 + `effects-unused.tsv` 15 = 829 base rows) plus
the level->music track map and the complete movie list, and pin `SfxID` to row order.

## Research
- `docs/research/dg-requirements-append.md` **V9.1** (814 rows; loader `Source/effects.cpp:141-165`
  `LoadEffectsData`, `TSFX`/`sgSFX`), **V9.2** (15 unused rows), **V9.3** (music per level/menu),
  **V9.4** (movie list, `Source/movie.cpp` 141 lines), **V12.10** (hf sound deltas), **V12.12**.
- TSVs: `assets/txtdata/sound/effects.tsv` (814), `assets/txtdata/sound/effects-unused.tsv` (15),
  `mods/hf/txtdata/sound/effects.tsv` (**917**: +103 appended, 10 modified),
  `mods/hf/txtdata/sound/effects-unused.tsv` (**1** row — replaced, -14).
- Enums: `SfxID` (`Source/sound_effect_enums.h:117`) — **row order == effects.tsv**, so the enum
  is derivable from the table; `sfx_flag` (`effects.cpp:132-140`): Stream, Misc, Ui, Monk, Rogue,
  Warrior, Sorcerer; `HeroSpeech` = 103 events (V10.7). Cow/quest ids: `Cow1`/`Cow2`
  (`sound_effect_enums.h:310-311`), `YepThatsACowAlright` = `:61`, `NaKrul4`/`NaKrul5`/`NaKrul6`
  (`monster.cpp:931-936`, `:4282-4290`).
- Music: `_music_id` (`Source/engine/sound.h:27-36`) = 8 tracks + `NUM_MUSIC` (TOWN, CATHEDRAL,
  CATACOMBS, CAVES, HELL, NEST, CRYPT, INTRO); `GetLevelMusic(dungeon_type)`
  (`engine/sound.cpp:310-329`, default -> INTRO, called `diablo.cpp:3404`); menu/attract cycle
  `NextTrack()` (`Source/menu.cpp:36-58`): spawn => INTRO always; otherwise
  INTRO -> CATACOMBS -> CAVES -> HELL -> (NEST if hf) -> (CRYPT if hf) -> INTRO;
  `RefreshMusic` `:60-62`; files `music\dlvlf.wav|mp3` etc., mp3-over-wav preference
  (`sound.cpp:61-64`), track tables `sound.cpp:143-146`; hf tracks from the `hfmusic` archive
  (`assets.cpp:570-586`); default rate 22050 Hz (`options.cpp:55,520`).
- Movies: `play_movie` (`movie.h:16`, `movie.cpp:33`, demo-mode early return `:34-36`),
  `PlayInGameMovie` (`movie.h:22`); the 10 `.smk` files in V9.4: `logo.smk` (`diablo.cpp:1320`),
  `Hellfire.smk` (`:1326`, `menu.cpp:98`), `diablo1.smk` (`:1328`, `menu.cpp:100`),
  `fprst3.smk` (`monster.cpp:2892`), `diabvic1/2/3.smk` + `diabend.smk`
  (`monster.cpp:4093/4097/4100/4103`), `loopdend.smk` (`:4113`). Port stance: stills + text by
  default, SMK decode optional (Volume I H11).

## Scope
`internal/gamedata/audiodat/`: `TSFX` table (814/917), `SfxID` enum generated in row order,
`sfx_flag` bitset, `UnusedSfx` table (15 base / 1 hf) as a documented dead set, music track enum
+ `LevelMusic(dungeon_type)` table + `NextTrack` cycle order (spawn/hf variants), movie manifest
(10 entries with trigger site + `userCanClose`) as data. Playback engine stays devilGo-ow6p /
epic-03; only the catalog and mapping live here.

## Acceptance criteria
- [ ] TDD first: row-count parity — Go `Effects` = **814** base / **917** hf (+103 appended, 10
      modified), `EffectsUnused` = **15** base / **1** hf, audio-domain total **829** base rows;
      `SfxID` enum length == table length for both variants (row-order coupling asserted).
- [ ] Every `sfxnr` value referenced by `textdat.tsv` (devilGo-h6wn, 338 rows) and by
      `towners.tsv`/class `sounds.tsv` resolves to a valid `SfxID` in the **base** build and in
      the **hf** build — cross-table referential integrity test.
- [ ] Music worked example: `GetLevelMusic` returns TOWN/CATHEDRAL/CATACOMBS/CAVES/HELL/NEST/
      CRYPT for the 7 dungeon types and INTRO for anything else; `NextTrack()` in spawn mode is a
      self-loop on INTRO, and in retail Diablo the cycle is INTRO -> CATACOMBS -> CAVES -> HELL ->
      INTRO, while Hellfire inserts NEST and CRYPT before wrapping.
- [ ] hf sound deltas by content: the 103 appended rows are the HF monster/quest voices and the
      10 modified rows differ from base in the documented columns only (byte-diff test);
      `NaKrul6` exists only in the hf catalog (assert base lacks it -> V5.4 gating).
- [ ] Movie manifest: exactly 10 entries, each with the cited trigger site and difficulty
      condition (`diabvic1/2/3` per difficulty, `fprst3` on the Skeleton King coffin path); no
      entry invented, and the fallback (stills + text) is the documented default.
- [ ] 1:1 fidelity mandate: every Go value is byte-identical to the TSV/C++ source; any deviation needs a recorded decision in the bean's Proof of Work.

## Libraries
- stdlib only (no third-party deps at this stage)

## Skills for coder
- /home/blaze/.agents/skills/go-dev/SKILL.md — generated enum from data, `go:embed` catalogs
- /home/blaze/.agents/skills/coding-rules/SKILL.md — catalog (data) vs playback (detail) split
- /home/blaze/.agents/skills/beans/SKILL.md — bean/PoW discipline
- /home/blaze/.agents/skills/ebiten/audio/SKILL.md — track/sfx API fit, 22050 Hz default, mp3-vs-wav

## Dependencies & blockers
blocked_by devilGo-jx08. Blocks devilGo-h6wn (text `sfxnr` cross-check), devilGo-778h (monster
voices), devilGo-pu3r (class `sounds.tsv` -> `HeroSpeech`), devilGo-i2cr (audit), and existing
devilGo-ow6p (sound engine) which consumes this catalog.

## Open questions
OPEN-QUESTION RESOLUTION 2026-10-16: RESEARCH-FIRST: generated SfxID enum diffed against sound_effect_enums.h at claim, mismatches recorded. Movie list: data-only here; stills+text playback owned by the movies/UI bean.
- Does generating `SfxID` from row order risk drift with the C++ header (310+ hand-written
  entries)? Compare generated enum names to `sound_effect_enums.h` and record any mismatch.
- Is the movie list data-only here, or does this bean also own the stills+text fallback spec?
  Recommend spec-only, implementation in epic-06.


## Fidelity & TDD gates (program mandate devilGo-i56j — binding)
- [ ] TDD: failing test written FIRST (golden / C++-fixture parity / unit — per domain) before implementation; PoW includes the failing→passing run.
- [ ] 1:1 parity evidence in PoW: state-hash vs C++ fixture, golden file, byte-compat roundtrip, or row-exact data table (pick per domain; no approximations).
- [ ] Any deviation from devilutionX behavior requires a recorded decision entry in this bean + requirements doc — otherwise it is a defect.

TRIAGE 2026-10-16 (pi-as-brainer wave, 93-bean sweep): NEEDS-WORK:wire-deps — 4 of 5 declared `blocks` edges missing from tracker (b8z9→h6wn/778h/pu3r/ow6p); content, citations, row counts, TDD-first AC all verified sound.


## IMPL NOTES (triage 2026-10-16)
1 root: Phase-7 data task (devilGo-b8z9) transcribing effects.tsv (814/917), music map, movie list into `internal/gamedata/audiodat`; draft, blocked_by devilGo-jx08 (wired ✓).
2 local: every spot-check passed — `dX/Source/effects.cpp:141-165` LoadEffectsData ✓ (skip col-1, flags+path cols); `sound_effect_enums.h:117` `enum class SfxID` ✓, Cow1/Cow2 at :310-311 ✓, YepThatsACowAlright at :61 ✓; `engine/sound.h:27-36` 8 tracks+NUM_MUSIC ✓; `sound.cpp:310-329` GetLevelMusic (default→INTRO) ✓; `menu.cpp:36-58` NextTrack spawn/hf cycle ✓ exact; `movie.cpp` = 141 lines ✓; all 8 movie trigger sites verified (diablo.cpp:1320/1326/1328, menu.cpp:98-100, monster.cpp:2892, 4093/4097/4100/4103, 4113, NaKrul4/5/6 at 931-936); TSV counts on disk: base effects 814 data rows / unused 15, hf 917 / 1 ✓ (wc -l 815/16/918/2 incl. headers); V9.1-V9.4/V12.10/V12.12 sections exist (append.md:2309/3168/3186/3198/3733/3793) and V12.12 says effects 814/917 ✓. All 4 skills paths exist ✓.
3 online: n/a (no external claims).
4 solution: wire the 4 declared reverse edges missing from tracker: h6wn(112), 778h(104), pu3r(114), ow6p(009) are NOT blocked_by b8z9 (only i2cr(118) is) — promoting those won't gate on this catalog.
5 understanding: diablo-layer data table (audiodat) feeding ow6p playback engine; clean catalog/engine split, no overlap with ow6p.
6 repro: `beans show devilGo-b8z9 --json | jq .blocked_by`; `beans show devilGo-h6wn --json | jq .blocked_by`; `wc -l` the 4 tsv files above; `sed -n '310,329p' dX/Source/engine/sound.cpp`.
7 fix: `beans update devilGo-b8z9 --blocking devilGo-h6n7 2>/dev/null || true; for t in devilGo-h6wn devilGo-778h devilGo-pu3r devilGo-ow6p; do beans update devilGo-b8z9 --blocking $t; done`
8 validation: re-run `beans show <t> --json` for the four: `blocked_by` must contain devilGo-b8z9.
9 related: jx08 (loader), ow6p (playback), i2cr (audit), h6wn/778h/pu3r (cross-table consumers).
VERDICT: NEEDS-WORK:wire-deps — 4 of 5 declared `blocks` edges missing from tracker (b8z9→h6wn/778h/pu3r/ow6p); content, citations, row counts, TDD-first AC all verified sound.

Fix pass 2026-10-16: Dep-wiring applied (fix pass): 4/5 missing edges wired (h6wn/778h + 2).

Fix pass 2026-10-16: NEEDS-WORK items addressed — verdict TRIAGED-OK (resolved).


