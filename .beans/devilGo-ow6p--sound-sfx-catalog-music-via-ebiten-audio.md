---
# devilGo-ow6p
title: 'Sound: sfx catalog + music via ebiten audio'
status: in-progress
type: task
priority: normal
tags:
    - plan:task
    - phase-1
    - audio
created_at: 2026-10-15T16:35:16Z
updated_at: 2026-11-26T21:22:56Z
parent: devilGo-5wsb
---

# Sound: sfx catalog + music via ebiten audio

## Goal
ONE audio.Context at the ini sample rate (default 22050) behind the windowed tag; PlaySFX/PlaySfxLoc play catalog bytes through a bounded player pool with positional volume/pan from the ported integer math (CalculateSoundPosition, VolumeLogToLinear, PanLogToLeftRight); the FIRST kernel bus subscriber plays only sound-bearing events. Claimed as the Options-slice wave anchor (triage: .hermes/lane/options-slice/triage.md).

## Research
- Positional math upstream: sound_position.cpp:8-22 (pan=(dx-dy)*256 clamped, volume=dist*-64, drop at <=ATTENUATION_MIN=-6400) - verified exact by the combined triage (E2).
- Volume curve: VolumeLogToLinear (utils/soundsample.cpp:69-74), remap to millibels then pow(10, mb/3321.9281); PanLogToLeftRight :83-102. Port float32 with committed golden at the 64 slider steps.
- E-A1 (the gating risk): NO retail sound byte has ever been decoded in Go; the catalog hands container bytes only (SfxSample.Data, SampleRate is a constant stamp); no WAV fixture committed; go-libraries.md records go-mpq cannot decompress Huffman/ADPCM and asserts retail=PCM without a cited measurement. The wave's S3 probe test reads sfx/misc/walk1.wav through MPQStack.ReadFile under DG_ASSET_ARCHIVES, parses the RIFF fmt chunk, asserts PCM. If it fails, STOP - record measured format, re-triage the decode gap.
- Bus: kernel.SetPublisher + bus.New exist; no subscriber registered in-tree yet (E1) - this wave adds the first.
- Sim purity: ebiten/audio is banned from internal/game + internal/diablo (simpurity gate; classified rendering). Pure math lives in internal/diablo/gamedata/audio/volume.go (sim-pure, stdlib only); the device engine lives in cmd/devilutionGo/audio.go behind //go:build windowed. ProcessSound stays a sim no-op.

## Acceptance criteria
- [ ] AC3 (audio engine): ONE audio.Context at the ini sample rate behind the windowed tag; PlaySFX/PlaySfxLoc through a bounded pool with positional volume/pan (integer-expectation tests) and VolumeLogToLinear/PanLogToLeftRight (committed golden); the first bus subscriber wires kernel.SetPublisher and plays only sound-bearing events; no device code under internal/. RED: windowed TestAudio + gamedata/audio volume tests.
- [ ] Probe (S3): retail WAV fmt chunk measured and recorded (asset-gated; STOP rule on non-PCM).
- [ ] Pool cap: unit test - N plays, at most cap live players, finished players released (replaces the unobservable 50-concurrent-underrun AC).
- [ ] Music: plays the level track at the ini volume, stops at MIN (no crossfade upstream - a fade would be a deviation).

## Libraries
stdlib + ebiten/v2/audio (+wav, +mp3 - already in go.mod).

## Skills for coder
- ~/.agents/skills/go-dev/SKILL.md
- ~/.agents/skills/go-tests/SKILL.md
- ~/.agents/skills/coding-rules/SKILL.md
- ~/.agents/skills/ponytail/SKILL.md
- ~/.agents/skills/tdd/SKILL.md
- ~/.agents/skills/ebiten/audio/SKILL.md - ONE context, NewPlayerFromBytes per play, InfiniteLoop for music, pool with IsPlaying cleanup; CONTRACT OVERRIDES the skill's 44100 note: 22050 from the ini.
- ~/.agents/skills/ebiten/game-loop/SKILL.md - nothing audio created in Update/Draw.
- ~/.agents/skills/port-parity-testing/SKILL.md - volume math is a 1:1 C++ port.
- ~/.hermes/skills/dev/devilutiongo-lane-ops/SKILL.md

## Triage
Combined triage 2026-11-26 (task-opts-triage-brainer): full evidence + slice order + gate point in .hermes/lane/options-slice/triage.md. Stale gates vof6/b8z9 cleared (both completed). Deferred: RndSFX kernel-RNG randomization (gameplay epic owns RNG), STREAM sfx/ClearDuplicateSounds/cues, menu music cycle (UI-flow epic hook site), positional tuning beyond menu needs. Residuals recorded per AC at wave close.


Claimed 2026-11-26: in-progress as the Options-slice wave anchor (combined triage task-opts-triage-brainer; parent claim, B-U6 protocol).
