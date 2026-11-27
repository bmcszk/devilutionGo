package engine

// TickRateDefault mirrors upstream's tickRate ctor default
// (Source/options.cpp:843).
const TickRateDefault = 20

// TickRateMin/TickRateMax come from sgGameInitInfo bounds enforced in
// gamemenu.cpp:276-277 (gnTickDelay = 1000 / tickRate).
const (
	TickRateMin = 20
	TickRateMax = 50
)

// DefaultAudioSampleRate is DEFAULT_AUDIO_SAMPLE_RATE
// (Source/options.cpp:55).
const DefaultAudioSampleRate = 22050

// Volume and positional-attenuation bounds (Source/engine/sound_defs.hpp:5-16).
const (
	// VolumeMin is the slider's floor: at it, sound and music switch off
	// (VOLUME_MIN, Source/engine/sound_defs.hpp:5).
	VolumeMin = -1600
	// VolumeMax is the slider's ceiling (VOLUME_MAX, :6).
	VolumeMax = 0
	// VolumeSteps is the Music/Sound slider granularity (:7).
	VolumeSteps = 64
	// AttenuationMin is the distance attenuation floor: a sound whose
	// computed volume is at or below it is dropped (:9).
	AttenuationMin = -6400
	// AttenuationMax is the distance attenuation ceiling (:10).
	AttenuationMax = 0
	// PanMin/PanMax bound the stereo pan value (:13,:14).
	PanMin = -6400
	PanMax = 6400
)

// Brightness slider bounds: 0..100 over 21 steps
// (Source/gamemenu.cpp:152-156).
const (
	// BrightnessMin is the gamma slider's floor.
	BrightnessMin = 0
	// BrightnessMax is the gamma slider's ceiling.
	BrightnessMax = 100
	// BrightnessSteps is the Gamma slider granularity
	// (gmenu_slider_steps(&sgOptionsMenu[2], 21), Source/gamemenu.cpp:154).
	BrightnessSteps = 21
	// DefaultBrightness is Graphics/"Brightness Correction"'s ctor default
	// (Source/options.cpp:796).
	DefaultBrightness = 0
)

// AudioSampleRates is the sampleRate entry list
// (Source/options.cpp:520).
var AudioSampleRates = [3]int{22050, 44100, 48000}

// SimOptions is the sim-relevant slice of the Gameplay/Audio categories:
// gameplay ticks per second and the audio output rate.
type SimOptions struct {
	TickRate   int
	SampleRate int
}

// UserOptions is the display/audio half of diablo.ini: the three keys the
// in-game Options submenu drives. It stays separate from SimOptions so the
// sim-relevant snapshot is not muddied by display knobs.
type UserOptions struct {
	// SoundVolume is Audio/"Sound Volume" (Source/options.cpp:514).
	SoundVolume int
	// MusicVolume is Audio/"Music Volume" (Source/options.cpp:516).
	MusicVolume int
	// Brightness is Graphics/"Brightness Correction" (Source/options.cpp:796).
	Brightness int
}

// DefaultSimOptions returns upstream's out-of-the-box values.
func DefaultSimOptions() SimOptions {
	return SimOptions{
		TickRate:   TickRateDefault,
		SampleRate: DefaultAudioSampleRate,
	}
}

// LoadSimOptions reads the typed tree from a parsed ini using the same
// section/key names as upstream categories "Game"->"Speed"
// (Source/options.cpp:842-843) and "Audio"->"Sample Rate"
// (Source/options.cpp:520). Missing or empty keys keep their defaults like
// upstream getters (options.cpp:226). Recorded deviation: upstream aborts
// via app_fatal on garbage values (GetInt/GetFloat error path), while this
// port returns the default and reports the error to the caller; whole-string
// parsing also rejects "+30" where upstream from_chars would accept it.
func LoadSimOptions(in *Ini) SimOptions {
	opts := DefaultSimOptions()
	if v, err := in.GetInt("Game", "Speed", opts.TickRate); err == nil && v >= TickRateMin && v <= TickRateMax {
		opts.TickRate = v
	}
	if v, err := in.GetInt("Audio", "Sample Rate", opts.SampleRate); err == nil {
		opts.SampleRate = v
	}
	return opts
}

// SaveSimOptions writes the typed tree back into the ini.
func (opts *SimOptions) SaveSimOptions(in *Ini) {
	in.SetInt("Game", "Speed", opts.TickRate)
	in.SetInt("Audio", "Sample Rate", opts.SampleRate)
}

// GnTickDelay maps tickRate to the engine tick delay like
// gnTickDelay = 1000 / nTickRate (Source/gamemenu.cpp:277,
// Source/diablo.cpp:146).
func (opts SimOptions) GnTickDelay() int {
	return 1000 / opts.TickRate
}

// ClampTickRate pins the rate into the upstream-accepted range so
// GnTickDelay never divides by zero.
func (opts *SimOptions) ClampTickRate() {
	if opts.TickRate < TickRateMin {
		opts.TickRate = TickRateMin
	}
	if opts.TickRate > TickRateMax {
		opts.TickRate = TickRateMax
	}
}

// ValidAudioSampleRate reports whether the rate is one of the upstream
// entries (Source/options.cpp:520).
func ValidAudioSampleRate(rate int) bool {
	for _, candidate := range AudioSampleRates {
		if candidate == rate {
			return true
		}
	}
	return false
}

// DefaultUserOptions returns upstream's out-of-the-box values: both volumes at
// VOLUME_MAX and brightness at 0 (Source/options.cpp:514,:516,:796).
func DefaultUserOptions() UserOptions {
	return UserOptions{SoundVolume: VolumeMax, MusicVolume: VolumeMax, Brightness: DefaultBrightness}
}

// LoadUserOptions reads the typed tree from a parsed ini using the upstream
// section/key names and clamps out-of-range values. Missing, empty or garbage
// values keep their defaults, and the clamp mirrors CapVolume
// (Source/engine/sound.cpp:168-171) applied by snd_init (:254-262).
func LoadUserOptions(in *Ini) UserOptions {
	opts := DefaultUserOptions()
	if v, err := in.GetInt("Audio", "Sound Volume", opts.SoundVolume); err == nil {
		opts.SoundVolume = CapVolume(v)
	}

	if v, err := in.GetInt("Audio", "Music Volume", opts.MusicVolume); err == nil {
		opts.MusicVolume = CapVolume(v)
	}

	if v, err := in.GetInt("Graphics", "Brightness Correction", opts.Brightness); err == nil {
		opts.Brightness = CapBrightness(v)
	}

	return opts
}

// SaveUserOptions writes the three keys in GetCategories order — Graphics
// before Audio (Source/options.h:905-920) — so a freshly written diablo.ini
// matches the shape a devilutionX build writes.
func (opts *UserOptions) SaveUserOptions(in *Ini) {
	in.SetInt("Graphics", "Brightness Correction", CapBrightness(opts.Brightness))
	in.SetInt("Audio", "Sound Volume", CapVolume(opts.SoundVolume))
	in.SetInt("Audio", "Music Volume", CapVolume(opts.MusicVolume))
}

// CapVolume clamps a volume into the slider range (CapVolume,
// Source/engine/sound.cpp:168-171).
func CapVolume(volume int) int {
	return min(max(volume, VolumeMin), VolumeMax)
}

// CapBrightness clamps the gamma level into the slider range
// (Source/gamemenu.cpp:152-156).
func CapBrightness(brightness int) int {
	return min(max(brightness, BrightnessMin), BrightnessMax)
}
