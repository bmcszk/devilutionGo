package engine_test

import (
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
