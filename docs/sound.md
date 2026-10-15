# SOUND ENGINE 1:1 SPEC FOR EBITEN PORT

## SN1. ARCHITECTURE MAP — C++ → Go package layout

| C++ File/Class | Lines | Purpose | Go Package / Type | Notes |
|---|---|---|---|---|
| Source/engine/sound.cpp | 413 | Engine pipeline init, music, volumes, hooks | `internal/audio/engine.go` | `snd_init/deinit`, `music_start/stop`, `snd_play_snd` |
| Source/engine/sound.h | 73 | Public API + `_music_id`, `TSnd` | `internal/audio/types.go` | `TSnd { start_tc, DSB SoundSample }` |
| Source/engine/sound_position.cpp | 18 | 3D position → volume/pan | `internal/audio/position.go` | `CalculateSoundPosition` |
| Source/engine/sound_defs.hpp | 24 | Constants: VOLUME_MIN=-1600..0, ATTENUATION_MIN=-6400, PAN_MIN=-6400..6400, VolumeScale=3321.9281, StereoSeparation=6000 | `internal/audio/constants.go` | Exact port required |
| Source/utils/soundsample.h/cpp | 300+ | `SoundSample` wrapper over Aulib/SDL_mixer, log→linear volume, pan L/R gains, duplicate pool | `internal/audio/sample.go` | Core math here |
| Source/effects.cpp | 7.9k | `PlaySfxPriv`, `PlaySfxLoc`, `RndSFX`, `sgSFX[814+15]` catalog, flags | `internal/sfx/catalog.go` | Hook from gameplay |

**Go package layout:**
```
internal/audio/
  engine.go      // snd_init, music, volumes, options callbacks
  sample.go      // SoundSample, VolumeLogToLinear, PanLogToLeftRight, duplicate pool
  position.go    // CalculateSoundPosition
  constants.go   // all #defines verbatim
  music.go       // streaming music, track tables
  context.go     // ebiten audio.Context singleton, sample-rate mgmt
internal/sfx/
  catalog.go     // effects.tsv → sgSFX table, lazy load, flags
  hooks.go       // PlaySfxPriv, PlaySfxLoc, RndSFX
```

## SN2. VOLUME/PAN MATH — BYTE-EXACT PORT

### Constants (sound_defs.hpp:9-17)
```go
const (
    VOLUME_MIN      = -1600
    VOLUME_MAX      = 0
    VOLUME_STEPS    = 64
    ATTENUATION_MIN = -6400
    ATTENUATION_MAX = 0
    PAN_MIN         = -6400
    PAN_MAX         = 6400
)
```

### soundsample.cpp: VolumeLogToLinear (lines 63-82)
```go
// VolumeScale = -1000 / ln(0.5) = 3321.9281
const VolumeScale = 3321.9281
const MillibelMin = -10000.0
const MillibelMax = 0.0
const LogBase = 10.0

// logVolume in [logMin..logMax], where logMin = ATTENUATION_MIN (-6400) for game sounds,
// VOLUME_MIN (-1600) for sliders; logMax = 0.
func VolumeLogToLinear(logVolume, logMin, logMax int) float32 {
    logScaled := math.Remap(float32(logMin), float32(logMax), MillibelMin, MillibelMax, float32(logVolume))
    return float32(math.Pow(LogBase, logScaled/VolumeScale))
}
```

### soundsample.cpp: PanLogToLeftRight (lines 84-108) — SDL3 path
```go
const StereoSeparation = 6000.0

func PanLogToLeftRight(logPan int) (leftGain, rightGain float32) {
    if logPan == 0 {
        return 1.0, 1.0
    }
    factor := float32(math.Pow(LogBase, -float32(abs(logPan))/StereoSeparation))
    if logPan < 0 {
        return 1.0, factor  // sound left → attenuate right
    }
    return factor, 1.0      // sound right → attenuate left
}
```

### sound_position.cpp: CalculateSoundPosition (lines 7-26)
```go
// pan = clamp((deltaX - deltaY) * 256, PAN_MIN, PAN_MAX)
// volume = -ApproxDistance * 64
// if volume <= ATTENUATION_MIN (-6400) → return false (don't play)
// *plPan = pan; *plVolume = volume
```

**Precision:** C++ uses `float` for VolumeScale/logScaled math → Go must use `float32` in these formulas. One worked example per function required in tests.

## SN3. SFX PLAYBACK SEMANTICS — 1:1 GATES

`snd_play_snd` (sound.cpp:205-250) dedup & play:
1. `if pSnd == nil || !gbSoundOn { return }`
2. **80 ms dedup**: `if SDL_GetTicks() - pSnd.start_tc < 80 { return }`
3. `if sound.IsPlaying() { sound = DuplicateSound(sound); if sound == nil { return } }`
4. `sound.PlayWithVolumeAndPan(lVolume, userVolume, lPan)`
5. `pSnd.start_tc = now`

`PlaySfxPriv` (effects.cpp:69-101) gates **in order**:
1. `if MyPlayer.pLvlLoad != 0 && gbIsMultiplayer { return }` (loading in MP)
2. `if !gbSndInited || !gbSoundOn || gbBufferMsgs != 0 { return }`
3. **Non-stream already playing**: `if !(flags & (sfx_STREAM|sfx_MISC)) && pSnd != nil && pSnd.isPlaying() { return }`
4. **Streaming**: `if flags & sfx_STREAM { StreamPlay(...); return }`
5. **Lazy load**: `if pSnd == nil { pSnd = sound_file_load(name) }`
6. **Accessibility cues volume**: `if id in [Weapon..Interact] { userVolume = audioCuesVolume } else { userVolume = soundVolume }`
7. `snd_play_snd(pSnd, lVolume, lPan, userVolume)`

`RndSFX` class variants (effects.cpp:112-128): Warrior69/Sorceror69/Rogue69 → random pitch variant.

## SN4. STREAMING + MUSIC

`music_start` (sound.cpp:338-368):
- Track tables (sound.cpp:130-151): `MusicTracks[8]` vs `SpawnMusicTracks[8]` — verbatim strings
- `stream = true` (unless `DISABLE_STREAMING_MUSIC`)
- `LoadAudioFile(path, stream, music)` → `music.SetVolume(musicVol, VOLUME_MIN, VOLUME_MAX)`
- `music.Play(0)` → `numIterations=0` = infinite loop
- `sgnMusicTrack = nTrack`

`GetLevelMusic(dungeon_type)` mapping (sound.cpp:324-330):
```go
Town→TMUSIC_TOWN, Cathedral→CATHEDRAL, Catacombs→CATACOMBS, Caves→CAVES, Hell→HELL, Nest→NEST, Crypt→CRYPT
```

`GetMp3Path` (sound.cpp:59-66): If file doesn't exist as WAV, try `.mp3` (for devilutionx.mpq overrides).

## SN5. CATALOG — effects.tsv LOADER

- Source/tables/effects.tsv: 814 + 15 unused = 829 rows (Volume II V9)
- Columns: `name`, `pszName` (file), `bFlags` (bitmask), `volume`, `pan`, etc.
- Flags from sound_defs.hpp: `sfx_STREAM=0x01`, `sfx_MISC=0x02`, `sfx_LIMITED=0x04`, etc.
- `sgSFX` array indexed by `SfxID` enum (auto-assigned from row order, `mappingId` in itemdat.cpp:585-600 pattern)

## SN6. AUDIO OPTIONS — RUNTIME RE-INIT CONSTRAINT

Options in Source/options.cpp + options.h (Audio section):
- `sampleRate` default **22050** (Diablo retail WAV rate) — **use same**
- `channels` default **2** (stereo output)
- `bufferSize` default **2048** samples
- `resamplingQuality` 0..4 (SDL_mixer)
- `resampler` string (default "linear")
- `device` string (empty = default)

**Critical constraint:** `ebiten.NewContext(sampleRate)` **can only be called once per process**. C++ `OptionAudioChanged()` callbacks re-init the audio subsystem. Go options:
- **A)** Mute all volumes (global multiplier 0) instead of re-init
- **B)** Full process restart (documented limitation)
- **C)** Recreate all players on new context (ebiten doesn't expose context swap)

**Decision:** Document Option A as chosen; Option B for sample-rate changes only.

## SN7. EBITEN CONSTRAINTS + DECISIONS

| C++ Feature | Ebiten Reality | Decision |
|---|---|---|
| Native pan | **None** | Port `StereoPanStream` from ebitengine/examples/audiopanning (verified online) |
| Global master volume | **None** (per-player only) | Maintain `masterVol float32` multiplier on all `SetVolume` calls |
| F32 stereo stream | `audio.Player` expects F32 LE stereo | Convert retail 8-bit u8 PCM WAV → F32 at load (decoder) |
| Audio context | One per process | `audio.NewContext(22050)` at init; sample-rate change = process restart |
| Music streaming | `io.Reader` + `NewInfiniteLoop` for loop | WAV → `audio.Decode`; MP3 → `go-mp3.Decode`; both feed `NewInfiniteLoop` |
| DuplicateSound pool | Aulib clones stream | `DuplicateSound` → `NewPlayerFromBytes` shared bytes (zero-copy) |
| Decoder formats | WAV/FLAC/OGG/MP3 | **Retail only WAV** (diabdat.mpq); MP3 optional via devilutionx.mpq; FLAC/OGG not needed |

## SN8. PARITY TESTS (TDD-FIRST)

| Test | Method | Acceptance |
|---|---|---|
| `TestVolumeLogToLinear` | Golden table: 100-point sweep logVolume∈[-6400,0] vs C++ probe | `float32` values bit-exact |
| `TestPanLogToLeftRight` | Golden table: pan∈[-6400,6400] step 100 | L/R gains bit-exact |
| `TestCalculateSoundPosition` | 8 directions × 10 distances vs C++ outputs | pan/volume int exact |
| `TestSfxDedup` | Mock clock; fire same sfx at t=0, t=40 (drop), t=80 (play) | 80 ms boundary exact |
| `TestCatalogRowCount` | effects.tsv rows = 829 (incl 15 unused) | Row count + first/last name |
| `TestMusicTrackTables` | MusicTracks/SpawnMusicTracks arrays verbatim | String equality |
| `TestMp3Fallback` | Missing WAV → MP3 path resolution | File existence order |

**Limitation:** Ebiten cannot capture audio output for waveform diff. Parity = API + math + catalog level only.

## SN9. EFFORT & RISKS

| Item | Effort | Risk |
|---|---|---|
| Audio context + constants + math | S | None (pure port) |
| StereoPanStream (mono→stereo upsample) | M | Pan law must match C++ `StereoSeparation=6000` tuning |
| Sample wrapper + duplicate pool | M | GC pressure from many short-lived players; use `NewPlayerFromBytes` shared bytes |
| SFX catalog loader (829 rows) | S | Row-count parity gate |
| Music streaming + MP3 fallback | M | `go-mp3` decode vs `audio.Decode` sample-rate alignment |
| Options volume mapping | S | Global multiplier pattern |

**Critical open question:** Does Diablo retail MPQ contain any Huffman/ADPCM compressed WAV? mpqfs reads them; suprsokr/go-mpq **does not support** Huffman (0x01) or ADPCM (0x40/0x80). If present → must port C++ `encrypt.cpp` decode or pre-extract assets. Check `diabdat.mpq` audio files.

---

**Online-verified Ebiten refs:**
- audiopanning example: https://github.com/hajimehoshi/ebiten/blob/v2.9.2/examples/audiopanning/main.go
- audio package: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/audio
- wav decode: https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/audio/wav
- mp3 decode: https://pkg.go.dev/github.com/hajimehoshi/go-mp3
- resound (optional): https://github.com/SolarLune/resound

**All constants and formulas trace to C++ file:line citations above.**
