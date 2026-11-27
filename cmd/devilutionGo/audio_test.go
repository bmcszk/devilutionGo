//go:build windowed

package main

import (
	"encoding/binary"
	"math"
	"testing"

	snd "github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// The sound device (Options slice S4a). Every test runs WITHOUT an audio
// device: the engine takes an injected player factory, so the pool, the volume
// application and the guards are scriptable.

// fakePlayer is one scripted ebiten player.
type fakePlayer struct {
	playing bool
	volume  float64
	closed  bool
	// finished reports IsPlaying false without Close, like a played-out voice.
	finished bool
}

func (p *fakePlayer) Play() { p.playing = true }

func (p *fakePlayer) SetVolume(volume float64) { p.volume = volume }

func (p *fakePlayer) IsPlaying() bool { return p.playing && !p.finished }

func (p *fakePlayer) Close() error {
	p.closed = true
	p.playing = false

	return nil
}

// scriptedPlayers records every player the engine asked for.
type scriptedPlayers struct {
	created []*fakePlayer
	// pcm and gains keep the exact bytes and gain of each request.
	gains []float64
}

// factory is the playerFactory the engine drives in the tests.
func (s *scriptedPlayers) factory(pcm []byte, gain float64) (sfxPlayer, error) {
	player := &fakePlayer{}
	// Mirror the production factory, which applies the gain to the player
	// (ebitenSfxFactory).
	player.SetVolume(gain)

	s.created = append(s.created, player)
	s.gains = append(s.gains, gain)

	return player, nil
}

// last returns the most recently created player.
func (s *scriptedPlayers) last(t *testing.T) *fakePlayer {
	t.Helper()

	if len(s.created) == 0 {
		t.Fatal("the engine created no player")
	}

	return s.created[len(s.created)-1]
}

func TestAudioEngineInitialisedFollowsTheDevice(t *testing.T) {
	// given an engine with no context and one with a scripted factory
	silent := newAudioEngine(nil, nil, nil, nil)
	scripted := newTestEngine(t, &scriptedPlayers{})

	// when asking each whether it initialised
	// then only the device-backed one reports gbSndInited
	if silent.Initialised() {
		t.Fatal("an engine with no context reports Initialised, want false")
	}

	if !scripted.Initialised() {
		t.Fatal("a device-backed engine reports not Initialised, want true")
	}

	// and a silent engine plays nothing without erroring
	if err := silent.PlaySFX(snd.SfxMenuMove); err != nil {
		t.Fatalf("a silent engine failed a play: %v", err)
	}
}

func TestAudioEnginePoolIsBoundedAndReleasesFinishedPlayers(t *testing.T) {
	// given a device-backed engine
	pool := &scriptedPlayers{}
	audio := newTestEngine(t, pool)

	// when playing far more effects than the cap allows
	for range sfxPoolCap * 4 {
		if err := audio.PlaySFX(snd.SfxMenuMove); err != nil {
			t.Fatalf("PlaySFX: %v", err)
		}
	}

	// then the live pool never exceeds the cap
	if got := audio.LivePlayers(); got > sfxPoolCap {
		t.Fatalf("live players = %d, want <= the cap %d", got, sfxPoolCap)
	}

	// when every playing voice finishes and another effect plays
	for _, player := range pool.created {
		player.finished = true
	}

	if err := audio.PlaySFX(snd.SfxMenuMove); err != nil {
		t.Fatalf("PlaySFX after the pool drained: %v", err)
	}

	// then the finished players were released and only the new one is live
	if got := audio.LivePlayers(); got != 1 {
		t.Fatalf("live players after the drain = %d, want 1", got)
	}
}

func TestAudioEngineVolumeAppliesToTheNextPlays(t *testing.T) {
	for _, tc := range []struct {
		name   string
		volume int
	}{
		{name: "full volume", volume: engine.VolumeMax},
		{name: "half slider", volume: engine.VolumeMin / 2},
		{name: "slider floor", volume: engine.VolumeMin},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given an engine at the case's sound volume
			pool := &scriptedPlayers{}
			audio := newTestEngine(t, pool)
			audio.SetSoundVolume(tc.volume)

			// when playing one effect at no distance
			if err := audio.PlaySFX(snd.SfxMenuMove); err != nil {
				t.Fatalf("PlaySFX: %v", err)
			}

			// then the player carries the linear gain of the user volume, and
			// the slider floor switches the family off entirely
			// (snd_init, Source/engine/sound.cpp:254-262)
			if !audio.SoundOn() {
				if len(pool.created) != 0 {
					t.Fatalf("sound off still created %d players", len(pool.created))
				}

				return
			}

			want := float64(snd.VolumeLogToLinear(tc.volume, engine.VolumeMin, engine.VolumeMax))
			if got := pool.last(t).volume; math.Abs(got-want) > 1e-9 {
				t.Fatalf("player gain = %v, want %v", got, want)
			}
		})
	}
}

func TestAudioEnginePlaySfxLocDropsPastTheFloor(t *testing.T) {
	// given an engine whose player stands at the origin
	pool := &scriptedPlayers{}
	audio := newTestEngine(t, pool)
	audio.playerTile = func() snd.Point { return snd.Point{} }

	// when a sound next to the player and one far away are played
	if err := audio.PlaySfxLoc(snd.SfxMenuMove, snd.Point{X: 1, Y: 0}); err != nil {
		t.Fatalf("PlaySfxLoc near: %v", err)
	}

	near := len(pool.created)

	if err := audio.PlaySfxLoc(snd.SfxMenuMove, snd.Point{X: 102, Y: 0}); err != nil {
		t.Fatalf("PlaySfxLoc far: %v", err)
	}

	// then only the audible one reached the pool: 102 tiles is exactly the
	// attenuation floor (CalculateSoundPosition returns false there)
	if near != 1 || len(pool.created) != near {
		t.Fatalf("players created = %d after the far sound, want %d (inaudible dropped)", len(pool.created), near)
	}
}

func TestAudioEngineMusicVolumeFloorStopsTheTrack(t *testing.T) {
	// given an engine with a loaded looping track
	pool := &scriptedPlayers{}
	audio := newTestEngine(t, pool)
	audio.newMusic = func([]byte, float64) (sfxPlayer, error) {
		player := &fakePlayer{}
		pool.created = append(pool.created, player)

		return player, nil
	}

	if err := audio.StartMusic("track", testWave(t)); err != nil {
		t.Fatalf("StartMusic: %v", err)
	}

	// when the music slider reaches the floor
	audio.SetMusicVolume(engine.VolumeMin)

	// then gbMusicOn is false, the track stopped and nothing is loaded
	// (sound_get_or_set_music_volume, Source/engine/sound.cpp:383-394)
	if audio.MusicOn() {
		t.Fatal("music still on at the slider floor, want off")
	}

	if audio.Track() != musicTrackNone {
		t.Fatalf("track = %q at the slider floor, want none", audio.Track())
	}

	if !pool.last(t).closed {
		t.Fatal("the music player was not closed at the slider floor")
	}
}

func TestAudioBusIgnoresEventsThatNameNoSound(t *testing.T) {
	// given an engine and every kind the kernel emits today
	pool := &scriptedPlayers{}
	audio := newTestEngine(t, pool)

	batch := []kernel.Event{
		{Tick: 1, Kind: kernel.KindTickProcessed},
		{Tick: 1, Kind: kernel.KindStickPush, A: 1},
		{Tick: 1, Kind: kernel.KindMouseAttack, A: 3, B: 4},
		{Tick: 1, Kind: kernel.KindLooted, A: 0},
		{Tick: 1, Kind: kernel.KindTickRateSet, A: 30},
	}

	// when the subscriber plays the batch
	playEventSounds(batch, audio)

	// then nothing played: no kernel event carries a sound id yet
	// (sfxForEvent's table is empty; Source/effects.cpp plays inside the sim)
	if len(pool.created) != 0 {
		t.Fatalf("the bus played %d sounds for sound-less events, want 0", len(pool.created))
	}
}

func TestAudioMixAppliesPerChannelGains(t *testing.T) {
	// given two stereo 16-bit frames
	pcm := []byte{
		0x00, 0x04, 0x00, 0x04,
		0x00, 0x04, 0x00, 0x04,
	}

	// when the left channel is halved and the right one is untouched
	mixed := mixPCM(pcm, 0.5, 1)

	// then only the left channel's sample changed
	if got := int16(binary.LittleEndian.Uint16(mixed[0:2])); got != 0x0200 {
		t.Fatalf("left sample = %d, want 0x0200", got)
	}

	if got := int16(binary.LittleEndian.Uint16(mixed[2:4])); got != 0x0400 {
		t.Fatalf("right sample = %d, want 0x0400", got)
	}

	// and equal unity gains pass the buffer through untouched
	if got := mixPCM(pcm, 1, 1); &got[0] != &pcm[0] {
		t.Fatal("unity gains copied the buffer, want the same slice")
	}
}

// newTestEngine builds a device-backed engine over the scripted player factory.
func newTestEngine(t *testing.T, pool *scriptedPlayers) *audioEngine {
	t.Helper()

	return newAudioEngine(nil, testCatalog(t), pool.factory, nil)
}

// testCatalog builds a one-effect catalog over a synthetic retail-shaped WAV.
func testCatalog(t *testing.T) *snd.SfxCatalog {
	t.Helper()

	source := snd.SfxSourceFunc(func(string) ([]byte, error) { return testWave(t), nil })

	effects := []snd.Effect{{ID: string(snd.SfxMenuMove), Flags: []string{"Misc"}, Path: `sfx\misc\walk1.wav`}}

	catalog, err := snd.NewSfxCatalog(effects, source)
	if err != nil {
		t.Fatalf("NewSfxCatalog: %v", err)
	}

	return catalog
}

// testWave builds a minimal linear-PCM WAV: stereo, 22050 Hz, 16-bit, two
// frames — the format the retail probe measured (probe_test.go).
func testWave(t *testing.T) []byte {
	t.Helper()

	const (
		channels      = 2
		sampleRate    = engine.DefaultAudioSampleRate
		bitsPerSample = 16
		frames        = 2
	)

	data := make([]byte, frames*channels*bitsPerSample/8)
	for i := 0; i < len(data); i += 2 {
		binary.LittleEndian.PutUint16(data[i:i+2], 0x0400)
	}

	return wavContainer(data, channels, sampleRate, bitsPerSample)
}

// wavContainer wraps PCM data in a RIFF/WAVE header.
func wavContainer(data []byte, channels, sampleRate, bitsPerSample int) []byte {
	const fmtSize = 16

	buf := make([]byte, 0, 12+8+fmtSize+8+len(data))

	buf = append(buf, "RIFF"...)
	buf = binary.LittleEndian.AppendUint32(buf, 0) // patched below
	buf = append(buf, "WAVE"...)

	buf = append(buf, "fmt "...)
	buf = binary.LittleEndian.AppendUint32(buf, fmtSize)
	buf = binary.LittleEndian.AppendUint16(buf, 1) // WAVE_FORMAT_PCM
	buf = binary.LittleEndian.AppendUint16(buf, uint16(channels))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(sampleRate))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(sampleRate*channels*bitsPerSample/8))
	buf = binary.LittleEndian.AppendUint16(buf, uint16(channels*bitsPerSample/8))
	buf = binary.LittleEndian.AppendUint16(buf, uint16(bitsPerSample))

	buf = append(buf, "data"...)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(data)))
	buf = append(buf, data...)

	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(buf)-8))

	return buf
}
