package audio_test

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
	"github.com/bmcszk/devilutionGo/internal/engine"
)

// The sound-side positional math (Options slice S3): CalculateSoundPosition
// (Source/engine/sound_position.cpp:8-22), ApproxDistance
// (Source/engine/point.hpp:118-130), VolumeLogToLinear and PanLogToLeftRight
// (Source/utils/soundsample.cpp:69-102).

// volumeGoldenPath is the committed golden file.
const volumeGoldenPath = "../../../testdata/audio/volume_golden.tsv"

// positionResult is one CalculateSoundPosition outcome under test.
type positionResult struct {
	volume  int
	pan     int
	audible bool
}

// soundPositionCase is one CalculateSoundPosition expectation.
type soundPositionCase struct {
	name        string
	dx, dy      int
	wantPan     int
	wantVolume  int
	wantAudible bool
}

// approxDistanceCase is one ApproxDistance expectation.
type approxDistanceCase struct {
	name   string
	dx, dy int
	want   int
}

// volumeGoldenRow is one committed VolumeLogToLinear point.
type volumeGoldenRow struct {
	step      int
	logVolume int
	gainBits  uint32
}

// panGoldenRow is one committed PanLogToLeftRight point.
type panGoldenRow struct {
	logPan              int
	leftBits, rightBits uint32
}

func TestCalculateSoundPosition(t *testing.T) {
	for _, tc := range soundPositionCases() {
		t.Run(tc.name, func(t *testing.T) {
			// given the player at the origin and a sound at the case's offset
			player := audio.Point{}
			sound := audio.Point{X: tc.dx, Y: tc.dy}

			// when computing the positional volume and pan
			volume, pan, audible := audio.CalculateSoundPosition(player, sound)

			// then the pan is (dx-dy)*256 clamped and the volume is
			// -64 * ApproxDistance, dropped at or below ATTENUATION_MIN
			got := positionResult{volume: volume, pan: pan, audible: audible}
			if msg := soundPositionMismatch(tc, got); msg != "" {
				t.Fatal(msg)
			}
		})
	}
}

func TestCalculateSoundPositionIsDirectional(t *testing.T) {
	// given the player and sounds five tiles to the +dx side and to the -dx side
	player := audio.Point{X: 5, Y: 5}
	right := audio.Point{X: 10, Y: 5}
	left := audio.Point{X: 0, Y: 5}

	// when computing each side's pan
	_, panRight, _ := audio.CalculateSoundPosition(player, right)
	_, panLeft, _ := audio.CalculateSoundPosition(player, left)

	// then the pan flips sign around the player (Source/sound_position.cpp:13-14)
	if panRight <= 0 {
		t.Fatalf("a sound to the +dx side panned %d, want > 0", panRight)
	}

	if panLeft >= 0 {
		t.Fatalf("a sound to the -dx side panned %d, want < 0", panLeft)
	}
}

func TestApproxDistance(t *testing.T) {
	for _, tc := range approxDistanceCases() {
		t.Run(tc.name, func(t *testing.T) {
			// given two tiles the case's offset apart
			from := audio.Point{X: 10, Y: 10}
			to := audio.Point{X: 10 + tc.dx, Y: 10 + tc.dy}

			// when measuring the approximate distance
			got := audio.ApproxDistance(from, to)

			// then it matches the integer approximation
			// (Source/engine/point.hpp:118-130)
			if got != tc.want {
				t.Fatalf("ApproxDistance(delta %d,%d) = %d, want %d", tc.dx, tc.dy, got, tc.want)
			}
		})
	}
}

func TestVolumeLogToLinearMatchesTheGolden(t *testing.T) {
	for _, row := range volumeGoldenRows(t) {
		t.Run(strconv.Itoa(row.step), func(t *testing.T) {
			// given the slider step's log volume
			logVolume := row.logVolume

			// when converting it to a linear gain
			gain := audio.VolumeLogToLinear(logVolume, audio.VolumeMin, audio.VolumeMax)

			// then the float32 result matches the committed golden bit for bit
			if got := math.Float32bits(gain); got != row.gainBits {
				t.Fatalf("VolumeLogToLinear(%d) bits = 0x%08x (%.9g), want 0x%08x (%.9g)",
					logVolume, got, gain, row.gainBits, math.Float32frombits(row.gainBits))
			}
		})
	}
}

func TestVolumeLogToLinearIsMonotonicAcrossTheSlider(t *testing.T) {
	// given the committed slider steps in ascending order
	rows := volumeGoldenRows(t)

	// when walking them
	previous := float32(0)

	// then the gain rises with every step and stays inside [0,1]
	for _, row := range rows {
		t.Run(strconv.Itoa(row.step), func(t *testing.T) {
			gain := audio.VolumeLogToLinear(row.logVolume, audio.VolumeMin, audio.VolumeMax)

			if msg := sliderStepMismatch(row.step, gain, previous); msg != "" {
				t.Fatal(msg)
			}
		})
	}
}

func TestPanLogToLeftRightMatchesTheGolden(t *testing.T) {
	for _, row := range panGoldenRows(t) {
		t.Run(strconv.Itoa(row.logPan), func(t *testing.T) {
			// given the case's log pan
			logPan := row.logPan

			// when converting it to per-channel gains
			left, right := audio.PanLogToLeftRight(logPan)

			// then both gains match the committed golden bit for bit
			// (Source/utils/soundsample.cpp:83-102)
			if math.Float32bits(left) != row.leftBits || math.Float32bits(right) != row.rightBits {
				t.Fatalf("PanLogToLeftRight(%d) = 0x%08x/0x%08x (%.9g/%.9g), want 0x%08x/0x%08x",
					logPan, math.Float32bits(left), math.Float32bits(right), left, right,
					row.leftBits, row.rightBits)
			}
		})
	}
}

func TestSoundBoundsAgreeWithTheSettingsSlice(t *testing.T) {
	// The same sound_defs.hpp bounds are declared in two packages: the audio
	// domain needs the attenuation/pan range, the settings sliders need the
	// volume range. Each row is judged on its own.
	pairs := []struct {
		name      string
		audioVal  int
		engineVal int
	}{
		{name: "VolumeMin", audioVal: audio.VolumeMin, engineVal: engine.VolumeMin},
		{name: "VolumeMax", audioVal: audio.VolumeMax, engineVal: engine.VolumeMax},
		{name: "VolumeSteps", audioVal: audio.VolumeSteps, engineVal: engine.VolumeSteps},
		{name: "AttenuationMin", audioVal: audio.AttenuationMin, engineVal: engine.AttenuationMin},
		{name: "PanMin", audioVal: audio.PanMin, engineVal: engine.PanMin},
		{name: "PanMax", audioVal: audio.PanMax, engineVal: engine.PanMax},
	}

	for _, pair := range pairs {
		t.Run(pair.name, func(t *testing.T) {
			// when reading both copies
			got, want := pair.audioVal, pair.engineVal

			// then they carry the same value
			if got != want {
				t.Fatalf("%s = %d in the audio domain, %d in engine; want one value",
					pair.name, got, want)
			}
		})
	}
}

// soundPositionCases lists the hand-computed integer expectations.
func soundPositionCases() []soundPositionCase {
	return []soundPositionCase{
		{name: "centre", dx: 0, dy: 0, wantPan: 0, wantVolume: 0, wantAudible: true},
		{name: "two tiles along x", dx: 2, dy: 0, wantPan: 512, wantVolume: -128, wantAudible: true},
		{name: "one tile diagonal", dx: 1, dy: 1, wantPan: 0, wantVolume: -64, wantAudible: true},
		{name: "pan clamps at the +max", dx: 1000, dy: -1000, wantPan: audio.PanMax, wantAudible: false},
		{name: "pan clamps at the -max", dx: -1000, dy: 1000, wantPan: audio.PanMin, wantAudible: false},
		{
			name: "just inside the attenuation floor", dx: 101, dy: 0,
			wantPan: audio.PanMax, wantVolume: -6336, wantAudible: true,
		},
		{
			name: "at the attenuation floor is dropped", dx: 102, dy: 0,
			wantPan: audio.PanMax, wantAudible: false,
		},
	}
}

// approxDistanceCases lists the approximation's known points.
func approxDistanceCases() []approxDistanceCase {
	return []approxDistanceCase{
		{name: "same tile", dx: 0, dy: 0, want: 0},
		{name: "one tile along x", dx: 1, dy: 0, want: 1},
		{name: "one tile diagonal", dx: 1, dy: 1, want: 1},
		{name: "the 3-4-5 triangle", dx: 3, dy: 4, want: 5},
		{name: "negative offsets are absolute", dx: -3, dy: -4, want: 5},
		{name: "a near-square offset takes the correction", dx: 20, dy: 19, want: 27},
		{name: "an axis run of 102", dx: 102, dy: 0, want: 100},
	}
}

// volumeGoldenRows loads the golden's [volume] section.
func volumeGoldenRows(t *testing.T) []volumeGoldenRow {
	t.Helper()

	var rows []volumeGoldenRow

	for _, fields := range goldenSections(t)["volume"] {
		if fields[0] == "step" {
			continue
		}

		rows = append(rows, volumeGoldenRow{
			step:      mustAtoi(t, fields[0]),
			logVolume: mustAtoi(t, fields[1]),
			gainBits:  mustHex(t, fields[2]),
		})
	}

	if len(rows) == 0 {
		t.Fatal("the golden carries no volume rows")
	}

	return rows
}

// panGoldenRows loads the golden's [pan] section.
func panGoldenRows(t *testing.T) []panGoldenRow {
	t.Helper()

	var rows []panGoldenRow

	for _, fields := range goldenSections(t)["pan"] {
		if fields[0] == "log_pan" {
			continue
		}

		rows = append(rows, panGoldenRow{
			logPan:    mustAtoi(t, fields[0]),
			leftBits:  mustHex(t, fields[1]),
			rightBits: mustHex(t, fields[3]),
		})
	}

	if len(rows) == 0 {
		t.Fatal("the golden carries no pan rows")
	}

	return rows
}

// goldenSections parses the golden file into tab-separated rows per [section].
func goldenSections(t *testing.T) map[string][][]string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(volumeGoldenPath))
	if err != nil {
		t.Fatalf("read the volume golden: %v", err)
	}

	sections := map[string][][]string{}
	current := ""

	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if name, ok := strings.CutPrefix(line, "["); ok {
			current = strings.TrimSuffix(name, "]")

			continue
		}

		sections[current] = append(sections[current], strings.Split(line, "\t"))
	}

	return sections
}

// mustAtoi parses a decimal field or fails the test.
func mustAtoi(t *testing.T, field string) int {
	t.Helper()

	value, err := strconv.Atoi(field)
	if err != nil {
		t.Fatalf("parse %q as int: %v", field, err)
	}

	return value
}

// mustHex parses a 0x-prefixed float32 bit pattern or fails the test.
func mustHex(t *testing.T, field string) uint32 {
	t.Helper()

	digits, ok := strings.CutPrefix(field, "0x")
	if !ok {
		t.Fatalf("parse %q as hex bits: missing 0x prefix", field)
	}

	value, err := strconv.ParseUint(digits, 16, 32)
	if err != nil {
		t.Fatalf("parse %q as hex bits: %v", field, err)
	}

	return uint32(value)
}

// soundPositionMismatch reports the first broken expectation, empty when clean.
func soundPositionMismatch(tc soundPositionCase, got positionResult) string {
	if got.pan != tc.wantPan || got.audible != tc.wantAudible {
		return fmt.Sprintf("CalculateSoundPosition(delta %d,%d) = pan %d audible %v, want pan %d audible %v",
			tc.dx, tc.dy, got.pan, got.audible, tc.wantPan, tc.wantAudible)
	}

	if got.audible && got.volume != tc.wantVolume {
		return fmt.Sprintf("CalculateSoundPosition(delta %d,%d) volume = %d, want %d",
			tc.dx, tc.dy, got.volume, tc.wantVolume)
	}

	return ""
}

// sliderStepMismatch reports a non-monotonic or out-of-range slider step.
func sliderStepMismatch(step int, gain, previous float32) string {
	if gain < previous {
		return fmt.Sprintf("step %d gain %.9g fell below the previous step %.9g", step, gain, previous)
	}

	if gain < 0 || gain > 1 {
		return fmt.Sprintf("step %d gain %.9g is outside [0,1]", step, gain)
	}

	return ""
}
