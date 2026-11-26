package engine_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/engine"
)

// Reference: Source/options.cpp:843 (tickRate default 20),
// Source/options.cpp:520 (Sample Rate default 22050 with entries
// {22050, 44100, 48000}), Source/gamemenu.cpp:277
// (gnTickDelay = 1000 / tickRate).

func TestSimOptionsDefaults(t *testing.T) {
	// when loading defaults
	opts := engine.DefaultSimOptions()

	// then they match the upstream ctor defaults
	if opts.TickRate != 20 {
		t.Errorf("TickRate = %d; want 20", opts.TickRate)
	}
	if opts.SampleRate != 22050 {
		t.Errorf("SampleRate = %d; want 22050", opts.SampleRate)
	}
}

func TestSimOptionsLoadReadsTypedKeys(t *testing.T) {
	// given an ini with the upstream section/key names
	in, err := engine.ParseIni([]byte("[Game]\nSpeed=30\n\n[Audio]\nSample Rate=48000\n"))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}

	// when loading the typed tree
	opts := engine.LoadSimOptions(in)

	// then both keys are picked up
	if opts.TickRate != 30 {
		t.Errorf("TickRate = %d; want 30", opts.TickRate)
	}
	if opts.SampleRate != 48000 {
		t.Errorf("SampleRate = %d; want 48000", opts.SampleRate)
	}
}

func TestSimOptionsLoadIgnoresUnknownKeysAndGarbage(t *testing.T) {
	// given an ini with unknown keys, unknown section, and garbage values
	in, err := engine.ParseIni([]byte("[Game]\nSpeed=abc\nBogus=1\n\n[Audio]\nSample Rate=xyz\n\n[Unknown]\nKey=1\n"))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}

	// when loading the typed tree
	opts := engine.LoadSimOptions(in)

	// then defaults survive (unknown keys ignored silently, ini.cpp:226)
	if opts.TickRate != 20 {
		t.Errorf("TickRate = %d; want default 20", opts.TickRate)
	}
	if opts.SampleRate != 22050 {
		t.Errorf("SampleRate = %d; want default 22050", opts.SampleRate)
	}
}

func TestSimOptionsRoundTripKeepsValues(t *testing.T) {
	// given loaded sim options
	in, err := engine.ParseIni([]byte("[Game]\nSpeed=42\n\n[Audio]\nSample Rate=44100\n"))
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}
	opts := engine.LoadSimOptions(in)

	// when saving them back into the same document
	opts.SaveSimOptions(in)

	// then a reload sees the same values
	reloaded := engine.LoadSimOptions(in)
	if reloaded.TickRate != 42 || reloaded.SampleRate != 44100 {
		t.Errorf("reloaded = %+v; want 42/44100", reloaded)
	}
}

func TestGnTickDelayMapping(t *testing.T) {
	// given tick rates across the accepted range
	for _, tc := range []struct {
		tickRate int
		want     int
	}{
		{20, 50},
		{25, 40},
		{50, 20},
	} {
		// when mapping to the engine tick delay
		opts := engine.SimOptions{TickRate: tc.tickRate}

		// then gnTickDelay = 1000 / tickRate (gamemenu.cpp:277)
		if got := opts.GnTickDelay(); got != tc.want {
			t.Errorf("GnTickDelay(%d) = %d; want %d", tc.tickRate, got, tc.want)
		}
	}
}

func TestClampTickRate(t *testing.T) {
	// given rates outside the accepted 20..50 window
	for _, tc := range []struct {
		in, want int
	}{
		{10, 20},
		{20, 20},
		{35, 35},
		{50, 50},
		{99, 50},
	} {
		// when clamping
		opts := engine.SimOptions{TickRate: tc.in}
		opts.ClampTickRate()

		// then the rate lands inside the window
		if opts.TickRate != tc.want {
			t.Errorf("ClampTickRate(%d) = %d; want %d", tc.in, opts.TickRate, tc.want)
		}
	}
}

func TestValidAudioSampleRate(t *testing.T) {
	// when probing rates against the upstream entry list
	for _, tc := range []struct {
		rate int
		want bool
	}{
		{22050, true},
		{44100, true},
		{48000, true},
		{8000, false},
		{96000, false},
	} {
		// then only listed rates are valid
		if got := engine.ValidAudioSampleRate(tc.rate); got != tc.want {
			t.Errorf("ValidAudioSampleRate(%d) = %v; want %v", tc.rate, got, tc.want)
		}
	}
}

// --- UserOptions: the display/audio half of diablo.ini (Options slice S1) ---
//
// Reference: Source/options.cpp:514 ("Sound Volume", VOLUME_MAX), :516
// ("Music Volume", VOLUME_MAX), :796 ("Brightness Correction", 0);
// Source/engine/sound_defs.hpp:5-16 (VOLUME_MIN/VOLUME_MAX/VOLUME_STEPS/
// ATTENUATION_MIN/PAN_MIN/PAN_MAX); Source/engine/sound.cpp:168-171 (CapVolume);
// Source/options.h:905-920 (GetCategories order: Graphics before Audio).

func TestUserOptionsDefaults(t *testing.T) {
	// when loading the out-of-the-box values
	opts := engine.DefaultUserOptions()

	// then both volumes sit at VOLUME_MAX and brightness at 0
	if opts.SoundVolume != 0 || opts.MusicVolume != 0 || opts.Brightness != 0 {
		t.Fatalf("DefaultUserOptions = %+v, want all zero", opts)
	}
}

func TestUserOptionsLoadKeepsDefaultsForAbsentOrUnreadableKeys(t *testing.T) {
	for _, tc := range []struct {
		name string
		doc  string
	}{
		{name: "absent keys", doc: "[Audio]\nSample Rate=22050\n"},
		{name: "empty values", doc: "[Audio]\nSound Volume=\nMusic Volume=\n\n[Graphics]\nBrightness Correction=\n"},
		{name: "garbage values", doc: "[Audio]\nSound Volume=loud\n\n[Graphics]\nBrightness Correction=dim\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given an ini carrying the case's keys
			in := mustParseIni(t, []byte(tc.doc))

			// when loading the typed tree
			got := engine.LoadUserOptions(in)

			// then the upstream defaults survive (Source/options.cpp:514-516, :796)
			if want := engine.DefaultUserOptions(); got != want {
				t.Fatalf("LoadUserOptions = %+v, want defaults %+v", got, want)
			}
		})
	}
}

func TestUserOptionsLoadReadsAndClampsTheKeys(t *testing.T) {
	for _, tc := range []struct {
		name string
		doc  string
		want engine.UserOptions
	}{
		{
			name: "typed keys are read",
			doc:  "[Audio]\nSound Volume=-800\nMusic Volume=-400\n\n[Graphics]\nBrightness Correction=42\n",
			want: engine.UserOptions{SoundVolume: -800, MusicVolume: -400, Brightness: 42},
		},
		{
			name: "out-of-range volumes clamp to the slider range",
			doc:  "[Audio]\nSound Volume=-9999\nMusic Volume=700\n",
			want: engine.UserOptions{SoundVolume: engine.VolumeMin, MusicVolume: engine.VolumeMax},
		},
		{
			name: "out-of-range brightness clamps to 0..100",
			doc:  "[Graphics]\nBrightness Correction=250\n",
			want: engine.UserOptions{Brightness: engine.BrightnessMax},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given an ini carrying the case's keys
			in := mustParseIni(t, []byte(tc.doc))

			// when loading the typed tree
			got := engine.LoadUserOptions(in)

			// then it matches the expectation
			if got != tc.want {
				t.Fatalf("LoadUserOptions = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestUserOptionsSaveWritesUpstreamSectionOrder(t *testing.T) {
	// given an empty document and non-default values
	in, err := engine.ParseIni(nil)
	if err != nil {
		t.Fatalf("ParseIni: %v", err)
	}

	opts := engine.UserOptions{SoundVolume: -800, MusicVolume: -400, Brightness: 42}

	// when saving them
	opts.SaveUserOptions(in)

	// then the sections are Graphics then Audio (GetCategories order,
	// Source/options.h:905-920) with the upstream key names
	want := "[Graphics]\r\nBrightness Correction=42\r\n\r\n[Audio]\r\nSound Volume=-800\r\nMusic Volume=-400"
	if got := strings.TrimSuffix(in.Serialize(), "\r\n"); got != want {
		t.Fatalf("Serialized =\n%q\nwant\n%q", got, want)
	}

	// and a reload sees the same values
	if reloaded := engine.LoadUserOptions(in); reloaded != opts {
		t.Fatalf("reloaded = %+v, want %+v", reloaded, opts)
	}
}

func TestUserOptionsSaveIsAChangeOnlyWhenAValueMoves(t *testing.T) {
	// given a document already carrying the values
	doc := "[Audio]\nSound Volume=-800\nMusic Volume=-400\n\n[Graphics]\nBrightness Correction=42\n"
	in := mustParseIni(t, []byte(doc))

	in.MarkAsUnchanged()

	// when saving the identical values back
	in2 := engine.LoadUserOptions(in)
	in2.SaveUserOptions(in)

	// then the document is not marked changed (SaveIni writes only a changed
	// ini, Source/options.cpp:169)
	if in.Changed() {
		t.Fatal("saving identical values marked the ini changed")
	}
}

func TestCapVolume(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   int
		want int
	}{
		{name: "below the minimum", in: -9999, want: engine.VolumeMin},
		{name: "at the minimum", in: engine.VolumeMin, want: engine.VolumeMin},
		{name: "in range", in: -800, want: -800},
		{name: "at the maximum", in: 0, want: engine.VolumeMax},
		{name: "above the maximum", in: 700, want: engine.VolumeMax},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given a volume outside or inside the slider range
			volume := tc.in

			// when clamping (CapVolume, Source/engine/sound.cpp:168-171)
			got := engine.CapVolume(volume)

			// then it lands inside the range
			if got != tc.want {
				t.Fatalf("CapVolume(%d) = %d, want %d", volume, got, tc.want)
			}
		})
	}
}

func TestCapBrightness(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   int
		want int
	}{
		{name: "below zero", in: -5, want: engine.BrightnessMin},
		{name: "in range", in: 42, want: 42},
		{name: "above 100", in: 250, want: engine.BrightnessMax},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// given a gamma level outside or inside the slider range
			level := tc.in

			// when clamping (Source/gamemenu.cpp:152-156)
			got := engine.CapBrightness(level)

			// then it lands inside the range
			if got != tc.want {
				t.Fatalf("CapBrightness(%d) = %d, want %d", level, got, tc.want)
			}
		})
	}
}
