//go:build windowed

package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"sync"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	snd "github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
	"github.com/bmcszk/devilutionGo/internal/engine"
)

// audio.go is the DEVICE half of the sound domain: it owns the one ebiten audio
// context and the bounded player pool (gbSndInited, Source/engine/sound.cpp:
// 254-262). The DECISIONS (volume, pan, audibility) live in the sim-pure
// internal/diablo/gamedata/audio package; nothing here is imported by the sim.
//
// Disclosed resource limits, not fidelity claims: upstream has no pool cap (the
// mixer queues every voice), so sfxPoolCap is our bound; and the per-channel
// pan is applied to the decoded PCM because ebiten players carry one gain.

// sfxPoolCap bounds the live sfx players. Upstream's mixer queues voices
// without a cap; this is the port's resource limit (disclosed).
const sfxPoolCap = 8

// musicTrackNone is the "no track loaded" sentinel (NUM_MUSIC, sound.cpp:340).
const musicTrackNone = ""

// sfxPlayer is the slice of an ebiten audio player the engine drives, so the
// pool and the volume seams are testable without an audio device.
type sfxPlayer interface {
	Play()
	SetVolume(volume float64)
	IsPlaying() bool
	Close() error
}

// playerFactory builds a player from decoded 16-bit stereo PCM at the given
// gain; the tests script it, production wires the ebiten context.
type playerFactory func(pcm []byte, gain float64) (sfxPlayer, error)

// musicLoader builds a looping music player from container bytes; the tests
// script it.
type musicLoader func(data []byte, gain float64) (sfxPlayer, error)

// audioEngine is the sound device: one context, one bounded sfx pool, at most
// one music player. A zero engine is "not initialised" (gbSndInited false), and
// every method tolerates that.
type audioEngine struct {
	ctx *ebitenaudio.Context

	catalog *snd.SfxCatalog
	// source reads track containers; the boot wires the archive stack.
	source snd.SfxSource
	// sampleRate is the ini Audio/"Sample Rate" (default 22050,
	// Source/options.cpp:55): the decoding rate for every container.
	sampleRate int
	newSfx     playerFactory
	newMusic   musicLoader

	// soundVolume/musicVolume are the log volumes ([-1600..0]), the same values
	// the ini carries (Source/options.cpp:514-516).
	soundVolume int
	musicVolume int

	// soundOn/musicOn are gbSoundOn/gbMusicOn (sound.cpp:257-262): volume at
	// VOLUME_MIN switches the family off.
	soundOn bool
	musicOn bool

	// playerTile reads the player's tile so PlaySfxLoc can attenuate; nil means
	// no positional source, so a positional sound is dropped.
	playerTile func() snd.Point

	// mu guards the pool and the music player: the bus subscriber runs on its
	// own goroutine while the frame loop plays.
	mu      sync.Mutex
	players []sfxPlayer
	music   sfxPlayer
	track   string
}

// newAudioEngine builds an engine over an injected context and player factories.
// A nil context or nil factories mean "no device": Initialised reports false and
// every play is a no-op.
func newAudioEngine(
	ctx *ebitenaudio.Context, catalog *snd.SfxCatalog,
	newSfx playerFactory, newMusic musicLoader,
) *audioEngine {
	sampleRate := engine.DefaultAudioSampleRate
	if ctx != nil {
		sampleRate = ctx.SampleRate()
	}

	return &audioEngine{
		ctx: ctx, catalog: catalog, sampleRate: sampleRate,
		newSfx: newSfx, newMusic: newMusic,
		soundVolume: engine.VolumeMax, musicVolume: engine.VolumeMax,
		soundOn: true, musicOn: true,
	}
}

// newDeviceAudioEngine opens the ONE audio context at the ini sample rate and
// wires the ebiten decoders (wav only: the probe measured every retail record as
// linear PCM, so no other decoder is needed — see probe_test.go).
func newDeviceAudioEngine(sampleRate int, catalog *snd.SfxCatalog) *audioEngine {
	ctx := ebitenaudio.NewContext(sampleRate)

	return newAudioEngine(ctx, catalog, ebitenSfxFactory(ctx), ebitenMusicLoader(ctx))
}

// Initialised reports gbSndInited: the Music/Sound rows arm only when this is
// true (X:gamemenu.cpp:126-134).
func (e *audioEngine) Initialised() bool {
	return e != nil && e.newSfx != nil
}

// SoundVolume reports the live log sound volume (sound_get_or_set_sound_volume's
// read arm, Source/engine/sound.cpp:396-404).
func (e *audioEngine) SoundVolume() int { return e.soundVolume }

// MusicVolume reports the live log music volume (Source/engine/sound.cpp:383-394).
func (e *audioEngine) MusicVolume() int { return e.musicVolume }

// SetSoundVolume stores a log sound volume and derives gbSoundOn from it
// (snd_init, Source/engine/sound.cpp:254-262).
func (e *audioEngine) SetSoundVolume(volume int) {
	e.soundVolume = engine.CapVolume(volume)
	e.soundOn = e.soundVolume > engine.VolumeMin

	if !e.soundOn {
		e.StopSounds()
	}
}

// SetMusicVolume stores a log music volume, derives gbMusicOn and re-applies the
// gain to the loaded track (sound_get_or_set_music_volume, Source/engine/
// sound.cpp:383-394).
func (e *audioEngine) SetMusicVolume(volume int) {
	e.musicVolume = engine.CapVolume(volume)
	e.musicOn = e.musicVolume > engine.VolumeMin

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.music == nil {
		return
	}

	if !e.musicOn {
		e.music.Close()
		e.music = nil
		e.track = musicTrackNone

		return
	}

	e.music.SetVolume(float64(linearVolume(e.musicVolume, engine.VolumeMin, engine.VolumeMax)))
}

// SoundOn reports gbSoundOn.
func (e *audioEngine) SoundOn() bool { return e.soundOn }

// MusicOn reports gbMusicOn.
func (e *audioEngine) MusicOn() bool { return e.musicOn }

// LivePlayers reports the players still in the pool: the pool-cap oracle.
func (e *audioEngine) LivePlayers() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	return len(e.players)
}

// Track reports the loaded music track ("" when none).
func (e *audioEngine) Track() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.track
}

// PlaySFX plays a catalog effect at the ini sound volume
// (PlaySFX -> PlaySfxPriv, Source/effects.cpp:69-98): the guards are "not
// initialised", "sound off", then the pool.
func (e *audioEngine) PlaySFX(id snd.SfxID) error {
	return e.PlaySfxAt(id, engine.AttenuationMax, 0)
}

// PlaySfxLoc plays a catalog effect attenuated by its tile distance from the
// player (PlaySfxLoc, Source/effects.cpp:208-232 -> sound_position.cpp:8-22). A
// sound past the attenuation floor is dropped, not an error.
func (e *audioEngine) PlaySfxLoc(id snd.SfxID, at snd.Point) error {
	if e.playerTile == nil {
		return nil
	}

	volume, pan, audible := snd.CalculateSoundPosition(e.playerTile(), at)
	if !audible {
		return nil
	}

	return e.PlaySfxAt(id, volume, pan)
}

// PlaySfxAt plays a catalog effect at a log volume and pan. The positional
// volume is folded into the user's sound volume exactly as upstream does
// (SoundSample::SetVolume, Source/utils/soundsample.h:83): the attenuation
// range scales the slider range.
func (e *audioEngine) PlaySfxAt(id snd.SfxID, logVolume, pan int) error {
	if !e.Initialised() || !e.soundOn || e.catalog == nil {
		return nil
	}

	sample, err := e.catalog.Load(id)
	if err != nil {
		return fmt.Errorf("load sfx %q: %w", id, err)
	}

	pcm, err := decodeWave(e.sampleRate, sample.Data)
	if err != nil {
		return fmt.Errorf("decode sfx %q: %w", id, err)
	}

	left, right := snd.PanLogToLeftRight(pan)

	// The gain spans the ATTENUATION range, not the slider range
	// (SoundSample::SetVolume(..., ATTENUATION_MIN, 0), soundsample.h:83).
	gain := float64(snd.VolumeLogToLinear(e.attenuatedVolume(logVolume), snd.AttenuationMin, snd.AttenuationMax))

	return e.startPlayer(mixPCM(pcm, left, right), gain)
}

// StopSounds closes every pooled sfx player (sound_stop, Source/engine/
// sound.cpp).
func (e *audioEngine) StopSounds() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, player := range e.players {
		player.Close()
	}

	e.players = nil
}

// StartMusic loads and loops a track at the ini music volume (music_start,
// Source/engine/sound.cpp:341-372): no crossfade, no fade-in — a fade would be
// a deviation. A track that cannot be read leaves music stopped.
func (e *audioEngine) StartMusic(track string, data []byte) error {
	if !e.Initialised() || e.newMusic == nil {
		return nil
	}

	e.StopMusic()

	if !e.musicOn || len(data) == 0 {
		return nil
	}

	gain := float64(linearVolume(e.musicVolume, engine.VolumeMin, engine.VolumeMax))

	player, err := e.newMusic(data, gain)
	if err != nil {
		return fmt.Errorf("start music %q: %w", track, err)
	}

	e.mu.Lock()
	e.music = player
	e.track = track
	e.mu.Unlock()

	player.Play()

	return nil
}

// StopMusic releases the loaded track and forgets it (music_stop,
// Source/engine/sound.cpp:333-337).
func (e *audioEngine) StopMusic() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.music != nil {
		e.music.Close()
	}

	e.music = nil
	e.track = musicTrackNone
}

// attenuatedVolume folds a positional attenuation into the user's sound volume
// exactly as upstream does (SoundSample::SetVolume, Source/utils/soundsample.h:
// 83): logSoundVolume + logUserVolume * (ATTENUATION_MIN / VOLUME_MIN). The
// result is NOT clamped to the slider range — it may fall below VOLUME_MIN,
// which is why the gain curve spans ATTENUATION_MIN..ATTENUATION_MAX.
func (e *audioEngine) attenuatedVolume(logVolume int) int {
	return logVolume + e.soundVolume*snd.AttenuationMin/engine.VolumeMin
}

// startPlayer releases finished voices, refuses past the cap, then plays.
func (e *audioEngine) startPlayer(pcm []byte, gain float64) error {
	player, err := e.newSfx(pcm, gain)
	if err != nil {
		return fmt.Errorf("create sfx player: %w", err)
	}

	e.mu.Lock()

	if !e.admitLocked(player) {
		e.mu.Unlock()
		player.Close()

		return nil
	}

	e.players = append(e.players, player)
	e.mu.Unlock()

	player.Play()

	return nil
}

// admitLocked drops finished players and reports whether a new one fits.
func (e *audioEngine) admitLocked(player sfxPlayer) bool {
	live := e.players[:0]

	for _, existing := range e.players {
		if existing.IsPlaying() {
			live = append(live, existing)

			continue
		}

		existing.Close()
	}

	e.players = live

	return len(e.players) < sfxPoolCap && player != nil
}

// linearVolume converts a log volume into the linear gain in [0,1].
func linearVolume(logVolume, logMin, logMax int) float32 {
	return snd.VolumeLogToLinear(logVolume, logMin, logMax)
}

// decodeWave decodes a retail WAV container into 16-bit stereo PCM at
// sampleRate. The probe measured every retail record as linear PCM, single or
// stereo (probe_test.go), which is exactly what ebiten's decoder takes.
func decodeWave(sampleRate int, data []byte) ([]byte, error) {
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("wav decode: %w", err)
	}

	pcm, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("wav read: %w", err)
	}

	return pcm, nil
}

// mixPCM applies per-channel gains to 16-bit little-endian stereo PCM
// (PanLogToLeftRight's output; ebiten players carry a single gain).
func mixPCM(pcm []byte, leftGain, rightGain float32) []byte {
	if leftGain == 1 && rightGain == 1 {
		return pcm
	}

	mixed := make([]byte, len(pcm))

	for i := 0; i+3 < len(pcm); i += 4 {
		mixed[i], mixed[i+1] = scaleSample(pcm[i:i+2], leftGain)
		mixed[i+2], mixed[i+3] = scaleSample(pcm[i+2:i+4], rightGain)
	}

	return mixed
}

// scaleSample scales one little-endian 16-bit sample by gain.
func scaleSample(sample []byte, gain float32) (byte, byte) {
	value := int16(uint16(sample[0]) | uint16(sample[1])<<8)
	scaled := int16(float32(value) * gain)

	return byte(uint16(scaled)), byte(uint16(scaled) >> 8)
}

// ebitenSfxFactory is the production player factory: one player per play
// (ebiten audio skill: never reuse a played Player).
func ebitenSfxFactory(ctx *ebitenaudio.Context) playerFactory {
	return func(pcm []byte, gain float64) (sfxPlayer, error) {
		player := ctx.NewPlayerFromBytes(pcm)
		player.SetVolume(gain)

		return player, nil
	}
}

// ebitenMusicLoader is the production music loader: an infinite loop over the
// decoded stream, which is how music_start loops (numIterations 0,
// Source/engine/sound.cpp:363).
func ebitenMusicLoader(ctx *ebitenaudio.Context) musicLoader {
	return func(data []byte, gain float64) (sfxPlayer, error) {
		stream, err := wav.DecodeWithSampleRate(ctx.SampleRate(), bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("music wav decode: %w", err)
		}

		player, err := ctx.NewPlayer(ebitenaudio.NewInfiniteLoop(stream, stream.Length()))
		if err != nil {
			return nil, fmt.Errorf("music player: %w", err)
		}

		player.SetVolume(gain)

		return player, nil
	}
}

// effectsSource is the archive path of the effects table
// (effects.cpp:154 reads "assets/txtdata/sound/effects.tsv").
const effectsPath = "assets/txtdata/sound/effects.tsv"

// loadSfxCatalog builds the effects catalog over the asset stack. A missing or
// broken corpus is reported and yields no catalog: silence, not a dead window.
func loadSfxCatalog(source snd.SfxSource) (*snd.SfxCatalog, error) {
	data, err := source.ReadFile(effectsPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", effectsPath, err)
	}

	effects, err := snd.ParseEffects(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", effectsPath, err)
	}

	catalog, err := snd.NewSfxCatalog(effects, source)
	if err != nil {
		return nil, fmt.Errorf("build sfx catalog: %w", err)
	}

	return catalog, nil
}

// startTrack plays a level track through the engine's own asset source
// (music_start(GetLevelMusic(leveltype)), Source/engine/sound.cpp:341-372).
func (e *audioEngine) startTrack(track string) {
	if e.source == nil {
		return
	}

	e.startMusicFromCatalog(e.source, track)
}

// startMusicFromCatalog loads a track's container bytes and plays it, stopping
// the music when the archive has no such track.
func (e *audioEngine) startMusicFromCatalog(source snd.SfxSource, track string) {
	data, err := source.ReadFile(snd.ArchivePath(track))
	if err != nil {
		slog.Warn("windowed boot: music track unavailable", "track", track, "err", err)
		e.StopMusic()

		return
	}

	if err := e.StartMusic(track, data); err != nil {
		slog.Warn("windowed boot: music start", "track", track, "err", err)
	}
}

// townMusicTrack is the town's level track (MusicTracks[TMUSIC_TOWN],
// Source/engine/sound.cpp:158-167). The wave's only session is the town: the
// dungeon tracks need GetLevelMusic driven by the live level type, which is a
// recorded deferral of the audio bean, not a silent substitution.
const townMusicTrack = `music\dtowne.wav`
