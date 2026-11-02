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

// AudioSampleRates is the sampleRate entry list
// (Source/options.cpp:520).
var AudioSampleRates = [3]int{22050, 44100, 48000}

// SimOptions is the sim-relevant slice of the Gameplay/Audio categories:
// gameplay ticks per second and the audio output rate.
type SimOptions struct {
	TickRate   int
	SampleRate int
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
// (Source/options.cpp:520). Unknown or missing keys keep their defaults
// because every getter falls back 1:1 (Source/options.cpp:226).
func LoadSimOptions(in *Ini) SimOptions {
	opts := DefaultSimOptions()
	if v, err := in.GetInt("Game", "Speed", opts.TickRate); err == nil {
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
